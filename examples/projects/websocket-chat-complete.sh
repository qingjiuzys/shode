#!/usr/bin/env shode
#
# WebSocket 聊天室完整示例
# 功能：实时聊天、房间管理、用户列表
#

# 启动 HTTP 服务器
StartHTTPServer "8098"

# 注册 WebSocket 路由
RegisterWebSocketRoute "/ws" "handleMessage"

# ==================== HTTP API ====================

# 1. 获取服务器统计信息
function getStats() {
    connCount := GetWebSocketConnectionCount
    rooms := ListWebSocketRooms
    
    stats := '{"connections":' $connCount ',"rooms":' $rooms '}'
    SetHTTPResponse 200 $stats
}
RegisterHTTPRoute "GET" "/api/stats" "function" "getStats"

# 2. 广播消息到所有连接
function broadcast() {
    body := GetHTTPBody
    BroadcastWebSocketMessage $body
    SetHTTPResponse 200 '{"status":"broadcasted"}'
}
RegisterHTTPRoute "POST" "/api/broadcast" "function" "broadcast"

# 3. 广播消息到特定房间
function broadcastToRoom() {
    room := GetHTTPQuery "room"
    body := GetHTTPBody
    
    err := BroadcastWebSocketMessageToRoom $room $body
    if $err; then
        SetHTTPResponse 400 '{"error":"Room not found"}'
        return
    fi
    
    SetHTTPResponse 200 '{"status":"broadcasted to room"}'
}
RegisterHTTPRoute "POST" "/api/broadcast-room" "function" "broadcastToRoom"

# 4. 获取房间信息
function getRoomInfo() {
    room := GetHTTPQuery "room"
    count := GetWebSocketRoomCount $room
    
    info := '{"room":"' $room '","users":' $count '}'
    SetHTTPResponse 200 $info
}
RegisterHTTPRoute "GET" "/api/room" "function" "getRoomInfo"

# 5. 静态文件服务
RegisterStaticRoute "/" "./examples/projects/public"

# ==================== WebSocket 处理 ====================

# WebSocket 消息处理函数
function handleMessage() {
    method := GetHTTPMethod
    path := GetHTTPPath
    
    # 处理 WebSocket 连接
    echo "WebSocket connection established"
    
    for i in $(seq 1 100000); do
        sleep 1
    done
}

# ==================== 使用说明 ====================

echo "========================================="
echo "  WebSocket Chat Server"
echo "========================================="
echo ""
echo "服务器信息："
echo "  WebSocket: ws://localhost:8098/ws"
echo "  HTTP API:  http://localhost:8098/api/"
echo ""
echo "API 端点："
echo "  GET  /api/stats         - 获取统计信息"
echo "  POST /api/broadcast     - 广播消息"
echo "  POST /api/broadcast-room?room=<name> - 广播到房间"
echo "  GET  /api/room?room=<name> - 获取房间信息"
echo ""
echo "示例："
echo "  # 广播消息"
echo "  curl -X POST http://localhost:8098/api/broadcast \\"
echo "    -d 'Hello everyone!'"
echo ""
echo "  # 获取统计"
echo "  curl http://localhost:8098/api/stats"
echo ""
echo "========================================="

# 保持服务器运行
for i in $(seq 1 100000); do sleep 1; done
