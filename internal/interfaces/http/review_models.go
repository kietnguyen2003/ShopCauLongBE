package http

import (
	appReview "kafka-order-demo/backend/internal/application/review"
	"time"
)

type createReviewRequest struct {
	Rating  *int    `json:"rating"`
	Comment *string `json:"comment"`
}

type updateReviewRequest struct {
	Rating int `json:"rating" binding:"required"`
}

type reviewResponse struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	UserID    uint      `json:"user_id"`
	Rating    *int      `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toReviewInput(req createReviewRequest, productID, userID uint) appReview.ReviewRequest {
	return appReview.ReviewRequest{
		ProductID: productID,
		UserID:    userID,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}
}

func toReviewUpdateInput(req updateReviewRequest, userID uint) appReview.ReviewUpdateRequest {
	return appReview.ReviewUpdateRequest{
		UserID: userID,
		Rating: req.Rating,
	}
}

func toReviewHTTPResponse(resp appReview.ReviewResponse) reviewResponse {
	return reviewResponse{
		ID:        resp.ID,
		ProductID: resp.ProductID,
		UserID:    resp.UserID,
		Rating:    resp.Rating,
		Comment:   resp.Comment,
		CreatedAt: resp.CreatedAt,
		UpdatedAt: resp.UpdatedAt,
	}
}

func toReviewHTTPResponses(responses []appReview.ReviewResponse) []reviewResponse {
	result := make([]reviewResponse, len(responses))
	for i, resp := range responses {
		result[i] = toReviewHTTPResponse(resp)
	}
	return result
}
