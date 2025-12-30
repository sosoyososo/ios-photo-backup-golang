package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ios-photo-backup/photo-backup-server/internal/config"
)

// FileStorage handles file storage operations
type FileStorage struct {
	config *config.Config
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(cfg *config.Config) *FileStorage {
	return &FileStorage{
		config: cfg,
	}
}

// SaveFile saves a file to the specified path
func (fs *FileStorage) SaveFile(filePath string, data []byte) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := config.EnsureDir(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ReadFile reads a file from the specified path
func (fs *FileStorage) ReadFile(filePath string) ([]byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return data, nil
}

// FileExists checks if a file exists
func (fs *FileStorage) FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("failed to check file: %w", err)
}

// DeleteFile deletes a file
func (fs *FileStorage) DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, that's okay
			return nil
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileSize returns the size of a file
func (fs *FileStorage) GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	return info.Size(), nil
}

// SetFileTimes sets the access and modification times of a file
func (fs *FileStorage) SetFileTimes(filePath string, atime, mtime time.Time) error {
	if err := os.Chtimes(filePath, atime, mtime); err != nil {
		return fmt.Errorf("failed to set file times: %w", err)
	}
	return nil
}

// CopyFile copies a file from src to dst
func (fs *FileStorage) CopyFile(src, dst string) error {
	// Ensure destination directory exists
	if err := config.EnsureDir(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}
