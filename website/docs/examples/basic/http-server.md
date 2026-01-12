# HTTP 服务器示例

## 简介

这个示例展示如何使用 Shode 创建一个简单的 HTTP 服务器，并注册基本路由。

## 代码

```shode
#!/usr/bin/env shode

# Simple HTTP Server Example
# This script demonstrates how to create an HTTP server using Shode

# Start HTTP server on port 9188
Println "Starting HTTP server on port 9188..."
StartHTTPServer "9188"

# Wait a moment for server to start
sleep 1

# Register a route that returns "hello world"
Println "Registering route /..."
RegisterRouteWithResponse "/" "hello world"

Println "HTTP server is running on http://localhost:9188"
Println "Visit http://localhost:9188 in your browser or use: curl http://localhost:9188"
Println ""
Println "Server is ready to accept requests"
```

## 运行方式

```bash
shode run examples/http_server.sh
```

## 测试

在另一个终端中测试：

```bash
# 访问根路径
curl http://localhost:9188

# 预期输出: hello world
```

## 预期输出

```
Starting HTTP server on port 9188...
Registering route /...
HTTP server is running on http://localhost:9188
Visit http://localhost:9188 in your browser or use: curl http://localhost:9188

Server is ready to accept requests
```

## 扩展

### 注册多个路由

```shode
# 注册不同的路由
RegisterRouteWithResponse "/" "Home"
RegisterRouteWithResponse "/api" "API Endpoint"
RegisterRouteWithResponse "/health" "OK"
```

### 使用函数处理器

```shode
function handleRequest() {
    SetHTTPResponse 200 "Hello from function handler"
}

RegisterHTTPRoute "GET" "/api" "function" "handleRequest"
```

## 相关文档

- [用户指南 - HTTP 服务器](../../guides/user-guide.md#6-http-服务器)
- [API 参考 - HTTP 函数](../../api/stdlib.md#http-服务器函数)
