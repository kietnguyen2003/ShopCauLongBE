package http

import (
	appOrder "kafka-order-demo/backend/internal/application/order"
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	"time"
)

type createOrderRequest struct {
	AddressID uint `json:"address_id"`
}

type updateOrderStatusRequest struct {
	Status string `json:"status"`
}

type orderItemResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Image       string    `json:"image"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type orderResponse struct {
	ID           uint                `json:"id"`
	UserID       uint                `json:"user_id"`
	OrderItems   []orderItemResponse `json:"order_items"`
	TotalAmount  float64             `json:"total_amount"`
	Status       string              `json:"status"`
	CustomerName string              `json:"customer_name"`
	Phone        string              `json:"phone"`
	Address      string              `json:"address"`
	Email        string              `json:"email"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

func toCreateOrderInput(req createOrderRequest, userID uint) appOrder.CreateOrderRequest {
	return appOrder.CreateOrderRequest{
		UserID:    userID,
		AddressID: req.AddressID,
	}
}

func toOrderHTTPResponse(resp appOrder.OrderResponse) orderResponse {
	items := make([]orderItemResponse, len(resp.OrderItems))
	for i, item := range resp.OrderItems {
		items[i] = orderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Name:        item.Name,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Image:       item.Image,
			Category:    item.Category,
			Description: item.Description,
			CreatedAt:   item.CreatedAt,
		}
	}

	return orderResponse{
		ID:           resp.ID,
		UserID:       resp.UserID,
		OrderItems:   items,
		TotalAmount:  resp.TotalAmount,
		Status:       resp.Status,
		CustomerName: resp.CustomerName,
		Phone:        resp.Phone,
		Address:      resp.Address,
		Email:        resp.Email,
		CreatedAt:    resp.CreatedAt,
		UpdatedAt:    resp.UpdatedAt,
	}
}

func toOrderHTTPResponses(responses []appOrder.OrderResponse) []orderResponse {
	result := make([]orderResponse, len(responses))
	for i, resp := range responses {
		result[i] = toOrderHTTPResponse(resp)
	}
	return result
}

func parseOrderStatus(value string) domainOrder.OrderStatus {
	return domainOrder.OrderStatus(value)
}
