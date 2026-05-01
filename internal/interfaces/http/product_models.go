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
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type productListResponse struct {
	Items      []productResponse `json:"items"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	Total      int64             `json:"total"`
	TotalPages int               `json:"total_pages"`
}

type productRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
}

type productUpdateRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	Image       *string  `json:"image"`
	Category    *string  `json:"category"`
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

func toProductUpdateInput(req productUpdateRequest) appProduct.ProductUpdateRequest {
	return appProduct.ProductUpdateRequest{
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
		Status:      resp.Status,
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

func toProductListHTTPResponse(resp appProduct.ProductListResponse) productListResponse {
	return productListResponse{
		Items:      toProductHTTPResponses(resp.Items),
		Page:       resp.Page,
		Limit:      resp.Limit,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}
