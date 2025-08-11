package product

import (
	"time"
	"errors"
)

// Product represents the product domain entity
type Product struct {
	ID          uint
	Name        string
	Description string
	Price       float64
	Stock       int
	Image       string
	Category    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ProductRepository defines the interface for product data access
type ProductRepository interface {
	Create(product *Product) error
	GetByID(id uint) (*Product, error)
	GetAll() ([]*Product, error)
	GetByCategory(category string) ([]*Product, error)
	Update(product *Product) error
	Delete(id uint) error
	UpdateStock(id uint, stock int) error
}

// NewProduct creates a new product with validation
func NewProduct(name, description string, price float64, stock int, image, category string) (*Product, error) {
	if name == "" {
		return nil, errors.New("product name cannot be empty")
	}
	if price <= 0 {
		return nil, errors.New("product price must be greater than 0")
	}
	if stock < 0 {
		return nil, errors.New("product stock cannot be negative")
	}

	return &Product{
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
		Image:       image,
		Category:    category,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// IsAvailable checks if product is available in stock
func (p *Product) IsAvailable(quantity int) bool {
	return p.Stock >= quantity
}

// UpdateStock updates product stock
func (p *Product) UpdateStock(newStock int) error {
	if newStock < 0 {
		return errors.New("stock cannot be negative")
	}
	p.Stock = newStock
	p.UpdatedAt = time.Now()
	return nil
}

// DecreaseStock decreases product stock by quantity
func (p *Product) DecreaseStock(quantity int) error {
	if !p.IsAvailable(quantity) {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}