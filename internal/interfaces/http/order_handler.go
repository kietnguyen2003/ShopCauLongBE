package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	appOrder "kafka-order-demo/backend/internal/application/order"
)

type OrderHandler struct {
	orderService *appOrder.Service
}

func NewOrderHandler(orderService *appOrder.Service) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.AddressID == 0 {
		errorResponse(c, http.StatusBadRequest, "address_id is required")
		return
	}

	// Get user ID from auth middleware
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	ord, err := h.orderService.CreateOrder(toCreateOrderInput(req, userID.(uint)))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create order successfully", toOrderHTTPResponse(*ord))
}

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	orders, err := h.orderService.GetOrdersByUser(userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get orders successfully", toOrderHTTPResponses(orders))
}

func (h *OrderHandler) GetAllOrders(c *gin.Context) {
	orders, err := h.orderService.GetAllOrders()
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get all orders successfully", toOrderHTTPResponses(orders))
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid order ID")
		return
	}

	var req updateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	status := parseOrderStatus(req.Status)
	err = h.orderService.UpdateOrderStatus(uint(id), status)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Order status updated successfully", nil)
}
