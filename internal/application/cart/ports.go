package cart

import (
	domainCart "kafka-order-demo/backend/internal/domain/cart"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type CartRepository interface {
	GetByUserID(userID uint) ([]*domainCart.CartItem, error)
	GetByIDAndUserID(id, userID uint) (*domainCart.CartItem, error)
	GetByUserIDAndProductID(userID, productID uint) (*domainCart.CartItem, error)
	Create(item *domainCart.CartItem) error
	Update(item *domainCart.CartItem) error
	Delete(id, userID uint) error
	Clear(userID uint) error
}

type ProductRepository interface {
	GetByID(id uint) (*domainProduct.Product, error)
}
