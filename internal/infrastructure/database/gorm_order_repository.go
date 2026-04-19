package database

import (
	appOrder "kafka-order-demo/backend/internal/application/order"
	"kafka-order-demo/backend/internal/domain/order"

	"gorm.io/gorm"
)

type GormOrder struct {
	ID           uint    `gorm:"primaryKey"`
	UserID       uint    `gorm:"not null"`
	TotalAmount  float64 `gorm:"not null"`
	Status       string  `gorm:"default:pending"`
	CustomerName string
	Phone        string
	Address      string
	Email        string
	CreatedAt    int64
	UpdatedAt    int64
	OrderItems   []GormOrderItem `gorm:"foreignKey:OrderID"`
}

func (GormOrder) TableName() string {
	return "orders"
}

type GormOrderItem struct {
	ID          uint    `gorm:"primaryKey"`
	OrderID     uint    `gorm:"not null"`
	ProductID   uint    `gorm:"not null"`
	Name        string  `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Quantity    int     `gorm:"not null"`
	Image       string
	Category    string
	Description string
	CreatedAt   int64
}

func (GormOrderItem) TableName() string {
	return "order_items"
}

type GormOrderRepository struct {
	db *gorm.DB
}

func NewGormOrderRepository(db *gorm.DB) *GormOrderRepository {
	return &GormOrderRepository{db: db}
}

var _ appOrder.OrderRepository = (*GormOrderRepository)(nil)

func (r *GormOrderRepository) Create(ord *order.Order) error {
	gormOrder := r.toGormOrder(ord)

	err := r.db.Create(gormOrder).Error
	if err != nil {
		return err
	}

	ord.ID = gormOrder.ID
	for i, item := range gormOrder.OrderItems {
		ord.OrderItems[i].ID = item.ID
	}

	return nil
}

func (r *GormOrderRepository) GetByID(id uint) (*order.Order, error) {
	var gormOrder GormOrder
	err := r.db.Preload("OrderItems").First(&gormOrder, id).Error
	if err != nil {
		return nil, err
	}

	return r.toDomainOrder(&gormOrder), nil
}

func (r *GormOrderRepository) GetByUserID(userID uint) ([]*order.Order, error) {
	var gormOrders []GormOrder
	err := r.db.Preload("OrderItems").Where("user_id = ?", userID).Find(&gormOrders).Error
	if err != nil {
		return nil, err
	}

	orders := make([]*order.Order, len(gormOrders))
	for i, gormOrder := range gormOrders {
		orders[i] = r.toDomainOrder(&gormOrder)
	}

	return orders, nil
}

func (r *GormOrderRepository) GetAll() ([]*order.Order, error) {
	var gormOrders []GormOrder
	err := r.db.Preload("OrderItems").Find(&gormOrders).Error
	if err != nil {
		return nil, err
	}

	orders := make([]*order.Order, len(gormOrders))
	for i, gormOrder := range gormOrders {
		orders[i] = r.toDomainOrder(&gormOrder)
	}

	return orders, nil
}

func (r *GormOrderRepository) Update(ord *order.Order) error {
	gormOrder := r.toGormOrder(ord)
	return r.db.Save(gormOrder).Error
}

func (r *GormOrderRepository) Delete(id uint) error {
	return r.db.Select("OrderItems").Delete(&GormOrder{ID: id}).Error
}

func (r *GormOrderRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM order_items; DELETE FROM orders;").Error
}

func (r *GormOrderRepository) toGormOrder(ord *order.Order) *GormOrder {
	gormItems := make([]GormOrderItem, len(ord.OrderItems))
	for i, item := range ord.OrderItems {
		gormItems[i] = GormOrderItem{
			ID:          item.ID,
			OrderID:     item.OrderID,
			ProductID:   item.ProductID,
			Name:        item.ProductSnapshot.Name,
			Price:       item.ProductSnapshot.Price,
			Quantity:    item.Quantity,
			Image:       item.ProductSnapshot.Image,
			Category:    item.ProductSnapshot.Category,
			Description: item.ProductSnapshot.Description,
			CreatedAt:   item.CreatedAt.Unix(),
		}
	}

	return &GormOrder{
		ID:           ord.ID,
		UserID:       ord.UserID,
		TotalAmount:  ord.TotalAmount,
		Status:       string(ord.Status),
		CustomerName: ord.CustomerName,
		Phone:        ord.Phone,
		Address:      ord.Address,
		Email:        ord.Email,
		CreatedAt:    ord.CreatedAt.Unix(),
		UpdatedAt:    ord.UpdatedAt.Unix(),
		OrderItems:   gormItems,
	}
}

func (r *GormOrderRepository) toDomainOrder(gormOrder *GormOrder) *order.Order {
	items := make([]order.OrderItem, len(gormOrder.OrderItems))
	for i, gormItem := range gormOrder.OrderItems {
		items[i] = order.OrderItem{
			ID:        gormItem.ID,
			OrderID:   gormItem.OrderID,
			ProductID: gormItem.ProductID,
			ProductSnapshot: order.ProductSnapshot{
				Name:        gormItem.Name,
				Price:       gormItem.Price,
				Image:       gormItem.Image,
				Category:    gormItem.Category,
				Description: gormItem.Description,
			},
			Quantity:  gormItem.Quantity,
			CreatedAt: timeFromUnix(gormItem.CreatedAt),
		}
	}

	return &order.Order{
		ID:           gormOrder.ID,
		UserID:       gormOrder.UserID,
		OrderItems:   items,
		TotalAmount:  gormOrder.TotalAmount,
		Status:       order.OrderStatus(gormOrder.Status),
		CustomerName: gormOrder.CustomerName,
		Phone:        gormOrder.Phone,
		Address:      gormOrder.Address,
		Email:        gormOrder.Email,
		CreatedAt:    timeFromUnix(gormOrder.CreatedAt),
		UpdatedAt:    timeFromUnix(gormOrder.UpdatedAt),
	}
}
