package order

import (
	"kafka-order-demo/backend/internal/domain/order"
	"kafka-order-demo/backend/internal/domain/product"
)

type Service struct {
	orderRepo   order.OrderRepository
	productRepo product.ProductRepository
}

type CreateOrderRequest struct {
	UserID       uint                     `json:"user_id"`
	CustomerName string                   `json:"customer_name"`
	Phone        string                   `json:"phone"`
	Address      string                   `json:"address"`
	Email        string                   `json:"email"`
	Items        []CreateOrderItemRequest `json:"items"`
}

type CreateOrderItemRequest struct {
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`
}

func NewService(orderRepo order.OrderRepository, productRepo product.ProductRepository) *Service {
	return &Service{
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *Service) CreateOrder(req CreateOrderRequest) (*order.Order, error) {
	// Create order
	ord, err := order.NewOrder(req.UserID, req.CustomerName, req.Phone, req.Address, req.Email)
	if err != nil {
		return nil, err
	}

	// Add items and validate stock
	for _, item := range req.Items {
		prod, err := s.productRepo.GetByID(item.ProductID)
		if err != nil {
			return nil, err
		}

		if !prod.IsAvailable(item.Quantity) {
			return nil, err
		}

		err = ord.AddItem(prod.ID, prod.Name, prod.Price, item.Quantity, prod.Image, prod.Category, prod.Description)
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

	// Save order
	err = s.orderRepo.Create(ord)
	if err != nil {
		return nil, err
	}

	return ord, nil
}

func (s *Service) GetOrdersByUser(userID uint) ([]*order.Order, error) {
	return s.orderRepo.GetByUserID(userID)
}

func (s *Service) GetAllOrders() ([]*order.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *Service) GetOrder(id uint) (*order.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *Service) UpdateOrderStatus(id uint, status order.OrderStatus) error {
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