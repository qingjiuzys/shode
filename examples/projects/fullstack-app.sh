#!/usr/bin/env shode
# Full-Stack Application Example
# Demonstrates a complete web application with frontend and API backend

StartHTTPServer "4000"

# Serve the frontend SPA (Single Page Application)
# All non-existent routes will fallback to index.html for client-side routing
RegisterStaticRouteAdvanced "/" "./frontend/dist" \
    "index.html" \
    "false" \
    "max-age=3600" \
    "true" \
    "index.html"

# API: Get all users
function getUsers() {
    SetHTTPResponse 200 '[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]'
}
RegisterHTTPRoute "GET" "/api/users" "function" "getUsers"

# API: Get single user
function getUser() {
    SetHTTPResponse 200 '{"id":1,"name":"Alice","email":"alice@example.com"}'
}
RegisterHTTPRoute "GET" "/api/users/1" "function" "getUser"

# API: Create user
function createUser() {
    SetHTTPResponse 201 '{"id":3,"name":"Charlie","status":"created"}'
}
RegisterHTTPRoute "POST" "/api/users" "function" "createUser"

# API: Health check
function healthCheck() {
    SetHTTPResponse 200 '{"status":"healthy","version":"1.0.0","timestamp":"2026-01-27T10:00:00Z"}'
}
RegisterHTTPRoute "GET" "/api/health" "function" "healthCheck"

Println "🌟 Full-Stack Application running at http://localhost:4000"
Println "✨ Frontend SPA served from /"
Println "🔌 API endpoints:"
Println "   GET    /api/users       - List all users"
Println "   GET    /api/users/1     - Get user by ID"
Println "   POST   /api/users       - Create new user"
Println "   GET    /api/health      - Health check"

for i in $(seq 1 100000); do sleep 1; done
