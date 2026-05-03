package review

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidProductID = errors.New("product ID cannot be zero")
	ErrInvalidUserID    = errors.New("user ID cannot be zero")
	ErrInvalidRating    = errors.New("rating must be between 1 and 5")
	ErrReviewRequired   = errors.New("rating or comment is required")
)

type Review struct {
	ID        uint
	ProductID uint
	UserID    uint
	Rating    *int
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewReview(productID, userID uint, rating *int, comment *string) (*Review, error) {
	normalizedComment := ""
	if comment != nil {
		normalizedComment = strings.TrimSpace(*comment)
	}

	if productID == 0 {
		return nil, ErrInvalidProductID
	}
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if rating != nil && !isValidRating(*rating) {
		return nil, ErrInvalidRating
	}
	if rating == nil && normalizedComment == "" {
		return nil, ErrReviewRequired
	}

	now := time.Now()
	return &Review{
		ProductID: productID,
		UserID:    userID,
		Rating:    rating,
		Comment:   normalizedComment,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *Review) UpdateRating(rating int) error {
	if !isValidRating(rating) {
		return ErrInvalidRating
	}

	r.Rating = &rating
	r.UpdatedAt = time.Now()
	return nil
}

func isValidRating(rating int) bool {
	return rating >= 1 && rating <= 5
}
