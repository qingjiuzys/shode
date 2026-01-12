#!/usr/bin/env shode

# Complete Spring-like Application Example
# Demonstrates all Spring features: IoC, Config, Web, Transaction, Repository

Println "=== Complete Spring-like Application ==="

# Step 1: Load Configuration
Println "Step 1: Loading configuration..."
configContent = '{
  "server": {
    "port": 9188,
    "host": "localhost"
  },
  "database": {
    "url": "sqlite:app.db"
  },
  "cache": {
    "enabled": true,
    "ttl": 300
  }
}'
WriteFile "test/tmp/application.json" configContent
LoadConfig "test/tmp/application.json"
Println "Configuration loaded"

# Step 2: Connect to Database
Println ""
Println "Step 2: Setting up database..."
port = GetConfigString "server.port" "9188"
dbUrl = GetConfigString "database.url" "sqlite:app.db"

ConnectDB "sqlite" dbUrl
Println "Database connected"

# Step 3: Create Schema
Println ""
Println "Step 3: Creating schema..."
ExecDB "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, email TEXT)"
Println "Schema created"

# Step 4: Repository Pattern (simulated)
function UserRepository() {
    Println "UserRepository: Repository operations"
}

function findAllUsers() {
    QueryDB "SELECT * FROM users"
    return GetQueryResult
}

function findUserById(id) {
    QueryRowDB "SELECT * FROM users WHERE id = ?" id
    return GetQueryResult
}

function createUser(name, email) {
    ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" name email
    return GetQueryResult
}

# Step 5: Service Layer (simulated)
function UserService() {
    Println "UserService: Service operations"
}

function getAllUsers() {
    # Check cache first
    cached = GetCache "users:all"
    if cached != "" {
        Println "Returning from cache"
        return cached
    }
    
    # Query database
    result = findAllUsers
    SetCache "users:all" result 300
    return result
}

function getUserById(id) {
    cacheKey = "user:" + id
    cached = GetCache cacheKey
    if cached != "" {
        return cached
    }
    
    result = findUserById id
    SetCache cacheKey result 300
    return result
}

function createNewUser(name, email) {
    # Create user
    result = createUser name email
    
    # Invalidate cache
    DeleteCache "users:all"
    
    return result
}

# Step 6: Controller Layer
function handleGetUsers() {
    result = getAllUsers
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleGetUser() {
    userId = GetHTTPQuery "id"
    result = getUserById userId
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleCreateUser() {
    body = GetHTTPBody
    # Parse body (simplified)
    name = "John Doe"
    email = "john@example.com"
    
    result = createNewUser name email
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 201 result
}

# Step 7: Start HTTP Server
Println ""
Println "Step 7: Starting HTTP server..."
StartHTTPServer port
sleep 1

# Register routes
Println "Registering routes..."
RegisterHTTPRoute "GET" "/api/users" "function" "handleGetUsers"
RegisterHTTPRoute "GET" "/api/users/:id" "function" "handleGetUser"
RegisterHTTPRoute "POST" "/api/users" "function" "handleCreateUser"
RegisterHTTPRoute "GET" "/health" "script" "SetHTTPResponse 200 'OK'"

Println ""
Println "=== Complete Spring-like Application is running ==="
Println "Server: http://localhost:" + port
Println ""
Println "Architecture:"
Println "  - Configuration Management: ✅"
Println "  - Database Connection: ✅"
Println "  - Repository Pattern: ✅"
Println "  - Service Layer: ✅"
Println "  - Controller Layer: ✅"
Println "  - Caching: ✅"
Println "  - HTTP Routing: ✅"
Println ""
Println "Available endpoints:"
Println "  GET  /api/users - List all users (cached)"
Println "  GET  /api/users/:id - Get user by ID (cached)"
Println "  POST /api/users - Create a user"
Println "  GET  /health - Health check"
Println ""
Println "Press Ctrl+C to stop the server"
