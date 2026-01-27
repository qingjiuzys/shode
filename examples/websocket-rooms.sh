#!/usr/bin/env shode
# WebSocket Rooms Example
# This demonstrates room-based messaging and broadcast features

# Start HTTP server
StartHTTPServer "8097"

# Register WebSocket route
RegisterWebSocketRoute "/ws" ""

# Register HTTP routes to manage WebSocket
function handleStats() {
    # Get connection count
    count=$(GetWebSocketConnectionCount)
    rooms=$(ListWebSocketRooms)

    SetHTTPResponse 200 "{
        \"status\": \"ok\",
        \"connections\": $count,
        \"rooms\": \"$rooms\"
    }"
}
RegisterHTTPRoute "GET" "/api/stats" "function" "handleStats"

function handleBroadcast() {
    # Get message from request body
    message=$(GetHTTPBody)

    # Broadcast to all connections
    BroadcastWebSocketMessage "$message"

    SetHTTPResponse 200 "{
        \"status\": \"broadcasted\",
        \"message\": \"$message\"
    }"
}
RegisterHTTPRoute "POST" "/api/broadcast" "function" "handleBroadcast"

function handleRoomBroadcast() {
    # Get room and message
    body=$(GetHTTPBody)
    # Parse JSON (simplified - assumes format: {"room":"name","message":"text"})
    room=$(echo "$body" | grep -o '"room":"[^"]*"' | cut -d'"' -f4)
    message=$(echo "$body" | grep -o '"message":"[^"]*"' | cut -d'"' -f4)

    # Broadcast to room
    BroadcastWebSocketMessageToRoom "$room" "$message"

    SetHTTPResponse 200 "{
        \"status\": \"broadcasted_to_room\",
        \"room\": \"$room\",
        \"message\": \"$message\"
    }"
}
RegisterHTTPRoute "POST" "/api/broadcast-room" "function" "handleRoomBroadcast"

Println "================================"
Println "WebSocket Rooms & Broadcast Demo"
Println "================================"
Println ""
Println "Server: http://localhost:8097"
Println "WebSocket: ws://localhost:8097/ws"
Println ""
Println "API Endpoints:"
Println "  GET  /api/stats       - Get connection statistics"
Println "  POST /api/broadcast   - Broadcast message to all"
Println "  POST /api/broadcast-room - Broadcast to specific room"
Println ""
Println "Example Usage:"
echo '  # Get stats'
echo '  curl http://localhost:8097/api/stats'
echo ''
echo '  # Broadcast to all clients'
echo '  curl -X POST -H "Content-Type: application/json" \\'
echo '    -d "{\"message\":\"Hello everyone!\"}" \\'
echo '    http://localhost:8097/api/broadcast'
echo ''
echo '  # Broadcast to room "general"'
echo '  curl -X POST -H "Content-Type: application/json" \\'
echo '    -d "{\"room\":\"general\",\"message\":\"Hello room!\"}" \\'
echo '    http://localhost:8097/api/broadcast-room'
echo ''
Println "Press Ctrl+C to stop the server"
Println "================================"

# Keep server running
for i in $(seq 1 100000); do
    sleep 1
done
