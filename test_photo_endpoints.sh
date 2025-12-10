#!/bin/bash

echo "Starting server on port 8082..."
./photo-backup-server --port 8082 &
SERVER_PID=$!

sleep 2

echo "=== Testing login ==="
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8082/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"newpassword123"}')
echo "$LOGIN_RESPONSE" | jq .

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.token')
echo -e "\n=== Token: $TOKEN ===\n"

echo "=== Testing photo index endpoint ==="
curl -s -X POST http://localhost:8082/photos/index \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "date": "2025-12-10",
    "photos": [
      {
        "local_id": "photo1",
        "creation_time": "2025-12-10T10:00:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      },
      {
        "local_id": "photo2",
        "creation_time": "2025-12-10T10:05:00Z",
        "file_extension": "jpg",
        "file_type": "image/jpeg"
      }
    ]
  }' | jq .

echo -e "\n=== Testing photo upload endpoint ==="
echo "Test file content" > /tmp/test_upload.txt
curl -s -X POST http://localhost:8082/photos/upload \
  -H "Authorization: Bearer $TOKEN" \
  -F "local_id=photo1" \
  -F "file_type=image/jpeg" \
  -F "file=@/tmp/test_upload.txt" | jq .

echo -e "\n=== Cleaning up ==="
kill $SERVER_PID 2>/dev/null
wait $SERVER_PID 2>/dev/null
rm -f /tmp/test_upload.txt
echo "Done!"
