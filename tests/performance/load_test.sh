#!/bin/bash

# Performance Test for Photo Backup Server
# Tests 10 concurrent uploads to verify server can handle multiple simultaneous requests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
SERVER_PORT=8085
SERVER_URL="http://localhost:${SERVER_PORT}"
NUM_CONCURRENT=10
TEST_DB="data/performance_test.db"

# Performance metrics
START_TIME=$(date +%s.%N)
END_TIME=0
TOTAL_DURATION=0
SUCCESS_COUNT=0
FAIL_COUNT=0
declare -a RESPONSE_TIMES

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

log_perf() {
    echo -e "${BLUE}[PERF]${NC} $1"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up test environment..."
    pkill -f "photo-backup-server --port ${SERVER_PORT}" 2>/dev/null || true
    rm -f "${TEST_DB}"
    rm -rf storage/performance_test
    sleep 2
}

# Setup function
setup() {
    log_info "Setting up performance test environment..."

    # Create performance test user
    ./photo-backup-cli user create --username "perf_test_user" --password "PerfTest123" --db-path "${TEST_DB}" 2>/dev/null || true

    # Start server
    log_info "Starting server on port ${SERVER_PORT}..."
    ./photo-backup-server --port "${SERVER_PORT}" --db-path "${TEST_DB}" > /tmp/perf_server.log 2>&1 &
    SERVER_PID=$!
    sleep 3

    # Verify server is running
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        log_error "Server failed to start"
        cat /tmp/perf_server.log
        exit 1
    fi

    log_info "Server started with PID ${SERVER_PID}"
}

