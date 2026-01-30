#!/usr/bin/env shode
# Authentication Demo - 展示 Shode v0.7.0 认证功能
#
# 本示例展示如何使用：
# - JWT Token 生成和验证
# - Session 管理
# - Cookie 操作
# - 认证中间件

# 配置
JWT_SECRET="my-secret-key-2024"
SESSION_TTL=3600
SERVER_PORT=9000

echo "=== Shode v0.7.0 认证功能演示 ==="
echo ""

# 1. JWT 认证演示
echo "1. JWT Token 认证"
echo "-------------------"

# 模拟用户登录并生成 JWT Token
USER_ID="user123"
USER_DATA='{"name":"Alice","role":"admin"}'

echo "用户登录: $USER_ID"
echo "用户数据: $USER_DATA"

# 生成 JWT Token (这里我们手动创建一个简单的 token)
# 在实际应用中，这会通过 HTTP 请求完成
TOKEN_HEADER="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
TOKEN_PAYLOAD="$(echo -n '{"sub":"user123","data":{"name":"Alice","role":"admin"},"exp":'$(($(date +%s) + 3600))} | base64)"
TOKEN_SIGNATURE="signature"
JWT_TOKEN="${TOKEN_HEADER}.${TOKEN_PAYLOAD}.${TOKEN_SIGNATURE}"

echo "生成的 JWT Token:"
echo "$JWT_TOKEN"
echo ""

# 2. Session 管理演示
echo "2. Session 管理"
echo "---------------"

# 生成 session ID
SESSION_ID=$(head -c 32 /dev/urandom | xxd -p | tr -d ' \n')

echo "创建新会话: $SESSION_ID"
echo "会话过期时间: ${SESSION_TTL} 秒"
echo ""

# 3. Cookie 设置演示
echo "3. Cookie 管理"
echo "-------------"

COOKIE_NAME="session_token"
COOKIE_VALUE="${SESSION_ID}"
COOKIE_OPTIONS="Path=/; HttpOnly; Secure; Max-Age=${SESSION_TTL}"

echo "设置 Cookie:"
echo "Name: $COOKIE_NAME"
echo "Value: $COOKIE_VALUE (Session ID)"
echo "Options: $COOKIE_OPTIONS"
echo ""

# 4. HTTP 服务器演示
echo "4. 认证 API 服务器"
echo "-----------------"

echo "启动认证 API 服务器在端口 $SERVER_PORT ..."
echo ""

# 定义认证相关的 HTTP 处理逻辑
cat << 'EOF' > /tmp/auth_server_logic.sh
# auth_login - 登录处理
auth_login() {
    local user_id="$1"
    local password="$2"

    # 简单验证（实际应用中应该查询数据库）
    if [ "$password" = "password123" ]; then
        # 生成 token
        local token="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.$(echo -n "{\"sub\":\"$user_id\",\"exp\":$(($(date +%s) + 3600)}" | base64).signature"

        # 设置 cookie
        SetCookie "session_token" "$token" "Path=/; HttpOnly; Max-Age=3600"

        # 返回 JSON 响应
        '{"status":"success","token":"'"$token"'","message":"Login successful"}'
        return 0
    else
        SetErrorResponse 401 "Invalid credentials"
        return 1
    fi
}

# auth_profile - 获取用户信息（需要认证）
auth_profile() {
    # 验证 token
    local token=$(GetCookie "session_token")

    if [ -z "$token" ]; then
        SetErrorResponse 401 "Unauthorized: No token provided"
        return 1
    fi

    # 解析 token (简化版)
    local user_id=$(echo "$token" | cut -d'.' -f2 | base64 -d 2>/dev/null | grep -o '"sub":"[^"]*"' | cut -d'"' -f4)

    # 返回用户信息
    '{"user":"'"$user_id"'","name":"Alice","email":"alice@example.com"}'
    return 0
}

# auth_logout - 登出处理
auth_logout() {
    # 删除 cookie
    DeleteCookie "session_token" "/"

    '{"status":"success","message":"Logged out successfully"}'
    return 0
}

