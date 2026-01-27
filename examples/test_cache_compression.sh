#!/usr/bin/env shode
# Test streaming Gzip compression and cache headers
# This example demonstrates the performance improvements in v0.5.1

# Start HTTP server
StartHTTPServer "8095"

# Register static route with Gzip enabled
# Usage: RegisterStaticRouteAdvanced "/" "directory" "indexFiles" "directoryBrowse" "cacheControl" "enableGzip" "spaFallback"
RegisterStaticRouteAdvanced "/" "examples/test_static" "index.html" "false" "max-age=3600" "true" ""

Println "================================"
Println "Streaming Gzip & Cache Test Server"
Println "================================"
Println ""
Println "Server running at: http://localhost:8095"
Println ""
Println "Features enabled:"
Println "  ✓ Streaming Gzip compression (memory efficient)"
Println "  ✓ ETag support (strong ETag based on file metadata)"
Println "  ✓ Last-Modified header"
Println "  ✓ Cache-Control: max-age=3600"
Println ""
Println "Test URLs:"
Println "  - http://localhost:8095/              (index page)"
Println "  - http://localhost:8095/test.html     (test page)"
Println "  - http://localhost:8095/style.css     (CSS file)"
Println ""
Println "Test commands in another terminal:"
Println "  # Check ETag and Last-Modified headers"
Println "  curl -I http://localhost:8095/test.html"
Println ""
Println "  # Test conditional request with If-None-Match (should return 304)"
Println "  ETAG=\$(curl -I http://localhost:8095/test.html 2>&1 | grep -i etag | cut -d' ' -f2 | tr -d '\\r')"
Println "  curl -I -H \"If-None-Match: \$ETAG\" http://localhost:8095/test.html"
Println ""
Println "  # Test conditional request with If-Modified-Since"
Println "  curl -I -H 'If-Modified-Since: Wed, 21 Oct 2015 07:28:00 GMT' http://localhost:8095/test.html"
Println ""
Println "  # Check compression (X-Compression-Ratio header)"
Println "  curl -I -H \"Accept-Encoding: gzip\" http://localhost:8095/test.html"
Println ""
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep server running
for i in $(seq 1 100000); do
    sleep 1
done
