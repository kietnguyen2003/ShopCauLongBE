package database

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&GormUser{}, &GormAddress{}, &GormCategory{}, &GormProduct{}, &GormOrder{}, &GormOrderItem{}, &GormCart{}, &GormCartItem{}, &GormReview{}); err != nil {
		return err
	}

	if err := db.Exec("DROP INDEX IF EXISTS idx_reviews_product_user").Error; err != nil {
		return err
	}

	return db.Exec("ALTER TABLE reviews ALTER COLUMN rating DROP NOT NULL").Error
}
