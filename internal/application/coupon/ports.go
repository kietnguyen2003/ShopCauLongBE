package coupon

import (
	domainCart "kafka-order-demo/backend/internal/domain/cart"
	domainCoupon "kafka-order-demo/backend/internal/domain/coupon"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
)

type CouponRepository interface {
	Create(coupon *domainCoupon.Coupon) error
	GetByID(id uint) (*domainCoupon.Coupon, error)
	GetByCode(code string) (*domainCoupon.Coupon, error)
	GetWithQuery(query CouponQuery) ([]*domainCoupon.Coupon, int64, error)
	GetUnusedActiveWithQuery(query CouponQuery, userID uint) ([]*domainCoupon.Coupon, int64, error)
	Update(coupon *domainCoupon.Coupon) error
	Delete(id uint) error
	CountRedemptionsByCouponIDAndUserID(couponID, userID uint) (int, error)
}

type CartRepository interface {
	GetByUserID(userID uint) ([]*domainCart.CartItem, error)
}

type ProductRepository interface {
	GetByID(id uint) (*domainProduct.Product, error)
}
