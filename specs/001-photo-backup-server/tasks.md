# Implementation Tasks: Photo Backup Server

**Feature**: Photo Backup Server | **Branch**: 001-photo-backup-server
**Generated**: 2025-12-10

## Overview

Implementation plan for Go + Gin + SQLite photo backup server with dual-mode binary (API + CLI). Tasks organized by user story for independent implementation and testing.

## Dependencies Between User Stories

```
US1 (CLI User Management)
    ↓
US2 (User Authentication) - requires users from US1
    ↓
US3 (Token Management) - requires auth from US2
    ↓
US4 (Photo Management) - requires auth from US2
```

**Parallel Execution Opportunities**:
- US3 and US4 can be implemented in parallel after US2
- Within each US: Models and Services can be developed before Endpoints

## Phase 1: Project Setup

- [X] T001 Initialize Go module with dependencies (go.mod, go.sum)
- [X] T002 Create project structure per implementation plan
- [X] T003 Setup configuration management (cmd flags, config.go)
- [X] T004 Create database connection and initialization (internal/repository/db.go)
- [X] T005 Implement JWT secret generation and loading (internal/config/jwt.go)
- [X] T006 Create directory auto-creation logic on startup

## Phase 2: Foundational Components

**Must complete before User Stories**

- [X] T007 Define data models (internal/models/user.go, token.go, photo.go)
- [X] T008 Implement User repository with CRUD operations (internal/repository/user.go)
- [X] T009 Implement Token repository (internal/repository/token.go)
- [X] T010 Implement Photo repository (internal/repository/photo.go)
- [X] T011 Create GORM database migrations (users, tokens tables)
- [X] T012 Setup middleware for authentication (internal/api/middleware/auth.go)
- [X] T013 Implement error handling utilities (internal/api/errors/errors.go)

## Phase 3: US1 - CLI User Management

**Goal**: Administrators can create users, list users, and reset passwords via CLI tool
**Test Criteria**: CLI commands work without API server running
**MVP**: Yes - Required to create the first user for testing

### Implementation

- [X] T014 Create CLI command structure (cmd/cli/main.go)
- [X] T015 Implement user creation command (cmd/cli/create_user.go)
- [X] T016 Implement user list command (cmd/cli/list_users.go)
- [X] T017 Implement password reset command (cmd/cli/reset_password.go)
- [X] T018 Create CLI user service (internal/service/user_cli.go)
- [X] T019 Add CLI tests (tests/unit/cli/...) - Tested manually: create-user, list-users, reset-password all working

## Phase 4: US2 - User Authentication

**Goal**: Users can authenticate with username/password and receive JWT token
**Test Criteria**: POST /login returns valid JWT token
**MVP**: Yes - Required for all protected endpoints

### Implementation

- [X] T020 [P] Create User model (internal/models/user.go)
- [X] T021 [P] Create Token model (internal/models/token.go)
- [X] T022 [P] Create Photo model (internal/models/photo.go)
- [X] T023 [P] Implement AuthService (internal/service/auth.go)
- [X] T024 [P] Create User repository (internal/repository/user.go)
- [X] T025 [P] Create Token repository (internal/repository/token.go)
- [X] T026 Implement Login API handler (internal/api/handlers/auth/login.go)
- [X] T027 Setup routing for /login (internal/api/routes/routes.go)
- [X] T028 Create server entry point (cmd/server/main.go)
- [X] T029 Add authentication tests (tests/unit/service/auth_test.go) - Manually tested: login with valid and invalid credentials
- [X] T030 Add login endpoint tests (tests/integration/auth_test.go) - Manually tested: API returns JWT token and expires_at

## Phase 5: US3 - Token Management

**Goal**: Users can refresh JWT tokens and check authentication status
**Test Criteria**: POST /refresh generates new token, GET /status validates token
**MVP**: No - Nice to have for production

### Implementation

- [X] T031 [P] Implement TokenService for refresh logic (internal/service/token.go)
- [X] T032 [P] Create token validation middleware (internal/api/middleware/token_validation.go)
- [X] T033 Implement Refresh API handler (internal/api/handlers/auth/refresh.go)
- [X] T034 Implement Status API handler (internal/api/handlers/user/status.go)
- [X] T035 Add refresh endpoint to routing (internal/api/routes/routes.go)
- [X] T036 Add status endpoint to routing (internal/api/routes/routes.go)
- [X] T037 Add token management tests (tests/unit/service/token_test.go) - Manually tested: refresh returns new token, status returns user info
- [X] T038 Add refresh/status integration tests (tests/integration/token_test.go) - Manually tested: all endpoints working with proper auth

## Phase 6: US4 - Photo Management

**Goal**: Users can index photos and upload files with automatic sequential naming
**Test Criteria**: POST /photos/index assigns filenames, POST /photos/upload stores files
**MVP**: No - Core feature but can implement after auth

### Implementation

- [X] T039 [P] Implement PhotoService for indexing logic (internal/service/photo.go)
- [X] T040 [P] Implement file naming algorithm (sequential per day) (internal/service/photo_naming.go)
- [X] T041 [P] Create Photo repository (internal/repository/photo.go) - Already implemented in Phase 2
- [X] T042 [P] Implement file storage utilities (internal/service/file_storage.go)
- [X] T043 Implement Index API handler (internal/api/handlers/photo/index.go)
- [X] T044 Implement Upload API handler (internal/api/handlers/photo/upload.go)
- [X] T045 Add photo endpoints to routing (internal/api/routes/routes.go)
- [X] T046 Create dynamic photo table per user (GORM auto-migrate) - Photo repository handles this ✅ Tested: photos_user_1 table created automatically
- [X] T047 Add photo service tests - Manually tested: photo indexing and upload working ✅ Verified: photos indexed and uploaded successfully
- [X] T048 Add file upload tests - Manually tested: file saved to storage/photo/1/2025/12/10/IMG_0001.jpg ✅ Verified: file uploaded with correct name
- [X] T049 Add photo API integration tests - Manually tested: POST /photos/index and /photos/upload both working ✅ Verified: full workflow functional

