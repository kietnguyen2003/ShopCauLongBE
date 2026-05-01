package http

import (
	appAddress "kafka-order-demo/backend/internal/application/address"
	"time"
)

type addressRequest struct {
	CustomerName string `json:"customer_name"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	Email        string `json:"email"`
	IsDefault    bool   `json:"is_default"`
}

type addressUpdateRequest struct {
	CustomerName *string `json:"customer_name"`
	Phone        *string `json:"phone"`
	Address      *string `json:"address"`
	Email        *string `json:"email"`
}

type addressResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	CustomerName string    `json:"customer_name"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	Email        string    `json:"email"`
	IsDefault    bool      `json:"is_default"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func toAddressInput(req addressRequest, userID uint) appAddress.AddressRequest {
	return appAddress.AddressRequest{
		UserID:       userID,
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		Address:      req.Address,
		Email:        req.Email,
		IsDefault:    req.IsDefault,
	}
}

func toAddressUpdateInput(req addressUpdateRequest) appAddress.AddressUpdateRequest {
	return appAddress.AddressUpdateRequest{
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		Address:      req.Address,
		Email:        req.Email,
	}
}

func toAddressHTTPResponse(resp appAddress.AddressResponse) addressResponse {
	return addressResponse{
		ID:           resp.ID,
		UserID:       resp.UserID,
		CustomerName: resp.CustomerName,
		Phone:        resp.Phone,
		Address:      resp.Address,
		Email:        resp.Email,
		IsDefault:    resp.IsDefault,
		CreatedAt:    resp.CreatedAt,
		UpdatedAt:    resp.UpdatedAt,
	}
}

func toAddressHTTPResponses(responses []appAddress.AddressResponse) []addressResponse {
	result := make([]addressResponse, len(responses))
	for i, resp := range responses {
		result[i] = toAddressHTTPResponse(resp)
	}
	return result
}
