# Tasks: Photo Extension Tracking

**Input**: Design documents from `/specs/002-photo-extension-tracking/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Not explicitly requested in feature specification - tests are optional

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**SKIPPED** - Existing photo backup server project already has proper structure

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema changes required by ALL user stories

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T001 Update Photo model in internal/models/photo.go to add UploadedExtensions field
- [x] T002 Restart server to trigger GORM AutoMigrate for new column

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Track Multiple File Extensions (Priority: P1) 🎯 MVP

**Goal**: Enable tracking of multiple uploaded file extensions for each photo

**Independent Test**: Upload two files with the same local_id but different extensions (e.g., jpg and heic), verify both extensions are tracked in database

### Implementation for User Story 1

- [x] T003 [P] [US1] Add GetUploadedExtensions method to internal/repository/photo.go
- [x] T004 [P] [US1] Add AddUploadedExtension method to internal/repository/photo.go
- [x] T005 [US1] Update PhotoService.UploadPhoto in internal/service/photo.go to call AddUploadedExtension
- [ ] T006 [US1] Test extension tracking by uploading same photo with different extensions

**Checkpoint**: At this point, User Story 1 should be fully functional - users can upload multiple formats and they're tracked

---

## Phase 4: User Story 2 - Overwrite Existing Files (Priority: P1)

**Goal**: Always overwrite files on upload without checking if they exist first

**Independent Test**: Upload a file, then upload a different file with same local_id and extension, verify second file completely replaces first

### Implementation for User Story 2

- [x] T007 [P] [US2] Remove file existence check logic from PhotoService.UploadPhoto in internal/service/photo.go
- [x] T008 [P] [US2] Simplify file save logic to always overwrite in internal/service/photo.go
- [ ] T009 [US2] Test overwrite behavior by uploading same file twice with different content

**Checkpoint**: At this point, User Story 2 should be fully functional - re-uploads always overwrite previous files

---

## Phase 5: User Story 3 - View Uploaded Extensions in Index (Priority: P2)

**Goal**: Index API returns list of uploaded extensions for each photo

**Independent Test**: After uploading files, call index API and verify response includes uploaded_extensions field with correct values

### Implementation for User Story 3

- [x] T010 [P] [US3] Update PhotoIndexResponse struct in internal/service/photo.go to include UploadedExtensions field
- [x] T011 [US3] Modify PhotoService.IndexPhotos in internal/service/photo.go to populate UploadedExtensions in response
- [ ] T012 [US3] Test index API returns correct extension lists for photos with 0, 1, and multiple uploads

**Checkpoint**: At this point, User Story 3 should be fully functional - users can see upload status in index response

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements and validation across all user stories

- [ ] T013 [P] Verify backward compatibility with existing photo records (no uploaded_extensions field)
- [ ] T014 [P] Review error handling for JSON parsing in extension list operations
- [ ] T015 Run integration test covering all three user stories end-to-end
- [ ] T016 Validate against contracts in contracts/index-api.md and contracts/upload-api.md

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: SKIPPED - project already exists
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3-5)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (US1 P1 → US2 P1 → US3 P2)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P1)**: Can start after Foundational (Phase 2) - No dependencies on US1 (can be done in parallel)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - No dependencies on US1/US2 (can be done in parallel)

### Within Each User Story

- Models before services
- Services before endpoints
- Core implementation before testing
- Story complete before moving to next priority

### Parallel Opportunities

- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- User Stories 1, 2, and 3 can all start in parallel after Foundational completes
- All tasks within User Story 1 marked [P] can run in parallel
- All tasks within User Story 2 marked [P] can run in parallel
- All tasks within User Story 3 marked [P] can run in parallel

---

## Parallel Example: User Stories 1, 2, and 3

After completing Foundational (Phase 2), all three user stories can be implemented in parallel:

```bash
# Team member 1: User Story 1 (Track Multiple Extensions)
Task: "Add GetUploadedExtensions method to internal/repository/photo.go"
Task: "Add AddUploadedExtension method to internal/repository/photo.go"
Task: "Update PhotoService.UploadPhoto to call AddUploadedExtension"

# Team member 2: User Story 2 (Overwrite Files)
Task: "Remove file existence check from PhotoService.UploadPhoto"
Task: "Simplify file save logic to always overwrite"

# Team member 3: User Story 3 (View Extensions in Index)
Task: "Update PhotoIndexResponse struct to include UploadedExtensions"
Task: "Modify IndexPhotos to populate UploadedExtensions in response"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 2: Foundational (Photo model update + AutoMigrate)
2. Complete Phase 3: User Story 1 (Track Multiple Extensions)
3. **STOP and VALIDATE**: Test extension tracking independently
4. Deploy/demo if ready

### Incremental Delivery

1. Complete Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP with extension tracking!)
3. Add User Story 2 → Test independently → Deploy/Demo (Now with overwrite behavior)
4. Add User Story 3 → Test independently → Deploy/Demo (Full feature with index visibility)
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Complete Foundational together (Photo model update + restart)
2. Once Foundational is done:
   - Developer A: User Story 1 (extension tracking)
   - Developer B: User Story 2 (overwrite behavior)
   - Developer C: User Story 3 (index visibility)
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence

**File Path Summary**:
- Model: `internal/models/photo.go`
- Repository: `internal/repository/photo.go`
- Service: `internal/service/photo.go`
- API Handlers: `internal/api/handlers/photo/*.go` (auto-updated via struct changes)