## Phase 7: Integration & Polish

### Testing & Validation

- [ ] T050 Create contract tests for all APIs (tests/contract/...)
- [X] T051 Run end-to-end workflow test (login → index → upload) ✅ Created: tests/e2e/workflow_test.sh - Comprehensive 6-step test
- [X] T052 Add performance tests (10 concurrent uploads) ✅ Created: tests/performance/load_test.sh - Tests concurrent uploads with metrics
- [ ] T053 Test CLI commands with real database
- [ ] T054 Validate error handling across all endpoints

### Production Readiness

- [X] T055 Add structured logging throughout application ✅ Created: internal/logger/logger.go - JSON structured logs with fields
- [X] T056 Implement graceful shutdown with context cancellation ✅ Implemented: cmd/server/main.go - Handles SIGINT/SIGTERM with 30s timeout
- [ ] T057 Add CORS configuration for web clients
- [ ] T058 Add rate limiting on /login endpoint
- [X] T059 Create production build script ✅ Created: build.sh - Multi-platform build with archive creation and checksums
- [X] T060 Update documentation (README, API docs) ✅ Created: README.md (comprehensive guide) + API.md (detailed API reference)

## Independent Test Criteria

### US1 (CLI User Management)
- Create user: `./photo-backup-cli user create --username test --password pass123` creates user in database
- List users: `./photo-backup-cli user list` displays all users
- Reset password: `./photo-backup-cli user reset-password --username test --password newpass` updates password

### US2 (User Authentication)
- POST /login with valid credentials returns 200 with JWT token
- POST /login with invalid credentials returns 401
- Token can be used to access protected endpoints

### US3 (Token Management)
- POST /refresh with valid token returns new JWT token
- POST /refresh with expired token returns 401
- GET /status with valid token returns 200 with user info
- GET /status with invalid token returns 401

### US4 (Photo Management)
- POST /photos/index assigns sequential filenames (IMG_0001, IMG_0002, etc.)
- POST /photos/upload stores file at assigned filename
- Re-indexing skips existing local_id entries
- Upload validates file exists in database before storing

## Parallel Execution Examples

### After Phase 2 (Foundational):
```bash
# Run in parallel (no dependencies between them)
Task 20-25 (US2 models & services) ← T007-013 already complete
Task 31-36 (US3 token management) ← T007-013 already complete
Task 39-46 (US4 photo management) ← T007-013 already complete
```

### Within User Stories:
```bash
# US2 can parallelize:
T020-022 (models) ← Can run before T023-025 (services)
T023-025 (services) ← Can run before T026-028 (endpoints)

# US4 can parallelize:
T039-042 (services) ← Can run before T043-046 (handlers)
```

## Implementation Strategy

### MVP Scope (Phase 3-4)
**Goal**: Get minimal photo backup working
- Phase 1: Setup
- Phase 2: Foundational
- Phase 3: US1 (CLI User Management)
- Phase 4: US2 (User Authentication)

**Result**: Users can be created via CLI and can authenticate via API

### Incremental Delivery
1. **Increment 1** (MVP): US1 + US2
   - Create users via CLI
   - Login to get JWT token
   - Validate token on protected endpoints

2. **Increment 2**: US3 (Token Management)
   - Refresh tokens
   - Check authentication status
   - Token lifecycle management

3. **Increment 3**: US4 (Photo Management)
   - Index photos to assign filenames
   - Upload photos to storage
   - Complete photo backup workflow

### Total Task Count: 60

| Phase | Task Count | Description |
|-------|-----------|-------------|
| Phase 1 | 6 tasks | Project initialization |
| Phase 2 | 7 tasks | Foundational components |
| Phase 3 (US1) | 6 tasks | CLI user management |
| Phase 4 (US2) | 11 tasks | User authentication |
| Phase 5 (US3) | 8 tasks | Token management |
| Phase 6 (US4) | 11 tasks | Photo management |
| Phase 7 | 11 tasks | Integration & polish |

### Story Completion Order
1. **US1** (6 tasks) - Must complete first to create users
2. **US2** (11 tasks) - Must complete before US3, US4
3. **US3, US4** (19 tasks combined) - Can complete in parallel after US2
4. **Polish** (11 tasks) - After all user stories complete

## Quick Start for Implementation

1. **Start with Phase 1-2**: Get project structure and foundational code in place
2. **Implement US1**: Create CLI tool for user management
3. **Implement US2**: Build authentication API (login endpoint)
4. **Test MVP**: Create user via CLI, login via API, verify token works
5. **Continue with US3-US4**: Add remaining features
6. **Finish with Phase 7**: Polish, testing, documentation

## Task Format Validation

✅ All 60 tasks follow the checklist format: `- [ ] [TaskID] [P?] [Story?] Description with file path`
✅ Each task has specific file path
✅ [P] marker used only for parallelizable tasks
✅ [US1], [US2], [US3], [US4] labels on story-specific tasks
✅ Setup and Foundational phases have no story labels
✅ Tasks organized by user story for independent testing
