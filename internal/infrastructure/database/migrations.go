package database

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&GormUser{}, &GormAddress{}, &GormCategory{}, &GormProduct{}, &GormOrder{}, &GormOrderItem{}, &GormCart{}, &GormCartItem{}, &GormReview{}, &GormCoupon{}, &GormCouponRedemption{}, &GormNotification{}); err != nil {
		return err
	}

	return nil
}
