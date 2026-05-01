package order

import (
	domainAddress "kafka-order-demo/backend/internal/domain/address"
	domainCart "kafka-order-demo/backend/internal/domain/cart"
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type OrderRepository interface {
	Create(order *domainOrder.Order) error
	CreateWithProductStockUpdates(order *domainOrder.Order, products []*domainProduct.Product) error
	GetByID(id uint) (*domainOrder.Order, error)
	GetByUserID(userID uint) ([]*domainOrder.Order, error)
	GetAll() ([]*domainOrder.Order, error)
	Update(order *domainOrder.Order) error
	Delete(id uint) error
	DeleteAll() error
}

type ProductRepository interface {
	GetByID(id uint) (*domainProduct.Product, error)
}

type AddressRepository interface {
	GetByIDAndUserID(id, userID uint) (*domainAddress.Address, error)
}

type CartRepository interface {
	GetByUserID(userID uint) ([]*domainCart.CartItem, error)
	Clear(userID uint) error
}
