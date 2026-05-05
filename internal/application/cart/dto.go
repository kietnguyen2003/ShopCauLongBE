package cart

import (
	domainCart "kafka-order-demo/backend/internal/domain/cart"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	"time"
)

type CartItemRequest struct {
	UserID    uint
	ProductID uint
	Quantity  int
}

type CartItemResponse struct {
	ID          uint
	ProductID   uint
	Name        string
	Description string
	Price       int64
	Quantity    int
	Image       string
	Subtotal    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CartResponse struct {
	Items       []CartItemResponse
	TotalAmount int64
	TotalItems  int
}

func toCartItemResponse(item *domainCart.CartItem, prod *domainProduct.Product) CartItemResponse {
	return CartItemResponse{
		ID:          item.ID,
		ProductID:   item.ProductID,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Quantity:    item.Quantity,
		Image:       prod.Image,
		Subtotal:    prod.Price * int64(item.Quantity),
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
}
