#!/usr/bin/env shode
# Error Pages Demo

StartHTTPServer "8092"

# Configure custom error pages
SetErrorPage 404 "examples/error-pages/404.html"
SetErrorPage 500 "examples/error-pages/500.html"

# Serve static files
RegisterStaticRoute "/" "./examples/test_static"

Println "Error Pages Demo Server"
Println "http://localhost:8092"
Println ""
Println "Try accessing:"
Println "  http://localhost:8092/  - Should work"
Println "  http://localhost:8092/nonexistent - Should show custom 404 page"
Println ""
Println "Custom 404 and 500 error pages are configured!"

for i in $(seq 1 100000); do sleep 1; done
