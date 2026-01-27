#!/usr/bin/env shode
# File Download Server
# Example of a file server optimized for downloads with caching

StartHTTPServer "5000"

# Main downloads directory - no directory browsing, long cache
RegisterStaticRouteAdvanced "/downloads" "./files" \
    "" \
    "false" \
    "max-age=86400, public" \
    "true" \
    ""

# Release notes directory - browsing enabled, medium cache
RegisterStaticRouteAdvanced "/releases" "./release-notes" \
    "index.html,README.md" \
    "true" \
    "max-age=3600" \
    "false" \
    ""

# API: Get file list
function listFiles() {
    SetHTTPResponse 200 '{"files":["app-v1.0.0.tar.gz","app-v1.1.0.tar.gz"],"total":2}'
}
RegisterHTTPRoute "GET" "/api/files" "function" "listFiles"

# API: Get latest version
function getLatest() {
    SetHTTPResponse 200 '{"version":"1.1.0","downloadUrl":"/downloads/app-v1.1.0.tar.gz"}'
}
RegisterHTTPRoute "GET" "/api/latest" "function" "getLatest"

Println "📦 File Download Server running at http://localhost:5000"
Println "⬇️  Downloads:     http://localhost:5000/downloads"
Println "📄 Release notes: http://localhost:5000/releases"
Println "🔌 API endpoints:"
Println "   GET /api/files  - List available files"
Println "   GET /api/latest - Get latest version info"

for i in $(seq 1 100000); do sleep 1; done
