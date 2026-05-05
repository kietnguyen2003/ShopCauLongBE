package database

import (
	"errors"
	"strings"

	appOrder "kafka-order-demo/backend/internal/application/order"
	"kafka-order-demo/backend/internal/domain/coupon"
	"kafka-order-demo/backend/internal/domain/order"
	"kafka-order-demo/backend/internal/domain/product"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormOrder struct {
	ID             uint  `gorm:"primaryKey"`
	UserID         uint  `gorm:"not null"`
	SubtotalAmount int64 `gorm:"default:0"`
	DiscountAmount int64 `gorm:"default:0"`
	TotalAmount    int64 `gorm:"not null"`
	CouponID       *uint
	CouponCode     string
	CouponCodes    string
	Status         string `gorm:"default:pending"`
	CustomerName   string
	Phone          string
	Address        string
	Email          string
	CreatedAt      int64
	UpdatedAt      int64
	OrderItems     []GormOrderItem `gorm:"foreignKey:OrderID"`
}

func (GormOrder) TableName() string {
	return "orders"
}

type GormOrderItem struct {
	ID          uint   `gorm:"primaryKey"`
	OrderID     uint   `gorm:"not null"`
	ProductID   uint   `gorm:"not null"`
	Name        string `gorm:"not null"`
	Price       int64  `gorm:"not null"`
	Quantity    int    `gorm:"not null"`
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

func (r *GormOrderRepository) CreateWithProductStockUpdates(ord *order.Order, products []*product.Product) error {
	_ = products
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := decreaseProductStocks(tx, ord); err != nil {
			return err
		}

		gormOrder := r.toGormOrder(ord)
		if err := tx.Create(gormOrder).Error; err != nil {
			return err
		}

		ord.ID = gormOrder.ID
		for i, item := range gormOrder.OrderItems {
			ord.OrderItems[i].ID = item.ID
		}

		return clearCartItems(tx, ord.UserID)
	})
}

func (r *GormOrderRepository) CreateWithProductStockUpdatesAndCoupons(ord *order.Order, products []*product.Product, redemptions []*coupon.Redemption) error {
	_ = products
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := decreaseProductStocks(tx, ord); err != nil {
			return err
		}

		gormOrder := r.toGormOrder(ord)
		if err := tx.Create(gormOrder).Error; err != nil {
			return err
		}

		ord.ID = gormOrder.ID
		for i, item := range gormOrder.OrderItems {
			ord.OrderItems[i].ID = item.ID
		}

		for _, redemption := range redemptions {
			gormCoupon, err := lockCouponForRedemption(tx, redemption.CouponID)
			if err != nil {
				return err
			}
			userUsedCount, err := countCouponRedemptionsByUser(tx, redemption.CouponID, redemption.UserID)
			if err != nil {
				return err
			}
			if userUsedCount >= gormCoupon.UsageLimitPerUser {
				return coupon.ErrCouponAlreadyUsed
			}

			redemption.OrderID = ord.ID
			gormRedemption := &GormCouponRedemption{
				CouponID:       redemption.CouponID,
				UserID:         redemption.UserID,
				OrderID:        redemption.OrderID,
				DiscountAmount: redemption.DiscountAmount,
				CreatedAt:      redemption.CreatedAt.Unix(),
			}
			if err := tx.Create(gormRedemption).Error; err != nil {
				return err
			}
			redemption.ID = gormRedemption.ID

			result := tx.Model(&GormCoupon{}).
				Where("id = ? AND (usage_limit IS NULL OR used_count < usage_limit)", redemption.CouponID).
				Update("used_count", gorm.Expr("used_count + 1"))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return coupon.ErrCouponUsageLimitReached
			}
		}

		return clearCartItems(tx, ord.UserID)
	})
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

func (r *GormOrderRepository) UpdateWithProductRestock(ord *order.Order) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		gormOrder := r.toGormOrder(ord)
		if err := restockOrderItems(tx, gormOrder.OrderItems); err != nil {
			return err
		}

		return tx.Model(&GormOrder{}).
			Where("id = ?", gormOrder.ID).
			Updates(map[string]interface{}{
				"status":     gormOrder.Status,
				"updated_at": gormOrder.UpdatedAt,
			}).Error
	})
}

