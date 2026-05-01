package http

import (
	appCart "kafka-order-demo/backend/internal/application/cart"
	"time"
)

type cartItemRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

type addCartItemsRequest struct {
	Items []cartItemRequest `json:"items"`
}

type updateCartItemRequest struct {
	Quantity int `json:"quantity"`
}

type cartItemResponse struct {
	ID          uint      `json:"id"`
	ProductID   uint      `json:"product_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	Image       string    `json:"image"`
	Subtotal    float64   `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type cartResponse struct {
	Items       []cartItemResponse `json:"items"`
	TotalAmount float64            `json:"total_amount"`
	TotalItems  int                `json:"total_items"`
}

func toCartItemInput(req cartItemRequest, userID uint) appCart.CartItemRequest {
	return appCart.CartItemRequest{
		UserID:    userID,
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	}
}

func toCartItemsInput(req addCartItemsRequest, userID uint) []appCart.CartItemRequest {
	items := make([]appCart.CartItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = toCartItemInput(item, userID)
	}

	return items
}

func toCartHTTPResponse(resp appCart.CartResponse) cartResponse {
	items := make([]cartItemResponse, len(resp.Items))
	for i, item := range resp.Items {
		items[i] = cartItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Quantity:    item.Quantity,
			Image:       item.Image,
			Subtotal:    item.Subtotal,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		}
	}

	return cartResponse{
		Items:       items,
		TotalAmount: resp.TotalAmount,
		TotalItems:  resp.TotalItems,
	}
}
