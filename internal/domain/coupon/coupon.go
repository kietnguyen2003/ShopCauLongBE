package coupon

import (
	"errors"
	"strings"
	"time"
)

const (
	DiscountTypePercentage  = "percentage"
	DiscountTypeFixedAmount = "fixed_amount"
)

var (
	ErrCouponCodeRequired       = errors.New("coupon code is required")
	ErrInvalidDiscountType      = errors.New("discount_type must be percentage or fixed_amount")
	ErrInvalidDiscountValue     = errors.New("discount_value must be greater than 0")
	ErrInvalidPercentageValue   = errors.New("percentage discount_value must be between 1 and 100")
	ErrInvalidMinOrderAmount    = errors.New("min_order_amount cannot be negative")
	ErrInvalidMaxDiscountAmount = errors.New("max_discount_amount must be greater than 0")
	ErrInvalidUsageLimit        = errors.New("usage_limit must be greater than 0")
	ErrInvalidUsageLimitPerUser = errors.New("usage_limit_per_user must be greater than 0")
	ErrInvalidCouponDates       = errors.New("end_at must be greater than start_at")
	ErrCouponInactive           = errors.New("coupon is inactive")
	ErrCouponNotStarted         = errors.New("coupon has not started yet")
	ErrCouponExpired            = errors.New("coupon expired")
	ErrMinOrderNotMet           = errors.New("order amount does not meet minimum requirement")
	ErrCouponUsageLimitReached  = errors.New("coupon usage limit reached")
	ErrCouponAlreadyUsed        = errors.New("you have already used this coupon")
)

type Coupon struct {
	ID                uint
	Code              string
	Name              string
	Description       string
	DiscountType      string
	DiscountValue     float64
	MinOrderAmount    float64
	MaxDiscountAmount *float64
	UsageLimit        *int
	UsedCount         int
	UsageLimitPerUser int
	StartAt           *time.Time
	EndAt             *time.Time
	IsActive          bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type Redemption struct {
	ID             uint
	CouponID       uint
	UserID         uint
	OrderID        uint
	DiscountAmount float64
	CreatedAt      time.Time
}

func NewCoupon(code, name, description, discountType string, discountValue, minOrderAmount float64, maxDiscountAmount *float64, usageLimit *int, usageLimitPerUser int, startAt, endAt *time.Time, isActive bool) (*Coupon, error) {
	coupon := &Coupon{
		Code:              NormalizeCode(code),
		Name:              name,
		Description:       description,
		DiscountType:      discountType,
		DiscountValue:     discountValue,
		MinOrderAmount:    minOrderAmount,
		MaxDiscountAmount: maxDiscountAmount,
		UsageLimit:        usageLimit,
		UsageLimitPerUser: usageLimitPerUser,
		StartAt:           startAt,
		EndAt:             endAt,
		IsActive:          isActive,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
	if coupon.UsageLimitPerUser == 0 {
		coupon.UsageLimitPerUser = 1
	}
	if err := coupon.ValidateDefinition(); err != nil {
		return nil, err
	}

	return coupon, nil
}

func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func (c *Coupon) ValidateDefinition() error {
	if c.Code == "" {
		return ErrCouponCodeRequired
	}
	if c.DiscountType != DiscountTypePercentage && c.DiscountType != DiscountTypeFixedAmount {
		return ErrInvalidDiscountType
	}
	if c.DiscountValue <= 0 {
		return ErrInvalidDiscountValue
	}
	if c.DiscountType == DiscountTypePercentage && (c.DiscountValue < 1 || c.DiscountValue > 100) {
		return ErrInvalidPercentageValue
	}
	if c.MinOrderAmount < 0 {
		return ErrInvalidMinOrderAmount
	}
	if c.MaxDiscountAmount != nil && *c.MaxDiscountAmount <= 0 {
		return ErrInvalidMaxDiscountAmount
	}
	if c.UsageLimit != nil && *c.UsageLimit <= 0 {
		return ErrInvalidUsageLimit
	}
	if c.UsageLimitPerUser <= 0 {
		return ErrInvalidUsageLimitPerUser
	}
	if c.StartAt != nil && c.EndAt != nil && !c.EndAt.After(*c.StartAt) {
		return ErrInvalidCouponDates
	}

	return nil
}

func (c *Coupon) ValidateForUse(subtotalAmount float64, userUsedCount int, now time.Time) error {
	if !c.IsActive {
		return ErrCouponInactive
	}
	if c.StartAt != nil && now.Before(*c.StartAt) {
		return ErrCouponNotStarted
	}
	if c.EndAt != nil && now.After(*c.EndAt) {
		return ErrCouponExpired
	}
	if subtotalAmount < c.MinOrderAmount {
		return ErrMinOrderNotMet
	}
	if c.UsageLimit != nil && c.UsedCount >= *c.UsageLimit {
		return ErrCouponUsageLimitReached
	}
	if userUsedCount >= c.UsageLimitPerUser {
		return ErrCouponAlreadyUsed
	}

	return nil
}

func (c *Coupon) CalculateDiscount(subtotalAmount float64) float64 {
	var discount float64
	switch c.DiscountType {
	case DiscountTypePercentage:
		discount = subtotalAmount * c.DiscountValue / 100
	case DiscountTypeFixedAmount:
		discount = c.DiscountValue
	}

	if c.MaxDiscountAmount != nil && discount > *c.MaxDiscountAmount {
		discount = *c.MaxDiscountAmount
	}
	if discount > subtotalAmount {
		discount = subtotalAmount
	}
	if discount < 0 {
		return 0
	}

	return discount
}
