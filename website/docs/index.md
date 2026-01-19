# Shode - 极简 HTTP 框架

> **30 秒上手 · 5 分钟精通 · 快速原型开发**

---

## 🚀 30 秒上手

### Hello World

```sh
#!/bin/sh
StartHTTPServer 8080
RegisterRouteWithResponse "/" "Hello World"
```

运行：
```bash
./shode run api.sh
```

测试：
```bash
curl http://localhost:8080/
```

**输出:** `Hello World`

---

## 📝 5 分钟精通

### 1. GET 请求

```sh
function getItems() {
    SetHTTPResponse 200 '{"items": ["apple", "banana"]}'
}

RegisterHTTPRoute "GET" "/items" "function" "getItems"
```

### 2. POST 请求

```sh
function addItem() {
    name = GetHTTPQuery "name"
    SetHTTPResponse 201 '{"added": "' + name + '"}'
}

RegisterHTTPRoute "POST" "/items" "function" "addItem"
```

### 3. 数据库操作

```sh
StartHTTPServer 8080
ConnectDB "sqlite" "app.db"

ExecDB "CREATE TABLE users (id INTEGER, name TEXT)"
ExecDB "INSERT INTO users (name) VALUES (?)" "Alice"

function getUsers() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 "$result"
}

RegisterHTTPRoute "GET" "/users" "function" "getUsers"
```

### 4. 完整 CRUD

```sh
# 创建
function create() {
    name = GetHTTPQuery "name"
    ExecDB "INSERT INTO items (name) VALUES (?)" name
    SetHTTPResponse 201 '{"success": true}'
}

# 读取
function getAll() {
    QueryDB "SELECT * FROM items"
    result = GetQueryResult
    SetHTTPResponse 200 "$result"
}

# 更新
function update() {
    id = GetHTTPQuery "id"
    name = GetHTTPQuery "name"
    ExecDB "UPDATE items SET name = ? WHERE id = ?" name id
    SetHTTPResponse 200 '{"success": true}'
}

# 删除
function delete() {
    id = GetHTTPQuery "id"
    ExecDB "DELETE FROM items WHERE id = ?" id
    SetHTTPResponse 204 ""
}

RegisterHTTPRoute "POST" "/items" "function" "create"
RegisterHTTPRoute "GET" "/items" "function" "getAll"
RegisterHTTPRoute "PUT" "/items" "function" "update"
RegisterHTTPRoute "DELETE" "/items" "function" "delete"
```

---

## 💾 缓存优化

```sh
function getData() {
    # 检查缓存
    cached = GetCache "data:all"

    if cached != "" {
        SetHTTPResponse 200 "$cached"
        return
    }

    # 查询数据库
    QueryDB "SELECT * FROM data"
    result = GetQueryResult

    # 存入缓存（5 分钟）
    SetCache "data:all" result 300
    SetHTTPResponse 200 "$result"
}
```

---

## 🔐 认证系统

```sh
function login() {
    username = GetHTTPQuery "username"
    password = GetHTTPQuery "password"

    # 密码哈希
    passwordHash = SHA256Hash password

    # 验证（示例）
    QueryRowDB "SELECT * FROM users WHERE username = ?" username
    result = GetQueryResult

    if Contains result passwordHash {
        # 生成会话令牌
        token = SHA256Hash username + "salt"
        SetCache "session:" + token username 3600
        SetHTTPResponse 200 '{"token":"' + token '"}'
    } else {
        SetHTTPResponse 401 '{"error":"Invalid credentials"}'
    }
}

function protected() {
    token = GetHTTPHeader "Authorization"
    username = GetCache "session:" + token

    if username == "" {
        SetHTTPResponse 401 '{"error":"Unauthorized"}'
        return
    }

    # 验证成功，继续处理
    SetHTTPResponse 200 '{"data":"protected"}'
}
```

---

## 🎯 框架对比

### Shode vs 其他框架

