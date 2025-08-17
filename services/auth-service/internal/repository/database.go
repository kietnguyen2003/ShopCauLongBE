package repository

import (
	"log"
	"time"
	
	"auth-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"golang.org/x/crypto/bcrypt"
)

func NewConnection(databaseURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData(db)
	return db
}

func seedData(db *gorm.DB) {
	var adminCount int64
	db.Model(&models.User{}).Where("is_admin = ?", true).Count(&adminCount)
	if adminCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		admin := models.User{
			Username:  "admin",
			Email:     "admin@demo.com",
			Password:  string(hashedPassword),
			IsAdmin:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(&admin)
		log.Println("Admin user created: admin / admin123")
	}
}