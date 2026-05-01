package address

import (
	domainAddress "kafka-order-demo/backend/internal/domain/address"
	"time"
)

type AddressRequest struct {
	UserID       uint
	CustomerName string
	Phone        string
	Address      string
	Email        string
	IsDefault    bool
}

type AddressUpdateRequest struct {
	CustomerName *string
	Phone        *string
	Address      *string
	Email        *string
}

type AddressResponse struct {
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

func toAddressResponse(address *domainAddress.Address) AddressResponse {
	return AddressResponse{
		ID:           address.ID,
		UserID:       address.UserID,
		CustomerName: address.CustomerName,
		Phone:        address.Phone,
		Address:      address.Address,
		Email:        address.Email,
		IsDefault:    address.IsDefault,
		CreatedAt:    address.CreatedAt,
		UpdatedAt:    address.UpdatedAt,
	}
}

func toAddressResponses(addresses []*domainAddress.Address) []AddressResponse {
	responses := make([]AddressResponse, len(addresses))
	for i, address := range addresses {
		responses[i] = toAddressResponse(address)
	}
	return responses
}
