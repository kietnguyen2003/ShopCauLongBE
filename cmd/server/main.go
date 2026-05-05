package main

import (
	"log"

	"kafka-order-demo/backend/internal/application/address"
	"kafka-order-demo/backend/internal/application/auth"
	"kafka-order-demo/backend/internal/application/cart"
	"kafka-order-demo/backend/internal/application/category"
	"kafka-order-demo/backend/internal/application/coupon"
	"kafka-order-demo/backend/internal/application/notification"
	"kafka-order-demo/backend/internal/application/order"
	"kafka-order-demo/backend/internal/application/product"
	"kafka-order-demo/backend/internal/application/review"
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
	addressRepo := database.NewGormAddressRepository(db)
	cartRepo := database.NewGormCartRepository(db)
	categoryRepo := database.NewGormCategoryRepository(db)
	productRepo := database.NewGormProductRepository(db)
	orderRepo := database.NewGormOrderRepository(db)
	reviewRepo := database.NewGormReviewRepository(db)
	couponRepo := database.NewGormCouponRepository(db)
	notificationRepo := database.NewGormNotificationRepository(db)
	passwordHasher := security.NewBcryptHasher()
	tokenProvider := security.NewJWTProvider(cfg.JWTSecret)
	notificationHub := httpHandlers.NewNotificationHub()

	// Initialize services
	addressService := address.NewService(addressRepo)
	authService := auth.NewService(userRepo, passwordHasher, tokenProvider)
	cartService := cart.NewService(cartRepo, productRepo)
	categoryService := category.NewService(categoryRepo)
	productService := product.NewService(productRepo, categoryRepo)
	couponService := coupon.NewService(couponRepo, cartRepo, productRepo)
	notificationService := notification.NewService(notificationRepo, notificationHub)
	orderService := order.NewService(orderRepo, productRepo, addressRepo, cartRepo, couponRepo, notificationService)
	reviewService := review.NewService(reviewRepo, productRepo, orderRepo)

	// Initialize handlers
	addressHandler := httpHandlers.NewAddressHandler(addressService)
	authHandler := httpHandlers.NewAuthHandler(authService, tokenProvider)
	cartHandler := httpHandlers.NewCartHandler(cartService)
	categoryHandler := httpHandlers.NewCategoryHandler(categoryService)
	productHandler := httpHandlers.NewProductHandler(productService)
	orderHandler := httpHandlers.NewOrderHandler(orderService)
	reviewHandler := httpHandlers.NewReviewHandler(reviewService)
	couponHandler := httpHandlers.NewCouponHandler(couponService)
	notificationHandler := httpHandlers.NewNotificationHandler(notificationService, notificationHub)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	// Setup routes
	setupRoutes(r, addressHandler, authHandler, cartHandler, orderHandler, productHandler, categoryHandler, reviewHandler, couponHandler, notificationHandler)

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func setupRoutes(r *gin.Engine, addressHandler *httpHandlers.AddressHandler, authHandler *httpHandlers.AuthHandler, cartHandler *httpHandlers.CartHandler, orderHandler *httpHandlers.OrderHandler, productHandler *httpHandlers.ProductHandler, categoryHandler *httpHandlers.CategoryHandler, reviewHandler *httpHandlers.ReviewHandler, couponHandler *httpHandlers.CouponHandler, notificationHandler *httpHandlers.NotificationHandler) {
	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-login", authHandler.AdminLogin)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)

		auth.GET("/me", authHandler.AuthMiddleware(), authHandler.Me)
		auth.POST("/logout", authHandler.AuthMiddleware(), authHandler.Logout)
		auth.POST("/refresh-token", authHandler.AuthMiddleware(), authHandler.RefreshToken)
		auth.PUT("/change-password", authHandler.AuthMiddleware(), authHandler.ChangePassword)
	}

	// Public routes
	r.GET("/api/products", productHandler.GetProducts)
	r.GET("/api/products/search", productHandler.SearchProducts)
	r.GET("/api/products/category/:category", productHandler.GetProductsByCategory)
	r.GET("/api/products/:id/reviews", reviewHandler.GetProductReviews)
	r.GET("/api/products/:id", productHandler.GetProductByID)
	r.GET("/api/categories", categoryHandler.GetCategories)
	r.GET("/api/categories/:id", categoryHandler.GetCategoryByID)

	// Protected routes
	api := r.Group("/api")
	api.Use(authHandler.AuthMiddleware())
	{
		// User routes
		api.GET("/addresses", addressHandler.GetAddresses)
		api.POST("/addresses", addressHandler.CreateAddress)
		api.PUT("/addresses/:id", addressHandler.UpdateAddress)
		api.DELETE("/addresses/:id", addressHandler.DeleteAddress)
		api.PATCH("/addresses/:id/default", addressHandler.SetDefaultAddress)
		api.GET("/cart", cartHandler.GetCart)
		api.POST("/cart/items", cartHandler.AddItem)
		api.PUT("/cart/items/:id", cartHandler.UpdateItem)
		api.DELETE("/cart/items/:id", cartHandler.DeleteItem)
		api.DELETE("/cart/clear", cartHandler.ClearCart)
		api.POST("/orders", orderHandler.CreateOrder)
		api.GET("/orders", orderHandler.GetUserOrders)
		api.POST("/products/:id/reviews", reviewHandler.CreateReview)
		api.PUT("/reviews/:id", reviewHandler.UpdateReview)
		api.DELETE("/reviews/:id", reviewHandler.DeleteReview)
		api.GET("/coupons", couponHandler.GetActiveCoupons)
		api.GET("/coupons/:id", couponHandler.GetActiveCoupon)
		api.POST("/coupons/validate", couponHandler.ValidateCoupon)
		api.GET("/notifications", notificationHandler.GetNotifications)
		api.GET("/notifications/unread-count", notificationHandler.GetUnreadCount)
		api.PATCH("/notifications/:id/read", notificationHandler.MarkAsRead)
		api.PATCH("/notifications/read-all", notificationHandler.MarkAllAsRead)
		api.GET("/notifications/ws", notificationHandler.StreamNotifications)

		// Admin routes
		admin := api.Group("/admin")
		admin.Use(authHandler.AdminMiddleware())
		{
			admin.GET("/orders", orderHandler.GetAllOrders)
			admin.PUT("/orders/:id", orderHandler.UpdateOrderStatus)
			admin.POST("/products", productHandler.CreateProduct)
			admin.PUT("/products/:id", productHandler.UpdateProduct)
			admin.DELETE("/products/:id", productHandler.DeleteProduct)
			admin.PATCH("/products/:id/stock", productHandler.UpdateProductStock)
			admin.POST("/categories", categoryHandler.CreateCategory)
			admin.PUT("/categories/:id", categoryHandler.UpdateCategory)
			admin.DELETE("/categories/:id", categoryHandler.DeleteCategory)
			admin.GET("/coupons", couponHandler.GetCoupons)
			admin.POST("/coupons", couponHandler.CreateCoupon)
			admin.GET("/coupons/:id", couponHandler.GetCoupon)
			admin.PUT("/coupons/:id", couponHandler.UpdateCoupon)
			admin.DELETE("/coupons/:id", couponHandler.DeleteCoupon)
			admin.PATCH("/coupons/:id/status", couponHandler.UpdateCouponStatus)
		}
	}
}
