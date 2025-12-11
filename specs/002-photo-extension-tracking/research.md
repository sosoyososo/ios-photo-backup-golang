# Research: Photo Extension Tracking Implementation

**Date**: 2025-12-11
**Feature**: 002-photo-extension-tracking

## Research Questions

### 1. How to store extension list in SQLite/GORM?

**Decision**: Store as JSON text field in SQLite

**Rationale**:
- SQLite doesn't have native array type
- GORM supports JSON fields using `gorm:"type:text"` with `json` tag
- JSON field allows flexible number of extensions per photo
- Simpler than creating a separate table for extensions
- Querying for specific extensions can use JSON_EXTRACT or GORM raw queries

**Implementation approach**:
```go
type Photo struct {
    // ... existing fields
    UploadedExtensions string `json:"uploaded_extensions" gorm:"type:text;default:'[]'"`
}
```

**Storage format**: `["jpg","heic","png"]` - JSON array of strings

---

### 2. Database migration strategy for existing data?

**Decision**: Use GORM AutoMigrate with default value handling

**Rationale**:
- Existing photo records need backward compatibility
- New column should default to empty JSON array `[]`
- GORM AutoMigrate will add column to existing tables
- No data migration needed - just add column with default

**Migration approach**:
- Add `UploadedExtensions` field to Photo model
- Set default: `gorm:"default:'[]'"` for empty array
- AutoMigrate will add column on next server start
- Existing records get default empty array value

---

### 3. How to update extension list during upload?

**Decision**: Append to JSON array, avoiding duplicates

**Rationale**:
- Need to add extension to list when file uploaded
- Must handle duplicates (same extension re-uploaded)
- GORM doesn't have built-in JSON array operations
- Use raw SQL or unmarshal/modify/remarshal

**Implementation approach**:
```go
// Option 1: Using raw SQL with JSON functions
db.Exec("UPDATE photos_user_? SET uploaded_extensions = ? WHERE local_id = ?",
    userID, JSON_ARRAY_APPEND(uploadedExtensions, '$', newExtension), localID)

// Option 2: Unmarshal/modify/remarshal (simpler)
var photo Photo
db.Where("local_id = ?", localID).First(&photo)
extensions := parseExtensions(photo.UploadedExtensions)
if !contains(extensions, newExtension) {
    extensions = append(extensions, newExtension)
}
photo.UploadedExtensions = marshalExtensions(extensions)
db.Save(&photo)
```

**Recommendation**: Use Option 2 (unmarshal/modify/remarshal) for simplicity and readability

---

### 4. API response format for Index API?

**Decision**: Add `uploaded_extensions` field to each photo in response

**Rationale**:
- Index API already returns list of photos
- Need to include which extensions are uploaded
- Must handle null/empty cases gracefully
- Client can verify upload status

**Response format**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "IMG_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["jpg", "heic"]
    }
  ]
}
```

**Empty case**: `[]` (empty array) when no files uploaded

---

### 5. Simplify Upload API - remove existence check?

**Decision**: Remove file existence check, always save file

**Rationale**:
- Upload API currently checks if file exists
- If exists and complete, skips upload
- If exists and incomplete, deletes and re-uploads
- New requirement: always overwrite regardless

**Current flow**:
1. Check if file exists
2. If exists, check size/completeness
3. If incomplete, delete and re-upload
4. Save file

**New flow**:
1. Always save file (overwrites if exists)
2. Update extension list
3. Update file count

**Benefits**:
- Simpler logic
- Ensures latest version always stored
- Faster uploads (no existence check)

---

## Implementation Summary

### Database Changes
- Add `UploadedExtensions` field to Photo model (JSON text, default `[]`)
- GORM AutoMigrate handles existing tables

### API Changes
- **Index API**: Add `uploaded_extensions` field to response
- **Upload API**: Remove file existence check, always overwrite

### Data Flow
1. **Index**: Returns extensions list with each photo
2. **Upload**: Saves file → Updates extensions list → Updates file count

### Migration Path
1. Add field to Photo model
2. Restart server (AutoMigrate creates column)
3. Existing records get default `[]`
4. New uploads populate field
