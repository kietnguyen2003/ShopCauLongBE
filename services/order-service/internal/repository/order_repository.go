package repository

import (
	"order-service/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *models.Order) error
	GetByID(id uint) (*models.Order, error)
	GetByUserID(userID uint) ([]*models.Order, error)
	GetAll() ([]*models.Order, error)
	Update(order *models.Order) error
	Delete(id uint) error
	DeleteAll() error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("OrderItems").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) GetByUserID(userID uint) ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.Preload("OrderItems").Where("user_id = ?", userID).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetAll() ([]*models.Order, error) {
	var orders []*models.Order
	err := r.db.Preload("OrderItems").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&models.Order{}, id).Error
}

func (r *orderRepository) DeleteAll() error {
	return r.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Order{}).Error
}