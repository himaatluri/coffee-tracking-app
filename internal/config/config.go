package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	Port          string
	DBConfig      DatabaseConfig
	MaxUploadSize int64
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Port: getEnvOrDefault("PORT", ":8080"),
		DBConfig: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "coffee_tracker"),
			SSLMode:  getEnvOrDefault("DB_SSL_MODE", "disable"),
		},
		MaxUploadSize: 5 << 20, // 5MB
	}
}

// getEnvOrDefault returns the value of an environment variable or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