| 特性 | Shode | Express.js | Flask | Spring Boot |
|--------|--------|-----------|-------|-------------|
| **代码量** | 5 行 | 20 行 | 15 行 | 50+ 行 |
| **学习曲线** | ⭐ 极简 | ⭐⭐ 中等 | ⭐⭐ 中等 | ⭐⭐⭐ 复杂 |
| **启动时间** | &lt;1 秒 | ~3 秒 | ~2 秒 | ~10 秒 |
| **数据库** | ✅ 内置 | ❌ 需额外库 | ❌ 需要 ORM | ✅ 内置 |
| **缓存** | ✅ 内置 | ❌ 需额外库 | ❌ 需额外库 | ❌ 需额外库 |
| **配置文件** | ✅ 无需 | ✅ package.json | ✅ requirements.txt | ⭐⭐ 多个文件 |
| **部署复杂度** | ✅ 单文件 | ⭐⭐ 需要打包 | ⭐ 需要 venv | ⭐⭐⭐ 容器/配置 |

### 示例对比 - Hello World

**Shode (3 行):**
```sh
StartHTTPServer 8080
RegisterRouteWithResponse "/" "Hello"
```

**Express.js (8 行):**
```javascript
const express = require('express');
const app = express();
app.get('/', (req, res) => res.send('Hello'));
app.listen(8080);
```

**Flask (7 行):**
```python
from flask import Flask
app = Flask(__name__)
@app.route('/')
def hello():
    return 'Hello'
app.run(port=8080)
```

**Spring Boot (20+ 行):**
```java
@SpringBootApplication
@RestController
public class App {
    @GetMapping("/")
    public String hello() {
        return "Hello";
    }
    public static void main(String[] args) {
        SpringApplication.run(App.class, args);
    }
}
// + pom.xml, application.yml
```

---

## 🌟 核心优势

### 1. 极致简单
- ✅ 单脚本文件即可运行
- ✅ 无需项目结构
- ✅ 无需构建步骤

### 2. 开箱即用
- ✅ HTTP 服务器内置
- ✅ SQLite/MySQL/PostgreSQL 开箱即用
- ✅ 内存缓存系统
- ✅ 安全检查（防注入）

### 3. 快速开发
- ✅ 从 idea 到运行 &lt;1 分钟
- ✅ 迭代速度极快
- ✅ 适合快速原型

---

## 📚 常用命令

### 服务器
```sh
StartHTTPServer 8080          # 启动
StopHTTPServer                # 停止
```

### 数据库
```sh
ConnectDB "sqlite" "app.db"           # 连接
QueryDB "SELECT * FROM users"          # 查询
ExecDB "INSERT INTO users..."            # 插入
GetQueryResult                         # 获取结果
```

### 缓存
```sh
SetCache "key" "value" 300     # 设置（5分钟）
GetCache "key"                   # 获取
DeleteCache "key"                 # 删除
```

### HTTP
```sh
GetHTTPMethod              # 获取方法
GetHTTPPath                # 获取路径
GetHTTPQuery "name"         # 获取参数
SetHTTPResponse 200 data    # 设置响应
SetHTTPHeader "Content-Type" "application/json"
```

---

## 🎯 何时选择 Shode

### ✅ 推荐使用

1. **快速原型** - 5 分钟构建可演示的 API
2. **脚本自动化** - 将 Shell 脚本升级为 Web API
3. **简单 CRUD 应用** - 数据管理后台
4. **学习 HTTP 原理** - 理解 RESTful 设计
5. **资源受限环境** - 容器化、IoT 设备

### ❌ 不推荐使用

1. **大型企业应用** → Spring Boot
2. **复杂前端应用** → Express.js
3. **数据科学项目** → Flask/FastAPI

---

## 💡 提示

- 🎯 从最简单的例子开始
- 🎯 使用内置函数，不要重复造轮子
- 🎯 先实现功能，再优化性能
- 🎯 利用缓存减少数据库查询
- 🎯 保持代码简洁，单文件即可运行

---

**开始时间**: 现在
**完成时间**: 5 分钟后
**下一步**: 创建你的第一个 API！
