package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Host string
	Port int

	// Directories
	DataDir    string
	StorageDir string

	// Database
	DatabasePath string

	// JWT
	JWTSecretPath string
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Host:          "0.0.0.0",
		Port:          8080,
		DataDir:       "./data",
		StorageDir:    "./storage",
		DatabasePath:  "./data/app.db",
		JWTSecretPath: "./data/jwt_secret.key",
	}
}

// Load loads configuration from command-line flags and environment variables
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Command-line flags
	flag.StringVar(&cfg.Host, "host", cfg.Host, "Server host address")
	flag.IntVar(&cfg.Port, "port", cfg.Port, "Server port")
	flag.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Data directory path")
	flag.StringVar(&cfg.StorageDir, "storage-dir", cfg.StorageDir, "Storage directory path")
	flag.StringVar(&cfg.DatabasePath, "db-path", cfg.DatabasePath, "Database file path")
	flag.StringVar(&cfg.JWTSecretPath, "jwt-secret-path", cfg.JWTSecretPath, "JWT secret file path")

	flag.Parse()

	// Environment variables override defaults
	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}
	if port := os.Getenv("PORT"); port != "" {
		if p, err := fmt.Sscanf(port, "%d", &cfg.Port); err != nil && p == 1 {
			return nil, fmt.Errorf("invalid PORT value: %s", port)
		}
	}

	return cfg, nil
}
