#!/usr/bin/env shode
#
# REST API 示例：带缓存的用户管理 API
# 功能：CRUD 操作 + 缓存优化
#

# 启动 HTTP 服务器
StartHTTPServer "8099"

# ==================== 数据库初始化 ====================

ConnectDB "sqlite" "./users.db"

# 创建表
ExecDB "CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)"

# ==================== API 路由 ====================

# 1. 获取用户列表（带缓存）
function listUsers() {
    # 检查缓存
    cached, exists := GetCache "users:list"
    if $exists; then
        SetHTTPResponse 200 $cached
        return
    fi
    
    # 查询数据库
    result := QueryDB "SELECT id, name, email FROM users ORDER BY id"
    
    # 缓存结果（5 分钟）
    SetCache "users:list" $result 300
    
    SetHTTPResponse 200 $result
}
RegisterHTTPRoute "GET" "/api/users" "function" "listUsers"

# 2. 获取单个用户（带缓存）
function getUser() {
    id := GetHTTPQuery "id"
    
    # 检查缓存
    cacheKey := "user:" $id
    cached, exists := GetCache $cacheKey
    if $exists; then
        SetHTTPResponse 200 $cached
        return
    fi
    
    # 查询数据库
    result := QueryRowDB "SELECT id, name, email FROM users WHERE id = ?" $id
    if IsEmpty $result; then
        SetHTTPResponse 404 '{"error":"User not found"}'
        return
    fi
    
    # 缓存结果（10 分钟）
    SetCache $cacheKey $result 600
    
    SetHTTPResponse 200 $result
}
RegisterHTTPRoute "GET" "/api/user" "function" "getUser"

# 3. 创建用户
function createUser() {
    name := GetHTTPQuery "name"
    email := GetHTTPQuery "email"
    
    # 验证输入
    if IsEmpty $name; then
        SetHTTPResponse 400 '{"error":"Name is required"}'
        return
    fi
    
    if IsEmpty $email; then
        SetHTTPResponse 400 '{"error":"Email is required"}'
        return
    fi
    
    # 插入数据库
    err := ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" $name $email
    if $err; then
        SetHTTPResponse 400 '{"error":"Email already exists"}'
        return
    fi
    
    # 清除列表缓存
    DeleteCache "users:list"
    
    SetHTTPResponse 201 '{"status":"created"}'
}
RegisterHTTPRoute "POST" "/api/users" "function" "createUser"

# 4. 更新用户
function updateUser() {
    id := GetHTTPQuery "id"
    name := GetHTTPQuery "name"
    email := GetHTTPQuery "email"
    
    # 更新数据库
    ExecDB "UPDATE users SET name = ?, email = ? WHERE id = ?" $name $email $id
    
    # 清除缓存
    DeleteCache "user:" $id
    DeleteCache "users:list"
    
    SetHTTPResponse 200 '{"status":"updated"}'
}
RegisterHTTPRoute "PUT" "/api/user" "function" "updateUser"

# 5. 删除用户
function deleteUser() {
    id := GetHTTPQuery "id"
    
    # 删除记录
    ExecDB "DELETE FROM users WHERE id = ?" $id
    
    # 清除缓存
    DeleteCache "user:" $id
    DeleteCache "users:list"
    
    SetHTTPResponse 200 '{"status":"deleted"}'
}
RegisterHTTPRoute "DELETE" "/api/user" "function" "deleteUser"

# 6. 清空所有缓存
function clearCache() {
    ClearCache
    SetHTTPResponse 200 '{"status":"cache cleared"}'
}
RegisterHTTPRoute "POST" "/api/cache/clear" "function" "clearCache"

# 7. 获取缓存统计
function cacheStats() {
    # 模拟缓存统计
    SetHTTPResponse 200 '{"message":"Cache statistics endpoint"}'
}
RegisterHTTPRoute "GET" "/api/cache/stats" "function" "cacheStats"

# ==================== 使用说明 ====================

echo "========================================="
echo "  REST API with Cache"
echo "========================================="
echo ""
echo "API 端点："
echo "  GET    /api/users          - 获取用户列表（缓存 5 分钟）"
echo "  GET    /api/user?id=<id>   - 获取用户详情（缓存 10 分钟）"
echo "  POST   /api/users          - 创建用户"
echo "          参数: name, email"
echo "  PUT    /api/user?id=<id>   - 更新用户"
echo "          参数: name, email"
echo "  DELETE /api/user?id=<id>   - 删除用户"
echo "  POST   /api/cache/clear    - 清空缓存"
echo "  GET    /api/cache/stats    - 缓存统计"
echo ""
echo "示例："
echo "  # 创建用户"
echo "  curl 'http://localhost:8099/api/users?name=Alice&email=alice@example.com' -X POST"
echo ""
echo "  # 获取用户列表"
echo "  curl http://localhost:8099/api/users"
echo ""
echo "  # 获取单个用户"
echo "  curl 'http://localhost:8099/api/user?id=1'"
echo ""
echo "  # 更新用户"
echo "  curl 'http://localhost:8099/api/user?id=1&name=Alice+Smith' -X PUT"
echo ""
echo "  # 删除用户"
echo "  curl 'http://localhost:8099/api/user?id=1' -X DELETE"
echo ""
echo "  # 清空缓存"
echo "  curl http://localhost:8099/api/cache/clear -X POST"
echo ""
echo "========================================="

# 保持服务器运行
for i in $(seq 1 100000); do sleep 1; done
