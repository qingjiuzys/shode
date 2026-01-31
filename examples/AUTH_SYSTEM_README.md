# Authentication System Example

A comprehensive authentication system demonstration using Shode, featuring JWT tokens, session management, cookies, and role-based access control.

## Features Demonstrated

### 1. User Registration with Password Hashing
- Secure password storage using SHA-256 hashing
- User profile creation with roles
- Email and metadata handling

### 2. Login with JWT Token Generation
- JWT (JSON Web Token) creation
- Token signing with HMAC-SHA256
- Base64 encoding for payload
- Secure credential verification

### 3. Session Management
- Session creation with unique UUID
- Time-based session expiration (TTL)
- Session storage in cache
- Session retrieval and validation

### 4. Cookie-Based Authentication
- HTTP cookie setting and retrieval
- Cookie expiration management
- Multiple cookie support
- Secure cookie handling

### 5. Protected Routes
- Authentication middleware simulation
- JWT token verification
- Session validation
- Access control logic

### 6. Role-Based Access Control (RBAC)
- Role definitions (admin, user)
- Permission mapping
- Access control checks
- Authorization logic

### 7. Logout and Cleanup
- Session invalidation
- Cookie removal
- Cache cleanup
- Security best practices

### 8. Token Refresh
- Token renewal before expiration
- Extended session duration
- Cookie updates
- Seamless user experience

## Running the Example

```bash
cd examples
./auth_system.sh
```

Or using the Shode CLI:

```bash
shode auth_system.sh
```

## Security Considerations

This example demonstrates several security best practices:

1. **Password Hashing**: Passwords are never stored in plain text
2. **JWT Signing**: Tokens are cryptographically signed
3. **Session TTL**: Sessions automatically expire
4. **Cookie Limits**: Cookies have max-age constraints
5. **Clean Logout**: Sessions are invalidated on logout

### Important Notes for Production

⚠️ **This is a demonstration example. For production use, consider:**

1. Use stronger password hashing (bcrypt, Argon2)
2. Use longer JWT secrets from environment variables
3. Implement HTTPS for secure cookie transmission
4. Add CSRF protection
5. Implement rate limiting for login attempts
6. Add password reset functionality
7. Use proper JSON parsing libraries
8. Implement refresh token rotation
9. Add audit logging
10. Use database instead of cache for user storage

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│       Authentication Flow       │
├─────────────────────────────────┤
│ 1. Register → Hash Password     │
│ 2. Login → Verify Credentials   │
│ 3. Generate JWT Token           │
│ 4. Create Session               │
│ 5. Set Auth Cookie              │
│ 6. Access Protected Routes      │
│ 7. Verify Token & Session       │
│ 8. Logout → Invalidate Session  │
└─────────────────────────────────┘
```

## API Flow

### Registration
```
POST /api/register
{
  "username": "alice",
  "password": "password123",
  "email": "alice@example.com"
}

→ Hash password with SHA-256
→ Store user in database
→ Return success
```

### Login
```
POST /api/login
{
  "username": "alice",
  "password": "password123"
}

→ Retrieve user from database
→ Verify password hash
→ Generate JWT token
→ Create session with TTL
→ Set auth cookie
→ Return token
```

### Protected Route
```
GET /api/dashboard
Headers: Cookie: auth_token=<jwt>

→ Verify JWT signature
→ Check session exists
→ Validate user permissions
→ Return protected data
```

### Logout
```
POST /api/logout
Headers: Cookie: auth_token=<jwt>

→ Invalidate session
→ Clear auth cookie
→ Return success
```

## Code Examples

### Middleware Integration

```javascript
// Authentication middleware for protected routes
function requireAuth(req, res, next) {
    const token = req.cookies.auth_token;

    if (!token) {
        return res.status(401).json({ error: "Unauthorized" });
    }

    try {
        const decoded = verifyJWT(token);
        const session = getSession(decoded.user);

        if (!session) {
            return res.status(401).json({ error: "Invalid session" });
        }

        req.user = decoded;
        next();
    } catch (error) {
        return res.status(401).json({ error: "Invalid token" });
    }
}
```

### Role-Based Access

```javascript
function requireRole(role) {
    return (req, res, next) => {
        if (req.user.role !== role) {
            return res.status(403).json({ error: "Forbidden" });
        }
        next();
    };
}

// Usage
app.get("/api/admin", requireAuth, requireRole("admin"), handler);
```

## Testing the Example

Run the script and observe the output:

```bash
$ ./auth_system.sh

=== Authentication System Demo ===

Part 1: User Registration
-------------------------
Registering user: alice
User registered (password hashed)

Part 2: Login and JWT Generation
-------------------------------
Login attempt for user: alice
JWT Token generated: eyJ1c2VyIjoiYWxpY2U...

...
```

## Related Examples

- `session_management.sh` - Basic session handling
- `user_management.sh` - CRUD operations for users
- `rate_limiting.sh` - API rate limiting

## Further Reading

- [JWT Specification](https://tools.ietf.org/html/rfc7519)
- [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
- [Session Management Best Practices](https://owasp.org/www-project-cheat-sheets/cheatsheets/Session_Management_Cheat_Sheet.html)
