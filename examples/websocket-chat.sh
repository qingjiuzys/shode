#!/usr/bin/env shode
# WebSocket Chat Room Example
# This example demonstrates WebSocket real-time communication

# Start HTTP server
StartHTTPServer "8096"

# Register WebSocket route at /ws
# The second parameter is the handler function name (optional for this example)
RegisterWebSocketRoute "/ws" "handleWebSocketMessage"

Println "================================"
Println "WebSocket Chat Room Server"
Println "================================"
Println ""
Println "Server running at: http://localhost:8096"
Println "WebSocket endpoint: ws://localhost:8096/ws"
Println ""
Println "Features:"
Println "  ✓ Real-time bidirectional communication"
Println "  ✓ Broadcast to all connected clients"
Println "  ✓ Room-based messaging"
Println "  ✓ Connection management"
Println ""
Println "Test with:"
echo '  1. Open browser console and run:'
echo '     const ws = new WebSocket("ws://localhost:8096/ws");'
echo '     ws.onmessage = (e) => console.log("Received:", e.data);'
echo '     ws.send("Hello from browser!");'
echo ''
echo '  2. Or use websocat:'
echo '     websocat ws://localhost:8096/ws'
echo ''
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep server running
for i in $(seq 1 100000); do
    sleep 1
done
