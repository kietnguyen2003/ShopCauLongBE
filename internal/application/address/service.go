package address

import (
	"errors"

	domainAddress "kafka-order-demo/backend/internal/domain/address"
)

type Service struct {
	addressRepo AddressRepository
}

func NewService(addressRepo AddressRepository) *Service {
	return &Service{addressRepo: addressRepo}
}

func (s *Service) GetAddresses(userID uint) ([]AddressResponse, error) {
	addresses, err := s.addressRepo.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	return toAddressResponses(addresses), nil
}

func (s *Service) CreateAddress(req AddressRequest) (*AddressResponse, error) {
	count, err := s.addressRepo.CountByUserID(req.UserID)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		req.IsDefault = true
	}

	address, err := domainAddress.NewAddress(req.UserID, req.CustomerName, req.Phone, req.Address, req.Email, req.IsDefault)
	if err != nil {
		return nil, err
	}

	if err := s.addressRepo.Create(address); err != nil {
		return nil, err
	}
	if address.IsDefault {
		if err := s.addressRepo.SetDefault(address.ID, address.UserID); err != nil {
			return nil, err
		}
		address.MarkDefault()
	}

	response := toAddressResponse(address)
	return &response, nil
}

func (s *Service) UpdateAddress(id, userID uint, req AddressUpdateRequest) (*AddressResponse, error) {
	address, err := s.addressRepo.GetByIDAndUserID(id, userID)
	if err != nil {
		return nil, errors.New("address not found")
	}

	customerName := address.CustomerName
	phone := address.Phone
	addressText := address.Address
	email := address.Email

	if req.CustomerName != nil {
		customerName = *req.CustomerName
	}
	if req.Phone != nil {
		phone = *req.Phone
	}
	if req.Address != nil {
		addressText = *req.Address
	}
	if req.Email != nil {
		email = *req.Email
	}

	if err := address.Update(customerName, phone, addressText, email); err != nil {
		return nil, err
	}
	if err := s.addressRepo.Update(address); err != nil {
		return nil, err
	}

	response := toAddressResponse(address)
	return &response, nil
}

func (s *Service) DeleteAddress(id, userID uint) error {
	if _, err := s.addressRepo.GetByIDAndUserID(id, userID); err != nil {
		return errors.New("address not found")
	}

	return s.addressRepo.Delete(id, userID)
}

func (s *Service) SetDefaultAddress(id, userID uint) error {
	if _, err := s.addressRepo.GetByIDAndUserID(id, userID); err != nil {
		return errors.New("address not found")
	}

	return s.addressRepo.SetDefault(id, userID)
}