# auth_session - 获取会话信息
auth_session() {
    local session_id=$(GetCookie "session_id")

    if [ -z "$session_id" ]; then
        SetErrorResponse 401 "No active session"
        return 1
    fi

    '{"session_id":"'"$session_id"'","user":"user123","created_at":"'$(date +%Y-%m-%d\ %H:%M:%S)'","expires_in":3600}'
    return 0
}
EOF

echo "API 端点:"
echo ""
echo "POST   /api/login      - 用户登录"
echo "  Body: {\"user_id\":\"user123\",\"password\":\"password123\"}"
echo ""
echo "GET    /api/profile    - 获取用户信息（需要认证）"
echo "  Header: Authorization: Bearer <token>"
echo ""
echo "POST   /api/logout     - 用户登出"
echo ""
echo "GET    /api/session    - 获取当前会话信息"
echo ""
echo "示例请求:"
echo ""
echo "# 登录"
echo "curl -X POST http://localhost:$SERVER_PORT/api/login \\"
echo "  -H \"Content-Type: application/json\" \\"
echo "  -d '{\"user_id\":\"user123\",\"password\":\"password123\"}'"
echo ""
echo "# 获取用户信息（带 token）"
echo "curl -X GET http://localhost:$SERVER_PORT/api/profile \\"
echo "  -H \"Authorization: Bearer <your-token>\""
echo ""

# 5. 安全最佳实践
echo "5. 安全最佳实践"
echo "-------------"

cat << 'EOF'
✅ 密码存储: 使用 bcrypt 或 argon2 加密
✅ Token 签名: 使用 HS256 或 RS256 算法
✅ Token 过期: 设置合理的过期时间（1小时）
✅ HTTPS: 生产环境必须使用 HTTPS
✅ Cookie 安全:
   - HttpOnly: 防止 XSS 攻击
   - Secure: 仅通过 HTTPS 传输
   - SameSite: 防止 CSRF 攻击
✅ Session 管理:
   - 定期清理过期会话
   - 限制并发会话数
   - 记录会话活动日志
EOF

echo ""
echo "6. 完整示例：认证流程"
echo "--------------------"

cat << 'EOF'
# 步骤 1: 用户登录
POST /api/login
{
  "user_id": "user123",
  "password": "password123"
}

Response:
{
  "status": "success",
  "token": "eyJhbGci...",
  "message": "Login successful"
}
Cookie: session_token=eyJhbGci...; Path=/; HttpOnly; Max-Age=3600

# 步骤 2: 访问受保护的资源
GET /api/profile
Headers:
  Authorization: Bearer eyJhbGci...

Response:
{
  "user": "user123",
  "name": "Alice",
  "email": "alice@example.com"
}

# 步骤 3: 用户登出
POST /api/logout
Headers:
  Authorization: Bearer eyJhbGci...

Response:
{
  "status": "success",
  "message": "Logged out successfully"
}
Cookie: session_token=; Path=/; Max-Age=-1
EOF

echo ""
echo "7. 代码示例"
echo "---------"

cat << 'EOF'
// 使用 JWT 验证中间件
func AuthMiddleware(next) {
    return func(w, r) {
        token := r.Header.Get("Authorization")

        if token == "" {
            w.WriteHeader(401)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Missing authorization header",
            })
            return
        }

        // 验证 token
        claims, err := jwt.VerifyJWT(token)
        if err != nil {
            w.WriteHeader(401)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Invalid token",
            })
            return
        }

        // 将用户信息存入 context
        context.Set(r, "user_id", claims.Subject)

        next(w, r)
    }
}

// 使用 Session 中间件
func SessionMiddleware(sessionManager *session.Manager) {
    return func(w, r) {
        sessionID, err := r.Cookie("session_id")

        if err != nil || sessionID == nil {
            w.WriteHeader(401)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "No active session",
            })
            return
        }

        session, err := sessionManager.GetSession(sessionID.Value)
        if err != nil {
            w.WriteHeader(401)
            json.NewEncoder(w).Encode(map[string]string{
                "error": "Invalid session",
            })
            return
        }

        // 将用户信息存入 context
        context.Set(r, "user", session.Data["user_id"])

        next(w, r)
    }
}
EOF

