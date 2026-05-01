package database

import (
	appAddress "kafka-order-demo/backend/internal/application/address"
	"kafka-order-demo/backend/internal/domain/address"
	"time"

	"gorm.io/gorm"
)

type GormAddress struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint `gorm:"not null;index"`
	CustomerName string
	Phone        string
	Address      string
	Email        string
	IsDefault    bool `gorm:"default:false"`
	CreatedAt    int64
	UpdatedAt    int64
}

func (GormAddress) TableName() string {
	return "addresses"
}

type GormAddressRepository struct {
	db *gorm.DB
}

func NewGormAddressRepository(db *gorm.DB) *GormAddressRepository {
	return &GormAddressRepository{db: db}
}

var _ appAddress.AddressRepository = (*GormAddressRepository)(nil)

func (r *GormAddressRepository) Create(address *address.Address) error {
	gormAddress := r.toGormAddress(address)
	if err := r.db.Create(gormAddress).Error; err != nil {
		return err
	}

	address.ID = gormAddress.ID
	return nil
}

func (r *GormAddressRepository) GetByUserID(userID uint) ([]*address.Address, error) {
	var gormAddresses []GormAddress
	if err := r.db.Where("user_id = ?", userID).Order("is_default DESC, id DESC").Find(&gormAddresses).Error; err != nil {
		return nil, err
	}

	addresses := make([]*address.Address, len(gormAddresses))
	for i, gormAddress := range gormAddresses {
		addresses[i] = r.toDomainAddress(&gormAddress)
	}

	return addresses, nil
}

func (r *GormAddressRepository) GetByIDAndUserID(id, userID uint) (*address.Address, error) {
	var gormAddress GormAddress
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&gormAddress).Error; err != nil {
		return nil, err
	}

	return r.toDomainAddress(&gormAddress), nil
}

func (r *GormAddressRepository) CountByUserID(userID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&GormAddress{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (r *GormAddressRepository) Update(address *address.Address) error {
	return r.db.Save(r.toGormAddress(address)).Error
}

func (r *GormAddressRepository) Delete(id, userID uint) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&GormAddress{}).Error
}

func (r *GormAddressRepository) SetDefault(id, userID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().Unix()
		if err := tx.Model(&GormAddress{}).
			Where("user_id = ?", userID).
			Updates(map[string]interface{}{
				"is_default": false,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		return tx.Model(&GormAddress{}).
			Where("id = ? AND user_id = ?", id, userID).
			Updates(map[string]interface{}{
				"is_default": true,
				"updated_at": now,
			}).Error
	})
}

func (r *GormAddressRepository) toGormAddress(address *address.Address) *GormAddress {
	return &GormAddress{
		ID:           address.ID,
		UserID:       address.UserID,
		CustomerName: address.CustomerName,
		Phone:        address.Phone,
		Address:      address.Address,
		Email:        address.Email,
		IsDefault:    address.IsDefault,
		CreatedAt:    address.CreatedAt.Unix(),
		UpdatedAt:    address.UpdatedAt.Unix(),
	}
}

func (r *GormAddressRepository) toDomainAddress(gormAddress *GormAddress) *address.Address {
	return &address.Address{
		ID:           gormAddress.ID,
		UserID:       gormAddress.UserID,
		CustomerName: gormAddress.CustomerName,
		Phone:        gormAddress.Phone,
		Address:      gormAddress.Address,
		Email:        gormAddress.Email,
		IsDefault:    gormAddress.IsDefault,
		CreatedAt:    timeFromUnix(gormAddress.CreatedAt),
		UpdatedAt:    timeFromUnix(gormAddress.UpdatedAt),
	}
}
