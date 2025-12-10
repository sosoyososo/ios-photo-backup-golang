# Feature Specification: Photo Backup Server

## Overview
Build a Go + Gin + SQLite application that combines an API server and command-line tool in a single codebase for photo backup and user management.

## Technical Stack
- **Language:** Go
- **Web Framework:** Gin
- **Database:** SQLite
- **ORM:** GORM
- **Authentication:** JWT
- **Storage:** Local filesystem

## Clarifications

### Session 2025-12-10
- Q: How should the server listen address be configured? → A: Command-line flags (e.g., --port 8080)
- Q: Where should the SQLite database file be stored? → A: Data directory (./data/app.db)
- Q: How should the JWT signing secret be provided? → A: Auto-generate and persist to file (data/jwt_secret.key), reuse if exists
- Q: Should the server auto-create missing directories on startup? → A: Yes - auto-create data/ and storage/ directories
- Q: Should database tables auto-migrate on server startup? → A: Yes - use GORM AutoMigrate
- Q: What should be the exact table names? → A: Plural form (users, tokens, photos_user_{id})

### Session 2025-12-09
- Q: Which ORM would you prefer for this project? → A: GORM (full-featured ORM with auto-migration capabilities)
- Q: How should duplicate photo detection work? → A: 1) Find by local_id in photo table, 2) Check file_count in database, 3) Re-verify actual file count in directory
- Q: How should JWT token refresh be implemented? → A: Dedicated /refresh endpoint for token lifecycle management
- Q: Which password hashing algorithm should be used? → A: bcrypt (industry standard with salt and work factor)
- Q: How should photo table store file information and naming work? → A: Store file_path and file_name separately; query database to count directory files for naming
- Q: What should be the JWT token expiration times? → A: Access token 1 hour, Refresh token 7 days
- Q: Should photo table store file extension or file type? → A: Store both file_extension and file_type (MIME string)
- Q: What is the upload workflow (Index API vs Upload API)? → A: 1) Call Index API to create records, 2) Call Upload API to upload files; check by local_id+extension, verify completeness, re-upload if incomplete
- Q: Which APIs should be implemented? → A: Only the 5 required APIs (login, refresh, status, index, upload) - no additional features to keep complexity low
- Q: How should the Index API work? → A: Submit date + batch of (local_id, creation_time, file_extension, file_type); server sorts by time, pre-assigns sequential filenames (IMG_0001.ext), creates DB records
- Q: How should reindex logic work? → A: Check local_id against database; if exists: SKIP (keep existing filename); if new: assign sequential filename
- Q: How to calculate next sequential number? → A: Count database records WHERE file_path contains the date
- Q: What is the numbering scope? → A: Per calendar date (each day resets to IMG_0000)
- Q: How to handle existing filesystem files? → A: Error if non-DB files exist (prevent conflicts)
- Q: Should we add token_type field? → A: No - use single token type (simplified)
- Q: Photo table primary key? → A: Use local_id as PRIMARY KEY
- Q: API authentication requirements? → A: Public: /login; Protected: /refresh, /status, /index, /upload (Authorization: Bearer)
- Q: File upload format? → A: multipart/form-data with binary file data
- Q: Error response format? → A: Detailed JSON with error, message, details, timestamp
- Q: Status endpoint purpose? → A: Validate JWT token is valid and not expired
- Q: Refresh API authentication? → A: Yes - requires Authorization: Bearer header, validates JWT before generating new token

## Functional Requirements

### Command Line Tool
The CLI tool provides three functions:
1. **Create User** - Create a new user account
2. **User List** - List all users
3. **Reset Password** - Reset user password

### API Server

#### Authentication & Authorization
- **Login:** Username and password authentication
- **JWT Token:** Use JWT to identify login state
- **Token Refresh:** Dedicated `/refresh` endpoint for token lifecycle management
- **Token Expiry:** Token expires in 7 days
- **Token Expiry Control:** SQLite secondary control for token expiration
- **Password Storage:** Store bcrypt hashed passwords
- **Password Validation:** Only validate length (minimum 8 characters), no complexity requirements

#### API Endpoints (Required Only)

**Authentication Requirements:**
- Public endpoints (no auth): /login
- Protected endpoints (require JWT): /refresh, /status, /photos/index, /photos/upload
- JWT Pass Method: Authorization header with Bearer token
  - Header: `Authorization: Bearer <jwt_token>`

**Authentication APIs:**
1. **Login** - POST /login
   - **Request:** `{"username": "string", "password": "string"}`
   - **Response:** `{"token": "jwt_token", "expires_at": "datetime"}`
   - **Errors:** 400 (invalid credentials), 500 (server error)

2. **Refresh Token** - POST /refresh
   - **Auth:** Required (Authorization: Bearer)
   - **Request:** `{}` (empty body - token is in Authorization header)
   - **Process:**
     1. Validate JWT token from Authorization header
     2. Check token exists in database and hasn't expired
     3. Generate new JWT token
     4. Delete old token from database
     5. Store new token in database
   - **Response:** `{"token": "new_jwt_token", "expires_at": "datetime"}`
   - **Errors:** 401 (invalid/expired token), 500 (server error)

**User APIs:**
3. **Online Status Check** - GET /status
   - **Auth:** Required (Authorization: Bearer)
   - **Response:** `{"status": "online", "user_id": 123, "username": "string"}`
   - **Errors:** 401 (invalid/expired token), 500 (server error)
   - **Purpose:** Validate JWT token is valid and not expired

