package order

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidOrderStatus    = errors.New("invalid order status")
	ErrOrderCannotBeModified = errors.New("order cannot be modified")
	ErrOrderMustHaveItems    = errors.New("order must contain at least one item")
	ErrInvalidProductID      = errors.New("product ID cannot be zero")
	ErrInvalidQuantity       = errors.New("quantity must be greater than 0")
	ErrInvalidPrice          = errors.New("price must be greater than 0")
	ErrProductNameRequired   = errors.New("product name cannot be empty")
)

// OrderStatus represents the order status
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// ProductSnapshot captures product data at purchase time.
type ProductSnapshot struct {
	Name        string
	Price       int64
	Image       string
	Category    string
	Description string
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID              uint
	OrderID         uint
	ProductID       uint
	ProductSnapshot ProductSnapshot
	Quantity        int
	CreatedAt       time.Time
}

// Order represents the order domain entity
type Order struct {
	ID             uint
	UserID         uint
	OrderItems     []OrderItem
	SubtotalAmount int64
	DiscountAmount int64
	TotalAmount    int64
	CouponID       *uint
	CouponCode     string
	CouponCodes    []string
	Status         OrderStatus
	CustomerName   string
	Phone          string
	Address        string
	Email          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
		UserID:         userID,
		CustomerName:   customerName,
		Phone:          phone,
		Address:        address,
		Email:          email,
		Status:         OrderStatusPending,
		SubtotalAmount: 0,
		DiscountAmount: 0,
		TotalAmount:    0,
		OrderItems:     make([]OrderItem, 0),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func NewProductSnapshot(name string, price int64, image, category, description string) (ProductSnapshot, error) {
	if name == "" {
		return ProductSnapshot{}, ErrProductNameRequired
	}
	if price <= 0 {
		return ProductSnapshot{}, ErrInvalidPrice
	}

	return ProductSnapshot{
		Name:        name,
		Price:       price,
		Image:       image,
		Category:    category,
		Description: description,
	}, nil
}

// AddItem adds an item to the order
func (o *Order) AddItem(productID uint, snapshot ProductSnapshot, quantity int) error {
	if !o.CanBeModified() {
		return ErrOrderCannotBeModified
	}
	if productID == 0 {
		return ErrInvalidProductID
	}
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	item := OrderItem{
		OrderID:         o.ID,
		ProductID:       productID,
		ProductSnapshot: snapshot,
		Quantity:        quantity,
		CreatedAt:       time.Now(),
	}

	o.OrderItems = append(o.OrderItems, item)
	o.calculateTotal()
	o.UpdatedAt = time.Now()

	return nil
}

// UpdateStatus updates the order status
func (o *Order) UpdateStatus(status OrderStatus) error {
	if !isValidStatus(status) {
		return ErrInvalidOrderStatus
	}

	if status == o.Status {
		return nil
	}

	switch o.Status {
	case OrderStatusPending:
		if status != OrderStatusConfirmed && status != OrderStatusCancelled {
			return ErrInvalidOrderStatus
		}
	case OrderStatusConfirmed, OrderStatusCancelled:
		return ErrInvalidOrderStatus
	default:
		return ErrInvalidOrderStatus
	}

	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// calculateTotal calculates the total amount of the order
func (o *Order) calculateTotal() {
	var total int64
	for _, item := range o.OrderItems {
		total += item.ProductSnapshot.Price * int64(item.Quantity)
	}
	o.SubtotalAmount = total
	o.TotalAmount = o.SubtotalAmount - o.DiscountAmount
}

func (o *Order) ApplyDiscount(couponID *uint, couponCodes []string, discountAmount int64) {
	if discountAmount < 0 {
		discountAmount = 0
	}
	if discountAmount > o.SubtotalAmount {
		discountAmount = o.SubtotalAmount
	}

	o.CouponID = couponID
	o.CouponCodes = normalizeCouponCodes(couponCodes)
	o.CouponCode = strings.Join(o.CouponCodes, ",")
	o.DiscountAmount = discountAmount
	o.TotalAmount = o.SubtotalAmount - o.DiscountAmount
	o.UpdatedAt = time.Now()
}

func normalizeCouponCodes(codes []string) []string {
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		trimmed := strings.TrimSpace(code)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
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

func (o *Order) ValidateForCreation() error {
	if len(o.OrderItems) == 0 {
		return ErrOrderMustHaveItems
	}
	if o.TotalAmount <= 0 {
		return ErrInvalidPrice
	}
	return nil
}

func isValidStatus(status OrderStatus) bool {
	return status == OrderStatusPending || status == OrderStatusConfirmed || status == OrderStatusCancelled
}
