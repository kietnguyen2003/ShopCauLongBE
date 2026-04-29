package database

import (
	appCategory "kafka-order-demo/backend/internal/application/category"
	"kafka-order-demo/backend/internal/domain/category"

	"gorm.io/gorm"
)

type GormCategory struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string
	Image       string
	CreatedAt   int64
	UpdatedAt   int64
}

func (GormCategory) TableName() string {
	return "categories"
}

type GormCategoryRepository struct {
	db *gorm.DB
}

func NewGormCategoryRepository(db *gorm.DB) *GormCategoryRepository {
	return &GormCategoryRepository{db: db}
}

var _ appCategory.CategoryRepository = (*GormCategoryRepository)(nil)

func (r *GormCategoryRepository) Create(category *category.Category) error {
	gormCategory := &GormCategory{
		Name:        category.Name,
		Description: category.Description,
		Image:       category.Image,
		CreatedAt:   category.CreatedAt.Unix(),
		UpdatedAt:   category.UpdatedAt.Unix(),
	}

	if err := r.db.Create(gormCategory).Error; err != nil {
		return err
	}

	category.ID = gormCategory.ID
	return nil
}

func (r *GormCategoryRepository) GetByID(id uint) (*category.Category, error) {
	var gormCategory GormCategory
	if err := r.db.First(&gormCategory, id).Error; err != nil {
		return nil, err
	}

	return r.toDomainCategory(&gormCategory), nil
}

func (r *GormCategoryRepository) GetAll() ([]*category.Category, error) {
	var gormCategories []GormCategory
	if err := r.db.Find(&gormCategories).Error; err != nil {
		return nil, err
	}

	categories := make([]*category.Category, len(gormCategories))
	for i, gc := range gormCategories {
		categories[i] = r.toDomainCategory(&gc)
	}

	return categories, nil
}

func (r *GormCategoryRepository) Update(category *category.Category) error {
	gormCategory := &GormCategory{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
		Image:       category.Image,
		CreatedAt:   category.CreatedAt.Unix(),
		UpdatedAt:   category.UpdatedAt.Unix(),
	}

	return r.db.Save(gormCategory).Error
}

func (r *GormCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&GormCategory{}, id).Error
}

func (r *GormCategoryRepository) toDomainCategory(gormCategory *GormCategory) *category.Category {
	return &category.Category{
		ID:          gormCategory.ID,
		Name:        gormCategory.Name,
		Description: gormCategory.Description,
		Image:       gormCategory.Image,
		CreatedAt:   timeFromUnix(gormCategory.CreatedAt),
		UpdatedAt:   timeFromUnix(gormCategory.UpdatedAt),
	}
}
