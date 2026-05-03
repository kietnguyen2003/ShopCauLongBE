package review

import (
	domainReview "kafka-order-demo/backend/internal/domain/review"
	"time"
)

type ReviewRequest struct {
	ProductID uint
	UserID    uint
	Rating    *int
	Comment   *string
}

type ReviewUpdateRequest struct {
	UserID uint
	Rating int
}

type ReviewResponse struct {
	ID        uint
	ProductID uint
	UserID    uint
	Rating    *int
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func toReviewResponse(review *domainReview.Review) ReviewResponse {
	return ReviewResponse{
		ID:        review.ID,
		ProductID: review.ProductID,
		UserID:    review.UserID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt,
		UpdatedAt: review.UpdatedAt,
	}
}

func toReviewResponses(reviews []*domainReview.Review) []ReviewResponse {
	responses := make([]ReviewResponse, len(reviews))
	for i, review := range reviews {
		responses[i] = toReviewResponse(review)
	}
	return responses
}
