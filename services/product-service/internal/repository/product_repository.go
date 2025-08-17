package repository

import (
	"product-service/internal/models"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *models.Product) error
	GetByID(id uint) (*models.Product, error)
	GetAll() ([]*models.Product, error)
	GetByCategory(category string) ([]*models.Product, error)
	Update(product *models.Product) error
	Delete(id uint) error
	UpdateStock(id uint, stock int) error
}

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) GetByID(id uint) (*models.Product, error) {
	var product models.Product
	err := r.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetAll() ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCategory(category string) ([]*models.Product, error) {
	var products []*models.Product
	err := r.db.Where("category = ?", category).Find(&products).Error
	return products, err
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uint) error {
	return r.db.Delete(&models.Product{}, id).Error
}

func (r *productRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&models.Product{}).Where("id = ?", id).Update("stock", stock).Error
}