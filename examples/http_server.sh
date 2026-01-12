#!/usr/bin/env shode

# Simple HTTP Server Example
# This script demonstrates how to create an HTTP server using Shode
# Usage: shode run http_server.sh

# Start HTTP server on port 9188
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"

# Wait a moment for server to start
sleep 1

# Register a route that returns "hello world"
Println "Registering route /..."
RegisterRouteWithResponse "/" "hello world"

Println "HTTP server is running on http://localhost:9188"
Println "Visit http://localhost:9188 in your browser or use: curl http://localhost:9188"
Println ""
Println "Server is ready to accept requests"
