package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// GenerateJWTSecret generates a new random JWT secret
func GenerateJWTSecret() (string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// LoadOrCreateJWTSecret loads an existing JWT secret from file or creates a new one
func LoadOrCreateJWTSecret(secretPath string) (string, error) {
	// Try to load existing secret
	if data, err := os.ReadFile(secretPath); err == nil {
		return string(data), nil
	}

	// Generate new secret
	secret, err := GenerateJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}

	// Save to file
	if err := os.WriteFile(secretPath, []byte(secret), 0600); err != nil {
		return "", fmt.Errorf("failed to save JWT secret: %w", err)
	}

	return secret, nil
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// EnsureDir ensures a directory exists, creating it if necessary
func EnsureDir(dir string, perm fs.FileMode) error {
	return os.MkdirAll(dir, perm)
}

// CreateOrUpdateFile creates a file with given content, creating parent directories if needed
func CreateOrUpdateFile(filePath string, content []byte, perm fs.FileMode) error {
	// Ensure parent directory exists
	if err := EnsureDir(filePath[:lastSlash(filePath)], 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, content, perm); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// lastSlash returns the index of the last slash in a path
func lastSlash(path string) int {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return i
		}
	}
	return len(path)
}
