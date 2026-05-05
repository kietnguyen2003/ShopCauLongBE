package http

import (
	appCoupon "kafka-order-demo/backend/internal/application/coupon"
	"time"
)

type couponRequest struct {
	Code              string     `json:"code"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	DiscountType      string     `json:"discount_type"`
	DiscountValue     int64      `json:"discount_value"`
	MinOrderAmount    int64      `json:"min_order_amount"`
	MaxDiscountAmount *int64     `json:"max_discount_amount"`
	UsageLimit        *int       `json:"usage_limit"`
	UsageLimitPerUser int        `json:"usage_limit_per_user"`
	StartAt           *time.Time `json:"start_at"`
	EndAt             *time.Time `json:"end_at"`
	IsActive          *bool      `json:"is_active"`
}

type couponUpdateRequest struct {
	Code              *string    `json:"code"`
	Name              *string    `json:"name"`
	Description       *string    `json:"description"`
	DiscountType      *string    `json:"discount_type"`
	DiscountValue     *int64     `json:"discount_value"`
	MinOrderAmount    *int64     `json:"min_order_amount"`
	MaxDiscountAmount *int64     `json:"max_discount_amount"`
	UsageLimit        *int       `json:"usage_limit"`
	UsageLimitPerUser *int       `json:"usage_limit_per_user"`
	StartAt           *time.Time `json:"start_at"`
	EndAt             *time.Time `json:"end_at"`
	IsActive          *bool      `json:"is_active"`
}

type couponStatusRequest struct {
	IsActive bool `json:"is_active"`
}

type validateCouponRequest struct {
	Code  string   `json:"code"`
	Codes []string `json:"coupon_codes"`
}

type couponResponse struct {
	ID                uint       `json:"id"`
	Code              string     `json:"code"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	DiscountType      string     `json:"discount_type"`
	DiscountValue     int64      `json:"discount_value"`
	MinOrderAmount    int64      `json:"min_order_amount"`
	MaxDiscountAmount *int64     `json:"max_discount_amount"`
	UsageLimit        *int       `json:"usage_limit"`
	UsedCount         int        `json:"used_count"`
	UsageLimitPerUser int        `json:"usage_limit_per_user"`
	StartAt           *time.Time `json:"start_at"`
	EndAt             *time.Time `json:"end_at"`
	IsActive          bool       `json:"is_active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type couponListResponse struct {
	Items      []couponResponse `json:"items"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

type validateCouponResponse struct {
	CouponCode     string   `json:"coupon_code"`
	CouponCodes    []string `json:"coupon_codes"`
	DiscountType   string   `json:"discount_type"`
	DiscountValue  int64    `json:"discount_value"`
	SubtotalAmount int64    `json:"subtotal_amount"`
	DiscountAmount int64    `json:"discount_amount"`
	TotalAmount    int64    `json:"total_amount"`
}

func toCouponInput(req couponRequest) appCoupon.CouponRequest {
	return appCoupon.CouponRequest{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		UsageLimitPerUser: req.UsageLimitPerUser,
		StartAt:           req.StartAt,
		EndAt:             req.EndAt,
		IsActive:          req.IsActive,
	}
}

func toCouponUpdateInput(req couponUpdateRequest) appCoupon.CouponUpdateRequest {
	return appCoupon.CouponUpdateRequest{
		Code:              req.Code,
		Name:              req.Name,
		Description:       req.Description,
		DiscountType:      req.DiscountType,
		DiscountValue:     req.DiscountValue,
		MinOrderAmount:    req.MinOrderAmount,
		MaxDiscountAmount: req.MaxDiscountAmount,
		UsageLimit:        req.UsageLimit,
		UsageLimitPerUser: req.UsageLimitPerUser,
		StartAt:           req.StartAt,
		EndAt:             req.EndAt,
		IsActive:          req.IsActive,
	}
}

func toCouponHTTPResponse(resp appCoupon.CouponResponse) couponResponse {
	return couponResponse{
		ID:                resp.ID,
		Code:              resp.Code,
		Name:              resp.Name,
		Description:       resp.Description,
		DiscountType:      resp.DiscountType,
		DiscountValue:     resp.DiscountValue,
		MinOrderAmount:    resp.MinOrderAmount,
		MaxDiscountAmount: resp.MaxDiscountAmount,
		UsageLimit:        resp.UsageLimit,
		UsedCount:         resp.UsedCount,
		UsageLimitPerUser: resp.UsageLimitPerUser,
		StartAt:           resp.StartAt,
		EndAt:             resp.EndAt,
		IsActive:          resp.IsActive,
		CreatedAt:         resp.CreatedAt,
		UpdatedAt:         resp.UpdatedAt,
	}
}

func toCouponHTTPResponses(responses []appCoupon.CouponResponse) []couponResponse {
	result := make([]couponResponse, len(responses))
	for i, resp := range responses {
		result[i] = toCouponHTTPResponse(resp)
	}
	return result
}

func toCouponListHTTPResponse(resp appCoupon.CouponListResponse) couponListResponse {
	return couponListResponse{
		Items:      toCouponHTTPResponses(resp.Items),
		Page:       resp.Page,
		Limit:      resp.Limit,
		Total:      resp.Total,
		TotalPages: resp.TotalPages,
	}
}

func toValidateCouponHTTPResponse(resp appCoupon.ValidateCouponResponse) validateCouponResponse {
	return validateCouponResponse{
		CouponCode:     resp.CouponCode,
		CouponCodes:    resp.CouponCodes,
		DiscountType:   resp.DiscountType,
		DiscountValue:  resp.DiscountValue,
		SubtotalAmount: resp.SubtotalAmount,
		DiscountAmount: resp.DiscountAmount,
		TotalAmount:    resp.TotalAmount,
	}
}
