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
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
