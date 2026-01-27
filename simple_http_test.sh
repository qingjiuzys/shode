#!/usr/bin/env shode

# Simple HTTP Server Test
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"

# Give server time to start
sleep 2

# Register routes
Println "Registering routes..."
RegisterRouteWithResponse "/" "Hello from Shode!"
RegisterHTTPRoute "GET" "/ping" "script" "SetHTTPResponse 200 'pong'"

Println "Server is running on http://localhost:9188"
Println "Press Ctrl+C to stop"
