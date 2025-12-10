package service

import (
	"fmt"
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
	FileExtension string    `json:"file_extension"`
	FileType      string    `json:"file_type"`
}

// PhotoIndexResponse represents a photo indexing response
type PhotoIndexResponse struct {
	LocalID  string `json:"local_id"`
	Filename string `json:"filename"`
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

		if existingPhoto != nil {
			// Photo already exists, keep existing filename
			responses = append(responses, PhotoIndexResponse{
				LocalID:  photo.LocalID,
				Filename: existingPhoto.FileName,
			})
			continue
		}

		// Generate filename
		filename := s.naming.GenerateFilename(photo.FileExtension, nextSequence)
		nextSequence++

		// Create photo record
		photoRecord := &models.Photo{
			LocalID:       photo.LocalID,
			CreationTime:  photo.CreationTime,
			FilePath:      dirPath,
			FileName:      filename,
			FileExtension: photo.FileExtension,
			FileType:      photo.FileType,
			FileCount:     0,
		}

		// Save to database
		if err := s.photoRepo.Create(photoRecord); err != nil {
			return nil, fmt.Errorf("failed to create photo record: %w", err)
		}

		responses = append(responses, PhotoIndexResponse{
			LocalID:  photo.LocalID,
			Filename: filename,
		})
	}

	return responses, nil
}

// UploadPhoto uploads a photo file
func (s *PhotoService) UploadPhoto(userID uint, localID, fileType string, fileData []byte) error {
	// Find photo record
	photo, err := s.photoRepo.FindByLocalID(localID)
	if err != nil {
		return fmt.Errorf("failed to find photo: %w", err)
	}
	if photo == nil {
		return fmt.Errorf("photo not found")
	}

	// Build full file path
	fullPath := photo.FilePath + photo.FileName

	// Check if file already exists
	exists, err := s.fileStorage.FileExists(fullPath)
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	if exists {
		// File exists, check if complete (placeholder - actual implementation would verify integrity)
		size, err := s.fileStorage.GetFileSize(fullPath)
		if err != nil {
			return fmt.Errorf("failed to get file size: %w", err)
		}

		// If file seems complete, skip upload
		if size > 0 && size == int64(len(fileData)) {
			return nil
		}

		// File is incomplete, delete and re-upload
		if err := s.fileStorage.DeleteFile(fullPath); err != nil {
			return fmt.Errorf("failed to delete incomplete file: %w", err)
		}
	}

	// Save file
	if err := s.fileStorage.SaveFile(fullPath, fileData); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	// Update file count
	if err := s.photoRepo.UpdateFileCount(localID, 1); err != nil {
		return fmt.Errorf("failed to update file count: %w", err)
	}

	return nil
}
