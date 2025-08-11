package main

import (
	"log"
	"kafka-order-demo/backend/config"
	"kafka-order-demo/backend/handlers"
	"kafka-order-demo/backend/kafka"
	"kafka-order-demo/backend/models"
	"kafka-order-demo/backend/websocket"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize config
	cfg := config.Load()

	// Initialize database
	db := models.InitDB(cfg.DatabaseURL)

	// Initialize Kafka
	kafkaManager := kafka.NewManager(cfg.KafkaBrokers)
	go kafkaManager.StartConsumers()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg.JWTSecret)
	orderHandler := handlers.NewOrderHandler(db, kafkaManager, wsHub)
	productHandler := handlers.NewProductHandler(db)
	wsHandler := websocket.NewHandler(wsHub)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Routes
	setupRoutes(r, authHandler, orderHandler, productHandler, wsHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, orderHandler *handlers.OrderHandler, productHandler *handlers.ProductHandler, wsHandler *websocket.Handler) {
	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-login", authHandler.AdminLogin)
	}

	// WebSocket routes
	r.GET("/ws/user/:userID", wsHandler.HandleUserConnection)
	r.GET("/ws/admin", wsHandler.HandleAdminConnection)

	// Public routes
	r.GET("/api/products", productHandler.GetProducts)
	r.GET("/api/products/:id", productHandler.GetProductByID)

	// Protected routes
	api := r.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	{
		// User routes
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders", orderHandler.GetUserOrders)

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(authHandler.AdminMiddleware())
		{
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.PUT("/orders/:id", orderHandler.UpdateOrderStatus)
		}
	}
}