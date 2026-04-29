package category

import (
	domainCategory "kafka-order-demo/backend/internal/domain/category"
	"time"
)

type CategoryRequest struct {
	Name        string
	Description string
	Image       string
}

type CategoryResponse struct {
	ID          uint
	Name        string
	Description string
	Image       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func toCategoryResponse(category *domainCategory.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Image:       category.Image,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
	}
}

func toCategoryResponses(categories []*domainCategory.Category) []CategoryResponse {
	responses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = toCategoryResponse(category)
	}
	return responses
}
