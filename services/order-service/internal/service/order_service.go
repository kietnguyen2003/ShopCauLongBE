package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"order-service/internal/models"
	"order-service/internal/repository"
)

type OrderService struct {
	orderRepo      repository.OrderRepository
	productService string // Product service URL
}

type CreateOrderRequest struct {
	UserID       uint                   `json:"user_id" binding:"required"`
	CustomerName string                 `json:"customer_name" binding:"required"`
	Phone        string                 `json:"phone" binding:"required"`
	Address      string                 `json:"address" binding:"required"`
	Email        string                 `json:"email" binding:"required,email"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}

type Product struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
}

func NewOrderService(orderRepo repository.OrderRepository, productServiceURL string) *OrderService {
	return &OrderService{
		orderRepo:      orderRepo,
		productService: productServiceURL,
	}
}

func (s *OrderService) CreateOrder(req CreateOrderRequest) (*models.Order, error) {
	order, err := models.NewOrder(req.UserID, req.CustomerName, req.Phone, req.Address, req.Email)
	if err != nil {
		return nil, err
	}

	err = s.orderRepo.Create(order)
	if err != nil {
		return nil, err
	}

	for _, item := range req.Items {
		product, err := s.getProduct(item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product with ID %d not found", item.ProductID)
		}

		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s", product.Name)
		}

		err = order.AddItem(
			product.ID,
			product.Name,
			product.Price,
			item.Quantity,
			product.Image,
			product.Category,
			product.Description,
		)
		if err != nil {
			return nil, err
		}

		err = s.decreaseProductStock(product.ID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %v", err)
		}
	}

	err = s.orderRepo.Update(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) GetOrders() ([]*models.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *OrderService) GetOrdersByUserID(userID uint) ([]*models.Order, error) {
	return s.orderRepo.GetByUserID(userID)
}

func (s *OrderService) GetOrderByID(id uint) (*models.Order, error) {
	return s.orderRepo.GetByID(id)
}

func (s *OrderService) UpdateOrderStatus(id uint, status models.OrderStatus) error {
	order, err := s.orderRepo.GetByID(id)
	if err != nil {
		return errors.New("order not found")
	}

	err = order.UpdateStatus(status)
	if err != nil {
		return err
	}

	return s.orderRepo.Update(order)
}

func (s *OrderService) getProduct(productID uint) (*Product, error) {
	url := fmt.Sprintf("%s/api/products/%d", s.productService, productID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("product not found")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var product Product
	err = json.Unmarshal(body, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (s *OrderService) decreaseProductStock(productID uint, quantity int) error {
	url := fmt.Sprintf("%s/api/products/%d/decrease-stock", s.productService, productID)
	
	payload := map[string]int{"quantity": quantity}
	payloadBytes, _ := json.Marshal(payload)
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to decrease product stock")
	}

	return nil
}