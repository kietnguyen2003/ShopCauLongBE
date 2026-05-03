package review

import (
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	domainReview "kafka-order-demo/backend/internal/domain/review"
)

type ReviewRepository interface {
	Create(review *domainReview.Review) error
	GetByID(id uint) (*domainReview.Review, error)
	GetByProductID(productID uint) ([]*domainReview.Review, error)
	Update(review *domainReview.Review) error
	Delete(id uint) error
}

type ProductRepository interface {
	GetByID(id uint) (*domainProduct.Product, error)
}

type OrderRepository interface {
	GetByUserID(userID uint) ([]*domainOrder.Order, error)
}
