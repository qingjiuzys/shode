#!/usr/bin/env shode

# Database initialization and setup

function initDatabase() {
    Println "Step 1: Initializing database..."
    dbPath = "test/tmp/library.db"
    ConnectDB "sqlite" dbPath
    Println "Database connected: " + dbPath
    
    # Create tables
    Println "Creating database schema..."
    ExecDB "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT UNIQUE NOT NULL, password_hash TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"
    ExecDB "CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE NOT NULL, description TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"
    ExecDB "CREATE TABLE IF NOT EXISTS books (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, author TEXT, isbn TEXT, category_id INTEGER, price REAL, stock INTEGER DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP, FOREIGN KEY(category_id) REFERENCES categories(id))"
    Println "Schema created"
    
    # Initialize default admin user (password: admin123)
    Println ""
    Println "Creating default admin user..."
    adminPassword = "admin123"
    passwordHash = SHA256Hash adminPassword
    ExecDB "INSERT OR IGNORE INTO users (username, password_hash) VALUES (?, ?)" "admin" passwordHash
    Println "Default admin user created (username: admin, password: admin123)"
    
    # Initialize default categories
    Println ""
    Println "Creating default categories..."
    ExecDB "INSERT OR IGNORE INTO categories (name, description) VALUES (?, ?)" "Fiction" "Fiction books"
    ExecDB "INSERT OR IGNORE INTO categories (name, description) VALUES (?, ?)" "Non-Fiction" "Non-fiction books"
    ExecDB "INSERT OR IGNORE INTO categories (name, description) VALUES (?, ?)" "Science" "Science books"
    ExecDB "INSERT OR IGNORE INTO categories (name, description) VALUES (?, ?)" "History" "History books"
    Println "Default categories created"
}
