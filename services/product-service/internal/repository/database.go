package repository

import (
	"log"
	"time"
	
	"product-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(databaseURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.Product{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData(db)
	return db
}

func seedData(db *gorm.DB) {
	var productCount int64
	db.Model(&models.Product{}).Count(&productCount)
	if productCount == 0 {
		now := time.Now()
		products := []models.Product{
			{
				Name:        "Áo Cầu Lông Yonex Pro",
				Description: "Áo cầu lông chính hãng Yonex, chất liệu thoáng mát, thấm hút mồ hôi tốt",
				Price:       450000,
				Stock:       20,
				Image:       "https://www.yonex.com/media/catalog/product/a/l/all_10700_629.png",
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
		for _, product := range products {
			db.Create(&product)
		}
		log.Println("Sample products created")
	}
}