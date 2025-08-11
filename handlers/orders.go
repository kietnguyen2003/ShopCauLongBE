package handlers

import (
	"net/http"
	"strconv"
	"time"

	"kafka-order-demo/backend/kafka"
	"kafka-order-demo/backend/models"
	"kafka-order-demo/backend/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db           *gorm.DB
	kafkaManager *kafka.Manager
	wsHub        *websocket.Hub
}

type CreateOrderRequest struct {
	CustomerName string              `json:"customerName" binding:"required"`
	Phone        string              `json:"phone" binding:"required"`
	Address      string              `json:"address" binding:"required"`
	Email        string              `json:"email" binding:"required"`
	Total        float64             `json:"total" binding:"required"`
	Items        []OrderItemRequest  `json:"items" binding:"required,min=1"`
}

type OrderItemRequest struct {
	ID          string  `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func NewOrderHandler(db *gorm.DB, kafkaManager *kafka.Manager, wsHub *websocket.Hub) *OrderHandler {
	return &OrderHandler{
		db:           db,
		kafkaManager: kafkaManager,
		wsHub:        wsHub,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create order
	order := models.Order{
		UserID:       userID,
		Status:       "pending",
		TotalAmount:  req.Total,
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		Address:      req.Address,
		Email:        req.Email,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order items
	for _, itemReq := range req.Items {
		// Convert string ID to uint
		productID, err := strconv.ParseUint(itemReq.ID, 10, 32)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		orderItem := models.OrderItem{
			OrderID:     order.ID,
			ProductID:   uint(productID),
			Name:        itemReq.Name,
			Price:       itemReq.Price,
			Quantity:    itemReq.Quantity,
			Image:       itemReq.Image,
			Category:    itemReq.Category,
			Description: itemReq.Description,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order item"})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit order"})
		return
	}

	// Load complete order with items
	var completeOrder models.Order
	if err := h.db.Preload("OrderItems").Preload("User").First(&completeOrder, order.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load order details"})
		return
	}

	// Publish order created event to Kafka
	orderEvent := kafka.OrderEvent{
		OrderID:     order.ID,
		UserID:      userID,
		Type:        "created",
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		Timestamp:   time.Now(),
	}

	if err := h.kafkaManager.PublishOrderEvent(orderEvent); err != nil {
		// Log error but don't fail the request
		c.Header("X-Warning", "Failed to publish order event")
	}

	// Send WebSocket notification to admin
	h.wsHub.BroadcastToAdmin(websocket.Message{
		Type: "new_order",
		Data: map[string]interface{}{
			"order_id":     order.ID,
			"user_id":      userID,
			"total_amount": order.TotalAmount,
			"status":       order.Status,
		},
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   completeOrder,
	})
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, err := GetUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user"})
		return
	}

	var orders []models.Order
	if err := h.db.Preload("OrderItems").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	var orders []models.Order
	if err := h.db.Preload("OrderItems").Preload("User").Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderIDParam := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate status
	validStatuses := []string{"pending", "confirmed", "shipped", "delivered", "cancelled"}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Find and update order
	var order models.Order
	if err := h.db.Preload("User").First(&order, uint(orderID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	oldStatus := order.Status
	order.Status = req.Status

	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Publish order updated event to Kafka
	orderEvent := kafka.OrderEvent{
		OrderID:     order.ID,
		UserID:      order.UserID,
		Type:        "updated",
		Status:      req.Status,
		TotalAmount: order.TotalAmount,
		Timestamp:   time.Now(),
	}

	if err := h.kafkaManager.PublishOrderEvent(orderEvent); err != nil {
		// Log error but don't fail the request
		c.Header("X-Warning", "Failed to publish order event")
	}

	// Send WebSocket notification to user
	h.wsHub.SendToUser(order.UserID, websocket.Message{
		Type: "order_updated",
		Data: map[string]interface{}{
			"order_id":   order.ID,
			"old_status": oldStatus,
			"new_status": req.Status,
			"message":    getStatusMessage(req.Status),
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"order":   order,
	})
}

func getStatusMessage(status string) string {
	switch status {
	case "confirmed":
		return "Đơn hàng của bạn đã được xác nhận"
	case "shipped":
		return "Đơn hàng của bạn đã được giao vận"
	case "delivered":
		return "Đơn hàng của bạn đã được giao thành công"
	case "cancelled":
		return "Đơn hàng của bạn đã bị hủy"
	default:
		return "Trạng thái đơn hàng của bạn đã thay đổi"
	}
}