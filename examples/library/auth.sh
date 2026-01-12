#!/usr/bin/env shode

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
