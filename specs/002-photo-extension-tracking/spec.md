# Feature Specification: Photo Extension Tracking

**Feature Branch**: `002-photo-extension-tracking`
**Created**: 2025-12-11
**Status**: Draft
**Input**: User description: "修改 photo 结构，增加一列，存储文件拓展名列表，代表上传了哪些后缀名的文件。index api返回这些后缀，表示这些文件已经上传。修改upload api 无需判断文件是否上传，如果已经上传直接覆盖。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Track Multiple File Extensions for Single Photo (Priority: P1)

A user uploads different versions of the same photo (e.g., original HEIC and converted JPEG). The system tracks which file extensions have been uploaded for each photo and prevents duplicate uploads.

**Why this priority**: This enables users to store multiple formats of the same photo without conflicts, which is essential for photo backup systems that handle different image formats.

**Independent Test**: Can be tested by uploading two files with the same local_id but different extensions, verifying the system tracks both extensions and allows both uploads.

**Acceptance Scenarios**:

1. **Given** a photo has been indexed with local_id "IMG_001", **When** user uploads a file with extension "jpg", **Then** the system records "jpg" as an uploaded extension for this photo.

2. **Given** a photo has extension "jpg" uploaded, **When** user uploads the same photo with extension "heic", **Then** the system accepts the upload and updates the extension list to include both "jpg" and "heic".

3. **Given** a photo has extensions "jpg" and "heic" uploaded, **When** user requests photo index, **Then** the response includes information showing both extensions have been uploaded.

---

### User Story 2 - Overwrite Existing Files on Re-upload (Priority: P1)

A user re-uploads a photo file. The system always overwrites the existing file instead of checking if it exists first, ensuring the latest version is always stored.

**Why this priority**: This simplifies the upload logic and ensures users always get the most up-to-date file stored, which is critical for photo backup integrity.

**Independent Test**: Can be tested by uploading a file, then uploading a different file with the same local_id and extension, verifying the second file completely replaces the first.

**Acceptance Scenarios**:

1. **Given** a photo file "IMG_0001.jpg" exists, **When** user uploads a new file with the same local_id and extension, **Then** the new file completely replaces the old file.

2. **Given** a photo has multiple uploaded extensions, **When** user re-uploads one extension, **Then** only that specific extension file is overwritten, other extensions remain unchanged.

---

### User Story 3 - View Uploaded Extensions in Index Response (Priority: P2)

A user wants to know which file formats have been successfully uploaded for each photo. The index API returns this information so users can verify upload status.

**Why this priority**: Provides transparency to users about upload status and helps them identify which files still need to be uploaded.

**Independent Test**: Can be tested by checking the index API response contains a field showing which extensions have been uploaded for each photo.

**Acceptance Scenarios**:

1. **Given** photos have been indexed, **When** user calls the index API, **Then** the response includes a list of uploaded extensions for each photo.

2. **Given** a photo has no files uploaded yet, **When** user calls the index API, **Then** the response shows an empty list or null for uploaded extensions.

3. **Given** a photo has two extensions uploaded, **When** user calls the index API, **Then** the response shows both extensions in the uploaded extensions list.

---

### Edge Cases

- What happens when a user uploads a file with an extension not in the original index request?
- How does the system handle concurrent uploads of the same photo with the same extension?
- What happens when the extensions list is corrupted or missing?
- How does the system behave when storage is full during an overwrite operation?
- What happens if a photo record exists but the actual files are missing from storage?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST add a new column to the photo data structure to store a list of uploaded file extensions.

- **FR-002**: System MUST update the extensions list whenever a file is successfully uploaded for a photo.

- **FR-003**: System MUST always overwrite existing files when uploading, regardless of whether a file with the same name already exists.

- **FR-004**: System MUST return the list of uploaded extensions in the index API response for each photo.

- **FR-005**: System MUST support multiple different extensions for the same photo (e.g., heic, jpg, png).

- **FR-006**: System MUST maintain the uploaded extensions list even when files are overwritten.

- **FR-007**: System MUST handle the case where a photo has no uploaded extensions (return empty list or null).

- **FR-008**: System MUST preserve existing photo data (local_id, creation_time, file_path, file_name, file_type) when adding extension tracking.

### Key Entities

- **Photo**: Represents a photo in the system with metadata including local_id, creation_time, file_path, file_name, file_type, and now a list of uploaded extensions. Each photo can have multiple file extensions uploaded.

- **Uploaded Extension**: A data element representing a file extension that has been successfully uploaded for a photo. Stored as part of the photo record.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can upload multiple file formats of the same photo (e.g., HEIC and JPEG) without conflicts, with each format tracked separately.

- **SC-002**: Re-uploading a file with the same local_id and extension always results in the new file completely replacing the old file (100% overwrite success rate).

- **SC-003**: Index API responses accurately reflect which file extensions have been uploaded for each photo (100% accuracy in tracking uploaded extensions).

- **SC-004**: Users can verify upload status for each photo format through the index API without making additional requests.

- **SC-005**: The system maintains backward compatibility with existing photo records that don't have extension lists (gracefully handles missing or null extension data).
