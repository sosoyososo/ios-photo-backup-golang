package repository

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
)

// InitDB initializes and returns a database connection
func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Enable logging to debug database issues
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// AutoMigrate runs database migrations for users and tokens tables
func AutoMigrate(db *gorm.DB) error {
	// Migrate User and Token models
	// Photo tables are created dynamically per user
	if err := db.AutoMigrate(
		&models.User{},
		&models.Token{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
