package order

import (
	"time"
	"errors"
)

// OrderStatus represents the order status
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// OrderItem represents an item in an order
type OrderItem struct {
	ID          uint
	OrderID     uint
	ProductID   uint
	Name        string
	Price       float64
	Quantity    int
	Image       string
	Category    string
	Description string
	CreatedAt   time.Time
}

// Order represents the order domain entity
type Order struct {
	ID           uint
	UserID       uint
	OrderItems   []OrderItem
	TotalAmount  float64
	Status       OrderStatus
	CustomerName string
	Phone        string
	Address      string
	Email        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// OrderRepository defines the interface for order data access
type OrderRepository interface {
	Create(order *Order) error
	GetByID(id uint) (*Order, error)
	GetByUserID(userID uint) ([]*Order, error)
	GetAll() ([]*Order, error)
	Update(order *Order) error
	Delete(id uint) error
	DeleteAll() error
}

// NewOrder creates a new order with validation
func NewOrder(userID uint, customerName, phone, address, email string) (*Order, error) {
	if userID == 0 {
		return nil, errors.New("user ID cannot be zero")
	}
	if customerName == "" {
		return nil, errors.New("customer name cannot be empty")
	}
	if phone == "" {
		return nil, errors.New("phone cannot be empty")
	}
	if address == "" {
		return nil, errors.New("address cannot be empty")
	}
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	return &Order{
		UserID:       userID,
		CustomerName: customerName,
		Phone:        phone,
		Address:      address,
		Email:        email,
		Status:       OrderStatusPending,
		TotalAmount:  0,
		OrderItems:   make([]OrderItem, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// AddItem adds an item to the order
func (o *Order) AddItem(productID uint, name string, price float64, quantity int, image, category, description string) error {
	if quantity <= 0 {
		return errors.New("quantity must be greater than 0")
	}
	if price <= 0 {
		return errors.New("price must be greater than 0")
	}

	item := OrderItem{
		OrderID:     o.ID,
		ProductID:   productID,
		Name:        name,
		Price:       price,
		Quantity:    quantity,
		Image:       image,
		Category:    category,
		Description: description,
		CreatedAt:   time.Now(),
	}

	o.OrderItems = append(o.OrderItems, item)
	o.calculateTotal()
	o.UpdatedAt = time.Now()

	return nil
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(status OrderStatus) error {
	if status != OrderStatusPending && status != OrderStatusConfirmed && status != OrderStatusCancelled {
		return errors.New("invalid order status")
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// calculateTotal calculates the total amount of the order
func (o *Order) calculateTotal() {
	total := 0.0
	for _, item := range o.OrderItems {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
}

// CanBeModified checks if order can be modified
func (o *Order) CanBeModified() bool {
	return o.Status == OrderStatusPending
}

// GetItemCount returns the total number of items in the order
func (o *Order) GetItemCount() int {
	count := 0
	for _, item := range o.OrderItems {
		count += item.Quantity
	}
	return count
}