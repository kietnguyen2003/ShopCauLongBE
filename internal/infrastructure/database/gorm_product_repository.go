package database

import (
	appProduct "kafka-order-demo/backend/internal/application/product"
	"kafka-order-demo/backend/internal/domain/product"
	"strings"
	"time"

	"gorm.io/gorm"
)

type GormProduct struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	Price       float64 `gorm:"not null"`
	Stock       int     `gorm:"default:0"`
	Image       string
	Category    string
	Status      string `gorm:"default:active"`
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
		Status:      productStatus(prod.Status),
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
	err := r.db.Where("id = ? AND (status = ? OR status = '' OR status IS NULL)", id, product.StatusActive).First(&gormProduct).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainProduct(&gormProduct), nil
}

func (r *GormProductRepository) GetAll() ([]*product.Product, error) {
	var gormProducts []GormProduct
	err := r.db.Where("status = ? OR status = '' OR status IS NULL", product.StatusActive).Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}

	products := make([]*product.Product, len(gormProducts))
	for i, gp := range gormProducts {
		products[i] = r.toDomainProduct(&gp)
	}

	return products, nil
}

func (r *GormProductRepository) GetWithQuery(query appProduct.ProductQuery) ([]*product.Product, int64, error) {
	var gormProducts []GormProduct
	var total int64

	dbQuery := r.db.Model(&GormProduct{}).Where("status = ? OR status = '' OR status IS NULL", product.StatusActive)
	if query.Search != "" {
		keyword := "%" + strings.ToLower(query.Search) + "%"
		dbQuery = dbQuery.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", keyword, keyword)
	}
	if query.Category != "" {
		dbQuery = dbQuery.Where("category = ?", query.Category)
	}
	if query.MinPrice != nil {
		dbQuery = dbQuery.Where("price >= ?", *query.MinPrice)
	}
	if query.MaxPrice != nil {
		dbQuery = dbQuery.Where("price <= ?", *query.MaxPrice)
	}

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := productOrderBy(query.Sort)
	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Order(orderBy).Limit(query.Limit).Offset(offset).Find(&gormProducts).Error; err != nil {
		return nil, 0, err
	}

	products := make([]*product.Product, len(gormProducts))
	for i, gp := range gormProducts {
		products[i] = r.toDomainProduct(&gp)
	}

	return products, total, nil
}

func (r *GormProductRepository) GetByCategory(category string) ([]*product.Product, error) {
	var gormProducts []GormProduct
	err := r.db.Where("category = ? AND (status = ? OR status = '' OR status IS NULL)", category, product.StatusActive).Find(&gormProducts).Error
	if err != nil {
		return nil, err
	}

	products := make([]*product.Product, len(gormProducts))
	for i, gp := range gormProducts {
		products[i] = r.toDomainProduct(&gp)
	}

	return products, nil
}

func (r *GormProductRepository) Search(keyword string) ([]*product.Product, error) {
	var gormProducts []GormProduct
	keyword = strings.ToLower(keyword)
	query := "%" + keyword + "%"
	err := r.db.Where("(status = ? OR status = '' OR status IS NULL) AND (LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(category) LIKE ?)", product.StatusActive, query, query, query).Find(&gormProducts).Error
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
		Status:      productStatus(prod.Status),
		CreatedAt:   prod.CreatedAt.Unix(),
		UpdatedAt:   prod.UpdatedAt.Unix(),
	}

	return r.db.Save(gormProduct).Error
}

func (r *GormProductRepository) Delete(id uint) error {
	return r.db.Model(&GormProduct{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     product.StatusInactive,
		"updated_at": time.Now().Unix(),
	}).Error
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
		Status:      productStatus(gormProduct.Status),
		CreatedAt:   timeFromUnix(gormProduct.CreatedAt),
		UpdatedAt:   timeFromUnix(gormProduct.UpdatedAt),
	}
}

func productStatus(status string) string {
	if status == "" {
		return product.StatusActive
	}

	return status
}

func productOrderBy(sort string) string {
	switch sort {
	case "price_asc":
		return "price ASC"
	case "price_desc":
		return "price DESC"
	case "name_asc":
		return "name ASC"
	case "name_desc":
		return "name DESC"
	case "oldest":
		return "created_at ASC"
	case "newest":
		return "created_at DESC"
	default:
		return "id DESC"
	}
}
