package config

import (
	"fmt"
)

// InitializeApp initializes the application by creating necessary directories
// and setting up the JWT secret
func InitializeApp(cfg *Config) error {
	// Create data directory
	if err := EnsureDir(cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create storage directory
	if err := EnsureDir(cfg.StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load or create JWT secret
	if _, err := LoadOrCreateJWTSecret(cfg.JWTSecretPath); err != nil {
		return fmt.Errorf("failed to setup JWT secret: %w", err)
	}

	return nil
}
