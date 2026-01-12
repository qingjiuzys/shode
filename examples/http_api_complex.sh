#!/usr/bin/env shode

# Complex HTTP API Example
# Demonstrates HTTP methods, request context, cache, and database operations

# Start HTTP server
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"
sleep 1

# Define handler functions
function handleGetUsers() {
    # Get users from cache or database
    cacheKey = "users:list"
    cached = GetCache cacheKey
    if cached != "" {
        SetHTTPResponse 200 cached
        return
    }
    
    # Query database
    QueryDB "SELECT id, name, email FROM users"
    result = GetQueryResult
    SetCache cacheKey result 300
    SetHTTPResponse 200 result
}

function handleCreateUser() {
    # Get request body
    body = GetHTTPBody
    # Parse body and insert into database
    # For demo: simple insert
    ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "John Doe" "john@example.com"
    SetHTTPResponse 201 "User created"
}

function handleGetUser() {
    # Get user ID from path or query
    userId = GetHTTPQuery "id"
    cacheKey = "user:" + userId
    cached = GetCache cacheKey
    if cached != "" {
        SetHTTPResponse 200 cached
        return
    }
    
    QueryRowDB "SELECT id, name, email FROM users WHERE id = ?" userId
    result = GetQueryResult
    SetCache cacheKey result 300
    SetHTTPResponse 200 result
}

# Register routes
RegisterHTTPRoute "GET" "/api/users" "function" "handleGetUsers"
RegisterHTTPRoute "POST" "/api/users" "function" "handleCreateUser"
RegisterHTTPRoute "GET" "/api/users/:id" "function" "handleGetUser"

# Simple script-based route
RegisterHTTPRoute "GET" "/api/health" "script" "SetHTTPResponse 200 'OK'"

Println "HTTP API server is running on http://localhost:9188"
Println "Available endpoints:"
Println "  GET  /api/users - List all users"
Println "  POST /api/users - Create a user"
Println "  GET  /api/users/:id - Get user by ID"
Println "  GET  /api/health - Health check"
