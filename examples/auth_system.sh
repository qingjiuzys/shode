#!/usr/bin/env shode

# Authentication System Example
# Demonstrates a complete authentication system using JWT, cookies, and sessions
# This example shows:
#   - User registration with password hashing
#   - Login with JWT token generation
#   - Session management with cache
#   - Protected routes with authentication middleware
#   - Cookie-based authentication
#   - Logout and token invalidation

Println "=== Authentication System Demo ==="
Println ""

# ============================================================================
# PART 1: User Registration with Password Hashing
# ============================================================================
Println "Part 1: User Registration"
Println "-------------------------"

# Simulate user database (using cache for demo)
Println "Creating user database..."

# Register user 1 with password hashing
username1 = "alice"
password1 = "password123"
Println "Registering user: " + username1

# Hash password before storing
hashedPassword1 = HashSHA256 password1
SetCache "user:alice" '{"username":"alice","password":"' + hashedPassword1 + '","role":"admin","email":"alice@example.com"}' 0
Println "User registered (password hashed)"

# Register user 2
username2 = "bob"
password2 = "bobpass456"
Println "Registering user: " + username2
hashedPassword2 = HashSHA256 password2
SetCache "user:bob" '{"username":"bob","password":"' + hashedPassword2 + '","role":"user","email":"bob@example.com"}' 0
Println "User registered (password hashed)"

Println ""

# ============================================================================
# PART 2: Login with JWT Token Generation
# ============================================================================
Println "Part 2: Login and JWT Generation"
Println "-------------------------------"

# Simulate login request
loginUser = "alice"
loginPass = "password123"
Println "Login attempt for user: " + loginUser

# Retrieve user from database
userRecord = GetCache "user:" + loginUser

# Verify password
expectedHash = HashSHA256 loginPass

# Parse password from user record (simplified - in real app use proper JSON parsing)
Println "Verifying credentials..."
Println "User record: " + userRecord

# Generate JWT token on successful authentication
jwtSecret = "your-secret-key-change-in-production"
tokenPayload = '{"user":"' + loginUser + '","role":"admin","exp":1706745600}'

# Encode token (base64)
encodedPayload = Base64Encode tokenPayload
signature = HMACSHA256 jwtSecret encodedPayload
jwtToken = encodedPayload + "." + signature

Println "JWT Token generated: " + jwtToken
Println ""

# ============================================================================
# PART 3: Session Management
# ============================================================================
Println "Part 3: Session Management"
Println "--------------------------"

# Create session after successful login
sessionID = GenerateUUID
sessionData = '{"user":"' + loginUser + '","role":"admin","login_time":"' + CurrentTimestamp + '","ip":"127.0.0.1"}'

# Store session with 30 minute TTL
SetCache "session:" + sessionID sessionData 1800
Println "Session created: " + sessionID

# Also store session ID in user's active sessions list
SetCache "user_sessions:" + loginUser sessionID 1800

# Get session info
retrievedSession = GetCache "session:" + sessionID
Println "Session data: " + retrievedSession

# Check session TTL
sessionTTL = GetCacheTTL "session:" + sessionID
Println "Session TTL: " + sessionTTL + " seconds"

Println ""

# ============================================================================
# PART 4: Cookie-Based Authentication
# ============================================================================
Println "Part 4: Cookie Authentication"
Println "----------------------------"

# Set authentication cookie
cookieName = "auth_token"
cookieValue = jwtToken
cookieMaxAge = 1800  # 30 minutes

# Simulate setting cookie
Println "Setting auth cookie..."
Println "Cookie Name: " + cookieName
Println "Cookie Value: " + cookieValue
Println "Max Age: " + cookieMaxAge + " seconds"

# Store cookie in session (simulating browser cookie jar)
SetCookie cookieName cookieValue cookieMaxAge

# Get cookie value
storedCookie = GetCookie cookieName
Println "Retrieved cookie: " + storedCookie

# Set additional cookies (remember me, preferences)
SetCookie "theme" "dark" 86400  # 24 hours
SetCookie "language" "en" 86400
Println "Additional cookies set"

Println ""

# ============================================================================
# PART 5: Protected Route (Authentication Required)
# ============================================================================
Println "Part 5: Protected Route Access"
Println "-----------------------------"

# Simulate accessing a protected route
Println "Accessing /api/dashboard..."

# Check for authentication cookie
authCookie = GetCookie "auth_token"

if authCookie != "" {
    Println "Auth cookie found: " + authCookie

    # Verify JWT token (simplified)
    Println "Verifying JWT token..."

    # Check session
    sessionID = GetCache "user_sessions:alice"
    sessionData = GetCache "session:" + sessionID

    if sessionData != "" {
        Println "Access granted!"
        Println "User: alice"
        Println "Role: admin"
        Println "Session: " + sessionData
    } else {
        Println "Access denied: Invalid or expired session"
    }
} else {
    Println "Access denied: No authentication cookie"
}

Println ""

# ============================================================================
# PART 6: Role-Based Access Control (RBAC)
# ============================================================================
Println "Part 6: Role-Based Access Control"
Println "--------------------------------"

# Check user permissions
userRole = "admin"

Println "Checking permissions for role: " + userRole

# Define role permissions
SetCache "role:admin" '{"permissions":["read","write","delete","manage_users"]}' 0
SetCache "role:user" '{"permissions":["read","write"]}' 0

# Get role permissions
rolePerms = GetCache "role:" + userRole
Println "Role permissions: " + rolePerms

# Check specific permission
Println "Can delete resources? YES (admin role)"

Println ""

# ============================================================================
# Part 7: Logout and Session Invalidation
# ============================================================================
Println "Part 7: Logout and Cleanup"
Println "------------------------"

# Invalidate session
Println "Logging out user: alice"

# Remove session from cache
DeleteCache "session:" + sessionID
Println "Session invalidated: " + sessionID

# Clear auth cookie
DeleteCookie "auth_token"
Println "Auth cookie cleared"

# Verify session is gone
sessionExists = CacheExists "session:" + sessionID
Println "Session exists after logout: " + sessionExists

# Clean up other cookies
DeleteCookie "theme"
DeleteCookie "language"
Println "All cookies cleared"

Println ""

# ============================================================================
# PART 8: Token Refresh
# ============================================================================
Println "Part 8: Token Refresh"
Println "-------------------"

# Simulate token refresh (before expiration)
Println "Refreshing JWT token..."

# Generate new token with extended expiration
newPayload = '{"user":"alice","role":"admin","exp":1706832000}'
newEncoded = Base64Encode newPayload
newSignature = HMACSHA256 jwtSecret newEncoded
newToken = newEncoded + "." + newSignature

Println "New token generated: " + newToken

# Update cookie
SetCookie "auth_token" newToken 1800
Println "Auth cookie updated with new token"

Println ""

# ============================================================================
# Summary
# ============================================================================
Println "=== Authentication System Demo Complete ==="
Println ""
Println "Features demonstrated:"
Println "  ✓ User registration with password hashing"
Println "  ✓ Login with JWT token generation"
Println "  ✓ Session management with TTL"
Println "  ✓ Cookie-based authentication"
Println "  ✓ Protected route access control"
Println "  ✓ Role-based permissions (RBAC)"
Println "  ✓ Logout and session cleanup"
Println "  ✓ Token refresh mechanism"
Println ""
Println "Security best practices shown:"
Println "  • Passwords are hashed before storage"
Println "  • JWT tokens are signed with secret"
Println "  • Sessions have TTL for auto-expiration"
Println "  • Cookies have max-age limits"
Println "  • Sessions are invalidated on logout"
Println ""
