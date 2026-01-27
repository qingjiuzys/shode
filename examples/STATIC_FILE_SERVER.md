# Static File Server - 综合指南

## 功能概述

Shode 静态文件服务器提供了完整的静态文件服务能力，支持多种高级功能。

## 快速开始

### 基础用法

```bash
#!/usr/bin/env shode

# 启动 HTTP 服务器
StartHTTPServer "8080"

# 注册静态文件路由
RegisterStaticRoute "/" "./public"

# 保持服务器运行
for i in $(seq 1 100000); do sleep 1; done
```

### 高级用法

```bash
# 完整配置语法
RegisterStaticRouteAdvanced [path] [directory] [indexFiles] [directoryBrowse] [cacheControl] [enableGzip] [spaFallback]
```

**参数说明：**
- `path`: URL 路径前缀（如："/", "/assets"）
- `directory`: 文件系统目录路径
- `indexFiles`: 索引文件列表（逗号分隔，如："index.html,index.htm"）
- `directoryBrowse`: 启用目录浏览（"true" 或 "false"）
- `cacheControl`: 缓存控制头（如："max-age=3600"）
- `enableGzip`: 启用 gzip 压缩（"true" 或 "false"）
- `spaFallback`: SPA 回退文件（如："index.html"）

## 使用示例

### 示例 1: 基础静态文件服务

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"
RegisterStaticRoute "/" "./public"
```

### 示例 2: 多个静态目录

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 主站文件
RegisterStaticRoute "/" "./public"

# 静态资源（CSS、JS、图片）
RegisterStaticRoute "/assets" "./static/assets"

# 文档
RegisterStaticRoute "/docs" "./documentation"
```

### 示例 3: 启用目录浏览

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 启用目录浏览，显示文件列表
RegisterStaticRouteAdvanced "/" "./public" "" "true" "" "" ""
```

### 示例 4: 配置缓存控制

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 设置 1 小时缓存
RegisterStaticRouteAdvanced "/" "./public" "" "false" "max-age=3600" "" ""
```

### 示例 5: SPA（单页应用）支持

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 所有路由回退到 index.html（适用于 React、Vue 等 SPA）
RegisterStaticRouteAdvanced "/" "./spa-build" "" "false" "" "" "index.html"
```

### 示例 6: 完整配置

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 完整功能：自定义索引文件、目录浏览、缓存控制
RegisterStaticRouteAdvanced "/" "./public" \
    "index.html,index.htm,default.html" \
    "true" \
    "max-age=3600, public" \
    "true" \
    ""

# API 端点
function handleAPI() {
    SetHTTPResponse 200 "API Status: OK"
}
RegisterHTTPRoute "GET" "/api/status" "function" "handleAPI"
```

## 支持的文件类型

静态文件服务器自动检测并设置正确的 MIME 类型：

| 文件扩展名 | MIME 类型 |
|-----------|----------|
| .html, .htm | text/html; charset=utf-8 |
| .css | text/css; charset=utf-8 |
| .js | application/javascript; charset=utf-8 |
| .json | application/json; charset=utf-8 |
| .xml | application/xml; charset=utf-8 |
| .png | image/png |
| .jpg, .jpeg | image/jpeg |
| .gif | image/gif |
| .svg | image/svg+xml |
| .ico | image/x-icon |
| .woff, .woff2 | font/woff2 |
| .ttf | font/ttf |
| .pdf | application/pdf |
| .zip | application/zip |
| .txt | text/plain; charset=utf-8 |
| .md | text/markdown; charset=utf-8 |

## 安全特性

### 1. 路径遍历防护
```bash
# 攻击尝试
curl http://localhost:8080/../../../etc/passwd

# 结果：403 Forbidden
```

### 2. 文件验证
- 检查目录存在性
- 验证可读性权限
- 相对路径自动转换为绝对路径

## 错误处理

### 404 - 文件未找到
当请求的文件不存在时，服务器返回 404 状态码。

### 403 - 禁止访问
当检测到路径遍历攻击时，返回 403 状态码。

### 500 - 服务器错误
当服务器内部错误时，返回 500 状态码。

## 性能优化

### 1. 缓存控制
```bash
# 设置浏览器缓存
RegisterStaticRouteAdvanced "/" "./public" "" "false" "max-age=3600" "" ""
```

### 2. Gzip 压缩
```bash
# 启用 gzip 压缩以减少传输大小（约 50% 压缩率）
RegisterStaticRouteAdvanced "/" "./public" "" "false" "" "true" ""
```

## 与 API 端点集成

静态文件服务器可以与 API 端点无缝集成：

