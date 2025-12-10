package repository

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
)

// UserRepository provides CRUD operations for users
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.User{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListAll lists all users
func (r *UserRepository) ListAll() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}
