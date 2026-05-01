package database

import (
	"errors"
	"time"

	appCart "kafka-order-demo/backend/internal/application/cart"
	"kafka-order-demo/backend/internal/domain/cart"

	"gorm.io/gorm"
)

const cartStatusActive = "active"

type GormCart struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    uint   `gorm:"not null;uniqueIndex"`
	Status    string `gorm:"default:active"`
	CreatedAt int64
	UpdatedAt int64
	Items     []GormCartItem `gorm:"foreignKey:CartID"`
}

func (GormCart) TableName() string {
	return "carts"
}

type GormCartItem struct {
	ID        uint `gorm:"primaryKey"`
	CartID    uint `gorm:"not null;index;uniqueIndex:idx_cart_product"`
	UserID    uint `gorm:"not null;index"`
	ProductID uint `gorm:"not null;index;uniqueIndex:idx_cart_product"`
	Quantity  int  `gorm:"not null"`
	CreatedAt int64
	UpdatedAt int64
}

func (GormCartItem) TableName() string {
	return "cart_items"
}

type GormCartRepository struct {
	db *gorm.DB
}

func NewGormCartRepository(db *gorm.DB) *GormCartRepository {
	return &GormCartRepository{db: db}
}

var _ appCart.CartRepository = (*GormCartRepository)(nil)

func (r *GormCartRepository) GetByUserID(userID uint) ([]*cart.CartItem, error) {
	gormCart, err := r.getActiveCartByUserID(userID)
	if err != nil {
		if errors.Is(err, appCart.ErrCartItemNotFound) {
			return []*cart.CartItem{}, nil
		}
		return nil, err
	}

	var gormItems []GormCartItem
	if err := r.db.Where("cart_id = ?", gormCart.ID).Find(&gormItems).Error; err != nil {
		return nil, err
	}

	items := make([]*cart.CartItem, len(gormItems))
	for i, item := range gormItems {
		items[i] = r.toDomainCartItem(&item, userID)
	}

	return items, nil
}

func (r *GormCartRepository) GetByIDAndUserID(id, userID uint) (*cart.CartItem, error) {
	var gormItem GormCartItem
	if err := r.db.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("cart_items.id = ? AND carts.user_id = ? AND carts.status = ?", id, userID, cartStatusActive).
		First(&gormItem).Error; err != nil {
		return nil, normalizeCartNotFound(err)
	}

	return r.toDomainCartItem(&gormItem, userID), nil
}

func (r *GormCartRepository) GetByUserIDAndProductID(userID, productID uint) (*cart.CartItem, error) {
	var gormItem GormCartItem
	if err := r.db.Joins("JOIN carts ON carts.id = cart_items.cart_id").
		Where("carts.user_id = ? AND carts.status = ? AND cart_items.product_id = ?", userID, cartStatusActive, productID).
		First(&gormItem).Error; err != nil {
		return nil, normalizeCartNotFound(err)
	}

	return r.toDomainCartItem(&gormItem, userID), nil
}

func (r *GormCartRepository) Create(item *cart.CartItem) error {
	gormCart, err := r.getOrCreateActiveCart(item.UserID)
	if err != nil {
		return err
	}

	gormItem := r.toGormCartItem(item, gormCart.ID)
	if err := r.db.Create(gormItem).Error; err != nil {
		return err
	}

	item.ID = gormItem.ID
	return nil
}

func (r *GormCartRepository) Update(item *cart.CartItem) error {
	gormCart, err := r.getActiveCartByUserID(item.UserID)
	if err != nil {
		return err
	}

	return r.db.Model(&GormCartItem{}).
		Where("id = ? AND cart_id = ?", item.ID, gormCart.ID).
		Updates(map[string]interface{}{
			"quantity":   item.Quantity,
			"updated_at": item.UpdatedAt.Unix(),
		}).Error
}

func (r *GormCartRepository) Delete(id, userID uint) error {
	gormCart, err := r.getActiveCartByUserID(userID)
	if err != nil {
		return err
	}

	return r.db.Where("id = ? AND cart_id = ?", id, gormCart.ID).Delete(&GormCartItem{}).Error
}

func (r *GormCartRepository) Clear(userID uint) error {
	gormCart, err := r.getActiveCartByUserID(userID)
	if err != nil {
		if errors.Is(err, appCart.ErrCartItemNotFound) {
			return nil
		}
		return err
	}

	return r.db.Where("cart_id = ?", gormCart.ID).Delete(&GormCartItem{}).Error
}

func (r *GormCartRepository) toGormCartItem(item *cart.CartItem, cartID uint) *GormCartItem {
	return &GormCartItem{
		ID:        item.ID,
		CartID:    cartID,
		UserID:    item.UserID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		CreatedAt: item.CreatedAt.Unix(),
		UpdatedAt: item.UpdatedAt.Unix(),
	}
}

func (r *GormCartRepository) toDomainCartItem(item *GormCartItem, userID uint) *cart.CartItem {
	return &cart.CartItem{
		ID:        item.ID,
		UserID:    userID,
		ProductID: item.ProductID,
		Quantity:  item.Quantity,
		CreatedAt: timeFromUnix(item.CreatedAt),
		UpdatedAt: timeFromUnix(item.UpdatedAt),
	}
}

func (r *GormCartRepository) getActiveCartByUserID(userID uint) (*GormCart, error) {
	var gormCart GormCart
	if err := r.db.Where("user_id = ? AND status = ?", userID, cartStatusActive).First(&gormCart).Error; err != nil {
		return nil, normalizeCartNotFound(err)
	}

	return &gormCart, nil
}

func (r *GormCartRepository) getOrCreateActiveCart(userID uint) (*GormCart, error) {
	gormCart, err := r.getActiveCartByUserID(userID)
	if err == nil {
		return gormCart, nil
	}
	if !errors.Is(err, appCart.ErrCartItemNotFound) {
		return nil, err
	}

	now := time.Now().Unix()
	gormCart = &GormCart{
		UserID:    userID,
		Status:    cartStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := r.db.Create(gormCart).Error; err != nil {
		return nil, err
	}

	return gormCart, nil
}

func normalizeCartNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return appCart.ErrCartItemNotFound
	}

	return err
}
