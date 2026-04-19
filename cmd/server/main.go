package main

import (
	"log"

	"kafka-order-demo/backend/internal/application/auth"
	"kafka-order-demo/backend/internal/application/order"
	"kafka-order-demo/backend/internal/application/product"
	"kafka-order-demo/backend/internal/infrastructure/config"
	"kafka-order-demo/backend/internal/infrastructure/database"
	"kafka-order-demo/backend/internal/infrastructure/security"
	httpHandlers "kafka-order-demo/backend/internal/interfaces/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database infrastructure
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	if err := database.SeedInitialData(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Initialize repositories
	userRepo := database.NewGormUserRepository(db)
	productRepo := database.NewGormProductRepository(db)
	orderRepo := database.NewGormOrderRepository(db)
	passwordHasher := security.NewBcryptHasher()
	tokenProvider := security.NewJWTProvider(cfg.JWTSecret)

	// Initialize services
	authService := auth.NewService(userRepo, passwordHasher, tokenProvider)
	productService := product.NewService(productRepo)
	orderService := order.NewService(orderRepo, productRepo)

	// Initialize handlers
	authHandler := httpHandlers.NewAuthHandler(authService, tokenProvider)
	productHandler := httpHandlers.NewProductHandler(productService)
	orderHandler := httpHandlers.NewOrderHandler(orderService)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Setup routes
	setupRoutes(r, authHandler, orderHandler, productHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func setupRoutes(r *gin.Engine, authHandler *httpHandlers.AuthHandler, orderHandler *httpHandlers.OrderHandler, productHandler *httpHandlers.ProductHandler) {
	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-login", authHandler.AdminLogin)
	}

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
