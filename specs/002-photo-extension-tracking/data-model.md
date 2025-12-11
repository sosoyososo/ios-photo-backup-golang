# Data Model: Photo Extension Tracking

**Feature**: 002-photo-extension-tracking
**Date**: 2025-12-11

## Schema Changes

### Photo Table Modification

**Table**: `photos_user_{user_id}` (dynamic per user)

**New Column**: `uploaded_extensions`

| Column | Type | Constraints | Default | Description |
|--------|------|-------------|---------|-------------|
| `local_id` | TEXT | PRIMARY KEY | - | Client-side unique identifier |
| `creation_time` | DATETIME | NOT NULL, INDEX | - | When photo was taken |
| `file_path` | TEXT | NOT NULL, INDEX | - | Directory path |
| `file_name` | TEXT | NOT NULL | - | Filename (without extension) |
| `file_type` | TEXT | NOT NULL | - | MIME type |
| `file_count` | INTEGER | DEFAULT 0 | 0 | Number of files uploaded |
| `uploaded_extensions` | TEXT | - | `'[]'` | **NEW**: JSON array of uploaded extensions |
| `created_at` | DATETIME | - | - | Record creation time |
| `updated_at` | DATETIME | - | - | Record update time |
| `deleted_at` | DATETIME | INDEX | - | Soft delete support |

### Field Details

#### uploaded_extensions
- **Type**: TEXT (JSON)
- **Format**: `["jpg","heic","png"]` - JSON array of strings
- **Default**: `'[]'` (empty JSON array)
- **Purpose**: Tracks which file extensions have been uploaded for this photo
- **Example**:
  - `[]` - No files uploaded
  - `["jpg"]` - Only JPEG uploaded
  - `["heic","jpg"]` - Both HEIC and JPEG uploaded

## State Transitions

### Photo Lifecycle

```
Indexed → [Upload JPG] → JPG Uploaded → [Upload HEIC] → Multiple Formats
   ↓              ↓                ↓                   ↓
No files    Extension: ["jpg"]  Extension: ["jpg"]  Extension: ["heic","jpg"]
```

### Extension List Updates

```
Initial:     []
After JPG:   ["jpg"]
After HEIC:  ["jpg", "heic"]
After ReJPG: ["jpg", "heic"]  (duplicates prevented)
```

## Relationships

```
User (1) ──→ (N) Photo Records
Photo (1) ──→ (N) Uploaded Extensions (stored as JSON array)
```

## Validation Rules

### Photo Record
- `local_id`: Must be unique per user
- `creation_time`: Must be valid datetime
- `file_path`: Must be valid directory path
- `file_name`: Non-empty, follows IMG_XXXX pattern
- `file_type`: Valid MIME type
- `file_count`: ≥ 0
- `uploaded_extensions`: Must be valid JSON array of strings
  - Each extension: lowercase alphanumeric (jpg, png, heic, etc.)
  - No duplicates allowed in array
  - Empty array `[]` is valid (no files uploaded)

### Extension Management
- Extensions added only after successful file upload
- Same extension can be re-uploaded (file overwritten)
- Extension list persists across re-uploads
- No removal of extensions (only additions)

## Example Records

### Record 1: Photo with no uploads
```json
{
  "local_id": "IMG_001",
  "creation_time": "2025-01-15T10:30:00Z",
  "file_path": "storage/photo/1/2025/01/15/",
  "file_name": "IMG_0001",
  "file_type": "image/jpeg",
  "file_count": 0,
  "uploaded_extensions": "[]",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

### Record 2: Photo with multiple formats uploaded
```json
{
  "local_id": "IMG_002",
  "creation_time": "2025-01-15T10:31:00Z",
  "file_path": "storage/photo/1/2025/01/15/",
  "file_name": "IMG_0002",
  "file_type": "image/heic",
  "file_count": 2,
  "uploaded_extensions": "[\"heic\",\"jpg\"]",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:35:00Z"
}
```

### Record 3: Photo with re-uploaded file
```json
{
  "local_id": "IMG_003",
  "creation_time": "2025-01-15T10:32:00Z",
  "file_path": "storage/photo/1/2025/01/15/",
  "file_name": "IMG_0003",
  "file_type": "image/jpeg",
  "file_count": 3,
  "uploaded_extensions": "[\"jpg\"]",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:40:00Z"
}
```

## Migration Strategy

### Step 1: Update Photo Model
Add `UploadedExtensions` field to `internal/models/photo.go`:
```go
type Photo struct {
    LocalID            string         `json:"local_id" gorm:"primaryKey"`
    CreationTime       time.Time      `json:"creation_time" gorm:"not null;index"`
    FilePath           string         `json:"file_path" gorm:"not null;index"`
    FileName           string         `json:"file_name" gorm:"not null;size:255"`
    FileType           string         `json:"file_type" gorm:"not null;size:50"`
    FileCount          int            `json:"file_count" gorm:"default:0"`
    UploadedExtensions string         `json:"uploaded_extensions" gorm:"type:text;default:'[]'"`
    CreatedAt          time.Time      `json:"created_at"`
    UpdatedAt          time.Time      `json:"updated_at"`
    DeletedAt          gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### Step 2: AutoMigrate
Server restart triggers GORM AutoMigrate:
- Detects new field
- Adds `uploaded_extensions` column to existing tables
- Sets default value `'[]'` for all records

### Step 3: Backward Compatibility
- Existing records get default `'[]'` value
- New uploads populate the field
- No data migration needed

## Indexes

**Existing Indexes**:
- PRIMARY KEY: `local_id`
- INDEX: `creation_time`
- INDEX: `file_path`
- INDEX: `deleted_at`

**No New Indexes Needed**:
- `uploaded_extensions` is JSON text, not suitable for indexing
- Can query with raw SQL if needed: `JSON_EXTRACT(uploaded_extensions, '$')`
- Default query is by `local_id` or `file_path` (date), already indexed
