# Index API Contract

**Endpoint**: `POST /photos/index`
**Feature**: 002-photo-extension-tracking

## Overview

Indexes a batch of photos and assigns sequential filenames. Returns the list of assigned files with uploaded extensions for each photo.

## Request

### Headers
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Body

```json
{
  "date": "2025-01-15",
  "photos": [
    {
      "local_id": "unique_identifier_1",
      "creation_time": "2025-01-15T10:30:00Z",
      "file_extension": "jpg",
      "file_type": "image/jpeg"
    },
    {
      "local_id": "unique_identifier_2",
      "creation_time": "2025-01-15T10:31:00Z",
      "file_extension": "heic",
      "file_type": "image/heic"
    }
  ]
}
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `date` | string | Yes | Date in YYYY-MM-DD format |
| `photos` | array | Yes | List of photos to index |
| `photos[].local_id` | string | Yes | Client-side unique identifier |
| `photos[].creation_time` | string | Yes | ISO 8601 timestamp |
| `photos[].file_extension` | string | Yes | File extension (jpg, png, heic, etc.) |
| `photos[].file_type` | string | Yes | MIME type |

## Response

### Success (200 OK)

```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "unique_identifier_1",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": []
    },
    {
      "local_id": "unique_identifier_2",
      "filename": "IMG_0002.heic",
      "uploaded_extensions": []
    }
  ]
}
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Always "success" on success |
| `date` | string | Echo of request date |
| `assigned_files` | array | List of indexed photos |
| `assigned_files[].local_id` | string | Echo of request local_id |
| `assigned_files[].filename` | string | Assigned filename with extension |
| `assigned_files[].uploaded_extensions` | array | **NEW**: List of uploaded extensions for this photo |

### uploaded_extensions Values

- `[]` - No files uploaded yet
- `["jpg"]` - Only JPEG uploaded
- `["heic","jpg"]` - Both HEIC and JPEG uploaded
- `["png"]` - Only PNG uploaded

### Error Responses

#### 400 Bad Request
```json
{
  "error": "Invalid request format",
  "details": "specific validation error"
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

## Changes from Previous Version

### Added Fields
- `assigned_files[].uploaded_extensions` - Array of uploaded file extensions for each photo

### Behavior Changes
- Response now includes upload status for each photo
- Users can see which file formats have been backed up
- Helps identify missing uploads without additional API calls

## Examples

### Example 1: Fresh Index (No Uploads)

**Request**:
```json
{
  "date": "2025-01-15",
  "photos": [
    {
      "local_id": "IMG_001",
      "creation_time": "2025-01-15T10:30:00Z",
      "file_extension": "jpg",
      "file_type": "image/jpeg"
    }
  ]
}
```

**Response**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "IMG_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": []
    }
  ]
}
```

### Example 2: Re-index with Uploaded Files

**Request**:
```json
{
  "date": "2025-01-15",
  "photos": [
    {
      "local_id": "IMG_001",
      "creation_time": "2025-01-15T10:30:00Z",
      "file_extension": "jpg",
      "file_type": "image/jpeg"
    }
  ]
}
```

**Response** (assuming JPG already uploaded):
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "IMG_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["jpg"]
    }
  ]
}
```

### Example 3: Multiple Formats

**Request**:
```json
{
  "date": "2025-01-15",
  "photos": [
    {
      "local_id": "IMG_001",
      "creation_time": "2025-01-15T10:30:00Z",
      "file_extension": "jpg",
      "file_type": "image/jpeg"
    }
  ]
}
```

**Response** (assuming both HEIC and JPG uploaded):
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "IMG_001",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["heic", "jpg"]
    }
  ]
}
```

## Notes

- Extension tracking is independent of the indexed extension
- A photo can have uploaded files with different extensions than what was indexed
- Empty array `[]` indicates no uploads yet
- Extensions are tracked in order of upload, but order is not guaranteed
