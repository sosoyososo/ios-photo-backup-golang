# Quickstart Guide: Photo Backup Server

## Overview

Photo Backup Server is a Go application that combines an API server and command-line tool for photo backup and user management. It provides JWT-based authentication, photo upload with automatic sequential naming, and user management via both API and CLI.

## Architecture

- **Single Binary**: Combines API server and CLI tool
- **Database**: SQLite for metadata, local filesystem for photo storage
- **Authentication**: JWT tokens with Bearer authorization
- **Storage Pattern**: `storage/photo/{user_id}/YYYY/MM/DD/IMG_xxxx.ext`

## Project Structure

```
src/
├── cmd/
│   ├── server/          # API server entry point
│   └── cli/             # CLI tool entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   ├── service/         # Business logic layer
│   ├── repository/      # Data access layer
│   ├── models/          # Data models
│   └── config/          # Configuration
└── storage/             # Photo storage (auto-created)

tests/
├── unit/                # Unit tests
├── integration/         # API integration tests
└── contract/            # API contract tests
```

## Prerequisites

- Go 1.21 or later
- Git

## Installation

1. **Clone and build**:
```bash
git clone <repository-url>
cd ios-photo-backup-server
go build -o photo-backup ./cmd/server
go build -o photo-backup-cli ./cmd/cli
```

2. **Initialize directories** (auto-created on first run):
```bash
./photo-backup --help
```

## Configuration

### Server Configuration

The server accepts command-line flags:

```bash
./photo-backup \
  --host 0.0.0.0 \
  --port 8080 \
  --data-dir ./data \
  --storage-dir ./storage
```

**Default Values**:
- Host: `0.0.0.0`
- Port: `8080`
- Data Directory: `./data/` (stores database and JWT secret)
- Storage Directory: `./storage/` (stores uploaded photos)

### Auto-Initialization

On first startup, the server automatically:
- Creates `data/` and `storage/` directories
- Generates JWT secret in `data/jwt_secret.key`
- Creates SQLite database at `data/app.db`
- Runs GORM AutoMigrate to create tables

## API Server Usage

### Start Server

```bash
./photo-backup --port 8080
```

Server will listen on http://localhost:8080

### CLI Tool Usage

The same binary provides CLI functionality:

```bash
# Create a user
./photo-backup-cli user create --username johndoe --password securepass

# List all users
./photo-backup-cli user list

# Reset user password
./photo-backup-cli user reset-password --username johndoe --password newpass
```

## API Reference

### Authentication

All protected endpoints require:
```
Authorization: Bearer <jwt_token>
```

#### 1. Login

**Endpoint**: `POST /login`

**Request**:
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"johndoe","password":"securepass"}'
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T12:00:00Z"
}
```

Use the token for subsequent requests.

#### 2. Refresh Token

**Endpoint**: `POST /refresh`

**Request**:
```bash
curl -X POST http://localhost:8080/refresh \
  -H "Authorization: Bearer <token>"
```

**Response**:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T12:00:00Z"
}
```

#### 3. Status Check

**Endpoint**: `GET /status`

**Request**:
```bash
curl http://localhost:8080/status \
  -H "Authorization: Bearer <token>"
```

**Response**:
```json
{
  "status": "online",
  "user_id": 123,
  "username": "johndoe"
}
```

### Photo Management

#### 4. Index Photos

Pre-assign filenames to a batch of photos.

**Endpoint**: `POST /photos/index`

**Request**:
```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-15",
    "photos": [
      {
        "local_id": "IMG_1234",
        "creation_time": "2025-01-15T10:30:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }'
```

**Response**:
```json
{
  "status": "success",
  "date": "2025-01-15",
  "assigned_files": [
    {
      "local_id": "IMG_1234",
      "filename": "IMG_0001.jpg"
    }
  ]
}
```

**File Naming Algorithm**:
- Sequential numbering per calendar date
- Starts at IMG_0000 each day
- Format: IMG_XXXX.ext (4-digit zero-padded)

#### 5. Upload Photo

Upload a photo file using the pre-assigned filename.

**Endpoint**: `POST /photos/upload`

**Request**:
```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer <token>" \
  -F "local_id=IMG_1234" \
  -F "file=@/path/to/photo.jpg" \
  -F "file_type=image/jpeg"
```

**Response**:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_1234",
  "filename": "IMG_0001.jpg"
}
```

**File Storage**:
```
storage/photo/{user_id}/2025/01/15/IMG_0001.jpg
```

## Workflow Example

```bash
# 1. Start server
./photo-backup --port 8080 &

# 2. Create user via CLI
./photo-backup-cli user create --username alice --password mypassword

# 3. Login
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"mypassword"}' | jq -r '.token')

# 4. Index photos
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-01-15",
    "photos": [
      {"local_id": "IMG_001", "creation_time": "2025-01-15T10:00:00Z", "file_extension": "jpg", "file_type": "image/jpeg"},
      {"local_id": "IMG_002", "creation_time": "2025-01-15T11:00:00Z", "file_extension": "jpg", "file_type": "image/jpeg"}
    ]
  }'

# 5. Upload photos
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "local_id=IMG_001" \
  -F "file=@photo1.jpg" \
  -F "file_type=image/jpeg"

curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "local_id=IMG_002" \
  -F "file=@photo2.jpg" \
  -F "file_type=image/jpeg"

# 6. Verify upload
curl http://localhost:8080/status \
  -H "Authorization: Bearer $TOKEN"
```

## Error Handling

All errors follow this format:

```json
{
  "error": "error_type",
  "message": "detailed message",
  "details": {},
  "timestamp": "2025-12-10T12:00:00Z"
}
```

**HTTP Status Codes**:
- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (invalid/missing token)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (orphaned files)
- `500` - Internal Server Error

## Testing

```bash
# Run unit tests
go test ./tests/unit/...

# Run integration tests
go test ./tests/integration/...

# Run contract tests
go test ./tests/contract/...

# Run all tests
go test ./...
```

## Troubleshooting

### Port Already in Use

```bash
# Use a different port
./photo-backup --port 8081
```

### Database Locked

- Ensure only one server instance is running
- Check file permissions on `data/` directory

### Upload Fails

1. Verify local_id was indexed first
2. Check file size (default limit: 50MB)
3. Verify file type matches MIME type
4. Check storage directory permissions

### Token Expired

- Use `/refresh` endpoint to get a new token
- Default expiration: 7 days

## Production Deployment

1. **Build for production**:
```bash
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o photo-backup ./cmd/server
```

2. **Use environment variables** (if using systemd):
```bash
# /etc/systemd/system/photo-backup.service
[Service]
ExecStart=/usr/local/bin/photo-backup --host 0.0.0.0 --port 8080
Environment=PORT=8080
```

3. **Setup reverse proxy** (Nginx):
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Authorization $http_authorization;
    }
}
```

4. **Backup data**:
```bash
# Backup database and photos
tar -czf backup-$(date +%Y%m%d).tar.gz data/ storage/
```

## Support

For issues and questions:
- Review the [Feature Specification](./spec.md)
- Check [API Contracts](./contracts/)
- See [Data Model](./data-model.md)
- Review [Research Notes](./research.md)
