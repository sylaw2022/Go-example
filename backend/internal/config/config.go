package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	Port     string
	LogLevel string
}

// LoadConfig loads the configuration from environment variables
func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		Port:     port,
		LogLevel: logLevel,
	}
}
