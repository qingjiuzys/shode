# Shode - Shell 脚本运行时

> **Shell 脚本 → Web 服务：30 秒上手，类似 Node.js 的执行平台**

---

## 🎯 核心定位

**Shode** 是一个 **Shell 脚本运行时平台**，类似 Node.js 是 JavaScript 运行时。

**核心价值**：将你熟悉的 Shell 脚本，在 30 秒内升级为完整的 HTTP Web 服务。

---

## 🆚 与 Node.js 对标

| 维度 | Shode | Node.js |
|------|--------|---------|
| **运行语言** | Shell 脚本 | JavaScript |
| **主要用途** | Shell 脚本 Web 化 | JavaScript 开发运行时 |
| **学习成本** | ⭐（0 分钟，你会 Shell 即可） | ⭐⭐（需要学习 JavaScript） |
| **HTTP 服务器** | ✅ 内置，1 行启动 | ✅ 需要框架 |
| **数据库** | ✅ 内置（SQLite/MySQL/PG） | ❌ 需要额外库 |
| **缓存** | ✅ 内置 | ❌ 需要 Redis/Memcached |
| **依赖管理** | ✅ 无需 npm | ⭐⭐ 需要 npm |
| **部署** | ✅ 单脚本文件 | ⭐⭐ 需要 build + 部署 |
| **适用场景** | 脚本自动化、快速原型 | 完整的前后端应用 |

---

## 💡 核心优势

### 1. 你已经会 Shell

```bash
# 你现有的 Shell 脚本，无需改语言
#!/bin/bash

# 检查服务器状态
if systemctl is-active nginx; then
    echo "Nginx is running"
fi

# 重启服务
systemctl restart mysql

# 备份数据库
mysqldump -u root -p database > backup.sql
```

**Shode 让你的 Shell 脚本立即变成 HTTP API**：
```sh
#!/bin/sh

StartHTTPServer 8080

function checkNginx() {
    status = IsHTTPServerRunning
    SetHTTPResponse 200 '{"nginx":"' + status '"}'
}

function restartMySQL() {
    ExecDB "RESTART TABLE users"
    SetHTTPResponse 200 '{"status":"restarted"}'
}

function backupDB() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/nginx" "function" "checkNginx"
RegisterHTTPRoute "POST" "/mysql/restart" "function" "restartMySQL"
RegisterHTTPRoute "GET" "/backup" "function" "backupDB"
```

**你的 Shell 脚本知识，直接复用！**

---

### 2. Shell 脚本 → Web API，0 学习成本

```bash
# 传统方式（需要 Web 开发）
# 1. 学习 Node.js/Express.js
# 2. 编写 API 服务器（30+ 行）
# 3. 配置路由、中间件
# 4. 安装依赖（npm install）
# 5. 构建、部署
# 总时间：2-3 小时

# Shode 方式（你已经会 Shell）
# 1. 添加 HTTP 服务器：1 行
# 2. 注册路由：1-2 行
# 3. 运行：shode run script.sh
# 总时间：30 秒
```

---

### 3. 内置所有 Web 基础设施

| 功能 | Shode | Node.js |
|------|--------|---------|
| HTTP 服务器 | ✅ 内置 | ❌ 需要 Express/Koa |
| 数据库 | ✅ 内置（3 种数据库） | ❌ 需要 mysql/pg/mongodb |
| 缓存 | ✅ 内置（内存缓存） | ❌ 需要 Redis |
| 认证 | ✅ 内置（会话、哈希） | ❌ 需要 passport/jwt |
| 日志 | ✅ 内置（Println） | ❌ 需要 winston/pino |

---

## 🚀 30 秒上手

### 场景 1：Shell 脚本 → HTTP API

**原脚本**：检查服务状态
```bash
#!/bin/bash

check_service() {
    if systemctl is-active nginx; then
        echo "Nginx running"
    else
        echo "Nginx stopped"
    fi
}

check_service
```

**Shode 升级**（30 秒完成）：
```sh
#!/bin/sh

StartHTTPServer 8080

function checkService() {
    if IsHTTPServerRunning == "true" {
        SetHTTPResponse 200 '{"status":"running"}'
    } else {
        SetHTTPResponse 200 '{"status":"stopped"}'
    }
}

RegisterHTTPRoute "GET" "/service/nginx" "function" "checkService"
```

