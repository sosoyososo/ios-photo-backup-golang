# Upload API Contract

**Endpoint**: `POST /photos/upload`
**Feature**: 002-photo-extension-tracking

## Overview

Uploads a photo file to the server. Files are always overwritten if they already exist. The uploaded extension is tracked in the photo record.

## Request

### Headers
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

### Form Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `file` | file | Yes | The photo file to upload (max 50MB) |
| `local_id` | string | Yes | The local_id from index response |
| `file_type` | string | Yes | MIME type of the file |

### Example

```
POST /photos/upload
Authorization: Bearer eyJhbGciOi...

file: [binary data]
local_id: IMG_001
file_type: image/jpeg
```

## Response

### Success (200 OK)

```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_001",
  "filename": "IMG_0001.jpg",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.jpg"
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always "success" on success |
| `message` | string | Success message |
| `local_id` | string | Echo of request local_id |
| `filename` | string | Filename with extension |
| `file_path` | string | Full path to uploaded file |

### Error Responses

#### 400 Bad Request
```json
{
  "error": "Bad request",
  "message": "file is required"
}
```

#### 401 Unauthorized
```json
{
  "error": "Unauthorized",
  "message": "Token not found or invalid"
}
```

#### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "detailed error message"
}
```

## Behavior Changes

### Previous Behavior (Before Feature)

1. Check if file exists at path
2. If exists:
   - Check file size
   - If size matches expected size → Skip upload
   - If size doesn't match → Delete file and re-upload
3. Save file
4. Update file count

### New Behavior (With Feature)

1. **Always save file** (overwrites if exists)
2. **Update extension list** (adds to JSON array if not present)
3. **Update file count**

### Key Changes

- **Removed**: File existence check
- **Removed**: File size verification
- **Removed**: Incomplete file detection
- **Added**: Extension list tracking
- **Simplified**: Always overwrite logic

## Upload Flow

```
Client Upload Request
       ↓
Parse multipart form
       ↓
Validate local_id exists
       ↓
Extract file extension
       ↓
Read file data
       ↓
Save file (OVERWRITE if exists)  ← Always overwrite
       ↓
Update extension list in DB
       ↓
Update file count
       ↓
Return success
```

## Extension Tracking

### Extension List Storage

The uploaded extension is added to the `uploaded_extensions` JSON array in the photo record.

### Examples

#### First Upload (JPG)
**Before**: `[]`
**After**: `["jpg"]`

#### Second Upload (HEIC of same photo)
**Before**: `["jpg"]`
**After**: `["jpg", "heic"]`

#### Re-upload Same Extension
**Before**: `["jpg", "heic"]`
**After**: `["jpg", "heic"]` (no duplicate)

### Extension List Rules

- Extensions stored as JSON array of strings
- No duplicate extensions (re-upload doesn't add duplicate)
- Extensions persist across re-uploads
- Order is not guaranteed (use set-like behavior)

## Examples

### Example 1: First Upload

**Request**:
```
POST /photos/upload
file: [binary jpg data]
local_id: IMG_001
file_type: image/jpeg
```

**Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_001",
  "filename": "IMG_0001.jpg",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.jpg"
}
```

**Database State**:
- `uploaded_extensions`: `["jpg"]`
- `file_count`: 1

### Example 2: Upload Different Extension

**Request**:
```
POST /photos/upload
file: [binary heic data]
local_id: IMG_001
file_type: image/heic
```

**Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_001",
  "filename": "IMG_0001.heic",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.heic"
}
```

**Database State**:
- `uploaded_extensions`: `["jpg", "heic"]`
- `file_count`: 2

### Example 3: Re-upload Same Extension (Overwrite)

**Request**:
```
POST /photos/upload
file: [binary jpg data - updated version]
local_id: IMG_001
file_type: image/jpeg
```

**Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_001",
  "filename": "IMG_0001.jpg",
  "file_path": "storage/photo/1/2025/01/15/IMG_0001.jpg"
}
```

**Database State**:
- `uploaded_extensions`: `["jpg", "heic"]` (unchanged)
- `file_count`: 3 (incremented)

**File System**:
- Old `IMG_0001.jpg` is **completely replaced** with new version

## Error Handling

### File Size Limit
- Maximum file size: 50MB
- Exceeding limit returns 400 Bad Request

### Invalid local_id
- Non-existent local_id returns 500 Internal Server Error
- Photo must be indexed before upload

### Missing File Extension
- File must have an extension (determined from filename)
- Missing extension returns 400 Bad Request

### Storage Full
- If storage is full, upload fails
- Returns 500 Internal Server Error
- Extension list is NOT updated if upload fails

## Implementation Notes

### File Overwrite
- Uses standard file write operations
- Overwrites entire file content
- No atomic rename needed (overwrite is safe)

### Extension List Update
- Fetch current photo record
- Parse `uploaded_extensions` JSON
- Check if extension already exists
- Add if new, skip if duplicate
- Save updated record

### File Count
- Incremented on every upload
- Counts total uploads regardless of extension
- Not decremented on overwrite
