#!/bin/bash

# Manual HTTP Server Test
# This script starts the server in the foreground and runs tests

set -e

echo "=========================================="
echo "Starting Shode HTTP Server..."
echo "=========================================="
echo ""

# Start server in background
./shode run simple_http_test.sh &
SERVER_PID=$!

echo "Server PID: $SERVER_PID"
echo "Waiting for server to start..."
sleep 3

# Check if server is still running
if ! ps -p $SERVER_PID > /dev/null; then
    echo "❌ Server failed to start or exited immediately"
    echo ""
    echo "Checking what happened..."
    wait $SERVER_PID || true
    exit 1
fi

echo "✓ Server is running (PID: $SERVER_PID)"
echo ""

# Test 1: Root endpoint
echo "=========================================="
echo "Test 1: GET /"
echo "=========================================="
curl -v http://localhost:9188/
echo ""

# Test 2: Ping endpoint
echo "=========================================="
echo "Test 2: GET /ping"
echo "=========================================="
curl -v http://localhost:9188/ping
echo ""

# Test 3: Non-existent endpoint
echo "=========================================="
echo "Test 3: GET /nonexistent (should be 404)"
echo "=========================================="
curl -v http://localhost:9188/nonexistent
echo ""

echo ""
echo "=========================================="
echo "Tests completed!"
echo "=========================================="
echo ""
echo "Server is still running. Press Ctrl+C to stop."
echo "Or run: kill $SERVER_PID"

# Keep server running
wait $SERVER_PID
