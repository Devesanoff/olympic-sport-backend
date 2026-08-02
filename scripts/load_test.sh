#!/bin/bash
set -e

# Load testing script using Vegeta Docker container.
# Simulates 200+ concurrent requests to /api/scan/access.

echo "=========================================================="
echo "Starting Accreditation System Performance Load Test"
echo "=========================================================="

# 1. Obtain JWT access token
echo "Obtaining JWT access token for scanner..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"scanner@olympic.com","password":"scanner123"}')

JWT_TOKEN=$(echo "$TOKEN_RESPONSE" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

if [ -z "$JWT_TOKEN" ]; then
  echo "Error: Failed to obtain JWT token. Server response was:"
  echo "$TOKEN_RESPONSE"
  exit 1
fi

echo "Successfully logged in. Token length: ${#JWT_TOKEN} characters."

# 2. Get Seeded QR Token
if [ ! -f "test_qr_token.txt" ]; then
  echo "Error: test_qr_token.txt not found. Please run the seeder first."
  exit 1
fi
QR_TOKEN=$(cat test_qr_token.txt)
echo "Using seeded QR token: $QR_TOKEN"

# 3. Create Vegeta Target Files
echo "Creating Vegeta target configuration..."
echo "{\"qr_token\":\"$QR_TOKEN\",\"zone_id\":1,\"direction\":\"IN\"}" > target_body.json

cat <<EOF > targets.txt
POST http://localhost:8080/api/scan/access
Content-Type: application/json
Authorization: Bearer $JWT_TOKEN
@target_body.json
EOF

# 4. Run Vegeta attack using local binary
# rate=200 means 200 requests per second.
# duration=5s runs the test for 5 seconds (total 1000 requests).
echo "Running load test (200 requests/sec for 5 seconds)..."
./vegeta attack -targets=targets.txt -rate=200 -duration=5s > results.bin

# 5. Display Vegeta Report
echo "=========================================================="
echo "Performance Results Report"
echo "=========================================================="
./vegeta report results.bin

# Clean up temp files
rm -f target_body.json targets.txt results.bin
echo "=========================================================="
echo "Load Test Completed!"
echo "=========================================================="
