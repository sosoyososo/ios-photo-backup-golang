# Implementation Plan: Photo Extension Tracking

**Branch**: `002-photo-extension-tracking` | **Date**: 2025-12-11 | **Spec**: [link](file:///Users/karsa/proj/photobackup/ios-photo-backup-server/specs/002-photo-extension-tracking/spec.md)
**Input**: Feature specification from `/specs/002-photo-extension-tracking/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

This feature adds extension tracking to the photo backup system. A new column will be added to the photo data structure to store a list of uploaded file extensions. The Index API will return this list to show which files have been uploaded, and the Upload API will be simplified to always overwrite files without checking for existence first.

## Technical Context

**Language/Version**: Go (latest stable)
**Primary Dependencies**: Gin (HTTP framework), GORM (ORM), SQLite (database), JWT (authentication), bcrypt (password hashing)
**Storage**: SQLite database for metadata, filesystem for actual photo files
**Testing**: Go testing framework (testing package)
**Target Platform**: Linux server
**Project Type**: Single server application
**Performance Goals**: Support efficient batch photo indexing and uploads
**Constraints**: 50MB max file upload size, sequential filename generation
**Scale/Scope**: Multi-user photo backup system with dynamic table per user

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

No constitution file found - proceeding without constitutional constraints.

## Project Structure

### Documentation (this feature)

```text
specs/002-photo-extension-tracking/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/
├── models/              # Data models (Photo struct)
│   └── photo.go
├── service/             # Business logic (PhotoService)
│   └── photo.go
├── repository/          # Data access (PhotoRepository)
│   └── photo.go
└── api/handlers/photo/  # API endpoints
    ├── index.go
    └── upload.go

cmd/
├── server/              # Server entry point
└── cli/                 # CLI commands

tests/
├── integration/         # Integration tests
└── unit/                # Unit tests
```

**Structure Decision**: Extending existing Go backend structure with modifications to photo-related modules

## Complexity Tracking

N/A - No constitutional violations

---

## Phase 0: Outline & Research ✓ COMPLETE

**Output**: `research.md`

**Research Completed**:
1. ✓ Database storage approach (JSON text field)
2. ✓ Migration strategy (GORM AutoMigrate with default)
3. ✓ Extension list update logic
4. ✓ API response format
5. ✓ Upload API simplification (remove existence check)

**Key Decisions**:
- Store extensions as JSON array in TEXT column
- Default value `'[]'` for backward compatibility
- Remove file existence check, always overwrite
- Add `uploaded_extensions` to Index API response

---

## Phase 1: Design & Contracts ✓ COMPLETE

**Outputs**:
- ✓ `data-model.md` - Updated schema with UploadedExtensions field
- ✓ `/contracts/index-api.md` - Modified Index API contract
- ✓ `/contracts/upload-api.md` - Simplified Upload API contract
- ✓ `quickstart.md` - Implementation guide with code examples
- ✓ Agent context updated (CLAUDE.md)

**Design Complete**:
- Photo model updated with `UploadedExtensions` field
- Repository methods for extension management
- Service layer updated for overwrite behavior
- API contracts updated for extension tracking
- Full implementation guide provided

---

## Phase 2: Implementation Tasks

**Next Step**: Run `/speckit.tasks` to generate detailed implementation tasks

**Implementation Scope**:
1. Update Photo model (internal/models/photo.go)
2. Add extension helpers to repository (internal/repository/photo.go)
3. Modify PhotoService.UploadPhoto (internal/service/photo.go)
4. Update PhotoIndexResponse struct (internal/service/photo.go)
5. Restart server for AutoMigrate
6. Test all scenarios

**Estimated Effort**: 4-6 hours
**Dependencies**: None (uses existing infrastructure)
**Risk Level**: Low (backward compatible, simple changes)
