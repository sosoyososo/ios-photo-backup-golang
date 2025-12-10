package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PhotoNaming handles sequential photo naming logic
type PhotoNaming struct{}

// NewPhotoNaming creates a new PhotoNaming instance
func NewPhotoNaming() *PhotoNaming {
	return &PhotoNaming{}
}

// GenerateFilename generates a sequential filename for a photo
// Format: IMG_XXXX.ext where XXXX is a 4-digit zero-padded number
func (n *PhotoNaming) GenerateFilename(extension string, sequenceNumber int) string {
	// Format: IMG_XXXX.ext (4-digit zero-padded)
	sequence := strconv.Itoa(sequenceNumber)
	paddedSequence := strings.Repeat("0", 4-len(sequence)) + sequence

	return fmt.Sprintf("IMG_%s.%s", paddedSequence, extension)
}

// ParseDate parses a date string in YYYY-MM-DD format
func (n *PhotoNaming) ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// GetDirectoryPath generates the directory path for a user's photos
// Format: storage/photo/{user_id}/{year}/{month}/{day}/
func (n *PhotoNaming) GetDirectoryPath(storageDir string, userID uint, date time.Time) string {
	year := date.Format("2006")
	month := date.Format("01")
	day := date.Format("02")

	return fmt.Sprintf("%s/photo/%d/%s/%s/%s/",
		strings.TrimRight(storageDir, "/"),
		userID,
		year,
		month,
		day,
	)
}

// GetNextSequenceNumber calculates the next sequence number for a date
// This is a placeholder - actual implementation will query the database
func (n *PhotoNaming) GetNextSequenceNumber(existingCount int) int {
	// Start from 1, so first photo is IMG_0001
	return existingCount + 1
}
