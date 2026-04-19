package database

import (
	"gorm.io/gorm"
	appProduct "kafka-order-demo/backend/internal/application/product"
	"kafka-order-demo/backend/internal/domain/product"
)

type GormProduct struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Price       float64 `gorm:"not null"`
	Stock       int     `gorm:"default:0"`
	Image       string
	Category    string
	CreatedAt   int64
	UpdatedAt   int64
}

func (GormProduct) TableName() string {
	return "products"
}

type GormProductRepository struct {
	db *gorm.DB
}

func NewGormProductRepository(db *gorm.DB) *GormProductRepository {
	return &GormProductRepository{db: db}
}

var _ appProduct.ProductRepository = (*GormProductRepository)(nil)

func (r *GormProductRepository) Create(prod *product.Product) error {
	gormProduct := &GormProduct{
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Stock:       prod.Stock,
		Image:       prod.Image,
		Category:    prod.Category,
		CreatedAt:   prod.CreatedAt.Unix(),
		UpdatedAt:   prod.UpdatedAt.Unix(),
	}

	err := r.db.Create(gormProduct).Error
	if err != nil {
		return err
	}

	prod.ID = gormProduct.ID
	return nil
}

func (r *GormProductRepository) GetByID(id uint) (*product.Product, error) {
	var gormProduct GormProduct
	err := r.db.First(&gormProduct, id).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainProduct(&gormProduct), nil
}

func (r *GormProductRepository) GetAll() ([]*product.Product, error) {
	var gormProducts []GormProduct
	err := r.db.Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}

	products := make([]*product.Product, len(gormProducts))
	for i, gp := range gormProducts {
		products[i] = r.toDomainProduct(&gp)
	}

	return products, nil
}

func (r *GormProductRepository) GetByCategory(category string) ([]*product.Product, error) {
	var gormProducts []GormProduct
	err := r.db.Where("category = ?", category).Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}

	products := make([]*product.Product, len(gormProducts))
	for i, gp := range gormProducts {
		products[i] = r.toDomainProduct(&gp)
	}

	return products, nil
}

func (r *GormProductRepository) Update(prod *product.Product) error {
	gormProduct := &GormProduct{
		ID:          prod.ID,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
		Stock:       prod.Stock,
		Image:       prod.Image,
		Category:    prod.Category,
		CreatedAt:   prod.CreatedAt.Unix(),
		UpdatedAt:   prod.UpdatedAt.Unix(),
	}

	return r.db.Save(gormProduct).Error
}

func (r *GormProductRepository) Delete(id uint) error {
	return r.db.Delete(&GormProduct{}, id).Error
}

func (r *GormProductRepository) UpdateStock(id uint, stock int) error {
	return r.db.Model(&GormProduct{}).Where("id = ?", id).Update("stock", stock).Error
}

func (r *GormProductRepository) toDomainProduct(gormProduct *GormProduct) *product.Product {
	return &product.Product{
		ID:          gormProduct.ID,
		Name:        gormProduct.Name,
		Description: gormProduct.Description,
		Price:       gormProduct.Price,
		Stock:       gormProduct.Stock,
		Image:       gormProduct.Image,
		Category:    gormProduct.Category,
		CreatedAt:   timeFromUnix(gormProduct.CreatedAt),
		UpdatedAt:   timeFromUnix(gormProduct.UpdatedAt),
	}
}
