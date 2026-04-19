package order

import (
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type OrderRepository interface {
	Create(order *domainOrder.Order) error
	GetByID(id uint) (*domainOrder.Order, error)
	GetByUserID(userID uint) ([]*domainOrder.Order, error)
	GetAll() ([]*domainOrder.Order, error)
	Update(order *domainOrder.Order) error
	Delete(id uint) error
	DeleteAll() error
}

type ProductRepository interface {
	GetByID(id uint) (*domainProduct.Product, error)
	Update(product *domainProduct.Product) error
}
