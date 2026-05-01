package product

import (
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	"time"
)

type ProductResponse struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	Image       string
	Category    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ProductListResponse struct {
	Items      []ProductResponse
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

type ProductRequest struct {
	Name        string
	Description string
	Price       float64
	Stock       int
	Image       string
	Category    string
}

type ProductUpdateRequest struct {
	Name        *string
	Description *string
	Price       *float64
	Stock       *int
	Image       *string
	Category    *string
}

type ProductQuery struct {
	Page     int
	Limit    int
	Search   string
	Category string
	MinPrice *float64
	MaxPrice *float64
	Sort     string
}

func toProductResponse(prod *domainProduct.Product) ProductResponse {
	return ProductResponse{
		ID:          prod.ID,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Stock:       prod.Stock,
		Image:       prod.Image,
		Category:    prod.Category,
		Status:      prod.Status,
		CreatedAt:   prod.CreatedAt,
		UpdatedAt:   prod.UpdatedAt,
	}
}

func toProductResponses(products []*domainProduct.Product) []ProductResponse {
	responses := make([]ProductResponse, len(products))
	for i, prod := range products {
		responses[i] = toProductResponse(prod)
	}
	return responses
}
