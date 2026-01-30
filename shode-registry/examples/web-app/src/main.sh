#!/bin/sh
# Example web application using Shode official packages

# Load dependencies
. "sh_modules/@shode/logger/index.sh"
. "sh_modules/@shode/config/index.sh"
. "sh_modules/@shode/http/index.sh"

# Initialize logger
SetLogLevel "info"

LogInfo "Starting Shode web application..."

# Load configuration
ConfigLoad "config.json"

# Get configuration values
API_HOST=$(ConfigGet "API_HOST" "localhost")
API_PORT=$(ConfigGet "API_PORT" "8080")

LogInfo "Configuration loaded"
LogInfo "API Host: $API_HOST"
LogInfo "API Port: $API_PORT"

# Start HTTP server
LogInfo "Starting HTTP server on $API_HOST:$API_PORT"

# Example HTTP request
LogInfo "Making test API request..."
response=$(HttpGet "http://$API_HOST:$API_PORT/health")
LogInfo "API Response: $response"

LogInfo "Application started successfully"
