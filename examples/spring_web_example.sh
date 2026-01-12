#!/usr/bin/env shode

# Spring-like Web Application Example
# Demonstrates controllers, middleware, and dependency injection

Println "=== Spring-like Web Application Example ==="

# Load configuration
Println "Loading configuration..."
LoadConfig "application.json"

# Get server port from config
port = GetConfigString "server.port" "9188"

# Start HTTP server
Println "Starting HTTP server on port " + port + "..."
StartHTTPServer port
sleep 1

# Define a service (simulated with function)
function UserService() {
    Println "UserService: Finding user..."
    return "User data"
}

# Define a controller handler
function handleGetUsers() {
    # Simulate service call
    result = UserService
    SetHTTPResponse 200 result
}

function handleGetUser() {
    # Get path parameter (simulated)
    userId = GetHTTPQuery "id"
    SetHTTPResponse 200 "User " + userId
}

function handleCreateUser() {
    # Get request body
    body = GetHTTPBody
    SetHTTPResponse 201 "User created: " + body
}

# Register routes
Println "Registering routes..."
RegisterHTTPRoute "GET" "/api/users" "function" "handleGetUsers"
RegisterHTTPRoute "GET" "/api/users/:id" "function" "handleGetUser"
RegisterHTTPRoute "POST" "/api/users" "function" "handleCreateUser"

# Health check route
RegisterHTTPRoute "GET" "/health" "script" "SetHTTPResponse 200 'OK'"

Println ""
Println "=== Spring-like Web Application is running ==="
Println "Server: http://localhost:" + port
Println ""
Println "Available endpoints:"
Println "  GET  /api/users - List all users"
Println "  GET  /api/users/:id - Get user by ID"
Println "  POST /api/users - Create a user"
Println "  GET  /health - Health check"
Println ""
Println "Features demonstrated:"
Println "  - Configuration management"
Println "  - HTTP server with routing"
Println "  - Service layer pattern"
Println "  - Controller handlers"
Println ""
Println "Press Ctrl+C to stop the server"
