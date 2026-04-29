package order

import (
	"errors"
	domainOrder "kafka-order-demo/backend/internal/domain/order"
	"log"
)

type Service struct {
	orderRepo   OrderRepository
	productRepo ProductRepository
}

func NewService(orderRepo OrderRepository, productRepo ProductRepository) *Service {
	return &Service{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest) (*OrderResponse, error) {
	// Create order
	ord, err := domainOrder.NewOrder(req.UserID, req.CustomerName, req.Phone, req.Address, req.Email)
	if err != nil {
		log.Println(err)
		return nil, errors.New("failed to create order")
	}

	// Add items and validate stock
	for _, item := range req.Items {
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, errors.New("product not found")
		}

		snapshot, err := domainOrder.NewProductSnapshot(prod.Name, prod.Price, prod.Image, prod.Category, prod.Description)
		if err != nil {
			return nil, err
		}

		err = ord.AddItem(prod.ID, snapshot, item.Quantity)
		if err != nil {
			return nil, err
		}

		// Decrease product stock
		err = prod.DecreaseStock(item.Quantity)
		if err != nil {
			return nil, err
		}

		err = s.productRepo.Update(prod)
		if err != nil {
			return nil, err
		}
	}

	err = ord.ValidateForCreation()
	if err != nil {
		return nil, err
	}

	// Save order
	err = s.orderRepo.Create(ord)
	if err != nil {
		return nil, err
	}

	response := toOrderResponse(ord)
	return &response, nil
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

	err = ord.UpdateStatus(status)
	if err != nil {
		return err
	}

	err = s.orderRepo.Update(ord)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) ClearAllOrders() error {
	return s.orderRepo.DeleteAll()
}
