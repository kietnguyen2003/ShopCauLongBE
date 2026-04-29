package http

import (
	appCategory "kafka-order-demo/backend/internal/application/category"
	"time"
)

type categoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image"`
}

type categoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func toCategoryInput(req categoryRequest) appCategory.CategoryRequest {
	return appCategory.CategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		Image:       req.Image,
	}
}

func toCategoryHTTPResponse(resp appCategory.CategoryResponse) categoryResponse {
	return categoryResponse{
		ID:          resp.ID,
		Name:        resp.Name,
		Description: resp.Description,
		Image:       resp.Image,
		CreatedAt:   resp.CreatedAt,
		UpdatedAt:   resp.UpdatedAt,
	}
}

func toCategoryHTTPResponses(responses []appCategory.CategoryResponse) []categoryResponse {
	result := make([]categoryResponse, len(responses))
	for i, resp := range responses {
		result[i] = toCategoryHTTPResponse(resp)
	}
	return result
}
