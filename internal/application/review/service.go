package review

import (
	"errors"
	domainReview "kafka-order-demo/backend/internal/domain/review"
)

var (
	ErrReviewForbidden  = errors.New("you can only modify your own review")
	ErrProductNotBought = errors.New("Bạn chưa mua sản phẩm")
)

type Service struct {
	reviewRepo  ReviewRepository
	productRepo ProductRepository
	orderRepo   OrderRepository
}

func NewService(reviewRepo ReviewRepository, productRepo ProductRepository, orderRepo OrderRepository) *Service {
	return &Service{
		reviewRepo:  reviewRepo,
		productRepo: productRepo,
		orderRepo:   orderRepo,
	}
}

func (s *Service) GetProductReviews(productID uint) ([]ReviewResponse, error) {
	if _, err := s.productRepo.GetByID(productID); err != nil {
		return nil, err
	}

	reviews, err := s.reviewRepo.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	return toReviewResponses(reviews), nil
}

func (s *Service) CreateReview(req ReviewRequest) (*ReviewResponse, error) {
	if _, err := s.productRepo.GetByID(req.ProductID); err != nil {
		return nil, err
	}
	if err := s.ensureUserBoughtProduct(req.UserID, req.ProductID); err != nil {
		return nil, err
	}

	review, err := domainReview.NewReview(req.ProductID, req.UserID, req.Rating, req.Comment)
	if err != nil {
		return nil, err
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, err
	}

	response := toReviewResponse(review)
	return &response, nil
}

func (s *Service) UpdateReview(id uint, req ReviewUpdateRequest) (*ReviewResponse, error) {
	review, err := s.reviewRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if review.UserID != req.UserID {
		return nil, ErrReviewForbidden
	}

	if err := review.UpdateRating(req.Rating); err != nil {
		return nil, err
	}

	if err := s.reviewRepo.Update(review); err != nil {
		return nil, err
	}

	response := toReviewResponse(review)
	return &response, nil
}

func (s *Service) DeleteReview(id, userID uint) error {
	review, err := s.reviewRepo.GetByID(id)
	if err != nil {
		return err
	}
	if review.UserID != userID {
		return ErrReviewForbidden
	}

	return s.reviewRepo.Delete(id)
}

func (s *Service) ensureUserBoughtProduct(userID, productID uint) error {
	orders, err := s.orderRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	for _, order := range orders {
		for _, item := range order.OrderItems {
			if item.ProductID == productID {
				return nil
			}
		}
	}

	return ErrProductNotBought
}
