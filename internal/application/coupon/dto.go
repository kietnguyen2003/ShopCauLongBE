package coupon

import (
	domainCoupon "kafka-order-demo/backend/internal/domain/coupon"
	"math"
	"time"
)

type CouponRequest struct {
	Code              string
	Name              string
	Description       string
	DiscountType      string
	DiscountValue     float64
	MinOrderAmount    float64
	MaxDiscountAmount *float64
	UsageLimit        *int
	UsageLimitPerUser int
	StartAt           *time.Time
	EndAt             *time.Time
	IsActive          *bool
}

type CouponUpdateRequest struct {
	Code              *string
	Name              *string
	Description       *string
	DiscountType      *string
	DiscountValue     *float64
	MinOrderAmount    *float64
	MaxDiscountAmount *float64
	UsageLimit        *int
	UsageLimitPerUser *int
	StartAt           *time.Time
	EndAt             *time.Time
	IsActive          *bool
}

type CouponStatusRequest struct {
	IsActive bool
}

type CouponQuery struct {
	Page         int
	Limit        int
	Search       string
	IsActive     *bool
	DiscountType string
}

type ValidateCouponRequest struct {
	UserID uint
	Code   string
	Codes  []string
}

type CouponResponse struct {
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

type CouponListResponse struct {
	Items      []CouponResponse
	Page       int
	Limit      int
	Total      int64
	TotalPages int
}

type ValidateCouponResponse struct {
	CouponCode     string
	CouponCodes    []string
	DiscountType   string
	DiscountValue  float64
	SubtotalAmount float64
	DiscountAmount float64
	TotalAmount    float64
}

func toCouponResponse(coupon *domainCoupon.Coupon) CouponResponse {
	return CouponResponse{
		ID:                coupon.ID,
		Code:              coupon.Code,
		Name:              coupon.Name,
		Description:       coupon.Description,
		DiscountType:      coupon.DiscountType,
		DiscountValue:     coupon.DiscountValue,
		MinOrderAmount:    coupon.MinOrderAmount,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		UsageLimit:        coupon.UsageLimit,
		UsedCount:         coupon.UsedCount,
		UsageLimitPerUser: coupon.UsageLimitPerUser,
		StartAt:           coupon.StartAt,
		EndAt:             coupon.EndAt,
		IsActive:          coupon.IsActive,
		CreatedAt:         coupon.CreatedAt,
		UpdatedAt:         coupon.UpdatedAt,
	}
}

func toCouponResponses(coupons []*domainCoupon.Coupon) []CouponResponse {
	responses := make([]CouponResponse, len(coupons))
	for i, coupon := range coupons {
		responses[i] = toCouponResponse(coupon)
	}
	return responses
}

func totalPages(total int64, limit int) int {
	if total == 0 || limit <= 0 {
		return 0
	}

	return int(math.Ceil(float64(total) / float64(limit)))
}
