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

// SaveFileStream saves a file from an io.Reader (streaming)
func (fs *FileStorage) SaveFileStream(filePath string, reader io.Reader) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := config.EnsureDir(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Stream copy
	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("failed to stream file: %w", err)
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

// GetChunkDir returns the temporary directory for storing chunks
func (fs *FileStorage) GetChunkDir(filePath string) string {
	return filePath + ".chunks"
}

// GetChunkPath returns the path for a specific chunk
func (fs *FileStorage) GetChunkPath(filePath string, chunkNumber int) string {
	chunkDir := fs.GetChunkDir(filePath)
	return filepath.Join(chunkDir, fmt.Sprintf("chunk_%03d", chunkNumber))
}

// SaveChunk saves a single chunk to the chunk directory
func (fs *FileStorage) SaveChunk(filePath string, chunkNumber int, data []byte) error {
	chunkDir := fs.GetChunkDir(filePath)

	// Ensure chunk directory exists
	if err := config.EnsureDir(chunkDir, 0755); err != nil {
		return fmt.Errorf("failed to create chunk directory: %w", err)
	}

	chunkPath := fs.GetChunkPath(filePath, chunkNumber)

	// Write chunk file
	if err := os.WriteFile(chunkPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chunk %d: %w", chunkNumber, err)
	}

	return nil
}

// MergeChunks merges all chunks into the final file
func (fs *FileStorage) MergeChunks(filePath string, totalChunks int) error {
	// Ensure destination directory exists
	dir := filepath.Dir(filePath)
	if err := config.EnsureDir(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Merge chunks in order
	for i := 0; i < totalChunks; i++ {
		chunkPath := fs.GetChunkPath(filePath, i)

		chunkData, err := os.ReadFile(chunkPath)
		if err != nil {
			return fmt.Errorf("failed to read chunk %d: %w", i, err)
		}

		if _, err := dst.Write(chunkData); err != nil {
			return fmt.Errorf("failed to write chunk %d: %w", i, err)
		}
	}

	return nil
}

// CleanupChunks removes all chunk files for a given file path
func (fs *FileStorage) CleanupChunks(filePath string) error {
	chunkDir := fs.GetChunkDir(filePath)

	if err := os.RemoveAll(chunkDir); err != nil {
		return fmt.Errorf("failed to cleanup chunks: %w", err)
	}

	return nil
}

// GetUploadedChunks returns the list of uploaded chunk numbers
func (fs *FileStorage) GetUploadedChunks(filePath string) ([]int, error) {
	chunkDir := fs.GetChunkDir(filePath)

	entries, err := os.ReadDir(chunkDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, fmt.Errorf("failed to read chunk directory: %w", err)
	}

	var chunks []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var chunkNum int
		if _, err := fmt.Sscanf(entry.Name(), "chunk_%d", &chunkNum); err == nil {
			chunks = append(chunks, chunkNum)
		}
	}

	return chunks, nil
}
