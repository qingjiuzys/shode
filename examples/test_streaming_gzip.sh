#!/usr/bin/env shode
# Test streaming Gzip compression and cache headers
# This example demonstrates the performance improvements in v0.5.1

# Start HTTP server
StartHTTPServer "8094"

# Create a test directory with a large file
TEST_DIR="/tmp/shode_test_static"
mkdir -p "$TEST_DIR"

# Create a 10MB test file for compression testing
Println "Creating 10MB test file..."
dd if=/dev/zero of="$TEST_DIR/large.txt" bs=1024 count=10240 2>/dev/null

# Create a test HTML file
echo '<html><body><h1>Test Page</h1></body></html>' > "$TEST_DIR/test.html"

# Register static route with Gzip enabled
# Usage: RegisterStaticRouteAdvanced "/" "directory" "indexFiles" "directoryBrowse" "cacheControl" "enableGzip" "spaFallback"
RegisterStaticRouteAdvanced "/" "$TEST_DIR" "index.html" "false" "max-age=3600" "true" ""

Println "================================"
Println "Streaming Gzip & Cache Test Server"
Println "================================"
Println ""
Println "Server running at: http://localhost:8094"
Println ""
Println "Features enabled:"
Println "  ✓ Streaming Gzip compression (memory efficient)"
Println "  ✓ ETag support (strong ETag based on file metadata)"
Println "  ✓ Last-Modified header"
Println "  ✓ Cache-Control: max-age=3600"
Println ""
Println "Test URLs:"
Println "  - http://localhost:8094/test.html  (small file with cache headers)"
Println "  - http://localhost:8094/large.txt   (10MB file with streaming compression)"
Println ""
Println "Test commands:"
echo '  # Check ETag and Last-Modified headers'
echo '  curl -I http://localhost:8094/test.html'
echo ''
echo '  # Test conditional request with If-None-Match (should return 304)'
echo '  ETAG=$(curl -I http://localhost:8094/test.html | grep -i etag | cut -d" " -f2)'
echo '  curl -I -H "If-None-Match: $ETAG" http://localhost:8094/test.html'
echo ''
echo '  # Check compression ratio (look at X-Compression-Ratio header)'
echo '  curl -I -H "Accept-Encoding: gzip" http://localhost:8094/large.txt'
echo ''
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep server running
for i in $(seq 1 100000); do
    sleep 1
done

# Cleanup on exit
# rm -rf "$TEST_DIR"
