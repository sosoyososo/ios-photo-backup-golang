# Photo Backup Server

A production-ready, dual-mode photo backup server built with Go, featuring a RESTful API and command-line interface for managing users, authentication, and photo uploads.

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

## ✨ Features

### 🔐 User Management (CLI)
- Create users with username/password authentication
- List all users in the system
- Reset user passwords
- Secure password hashing with bcrypt

### 🔑 Authentication (API)
- JWT-based authentication
- Token-based API access
- Token refresh mechanism
- Authentication status validation

### 📸 Photo Management
- **Index Photos**: Batch process photos with automatic sequential naming
- **Upload Files**: Multipart file upload with validation
- **Sequential Naming**: Automatic filename generation (IMG_0001.jpg, IMG_0002.jpg, etc.)
- **Per-User Storage**: Isolated storage with dynamic database tables
- **File Organization**: Photos organized by date (YYYY/MM/DD structure)
- **Extension Tracking**: Track multiple file formats per photo (e.g., HEIC + JPEG)
- **Overwrite Support**: Always overwrite files on re-upload (ensures latest version)
- **Upload Status**: View which formats have been uploaded via Index API response

### 🏗️ Production-Ready
- Structured JSON logging
- Graceful shutdown with signal handling
- SQLite database with GORM
- Cross-platform binaries (Darwin, Linux, Windows)
- Comprehensive test suite
- Health check endpoint

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- SQLite (included)

### Building from Source

```bash
# Clone the repository
git clone <repository-url>
cd ios-photo-backup-server

# Build for current platform
./build.sh

# Build for all platforms
./build.sh -a

# Build for specific platform
./build.sh -p linux/amd64
```

### Installation

#### Option 1: Use Pre-built Binaries

Download the latest release from the [releases page](releases):

```bash
# Extract archive
tar -xzf photo-backup-server-1.0.0-darwin-arm64.tar.gz

# Make binaries executable
chmod +x photo-backup-server photo-backup-cli
```

#### Option 2: Build from Source

```bash
# Build server
go build -o photo-backup-server ./cmd/server

# Build CLI
go build -o photo-backup-cli ./cmd/cli
```

### First-Time Setup

1. **Create an Admin User**:
```bash
./photo-backup-cli user create --username admin --password "YourSecurePassword"
```

2. **Start the Server**:
```bash
./photo-backup-server --port 8080
```

3. **Verify Server is Running**:
```bash
curl http://localhost:8080/health
```

## 📖 Usage Guide

### CLI Commands

#### Create User
```bash
./photo-backup-cli user create --username <username> --password <password>

# Example
./photo-backup-cli user create --username john --password "SecurePass123"
```

#### List Users
```bash
./photo-backup-cli user list
```

#### Reset Password
```bash
./photo-backup-cli user reset-password --username <username> --password <newpassword>

# Example
./photo-backup-cli user reset-password --username john --password "NewPassword456"
```

### API Endpoints

#### Authentication

**Login**
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "YourSecurePassword"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T21:34:47+08:00"
}
```

**Refresh Token**
```bash
curl -X POST http://localhost:8080/refresh \
  -H "Authorization: Bearer <token>"
```

**Check Status**
```bash
curl -X GET http://localhost:8080/status \
  -H "Authorization: Bearer <token>"
```

#### Photo Management

**Index Photos** (Assign filenames)
```bash
curl -X POST http://localhost:8080/photos/index \
  -H "Authorization: Bearer <token>" \
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

Response:
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

**Note**: The `uploaded_extensions` field shows which file formats have been uploaded for each photo. An empty array `[]` means no files uploaded yet.

**Upload Photo**
```bash
curl -X POST http://localhost:8080/photos/upload \
  -H "Authorization: Bearer <token>" \
  -F "local_id=IMG_1234" \
  -F "file_type=image/jpeg" \
  -F "file=@/path/to/photo.jpg"
```

Response:
```json
{
  "status": "success",
  "message": "File uploaded",
  "local_id": "IMG_1234",
  "filename": "IMG_0001.jpg"
}
```

## ⚙️ Configuration

### Command-Line Flags

#### Server
```bash
./photo-backup-server [flags]

Flags:
  --port int          Server port (default 8080)
  --host string       Server host (default "0.0.0.0")
  --db-path string    Database file path (default "./data/app.db")
  --storage-dir string Storage directory (default "./storage")
  --jwt-secret-path string JWT secret file path (default "./jwt_secret.key")
```

#### CLI
```bash
./photo-backup-cli [command] [flags]

Flags:
  --db-path string    Database file path (default "./data/app.db")
```

### Configuration File

You can create a `.env` file or use environment variables:

```bash
PORT=8080
HOST=0.0.0.0
DB_PATH=./data/app.db
STORAGE_DIR=./storage
JWT_SECRET_PATH=./jwt_secret.key
```

## 📁 Project Structure

```
ios-photo-backup-server/
├── cmd/
│   ├── server/           # Server entry point
│   └── cli/              # CLI entry point
├── internal/
│   ├── api/
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # Gin middleware
│   │   ├── routes/       # Route definitions
│   │   └── errors/       # Error handling
│   ├── config/           # Configuration management
│   ├── logger/           # Structured logging
│   ├── models/           # Data models
│   ├── repository/       # Database layer
│   └── service/          # Business logic
├── tests/
│   ├── e2e/              # End-to-end tests
│   └── performance/      # Performance tests
├── specs/                # Specifications and tasks
├── dist/                 # Build artifacts
└── build.sh              # Production build script
```

## 🧪 Testing

### Run All Tests
```bash
# End-to-end workflow test
./tests/e2e/workflow_test.sh

# Performance test (10 concurrent uploads)
./tests/performance/load_test.sh

# Quick server test
./test_server_features.sh
```

