#!/usr/bin/env shode

# Data Aggregation Example
# Demonstrates data aggregation with caching

Println "=== Data Aggregation Demo ==="

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "sales.db"
Println "Database connected"

# Create sales table
Println "Creating sales table..."
ExecDB "CREATE TABLE IF NOT EXISTS sales (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id INTEGER, amount REAL, sale_date DATE)"

# Insert sales data
Println "Inserting sales data..."
ExecDB "INSERT INTO sales (product_id, amount, sale_date) VALUES (?, ?, ?)" "1" "100.00" "2024-01-01"
ExecDB "INSERT INTO sales (product_id, amount, sale_date) VALUES (?, ?, ?)" "1" "150.00" "2024-01-02"
ExecDB "INSERT INTO sales (product_id, amount, sale_date) VALUES (?, ?, ?)" "2" "200.00" "2024-01-01"
ExecDB "INSERT INTO sales (product_id, amount, sale_date) VALUES (?, ?, ?)" "2" "250.00" "2024-01-02"
ExecDB "INSERT INTO sales (product_id, amount, sale_date) VALUES (?, ?, ?)" "1" "120.00" "2024-01-03"
Println "Sales data inserted"

# Aggregate: Total sales by product
Println "Aggregating sales by product..."
QueryDB "SELECT product_id, SUM(amount) as total, COUNT(*) as count FROM sales GROUP BY product_id"
aggregateResult = GetQueryResult
Println "Aggregation result: " + aggregateResult

# Cache aggregated result for 1 hour
Println "Caching aggregation result..."
SetCache "sales:by_product" aggregateResult 3600
Println "Result cached for 1 hour"

# Retrieve from cache
Println "Retrieving from cache..."
cached = GetCache "sales:by_product"
if cached != "" {
    Println "Cache hit! Retrieved from cache: " + cached
} else {
    Println "Cache miss"
}

# Aggregate: Daily sales
Println "Aggregating daily sales..."
QueryDB "SELECT sale_date, SUM(amount) as daily_total, COUNT(*) as transaction_count FROM sales GROUP BY sale_date ORDER BY sale_date"
dailyResult = GetQueryResult
Println "Daily sales: " + dailyResult

# Cache daily aggregation
SetCache "sales:by_date" dailyResult 3600

# Get all sales cache keys
Println "Getting all sales cache keys..."
salesKeys = GetCacheKeys "sales:*"
Println "Sales cache keys: " + salesKeys

# Close database
Println "Closing database..."
CloseDB
Println "Database closed"

Println ""
Println "=== Data Aggregation Demo Complete ==="
