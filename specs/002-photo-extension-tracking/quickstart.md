# Quick Start Guide: Photo Extension Tracking

**Feature**: 002-photo-extension-tracking
**Date**: 2025-12-11

## Overview

This guide walks through implementing the Photo Extension Tracking feature, which adds support for tracking uploaded file extensions and simplifies the upload process.

## Prerequisites

- Go (latest stable)
- Existing photo backup server codebase
- SQLite database
- GORM (Go ORM)

## Implementation Steps

### Step 1: Update Photo Model

**File**: `internal/models/photo.go`

Add the `UploadedExtensions` field to the Photo struct:

```go
type Photo struct {
    LocalID            string         `json:"local_id" gorm:"primaryKey"`
    CreationTime       time.Time      `json:"creation_time" gorm:"not null;index"`
    FilePath           string         `json:"file_path" gorm:"not null;index"`
    FileName           string         `json:"file_name" gorm:"not null;size:255"`
    FileType           string         `json:"file_type" gorm:"not null;size:50"`
    FileCount          int            `json:"file_count" gorm:"default:0"`
    UploadedExtensions string         `json:"uploaded_extensions" gorm:"type:text;default:'[]'"` // NEW
    CreatedAt          time.Time      `json:"created_at"`
    UpdatedAt          time.Time      `json:"updated_at"`
    DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### Step 2: Update Photo Repository

**File**: `internal/repository/photo.go`

Add helper methods to manage extension list:

```go
// GetUploadedExtensions returns the list of uploaded extensions
func (r *PhotoRepository) GetUploadedExtensions(localID string) ([]string, error) {
    if err := r.ensureTableExists(); err != nil {
        return nil, err
    }
    var photo models.Photo
    if err := r.db.Table(r.tableName).Where("local_id = ?", localID).First(&photo).Error; err != nil {
        return nil, err
    }

    // Parse JSON array
    extensions := make([]string, 0)
    if photo.UploadedExtensions != "" && photo.UploadedExtensions != "[]" {
        if err := json.Unmarshal([]byte(photo.UploadedExtensions), &extensions); err != nil {
            return nil, fmt.Errorf("failed to parse extensions: %w", err)
        }
    }
    return extensions, nil
}

// AddUploadedExtension adds an extension to the list if not exists
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
            return nil // Already exists, no update needed
        }
    }

    // Add new extension
    extensions = append(extensions, extension)

    // Update database
    extensionsJSON, err := json.Marshal(extensions)
    if err != nil {
        return fmt.Errorf("failed to marshal extensions: %w", err)
    }

    if err := r.db.Table(r.tableName).Where("local_id = ?", localID).Update("uploaded_extensions", string(extensionsJSON)).Error; err != nil {
        return fmt.Errorf("failed to update extensions: %w", err)
    }

    return nil
}
```

Don't forget to add the import:
```go
import "encoding/json"
```

### Step 3: Update Photo Service

**File**: `internal/service/photo.go`

Modify the `UploadPhoto` method to:
1. Remove file existence check
2. Always save file (overwrite)
3. Add extension to tracking list

```go
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

    // NEW: Always save file (overwrites if exists)
    if err := s.fileStorage.SaveFile(fullPath, fileData); err != nil {
        return fmt.Errorf("failed to save file: %w", err)
    }

    // NEW: Add extension to tracking list
    if err := s.photoRepo.AddUploadedExtension(localID, fileExtension); err != nil {
        return fmt.Errorf("failed to update extension list: %w", err)
    }

    // Update file count
    if err := s.photoRepo.UpdateFileCount(localID, 1); err != nil {
        return fmt.Errorf("failed to update file count: %w", err)
    }

    return nil
}
```

Also update the `IndexPhotos` method to include extension list in response:

```go
// In IndexPhotos, when creating response:
for _, photo := range sortedPhotos {
    // ... existing code ...

    // NEW: Get uploaded extensions for response
    extensions, err := s.photoRepo.GetUploadedExtensions(photo.LocalID)
    if err != nil {
        extensions = []string{} // Empty on error
    }

    // Response includes extension for client to use
    responses = append(responses, PhotoIndexResponse{
        LocalID:             photo.LocalID,
        Filename:            filename + "." + photo.FileExtension,
        UploadedExtensions:  extensions, // NEW
    })
}
```

Update the `PhotoIndexResponse` struct to include the new field:

```go
type PhotoIndexResponse struct {
    LocalID             string   `json:"local_id"`
    Filename            string   `json:"filename"`
    UploadedExtensions  []string `json:"uploaded_extensions"` // NEW
}
```

### Step 4: Update Index API Handler

**File**: `internal/api/handlers/photo/index.go`

The handler already uses `PhotoIndexResponse`, so the field will be automatically included in the JSON response. No changes needed to the handler itself.

### Step 5: Restart Server (Database Migration)

Restart the server to trigger GORM AutoMigrate:

```bash
go run cmd/server/main.go
```

GORM will automatically:
- Detect the new `UploadedExtensions` field
- Add the column to existing `photos_user_*` tables
- Set default value `'[]'` for all existing records

### Step 6: Test Implementation

#### Test 1: Index Photos (No Uploads)

```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-15",
    "photos": [
      {
        "local_id": "TEST_001",
        "creation_time": "2025-01-15T10:30:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

**Expected Response**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "TEST_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": []
    }
  ]
}
```

#### Test 2: Upload File

```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/test.jpg" \
  -F "local_id=TEST_001" \
  -F "file_type=image/jpeg"
```

**Expected Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "TEST_001",
  "filename": "IMG_0001.jpg",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.jpg"
}
```

#### Test 3: Re-index (Verify Extension Tracking)

```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-15",
    "photos": [
      {
        "local_id": "TEST_001",
        "creation_time": "2025-01-15T10:30:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

**Expected Response**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "TEST_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["jpg"]
    }
  ]
}
```

#### Test 4: Upload Different Extension

```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/test.heic" \
  -F "local_id=TEST_001" \
  -F "file_type=image/heic"
```

**Expected Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "TEST_001",
  "filename": "IMG_0001.heic",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.heic"
}
```

#### Test 5: Re-index (Verify Both Extensions)

```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-15",
    "photos": [
      {
        "local_id": "TEST_001",
        "creation_time": "2025-01-15T10:30:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

**Expected Response**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "TEST_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["jpg", "heic"]
    }
  ]
}
```

## Verification Checklist

- [ ] Photo model updated with `UploadedExtensions` field
- [ ] Repository has methods to get and add extensions
- [ ] Photo service `UploadPhoto` always overwrites files
- [ ] Photo service adds extensions to tracking list
- [ ] Index API response includes `uploaded_extensions` field
- [ ] Server starts without errors (AutoMigrate successful)
- [ ] Test 1: Index returns empty extension list
- [ ] Test 2: Upload works correctly
- [ ] Test 3: Re-index shows uploaded extension
- [ ] Test 4: Different extension upload works
- [ ] Test 5: Re-index shows both extensions

## Rollback Plan

If issues occur, rollback is simple:

1. Remove `UploadedExtensions` field from Photo model
2. Remove extension-related repository methods
3. Restore previous `UploadPhoto` implementation (with existence check)
4. Restart server
5. GORM AutoMigrate will drop the column

## Performance Considerations

- JSON field parsing adds minimal overhead
- Extension list is small (typically 1-5 extensions per photo)
- No new database indexes needed
- File overwrite is O(1) operation

## Security Considerations

- Extension validation: Only allow valid image extensions
- File type validation: Verify MIME type matches extension
- File size limits: Enforce 50MB limit (existing)
- Path traversal: Prevent (already handled by existing code)
