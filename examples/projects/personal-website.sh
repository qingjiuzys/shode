#!/usr/bin/env shode
# Personal Website / Blog Server
# This example demonstrates hosting a simple personal website with a blog

# Start HTTP server on port 3000
StartHTTPServer "3000"

# Serve static website files from public directory
RegisterStaticRoute "/" "examples/projects/public"

# Optional: Add a simple API endpoint for visitor stats
function getStats() {
    SetHTTPResponse 200 '{"visitors":1243,"posts":42,"lastUpdated":"2026-01-27"}'
}
RegisterHTTPRoute "GET" "/api/stats" "function" "getStats"

Println "🚀 Personal website running at http://localhost:3000"
Println "📝 Blog posts available at /blog/"
Println "📊 Stats API at /api/stats"

# Keep server running
for i in $(seq 1 100000); do sleep 1; done
