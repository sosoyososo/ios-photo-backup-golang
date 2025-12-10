#!/bin/bash

echo "Testing Photo Backup Server with new features..."
echo ""

# Test 1: Start server
echo "1. Starting server..."
./photo-backup-server --port 8086 > /tmp/server_test.log 2>&1 &
SERVER_PID=$!
sleep 3

# Test 2: Health check
echo "2. Testing health endpoint..."
curl -s http://localhost:8086/health
echo ""
echo ""

# Test 3: Login
echo "3. Testing login..."
TOKEN=$(curl -s -X POST http://localhost:8086/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"newpassword123"}' | jq -r '.token')
echo "Token: ${TOKEN:0:50}..."
echo ""

# Test 4: Check logs
echo "4. Checking structured logs..."
echo "=== Server startup logs ==="
cat /tmp/server_test.log | grep -E "(Starting|Configuration|Application|Database|Routes|Server)" | head -10
echo ""

# Test 5: Graceful shutdown
echo "5. Testing graceful shutdown..."
kill -TERM $SERVER_PID
sleep 2
cat /tmp/server_test.log | grep -E "(shutdown|graceful)" | tail -5
echo ""

# Cleanup
pkill -f "photo-backup-server --port 8086" 2>/dev/null

echo "Test completed!"
