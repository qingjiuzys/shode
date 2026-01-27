#!/bin/bash

# ========================================
# Automated HTTP Server Test Runner
# ========================================
# This script automatically tests all HTTP server features

set -e

PORT=9188
BASE_URL="http://localhost:$PORT"
SERVER_PID=""
TESTS_PASSED=0
TESTS_FAILED=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test functions
test_start() {
    echo -e "${BLUE}Testing: $1${NC}"
}

test_pass() {
    echo -e "${GREEN}✓ PASS: $1${NC}"
    ((TESTS_PASSED++))
}

test_fail() {
    echo -e "${RED}✗ FAIL: $1${NC}"
    echo -e "${RED}  Expected: $2${NC}"
    echo -e "${RED}  Got: $3${NC}"
    ((TESTS_FAILED++))
}

test_info() {
    echo -e "${YELLOW}ℹ INFO: $1${NC}"
}

# Cleanup function
cleanup() {
    if [ -n "$SERVER_PID" ]; then
        echo ""
        test_info "Stopping server (PID: $SERVER_PID)..."
        kill $SERVER_PID 2>/dev/null || true
        wait $SERVER_PID 2>/dev/null || true
        echo "Server stopped"
    fi
}

# Set trap to cleanup on exit
trap cleanup EXIT INT TERM

# Wait for server to be ready
wait_for_server() {
    local max_attempts=30
    local attempt=0

    echo "Waiting for server to start..."
    while [ $attempt -lt $max_attempts ]; do
        if curl -s "$BASE_URL/" > /dev/null 2>&1; then
            echo "Server is ready!"
            return 0
        fi
        ((attempt++))
        sleep 0.5
    done

    echo "Failed to start server"
    return 1
}

# ========================================
# Start HTTP Server
# ========================================
echo "=========================================="
echo "Shode HTTP Server Automated Test Suite"
echo "=========================================="
echo ""

test_start "Starting Shode HTTP server on port $PORT..."
./shode run test_http_complete.sh > /tmp/shode_server.log 2>&1 &
SERVER_PID=$!

sleep 2

# Check if server started
if ! ps -p $SERVER_PID > /dev/null; then
    test_fail "Server start" "Server process running" "Process died"
    cat /tmp/shode_server.log
    exit 1
fi

if ! wait_for_server; then
    test_fail "Server readiness" "Server to respond" "Timeout"
    cat /tmp/shode_server.log
    exit 1
fi

test_pass "Server started"
echo ""

# ========================================
# Test Suite
# ========================================

echo "=========================================="
echo "Running HTTP Tests"
echo "=========================================="
echo ""

# Test 1: Welcome endpoint
test_start "GET / - Welcome message"
RESULT=$(curl -s "$BASE_URL/")
if [ "$RESULT" = "Welcome to Shode HTTP Server" ]; then
    test_pass "GET /" "Welcome message" "$RESULT"
else
    test_fail "GET /" "Welcome to Shode HTTP Server" "$RESULT"
fi
echo ""

# Test 2: Hello with query parameter
test_start "GET /hello?name=Alice - Query parameter"
RESULT=$(curl -s "$BASE_URL/hello?name=Alice")
if [ "$RESULT" = "Hello, Alice!" ]; then
    test_pass "GET /hello?name=Alice" "Hello, Alice!" "$RESULT"
else
    test_fail "GET /hello?name=Alice" "Hello, Alice!" "$RESULT"
fi
echo ""

# Test 3: Hello without query parameter (default value)
test_start "GET /hello - Default query parameter"
RESULT=$(curl -s "$BASE_URL/hello")
if [ "$RESULT" = "Hello, World!" ]; then
    test_pass "GET /hello (default)" "Hello, World!" "$RESULT"
else
    test_fail "GET /hello (default)" "Hello, World!" "$RESULT"
fi
echo ""

# Test 4: POST echo
test_start "POST /echo - Echo request body"
RESULT=$(curl -s -X POST "$BASE_URL/echo" -d "Hello from curl")
if [[ "$RESULT" == *"Method: POST"* ]] && [[ "$RESULT" == *"Body: Hello from curl"* ]]; then
    test_pass "POST /echo" "Echo with body" "$RESULT"
else
    test_fail "POST /echo" "Echo with method and body" "$RESULT"
fi
echo ""

# Test 5: PUT update
test_start "PUT /update - Update endpoint"
RESULT=$(curl -s -X PUT "$BASE_URL/update" -d "Update data")
if [ "$RESULT" = "PUT received: Update data" ]; then
    test_pass "PUT /update" "PUT received: Update data" "$RESULT"
else
    test_fail "PUT /update" "PUT received: Update data" "$RESULT"
fi
echo ""

