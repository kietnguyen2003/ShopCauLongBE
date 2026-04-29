package database

import (
	"time"

	"gorm.io/gorm"
)

func SeedInitialData(db *gorm.DB) error {
	if err := seedAdminUser(db); err != nil {
		return err
	}

	if err := seedCategories(db); err != nil {
		return err
	}

	if err := seedProducts(db); err != nil {
		return err
	}

	return nil
}

func seedAdminUser(db *gorm.DB) error {
	var adminCount int64
	if err := db.Model(&GormUser{}).Where("is_admin = ?", true).Count(&adminCount).Error; err != nil {
		return err
	}

	if adminCount > 0 {
		return nil
	}

	hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi"
	admin := GormUser{
		Username:  "admin",
		Email:     "admin@demo.com",
		Password:  hashedPassword,
		IsAdmin:   true,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	return db.Create(&admin).Error
}

func seedCategories(db *gorm.DB) error {
	var categoryCount int64
	if err := db.Model(&GormCategory{}).Count(&categoryCount).Error; err != nil {
		return err
	}

	if categoryCount > 0 {
		return nil
	}

	now := time.Now().Unix()
	categories := []GormCategory{
		{
			Name:        "clothing",
			Description: "Quan ao cau long",
			Image:       "https://cdn.shopvnb.com/uploads/gallery/ao-cau-long-mizuno-vm1076-nam-xanh_1728499552.webp",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "racket",
			Description: "Vot cau long",
			Image:       "https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-yonex-arcsaber-11-pro-chinh-hang-1.webp",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	return db.Create(&categories).Error
}

func seedProducts(db *gorm.DB) error {
	var productCount int64
	if err := db.Model(&GormProduct{}).Count(&productCount).Error; err != nil {
		return err
	}

	if productCount > 0 {
		return nil
	}

	now := time.Now().Unix()
	products := []GormProduct{
		{
			Name:        "Áo Cầu Lông Yonex Pro",
			Description: "Áo cầu lông chính hãng Yonex, chất liệu thoáng mát, thấm hút mồ hôi tốt",
			Price:       450000,
			Stock:       20,
			Image:       "https://www.yonex.com/media/catalog/product/a/l/all_10700_629.png?quality=80&bg-color=248,248,248,0.75&fit=bounds&height=300&width=240&canvas=240:300",
			Category:    "clothing",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Quần Cầu Lông Victor",
			Description: "Quần cầu lông Victor, thiết kế năng động, thoải mái khi vận động",
			Price:       380000,
			Stock:       15,
			Image:       "https://cdn.shopvnb.com/uploads/gallery/quan-cau-long-victor-q41-nam-xanh_1715654738.webp",
			Category:    "clothing",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Áo Cầu Lông Mizuno",
			Description: "Giày cầu lông Mizuno chuyên nghiệp, đế chống trượt, hỗ trợ di chuyển tốt",
			Price:       1200000,
			Stock:       10,
			Image:       "https://cdn.shopvnb.com/img/300x300/uploads/gallery/ao-cau-long-mizuno-vm1076-nam-xanh_1728499552.webp",
			Category:    "clothing",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Vợt Cầu Lông Yonex Arcsaber 11",
			Description: "Vợt cầu lông Yonex Arcsaber 11, công nghệ tiên tiến, phù hợp cho người chơi chuyên nghiệp",
			Price:       3500000,
			Stock:       8,
			Image:       "https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-yonex-arcsaber-11-pro-chinh-hang-1.webp",
			Category:    "racket",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "Vợt Cầu Lông Victor Jetspeed S12",
			Description: "Vợt Victor Jetspeed S12, thiết kế aerodynamic, tốc độ swing nhanh",
			Price:       2800000,
			Stock:       12,
			Image:       "https://cdn.shopvnb.com/img/600x600/uploads/san_pham/combo-mua-vot-cau-long-victor-jetspeed-12-js-12-tang-vot-js120-2-vot-victor-tk9-1.webp",
			Category:    "racket",
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	return db.Create(&products).Error
}
