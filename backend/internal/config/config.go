// Package config provides application configuration loaded from environment variables.
package config

import (
	"os"
)

// Config holds all configuration values for the Ponte server.
type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

// Load reads configuration from environment variables, applying sensible defaults.
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://ponte:ponte@localhost:5432/ponte?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "ponte-dev-secret-change-in-production"),
		Port:        getEnv("PORT", "8080"),
	}
}

// getEnv returns the value of an environment variable or a fallback default.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
