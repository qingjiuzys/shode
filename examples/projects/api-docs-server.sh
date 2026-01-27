#!/usr/bin/env shode
# API Documentation Server
# This example demonstrates hosting API documentation with directory browsing enabled

StartHTTPServer "8080"

# Serve documentation with directory browsing enabled
# This allows users to browse through different API version docs
RegisterStaticRouteAdvanced "/docs" "./documentation" \
    "index.html,README.md" \
    "true" \
    "max-age=3600" \
    "true" \
    ""

# Serve static assets (images, code samples) with longer cache
RegisterStaticRouteAdvanced "/assets" "./docs-assets" \
    "" \
    "false" \
    "max-age=86400" \
    "true" \
    ""

# Add a simple search endpoint (placeholder)
function searchDocs() {
    SetHTTPResponse 200 '{"results":["Endpoint 1","Endpoint 2"],"total":2}'
}
RegisterHTTPRoute "GET" "/api/search" "function" "searchDocs"

Println "📚 API Documentation Server running at http://localhost:8080"
Println "📖 Browse docs at http://localhost:8080/docs"
Println "🔍 Search API at http://localhost:8080/api/search?q=endpoint"

for i in $(seq 1 100000); do sleep 1; done