# Get authentication token
get_token() {
    local response=$(curl -s -X POST "${SERVER_URL}/login" \
        -H "Content-Type: application/json" \
        -d '{"username":"perf_test_user","password":"PerfTest123"}')

    TOKEN=$(echo "$response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

    if [ -z "$TOKEN" ]; then
        log_error "Failed to get authentication token"
        echo "$response"
        return 1
    fi

    return 0
}

# Single upload test
single_upload() {
    local upload_id=$1
    local upload_start=$(date +%s.%N)

    # Create unique test file
    echo "Performance test file ${upload_id} - $(date)" > "/tmp/perf_test_${upload_id}.txt"

    # Upload file
    local response=$(curl -s -w "\n%{http_code}" -X POST "${SERVER_URL}/photos/upload" \
        -H "Authorization: Bearer ${TOKEN}" \
        -F "local_id=perf_photo_${upload_id}" \
        -F "file_type=image/jpeg" \
        -F "file=@/tmp/perf_test_${upload_id}.txt")

    local http_code=$(echo "$response" | tail -n1)
    local upload_end=$(date +%s.%N)
    local duration=$(echo "$upload_end - $upload_start" | bc)

    # Store response time
    RESPONSE_TIMES[$upload_id]=$duration

    if [ "$http_code" = "200" ] && echo "$response" | grep -q '"status":"success"'; then
        ((SUCCESS_COUNT++))
        return 0
    else
        ((FAIL_COUNT++))
        log_error "Upload ${upload_id} failed with HTTP ${http_code}"
        return 1
    fi
}

# Worker function for concurrent uploads
worker() {
    local worker_id=$1
    local start_photo=$((worker_id * (NUM_CONCURRENT / 4) + 1))
    local end_photo=$((start_photo + (NUM_CONCURRENT / 4) - 1))

    for i in $(seq $start_photo $end_photo); do
        single_upload $i
    done
}

# Calculate statistics
calculate_stats() {
    local total=0
    local count=${#RESPONSE_TIMES[@]}
    local min=999999
    local max=0
    local avg=0

    for time in "${RESPONSE_TIMES[@]}"; do
        total=$(echo "$time + $total" | bc)
        if (( $(echo "$time < $min" | bc -l) )); then
            min=$time
        fi
        if (( $(echo "$time > $max" | bc) )); then
            max=$time
        fi
    done

    if [ $count -gt 0 ]; then
        avg=$(echo "scale=3; $total / $count" | bc)
    fi

    echo "$avg|$min|$max"
}

# Main performance test
run_performance_test() {
    log_test "Running performance test with ${NUM_CONCURRENT} concurrent uploads..."

    # Index photos first
    log_info "Indexing ${NUM_CONCURRENT} photos..."
    local index_payload="{"
    index_payload+='"date":"2025-12-10","photos":['
    for i in $(seq 1 $NUM_CONCURRENT); do
        index_payload+='{"local_id":"perf_photo_'$i'","creation_time":"2025-12-10T10:00:00Z","file_extension":"jpg","file_type":"image/jpeg"}'
        if [ $i -lt $NUM_CONCURRENT ]; then
            index_payload+=","
        fi
    done
    index_payload+=']}'

    local index_response=$(curl -s -X POST "${SERVER_URL}/photos/index" \
        -H "Authorization: Bearer ${TOKEN}" \
        -H "Content-Type: application/json" \
        -d "$index_payload")

    if ! echo "$index_response" | grep -q '"status":"success"'; then
        log_error "Photo indexing failed"
        echo "$index_response"
        return 1
    fi

    log_info "Photo indexing completed"

    # Run concurrent uploads
    log_info "Starting ${NUM_CONCURRENT} concurrent uploads..."

    # Launch background workers
    for worker_id in 0 1 2 3; do
        worker $worker_id &
    done

    # Wait for all workers
    wait

    END_TIME=$(date +%s.%N)
    TOTAL_DURATION=$(echo "$END_TIME - $START_TIME" | bc)

    return 0
}

# Print performance report
print_report() {
    echo ""
    echo "========================================"
    echo "  Performance Test Report"
    echo "========================================"
    echo ""
    log_perf "Configuration:"
    echo "  - Concurrent Uploads: ${NUM_CONCURRENT}"
    echo "  - Test Database: ${TEST_DB}"
    echo "  - Server Port: ${SERVER_PORT}"
    echo ""
    log_perf "Results:"
    echo "  - Successful Uploads: ${GREEN}${SUCCESS_COUNT}${NC}"
    echo "  - Failed Uploads: ${RED}${FAIL_COUNT}${NC}"
    echo "  - Total Duration: ${TOTAL_DURATION}s"
    echo ""

    # Calculate and display statistics
    local stats=$(calculate_stats)
    local avg=$(echo $stats | cut -d'|' -f1)
    local min=$(echo $stats | cut -d'|' -f2)
    local max=$(echo $stats | cut -d'|' -f3)

    log_perf "Response Time Statistics:"
    echo "  - Average: ${avg}s"
    echo "  - Minimum: ${min}s"
    echo "  - Maximum: ${max}s"
    echo ""

    # Calculate throughput
    local throughput=$(echo "scale=2; $SUCCESS_COUNT / $TOTAL_DURATION" | bc)
    log_perf "Throughput: ${throughput} uploads/second"
    echo ""

    # Success rate
    local success_rate=$(echo "scale=1; $SUCCESS_COUNT * 100 / $NUM_CONCURRENT" | bc)
    echo "  - Success Rate: ${success_rate}%"
    echo ""

    # Benchmark criteria
    echo "========================================"
    echo "  Benchmark Criteria"
    echo "========================================"
    echo ""

    local benchmark_pass=true

    if [ $FAIL_COUNT -gt 0 ]; then
        echo -e "${RED}✗${NC} All uploads should succeed"
        benchmark_pass=false
    else
        echo -e "${GREEN}✓${NC} All uploads succeeded"
    fi

    if (( $(echo "$avg > 5.0" | bc -l) )); then
        echo -e "${RED}✗${NC} Average response time should be < 5s (got ${avg}s)"
        benchmark_pass=false
    else
        echo -e "${GREEN}✓${NC} Average response time is acceptable (< 5s)"
    fi

    if (( $(echo "$throughput < 2.0" | bc -l) )); then
        echo -e "${RED}✗${NC} Throughput should be > 2 uploads/second (got ${throughput})"
        benchmark_pass=false
    else
        echo -e "${GREEN}✓${NC} Throughput is acceptable (> 2 uploads/sec)"
    fi

    echo ""
    echo "========================================"

    if [ "$benchmark_pass" = true ]; then
        log_info "Performance test passed! ✓"
        return 0
    else
        log_error "Performance test failed! ✗"
        return 1
    fi
}

# Main execution
main() {
    echo "========================================"
    echo "  Photo Backup Server - Performance Test"
    echo "========================================"
    echo ""

    # Trap to ensure cleanup
    trap cleanup EXIT

    # Setup
    setup

    # Get authentication token
    log_info "Authenticating..."
    if ! get_token; then
        log_error "Authentication failed"
        exit 1
    fi
    log_info "Authentication successful"

    # Run performance test
    echo ""
    if ! run_performance_test; then
        log_error "Performance test execution failed"
        exit 1
    fi

    # Print report
    print_report
}

# Run main
main
