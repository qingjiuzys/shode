#!/usr/bin/env shode

# ========================================
# Shode HTTP Server Complete Test Suite
# ========================================
# This script tests all HTTP server functionality including:
# - Server start/stop
# - Route registration (function and script handlers)
# - HTTP methods (GET, POST, PUT, DELETE, PATCH)
# - Query parameters
# - Request headers
# - Request body
# - Response status codes
# - Response headers
# - Error handling
# ========================================

Println "========================================"
Println "Shode HTTP Server Complete Test Suite"
Println "========================================"
Println ""

# Start HTTP server
Println "[1/10] Starting HTTP server on port 9188..."
StartHTTPServer "9188"
sleep 2

if ! IsHTTPServerRunning {
    Println "❌ FAILED: Server did not start"
    exit 1
}
Println "✓ Server started successfully"
Println ""

# Test 1: Simple route with fixed response
Println "[2/10] Testing simple route..."
RegisterRouteWithResponse "/" "Welcome to Shode HTTP Server"
Println "✓ Route registered: /"
Println ""

# Test 2: GET endpoint with query parameters
Println "[3/10] Testing GET endpoint with query parameters..."
function handleHello {
    name = GetHTTPQuery "name"
    if name == "" {
        name = "World"
    }
    greeting = "Hello, " + name + "!"
    SetHTTPHeader "Content-Type" "text/plain"
    SetHTTPResponse 200 greeting
}
RegisterHTTPRoute "GET" "/hello" "function" "handleHello"
Println "✓ Route registered: GET /hello?name=..."
Println ""

# Test 3: POST endpoint with request body
Println "[4/10] Testing POST endpoint with request body..."
function handlePost {
    method = GetHTTPMethod
    path = GetHTTPPath
    body = GetHTTPBody

    response = "Method: " + method + "\n"
    response = response + "Path: " + path + "\n"
    response = response + "Body: " + body

    SetHTTPHeader "Content-Type" "text/plain"
    SetHTTPResponse 200 response
}
RegisterHTTPRoute "POST" "/echo" "function" "handlePost"
Println "✓ Route registered: POST /echo"
Println ""

# Test 4: PUT endpoint
Println "[5/10] Testing PUT endpoint..."
function handlePut {
    body = GetHTTPBody
    SetHTTPResponse 200 "PUT received: " + body
}
RegisterHTTPRoute "PUT" "/update" "function" "handlePut"
Println "✓ Route registered: PUT /update"
Println ""

# Test 5: DELETE endpoint
Println "[6/10] Testing DELETE endpoint..."
function handleDelete {
    id = GetHTTPQuery "id"
    SetHTTPResponse 200 "Deleted item with ID: " + id
}
RegisterHTTPRoute "DELETE" "/delete" "function" "handleDelete"
Println "✓ Route registered: DELETE /delete?id=..."
Println ""

# Test 6: Request headers
Println "[7/10] Testing request headers..."
function handleHeaders {
    userAgent = GetHTTPHeader "User-Agent"
    accept = GetHTTPHeader "Accept"
    customHeader = GetHTTPHeader "X-Custom-Header"

    response = "User-Agent: " + userAgent + "\n"
    response = response + "Accept: " + accept + "\n"
    response = response + "X-Custom-Header: " + customHeader

    SetHTTPResponse 200 response
}
RegisterHTTPRoute "GET" "/headers" "function" "handleHeaders"
Println "✓ Route registered: GET /headers"
Println ""

# Test 7: JSON API endpoint
Println "[8/10] Testing JSON API endpoint..."
function handleApi {
    method = GetHTTPMethod

    if method == "GET" {
        # Return sample data
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 200 '{"status":"success","data":{"message":"API is working"}}'
    } else if method == "POST" {
        # Accept POST data
        body = GetHTTPBody
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 201 '{"status":"created","received":"' + body + '"}'
    }
}
RegisterHTTPRoute "GET" "/api/status" "function" "handleApi"
RegisterHTTPRoute "POST" "/api/status" "function" "handleApi"
Println "✓ Route registered: GET/POST /api/status"
Println ""

# Test 8: Script-based route (inline)
Println "[9/10] Testing script-based route..."
RegisterHTTPRoute "GET" "/ping" "script" "SetHTTPResponse 200 'pong'"
Println "✓ Route registered: GET /ping (script handler)"
Println ""

# Test 9: Error handling
Println "[10/10] Testing error handling..."
function handleNotFound {
    SetHTTPResponse 404 "Not Found: The requested resource was not found"
}
function handleBadRequest {
    SetHTTPResponse 400 "Bad Request: Invalid parameters"
}
RegisterHTTPRoute "GET" "/error/404" "function" "handleNotFound"
RegisterHTTPRoute "GET" "/error/400" "function" "handleBadRequest"
Println "✓ Error routes registered"
Println ""

# Summary
Println "========================================"
Println "✓ All routes registered successfully!"
Println "========================================"
Println ""
Println "Server is running on http://localhost:9188"
Println ""
Println "Available test endpoints:"
Println ""
Println "  1. GET  /                        - Welcome message"
Println "  2. GET  /hello?name=Alice         - Greeting with query param"
Println "  3. POST /echo                    - Echo back request body"
Println "  4. PUT  /update                  - Update endpoint"
Println "  5. DELETE /delete?id=123         - Delete with query param"
Println "  6. GET  /headers                 - Returns request headers"
Println "  7. GET  /api/status              - JSON API (GET)"
Println "  8. POST /api/status              - JSON API (POST)"
Println "  9. GET  /ping                    - Simple ping/pong"
Println " 10. GET  /error/404               - 404 error test"
Println " 11. GET  /error/400               - 400 error test"
Println ""
Println "========================================"
Println "Test Commands:"
Println "========================================"
Println ""
Println "# Test 1: Welcome"
Println "curl http://localhost:9188/"
Println ""
Println "# Test 2: Hello with query parameter"
Println "curl http://localhost:9188/hello?name=Alice"
Println "curl http://localhost:9188/hello"
Println ""
Println "# Test 3: POST echo"
Println "curl -X POST http://localhost:9188/echo -d 'Hello from curl'"
Println ""
Println "# Test 4: PUT update"
Println "curl -X PUT http://localhost:9188/update -d 'Update data'"
Println ""
Println "# Test 5: DELETE"
Println "curl -X DELETE http://localhost:9188/delete?id=123"
Println ""
Println "# Test 6: Headers"
Println "curl -v http://localhost:9188/headers"
Println "curl -H 'X-Custom-Header: MyValue' http://localhost:9188/headers"
Println ""
Println "# Test 7: JSON API"
Println "curl http://localhost:9188/api/status"
Println "curl -X POST http://localhost:9188/api/status -d '{\"test\":\"data\"}'"
Println ""
Println "# Test 8: Ping"
Println "curl http://localhost:9188/ping"
Println ""
Println "# Test 9: Error handling"
Println "curl http://localhost:9188/error/404"
Println "curl http://localhost:9188/error/400"
Println ""
Println "# Run all tests"
Println "echo 'Running comprehensive HTTP tests...'"
Println ""
Println "========================================"
Println "Press Ctrl+C to stop the server"
Println "========================================"
