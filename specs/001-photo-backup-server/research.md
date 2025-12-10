# Research: Photo Backup Server

## Research Summary

This document captures research findings to resolve ambiguities in the Technical Context and identify best practices for the Go + Gin + SQLite photo backup server.

## Go Testing Framework Selection

**Decision**: Use Go's built-in testing package with testify for assertions

**Rationale**:
- Go's built-in `testing` package is the standard and well-supported
- `testify/assert` provides better assertions and failure reporting
- No external test runner required (uses `go test`)
- Gin provides `net/http/httptest` for HTTP testing
- GORM provides test utilities for database testing

**Implementation Strategy**:
1. Unit tests: `testing` package + `testify/assert`
2. Integration tests: Use Docker containers or in-memory SQLite for DB tests
3. API contract tests: httptest + custom response matchers
4. Test structure: Table-driven tests (Go best practice)

**Testing Structure**:
```
tests/
├── unit/           # Pure functions, services
├── integration/    # DB operations, API endpoints
└── contract/       # API response validation
```

## Performance Goals

**Decision**: Define performance targets for single-server deployment

**Rationale**: Photo backup servers handle large file uploads and sequential naming queries

**Performance Targets**:
- **Throughput**: Support 10 concurrent photo uploads
- **Latency**: <200ms p95 for API responses (excluding file upload)
- **File Upload**: Support files up to 50MB (configurable)
- **Database**: Efficient queries for sequential numbering (max 1000 photos/day)
- **Concurrent Users**: Support 100 registered users on single server

**Measurement**:
- Use Go's built-in pprof for profiling
- Add request duration logging
- Monitor SQLite query performance

## Best Practices for Gin + GORM + SQLite

**Database Pattern**:
- Use connection pooling (GORM handles this)
- Separate read/write operations if needed (not for single-user scenario)
- Use transactions for multi-step operations (photo index + upload)

**Gin Best Practices**:
- Use Gin middleware for authentication, logging, CORS
- Structured error responses (already defined in spec)
- Graceful shutdown with context cancellation
- Use `gin.Mode(gin.ReleaseMode)` for production

**JWT Implementation**:
- Use `github.com/golang-jwt/jwt/v5` (latest stable)
- Store tokens in SQLite (already specified)
- Implement token rotation on refresh (delete old, create new)

**Project Structure**:
- `internal/` package pattern (prevents external imports)
- `cmd/` for binaries
- Clear separation: handlers → services → repository → models

**Configuration**:
- Use `github.com/spf13/viper` or `github.com/urfave/cli/v2` for flags
- Environment variables with fallbacks
- Initialize directories and database on startup (already clarified)

## Security Considerations

**Implemented in Spec**:
- bcrypt password hashing ✅
- JWT token authentication ✅
- Bearer token authorization ✅

**Additional Recommendations**:
- Rate limiting on /login endpoint (prevent brute force)
- CORS configuration for web clients
- Input validation on all endpoints
- Secure file upload (validate file type, size limits)

## Technology Stack Confirmation

**All technologies confirmed**:
- Go (latest stable)
- Gin (HTTP framework)
- GORM (ORM)
- SQLite (database)
- JWT (authentication)
- bcrypt (password hashing)

**No additional dependencies required** beyond what's in the spec.

## Resolution of NEEDS CLARIFICATION

✅ **Testing**: Go testing package + testify for assertions
✅ **Performance**: Targets defined for concurrent uploads, API latency, file size
✅ **All ambiguities resolved** - ready for Phase 1 design
