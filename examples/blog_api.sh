#!/usr/bin/env shode

# Blog API Example
# Demonstrates a blog API with posts, comments, views, and caching

Println "=== Blog API Server ==="

# Start HTTP server
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"
sleep 1

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "blog.db"
Println "Database connected"

# Create posts table
Println "Creating posts table..."
ExecDB "CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT, content TEXT, author_id INTEGER, views INTEGER DEFAULT 0, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"

# Create comments table
Println "Creating comments table..."
ExecDB "CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY AUTOINCREMENT, post_id INTEGER, author TEXT, content TEXT, created_at DATETIME DEFAULT CURRENT_TIMESTAMP)"

# Insert sample post
Println "Creating sample post..."
ExecDB "INSERT INTO posts (title, content, author_id) VALUES (?, ?, ?)" "Hello Shode" "This is my first blog post using Shode!" "1"

# Define handler functions
function handleGetPosts() {
    # Check cache
    cached = GetCache "posts:list"
    if cached != "" {
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 200 cached
        return
    }
    
    # Query posts with comment count
    QueryDB "SELECT p.*, COUNT(c.id) as comment_count FROM posts p LEFT JOIN comments c ON p.id = c.post_id GROUP BY p.id ORDER BY p.created_at DESC"
    result = GetQueryResult
    
    # Cache for 5 minutes
    SetCache "posts:list" result 300
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleGetPost() {
    # Get post ID from query
    postId = GetHTTPQuery "id"
    
    # Check cache
    cacheKey = "post:" + postId
    cached = GetCache cacheKey
    if cached != "" {
        SetHTTPHeader "Content-Type" "application/json"
        SetHTTPResponse 200 cached
        return
    }
    
    # Query post with comments
    QueryDB "SELECT p.*, COUNT(c.id) as comment_count FROM posts p LEFT JOIN comments c ON p.id = c.post_id WHERE p.id = ? GROUP BY p.id" postId
    result = GetQueryResult
    
    # Increment view count
    ExecDB "UPDATE posts SET views = views + 1 WHERE id = ?" postId
    
    # Cache for 10 minutes
    SetCache cacheKey result 600
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

function handleCreatePost() {
    # Get request data (simplified)
    title = GetHTTPQuery "title"
    content = GetHTTPQuery "content"
    authorId = GetHTTPQuery "author_id"
    
    # Insert post
    ExecDB "INSERT INTO posts (title, content, author_id) VALUES (?, ?, ?)" title content authorId
    result = GetQueryResult
    
    # Invalidate cache
    DeleteCache "posts:list"
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 201 result
}

function handleAddComment() {
    # Get request data
    postId = GetHTTPQuery "post_id"
    author = GetHTTPQuery "author"
    content = GetHTTPQuery "content"
    
    # Insert comment
    ExecDB "INSERT INTO comments (post_id, author, content) VALUES (?, ?, ?)" postId author content
    result = GetQueryResult
    
    # Invalidate post cache
    cacheKey = "post:" + postId
    DeleteCache cacheKey
    DeleteCache "posts:list"
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 201 result
}

function handleGetComments() {
    # Get post ID from query
    postId = GetHTTPQuery "post_id"
    
    # Query comments
    QueryDB "SELECT * FROM comments WHERE post_id = ? ORDER BY created_at DESC" postId
    result = GetQueryResult
    
    SetHTTPHeader "Content-Type" "application/json"
    SetHTTPResponse 200 result
}

# Register routes
RegisterHTTPRoute "GET" "/api/posts" "function" "handleGetPosts"
RegisterHTTPRoute "GET" "/api/post" "function" "handleGetPost"
RegisterHTTPRoute "POST" "/api/posts" "function" "handleCreatePost"
RegisterHTTPRoute "POST" "/api/comments" "function" "handleAddComment"
RegisterHTTPRoute "GET" "/api/comments" "function" "handleGetComments"

Println ""
Println "=== Blog API is running ==="
Println "Server: http://localhost:9188"
Println ""
Println "Available endpoints:"
Println "  GET  /api/posts - List all posts with comment counts (cached)"
Println "  GET  /api/post?id=1 - Get post by ID with views increment (cached)"
Println "  POST /api/posts?title=Title&content=Content&author_id=1 - Create a post"
Println "  POST /api/comments?post_id=1&author=Alice&content=Great! - Add a comment"
Println "  GET  /api/comments?post_id=1 - Get comments for a post"
Println ""
Println "Press Ctrl+C to stop the server"
