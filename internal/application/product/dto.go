package product

import (
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	"time"
)

type ProductResponse struct {
	ID          uint
	Name        string
	Description string
	Price       int64
	Stock       int
	Image       string
	CategoryID  uint
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
	Price       int64
	Stock       int
	Image       string
	CategoryID  uint
	Category    string
}

type ProductUpdateRequest struct {
	Name        *string
	Description *string
	Price       *int64
	Stock       *int
	Image       *string
	CategoryID  *uint
	Category    *string
}

type ProductQuery struct {
	Page       int
	Limit      int
	Search     string
	CategoryID uint
	Category   string
	MinPrice   *int64
	MaxPrice   *int64
	Sort       string
}

func toProductResponse(prod *domainProduct.Product) ProductResponse {
	return ProductResponse{
		ID:          prod.ID,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Stock:       prod.Stock,
		Image:       prod.Image,
		CategoryID:  prod.CategoryID,
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