**Photo APIs:**
4. **Index API** - POST /photos/index
   - **Auth:** Required (Authorization: Bearer)
   - **Request:** `{"date": "YYYY-MM-DD", "photos": [{"local_id": "string", "creation_time": "datetime", "file_extension": "jpg", "file_type": "image/jpeg"}]}`
   - **Process:**
     1. Receive batch of photos for specified date
     2. Sort photos by creation_time within the date
     3. Calculate directory path: storage/photo/{user_id}/YYYY/MM/DD/
     4. **File Naming Algorithm:**
        - Next Number: Count database records WHERE file_path contains the date
        - Scope: Per calendar date (each day resets to IMG_0000)
        - Example: For 2025-01-15, count records with path containing "2025/01/15"
     5. **Reindex Logic:**
        - Check each photo's local_id against database
        - If local_id exists: SKIP (ignore, keep existing filename)
        - If local_id is new: Assign next sequential filename (IMG_0001.ext, IMG_0002.ext, etc.)
     6. **Conflict Detection:**
        - Before indexing, scan directory for existing IMG_xxxx files
        - If non-DB files exist: Return error (cannot determine numbering)
        - Only allow indexing if directory is clean or only contains DB-tracked files
     7. Create database records only for new local_ids
   - **Response:** `{"status": "success", "date": "YYYY-MM-DD", "assigned_files": [{"local_id": "string", "filename": "IMG_0001.jpg"}]}`
   - **Errors:** 400 (bad request), 401 (unauthorized), 409 (conflict - orphaned files), 500 (server error)

5. **Photo Upload Service** - POST /photos/upload
   - **Auth:** Required (Authorization: Bearer)
   - **Request Format:** multipart/form-data
     - Fields: `local_id` (string), `file` (binary), `file_type` (string)
   - **Upload Process:**
     1. Look up photo record by local_id
     2. Check if file already exists at file_path + file_name
     3. If file exists:
        - Verify file integrity/completeness
        - If complete: skip upload, return success
        - If incomplete: delete old file, proceed to step 4
     4. If file doesn't exist or was incomplete:
        - Save file_data to pre-assigned filename (file_path + file_name)
        - Update file_count in database record
   - **Response:** `{"status": "success", "message": "File uploaded", "local_id": "string", "filename": "IMG_0001.jpg"}`
   - **Errors:** 400 (bad request), 401 (unauthorized), 404 (local_id not found), 500 (server error)

#### Error Response Format
**Structure:** `{"error": "error_type", "message": "detailed message", "details": {}, "timestamp": "datetime"}`

**HTTP Status Codes:**
- 200 - Success
- 400 - Bad Request (invalid input)
- 401 - Unauthorized (invalid/missing token)
- 404 - Not Found (resource doesn't exist)
- 409 - Conflict (data conflict, e.g., orphaned files)
- 500 - Internal Server Error

#### Photo Storage System
- **Storage Path Pattern:** `storage/photo/{user_id}/year/month/day/IMG_00xx.fileextension`
- **Naming Convention:** Each day folder starts from IMG_0000
- **File Extensions:** Same name may have different extensions
- **Database Tracking:** SQLite stores per-user photo tables:
  - Table name includes user ID (e.g., `photos_user_123`)
  - Columns: local_id, creation_time, file storage path, file extension, file count

## Data Model

### User Table (`users`)
- User ID (INTEGER PRIMARY KEY)
- Username (VARCHAR UNIQUE NOT NULL)
- Password (VARCHAR NOT NULL) - bcrypt hash
- Created at (DATETIME NOT NULL)
- Updated at (DATETIME NOT NULL)

### Token Table (`tokens`)
- Token ID (INTEGER PRIMARY KEY)
- User ID (INTEGER NOT NULL, FOREIGN KEY → users.id)
- Token value (TEXT NOT NULL, UNIQUE) - JWT token
- Created at (DATETIME NOT NULL)
- Expires at (DATETIME NOT NULL)

### Photo Tables (Per User - `photos_user_{user_id}`)
- **Naming Convention:** `photos_user_{user_id}`
- **Primary Key:** local_id (UNIQUE, NOT NULL)
- **Columns:**
  - local_id (PRIMARY KEY) - Client-side photo ID
  - creation_time (DATETIME NOT NULL)
  - file_path (TEXT NOT NULL) - Directory path only
  - file_name (VARCHAR NOT NULL) - Filename only (e.g., IMG_0001.jpg)
  - file_extension (VARCHAR NOT NULL) - Extension (e.g., "jpg", "png")
  - file_type (VARCHAR NOT NULL) - MIME type string (e.g., "image/jpeg")
  - file_count (INTEGER NOT NULL DEFAULT 0) - Number of files for this local_id

## Server Initialization

### Startup Configuration
- **Listen Address:** Configurable via command-line flags (e.g., `--host`, `--port`)
- **Default Port:** 8080 (if not specified)

### Directory Structure
- **Data Directory:** `./data/` - stores database and secrets
- **Storage Directory:** `./storage/` - stores uploaded photos
- **Auto-Create:** Server creates data/ and storage/ directories on startup if missing

### Database Initialization
- **Database File:** `./data/app.db`
- **Auto-Migrate:** Use GORM AutoMigrate on startup to create/update tables
- **Tables:** users, tokens (photo tables created dynamically per user)

### JWT Secret Management
- **Secret File:** `./data/jwt_secret.key`
- **Generation:** Auto-generate random secret on first run
- **Persistence:** Reuse existing secret file if present (ensures token validity across restarts)

## Non-Functional Requirements
- Single codebase for both CLI and API server
- Local file storage
- JWT-based authentication
- SQLite for persistent storage
- **Scope Control:** Only implement specified APIs - no additional features
- **Task Tracking:** Use local documentation (no GitHub issues)
