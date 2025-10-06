package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	RedisURL         string
	Port             string
	GinMode          string
	JWTSecret        string
	JWTExpiry        string
	SuperAdminEmail  string
	SuperAdminPassword string
}

func Load() *Config {
	return &Config{
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/breakoutglobe?sslmode=disable"),
		RedisURL:           getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:               getEnv("PORT", "8080"),
		GinMode:            getEnv("GIN_MODE", "debug"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          getEnv("JWT_EXPIRY", "24h"),
		SuperAdminEmail:    getEnv("SUPERADMIN_EMAIL", ""),
		SuperAdminPassword: getEnv("SUPERADMIN_PASSWORD", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}