package cart

import (
	"errors"
	"fmt"

	domainCart "kafka-order-demo/backend/internal/domain/cart"
)

var ErrCartItemNotFound = errors.New("cart item not found")

type Service struct {
	cartRepo    CartRepository
	productRepo ProductRepository
}

func NewService(cartRepo CartRepository, productRepo ProductRepository) *Service {
	return &Service{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *Service) GetCart(userID uint) (*CartResponse, error) {
	items, err := s.cartRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	responses := make([]CartItemResponse, 0, len(items))
	var totalAmount int64
	var totalItems int
	for _, item := range items {
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			continue
		}

		response := toCartItemResponse(item, prod)
		responses = append(responses, response)
		totalAmount += response.Subtotal
		totalItems += item.Quantity
	}

	return &CartResponse{
		Items:       responses,
		TotalAmount: totalAmount,
		TotalItems:  totalItems,
	}, nil
}

func (s *Service) AddItem(req CartItemRequest) (*CartResponse, error) {
	return s.AddItems([]CartItemRequest{req})
}

func (s *Service) AddItems(reqs []CartItemRequest) (*CartResponse, error) {
	if len(reqs) == 0 {
		return nil, errors.New("cart items cannot be empty")
	}

	itemsToSave := make([]*domainCart.CartItem, 0, len(reqs))
	for _, req := range reqs {
		item, err := s.prepareCartItem(req)
		if err != nil {
			return nil, err
		}
		itemsToSave = append(itemsToSave, item)
	}

	for _, item := range itemsToSave {
		if item.ID == 0 {
			if err := s.cartRepo.Create(item); err != nil {
				return nil, err
			}
			continue
		}

		if err := s.cartRepo.Update(item); err != nil {
			return nil, err
		}
	}

	return s.GetCart(reqs[0].UserID)
}

func (s *Service) prepareCartItem(req CartItemRequest) (*domainCart.CartItem, error) {
	prod, err := s.productRepo.GetByID(req.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	item, err := s.cartRepo.GetByUserIDAndProductID(req.UserID, req.ProductID)
	if err != nil {
		if !errors.Is(err, ErrCartItemNotFound) {
			return nil, err
		}

		item, err = domainCart.NewCartItem(req.UserID, req.ProductID, req.Quantity)
		if err != nil {
			return nil, err
		}
	} else if err := item.UpdateQuantity(item.Quantity + req.Quantity); err != nil {
		return nil, err
	}

	if item.Quantity > prod.Stock {
		return nil, fmt.Errorf("sản phẩm %s không còn đủ hàng", prod.Name)
	}

	return item, nil
}

func (s *Service) UpdateItemQuantity(userID, itemID uint, quantity int) (*CartResponse, error) {
	item, err := s.cartRepo.GetByIDAndUserID(itemID, userID)
	if err != nil {
		return nil, errors.New("cart item not found")
	}

	prod, err := s.productRepo.GetByID(item.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if quantity > prod.Stock {
		return nil, fmt.Errorf("sản phẩm %s không còn đủ hàng", prod.Name)
	}

	if err := item.UpdateQuantity(quantity); err != nil {
		return nil, err
	}
	if err := s.cartRepo.Update(item); err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

func (s *Service) DeleteItem(userID, itemID uint) (*CartResponse, error) {
	if _, err := s.cartRepo.GetByIDAndUserID(itemID, userID); err != nil {
		return nil, errors.New("cart item not found")
	}
	if err := s.cartRepo.Delete(itemID, userID); err != nil {
		return nil, err
	}

	return s.GetCart(userID)
}

func (s *Service) ClearCart(userID uint) error {
	return s.cartRepo.Clear(userID)
}
