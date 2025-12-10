package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account in the system
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null;size:255"`
	PasswordHash string         `json:"-" gorm:"not null;size:255"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
