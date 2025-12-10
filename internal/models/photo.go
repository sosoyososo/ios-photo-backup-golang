package models

import (
	"time"

	"gorm.io/gorm"
)

// Photo represents a photo in the system
// This model is used for dynamic tables (photos_user_{user_id})
type Photo struct {
	LocalID       string         `json:"local_id" gorm:"primaryKey"`
	CreationTime  time.Time      `json:"creation_time" gorm:"not null;index"`
	FilePath      string         `json:"file_path" gorm:"not null;index"`
	FileName      string         `json:"file_name" gorm:"not null;size:255"`
	FileType      string         `json:"file_type" gorm:"not null;size:50"`
	FileCount     int            `json:"file_count" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName is not defined here because Photo models
// use dynamic table names based on user ID
