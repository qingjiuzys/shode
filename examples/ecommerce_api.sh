#!/usr/bin/env shode

# E-Commerce API Example
# Demonstrates a complete e-commerce API with products, orders, and caching

Println "=== E-Commerce API Server ==="

# Start HTTP server
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"
sleep 1

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "ecommerce.db"
Println "Database connected"

# Create products table
Println "Creating products table..."
ExecDB "CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL, stock INTEGER, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"

# Create orders table
Println "Creating orders table..."
ExecDB "CREATE TABLE IF NOT EXISTS orders (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, total REAL, status TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"

# Insert sample products
Println "Inserting sample products..."
ExecDB "INSERT OR IGNORE INTO products (name, price, stock) VALUES (?, ?, ?)" "iPhone 15" "999.99" "10"
ExecDB "INSERT OR IGNORE INTO products (name, price, stock) VALUES (?, ?, ?)" "MacBook Pro" "1999.99" "5"
ExecDB "INSERT OR IGNORE INTO products (name, price, stock) VALUES (?, ?, ?)" "AirPods Pro" "249.99" "20"

# Define handler functions
function handleGetProducts() {
    # Check cache first
    cached = GetCache "products:list"
    if cached != "" {
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 200 cached
        return
    }
    
    # Query database
    QueryDB "SELECT id, name, price, stock FROM products WHERE stock > 0 ORDER BY price"
    result = GetQueryResult
    
    # Cache for 5 minutes
    SetCache "products:list" result 300
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleGetProduct() {
    # Get product ID from query
    productId = GetHTTPQuery "id"
    
    # Check cache
    cacheKey = "product:" + productId
    cached = GetCache cacheKey
    if cached != "" {
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 200 cached
        return
    }
    
    # Query database
    QueryRowDB "SELECT * FROM products WHERE id = ?" productId
    result = GetQueryResult
    
    # Cache for 10 minutes
    SetCache cacheKey result 600
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleCreateOrder() {
    # Get request body (simplified - in real app, parse JSON)
    body = GetHTTPBody
    
    # Insert order
    ExecDB "INSERT INTO orders (user_id, total, status) VALUES (?, ?, ?)" "1" "1249.98" "pending"
    result = GetQueryResult
    
    # Invalidate products cache
    DeleteCache "products:list"
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 201 result
}

function handleGetOrders() {
    # Get user ID from query
    userId = GetHTTPQuery "user_id"
    
    # Query orders
    QueryDB "SELECT * FROM orders WHERE user_id = ? ORDER BY created_at DESC" userId
    result = GetQueryResult
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

# Register routes
RegisterHTTPRoute "GET" "/api/products" "function" "handleGetProducts"
RegisterHTTPRoute "GET" "/api/product" "function" "handleGetProduct"
RegisterHTTPRoute "POST" "/api/orders" "function" "handleCreateOrder"
RegisterHTTPRoute "GET" "/api/orders" "function" "handleGetOrders"

# Health check route
RegisterHTTPRoute "GET" "/api/health" "script" "SetHTTPResponse 200 'OK'"

Println ""
Println "=== E-Commerce API is running ==="
Println "Server: http://localhost:9188"
Println ""
Println "Available endpoints:"
Println "  GET  /api/products?user_id=1 - List all products (cached)"
Println "  GET  /api/product?id=1 - Get product by ID (cached)"
Println "  POST /api/orders - Create a new order"
Println "  GET  /api/orders?user_id=1 - Get user orders"
Println "  GET  /api/health - Health check"
Println ""
Println "Press Ctrl+C to stop the server"
