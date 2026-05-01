package cart

import (
	"errors"
	"time"
)

var (
	ErrInvalidUserID    = errors.New("user ID cannot be zero")
	ErrInvalidProductID = errors.New("product ID cannot be zero")
	ErrInvalidQuantity  = errors.New("quantity must be greater than 0")
)

type CartItem struct {
	ID        uint
	UserID    uint
	ProductID uint
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewCartItem(userID, productID uint, quantity int) (*CartItem, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if productID == 0 {
		return nil, ErrInvalidProductID
	}
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	now := time.Now()
	return &CartItem{
		UserID:    userID,
		ProductID: productID,
		Quantity:  quantity,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (i *CartItem) UpdateQuantity(quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	i.Quantity = quantity
	i.UpdatedAt = time.Now()
	return nil
}
