# Data Model: Photo Backup Server

## Entities

### User

**Purpose**: Represents a user account in the system

**Fields**:
- `ID` (uint, PRIMARY KEY) - Auto-incrementing user ID
- `Username` (string, UNIQUE, NOT NULL) - Unique username for login
- `PasswordHash` (string, NOT NULL) - bcrypt hashed password
- `CreatedAt` (time.Time, NOT NULL) - Account creation timestamp
- `UpdatedAt` (time.Time, NOT NULL) - Last update timestamp

**Validation Rules**:
- Username: 3-50 characters, alphanumeric only
- Password: Minimum 8 characters (no complexity requirements per spec)

**Relationships**:
- Has many Tokens
- Has one dynamic Photo Table (photos_user_{user_id})

---

### Token

**Purpose**: Stores active JWT tokens for authentication

**Fields**:
- `ID` (uint, PRIMARY KEY) - Auto-incrementing token ID
- `UserID` (uint, NOT NULL, FOREIGN KEY → users.id) - Associated user
- `TokenValue` (string, UNIQUE, NOT NULL) - JWT token string
- `CreatedAt` (time.Time, NOT NULL) - Token creation timestamp
- `ExpiresAt` (time.Time, NOT NULL) - Token expiration timestamp

**Validation Rules**:
- Token expires 7 days after creation
- Each user can have multiple tokens (until refresh deletes old one)

**Lifecycle**:
- Created on successful login
- Deleted on token refresh (old token removed)
- Validated on each protected API call

---

### Photo (Dynamic Table: photos_user_{user_id})

**Purpose**: Stores metadata for photos belonging to a specific user

**Table Naming**: `photos_user_{user_id}` where {user_id} is the user's ID

**Fields**:
- `LocalID` (string, PRIMARY KEY) - Client-side photo identifier
- `CreationTime` (time.Time, NOT NULL) - When photo was taken
- `FilePath` (string, NOT NULL) - Directory path only (e.g., "storage/photo/123/2025/01/15")
- `FileName` (string, NOT NULL) - Filename only (e.g., "IMG_0001.jpg")
- `FileExtension` (string, NOT NULL) - Extension without dot (e.g., "jpg", "png")
- `FileType` (string, NOT NULL) - MIME type (e.g., "image/jpeg")
- `FileCount` (int, DEFAULT 0) - Number of files uploaded for this local_id

**Naming Algorithm**:
- Sequential numbering per calendar date
- Format: IMG_XXXX.ext where XXXX is 4-digit zero-padded number
- Example: IMG_0001.jpg, IMG_0002.png, etc.
- Resets to IMG_0000 each day

**Validation Rules**:
- File extensions must be valid image types (jpg, jpeg, png, heic, etc.)
- CreationTime must be parseable datetime
- FileCount ≥ 0

**File Storage Pattern**:
```
storage/photo/{user_id}/{year}/{month}/{day}/{filename}
```

**Example**:
```
storage/photo/123/2025/01/15/IMG_0001.jpg
storage/photo/123/2025/01/15/IMG_0002.png
```

---

## Entity Relationships

```text
User (1) -----> (N) Token
User (1) -----> (1) Photo Table (photos_user_{user_id})
Photo Table (1) -----> (N) Photo Record
```

---

## Database Schema (SQL)

```sql
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Tokens table
CREATE TABLE tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_value TEXT UNIQUE NOT NULL,
    created_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Photo table (example for user 123)
CREATE TABLE photos_user_123 (
    local_id TEXT PRIMARY KEY,
    creation_time DATETIME NOT NULL,
    file_path TEXT NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_extension VARCHAR(10) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    file_count INTEGER NOT NULL DEFAULT 0
);
```

---

## GORM Models (Go)

```go
type User struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    Username     string    `json:"username" gorm:"uniqueIndex;not null;size:255"`
    PasswordHash string    `json:"-" gorm:"not null;size:255"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
    Tokens       []Token   `json:"-"`
}

type Token struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    UserID    uint      `json:"-" gorm:"not null"`
    TokenValue string   `json:"-" gorm:"uniqueIndex;not null"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
    User      User      `json:"-"`
}

type Photo struct {
    LocalID       string    `json:"local_id" gorm:"primaryKey"`
    CreationTime  time.Time `json:"creation_time" gorm:"not null"`
    FilePath      string    `json:"file_path" gorm:"not null"`
    FileName      string    `json:"file_name" gorm:"not null;size:255"`
    FileExtension string    `json:"file_extension" gorm:"not null;size:10"`
    FileType      string    `json:"file_type" gorm:"not null;size:50"`
    FileCount     int       `json:"file_count" gorm:"default:0"`
}
```

---

## State Transitions

### User Lifecycle
```
Created → Active → (never deleted per spec)
```

### Token Lifecycle
```
Created → Active → Expired/Deleted
         ↓
    Refresh creates new token
```

### Photo Lifecycle
```
Indexed → Uploaded → Complete
    ↓
Re-indexed (if already exists, skip)
```

---

## Indexing Strategy

**Users Table**:
- Primary Key: `id`
- Unique Index: `username`

**Tokens Table**:
- Primary Key: `id`
- Unique Index: `token_value`
- Index: `user_id` (for cleanup)
- Index: `expires_at` (for cleanup queries)

**Photo Tables** (per user):
- Primary Key: `local_id`
- Index: `creation_time` (for sorting during index API)
- Index: `file_path` (for counting files per date)
