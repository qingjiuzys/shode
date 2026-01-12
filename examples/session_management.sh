#!/usr/bin/env shode

# Session Management Example
# Demonstrates session management using cache with TTL

Println "=== Session Management Demo ==="

# Create multiple sessions
Println "Creating sessions..."
SetCache "session:abc123" '{"username":"alice","role":"admin","login_time":"2024-01-01T10:00:00Z"}' 1800
SetCache "session:def456" '{"username":"bob","role":"user","login_time":"2024-01-01T11:00:00Z"}' 1800
SetCache "session:ghi789" '{"username":"charlie","role":"user","login_time":"2024-01-01T12:00:00Z"}' 1800
Println "3 sessions created (30 minute TTL)"

# Retrieve a session
Println "Retrieving session abc123..."
session1 = GetCache "session:abc123"
if session1 != "" {
    Println "Session found: " + session1
} else {
    Println "Session not found"
}

# Check if session exists
Println "Checking if session def456 exists..."
exists = CacheExists "session:def456"
Println "Session exists: " + exists

# Get TTL for a session
Println "Getting TTL for session ghi789..."
ttl = GetCacheTTL "session:ghi789"
Println "Remaining TTL: " + ttl + " seconds"

# Get all session keys using pattern matching
Println "Getting all session keys..."
allSessions = GetCacheKeys "session:*"
Println "All session keys:"
Println allSessions

# Update a session
Println "Updating session abc123..."
SetCache "session:abc123" '{"username":"alice","role":"admin","login_time":"2024-01-01T10:00:00Z","last_activity":"2024-01-01T10:30:00Z"}' 1800
Println "Session updated"

# Invalidate a session (logout)
Println "Invalidating session def456 (logout)..."
DeleteCache "session:def456"
Println "Session invalidated"

# Verify session was deleted
Println "Verifying session def456 was deleted..."
existsAfter = CacheExists "session:def456"
Println "Session exists after deletion: " + existsAfter

# Clear all sessions (emergency logout all)
Println "Clearing all sessions..."
ClearCache
Println "All sessions cleared"

Println ""
Println "=== Session Management Demo Complete ==="
