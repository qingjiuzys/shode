#!/usr/bin/env shode

# Database Example
# Demonstrates database operations with MySQL, PostgreSQL, and SQLite

Println "Database Example"

# Connect to SQLite (for testing)
ConnectDB "sqlite" ":memory:"
Println "Connected to SQLite"

# Create table
ExecDB "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)"
Println "Table created"

# Insert data
ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "Alice" "alice@example.com"
ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "Bob" "bob@example.com"
Println "Data inserted"

# Query data
QueryDB "SELECT * FROM users"
result = GetQueryResult
Println "Query result: " + result

# Query single row
QueryRowDB "SELECT * FROM users WHERE id = ?" "1"
row = GetQueryResult
Println "Single row: " + row

# Update data
ExecDB "UPDATE users SET email = ? WHERE id = ?" "alice.new@example.com" "1"
Println "Data updated"

# Close connection
CloseDB
Println "Connection closed"