```bash
#!/usr/bin/env shode
StartHTTPServer "8080"

# 静态文件服务
RegisterStaticRoute "/" "./public"

# API 端点
function getAPI() {
    SetHTTPResponse 200 '{"status":"ok","data":[1,2,3]}'
}
RegisterHTTPRoute "GET" "/api/data" "function" "getAPI"

function postAPI() {
    SetHTTPResponse 201 '{"message":"Created"}'
}
RegisterHTTPRoute "POST" "/api/data" "function" "postAPI"
```

## 故障排除

### 问题：目录浏览不工作
**解决：** 确保将 `directoryBrowse` 参数设置为 `"true"`

### 问题：找不到文件
**解决：** 检查目录路径是否正确（相对于脚本运行目录或使用绝对路径）

### 问题：MIME 类型不正确
**解决：** 当前支持常见文件类型。如需添加新类型，请修改 `getContentType()` 函数

## 最佳实践

1. **使用绝对路径**：避免相对路径的歧义
2. **配置合适的缓存**：静态资源可以设置较长缓存时间
3. **禁用目录浏览**：生产环境建议禁用目录浏览
4. **使用 SPA fallback**：对于单页应用，设置回退到 index.html

## 完整示例

参见 `examples/static_file_server.sh` 和 `examples/static_advanced.sh`。

## 🌟 真实项目示例

我们提供了多个真实场景的完整项目示例，展示如何在不同情况下使用静态文件服务器：

### 📄 个人网站/博客
**文件**: `examples/projects/personal-website.sh`

**特点**:
- 静态 HTML 页面服务
- 博客文章列表
- 统计信息 API
- 简洁的响应式设计

**运行**:
```bash
./shode run examples/projects/personal-website.sh
# 访问 http://localhost:3000
```

**包含内容**:
- 首页 `/`
- 博客 `/blog/`
- 关于页面 `/about.html`
- 统计 API `/api/stats`

---

### 📚 API 文档服务器
**文件**: `examples/projects/api-docs-server.sh`

**特点**:
- 目录浏览功能（便于文档导航）
- 多文档版本支持
- 静态资源缓存优化
- 搜索 API 端点

**运行**:
```bash
./shode run examples/projects/api-docs-server.sh
# 访问 http://localhost:8080/docs
```

**包含内容**:
- 文档浏览器 `/docs`
- 静态资源 `/assets`
- 搜索 API `/api/search`

---

### 🚀 全栈应用
**文件**: `examples/projects/fullstack-app.sh`

**特点**:
- SPA（单页应用）支持
- 完整的 RESTful API
- CRUD 操作
- 健康检查端点
- JSON 数据响应

**运行**:
```bash
./shode run examples/projects/fullstack-app.sh
# 访问 http://localhost:4000
```

**API 端点**:
- `GET /api/users` - 获取所有用户
- `GET /api/users/1` - 获取单个用户
- `POST /api/users` - 创建新用户
- `GET /api/health` - 健康检查

---

### 📦 文件下载服务器
**文件**: `examples/projects/file-server.sh`

**特点**:
- 下载优化（长缓存时间）
- 发布说明目录浏览
- 最新版本 API
- 文件列表 API

**运行**:
```bash
./shode run examples/projects/file-server.sh
# 访问 http://localhost:5000
```

**包含内容**:
- 文件下载 `/downloads`
- 发布说明 `/releases`（可浏览）
- 文件列表 API `/api/files`
- 最新版本 API `/api/latest`

---

### 查看所有项目示例

更多项目示例和详细信息，请参阅：
**[项目示例文档](examples/projects/README.md)**

## API 参考

### 函数列表

- `StartHTTPServer(port)` - 启动 HTTP 服务器
- `RegisterStaticRoute(path, directory)` - 注册基础静态路由
- `RegisterStaticRouteAdvanced(path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, spaFallback)` - 注册高级静态路由
- `RegisterHTTPRoute(method, path, type, handler)` - 注册 HTTP 路由
- `SetHTTPResponse(status, body)` - 设置 HTTP 响应

## 更新日志

### v0.5.0 (当前版本)
- ✅ 基础静态文件服务
- ✅ 目录浏览功能
- ✅ 缓存控制支持
- ✅ **Gzip 压缩**（约 50% 压缩率）
- ✅ 路径遍历防护
- ✅ 自动 MIME 类型检测
- ✅ 多路由支持
- ✅ SPA 回退支持
- ✅ 与 API 端点集成
- ✅ **完整项目示例**（个人网站、API 文档、全栈应用、文件服务器）

### 未来计划
- ⏳ Range 请求支持（断点续传）
- ⏳ 自定义错误页面
- ⏳ 请求日志记录
- ⏳ HTTP/2 支持
- ⏳ WebSocket 支持
