package database

import (
	"strings"
	"time"

	appCoupon "kafka-order-demo/backend/internal/application/coupon"
	"kafka-order-demo/backend/internal/domain/coupon"

	"gorm.io/gorm"
)

type GormCoupon struct {
	ID                uint   `gorm:"primaryKey"`
	Code              string `gorm:"unique;not null"`
	Name              string
	Description       string
	DiscountType      string  `gorm:"not null"`
	DiscountValue     float64 `gorm:"not null"`
	MinOrderAmount    float64 `gorm:"default:0"`
	MaxDiscountAmount *float64
	UsageLimit        *int
	UsedCount         int `gorm:"default:0"`
	UsageLimitPerUser int `gorm:"default:1"`
	StartAt           *int64
	EndAt             *int64
	IsActive          bool `gorm:"default:true"`
	CreatedAt         int64
	UpdatedAt         int64
}

func (GormCoupon) TableName() string {
	return "coupons"
}

type GormCouponRedemption struct {
	ID             uint    `gorm:"primaryKey"`
	CouponID       uint    `gorm:"not null;index"`
	UserID         uint    `gorm:"not null;index"`
	OrderID        uint    `gorm:"not null;index"`
	DiscountAmount float64 `gorm:"not null"`
	CreatedAt      int64
}

func (GormCouponRedemption) TableName() string {
	return "coupon_redemptions"
}

type GormCouponRepository struct {
	db *gorm.DB
}

func NewGormCouponRepository(db *gorm.DB) *GormCouponRepository {
	return &GormCouponRepository{db: db}
}

var _ appCoupon.CouponRepository = (*GormCouponRepository)(nil)

func (r *GormCouponRepository) Create(coupon *coupon.Coupon) error {
	gormCoupon := r.toGormCoupon(coupon)
	if err := r.db.Create(gormCoupon).Error; err != nil {
		return err
	}

	coupon.ID = gormCoupon.ID
	return nil
}

func (r *GormCouponRepository) GetByID(id uint) (*coupon.Coupon, error) {
	var gormCoupon GormCoupon
	if err := r.db.First(&gormCoupon, id).Error; err != nil {
		return nil, err
	}

	return r.toDomainCoupon(&gormCoupon), nil
}

func (r *GormCouponRepository) GetByCode(code string) (*coupon.Coupon, error) {
	var gormCoupon GormCoupon
	if err := r.db.Where("code = ?", coupon.NormalizeCode(code)).First(&gormCoupon).Error; err != nil {
		return nil, err
	}

	return r.toDomainCoupon(&gormCoupon), nil
}

func (r *GormCouponRepository) GetWithQuery(query appCoupon.CouponQuery) ([]*coupon.Coupon, int64, error) {
	var gormCoupons []GormCoupon
	var total int64

	dbQuery := applyCouponQuery(r.db.Model(&GormCoupon{}), query)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Order("id DESC").Limit(query.Limit).Offset(offset).Find(&gormCoupons).Error; err != nil {
		return nil, 0, err
	}

	coupons := make([]*coupon.Coupon, len(gormCoupons))
	for i, gormCoupon := range gormCoupons {
		coupons[i] = r.toDomainCoupon(&gormCoupon)
	}

	return coupons, total, nil
}

func (r *GormCouponRepository) GetUnusedActiveWithQuery(query appCoupon.CouponQuery, userID uint) ([]*coupon.Coupon, int64, error) {
	var gormCoupons []GormCoupon
	var total int64

	isActive := true
	query.IsActive = &isActive
	dbQuery := applyCouponQuery(r.db.Model(&GormCoupon{}), query).
		Where("NOT EXISTS (?)",
			r.db.Model(&GormCouponRedemption{}).
				Select("1").
				Where("coupon_redemptions.coupon_id = coupons.id AND coupon_redemptions.user_id = ?", userID),
		)

	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.Limit
	if err := dbQuery.Order("id DESC").Limit(query.Limit).Offset(offset).Find(&gormCoupons).Error; err != nil {
		return nil, 0, err
	}

	coupons := make([]*coupon.Coupon, len(gormCoupons))
	for i, gormCoupon := range gormCoupons {
		coupons[i] = r.toDomainCoupon(&gormCoupon)
	}

	return coupons, total, nil
}

