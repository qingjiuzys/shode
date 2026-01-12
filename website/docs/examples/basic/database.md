# 数据库操作示例

## 简介

这个示例展示如何使用 Shode 连接数据库（MySQL、PostgreSQL、SQLite），执行查询和更新操作。

## 代码

```shode
#!/usr/bin/env shode

# Database Example
# Demonstrates database operations with MySQL, PostgreSQL, and SQLite

Println "Database Example"

# Connect to SQLite (for testing)
ConnectDB "sqlite" ":memory:"
Println "Connected to SQLite"

# Create table
ExecDB "CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)"
Println "Table created"

# Insert data
ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "Alice" "alice@example.com"
ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "Bob" "bob@example.com"
Println "Data inserted"

# Query data
QueryDB "SELECT * FROM users"
result = GetQueryResult
Println "Query result: " + result

# Query single row
QueryRowDB "SELECT * FROM users WHERE id = ?" "1"
row = GetQueryResult
Println "Single row: " + row

# Update data
ExecDB "UPDATE users SET email = ? WHERE id = ?" "alice.new@example.com" "1"
Println "Data updated"

# Close connection
CloseDB
Println "Connection closed"
```

## 运行方式

```bash
shode run examples/database_example.sh
```

## 支持的数据库

### SQLite

```shode
# 内存数据库
ConnectDB "sqlite" ":memory:"

# 文件数据库
ConnectDB "sqlite" "app.db"
```

### MySQL

```shode
ConnectDB "mysql" "user:password@tcp(localhost:3306)/dbname"
```

### PostgreSQL

```shode
ConnectDB "postgres" "postgres://user:password@localhost/dbname?sslmode=disable"
```

## 功能说明

### 查询操作

- `QueryDB` - 查询多行数据
- `QueryRowDB` - 查询单行数据
- `GetQueryResult` - 获取查询结果（JSON 格式）

### 更新操作

- `ExecDB` - 执行 INSERT、UPDATE、DELETE 等操作
- 支持参数化查询，防止 SQL 注入

### 连接管理

- `ConnectDB` - 连接数据库
- `IsDBConnected` - 检查连接状态
- `CloseDB` - 关闭连接

## 使用场景

- **数据持久化**: 存储应用数据
- **API 后端**: 为 HTTP API 提供数据存储
- **数据分析**: 查询和分析数据
- **配置存储**: 存储应用配置

## 相关文档

- [用户指南 - 数据库操作](../../guides/user-guide.md#8-数据库操作)
- [API 参考 - 数据库函数](../../api/stdlib.md#数据库函数)
