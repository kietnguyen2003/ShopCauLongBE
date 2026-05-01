package address

import (
	"errors"
	"time"
)

var (
	ErrInvalidUserID     = errors.New("user ID cannot be zero")
	ErrCustomerNameEmpty = errors.New("customer name cannot be empty")
	ErrPhoneEmpty        = errors.New("phone cannot be empty")
	ErrAddressEmpty      = errors.New("address cannot be empty")
	ErrEmailEmpty        = errors.New("email cannot be empty")
)

type Address struct {
	ID           uint
	UserID       uint
	CustomerName string
	Phone        string
	Address      string
	Email        string
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewAddress(userID uint, customerName, phone, address, email string, isDefault bool) (*Address, error) {
	if userID == 0 {
		return nil, ErrInvalidUserID
	}
	if customerName == "" {
		return nil, ErrCustomerNameEmpty
	}
	if phone == "" {
		return nil, ErrPhoneEmpty
	}
	if address == "" {
		return nil, ErrAddressEmpty
	}
	if email == "" {
		return nil, ErrEmailEmpty
	}

	now := time.Now()
	return &Address{
		UserID:       userID,
		CustomerName: customerName,
		Phone:        phone,
		Address:      address,
		Email:        email,
		IsDefault:    isDefault,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (a *Address) Update(customerName, phone, address, email string) error {
	if customerName == "" {
		return ErrCustomerNameEmpty
	}
	if phone == "" {
		return ErrPhoneEmpty
	}
	if address == "" {
		return ErrAddressEmpty
	}
	if email == "" {
		return ErrEmailEmpty
	}

	a.CustomerName = customerName
	a.Phone = phone
	a.Address = address
	a.Email = email
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Address) MarkDefault() {
	a.IsDefault = true
	a.UpdatedAt = time.Now()
}
