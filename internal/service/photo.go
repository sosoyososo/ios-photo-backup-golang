package service

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
	"github.com/ios-photo-backup/photo-backup-server/internal/repository"
)

// PhotoService handles photo operations
type PhotoService struct {
	photoRepo     *repository.PhotoRepository
	naming        *PhotoNaming
	fileStorage   *FileStorage
	storageDir    string
}

// NewPhotoService creates a new PhotoService
func NewPhotoService(
	photoRepo *repository.PhotoRepository,
	naming *PhotoNaming,
	fileStorage *FileStorage,
	storageDir string,
) *PhotoService {
	return &PhotoService{
		photoRepo:   photoRepo,
		naming:      naming,
		fileStorage: fileStorage,
		storageDir:  storageDir,
	}
}

// PhotoIndexRequest represents a photo indexing request
type PhotoIndexRequest struct {
	LocalID       string    `json:"local_id"`
	CreationTime  time.Time `json:"creation_time"`
	FileType      string    `json:"file_type"` // File extension (e.g., "jpg", "heic", "png")
}

// PhotoIndexResponse represents a photo indexing response
type PhotoIndexResponse struct {
	LocalID            string   `json:"local_id"`
	UploadedExtensions []string `json:"uploaded_extensions"`
}

// IndexPhotos indexes a batch of photos and assigns filenames
func (s *PhotoService) IndexPhotos(userID uint, dateStr string, photos []PhotoIndexRequest) ([]PhotoIndexResponse, error) {
	// Parse date
	date, err := s.naming.ParseDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Get directory path
	dirPath := s.naming.GetDirectoryPath(s.storageDir, userID, date)

	// Check for orphaned files (non-DB files in directory)
	// This is a placeholder - actual implementation would scan directory
	// and check against database

	// Get existing photo count for this date
	existingCount, err := s.photoRepo.GetCountByDate(dateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing photo count: %w", err)
	}

	// Sort photos by creation time
	sortedPhotos := make([]PhotoIndexRequest, len(photos))
	copy(sortedPhotos, photos)
	sort.Slice(sortedPhotos, func(i, j int) bool {
		return sortedPhotos[i].CreationTime.Before(sortedPhotos[j].CreationTime)
	})

	// Assign filenames
	var responses []PhotoIndexResponse
	nextSequence := s.naming.GetNextSequenceNumber(existingCount)

	for _, photo := range sortedPhotos {
		// Check if photo already exists (re-index logic)
		existingPhoto, err := s.photoRepo.FindByLocalID(photo.LocalID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing photo: %w", err)
		}

		// Get uploaded extensions for this photo
		uploadedExtensions, err := s.photoRepo.GetUploadedExtensions(photo.LocalID)
		if err != nil {
			return nil, fmt.Errorf("failed to get uploaded extensions: %w", err)
		}

		if existingPhoto != nil {
			// Photo already exists, return existing photo info
			responses = append(responses, PhotoIndexResponse{
				LocalID:            photo.LocalID,
				UploadedExtensions: uploadedExtensions,
			})
			continue
		}

		// Generate filename (without extension)
		filename := s.naming.GenerateFilename(nextSequence)
		nextSequence++

		// Create photo record
		photoRecord := &models.Photo{
			LocalID:       photo.LocalID,
			CreationTime:  photo.CreationTime,
			FilePath:      dirPath,
			FileName:      filename,
			FileType:      photo.FileType,
			FileCount:     0,
		}

		// Save to database
		if err := s.photoRepo.Create(photoRecord); err != nil {
			return nil, fmt.Errorf("failed to create photo record: %w", err)
		}

		// Response with uploaded extensions (empty for new photos)
		responses = append(responses, PhotoIndexResponse{
			LocalID:            photo.LocalID,
			UploadedExtensions: uploadedExtensions,
		})
	}

	return responses, nil
}

// UploadPhoto uploads a photo file
func (s *PhotoService) UploadPhoto(userID uint, localID, fileExtension, fileType string, fileData []byte) error {
	// Find photo record
	photo, err := s.photoRepo.FindByLocalID(localID)
	if err != nil {
		return fmt.Errorf("failed to find photo: %w", err)
	}
	if photo == nil {
		return fmt.Errorf("photo not found")
	}

	// Build full file path with extension
	fullPath := photo.FilePath + photo.FileName + "." + fileExtension

	// Save file (always overwrites if exists)
	if err := s.fileStorage.SaveFile(fullPath, fileData); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Set file timestamps using photo's creation time
	if err := s.fileStorage.SetFileTimes(fullPath, photo.CreationTime, photo.CreationTime); err != nil {
		return fmt.Errorf("failed to set file times: %w", err)
	}

	// Add extension to tracking list
	if err := s.photoRepo.AddUploadedExtension(localID, fileExtension); err != nil {
		return fmt.Errorf("failed to update extension list: %w", err)
	}

	// Update file count
	if err := s.photoRepo.UpdateFileCount(localID, 1); err != nil {
		return fmt.Errorf("failed to update file count: %w", err)
	}

	return nil
}

// UploadPhotoStream uploads a photo file using streaming (no memory buffering)
func (s *PhotoService) UploadPhotoStream(userID uint, localID, fileExtension, fileType string, reader io.Reader) error {
	// Find photo record
	photo, err := s.photoRepo.FindByLocalID(localID)
	if err != nil {
		return fmt.Errorf("failed to find photo: %w", err)
	}
	if photo == nil {
		return fmt.Errorf("photo not found")
	}

	// Build full file path with extension
	fullPath := photo.FilePath + photo.FileName + "." + fileExtension

	// Save file using streaming (always overwrites if exists)
	if err := s.fileStorage.SaveFileStream(fullPath, reader); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Set file timestamps using photo's creation time
	if err := s.fileStorage.SetFileTimes(fullPath, photo.CreationTime, photo.CreationTime); err != nil {
		return fmt.Errorf("failed to set file times: %w", err)
	}

	// Add extension to tracking list
	if err := s.photoRepo.AddUploadedExtension(localID, fileExtension); err != nil {
		return fmt.Errorf("failed to update extension list: %w", err)
	}

	// Update file count
	if err := s.photoRepo.UpdateFileCount(localID, 1); err != nil {
		return fmt.Errorf("failed to update file count: %w", err)
	}

	return nil
}
