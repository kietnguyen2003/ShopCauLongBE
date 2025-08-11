package models

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"unique;not null"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"`
	IsAdmin   bool      `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"default:0"`
	Image       string    `json:"image"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID           uint        `json:"id" gorm:"primaryKey"`
	UserID       uint        `json:"user_id" gorm:"not null"`
	User         User        `json:"user" gorm:"foreignKey:UserID"`
	OrderItems   []OrderItem `json:"order_items" gorm:"foreignKey:OrderID"`
	TotalAmount  float64     `json:"total_amount" gorm:"not null"`
	Status       string      `json:"status" gorm:"default:pending"`
	CustomerName string      `json:"customer_name"`
	Phone        string      `json:"phone"`
	Address      string      `json:"address"`
	Email        string      `json:"email"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	OrderID     uint      `json:"order_id" gorm:"not null"`
	ProductID   uint      `json:"product_id" gorm:"not null"`
	Product     Product   `json:"product" gorm:"foreignKey:ProductID"`
	Name        string    `json:"name" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"`
	Quantity    int       `json:"quantity" gorm:"not null"`
	Image       string    `json:"image"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func InitDB(databaseURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto migrate schemas
	err = db.AutoMigrate(&User{}, &Product{}, &Order{}, &OrderItem{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed initial data
	seedData(db)

	return db
}

func seedData(db *gorm.DB) {
	// Create admin user if not exists
	var adminCount int64
	db.Model(&User{}).Where("is_admin = ?", true).Count(&adminCount)
	if adminCount == 0 {
		// Hash password: admin123
		hashedPassword := "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi" // bcrypt hash of "admin123"
		admin := User{
			Username: "admin",
			Email:    "admin@demo.com",
			Password: hashedPassword,
			IsAdmin:  true,
		}
		db.Create(&admin)
	}

	// Create sample products if not exists
	var productCount int64
	db.Model(&Product{}).Count(&productCount)
	if productCount == 0 {
		products := []Product{
			// Quần áo
			{
				Name:        "Áo Cầu Lông Yonex Pro",
				Description: "Áo cầu lông chính hãng Yonex, chất liệu thoáng mát, thấm hút mồ hôi tốt",
				Price:       450000,
				Stock:       20,
				Image:       "https://www.yonex.com/media/catalog/product/a/l/all_10700_629.png?quality=80&bg-color=248,248,248,0.75&fit=bounds&height=300&width=240&canvas=240:300",
				Category:    "clothing",
			},
			{
				Name:        "Quần Cầu Lông Victor",
				Description: "Quần cầu lông Victor, thiết kế năng động, thoải mái khi vận động",
				Price:       380000,
				Stock:       15,
				Image:       "https://cdn.shopvnb.com/uploads/gallery/quan-cau-long-victor-q41-nam-xanh_1715654738.webp",
				Category:    "clothing",
			},
			{
				Name:        "Áo Cầu Lông Mizuno",
				Description: "Giày cầu lông Mizuno chuyên nghiệp, đế chống trượt, hỗ trợ di chuyển tốt",
				Price:       1200000,
				Stock:       10,
				Image:       "https://cdn.shopvnb.com/img/300x300/uploads/gallery/ao-cau-long-mizuno-vm1076-nam-xanh_1728499552.webp",
				Category:    "clothing",
			},

			// Vợt
			{
				Name:        "Vợt Cầu Lông Yonex Arcsaber 11",
				Description: "Vợt cầu lông Yonex Arcsaber 11, công nghệ tiên tiến, phù hợp cho người chơi chuyên nghiệp",
				Price:       3500000,
				Stock:       8,
				Image:       "https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-yonex-arcsaber-11-pro-chinh-hang-1.webp",
				Category:    "racket",
			},
			{
				Name:        "Vợt Cầu Lông Victor Jetspeed S12",
				Description: "Vợt Victor Jetspeed S12, thiết kế aerodynamic, tốc độ swing nhanh",
				Price:       2800000,
				Stock:       12,
				Image:       "https://cdn.shopvnb.com/img/600x600/uploads/san_pham/combo-mua-vot-cau-long-victor-jetspeed-12-js-12-tang-vot-js120-2-vot-victor-tk9-1.webp",
				Category:    "racket",
			},
			{
				Name:        "Vợt Cầu Lông Li-Ning Windstorm 78",
				Description: "Vợt Li-Ning Windstorm 78, cân bằng tốt giữa sức mạnh và kiểm soát",
				Price:       2200000,
				Stock:       6,
				Image:       "https://cdn.shopvnb.com/uploads/san_pham/vot-cau-long-lining-windstorm-78-trang-vang-chinh-hang-3.webp",
				Category:    "racket",
			},

			// Dụng cụ
			{
				Name:        "Cầu Lông Yonex Mavis 350",
				Description: "Cầu lông Yonex Mavis 350, chất lượng cao, độ bền tốt (hộp 12 quả)",
				Price:       180000,
				Stock:       50,
				Image:       "https://cdn.shopvnb.com/uploads/san_pham/ong-cau-long-nhua-yonex-mav-350-6-in-1-vang-1.webp",
				Category:    "equipment",
			},
			{
				Name:        "Túi Đựng Vợt Cầu Lông",
				Description: "Túi đựng vợt cầu lông, chứa được 6 cây vợt, có ngăn phụ đựng phụ kiện",
				Price:       320000,
				Stock:       25,
				Image:       "https://cdn.shopvnb.com/uploads/san_pham/tui-dung-vot-cau-long-kason-fbjk022-3000-xanh-duong-3.webp",
				Category:    "equipment",
			},
			{
				Name:        "Dây Cước Vợt Cầu Lông",
				Description: "Dây cước vợt cầu lông chất lượng cao, độ đàn hồi tốt",
				Price:       150000,
				Stock:       30,
				Image:       "https://cdn.shopvnb.com/uploads/gallery/day-cuoc-cang-vot-lining-l64-noi-dia-1_1752883197.webp",
				Category:    "equipment",
			},
			{
				Name:        "Băng Quấn Cán Vợt",
				Description: "Băng quấn cán vợt, chống trượt, thấm hút mồ hôi",
				Price:       45000,
				Stock:       40,
				Image:       "https://cdn.shopvnb.com/img/600x600/uploads/san_pham/quan-can-yonex-xin-ac-109-cai-1.webp",
				Category:    "equipment",
			},
		}
		for _, product := range products {
			db.Create(&product)
		}
	}
}