# Test 6: DELETE with query parameter
test_start "DELETE /delete?id=123 - Delete with ID"
RESULT=$(curl -s -X DELETE "$BASE_URL/delete?id=123")
if [ "$RESULT" = "Deleted item with ID: 123" ]; then
    test_pass "DELETE /delete?id=123" "Deleted item with ID: 123" "$RESULT"
else
    test_fail "DELETE /delete?id=123" "Deleted item with ID: 123" "$RESULT"
fi
echo ""

# Test 7: Headers endpoint
test_start "GET /headers - Request headers"
RESULT=$(curl -s "$BASE_URL/headers")
if [[ "$RESULT" == *"User-Agent:"* ]] && [[ "$RESULT" == *"Accept:"* ]]; then
    test_pass "GET /headers" "Headers returned" "$RESULT"
else
    test_fail "GET /headers" "Headers with User-Agent and Accept" "$RESULT"
fi
echo ""

# Test 8: Custom header
test_start "GET /headers - Custom header"
RESULT=$(curl -s -H "X-Custom-Header: TestValue" "$BASE_URL/headers")
if [[ "$RESULT" == *"X-Custom-Header: TestValue"* ]]; then
    test_pass "GET /headers with custom header" "Custom header received" "$RESULT"
else
    test_fail "GET /headers with custom header" "X-Custom-Header: TestValue" "$RESULT"
fi
echo ""

# Test 9: JSON API GET
test_start "GET /api/status - JSON API"
RESULT=$(curl -s "$BASE_URL/api/status")
if [[ "$RESULT" == *"\"status\":\"success\""* ]] && [[ "$RESULT" == *"\"data\""* ]]; then
    test_pass "GET /api/status" "JSON response" "$RESULT"
else
    test_fail "GET /api/status" "JSON with status:success" "$RESULT"
fi
echo ""

# Test 10: JSON API POST
test_start "POST /api/status - JSON POST"
RESULT=$(curl -s -X POST "$BASE_URL/api/status" -d '{"test":"data"}')
if [[ "$RESULT" == *"\"status\":\"created\""* ]] && [[ "$RESULT" == *"\"received\""* ]]; then
    test_pass "POST /api/status" "JSON created response" "$RESULT"
else
    test_fail "POST /api/status" "JSON with status:created" "$RESULT"
fi
echo ""

# Test 11: Ping
test_start "GET /ping - Ping/pong"
RESULT=$(curl -s "$BASE_URL/ping")
if [ "$RESULT" = "pong" ]; then
    test_pass "GET /ping" "pong" "$RESULT"
else
    test_fail "GET /ping" "pong" "$RESULT"
fi
echo ""

# Test 12: 404 error
test_start "GET /error/404 - 404 error"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/error/404")
if [ "$HTTP_CODE" = "404" ]; then
    test_pass "GET /error/404" "404 status code" "$HTTP_CODE"
else
    test_fail "GET /error/404" "404" "$HTTP_CODE"
fi
echo ""

# Test 13: 400 error
test_start "GET /error/400 - 400 error"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/error/400")
if [ "$HTTP_CODE" = "400" ]; then
    test_pass "GET /error/400" "400 status code" "$HTTP_CODE"
else
    test_fail "GET /error/400" "400" "$HTTP_CODE"
fi
echo ""

# Test 14: Non-existent route (should return 404)
test_start "GET /nonexistent - Non-existent route"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/nonexistent")
if [ "$HTTP_CODE" = "404" ]; then
    test_pass "GET /nonexistent" "404 for unknown route" "$HTTP_CODE"
else
    test_fail "GET /nonexistent" "404" "$HTTP_CODE"
fi
echo ""

# Test 15: Response headers
test_start "Response Content-Type header"
CONTENT_TYPE=$(curl -s -I "$BASE_URL/api/status" | grep -i "Content-Type" | cut -d' ' -f2 | tr -d '\r')
if [[ "$CONTENT_TYPE" == *"application/json"* ]]; then
    test_pass "Content-Type header" "application/json" "$CONTENT_TYPE"
else
    test_fail "Content-Type header" "application/json" "$CONTENT_TYPE"
fi
echo ""

# ========================================
# Test Summary
# ========================================
echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo ""
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
TOTAL_TESTS=$((TESTS_PASSED + TESTS_FAILED))
echo "Total Tests:  $TOTAL_TESTS"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}=========================================="
    echo "✓ All tests passed!"
    echo "==========================================${NC}"
    exit 0
else
    echo -e "${RED}=========================================="
    echo "✗ Some tests failed"
    echo "==========================================${NC}"
    echo ""
    echo "Server log:"
    cat /tmp/shode_server.log
    exit 1
fi
