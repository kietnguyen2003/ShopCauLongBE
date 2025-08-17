package config

import (
	"os"
)

type Config struct {
	Port               string
	DatabaseURL        string
	ProductServiceURL  string
}

func Load() *Config {
	return &Config{
		Port:              getEnv("PORT", "8083"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/order_db?sslmode=disable"),
		ProductServiceURL: getEnv("PRODUCT_SERVICE_URL", "http://localhost:8082"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}