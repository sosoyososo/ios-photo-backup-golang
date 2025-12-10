package models

import (
	"time"

	"gorm.io/gorm"
)

// Token represents a JWT token in the system
type Token struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"-" gorm:"not null;index"`
	TokenValue string        `json:"-" gorm:"uniqueIndex;not null"`
	CreatedAt time.Time      `json:"created_at"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"index"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for Token model
func (Token) TableName() string {
	return "tokens"
}
