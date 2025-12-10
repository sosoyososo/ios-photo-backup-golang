#!/bin/bash

# End-to-End Workflow Test for Photo Backup Server
# Tests complete flow: login → index photos → upload files

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test configuration
SERVER_PORT=8084
SERVER_URL="http://localhost:${SERVER_PORT}"
TEST_USERNAME="e2e_test_user"
TEST_PASSWORD="TestPass123"
TEST_DB="data/e2e_test.db"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    pkill -f "photo-backup-server --port ${SERVER_PORT}" 2>/dev/null || true
    rm -f "${TEST_DB}"
    rm -rf storage/e2e_test
    sleep 2
}

# Setup function
setup() {
    log_info "Setting up test environment..."
    cleanup

    # Create test user via CLI
    log_test "Creating test user '${TEST_USERNAME}'..."
    ./photo-backup-cli user create --username "${TEST_USERNAME}" --password "${TEST_PASSWORD}" --db-path "${TEST_DB}" 2>/dev/null

    # Start server in background
    log_info "Starting server on port ${SERVER_PORT}..."
    ./photo-backup-server --port "${SERVER_PORT}" --db-path "${TEST_DB}" > /tmp/e2e_server.log 2>&1 &
    SERVER_PID=$!
    sleep 3

    # Verify server is running
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        log_error "Server failed to start"
        cat /tmp/e2e_server.log
        exit 1
    fi
    log_info "Server started with PID ${SERVER_PID}"
}

# Test function
run_test() {
    local test_name="$1"
    local test_command="$2"

    log_test "${test_name}..."

    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASS${NC}: ${test_name}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}: ${test_name}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Test 1: User Authentication (Login)
test_login() {
    local response=$(curl -s -X POST "${SERVER_URL}/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"${TEST_USERNAME}\",\"password\":\"${TEST_PASSWORD}\"}")

    TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$TOKEN" ]; then
        log_error "Login failed - no token received"
        echo "$response"
        return 1
    fi

    log_info "Login successful, token received"
    return 0
}

# Test 2: Photo Indexing
test_photo_index() {
    local response=$(curl -s -X POST "${SERVER_URL}/photos/index" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "date": "2025-12-10",
            "photos": [
                {
                    "local_id": "e2e_photo_001",
                    "creation_time": "2025-12-10T10:00:00Z",
                    "file_extension": "jpg",
                    "file_type": "image/jpeg"
                },
                {
                    "local_id": "e2e_photo_002",
                    "creation_time": "2025-12-10T10:05:00Z",
                    "file_extension": "jpg",
                    "file_type": "image/jpeg"
                }
            ]
        }')

    if ! echo "$response" | grep -q '"status":"success"'; then
        log_error "Photo indexing failed"
        echo "$response"
        return 1
    fi

    # Verify filenames were assigned
    if ! echo "$response" | grep -q "IMG_0001.jpg"; then
        log_error "Expected filename IMG_0001.jpg not found"
        echo "$response"
        return 1
    fi

    if ! echo "$response" | grep -q "IMG_0002.jpg"; then
        log_error "Expected filename IMG_0002.jpg not found"
        echo "$response"
        return 1
    fi

    log_info "Photo indexing successful - filenames assigned"
    return 0
}

# Test 3: Photo Upload
test_photo_upload() {
    # Create test file
    echo "E2E Test File Content" > /tmp/e2e_test_file.txt

    local response=$(curl -s -X POST "${SERVER_URL}/photos/upload" \
        -H "Authorization: Bearer ${TOKEN}" \
        -F "local_id=e2e_photo_001" \
        -F "file_type=image/jpeg" \
        -F "file=@/tmp/e2e_test_file.txt")

    if ! echo "$response" | grep -q '"status":"success"'; then
        log_error "Photo upload failed"
        echo "$response"
        return 1
    fi

    log_info "Photo upload successful"
    return 0
}

# Test 4: Verify File Storage
test_file_storage() {
    local file_path="storage/e2e_test/photo/1/2025/12/10/IMG_0001.jpg"

    if [ ! -f "$file_path" ]; then
        log_error "Uploaded file not found at expected location: $file_path"
        ls -la storage/e2e_test/photo/1/2025/12/10/ 2>/dev/null || true
        return 1
    fi

    log_info "File successfully stored at $file_path"
    return 0
}

# Test 5: Verify Database Record
test_database_record() {
    local count=$(sqlite3 "${TEST_DB}" "SELECT COUNT(*) FROM photos_user_1 WHERE local_id='e2e_photo_001';" 2>/dev/null || echo "0")

    if [ "$count" != "1" ]; then
        log_error "Database record not found for e2e_photo_001"
        sqlite3 "${TEST_DB}" "SELECT * FROM photos_user_1;" 2>/dev/null || true
        return 1
    fi

    log_info "Database record verified"
    return 0
}

# Test 6: Re-indexing (should skip existing photos)
test_reindex() {
    local response=$(curl -s -X POST "${SERVER_URL}/photos/index" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d '{
            "date": "2025-12-10",
            "photos": [
                {
                    "local_id": "e2e_photo_001",
                    "creation_time": "2025-12-10T10:00:00Z",
                    "file_extension": "jpg",
                    "file_type": "image/jpeg"
                }
            ]
        }')

    if ! echo "$response" | grep -q "IMG_0001.jpg"; then
        log_error "Re-indexing failed - existing photo not preserved"
        echo "$response"
        return 1
    fi

    log_info "Re-indexing successful - existing photo preserved"
    return 0
}

# Main execution
main() {
    echo "========================================"
    echo "  Photo Backup Server - E2E Workflow Test"
    echo "========================================"
    echo ""

    # Trap to ensure cleanup
    trap cleanup EXIT

    # Setup
    setup

    # Run tests
    echo ""
    log_info "Running end-to-end tests..."
    echo ""

    run_test "T1: User Authentication (Login)" "test_login"
    run_test "T2: Photo Indexing" "test_photo_index"
    run_test "T3: Photo Upload" "test_photo_upload"
    run_test "T4: File Storage Verification" "test_file_storage"
    run_test "T5: Database Record Verification" "test_database_record"
    run_test "T6: Re-indexing (Skip Existing)" "test_reindex"

    # Print summary
    echo ""
    echo "========================================"
    echo "  Test Summary"
    echo "========================================"
    echo -e "Tests Passed: ${GREEN}${TESTS_PASSED}${NC}"
    echo -e "Tests Failed: ${RED}${TESTS_FAILED}${NC}"
    echo "========================================"

    if [ $TESTS_FAILED -eq 0 ]; then
        log_info "All tests passed! ✓"
        return 0
    else
        log_error "Some tests failed! ✗"
        return 1
    fi
}

# Run main
main
