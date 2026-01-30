# Shode 示例项目

本目录包含完整的 Shode 示例项目，展示各种功能和最佳实践。

## 📁 项目列表

### 1. WebSocket 聊天室 (websocket-chat-complete.sh)

完整的实时聊天应用，展示 WebSocket 功能。

**功能：**
- ✅ 实时消息收发
- ✅ 房间管理
- ✅ 用户统计
- ✅ 消息广播
- ✅ Web 界面

**运行：**
```bash
./examples/projects/websocket-chat-complete.sh
```

**访问：**
- WebSocket: `ws://localhost:8098/ws`
- Web 界面: `http://localhost:8098/`
- API: `http://localhost:8098/api/`

### 2. REST API with Cache (rest-api-with-cache.sh)

带缓存优化的 RESTful API 示例。

**功能：**
- ✅ CRUD 操作
- ✅ SQLite 数据库
- ✅ 内存缓存
- ✅ 缓存失效策略

**运行：**
```bash
./examples/projects/rest-api-with-cache.sh
```

**API 端点：**
```bash
# 获取用户列表
curl http://localhost:8099/api/users

# 创建用户
curl 'http://localhost:8099/api/users?name=Alice&email=alice@example.com' -X POST

# 获取单个用户
curl 'http://localhost:8099/api/user?id=1'

# 更新用户
curl 'http://localhost:8099/api/user?id=1&name=Alice+Smith' -X PUT

# 删除用户
curl 'http://localhost:8099/api/user?id=1' -X DELETE
```

### 3. 静态文件服务器

提供静态文件服务。

**运行：**
```bash
./examples/projects/personal-website.sh
```

### 4. API 文档服务器

API 文档浏览和搜索。

**运行：**
```bash
./examples/projects/api-docs-server.sh
```

### 5. 全栈应用

SPA + RESTful API 的完整应用。

**运行：**
```bash
./examples/projects/fullstack-app.sh
```

---

## 🚀 快速开始

### 1. 选择项目

```bash
cd examples/projects
ls -la
```

### 2. 运行项目

```bash
# 直接运行
./websocket-chat-complete.sh

# 或使用 shode 运行
shode run websocket-chat-complete.sh
```

### 3. 访问应用

打开浏览器访问对应的 URL。

---

## 📚 学习路径

### 初学者

1. **personal-website.sh** - 最简单，静态文件服务
2. **api-docs-server.sh** - 添加 API 端点
3. **websocket-chat-complete.sh** - WebSocket 基础

### 中级

1. **rest-api-with-cache.sh** - 数据库 + 缓存
2. **fullstack-app.sh** - 前后端集成
3. **file-server.sh** - 文件上传下载

### 高级

1. **error-pages-demo.sh** - 自定义错误页面
2. **template-demo.sh** - 模板引擎
3. **websocket-rooms.sh** - 高级 WebSocket 功能

---

## 🛠️ 项目结构

```
examples/projects/
├── public/                      # 静态资源
│   └── index.html              # 聊天室前端
├── websocket-chat-complete.sh  # WebSocket 聊天室
├── rest-api-with-cache.sh      # REST API 示例
├── personal-website.sh         # 个人网站
├── api-docs-server.sh          # API 文档服务器
├── fullstack-app.sh            # 全栈应用
├── file-server.sh              # 文件服务器
├── error-pages-demo.sh         # 错误页面演示
└── template-demo.sh            # 模板演示
```

---

## 💡 最佳实践

这些示例展示了以下最佳实践：

### 1. 错误处理

```bash
function handleRequest() {
    # 验证输入
    if IsEmpty $input; then
        SetHTTPResponse 400 '{"error":"Invalid input"}'
        return
    fi
    
    # 处理请求
    result, err := ProcessRequest $input
    if $err; then
        SetHTTPResponse 500 '{"error":"Internal error"}'
        return
    fi
    
    SetHTTPResponse 200 $result
}
```

### 2. 缓存策略

```bash
# 先检查缓存
cached, exists := GetCache "key"
if $exists; then
    SetHTTPResponse 200 $cached
    return
fi

# 查询数据
data := QueryDB "SELECT * FROM table"

# 存入缓存
SetCache "key" $data 3600
```

### 3. 并发安全

```bash
# 使用锁保护共享状态
sl.httpMu.Lock()
defer sl.httpMu.Unlock()
```

---

## 📖 相关文档

- [API 参考](../../docs/API_REFERENCE.md)
- [最佳实践](../../docs/BEST_PRACTICES.md)
- [WebSocket 指南](../../docs/WEBSOCKET_GUIDE.md)
- [编码规范](../../docs/CODING_STANDARDS.md)

---

## 🤝 贡献

欢迎提交更多示例项目！

提交前请确保：
- ✅ 代码遵循编码规范
- ✅ 有完整的注释
- ✅ 包含使用说明
- ✅ 提供示例输出

---

**Happy Coding with Shode!** 🚀
