package config

import (
	"os"
	"strings"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	KafkaBrokers []string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://user:password@localhost:5432/kafka_demo?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "c30458d674115284b466558488a3d413"),
		KafkaBrokers: getKafkaBrokers(),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getKafkaBrokers() []string {
	brokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	return strings.Split(brokers, ",")
}
