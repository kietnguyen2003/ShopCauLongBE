package http

import (
	"errors"
	"net/http"
	"strconv"

	appCoupon "kafka-order-demo/backend/internal/application/coupon"

	"github.com/gin-gonic/gin"
)

const (
	defaultCouponPage  = 1
	defaultCouponLimit = 20
	maxCouponLimit     = 100
)

type CouponHandler struct {
	couponService *appCoupon.Service
}

func NewCouponHandler(couponService *appCoupon.Service) *CouponHandler {
	return &CouponHandler{
		couponService: couponService,
	}
}

func (h *CouponHandler) GetCoupons(c *gin.Context) {
	query, err := parseCouponQuery(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupons, err := h.couponService.GetCoupons(query)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get coupons successfully", toCouponListHTTPResponse(*coupons))
}

func (h *CouponHandler) GetActiveCoupons(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	query, err := parseCouponQuery(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupons, err := h.couponService.GetAvailableCouponsForUser(userID.(uint), query)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Get coupons successfully", toCouponListHTTPResponse(*coupons))
}

func (h *CouponHandler) CreateCoupon(c *gin.Context) {
	var req couponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupon, err := h.couponService.CreateCoupon(toCouponInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusCreated, "Create coupon successfully", toCouponHTTPResponse(*coupon))
}

func (h *CouponHandler) GetCoupon(c *gin.Context) {
	id, err := parseCouponID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid coupon ID")
		return
	}

	coupon, err := h.couponService.GetCoupon(id)
	if err != nil {
		errorResponse(c, http.StatusNotFound, "coupon not found")
		return
	}

	successResponse(c, http.StatusOK, "Get coupon successfully", toCouponHTTPResponse(*coupon))
}

func (h *CouponHandler) GetActiveCoupon(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	id, err := parseCouponID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid coupon ID")
		return
	}

	coupon, err := h.couponService.GetAvailableCouponForUser(id, userID.(uint))
	if err != nil {
		errorResponse(c, http.StatusNotFound, "coupon not found")
		return
	}

	successResponse(c, http.StatusOK, "Get coupon successfully", toCouponHTTPResponse(*coupon))
}

func (h *CouponHandler) UpdateCoupon(c *gin.Context) {
	id, err := parseCouponID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid coupon ID")
		return
	}

	var req couponUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	coupon, err := h.couponService.UpdateCoupon(id, toCouponUpdateInput(req))
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update coupon successfully", toCouponHTTPResponse(*coupon))
}

func (h *CouponHandler) DeleteCoupon(c *gin.Context) {
	id, err := parseCouponID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid coupon ID")
		return
	}

	if err := h.couponService.DeleteCoupon(id); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Delete coupon successfully", nil)
}

func (h *CouponHandler) UpdateCouponStatus(c *gin.Context) {
	id, err := parseCouponID(c)
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "Invalid coupon ID")
		return
	}

	var req couponStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.couponService.UpdateCouponStatus(id, req.IsActive); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Update coupon status successfully", nil)
}

func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		errorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req validateCouponRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.couponService.ValidateCoupon(appCoupon.ValidateCouponRequest{
		UserID: userID.(uint),
		Code:   req.Code,
		Codes:  req.Codes,
	})
	if err != nil {
		errorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	successResponse(c, http.StatusOK, "Coupon is valid", toValidateCouponHTTPResponse(*result))
}

func parseCouponQuery(c *gin.Context) (appCoupon.CouponQuery, error) {
	page, err := parsePositiveIntQuery(c, "page", defaultCouponPage)
	if err != nil {
		return appCoupon.CouponQuery{}, err
	}

	limit, err := parsePositiveIntQuery(c, "limit", defaultCouponLimit)
	if err != nil {
		return appCoupon.CouponQuery{}, err
	}
	if limit > maxCouponLimit {
		limit = maxCouponLimit
	}

	isActive, err := parseOptionalBoolQuery(c, "is_active")
	if err != nil {
		return appCoupon.CouponQuery{}, err
	}

	return appCoupon.CouponQuery{
		Page:         page,
		Limit:        limit,
		Search:       c.Query("search"),
		IsActive:     isActive,
		DiscountType: c.Query("discount_type"),
	}, nil
}

func parseOptionalBoolQuery(c *gin.Context, key string) (*bool, error) {
	value := c.Query(key)
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, errors.New(key + " must be a boolean")
	}

	return &parsed, nil
}

func parseCouponID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint(id), nil
}
