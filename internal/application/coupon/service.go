package coupon

import (
	"errors"
	domainCoupon "kafka-order-demo/backend/internal/domain/coupon"
	"strings"
	"time"
)

var ErrCartIsEmpty = errors.New("cart is empty")

type Service struct {
	couponRepo  CouponRepository
	cartRepo    CartRepository
	productRepo ProductRepository
}

func NewService(couponRepo CouponRepository, cartRepo CartRepository, productRepo ProductRepository) *Service {
	return &Service{
		couponRepo:  couponRepo,
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *Service) CreateCoupon(req CouponRequest) (*CouponResponse, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	coupon, err := domainCoupon.NewCoupon(req.Code, req.Name, req.Description, req.DiscountType, req.DiscountValue, req.MinOrderAmount, req.MaxDiscountAmount, req.UsageLimit, req.UsageLimitPerUser, req.StartAt, req.EndAt, isActive)
	if err != nil {
		return nil, err
	}

	if err := s.couponRepo.Create(coupon); err != nil {
		return nil, err
	}

	response := toCouponResponse(coupon)
	return &response, nil
}

func (s *Service) GetCoupons(query CouponQuery) (*CouponListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	coupons, total, err := s.couponRepo.GetWithQuery(query)
	if err != nil {
		return nil, err
	}

	return &CouponListResponse{
		Items:      toCouponResponses(coupons),
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages(total, query.Limit),
	}, nil
}

func (s *Service) GetAvailableCouponsForUser(userID uint, query CouponQuery) (*CouponListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.Limit <= 0 {
		query.Limit = 20
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	coupons, total, err := s.couponRepo.GetUnusedActiveWithQuery(query, userID)
	if err != nil {
		return nil, err
	}

	return &CouponListResponse{
		Items:      toCouponResponses(coupons),
		Page:       query.Page,
		Limit:      query.Limit,
		Total:      total,
		TotalPages: totalPages(total, query.Limit),
	}, nil
}

func (s *Service) GetCoupon(id uint) (*CouponResponse, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := toCouponResponse(coupon)
	return &response, nil
}

func (s *Service) GetAvailableCouponForUser(id, userID uint) (*CouponResponse, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if !coupon.IsActive {
		return nil, errors.New("coupon not found")
	}
	userUsedCount, err := s.couponRepo.CountRedemptionsByCouponIDAndUserID(coupon.ID, userID)
	if err != nil {
		return nil, err
	}
	if userUsedCount > 0 {
		return nil, errors.New("coupon not found")
	}

	response := toCouponResponse(coupon)
	return &response, nil
}

func (s *Service) UpdateCoupon(id uint, req CouponUpdateRequest) (*CouponResponse, error) {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if req.Code != nil {
		coupon.Code = domainCoupon.NormalizeCode(*req.Code)
	}
	if req.Name != nil {
		coupon.Name = *req.Name
	}
	if req.Description != nil {
		coupon.Description = *req.Description
	}
	if req.DiscountType != nil {
		coupon.DiscountType = *req.DiscountType
	}
	if req.DiscountValue != nil {
		coupon.DiscountValue = *req.DiscountValue
	}
	if req.MinOrderAmount != nil {
		coupon.MinOrderAmount = *req.MinOrderAmount
	}
	if req.MaxDiscountAmount != nil {
		coupon.MaxDiscountAmount = req.MaxDiscountAmount
	}
	if req.UsageLimit != nil {
		coupon.UsageLimit = req.UsageLimit
	}
	if req.UsageLimitPerUser != nil {
		coupon.UsageLimitPerUser = *req.UsageLimitPerUser
	}
	if req.StartAt != nil {
		coupon.StartAt = req.StartAt
	}
	if req.EndAt != nil {
		coupon.EndAt = req.EndAt
	}
	if req.IsActive != nil {
		coupon.IsActive = *req.IsActive
	}
	coupon.UpdatedAt = time.Now()

	if err := coupon.ValidateDefinition(); err != nil {
		return nil, err
	}
	if err := s.couponRepo.Update(coupon); err != nil {
		return nil, err
	}

	response := toCouponResponse(coupon)
	return &response, nil
}

func (s *Service) DeleteCoupon(id uint) error {
	if _, err := s.couponRepo.GetByID(id); err != nil {
		return err
	}

	return s.couponRepo.Delete(id)
}

func (s *Service) UpdateCouponStatus(id uint, isActive bool) error {
	coupon, err := s.couponRepo.GetByID(id)
	if err != nil {
		return err
	}

	coupon.IsActive = isActive
	coupon.UpdatedAt = time.Now()
	return s.couponRepo.Update(coupon)
}

func (s *Service) ValidateCoupon(req ValidateCouponRequest) (*ValidateCouponResponse, error) {
	subtotal, err := s.calculateCartSubtotal(req.UserID)
	if err != nil {
		return nil, err
	}
	return s.ValidateCouponsForAmount(req.UserID, normalizeCouponCodes(req.Code, req.Codes), subtotal)
}

func (s *Service) ValidateCouponForAmount(userID uint, code string, subtotalAmount int64) (*ValidateCouponResponse, error) {
	return s.ValidateCouponsForAmount(userID, []string{code}, subtotalAmount)
}

func (s *Service) ValidateCouponsForAmount(userID uint, codes []string, subtotalAmount int64) (*ValidateCouponResponse, error) {
	if len(codes) == 0 {
		return nil, domainCoupon.ErrCouponCodeRequired
	}

	var couponCodes []string
	var discountType string
	var discountValue int64
	var totalDiscount int64
	seenCodes := make(map[string]struct{})

	for _, code := range codes {
		coupon, discountAmount, err := s.validateSingleCouponForAmount(userID, code, subtotalAmount)
		if err != nil {
			return nil, err
		}
		if _, exists := seenCodes[coupon.Code]; exists {
			return nil, errors.New("duplicate coupon code")
		}
		seenCodes[coupon.Code] = struct{}{}

		couponCodes = append(couponCodes, coupon.Code)
		if len(couponCodes) == 1 {
			discountType = coupon.DiscountType
			discountValue = coupon.DiscountValue
		}
		totalDiscount += discountAmount
	}

	if totalDiscount > subtotalAmount {
		totalDiscount = subtotalAmount
	}

	return &ValidateCouponResponse{
		CouponCode:     strings.Join(couponCodes, ","),
		CouponCodes:    couponCodes,
		DiscountType:   discountType,
		DiscountValue:  discountValue,
		SubtotalAmount: subtotalAmount,
		DiscountAmount: totalDiscount,
		TotalAmount:    subtotalAmount - totalDiscount,
	}, nil
}

func (s *Service) validateSingleCouponForAmount(userID uint, code string, subtotalAmount int64) (*domainCoupon.Coupon, int64, error) {
	normalizedCode := domainCoupon.NormalizeCode(code)
	if normalizedCode == "" {
		return nil, 0, domainCoupon.ErrCouponCodeRequired
	}

	coupon, err := s.couponRepo.GetByCode(normalizedCode)
	if err != nil {
		return nil, 0, errors.New("coupon not found")
	}

	userUsedCount, err := s.couponRepo.CountRedemptionsByCouponIDAndUserID(coupon.ID, userID)
	if err != nil {
		return nil, 0, err
	}
	if err := coupon.ValidateForUse(subtotalAmount, userUsedCount, time.Now()); err != nil {
		return nil, 0, err
	}

	return coupon, coupon.CalculateDiscount(subtotalAmount), nil
}

func (s *Service) calculateCartSubtotal(userID uint) (int64, error) {
	items, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return 0, err
	}
	if len(items) == 0 {
		return 0, ErrCartIsEmpty
	}

	var subtotal int64
	for _, item := range items {
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return 0, err
		}
		subtotal += prod.Price * int64(item.Quantity)
	}

	return subtotal, nil
}

func normalizeCouponCodes(legacyCode string, codes []string) []string {
	result := make([]string, 0, len(codes)+1)
	if legacyCode != "" {
		result = append(result, legacyCode)
	}
	for _, code := range codes {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
