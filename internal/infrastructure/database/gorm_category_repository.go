package database

import (
	appCategory "kafka-order-demo/backend/internal/application/category"
	"kafka-order-demo/backend/internal/domain/category"
	"time"

	"gorm.io/gorm"
)

type GormCategory struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string
	Image       string
	Status      string `gorm:"default:active"`
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
		Status:      categoryStatus(category.Status),
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
	if err := r.db.Where("id = ? AND (status = ? OR status = '' OR status IS NULL)", id, category.StatusActive).First(&gormCategory).Error; err != nil {
		return nil, err
	}

	return r.toDomainCategory(&gormCategory), nil
}

func (r *GormCategoryRepository) GetByName(name string) (*category.Category, error) {
	var gormCategory GormCategory
	if err := r.db.Where("name = ? AND (status = ? OR status = '' OR status IS NULL)", name, category.StatusActive).First(&gormCategory).Error; err != nil {
		return nil, err
	}

	return r.toDomainCategory(&gormCategory), nil
}

func (r *GormCategoryRepository) GetAll() ([]*category.Category, error) {
	var gormCategories []GormCategory
	if err := r.db.Where("status = ? OR status = '' OR status IS NULL", category.StatusActive).Find(&gormCategories).Error; err != nil {
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
		Status:      categoryStatus(category.Status),
		CreatedAt:   category.CreatedAt.Unix(),
		UpdatedAt:   category.UpdatedAt.Unix(),
	}

	return r.db.Save(gormCategory).Error
}

func (r *GormCategoryRepository) Delete(id uint) error {
	return r.db.Model(&GormCategory{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     category.StatusInactive,
		"updated_at": time.Now().Unix(),
	}).Error
}

func (r *GormCategoryRepository) toDomainCategory(gormCategory *GormCategory) *category.Category {
	return &category.Category{
		ID:          gormCategory.ID,
		Name:        gormCategory.Name,
		Description: gormCategory.Description,
		Image:       gormCategory.Image,
		Status:      categoryStatus(gormCategory.Status),
		CreatedAt:   timeFromUnix(gormCategory.CreatedAt),
		UpdatedAt:   timeFromUnix(gormCategory.UpdatedAt),
	}
}

func categoryStatus(status string) string {
	if status == "" {
		return category.StatusActive
	}

	return status
}
