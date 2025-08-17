package main

import (
	"log"

	"order-service/internal/config"
	"order-service/internal/handlers"
	"order-service/internal/repository"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := repository.NewConnection(cfg.DatabaseURL)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, cfg.ProductServiceURL)
	orderHandler := handlers.NewOrderHandler(orderService)

	r := gin.Default()

	api := r.Group("/api")
	{
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders", orderHandler.GetUserOrders)
		api.GET("/orders/:id", orderHandler.GetOrderByID)
		api.PUT("/orders/:id", orderHandler.UpdateOrderStatus)
		
		admin := api.Group("/admin")
		{
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.PUT("/orders/:id", orderHandler.UpdateOrderStatus)
		}
	}

	log.Printf("Order Service starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}