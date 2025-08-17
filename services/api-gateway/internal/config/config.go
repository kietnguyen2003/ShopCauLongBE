package config

import (
	"os"
)

type Config struct {
	Port              string
	AuthServiceURL    string
	ProductServiceURL string
	OrderServiceURL   string
	JWTSecret         string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8080"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
		OrderServiceURL:   getEnv("ORDER_SERVICE_URL", "http://localhost:8083"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}