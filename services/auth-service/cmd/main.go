package main

import (
	"log"

	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/repository"
	"auth-service/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := repository.NewConnection(cfg.DatabaseURL)

	userRepo := repository.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authService, cfg.JWTSecret)

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/admin-login", authHandler.AdminLogin)
		auth.POST("/validate", authHandler.ValidateToken)
	}

	log.Printf("Auth Service starting on port %s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}