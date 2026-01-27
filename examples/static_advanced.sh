#!/usr/bin/env shode
# Advanced Static File Server Example
# This demonstrates directory browsing, cache control, and other advanced features

# Start HTTP server
StartHTTPServer "8083"

# 1. Register static route with directory browsing enabled
# Usage: RegisterStaticRouteAdvanced "/" "directory" "indexFiles" "directoryBrowse" "cacheControl" "enableGzip" "spaFallback"
RegisterStaticRouteAdvanced "/" "examples/test_static" "index.html,index.htm" "true" "max-age=3600" "false" ""

# 2. Register an API endpoint
function handleAPI() {
    SetHTTPResponse 200 "API Status: OK"
}
RegisterHTTPRoute "GET" "/api/status" "function" "handleAPI"

Println "================================"
Println "Advanced Static File Server"
Println "================================"
Println ""
Println "Server running at: http://localhost:8083"
Println ""
Println "Features:"
Println "  - Directory browsing ENABLED"
Println "  - Cache control: max-age=3600"
Println "  - Index files: index.html, index.htm"
Println ""
Println "Try these URLs:"
Println "  - http://localhost:8083/              (directory listing)"
Println "  - http://localhost:8083/test.html   (test page)"
Println "  - http://localhost:8083/api/status   (API endpoint)"
Println ""
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep server running
for i in $(seq 1 100000); do
    sleep 1
done
