package repository

import (
	"log"
	
	"order-service/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewConnection(databaseURL string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.AutoMigrate(&models.Order{}, &models.OrderItem{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	return db
}