func (r *GormOrderRepository) Delete(id uint) error {
	return r.db.Select("OrderItems").Delete(&GormOrder{ID: id}).Error
}

func (r *GormOrderRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM order_items; DELETE FROM orders;").Error
}

func lockCouponForRedemption(tx *gorm.DB, couponID uint) (*GormCoupon, error) {
	var gormCoupon GormCoupon
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&gormCoupon, couponID).Error; err != nil {
		return nil, err
	}

	return &gormCoupon, nil
}

func countCouponRedemptionsByUser(tx *gorm.DB, couponID, userID uint) (int, error) {
	var count int64
	if err := tx.Model(&GormCouponRedemption{}).
		Where("coupon_id = ? AND user_id = ?", couponID, userID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func decreaseProductStocks(tx *gorm.DB, ord *order.Order) error {
	quantitiesByProductID := make(map[uint]int)
	for _, item := range ord.OrderItems {
		quantitiesByProductID[item.ProductID] += item.Quantity
	}

	for productID, quantity := range quantitiesByProductID {
		result := tx.Model(&GormProduct{}).
			Where("id = ? AND stock >= ?", productID, quantity).
			Updates(map[string]interface{}{
				"stock":      gorm.Expr("stock - ?", quantity),
				"updated_at": time.Now().Unix(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return product.ErrInsufficientStock
		}
	}

	return nil
}

func clearCartItems(tx *gorm.DB, userID uint) error {
	var gormCart GormCart
	if err := tx.Where("user_id = ? AND status = ?", userID, cartStatusActive).First(&gormCart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	return tx.Where("cart_id = ?", gormCart.ID).Delete(&GormCartItem{}).Error
}

func restockOrderItems(tx *gorm.DB, orderItems []GormOrderItem) error {
	quantitiesByProductID := make(map[uint]int)
	for _, item := range orderItems {
		quantitiesByProductID[item.ProductID] += item.Quantity
	}

	for productID, quantity := range quantitiesByProductID {
		result := tx.Model(&GormProduct{}).
			Where("id = ?", productID).
			Updates(map[string]interface{}{
				"stock":      gorm.Expr("stock + ?", quantity),
				"updated_at": time.Now().Unix(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
	}

	return nil
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
		ID:             ord.ID,
		UserID:         ord.UserID,
		SubtotalAmount: ord.SubtotalAmount,
		DiscountAmount: ord.DiscountAmount,
		TotalAmount:    ord.TotalAmount,
		CouponID:       ord.CouponID,
		CouponCode:     ord.CouponCode,
		CouponCodes:    strings.Join(ord.CouponCodes, ","),
		Status:         string(ord.Status),
		CustomerName:   ord.CustomerName,
		Phone:          ord.Phone,
		Address:        ord.Address,
		Email:          ord.Email,
		CreatedAt:      ord.CreatedAt.Unix(),
		UpdatedAt:      ord.UpdatedAt.Unix(),
		OrderItems:     gormItems,
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
		ID:             gormOrder.ID,
		UserID:         gormOrder.UserID,
		OrderItems:     items,
		SubtotalAmount: gormOrder.SubtotalAmount,
		DiscountAmount: gormOrder.DiscountAmount,
		TotalAmount:    gormOrder.TotalAmount,
		CouponID:       gormOrder.CouponID,
		CouponCode:     gormOrder.CouponCode,
		CouponCodes:    splitCouponCodes(gormOrder.CouponCodes, gormOrder.CouponCode),
		Status:         order.OrderStatus(gormOrder.Status),
		CustomerName:   gormOrder.CustomerName,
		Phone:          gormOrder.Phone,
		Address:        gormOrder.Address,
		Email:          gormOrder.Email,
		CreatedAt:      timeFromUnix(gormOrder.CreatedAt),
		UpdatedAt:      timeFromUnix(gormOrder.UpdatedAt),
	}
}

func splitCouponCodes(couponCodes, legacyCouponCode string) []string {
	source := couponCodes
	if source == "" {
		source = legacyCouponCode
	}
	if source == "" {
		return nil
	}

	parts := strings.Split(source, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
