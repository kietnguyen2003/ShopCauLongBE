package order

import (
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	"time"
)

type CreateOrderRequest struct {
	UserID       uint
	AddressID    uint
	CouponCode   string
	CouponCodes  []string
	CustomerName string
	Phone        string
	Address      string
	Email        string
	Items        []CreateOrderItemRequest
}

type CreateOrderItemRequest struct {
	ProductID uint
	Quantity  int
}

type OrderItemResponse struct {
	ID          uint
	ProductID   uint
	Name        string
	Price       float64
	Quantity    int
	Image       string
	Category    string
	Description string
	CreatedAt   time.Time
}

type OrderResponse struct {
	ID             uint
	UserID         uint
	OrderItems     []OrderItemResponse
	SubtotalAmount float64
	DiscountAmount float64
	TotalAmount    float64
	CouponCode     string
	CouponCodes    []string
	Status         string
	CustomerName   string
	Phone          string
	Address        string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func toOrderResponse(ord *domainOrder.Order) OrderResponse {
	items := make([]OrderItemResponse, len(ord.OrderItems))
	for i, item := range ord.OrderItems {
		items[i] = OrderItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Name:        item.ProductSnapshot.Name,
			Price:       item.ProductSnapshot.Price,
			Quantity:    item.Quantity,
			Image:       item.ProductSnapshot.Image,
			Category:    item.ProductSnapshot.Category,
			Description: item.ProductSnapshot.Description,
			CreatedAt:   item.CreatedAt,
		}
	}

	return OrderResponse{
		ID:             ord.ID,
		UserID:         ord.UserID,
		OrderItems:     items,
		SubtotalAmount: ord.SubtotalAmount,
		DiscountAmount: ord.DiscountAmount,
		TotalAmount:    ord.TotalAmount,
		CouponCode:     ord.CouponCode,
		CouponCodes:    ord.CouponCodes,
		Status:         string(ord.Status),
		CustomerName:   ord.CustomerName,
		Phone:          ord.Phone,
		Address:        ord.Address,
		Email:          ord.Email,
		CreatedAt:      ord.CreatedAt,
		UpdatedAt:      ord.UpdatedAt,
	}
}

func toOrderResponses(orders []*domainOrder.Order) []OrderResponse {
	responses := make([]OrderResponse, len(orders))
	for i, ord := range orders {
		responses[i] = toOrderResponse(ord)
	}
	return responses
}
