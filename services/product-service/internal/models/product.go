package models

import (
	"time"
	"errors"
)

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"not null"`
	Image       string    `json:"image"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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

func (p *Product) IsAvailable(quantity int) bool {
	return p.Stock >= quantity
}

func (p *Product) UpdateStock(newStock int) error {
	if newStock < 0 {
		return errors.New("stock cannot be negative")
	}
	p.Stock = newStock
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) DecreaseStock(quantity int) error {
	if !p.IsAvailable(quantity) {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}