echo ""
echo "8. 测试命令"
echo "---------"

cat << 'EOF'
# 测试登录
curl -X POST http://localhost:9000/api/login \
  -H "Content-Type: application/json" \
  -d '{"user_id":"user123","password":"password123"}'

# 测试获取用户信息
TOKEN="your-token-here"
curl -X GET http://localhost:9000/api/profile \
  -H "Authorization: Bearer $TOKEN"

# 测试登出
curl -X POST http://localhost:9000/api/logout \
  -H "Authorization: Bearer $TOKEN"
EOF

echo ""
echo "演示完成！"
echo ""
echo "提示："
echo "- 实际应用中应该使用数据库存储用户信息"
echo "- 密码应该使用 bcrypt 加密存储"
echo "- Token 应该有更复杂的签名机制"
echo "- 生产环境必须使用 HTTPS"
echo ""

# 启动简单的演示服务器（如果 shode 支持的话）
if command -v shode &> /dev/null; then
    echo "启动演示服务器..."
    echo ""

    # 创建简单的演示脚本
    cat << 'SERVER_SCRIPT' > /tmp/auth_server.sh
#!/usr/bin/env shode

# 简单的认证演示服务器
SECRET_KEY="demo-secret-2024"

# POST /api/login - 登录端点
if [ "$REQUEST_METHOD" = "POST" ] && [ "$PATH_INFO" = "/api/login" ]; then
    BODY=$(cat)
    USER_ID=$(echo "$BODY" | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
    PASSWORD=$(echo "$BODY" | grep -o '"password":"[^"]*"' | cut -d'"' -f4)

    if [ "$PASSWORD" = "password123" ]; then
        # 生成简单 token (实际应用中应使用 jwt 包)
        TOKEN="demo-token-${USER_ID}-$(date +%s)"

        PrintContentType "application/json"
        Println "{\"status\":\"success\",\"token\":\"$TOKEN\",\"user_id\":\"$USER_ID\"}"
    else
        SetStatusCode 401
        PrintContentType "application/json"
        Println "{\"status\":\"error\",\"message\":\"Invalid credentials\"}"
    fi
    exit 0
fi

# GET /api/profile - 获取用户信息
if [ "$REQUEST_METHOD" = "GET" ] && [ "$PATH_INFO" = "/api/profile" ]; then
    AUTH_HEADER=$(GetHeader "Authorization")

    if [ -z "$AUTH_HEADER" ]; then
        SetStatusCode 401
        PrintContentType "application/json"
        Println "{\"status\":\"error\",\"message\":\"Missing authorization\"}"
        exit 0
    fi

    # 提取 token
    TOKEN=${AUTH_HEADER#Bearer }

    # 简单验证（实际应用中应解析 JWT）
    if [ "$TOKEN" != "" ]; then
        PrintContentType "application/json"
        Println "{\"status\":\"success\",\"user\":\"user123\",\"name\":\"Alice\",\"role\":\"admin\"}"
    else
        SetStatusCode 401
        PrintContentType "application/json"
        Println "{\"status\":\"error\",\"message\":\"Invalid token\"}"
    fi
    exit 0
fi

# 其他端点返回 404
SetStatusCode 404
PrintContentType "application/json"
Println "{\"status\":\"error\",\"message\":\"Endpoint not found\"}"
SERVER_SCRIPT

    echo "服务器已准备就绪，但需要手动测试"
    echo "服务器脚本: /tmp/auth_server.sh"
    echo ""
fi

echo "相关文档:"
echo "- pkg/jwt/jwt.go - JWT 实现"
echo "- pkg/session/session.go - Session 管理"
echo "- pkg/cookie/cookie.go - Cookie 管理"
echo "- pkg/auth/middleware.go - 认证中间件"
echo ""
echo "✅ v0.7.0 认证功能演示完成！"
