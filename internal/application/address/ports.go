package address

import domainAddress "kafka-order-demo/backend/internal/domain/address"

type AddressRepository interface {
	Create(address *domainAddress.Address) error
	GetByUserID(userID uint) ([]*domainAddress.Address, error)
	GetByIDAndUserID(id, userID uint) (*domainAddress.Address, error)
	CountByUserID(userID uint) (int64, error)
	Update(address *domainAddress.Address) error
	Delete(id, userID uint) error
	SetDefault(id, userID uint) error
}