**运行**：
```bash
./shode run api.sh
curl http://localhost:8080/service/nginx
```

**输出**：`{"status":"stopped"}`

---

### 场景 2：数据库查询 → HTTP API

**原脚本**：查询数据库
```bash
#!/bin/bash

query_users() {
    mysql -u root -p -e "SELECT * FROM users"
}

query_users
```

**Shode 升级**（30 秒完成）：
```sh
#!/bin/sh

StartHTTPServer 8080
ConnectDB "mysql" "root:password@tcp(3306)/database"

function getUsers() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/users" "function" "getUsers"
```

**运行**：
```bash
./shode run api.sh
curl http://localhost:8080/users
```

**输出**：完整的 JSON 数据

---

## 📖 典型应用场景

### 场景 1：运维脚本 Web 化

**传统方式**：手动 SSH，执行命令
```bash
# 需要登录到服务器，手动执行命令
ssh server "systemctl restart nginx"
```

**Shode 方式**：HTTP API 控制
```sh
StartHTTPServer 8080

function restartNginx() {
    Exec "systemctl restart nginx"
    SetHTTPResponse 200 '{"status":"restarted"}'
}

RegisterHTTPRoute "POST" "/nginx/restart" "function" "restartNginx"
```

**调用**：
```bash
curl -X POST http://localhost:8080/nginx/restart
```

---

### 场景 2：定时任务 → REST API

**传统方式**：Cron + Shell 脚本
```bash
# crontab -e
# 0 2 * * * /path/to/backup.sh
```

**Shode 方式**：定时任务 + HTTP API
```sh
StartHTTPServer 8080

function runBackup() {
    QueryDB "SELECT * FROM data"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/backup" "function" "runBackup"
```

**调用**：其他系统的定时任务
```bash
# 0 2 * * * curl http://localhost:8080/backup
```

---

### 场景 3：数据导出 → API

**传统方式**：生成文件，SCP 传输
```bash
# 生成文件
mysql -u root -p -e "SELECT * FROM users" > users.csv
# 传输
scp users.csv remote:/tmp/
```

**Shode 方式**：直接 HTTP 调用
```sh
StartHTTPServer 8080
ConnectDB "mysql" "root:password@tcp(3306)/database"

function exportUsers() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/export/users" "function" "exportUsers"
```

**调用**：
```bash
curl http://localhost:8080/export/users
```

---

## 🎯 Node.js 和 Shode 的互补关系

### Node.js 适合：
- ✅ 完整的前后端应用
- ✅ 复杂的 Web 应用
- ✅ 需要丰富框架生态的项目
- ✅ 团队协作的大型项目
- ✅ 需要类型安全的企业应用

### Shode 适合：
- ✅ 将现有 Shell 脚本升级为 Web 服务
- ✅ 运维脚本 Web 化（30 秒）
- ✅ 快速原型验证
- ✅ 内部工具 HTTP 化
- ✅ 自动化脚本的 REST API

**不是替代，而是互补**：
```bash
# 复杂应用：使用 Node.js
Node.js + Express + React = 完整的 Web 应用

# 脚本升级：使用 Shode
Shell 脚本 + Shode = 快速 Web API
```

---

## 💾 实际示例对比

### 示例 1：数据库查询 API

**Node.js 版本**（需要学习、配置、依赖）：
```javascript
// 需要：npm install express mysql2
// 需要：学习 JavaScript
// 需要：40+ 行代码
const express = require('express');
const mysql = require('mysql2');

const app = express();
const pool = mysql.createPool({
    host: 'localhost',
    user: 'root',
    password: 'password',
    database: 'mydb'
});

app.get('/users', (req, res) => {
    pool.query('SELECT * FROM users', (err, results) => {
        if (err) {
            res.status(500).json({error: err.message});
        } else {
            res.json({users: results});
        }
    });
});

app.listen(8080);
```

**Shode 版本**（你的 Shell 知识直接用）：
```sh
# 无需学习新语言
# 无需安装依赖
# 无需 40+ 行代码

StartHTTPServer 8080
ConnectDB "mysql" "root:password@tcp(3306)/mydb"

function getUsers() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/users" "function" "getUsers"
```

