#!/usr/bin/env shode

# Library Management System - Main Entry Point
# 
# This file loads all module files and starts the HTTP server.
# 
# File Structure:
#   examples/library_management.sh  - Main entry point (this file)
#   examples/library/database.sh    - Database initialization
#   examples/library/auth.sh        - Authentication functions (login, checkAuth)
#   examples/library/categories.sh  - Category management functions
#   examples/library/books.sh       - Book management functions
#   examples/library/handlers.sh    - HTTP route handlers
#
# Usage:
#   ./shode run examples/library_management.sh

Println "=== Library Management System ==="
Println ""

# Load all module files
# Since Shode functions are parsed at script load time, we need to include
# all function definitions. We'll do this by executing a combined file.

Println "Loading all modules..."
# Execute the combined modules file which contains all function definitions
shode run examples/library/all_modules.sh

Println "Modules loaded"
Println ""

# Initialize database
Println "Initializing database..."
initDatabase

# Start HTTP Server
Println ""
Println "Starting HTTP server..."
port = "9188"
StartHTTPServer port
sleep 1

# Register routes
Println "Registering routes..."

# Authentication
RegisterHTTPRoute "POST" "/api/login" "function" "handleLogin"

# Category management
RegisterHTTPRoute "GET" "/api/categories" "function" "handleListCategories"
RegisterHTTPRoute "POST" "/api/categories" "function" "handleCreateCategory"
RegisterHTTPRoute "PUT" "/api/categories" "function" "handleUpdateCategory"
RegisterHTTPRoute "DELETE" "/api/categories" "function" "handleDeleteCategory"

# Book management
RegisterHTTPRoute "GET" "/api/books" "function" "handleListBooks"
RegisterHTTPRoute "GET" "/api/books/:id" "function" "handleGetBook"
RegisterHTTPRoute "POST" "/api/books" "function" "handleCreateBook"
RegisterHTTPRoute "PUT" "/api/books" "function" "handleUpdateBook"
RegisterHTTPRoute "DELETE" "/api/books" "function" "handleDeleteBook"

# Health check
RegisterHTTPRoute "GET" "/health" "script" "SetHTTPResponse 200 'OK'"

Println ""
Println "=== Library Management System is running ==="
Println "Server: http://localhost:" + port
Println ""
Println "API Endpoints:"
Println "  Authentication:"
Println "    POST /api/login?username=admin&password=admin123"
Println ""
Println "  Categories:"
Println "    GET    /api/categories - List all categories"
Println "    POST   /api/categories?name=Tech&description=Technology - Create category"
Println "    PUT    /api/categories?id=1&name=Technology - Update category"
Println "    DELETE /api/categories?id=1 - Delete category"
Println ""
Println "  Books:"
Println "    GET    /api/books - List all books"
Println "    GET    /api/books?category_id=1 - List books by category"
Println "    GET    /api/books/:id - Get book by ID"
Println "    POST   /api/books?title=Book&author=Author&category_id=1&price=29.99&stock=10 - Create book"
Println "    PUT    /api/books?id=1&title=New Title - Update book"
Println "    DELETE /api/books?id=1 - Delete book"
Println ""
Println "  Health:"
Println "    GET    /health - Health check"
Println ""
Println "Usage Example:"
Println "  1. Login: curl 'http://localhost:" + port + "/api/login?username=admin&password=admin123'"
Println "  2. Get token from response"
Println "  3. Use token in Authorization header: curl -H 'Authorization: <token>' 'http://localhost:" + port + "/api/books'"
Println ""
Println "Press Ctrl+C to stop the server"
