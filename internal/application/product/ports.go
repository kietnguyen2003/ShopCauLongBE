package product

import (
	domainCategory "kafka-order-demo/backend/internal/domain/category"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type ProductRepository interface {
	Create(product *domainProduct.Product) error
	GetByID(id uint) (*domainProduct.Product, error)
	GetAll() ([]*domainProduct.Product, error)
	GetWithQuery(query ProductQuery) ([]*domainProduct.Product, int64, error)
	GetByCategory(category string) ([]*domainProduct.Product, error)
	Search(keyword string) ([]*domainProduct.Product, error)
	Update(product *domainProduct.Product) error
	Delete(id uint) error
	UpdateStock(id uint, stock int) error
}

type CategoryRepository interface {
	GetByID(id uint) (*domainCategory.Category, error)
	GetByName(name string) (*domainCategory.Category, error)
}