**对比**：
- 学习时间：2-3 小时（Node.js） vs 0 分钟（Shode）
- 代码量：40 行 vs 5 行
- 依赖管理：npm install vs 无需
- 配置复杂度：高 vs 零

---

### 示例 2：简单 Hello World

**Node.js 版本**：
```javascript
// 需要：npm init, npm install express
// 需要：学习 Express 路由
// 需要：配置 package.json

const express = require('express');
const app = express();
app.get('/', (req, res) => res.send('Hello'));
app.listen(8080);
```

**Shode 版本**：
```sh
# 你已经会的 Shell 命令风格
StartHTTPServer 8080
RegisterRouteWithResponse "/" "Hello"
```

---

## 🚀 开始使用

### 第一步：你的第一个 API

```sh
cat > my_api.sh << 'EOF'
#!/bin/sh

StartHTTPServer 8080

function hello() {
    SetHTTPResponse 200 "Hello from Shode!"
}

RegisterHTTPRoute "GET" "/" "function" "hello"
EOF

./shode run my_api.sh
curl http://localhost:8080/
```

### 第二步：添加数据库

```sh
cat > db_api.sh << 'EOF'
#!/bin/sh

StartHTTPServer 8080
ConnectDB "sqlite" "app.db"

ExecDB "CREATE TABLE users (id INTEGER, name)"

function getAll() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

function add() {
    name = GetHTTPQuery "name"
    ExecDB "INSERT INTO users (name) VALUES (?)" name
    SetHTTPResponse 201 '{"success":true}'
}

RegisterHTTPRoute "GET" "/users" "function" "getAll"
RegisterHTTPRoute "POST" "/users" "function" "add"
EOF

./shode run db_api.sh
curl http://localhost:8080/users
curl -X POST 'http://localhost:8080/users?name=Alice'
```

### 第三步：添加缓存

```sh
cat > cache_api.sh << 'EOF'
#!/bin/sh

StartHTTPServer 8080
ConnectDB "sqlite" "app.db"

function getCached() {
    data = GetCache "users"

    if data != "" {
        SetHTTPResponse 200 "$data"
        return
    }

    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetCache "users" result 300
    SetHTTPResponse 200 result
}

RegisterHTTPRoute "GET" "/users" "function" "getCached"
EOF

./shode run cache_api.sh
```

---

## 🎯 核心差异总结

### Node.js：
- ✅ JavaScript 运行时
- ✅ 适合全栈开发
- ✅ 生态系统成熟
- ✅ 适合复杂项目

### Shode：
- ✅ Shell 脚本运行时
- ✅ 将 Shell 脚本升级为 Web 服务
- ✅ 0 学习成本（你已经会 Shell）
- ✅ 开箱即用（HTTP/DB/Cache）
- ✅ 适合脚本自动化、运维、快速原型

---

## 💡 真正的价值主张

**不是让你学习 Shode Script 语法，而是让你用 Shell 脚本做 Web 服务**

```
你的知识：Shell 脚本
+
Shode 的能力：HTTP + DB + Cache
=
30 秒内可用的 Web API
```

---

## 🚀 快速开始

```bash
# 1. 安装
go install github.com/com_818cloud/shode@latest

# 2. 运行第一个 API
cat > api.sh << 'EOF'
StartHTTPServer 8080
function api() {
    SetHTTPResponse 200 '{"status":"ok"}'
}
RegisterHTTPRoute "GET" "/" "function" "api"
EOF

./shode run api.sh

# 3. 测试
curl http://localhost:8080/
```

---

## 📖 更多场景

- [运维脚本 Web 化](#运维脚本-web-化)
- [定时任务 REST API](#定时任务--rest-api)
- [数据导出 API](#数据导出--api)
- [监控系统](#监控系统)
- [自动化工具](#自动化工具)

---

## 🎯 何时选择 Shode

### ✅ 选择 Shode：
1. 你已经熟悉 Shell 脚本
2. 需要将现有脚本升级为 Web 服务
3. 想快速验证想法（30 秒）
4. 不想学习新语言（Node.js）
5. 内部工具和服务
6. 自动化脚本

### ❌ 选择 Node.js：
1. 开发新的 Web 应用
2. 需要复杂的前后端集成
3. 需要框架生态
4. 团队协作的大型项目

---

**开始时间**: 现在
**目标**: 30 秒内运行你的第一个 Shell 脚本 Web API
