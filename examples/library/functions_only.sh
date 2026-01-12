# Authentication functions
# Login Function
function login(username, password) {
    # Hash the password
    passwordHash = SHA256Hash password
    
    # Query user from database
    QueryRowDB "SELECT id, username, password_hash FROM users WHERE username = ?" username
    result = GetQueryResult
    
    # Check if user exists and password matches
    if Contains result passwordHash {
        # Generate session token (simplified)
        sessionToken = SHA256Hash username + passwordHash + "session"
        SetCache "session:" + sessionToken username 3600
        SetEnv "login_token" sessionToken
        SetEnv "login_success" "true"
    } else {
        SetEnv "login_token" ""
        SetEnv "login_success" "false"
    }
}
# Authentication Middleware
function checkAuth() {
    token = GetHTTPHeader "Authorization"
    if token == "" {
        SetHTTPResponse 401 "Unauthorized: Missing token"
        SetEnv "auth_valid" "false"
        return
    }
    
    # Remove "Bearer " prefix if present
    if Contains token "Bearer " {
        token = Replace token "Bearer " ""
    }
    
    # Check session in cache
    cacheKey = "session:" + token
    username = GetCache cacheKey
    if username == "" {
        SetHTTPResponse 401 "Unauthorized: Invalid token"
        SetEnv "auth_valid" "false"
        return
    }
    
    SetEnv "current_user" username
    SetEnv "auth_valid" "true"
}
# Book Management Functions
function listBooks() {
    categoryId = GetHTTPQuery "category_id"
    
    if categoryId != "" {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.category_id = ? ORDER BY b.title" categoryId
    } else {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id ORDER BY b.title"
    }
    
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function getBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    QueryRowDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = ?" bookId
    result = GetQueryResult
    
    if result == "" {
        SetHTTPResponse 404 "Book not found"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function createBook() {
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if title == "" {
        SetHTTPResponse 400 "Book title is required"
        return
    }
    
    if stock == "" {
        stock = "0"
    }
    if price == "" {
        price = "0"
    }
    
    ExecDB "INSERT INTO books (title, author, isbn, category_id, price, stock) VALUES (?, ?, ?, ?, ?, ?)" title author isbn categoryId price stock
    SetHTTPResponse 201 "Book created"
}
function updateBook() {
    bookId = GetHTTPQuery "id"
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    if title != "" {
        ExecDB "UPDATE books SET title = ? WHERE id = ?" title bookId
    }
    if author != "" {
        ExecDB "UPDATE books SET author = ? WHERE id = ?" author bookId
    }
    if isbn != "" {
        ExecDB "UPDATE books SET isbn = ? WHERE id = ?" isbn bookId
    }
    if categoryId != "" {
        ExecDB "UPDATE books SET category_id = ? WHERE id = ?" categoryId bookId
    }
    if price != "" {
        ExecDB "UPDATE books SET price = ? WHERE id = ?" price bookId
    }
    if stock != "" {
        ExecDB "UPDATE books SET stock = ? WHERE id = ?" stock bookId
    }
    
    SetHTTPResponse 200 "Book updated"
}
function deleteBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    ExecDB "DELETE FROM books WHERE id = ?" bookId
    SetHTTPResponse 200 "Book deleted"
}
# Category Management Functions
function listCategories() {
    QueryDB "SELECT id, name, description FROM categories ORDER BY name"
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function createCategory() {
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if name == "" {
        SetHTTPResponse 400 "Category name is required"
        return
    }
    
    ExecDB "INSERT INTO categories (name, description) VALUES (?, ?)" name description
    SetHTTPResponse 201 "Category created"
}
function updateCategory() {
    categoryId = GetHTTPQuery "id"
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    if name != "" {
        ExecDB "UPDATE categories SET name = ? WHERE id = ?" name categoryId
    }
    if description != "" {
        ExecDB "UPDATE categories SET description = ? WHERE id = ?" description categoryId
    }
    
    SetHTTPResponse 200 "Category updated"
}
function deleteCategory() {
    categoryId = GetHTTPQuery "id"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    # Check if category has books
    QueryRowDB "SELECT COUNT(*) FROM books WHERE category_id = ?" categoryId
    bookCount = GetQueryResult
    
    if Contains bookCount "0" {
        ExecDB "DELETE FROM categories WHERE id = ?" categoryId
        SetHTTPResponse 200 "Category deleted"
    } else {
        SetHTTPResponse 400 "Cannot delete category with books"
    }
}
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
# HTTP Route Handlers
function handleLogin() {
    username = GetHTTPQuery "username"
    password = GetHTTPQuery "password"
    
    if username == "" || password == "" {
        SetHTTPResponse 400 "Username and password are required"
        return
    }
    
    login username password
    token = GetEnv "login_token"
    success = GetEnv "login_success"
    
    if success == "false" || token == "" {
        SetHTTPResponse 401 "Invalid username or password"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    response = "{\"token\":\"" + token + "\",\"username\":\"" + username + "\"}"
    SetHTTPResponse 200 response
}
function handleListCategories() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listCategories
}
function handleCreateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createCategory
}
function handleUpdateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateCategory
}
function handleDeleteCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteCategory
}
function handleListBooks() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listBooks
}
function handleGetBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    getBook
}
function handleCreateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createBook
}
function handleUpdateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateBook
}
function handleDeleteBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteBook
}
# Authentication functions
# Login Function
function login(username, password) {
    # Hash the password
    passwordHash = SHA256Hash password
    
    # Query user from database
    QueryRowDB "SELECT id, username, password_hash FROM users WHERE username = ?" username
    result = GetQueryResult
    
    # Check if user exists and password matches
    if Contains result passwordHash {
        # Generate session token (simplified)
        sessionToken = SHA256Hash username + passwordHash + "session"
        SetCache "session:" + sessionToken username 3600
        SetEnv "login_token" sessionToken
        SetEnv "login_success" "true"
    } else {
        SetEnv "login_token" ""
        SetEnv "login_success" "false"
    }
}
# Authentication Middleware
function checkAuth() {
    token = GetHTTPHeader "Authorization"
    if token == "" {
        SetHTTPResponse 401 "Unauthorized: Missing token"
        SetEnv "auth_valid" "false"
        return
    }
    
    # Remove "Bearer " prefix if present
    if Contains token "Bearer " {
        token = Replace token "Bearer " ""
    }
    
    # Check session in cache
    cacheKey = "session:" + token
    username = GetCache cacheKey
    if username == "" {
        SetHTTPResponse 401 "Unauthorized: Invalid token"
        SetEnv "auth_valid" "false"
        return
    }
    
    SetEnv "current_user" username
    SetEnv "auth_valid" "true"
}
# Book Management Functions
function listBooks() {
    categoryId = GetHTTPQuery "category_id"
    
    if categoryId != "" {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.category_id = ? ORDER BY b.title" categoryId
    } else {
        QueryDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id ORDER BY b.title"
    }
    
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function getBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    QueryRowDB "SELECT b.id, b.title, b.author, b.isbn, c.name as category, b.price, b.stock FROM books b LEFT JOIN categories c ON b.category_id = c.id WHERE b.id = ?" bookId
    result = GetQueryResult
    
    if result == "" {
        SetHTTPResponse 404 "Book not found"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function createBook() {
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if title == "" {
        SetHTTPResponse 400 "Book title is required"
        return
    }
    
    if stock == "" {
        stock = "0"
    }
    if price == "" {
        price = "0"
    }
    
    ExecDB "INSERT INTO books (title, author, isbn, category_id, price, stock) VALUES (?, ?, ?, ?, ?, ?)" title author isbn categoryId price stock
    SetHTTPResponse 201 "Book created"
}
function updateBook() {
    bookId = GetHTTPQuery "id"
    title = GetHTTPQuery "title"
    author = GetHTTPQuery "author"
    isbn = GetHTTPQuery "isbn"
    categoryId = GetHTTPQuery "category_id"
    price = GetHTTPQuery "price"
    stock = GetHTTPQuery "stock"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    if title != "" {
        ExecDB "UPDATE books SET title = ? WHERE id = ?" title bookId
    }
    if author != "" {
        ExecDB "UPDATE books SET author = ? WHERE id = ?" author bookId
    }
    if isbn != "" {
        ExecDB "UPDATE books SET isbn = ? WHERE id = ?" isbn bookId
    }
    if categoryId != "" {
        ExecDB "UPDATE books SET category_id = ? WHERE id = ?" categoryId bookId
    }
    if price != "" {
        ExecDB "UPDATE books SET price = ? WHERE id = ?" price bookId
    }
    if stock != "" {
        ExecDB "UPDATE books SET stock = ? WHERE id = ?" stock bookId
    }
    
    SetHTTPResponse 200 "Book updated"
}
function deleteBook() {
    bookId = GetHTTPQuery "id"
    
    if bookId == "" {
        SetHTTPResponse 400 "Book ID is required"
        return
    }
    
    ExecDB "DELETE FROM books WHERE id = ?" bookId
    SetHTTPResponse 200 "Book deleted"
}
# Category Management Functions
function listCategories() {
    QueryDB "SELECT id, name, description FROM categories ORDER BY name"
    result = GetQueryResult
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}
function createCategory() {
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if name == "" {
        SetHTTPResponse 400 "Category name is required"
        return
    }
    
    ExecDB "INSERT INTO categories (name, description) VALUES (?, ?)" name description
    SetHTTPResponse 201 "Category created"
}
function updateCategory() {
    categoryId = GetHTTPQuery "id"
    name = GetHTTPQuery "name"
    description = GetHTTPQuery "description"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    if name != "" {
        ExecDB "UPDATE categories SET name = ? WHERE id = ?" name categoryId
    }
    if description != "" {
        ExecDB "UPDATE categories SET description = ? WHERE id = ?" description categoryId
    }
    
    SetHTTPResponse 200 "Category updated"
}
function deleteCategory() {
    categoryId = GetHTTPQuery "id"
    
    if categoryId == "" {
        SetHTTPResponse 400 "Category ID is required"
        return
    }
    
    # Check if category has books
    QueryRowDB "SELECT COUNT(*) FROM books WHERE category_id = ?" categoryId
    bookCount = GetQueryResult
    
    if Contains bookCount "0" {
        ExecDB "DELETE FROM categories WHERE id = ?" categoryId
        SetHTTPResponse 200 "Category deleted"
    } else {
        SetHTTPResponse 400 "Cannot delete category with books"
    }
}
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
# HTTP Route Handlers
function handleLogin() {
    username = GetHTTPQuery "username"
    password = GetHTTPQuery "password"
    
    if username == "" || password == "" {
        SetHTTPResponse 400 "Username and password are required"
        return
    }
    
    login username password
    token = GetEnv "login_token"
    success = GetEnv "login_success"
    
    if success == "false" || token == "" {
        SetHTTPResponse 401 "Invalid username or password"
        return
    }
    
    SetHTTPHeader "Content-Type" "application/json"
    response = "{\"token\":\"" + token + "\",\"username\":\"" + username + "\"}"
    SetHTTPResponse 200 response
}
function handleListCategories() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listCategories
}
function handleCreateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createCategory
}
function handleUpdateCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateCategory
}
function handleDeleteCategory() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteCategory
}
function handleListBooks() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    listBooks
}
function handleGetBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    getBook
}
function handleCreateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    createBook
}
function handleUpdateBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    updateBook
}
function handleDeleteBook() {
    checkAuth
    authValid = GetEnv "auth_valid"
    if authValid == "false" {
        return
    }
    deleteBook
}
