# Shode Project Examples

This directory contains real-world project examples demonstrating various use cases of Shode's static file server and HTTP capabilities.

## Available Examples

### 1. Personal Website / Blog (`personal-website.sh`)

**Use Case:** Hosting a simple personal website or blog

**Features:**
- Static HTML pages
- Blog section with multiple posts
- RESTful API for statistics
- Simple, clean design

**Run:**
```bash
./shode run examples/projects/personal-website.sh
```

**Access:** http://localhost:3000

**Includes:**
- Home page at `/`
- Blog section at `/blog/`
- About page at `/about.html`
- Stats API at `/api/stats`

---

### 2. API Documentation Server (`api-docs-server.sh`)

**Use Case:** Hosting searchable API documentation

**Features:**
- Directory browsing enabled for easy navigation
- Multiple documentation versions
- Static assets with caching
- Search API endpoint

**Run:**
```bash
./shode run examples/projects/api-docs-server.sh
```

**Access:** http://localhost:8080

**Includes:**
- Documentation browser at `/docs`
- Static assets at `/assets`
- Search API at `/api/search`

---

### 3. Full-Stack Application (`fullstack-app.sh`)

**Use Case:** Complete web application with frontend and backend

**Features:**
- SPA (Single Page Application) support
- RESTful API with CRUD operations
- Client-side routing fallback
- Health check endpoint
- JSON responses

**Run:**
```bash
./shode run examples/projects/fullstack-app.sh
```

**Access:** http://localhost:4000

**API Endpoints:**
- `GET /api/users` - List all users
- `GET /api/users/1` - Get user by ID
- `POST /api/users` - Create new user
- `GET /api/health` - Health check

---

### 4. File Download Server (`file-server.sh`)

**Use Case:** Distributing software releases and files

**Features:**
- Optimized for downloads (long cache times)
- Separate sections for downloads and release notes
- Directory browsing for release notes
- Latest version API

**Run:**
```bash
./shode run examples/projects/file-server.sh
```

**Access:** http://localhost:5000

**Includes:**
- File downloads at `/downloads`
- Release notes at `/releases` (with browsing)
- File list API at `/api/files`
- Latest version API at `/api/latest`

---

## Quick Start

1. **Choose an example** that matches your use case
2. **Run the script:**
   ```bash
   ./shode run examples/projects/[example-name].sh
   ```
3. **Open your browser** and navigate to the specified port
4. **Stop the server:** Press `Ctrl+C`

## Customization

Each example can be easily customized:

### Change Port
Edit the `StartHTTPServer` line:
```bash
StartHTTPServer "8080"  # Change to your preferred port
```

### Change Directory
Edit the `RegisterStaticRoute` line:
```bash
RegisterStaticRoute "/" "./your-directory"
```

### Add API Endpoints
```bash
function yourFunction() {
    SetHTTPResponse 200 '{"status":"ok"}'
}
RegisterHTTPRoute "GET" "/api/endpoint" "function" "yourFunction"
```

## Common Patterns

### Serving a React/Vue SPA
```bash
RegisterStaticRouteAdvanced "/" "./dist" \
    "index.html" \
    "false" \
    "max-age=3600" \
    "true" \
    "index.html"  # SPA fallback
```

### Enabling Directory Browsing
```bash
RegisterStaticRouteAdvanced "/docs" "./docs" \
    "" \
    "true" \
    "" \
    "false" \
    ""
```

### Long Cache for Assets
```bash
RegisterStaticRouteAdvanced "/static" "./static" \
    "" \
    "false" \
    "max-age=86400" \
    "true" \
    ""
```

## Next Steps

- Check the [Static File Server Documentation](../STATIC_FILE_SERVER.md) for detailed API reference
- See the [Implementation Guide](../../docs/STATIC_FILE_SERVER_IMPLEMENTATION.md) for technical details
- Explore the source code to understand how each feature works

## Troubleshooting

**Port already in use:**
```bash
# Find process using the port
lsof -i :8080

# Kill the process
kill -9 [PID]
```

**Files not found:**
- Make sure you're running the script from the project root directory
- Check that the specified directories exist relative to where you run the script

**API not responding:**
- Check the server logs for error messages
- Verify the function is defined before calling `RegisterHTTPRoute`
- Ensure you're using the correct HTTP method (GET, POST, etc.)

## Contributing

Have a great example to share? Feel free to submit a pull request!
