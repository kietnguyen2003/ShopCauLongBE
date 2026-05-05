package order

import (
	"errors"
	"fmt"
	domainCoupon "kafka-order-demo/backend/internal/domain/coupon"
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	domainProduct "kafka-order-demo/backend/internal/domain/product"
	"log"
	"strings"
	"time"
)

type Service struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
	addressRepo AddressRepository
	cartRepo    CartRepository
	couponRepo  CouponRepository
	notifier    NotificationPublisher
}

func NewService(orderRepo OrderRepository, productRepo ProductRepository, addressRepo AddressRepository, cartRepo CartRepository, couponRepo CouponRepository, notifier NotificationPublisher) *Service {
	return &Service{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		addressRepo: addressRepo,
		cartRepo:    cartRepo,
		couponRepo:  couponRepo,
		notifier:    notifier,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest) (*OrderResponse, error) {
	if req.AddressID != 0 {
		address, err := s.addressRepo.GetByIDAndUserID(req.AddressID, req.UserID)
		if err != nil {
			return nil, errors.New("address not found")
		}

		req.CustomerName = address.CustomerName
		req.Phone = address.Phone
		req.Address = address.Address
		req.Email = address.Email

		cartItems, err := s.cartRepo.GetByUserID(req.UserID)
		if err != nil {
			return nil, err
		}
		req.Items = make([]CreateOrderItemRequest, len(cartItems))
		for i, item := range cartItems {
			req.Items[i] = CreateOrderItemRequest{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}
	}

	// Create order
	ord, err := domainOrder.NewOrder(req.UserID, req.CustomerName, req.Phone, req.Address, req.Email)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to create order")
	}

	productsByID := make(map[uint]*domainProduct.Product)
	quantitiesByProductID := make(map[uint]int)

	// Add items and validate stock
	for _, item := range req.Items {
		prod, exists := productsByID[item.ProductID]
		if !exists {
			var err error
			prod, err = s.productRepo.GetByID(item.ProductID)
			if err != nil {
				return nil, errors.New("product not found")
			}
			productsByID[item.ProductID] = prod
		}

		snapshot, err := domainOrder.NewProductSnapshot(prod.Name, prod.Price, prod.Image, prod.Category, prod.Description)
		if err != nil {
			return nil, err
		}

		err = ord.AddItem(prod.ID, snapshot, item.Quantity)
		if err != nil {
			return nil, err
		}

		quantitiesByProductID[prod.ID] += item.Quantity
		if prod.Stock < quantitiesByProductID[prod.ID] {
			return nil, fmt.Errorf("sản phẩm %s không còn đủ hàng", prod.Name)
		}
	}

	err = ord.ValidateForCreation()
	if err != nil {
		return nil, err
	}

	var redemptions []*domainCoupon.Redemption
	couponCodes := normalizeRequestedCouponCodes(req.CouponCode, req.CouponCodes)
	if len(couponCodes) > 0 {
		coupons, discountAmount, err := s.validateCouponsForOrder(req.UserID, couponCodes, ord.SubtotalAmount)
		if err != nil {
			return nil, err
		}

		appliedCodes := make([]string, len(coupons))
		appliedCodes = appliedCodes[:0]
		now := time.Now()
		remainingDiscount := discountAmount
		for _, coupon := range coupons {
			if remainingDiscount <= 0 {
				break
			}
			couponDiscount := coupon.CalculateDiscount(ord.SubtotalAmount)
			if couponDiscount > remainingDiscount {
				couponDiscount = remainingDiscount
			}
			if couponDiscount <= 0 {
				continue
			}
			remainingDiscount -= couponDiscount
			appliedCodes = append(appliedCodes, coupon.Code)
			redemptions = append(redemptions, &domainCoupon.Redemption{
				CouponID:       coupon.ID,
				UserID:         req.UserID,
				DiscountAmount: couponDiscount,
				CreatedAt:      now,
			})
		}

		if len(redemptions) > 0 {
			ord.ApplyDiscount(&redemptions[0].CouponID, appliedCodes, discountAmount)
		}
	}

	productsToUpdate := make([]*domainProduct.Product, 0, len(productsByID))
	for productID, prod := range productsByID {
		if err := prod.DecreaseStock(quantitiesByProductID[productID]); err != nil {
			return nil, err
		}
		productsToUpdate = append(productsToUpdate, prod)
	}

	if len(redemptions) > 0 {
		err = s.orderRepo.CreateWithProductStockUpdatesAndCoupons(ord, productsToUpdate, redemptions)
	} else {
		err = s.orderRepo.CreateWithProductStockUpdates(ord, productsToUpdate)
	}
	if err != nil {
		return nil, err
	}

	response := toOrderResponse(ord)
	return &response, nil
}

func (s *Service) validateCouponsForOrder(userID uint, codes []string, subtotalAmount int64) ([]*domainCoupon.Coupon, int64, error) {
	coupons := make([]*domainCoupon.Coupon, 0, len(codes))
	seenCodes := make(map[string]struct{})
	var totalDiscount int64

	for _, code := range codes {
		normalizedCode := domainCoupon.NormalizeCode(code)
		if normalizedCode == "" {
			continue
		}
		if _, exists := seenCodes[normalizedCode]; exists {
			return nil, 0, errors.New("duplicate coupon code")
		}
		seenCodes[normalizedCode] = struct{}{}

		coupon, discountAmount, err := s.validateCouponForOrder(userID, normalizedCode, subtotalAmount)
		if err != nil {
			return nil, 0, err
		}
		coupons = append(coupons, coupon)
		totalDiscount += discountAmount
	}

	if len(coupons) == 0 {
		return nil, 0, domainCoupon.ErrCouponCodeRequired
	}
	if totalDiscount > subtotalAmount {
		totalDiscount = subtotalAmount
	}

	return coupons, totalDiscount, nil
}

func (s *Service) validateCouponForOrder(userID uint, code string, subtotalAmount int64) (*domainCoupon.Coupon, int64, error) {
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

func normalizeRequestedCouponCodes(legacyCode string, codes []string) []string {
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

func (s *Service) GetOrdersByUser(userID uint) ([]OrderResponse, error) {
	orders, err := s.orderRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return toOrderResponses(orders), nil
}

func (s *Service) GetAllOrders() ([]OrderResponse, error) {
	orders, err := s.orderRepo.GetAll()
	if err != nil {
		return nil, err
	}

	return toOrderResponses(orders), nil
}

func (s *Service) GetOrder(id uint) (*OrderResponse, error) {
	ord, err := s.orderRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	response := toOrderResponse(ord)
	return &response, nil
}

func (s *Service) UpdateOrderStatus(id uint, status domainOrder.OrderStatus) error {
	ord, err := s.orderRepo.GetByID(id)
	if err != nil {
		return err
	}

	wasCancelled := ord.Status == domainOrder.OrderStatusCancelled
	err = ord.UpdateStatus(status)
	if err != nil {
		return err
	}

	if status == domainOrder.OrderStatusCancelled && !wasCancelled {
		err = s.orderRepo.UpdateWithProductRestock(ord)
		if err != nil {
			return err
		}

		s.notifyOrderStatusChanged(ord)
		return nil
	}

	err = s.orderRepo.Update(ord)
	if err != nil {
		return err
	}

	s.notifyOrderStatusChanged(ord)
	return nil
}

func (s *Service) ClearAllOrders() error {
	return s.orderRepo.DeleteAll()
}

func (s *Service) notifyOrderStatusChanged(ord *domainOrder.Order) {
	if s.notifier == nil {
		return
	}
	if _, err := s.notifier.CreateOrderStatusNotification(ord.UserID, ord.ID, string(ord.Status)); err != nil {
		log.Println("failed to create order status notification:", err)
	}
}
