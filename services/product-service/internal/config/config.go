package config

import (
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8082"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/product_db?sslmode=disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}