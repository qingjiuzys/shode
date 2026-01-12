# Spring 功能示例

## 简介

这个示例展示了 Shode 的 Spring 化功能，包括 IoC 容器、配置管理、Web 层、事务管理等企业级特性。

## 功能特性

- **IoC 容器**: Bean 管理和依赖注入
- **配置管理**: 多源配置、类型安全访问
- **Web 层**: HTTP 服务器、路由注册
- **数据访问**: 数据库连接、查询操作
- **事务管理**: 事务支持（框架已就绪）

## 代码概览

查看完整代码：`examples/spring_complete_example.sh`

### 1. 配置管理

```shode
# 创建配置文件
configContent = '{"server":{"port":9188},"database":{"url":"sqlite:app.db"}}'
WriteFile "application.json" configContent

# 加载配置
LoadConfig "application.json"

# 读取配置
port = GetConfigString "server.port" "9188"
dbUrl = GetConfigString "database.url" "sqlite:app.db"
```

### 2. IoC 容器

```shode
# 注册 Bean（注意：需要函数引用支持）
# RegisterBean "userService" "singleton" createUserService

# 获取 Bean
# userService = GetBean "userService"

# 检查 Bean 是否存在
# exists = ContainsBean "userService"
```

### 3. Web 层

```shode
# 启动 HTTP 服务器
StartHTTPServer port

# 定义控制器函数
function UserController() {
    QueryDB "SELECT * FROM users"
    result = GetQueryResult
    SetHTTPResponse 200 result
}

# 注册路由
RegisterHTTPRoute "GET" "/api/users" "function" "UserController"
```

### 4. 数据访问

```shode
# 连接数据库
ConnectDB "sqlite" dbUrl

# 创建表
ExecDB "CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)"

# 查询数据
QueryDB "SELECT * FROM users"
result = GetQueryResult
```

## 运行方式

```bash
shode run examples/spring_complete_example.sh
```

## 其他 Spring 示例

### IoC 容器示例

```bash
shode run examples/spring_ioc_example.sh
```

### 配置管理示例

```bash
shode run examples/spring_config_example.sh
```

### Web 应用示例

```bash
shode run examples/spring_web_example.sh
```

### 事务管理示例

```bash
shode run examples/spring_transaction_example.sh
```

## Spring 功能对比

| Spring 功能 | Shode 实现 | 状态 |
|-----------|----------|------|
| IoC 容器 | ✅ | 完成 |
| 依赖注入 | ✅ | 完成 |
| 配置管理 | ✅ | 完成 |
| 注解系统 | ✅ | 基础完成 |
| 中间件 | ✅ | 完成 |
| 控制器 | ✅ | 完成 |
| 事务管理 | ✅ | 框架完成 |
| Repository | ✅ | 完成 |
| AOP | ✅ | 完成 |

## 使用场景

- **企业级应用**: 构建完整的 Web 应用
- **微服务**: 快速开发微服务接口
- **API 服务**: 提供 RESTful API
- **数据驱动应用**: 数据库驱动的应用

## 相关文档

- [用户指南 - Spring 化功能](../../guides/user-guide.md#spring-化功能)
- [开发文档 - Spring 化改造](../../../docs/development/SPRING_MIGRATION.md)
