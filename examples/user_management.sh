#!/usr/bin/env shode

# User Management System Example
# Demonstrates complete CRUD operations with caching

Println "=== User Management System ==="

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "users.db"
Println "Database connected"

# Create users table
Println "Creating users table..."
ExecDB "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE, email TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"

# Insert sample users
Println "Inserting sample users..."
ExecDB "INSERT OR IGNORE INTO users (username, email) VALUES (?, ?)" "alice" "alice@example.com"
ExecDB "INSERT OR IGNORE INTO users (username, email) VALUES (?, ?)" "bob" "bob@example.com"
ExecDB "INSERT OR IGNORE INTO users (username, email) VALUES (?, ?)" "charlie" "charlie@example.com"
Println "Sample users inserted"

# Query all users and cache
Println "Querying all users..."
QueryDB "SELECT id, username, email FROM users ORDER BY id"
result = GetQueryResult
Println "Query result: " + result

# Cache the result
Println "Caching user list..."
SetCache "users:all" result 60
Println "User list cached for 60 seconds"

# Retrieve from cache
Println "Retrieving from cache..."
cached = GetCache "users:all"
if cached != "" {
    Println "Cache hit! Retrieved from cache"
} else {
    Println "Cache miss"
}

# Query single user
Println "Querying single user (alice)..."
QueryRowDB "SELECT * FROM users WHERE username = ?" "alice"
singleResult = GetQueryResult
Println "Single user result: " + singleResult

# Update user
Println "Updating user email..."
ExecDB "UPDATE users SET email = ? WHERE username = ?" "alice.new@example.com" "alice"
Println "User updated"

# Invalidate cache after update
Println "Invalidating cache..."
DeleteCache "users:all"
Println "Cache invalidated"

# Verify cache is deleted
Println "Checking if cache exists..."
exists = CacheExists "users:all"
Println "Cache exists: " + exists

# Query updated user
Println "Querying updated user..."
QueryRowDB "SELECT * FROM users WHERE username = ?" "alice"
updatedResult = GetQueryResult
Println "Updated user result: " + updatedResult

# Close database
Println "Closing database connection..."
CloseDB
Println "Database closed"

Println ""
Println "=== User Management Demo Complete ==="
