#!/usr/bin/env shode
# Static File Server Example
# This example demonstrates serving static files (HTML, CSS, JS, images, etc.)

# Start HTTP server on port 8080
StartHTTPServer "8080"

# Register static file route
# Use the test_static directory in the same location as this script
RegisterStaticRoute "/" "examples/test_static"

# Register an API endpoint alongside static files
function handleAPI() {
    SetHTTPResponse 200 "API is working correctly"
}
RegisterHTTPRoute "GET" "/api/status" "function" "handleAPI"

# Print server information
Println "================================"
Println "Shode Static File Server"
Println "================================"
Println ""
Println "Server running at: http://localhost:8080"
Println ""
Println "Try these URLs:"
Println "  - http://localhost:8080/           (index page)"
Println "  - http://localhost:8080/test.html  (test page)"
Println "  - http://localhost:8080/style.css  (CSS file)"
Println "  - http://localhost:8080/script.js  (JavaScript file)"
Println "  - http://localhost:8080/api/status (API endpoint)"
Println ""
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep the script running (use a finite loop for safety)
for i in $(seq 1 100000); do
    sleep 1
done