func (r *GormCouponRepository) Update(coupon *coupon.Coupon) error {
	return r.db.Save(r.toGormCoupon(coupon)).Error
}

func (r *GormCouponRepository) Delete(id uint) error {
	return r.db.Model(&GormCoupon{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now().Unix(),
	}).Error
}

func (r *GormCouponRepository) CountRedemptionsByCouponIDAndUserID(couponID, userID uint) (int, error) {
	var count int64
	if err := r.db.Model(&GormCouponRedemption{}).Where("coupon_id = ? AND user_id = ?", couponID, userID).Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

func applyCouponQuery(dbQuery *gorm.DB, query appCoupon.CouponQuery) *gorm.DB {
	if query.Search != "" {
		keyword := "%" + strings.ToLower(query.Search) + "%"
		dbQuery = dbQuery.Where("LOWER(code) LIKE ? OR LOWER(name) LIKE ?", keyword, keyword)
	}
	if query.IsActive != nil {
		dbQuery = dbQuery.Where("is_active = ?", *query.IsActive)
	}
	if query.DiscountType != "" {
		dbQuery = dbQuery.Where("discount_type = ?", query.DiscountType)
	}

	return dbQuery
}

func (r *GormCouponRepository) toGormCoupon(coupon *coupon.Coupon) *GormCoupon {
	return &GormCoupon{
		ID:                coupon.ID,
		Code:              coupon.Code,
		Name:              coupon.Name,
		Description:       coupon.Description,
		DiscountType:      coupon.DiscountType,
		DiscountValue:     coupon.DiscountValue,
		MinOrderAmount:    coupon.MinOrderAmount,
		MaxDiscountAmount: coupon.MaxDiscountAmount,
		UsageLimit:        coupon.UsageLimit,
		UsedCount:         coupon.UsedCount,
		UsageLimitPerUser: coupon.UsageLimitPerUser,
		StartAt:           unixPtr(coupon.StartAt),
		EndAt:             unixPtr(coupon.EndAt),
		IsActive:          coupon.IsActive,
		CreatedAt:         coupon.CreatedAt.Unix(),
		UpdatedAt:         coupon.UpdatedAt.Unix(),
	}
}

func (r *GormCouponRepository) toDomainCoupon(gormCoupon *GormCoupon) *coupon.Coupon {
	return &coupon.Coupon{
		ID:                gormCoupon.ID,
		Code:              gormCoupon.Code,
		Name:              gormCoupon.Name,
		Description:       gormCoupon.Description,
		DiscountType:      gormCoupon.DiscountType,
		DiscountValue:     gormCoupon.DiscountValue,
		MinOrderAmount:    gormCoupon.MinOrderAmount,
		MaxDiscountAmount: gormCoupon.MaxDiscountAmount,
		UsageLimit:        gormCoupon.UsageLimit,
		UsedCount:         gormCoupon.UsedCount,
		UsageLimitPerUser: gormCoupon.UsageLimitPerUser,
		StartAt:           timePtrFromUnix(gormCoupon.StartAt),
		EndAt:             timePtrFromUnix(gormCoupon.EndAt),
		IsActive:          gormCoupon.IsActive,
		CreatedAt:         timeFromUnix(gormCoupon.CreatedAt),
		UpdatedAt:         timeFromUnix(gormCoupon.UpdatedAt),
	}
}

func unixPtr(value *time.Time) *int64 {
	if value == nil {
		return nil
	}
	unix := value.Unix()
	return &unix
}

func timePtrFromUnix(value *int64) *time.Time {
	if value == nil {
		return nil
	}
	parsed := timeFromUnix(*value)
	return &parsed
}
