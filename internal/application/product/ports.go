package product

import domainProduct "kafka-order-demo/backend/internal/domain/product"

type ProductRepository interface {
	Create(product *domainProduct.Product) error
	GetByID(id uint) (*domainProduct.Product, error)
	GetAll() ([]*domainProduct.Product, error)
	GetByCategory(category string) ([]*domainProduct.Product, error)
	Update(product *domainProduct.Product) error
	Delete(id uint) error
	UpdateStock(id uint, stock int) error
}
