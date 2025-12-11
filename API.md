# Photo Backup Server - API Documentation

**Version**: 1.0.0
**Base URL**: `http://localhost:8080`

## Table of Contents

1. [Authentication](#authentication)
2. [Health Check](#health-check)
3. [User Authentication Endpoints](#user-authentication-endpoints)
4. [Token Management Endpoints](#token-management-endpoints)
5. [Photo Management Endpoints](#photo-management-endpoints)
6. [Error Handling](#error-handling)
7. [Rate Limiting](#rate-limiting)

---

## Authentication

The Photo Backup Server uses JWT (JSON Web Tokens) for authentication.

### Token Format
```
Authorization: Bearer <jwt_token>
```

### Token Claims
```json
{
  "user_id": 1,
  "username": "admin",
  "exp": 1765374491,
  "iat": 1764774491
}
```

- **user_id**: Unique user identifier
- **username**: User's username
- **exp**: Expiration timestamp (Unix epoch)
- **iat**: Issued at timestamp (Unix epoch)

---

## Health Check

### GET /health

Check if the server is running and healthy.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok",
  "timestamp": 1765374491
}
```

**Status Codes**:
- `200` - Server is healthy

---

## User Authentication Endpoints

### POST /login

Authenticate user and receive JWT token.

**Endpoint**: `POST /login`

**Headers**:
```
Content-Type: application/json
```

**Request Body**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "SecurePassword123"
  }'
```

**Success Response** (`200`):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T21:34:47+08:00"
}
```

**Error Response** (`401`):
```json
{
  "error": "unauthorized",
  "message": "invalid credentials"
}
```

**Error Response** (`400`):
```json
{
  "error": "bad_request",
  "message": "Invalid request format"
}
```

---

## Token Management Endpoints

### POST /refresh

Refresh an existing JWT token. Invalidates the old token and issues a new one.

**Endpoint**: `POST /refresh`

**Headers**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body**:
```json
{}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/refresh \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{}'
```

**Success Response** (`200`):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T21:34:47+08:00"
}
```

**Error Response** (`401`):
```json
{
  "error": "unauthorized",
  "message": "token not found or expired"
}
```

### GET /status

Check authentication status and user information.

**Endpoint**: `GET /status`

**Headers**:
```
Authorization: Bearer <jwt_token>
```

**Example Request**:
```bash
curl -X GET http://localhost:8080/status \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**Success Response** (`200`):
```json
{
  "status": "online",
  "user_id": 1,
  "username": "admin"
}
```

**Error Response** (`401`):
```json
{
  "error": "unauthorized",
  "message": "token not found or expired"
}
```

---

## Photo Management Endpoints

All photo endpoints require authentication via Bearer token.

### POST /photos/index

Index a batch of photos and assign sequential filenames.

**Endpoint**: `POST /photos/index`

**Headers**:
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "date": "string (required, format: YYYY-MM-DD)",
  "photos": [
    {
      "local_id": "string (required, unique identifier)",
      "creation_time": "string (required, ISO 8601 format)",
      "file_extension": "string (required, e.g., 'jpg', 'png', 'heic')",
      "file_type": "string (required, e.g., 'image/jpeg')"
    }
  ]
}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-12-10",
    "photos": [
      {
        "local_id": "IMG_1234",
        "creation_time": "2025-12-10T10:00:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      },
      {
        "local_id": "IMG_1235",
        "creation_time": "2025-12-10T10:05:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

**Success Response** (`200`):
```json
{
  "status": "success",
  "date": "2025-12-10",
  "assigned_files": [
    {
      "local_id": "IMG_1234",
      "filename": "IMG_0001.jpg",
      "uploaded_extensions": ["jpg"]
    },
    {
      "local_id": "IMG_1235",
      "filename": "IMG_0002.jpg",
      "uploaded_extensions": []
    }
  ]
}
```

**Response Fields**:
- `uploaded_extensions` (array): List of file extensions that have been uploaded for this photo
  - `[]` - No files uploaded yet
  - `["jpg"]` - Only JPEG uploaded
  - `["heic", "jpg"]` - Both HEIC and JPEG uploaded

**Filename Generation Rules**:
- Format: `IMG_XXXX.ext` where XXXX is a 4-digit zero-padded number
- Sequence starts from 0001 for each date
- Existing photos are preserved (no re-indexing)
- File extension is preserved from request

**Error Response** (`400`):
```json
{
  "error": "bad_request",
  "message": "Invalid date format, expected YYYY-MM-DD"
}
```

**Error Response** (`401`):
```json
{
  "error": "unauthorized",
  "message": "Token not found"
}
```

**Error Response** (`500`):
```json
{
  "error": "internal_error",
  "message": "failed to create photo record: ..."
}
```

### POST /photos/upload

Upload a photo file to the server.

**Endpoint**: `POST /photos/upload`

**Headers**:
```
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data
```

**Form Fields**:
- `local_id` (string, required): Must match a local_id from indexing
- `file_type` (string, required): MIME type (e.g., 'image/jpeg')
- `file` (file, required): Photo file (max 50MB)

**Upload Behavior**:
- **Overwrite**: Files are always overwritten if they already exist (ensures latest version)
- **Extension Tracking**: Uploaded extensions are tracked and can be viewed in Index API response
- **Multiple Formats**: Same photo can have multiple formats uploaded (e.g., HEIC + JPEG)

**Example Request**:
```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "local_id=IMG_1234" \
  -F "file_type=image/jpeg" \
  -F "file=@/path/to/photo.jpg"
```

**Success Response** (`200`):
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_1234",
  "filename": "IMG_0001.jpg"
}
```

**File Storage Path**:
```
storage/photo/<user_id>/<year>/<month>/<day>/<filename.ext>
```

**Example**:
```
storage/photo/1/2025/12/10/IMG_0001.jpg
```

**Error Response** (`400`):
```json
{
  "error": "bad_request",
  "message": "local_id is required"
}
```

**Error Response** (`401`):
```json
{
  "error": "unauthorized",
  "message": "Token not found"
}
```

**Error Response** (`500`):
```json
{
  "error": "internal_error",
  "message": "photo not found"
}
```

---

## Error Handling

All error responses follow a consistent format:

```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

### Error Types

| Error Type | HTTP Status | Description |
|------------|-------------|-------------|
| `unauthorized` | 401 | Authentication failed or token missing/invalid |
| `bad_request` | 400 | Invalid request format or missing required fields |
| `internal_error` | 500 | Server-side error occurred |

### Common Error Scenarios

**Invalid Credentials** (401):
```json
{
  "error": "unauthorized",
  "message": "invalid credentials"
}
```

**Missing Token** (401):
```json
{
  "error": "unauthorized",
  "message": "Authorization header is required"
}
```

**Invalid Token Format** (401):
```json
{
  "error": "unauthorized",
  "message": "Invalid authorization header format"
}
```

**Missing Required Field** (400):
```json
{
  "error": "bad_request",
  "message": "local_id is required"
}
```

**Invalid JSON** (400):
```json
{
  "error": "bad_request",
  "message": "Invalid request format"
}
```

**Database Error** (500):
```json
{
  "error": "internal_error",
  "message": "failed to create photo record: ..."
}
```

---

## Rate Limiting

**Current Status**: Not implemented

**Planned Implementation**:
- Rate limit on `/login` endpoint: 5 requests per minute per IP
- Rate limit on photo upload: 10 requests per minute per authenticated user
- Rate limit headers in responses:
  ```
  X-RateLimit-Limit: 10
  X-RateLimit-Remaining: 9
  X-RateLimit-Reset: 1765378091
  ```

---

## Complete Workflow Example

### Step 1: Login
```bash
RESPONSE=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"SecurePass123"}')

TOKEN=$(echo $RESPONSE | jq -r '.token')
```

### Step 2: Index Photos
```bash
curl -s -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-12-10",
    "photos": [
      {
        "local_id": "IMG_1234",
        "creation_time": "2025-12-10T10:00:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

### Step 3: Upload File
```bash
curl -s -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "local_id=IMG_1234" \
  -F "file_type=image/jpeg" \
  -F "file=@/path/to/photo.jpg"
```

### Step 4: Verify Upload (Optional)
```bash
curl -s http://localhost:8080/status \
  -H "Authorization: Bearer $TOKEN"
```

---

## Postman Collection

A Postman collection is available in the `docs/` directory with pre-configured requests for all endpoints.

### Import Instructions
1. Open Postman
2. Click "Import"
3. Select the collection file
4. Configure environment variables:
   - `base_url`: `http://localhost:8080`
   - `token`: (will be set after login)

---

## OpenAPI Specification

OpenAPI 3.0 specification is available at:
```
docs/openapi.yaml
```

You can use it with:
- Swagger UI
- Postman
- Insomnia
- Any OpenAPI-compatible tool

---

## SDK Examples

### cURL

See individual endpoint examples above.

### JavaScript (Fetch API)

```javascript
// Login
const loginResponse = await fetch('http://localhost:8080/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'admin',
    password: 'SecurePass123'
  })
});
const { token } = await loginResponse.json();

// Index photos
await fetch('http://localhost:8080/photos/index', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    date: '2025-12-10',
    photos: [{
      local_id: 'IMG_1234',
      creation_time: '2025-12-10T10:00:00Z',
      file_extension: 'jpg',
      file_type: 'image/jpeg'
    }]
  })
});

// Upload photo
const formData = new FormData();
formData.append('local_id', 'IMG_1234');
formData.append('file_type', 'image/jpeg');
formData.append('file', fileInput.files[0]);

await fetch('http://localhost:8080/photos/upload', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: formData
});
```

---

## Database Schema

### Photo Table Structure

**Table Name**: `photos_user_<user_id>` (dynamic per user)

**Columns**:

| Column | Type | Description |
|--------|------|-------------|
| `local_id` | TEXT (PRIMARY KEY) | Client-side unique identifier |
| `creation_time` | DATETIME | When photo was taken |
| `file_path` | TEXT | Directory path (e.g., `storage/photo/1/2025/01/15/`) |
| `file_name` | TEXT | Filename without extension (e.g., `IMG_0001`) |
| `file_type` | TEXT | MIME type (e.g., `image/jpeg`) |
| `file_count` | INTEGER | Number of files uploaded (default: 0) |
| `uploaded_extensions` | TEXT | **NEW**: JSON array of uploaded extensions (e.g., `["jpg","heic"]`) |
| `created_at` | DATETIME | Record creation time |
| `updated_at` | DATETIME | Record last update time |
| `deleted_at` | DATETIME | Soft delete support |

**Example Record**:
```json
{
  "local_id": "IMG_001",
  "creation_time": "2025-01-15T10:30:00Z",
  "file_path": "storage/photo/1/2025/01/15/",
  "file_name": "IMG_0001",
  "file_type": "image/jpeg",
  "file_count": 2,
  "uploaded_extensions": "[\"jpg\",\"heic\"]",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:35:00Z"
}
```

### Extension Tracking

The `uploaded_extensions` field stores a JSON array of file extensions that have been successfully uploaded for each photo:

- `[]` - No files uploaded yet
- `["jpg"]` - Only JPEG uploaded
- `["heic","jpg"]` - Both HEIC and JPEG uploaded

**File Path Format**: `{file_path}{file_name}.{extension}`
**Example**: `storage/photo/1/2025/01/15/IMG_0001.jpg`

---

## Testing

### Unit Tests
```bash
go test ./internal/... -v
```

### Integration Tests
```bash
./tests/e2e/workflow_test.sh
./tests/performance/load_test.sh
```

---

## Support

For API-related questions:
- Check the [README.md](README.md)
- Review log files in `logs/` directory
- Open an issue on GitHub

---

**API Version**: 1.0.0
**Last Updated**: 2025-12-11

**Changelog**:

### Version 1.0.0 (2025-12-11)
- **NEW**: Extension tracking support - Track multiple file formats per photo
- **NEW**: Upload status in Index API response via `uploaded_extensions` field
- **NEW**: Overwrite behavior - Files are always overwritten on re-upload
- **NEW**: Database schema updated with `uploaded_extensions` column
