package database

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&GormUser{}, &GormCategory{}, &GormProduct{}, &GormOrder{}, &GormOrderItem{}, &GormCart{}, &GormCartItem{})
}
