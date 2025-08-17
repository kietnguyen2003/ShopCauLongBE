package main

import (
	"log"

	"api-gateway/internal/config"
	"api-gateway/internal/handlers"
	"api-gateway/internal/proxy"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	auth := r.Group("/auth")
	{
		auth.POST("/register", proxy.ProxyRequest(cfg.AuthServiceURL))
		auth.POST("/login", proxy.ProxyRequest(cfg.AuthServiceURL))
		auth.POST("/admin-login", proxy.ProxyRequest(cfg.AuthServiceURL))
		auth.POST("/validate", proxy.ProxyRequest(cfg.AuthServiceURL))
	}

	r.GET("/api/products", proxy.ProxyRequest(cfg.ProductServiceURL))
	r.GET("/api/products/:id", proxy.ProxyRequest(cfg.ProductServiceURL))

	api := r.Group("/api")
	api.Use(handlers.AuthMiddleware(cfg.JWTSecret))
	{
		api.POST("/orders", proxy.ProxyWithUserContext(cfg.OrderServiceURL))
		api.GET("/orders", proxy.ProxyWithUserContext(cfg.OrderServiceURL))
		api.GET("/orders/:id", proxy.ProxyRequest(cfg.OrderServiceURL))

		admin := api.Group("/admin")
		admin.Use(handlers.AdminMiddleware())
		{
			admin.GET("/orders", proxy.ProxyRequest(cfg.OrderServiceURL))
			admin.PUT("/orders/:id", proxy.ProxyRequest(cfg.OrderServiceURL))

			admin.POST("/products", proxy.ProxyRequest(cfg.ProductServiceURL))
			admin.PUT("/products/:id/stock", proxy.ProxyRequest(cfg.ProductServiceURL))
			admin.DELETE("/products/:id", proxy.ProxyRequest(cfg.ProductServiceURL))
		}
	}

	log.Printf("API Gateway starting on port %s", cfg.Port)
	log.Printf("Routing to services:")
	log.Printf("  Auth Service: %s", cfg.AuthServiceURL)
	log.Printf("  Product Service: %s", cfg.ProductServiceURL)
	log.Printf("  Order Service: %s", cfg.OrderServiceURL)
	
	log.Fatal(r.Run(":" + cfg.Port))
}