### Test Coverage

The project includes:
- ✅ End-to-end workflow validation
- ✅ Performance testing (10 concurrent uploads)
- ✅ API integration tests
- ✅ Authentication flow tests
- ✅ Photo indexing and upload tests

## 📊 Architecture

### Technology Stack
- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: SQLite with GORM ORM
- **Authentication**: JWT (golang-jwt/jwt/v5)
- **Password Hashing**: bcrypt
- **Logging**: Structured JSON logging

### Database Schema

#### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

#### Tokens Table
```sql
CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    token_value VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP,
    expires_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

#### Photos Table (Per User)
```sql
CREATE TABLE photos_user_<user_id> (
    local_id VARCHAR(255) PRIMARY KEY,
    creation_time TIMESTAMP,
    file_path VARCHAR(255),
    file_name VARCHAR(255),
    file_type VARCHAR(50),
    file_count INTEGER DEFAULT 0,
    uploaded_extensions TEXT DEFAULT '[]',  -- NEW: JSON array of uploaded extensions
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### File Organization

Photos are stored with the following structure:
```
storage/
└── photo/
    └── <user_id>/
        └── <year>/
            └── <month>/
                └── <day>/
                    ├── IMG_0001.jpg
                    ├── IMG_0002.jpg
                    └── ...
```

## 🔒 Security

- **Password Hashing**: Uses bcrypt with salt
- **JWT Tokens**: Signed with HMAC-SHA256
- **Token Expiration**: 7-day validity period
- **Token Rotation**: Old tokens deleted on refresh
- **Authentication Required**: All photo endpoints protected
- **Input Validation**: Request validation on all endpoints

## 📈 Performance

### Benchmarks
- **Concurrent Uploads**: 10 simultaneous uploads supported
- **Average Response Time**: < 500ms for typical operations
- **Throughput**: > 2 uploads/second
- **Success Rate**: 100% for valid requests

### Optimization Features
- GORM query optimization
- Minimal memory allocations
- Efficient file I/O
- Database connection pooling

## 🛠️ Development

### Building for Development
```bash
# Build server
go build -o photo-backup-server ./cmd/server

# Build CLI
go build -o photo-backup-cli ./cmd/cli

# Run tests
go test ./... -v
```

### Running in Development Mode
```bash
# Start server with debug logging
./photo-backup-server --port 8080

# Monitor logs
tail -f logs/photo-backup-*.log
```

## 📝 API Reference

### Response Format

All API responses follow a consistent format:

**Success Response**:
```json
{
  "status": "success",
  "data": { ... }
}
```

**Error Response**:
```json
{
  "error": "error_type",
  "message": "Human-readable error message"
}
```

### HTTP Status Codes
- `200` - Success
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (invalid or missing token)
- `500` - Internal Server Error

### Rate Limiting
Currently not implemented (see Future Enhancements)

## 🚀 Deployment

### Using Pre-built Binaries

1. **Download** the appropriate binary for your platform
2. **Extract** the archive
3. **Configure** using flags or environment variables
4. **Run** the server
5. **Monitor** logs in the `logs/` directory

### Docker (Future)

Docker support is planned for a future release.

### Systemd Service

Example systemd unit file:

```ini
[Unit]
Description=Photo Backup Server
After=network.target

[Service]
Type=simple
User=photo-backup
WorkingDirectory=/opt/photo-backup
ExecStart=/opt/photo-backup/photo-backup-server --port 8080
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## 🐛 Troubleshooting

### Common Issues

**Server won't start**
- Check if port is already in use: `lsof -i :8080`
- Verify database directory permissions
- Check logs in `logs/` directory

**Authentication fails**
- Ensure user exists: `./photo-backup-cli user list`
- Verify password is correct
- Check token expiration time

**Photo upload fails**
- Verify local_id was indexed first
- Check file size (max 50MB)
- Ensure storage directory is writable

**Database locked**
- Stop all server instances
- Remove `.db-wal` and `.db-shm` files if needed
- Restart server

### Logs

Structured JSON logs are written to:
- `logs/photo-backup-YYYYMMDD-HHMMSS.log`

View logs:
```bash
tail -f logs/photo-backup-*.log
```

Parse JSON logs:
```bash
cat logs/photo-backup-*.log | jq '.'
```

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [GORM](https://gorm.io/)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

## 📞 Support

For issues and questions:
- Open an issue on GitHub
- Check existing documentation
- Review log files for errors

## 📋 Changelog

### Version 1.0.0 (2025-12-11)

#### ✨ New Features
- **Extension Tracking**: Track multiple file formats per photo (e.g., HEIC + JPEG)
- **Upload Status**: View which formats have been uploaded via Index API response
- **Overwrite Support**: Files are always overwritten on re-upload (ensures latest version)
- **Database Schema**: Added `uploaded_extensions` column to photo tables

#### 📚 Documentation
- Updated API documentation with extension tracking details
- Added database schema documentation
- Added upload behavior documentation

#### 🔧 Technical Details
- Storage format: JSON array of extensions (e.g., `["jpg","heic"]`)
- Backward compatible: Existing photo records get default empty array
- GORM AutoMigrate handles schema updates automatically

---

## 🗺️ Roadmap

### Upcoming Features
- [ ] Contract tests for all APIs
- [ ] CLI commands integration tests
- [ ] Comprehensive error handling validation
- [ ] CORS configuration for web clients
- [ ] Rate limiting on /login endpoint
- [ ] Docker support
- [ ] Prometheus metrics
- [ ] Web UI for photo management
- [ ] Cloud storage support (S3, etc.)
- [ ] Video file support

---

**Version**: 1.0.0
**Last Updated**: 2025-12-11
