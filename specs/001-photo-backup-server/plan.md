# Implementation Plan: Photo Backup Server

**Branch**: `001-photo-backup-server` | **Date**: 2025-12-10 | **Spec**: spec.md
**Input**: Feature specification from `/specs/001-photo-backup-server/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a Go + Gin + SQLite application that combines an API server and command-line tool in a single codebase for photo backup and user management. The system provides JWT-based authentication, photo upload with automatic sequential naming, and user management via both API and CLI.

## Technical Context

**Language/Version**: Go (latest stable)
**Primary Dependencies**: Gin (HTTP framework), GORM (ORM), SQLite (database), JWT (authentication), bcrypt (password hashing)
**Storage**: SQLite for metadata, Local filesystem for photo storage
**Testing**: Go testing package + testify for assertions
**Target Platform**: Linux/macOS server
**Project Type**: Single binary with dual mode (API server + CLI tool)
**Performance Goals**: 10 concurrent uploads, <200ms p95 API latency, 50MB max file size
**Constraints**: Local storage only, no external dependencies for photo files
**Scale/Scope**: Single-server deployment, multi-user support via per-user data isolation

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Core Principles Evaluated:**

✅ **CLI Interface**: Project provides CLI tool for user management (create user, list users, reset password)
✅ **Test-First (NON-NEGOTIABLE)**: Need to define testing strategy before implementation
⚠️ **Observability**: Need structured logging and error handling (partially covered in spec)
✅ **Simplicity**: Single binary design minimizes deployment complexity
✅ **Scope Control**: Spec explicitly limits to 5 APIs only

**Gates Status:**
- ✅ PASS: CLI functionality defined
- ✅ PASS: Testing framework selected (Go testing + testify)
- ✅ PASS: Simplicity maintained (single binary, minimal APIs)
- ✅ PASS: Clear scope boundaries defined
- ✅ PASS: Project structure documented
- ✅ PASS: All violations resolved

**Violations Requiring Justification:**
None - all issues resolved in Phase 0

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)
<!--
  ACTION REQUIRED: Replace the placeholder tree below with the concrete layout
  for this feature. Delete unused options and expand the chosen structure with
  real paths (e.g., apps/admin, packages/something). The delivered plan must
  not include Option labels.
-->

```text
# [REMOVE IF UNUSED] Option 1: Single project (DEFAULT)
src/
├── models/
├── services/
├── cli/
└── lib/

tests/
├── contract/
├── integration/
└── unit/

# [REMOVE IF UNUSED] Option 2: Web application (when "frontend" + "backend" detected)
backend/
├── src/
│   ├── models/
│   ├── services/
│   └── api/
└── tests/

frontend/
├── src/
│   ├── components/
│   ├── pages/
│   └── services/
└── tests/

# [REMOVE IF UNUSED] Option 3: Mobile + API (when "iOS/Android" detected)
api/
└── [same as backend above]

ios/ or android/
└── [platform-specific structure: feature modules, UI flows, platform tests]
```

**Structure Decision**: Single project structure (Option 1) - dual-mode binary (API server + CLI tool)

```text
src/
├── cmd/
│   ├── server/          # API server entry point
│   └── cli/             # CLI tool entry point
├── internal/
│   ├── api/             # HTTP handlers and routing
│   │   ├── auth/        # Authentication endpoints
│   │   ├── handlers/    # Request handlers
│   │   └── routes/      # Route definitions
│   ├── service/         # Business logic layer
│   │   ├── auth.go      # Authentication service
│   │   ├── photo.go     # Photo upload service
│   │   └── user.go      # User management service
│   ├── repository/      # Data access layer
│   │   ├── db.go        # Database connection
│   │   ├── user.go      # User repository
│   │   └── photo.go     # Photo repository
│   ├── models/          # Data models
│   │   ├── user.go      # User model
│   │   ├── token.go     # Token model
│   │   └── photo.go     # Photo model
│   └── config/          # Configuration
│       └── config.go    # App configuration
└── storage/             # Photo storage directory (runtime created)

tests/
├── unit/                # Unit tests
├── integration/         # API integration tests
└── contract/            # API contract tests
```

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations to justify - all requirements met with minimal complexity.

## Planning Complete

### Phase 0: Research ✅
- Resolved testing framework selection
- Defined performance goals
- Identified best practices for Gin + GORM + SQLite
- Research documented in `research.md`

### Phase 1: Design & Contracts ✅
- Created data model with entities, relationships, and GORM models
- Generated OpenAPI 3.0 contracts for all 5 APIs
- Documented quickstart guide with examples
- Updated agent context for Claude Code

### Deliverables
- ✅ `plan.md` - This implementation plan
- ✅ `research.md` - Research findings and technology decisions
- ✅ `data-model.md` - Complete data model documentation
- ✅ `quickstart.md` - Developer getting started guide
- ✅ `contracts/` - OpenAPI specifications for all endpoints
  - `login.yaml` - Authentication endpoint
  - `refresh.yaml` - Token refresh endpoint
  - `status.yaml` - Status check endpoint
  - `index.yaml` - Photo indexing endpoint
  - `upload.yaml` - Photo upload endpoint
  - `errors.yaml` - Standardized error responses

### Next Steps
Run `/speckit.tasks` to generate the implementation task list for development.
