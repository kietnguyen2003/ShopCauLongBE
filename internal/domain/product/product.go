package product

import (
	"errors"
	"time"
)

var ErrInsufficientStock = errors.New("insufficient stock")

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// Product represents the product domain entity
type Product struct {
	ID          uint
	Name        string
	Description string
	Price       int64
	Stock       int
	Image       string
	CategoryID  uint
	Category    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewProduct creates a new product with validation
func NewProduct(name, description string, price int64, stock int, image string, categoryID uint, category string) (*Product, error) {
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
		CategoryID:  categoryID,
		Category:    category,
		Status:      StatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// IsAvailable checks if product is available in stock
func (p *Product) IsAvailable(quantity int) bool {
	return quantity > 0 && p.Stock >= quantity
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

// UpdateDetails updates product information with validation
func (p *Product) UpdateDetails(name, description string, price int64, stock int, image string, categoryID uint, category string) error {
	if name == "" {
		return errors.New("product name cannot be empty")
	}
	if price <= 0 {
		return errors.New("product price must be greater than 0")
	}
	if stock < 0 {
		return errors.New("product stock cannot be negative")
	}

	p.Name = name
	p.Description = description
	p.Price = price
	p.Stock = stock
	p.Image = image
	p.CategoryID = categoryID
	p.Category = category
	p.UpdatedAt = time.Now()
	return nil
}

// Deactivate marks product as unavailable without deleting its data.
func (p *Product) Deactivate() {
	p.Status = StatusInactive
	p.UpdatedAt = time.Now()
}

// DecreaseStock decreases product stock by quantity
func (p *Product) DecreaseStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if !p.IsAvailable(quantity) {
		return ErrInsufficientStock
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}
