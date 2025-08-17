package main

import (
	"log"

	"product-service/internal/config"
	"product-service/internal/handlers"
	"product-service/internal/repository"
	"product-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := repository.NewConnection(cfg.DatabaseURL)

	productRepo := repository.NewProductRepository(db)
	productService := service.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.GET("/products", productHandler.GetProducts)
		api.GET("/products/:id", productHandler.GetProductByID)
		api.POST("/products", productHandler.CreateProduct)
		api.PUT("/products/:id/stock", productHandler.UpdateStock)
		api.POST("/products/:id/decrease-stock", productHandler.DecreaseStock)
		api.DELETE("/products/:id", productHandler.DeleteProduct)
	}

	log.Printf("Product Service starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}