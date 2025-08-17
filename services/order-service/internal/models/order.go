package models

import (
	"time"
	"errors"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID          uint        `json:"id" gorm:"primaryKey"`
	OrderID     uint        `json:"order_id"`
	ProductID   uint        `json:"product_id"`
	Name        string      `json:"name"`
	Price       float64     `json:"price"`
	Quantity    int         `json:"quantity"`
	Image       string      `json:"image"`
	Category    string      `json:"category"`
	Description string      `json:"description"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Order struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	UserID       uint          `json:"user_id"`
	OrderItems   []OrderItem   `json:"order_items" gorm:"foreignKey:OrderID"`
	TotalAmount  float64       `json:"total_amount"`
	Status       OrderStatus   `json:"status"`
	CustomerName string        `json:"customer_name"`
	Phone        string        `json:"phone"`
	Address      string        `json:"address"`
	Email        string        `json:"email"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

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

func (o *Order) UpdateStatus(status OrderStatus) error {
	if status != OrderStatusPending && status != OrderStatusConfirmed && status != OrderStatusCancelled {
		return errors.New("invalid order status")
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) calculateTotal() {
	total := 0.0
	for _, item := range o.OrderItems {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
}

func (o *Order) CanBeModified() bool {
	return o.Status == OrderStatusPending
}

func (o *Order) GetItemCount() int {
	count := 0
	for _, item := range o.OrderItems {
		count += item.Quantity
	}
	return count
}