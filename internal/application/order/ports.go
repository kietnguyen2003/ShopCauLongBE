package order

import (
	appNotification "kafka-order-demo/backend/internal/application/notification"
	domainAddress "kafka-order-demo/backend/internal/domain/address"
	domainCart "kafka-order-demo/backend/internal/domain/cart"
	domainCoupon "kafka-order-demo/backend/internal/domain/coupon"
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type OrderRepository interface {
	Create(order *domainOrder.Order) error
	CreateWithProductStockUpdates(order *domainOrder.Order, products []*domainProduct.Product) error
	CreateWithProductStockUpdatesAndCoupons(order *domainOrder.Order, products []*domainProduct.Product, redemptions []*domainCoupon.Redemption) error
	GetByID(id uint) (*domainOrder.Order, error)
	GetByUserID(userID uint) ([]*domainOrder.Order, error)
	GetAll() ([]*domainOrder.Order, error)
	Update(order *domainOrder.Order) error
	UpdateWithProductRestock(order *domainOrder.Order) error
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

type CouponRepository interface {
	GetByCode(code string) (*domainCoupon.Coupon, error)
	CountRedemptionsByCouponIDAndUserID(couponID, userID uint) (int, error)
}

type NotificationPublisher interface {
	CreateOrderStatusNotification(userID, orderID uint, status string) (*appNotification.NotificationResponse, error)
}
