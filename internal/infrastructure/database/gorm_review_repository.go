package database

import (
	appReview "kafka-order-demo/backend/internal/application/review"
	"kafka-order-demo/backend/internal/domain/review"

	"gorm.io/gorm"
)

type GormReview struct {
	ID        uint `gorm:"primaryKey"`
	ProductID uint `gorm:"not null"`
	UserID    uint `gorm:"not null"`
	Rating    *int
	Comment   string
	CreatedAt int64
	UpdatedAt int64
}

func (GormReview) TableName() string {
	return "reviews"
}

type GormReviewRepository struct {
	db *gorm.DB
}

func NewGormReviewRepository(db *gorm.DB) *GormReviewRepository {
	return &GormReviewRepository{db: db}
}

var _ appReview.ReviewRepository = (*GormReviewRepository)(nil)

func (r *GormReviewRepository) Create(review *review.Review) error {
	gormReview := r.toGormReview(review)

	if err := r.db.Create(gormReview).Error; err != nil {
		return err
	}

	review.ID = gormReview.ID
	return nil
}

func (r *GormReviewRepository) GetByID(id uint) (*review.Review, error) {
	var gormReview GormReview
	if err := r.db.First(&gormReview, id).Error; err != nil {
		return nil, err
	}

	return r.toDomainReview(&gormReview), nil
}

func (r *GormReviewRepository) GetByProductID(productID uint) ([]*review.Review, error) {
	var gormReviews []GormReview
	if err := r.db.Where("product_id = ?", productID).Order("id DESC").Find(&gormReviews).Error; err != nil {
		return nil, err
	}

	reviews := make([]*review.Review, len(gormReviews))
	for i, gormReview := range gormReviews {
		reviews[i] = r.toDomainReview(&gormReview)
	}

	return reviews, nil
}

func (r *GormReviewRepository) Update(review *review.Review) error {
	return r.db.Save(r.toGormReview(review)).Error
}

func (r *GormReviewRepository) Delete(id uint) error {
	return r.db.Delete(&GormReview{}, id).Error
}

func (r *GormReviewRepository) toGormReview(review *review.Review) *GormReview {
	return &GormReview{
		ID:        review.ID,
		ProductID: review.ProductID,
		UserID:    review.UserID,
		Rating:    review.Rating,
		Comment:   review.Comment,
		CreatedAt: review.CreatedAt.Unix(),
		UpdatedAt: review.UpdatedAt.Unix(),
	}
}

func (r *GormReviewRepository) toDomainReview(gormReview *GormReview) *review.Review {
	return &review.Review{
		ID:        gormReview.ID,
		ProductID: gormReview.ProductID,
		UserID:    gormReview.UserID,
		Rating:    gormReview.Rating,
		Comment:   gormReview.Comment,
		CreatedAt: timeFromUnix(gormReview.CreatedAt),
		UpdatedAt: timeFromUnix(gormReview.UpdatedAt),
	}
}
