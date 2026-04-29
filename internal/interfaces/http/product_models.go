package http

import (
	appProduct "kafka-order-demo/backend/internal/application/product"
	"time"
)

type productResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Image       string    `json:"image"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type productRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
}

type updateProductStockRequest struct {
	Stock int `json:"stock"`
}

func toProductInput(req productRequest) appProduct.ProductRequest {
	return appProduct.ProductRequest{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Image:       req.Image,
		Category:    req.Category,
	}
}

func toProductHTTPResponse(resp appProduct.ProductResponse) productResponse {
	return productResponse{
		ID:          resp.ID,
		Name:        resp.Name,
		Description: resp.Description,
		Price:       resp.Price,
		Stock:       resp.Stock,
		Image:       resp.Image,
		Category:    resp.Category,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
	}
}

func toProductHTTPResponses(responses []appProduct.ProductResponse) []productResponse {
	result := make([]productResponse, len(responses))
	for i, resp := range responses {
		result[i] = toProductHTTPResponse(resp)
	}
	return result
}
