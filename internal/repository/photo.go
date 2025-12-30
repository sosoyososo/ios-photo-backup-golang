package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/ios-photo-backup/photo-backup-server/internal/models"
)

// PhotoRepository provides CRUD operations for photos
// Uses dynamic table names based on user ID
type PhotoRepository struct {
	db      *gorm.DB
	userID  uint
	tableName string
}

// NewPhotoRepository creates a new PhotoRepository for a specific user
func NewPhotoRepository(db *gorm.DB, userID uint) *PhotoRepository {
	return &PhotoRepository{
		db:      db,
		userID:  userID,
		tableName: fmt.Sprintf("photos_user_%d", userID),
	}
}

// ensureTableExists creates the photo table if it doesn't exist and migrates schema if needed
func (r *PhotoRepository) ensureTableExists() error {
	// Check if table already exists
	if r.db.Migrator().HasTable(r.tableName) {
		return nil
	}

	// Create the table with the custom name
	if err := r.db.Table(r.tableName).AutoMigrate(&models.Photo{}); err != nil {
		return fmt.Errorf("failed to migrate photo table: %w", err)
	}
	return nil
}

// Create creates a new photo record
func (r *PhotoRepository) Create(photo *models.Photo) error {
	if err := r.ensureTableExists(); err != nil {
		return err
	}
	if err := r.db.Table(r.tableName).Create(photo).Error; err != nil {
		return fmt.Errorf("failed to create photo: %w", err)
	}
	return nil
}

// CreateBatch creates multiple photo records
func (r *PhotoRepository) CreateBatch(photos []models.Photo) error {
	if err := r.ensureTableExists(); err != nil {
		return err
	}
	if err := r.db.Table(r.tableName).CreateInBatches(photos, 100).Error; err != nil {
		return fmt.Errorf("failed to create photo batch: %w", err)
	}
	return nil
}

// FindByLocalID finds a photo by local_id
func (r *PhotoRepository) FindByLocalID(localID string) (*models.Photo, error) {
	if err := r.ensureTableExists(); err != nil {
		return nil, err
	}
	var photo models.Photo
	if err := r.db.Table(r.tableName).Where("local_id = ?", localID).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find photo: %w", err)
	}
	return &photo, nil
}

// GetByDate gets all photos for a specific date
func (r *PhotoRepository) GetByDate(date string) ([]models.Photo, error) {
	if err := r.ensureTableExists(); err != nil {
		return nil, err
	}
	var photos []models.Photo
	if err := r.db.Table(r.tableName).Where("file_path LIKE ?", "%"+date+"%").Find(&photos).Error; err != nil {
		return nil, fmt.Errorf("failed to get photos by date: %w", err)
	}
	return photos, nil
}

// GetCountByDate gets the count of photos for a specific date
func (r *PhotoRepository) GetCountByDate(date string) (int, error) {
	if err := r.ensureTableExists(); err != nil {
		return 0, err
	}
	// Convert date format from YYYY-MM-DD to YYYY/MM/DD to match file_path structure
	dateForPath := strings.ReplaceAll(date, "-", "/")
	var count int64
	if err := r.db.Table(r.tableName).Where("file_path LIKE ?", "%"+dateForPath+"%").Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count photos by date: %w", err)
	}
	return int(count), nil
}

// Update updates a photo record
func (r *PhotoRepository) Update(photo *models.Photo) error {
	if err := r.ensureTableExists(); err != nil {
		return err
	}
	if err := r.db.Table(r.tableName).Save(photo).Error; err != nil {
		return fmt.Errorf("failed to update photo: %w", err)
	}
	return nil
}

// UpdateFileCount updates the file count for a photo
func (r *PhotoRepository) UpdateFileCount(localID string, count int) error {
	if err := r.ensureTableExists(); err != nil {
		return err
	}
	if err := r.db.Table(r.tableName).Where("local_id = ?", localID).Update("file_count", count).Error; err != nil {
		return fmt.Errorf("failed to update file count: %w", err)
	}
	return nil
}

// GetLocalIDs gets all local_ids for a date
func (r *PhotoRepository) GetLocalIDs(date string) ([]string, error) {
	if err := r.ensureTableExists(); err != nil {
		return nil, err
	}
	var localIDs []string
	if err := r.db.Table(r.tableName).Where("file_path LIKE ?", "%"+date+"%").Pluck("local_id", &localIDs).Error; err != nil {
		return nil, fmt.Errorf("failed to get local IDs: %w", err)
	}
	return localIDs, nil
}

// GetUploadedExtensions returns the list of uploaded extensions for a photo
func (r *PhotoRepository) GetUploadedExtensions(localID string) ([]string, error) {
	if err := r.ensureTableExists(); err != nil {
		return nil, err
	}
	var photo models.Photo
	if err := r.db.Table(r.tableName).Where("local_id = ?", localID).First(&photo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get photo: %w", err)
	}

	// Parse JSON array from uploaded_extensions field
	extensions := make([]string, 0)
	if photo.UploadedExtensions != "" && photo.UploadedExtensions != "[]" {
		if err := json.Unmarshal([]byte(photo.UploadedExtensions), &extensions); err != nil {
			return nil, fmt.Errorf("failed to parse uploaded extensions: %w", err)
		}
	}
	return extensions, nil
}

// AddUploadedExtension adds an extension to the uploaded extensions list if not already present
func (r *PhotoRepository) AddUploadedExtension(localID string, extension string) error {
	if err := r.ensureTableExists(); err != nil {
		return err
	}

	// Get current extensions
	extensions, err := r.GetUploadedExtensions(localID)
	if err != nil {
		return err
	}

	// Check if extension already exists
	for _, ext := range extensions {
		if ext == extension {
			// Extension already exists, no need to update
			return nil
		}
	}

	// Add new extension
	extensions = append(extensions, extension)

	// Marshal back to JSON
	extensionsJSON, err := json.Marshal(extensions)
	if err != nil {
		return fmt.Errorf("failed to marshal extensions: %w", err)
	}

	// Update database
	if err := r.db.Table(r.tableName).Where("local_id = ?", localID).Update("uploaded_extensions", string(extensionsJSON)).Error; err != nil {
		return fmt.Errorf("failed to update uploaded extensions: %w", err)
	}

	return nil
}

// GetAll retrieves all photo records for the user
func (r *PhotoRepository) GetAll() ([]models.Photo, error) {
	if err := r.ensureTableExists(); err != nil {
		return nil, err
	}
	var photos []models.Photo
	if err := r.db.Table(r.tableName).Find(&photos).Error; err != nil {
		return nil, fmt.Errorf("failed to get all photos: %w", err)
	}
	return photos, nil
}
