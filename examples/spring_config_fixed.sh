#!/usr/bin/env shode

# Spring-like Configuration Management Example (Fixed)
# Demonstrates configuration loading and access

Println "=== Configuration Management Example ==="

# Create a sample configuration file (single line JSON)
configContent = '{"server":{"port":9188,"host":"localhost"},"database":{"url":"sqlite:app.db","pool":{"max":10,"min":2}},"cache":{"enabled":true,"ttl":3600}}'

WriteFile "test/tmp/application.json" configContent
Println "Created test/tmp/application.json"

# Load configuration
Println "Loading configuration..."
LoadConfig "test/tmp/application.json"
Println "Configuration loaded"

# Access configuration values
Println ""
Println "Configuration values:"
port = GetConfigString "server.port" "8080"
Println "Server port: " + port

host = GetConfigString "server.host" "0.0.0.0"
Println "Server host: " + host

dbUrl = GetConfigString "database.url" ""
Println "Database URL: " + dbUrl

maxPool = GetConfigInt "database.pool.max" 5
Println "Database pool max: " + maxPool

cacheEnabled = GetConfigBool "cache.enabled" false
Println "Cache enabled: " + cacheEnabled

# Set configuration programmatically
Println ""
Println "Setting configuration..."
SetConfig "server.port" "9199"
newPort = GetConfigString "server.port" "8080"
Println "New server port: " + newPort

Println ""
Println "=== Configuration Example Complete ==="
