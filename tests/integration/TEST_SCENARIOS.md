# 复杂测试场景说明

本文档描述了为Shode HTTP服务器、缓存和数据库功能创建的复杂测试场景。

## 测试文件

### 1. `http_api_complex_test.go` - HTTP API复杂场景测试

#### TestRESTfulAPIWithCache
- **场景**: RESTful API与缓存结合
- **测试内容**:
  - GET请求返回数据（带缓存）
  - POST请求创建资源
  - HTTP方法路由匹配
- **验证点**: 状态码、响应内容、路由注册

#### TestDatabaseWithCache
- **场景**: 数据库查询与缓存结合
- **测试内容**:
  - 创建表并插入数据
  - 查询数据并缓存结果
  - 从缓存中检索数据
- **验证点**: 数据查询、缓存存储和检索

#### TestHTTPRequestContext
- **场景**: HTTP请求上下文访问
- **测试内容**:
  - 获取HTTP方法
  - 获取请求路径
  - 获取查询参数
- **验证点**: 请求上下文函数正常工作

#### TestCompleteUserWorkflow
- **场景**: 完整的用户管理工作流
- **测试内容**:
  1. 连接数据库
  2. 创建用户表
  3. 插入多个用户
  4. 查询所有用户并缓存
  5. 查询单个用户
  6. 更新用户
  7. 使缓存失效
- **验证点**: 完整的CRUD操作流程、缓存管理

#### TestCacheTTLAndExpiration
- **场景**: 缓存TTL和过期机制
- **测试内容**:
  - 设置带TTL的缓存
  - 检查缓存是否存在
  - 获取剩余TTL
  - 等待过期
- **验证点**: TTL机制、过期清理

#### TestHTTPMethodsComprehensive
- **场景**: 所有HTTP方法支持
- **测试内容**:
  - GET, POST, PUT, DELETE, PATCH方法
  - 不同路径的路由注册
  - 方法匹配验证
- **验证点**: 所有HTTP方法正常工作

#### TestCachePatternMatching
- **场景**: 缓存键模式匹配
- **测试内容**:
  - 设置多个带模式的缓存键
  - 使用通配符模式查找键
  - 前缀、后缀、包含匹配
- **验证点**: 模式匹配功能

#### TestDatabaseTransactionSimulation
- **场景**: 数据库事务模拟（账户转账）
- **测试内容**:
  - 创建账户表
  - 初始化账户余额
  - 执行转账操作（扣除和增加）
  - 验证余额变化
- **验证点**: 多步骤数据库操作、数据一致性

### 2. `real_world_scenarios_test.go` - 真实世界场景测试

#### TestECommerceAPI
- **场景**: 电商API（产品、订单管理）
- **测试内容**:
  - 产品表管理
  - 订单创建
  - 产品查询和缓存
  - 订单查询
- **验证点**: 电商业务逻辑、数据关联

#### TestBlogAPI
- **场景**: 博客API（文章、评论管理）
- **测试内容**:
  - 文章创建
  - 评论添加
  - JOIN查询（文章+评论数）
  - 浏览量统计
  - 数据缓存
- **验证点**: 复杂查询、数据聚合、缓存策略

#### TestAPIRateLimitingWithCache
- **场景**: 使用缓存的API限流
- **测试内容**:
  - 使用缓存跟踪用户请求次数
  - 实现简单的限流逻辑
  - 验证限流生效
- **验证点**: 缓存用于业务逻辑、限流机制

#### TestSessionManagementWithCache
- **场景**: 使用缓存的会话管理
- **测试内容**:
  - 创建多个会话
  - 存储会话数据到缓存
  - 检索会话
  - 模式匹配查找所有会话
  - 会话失效
- **验证点**: 会话管理、缓存TTL、批量操作

#### TestDataAggregationWithCache
- **场景**: 数据聚合与缓存
- **测试内容**:
  - 创建销售数据表
  - 插入销售记录
  - 按产品聚合销售额
  - 缓存聚合结果
- **验证点**: SQL聚合函数、缓存复杂数据

#### TestHTTPAPIWithDatabaseAndCache
- **场景**: 完整的HTTP API（数据库+缓存）
- **测试内容**:
  - 启动HTTP服务器
  - 连接数据库
  - 注册GET和POST路由
  - 测试HTTP请求
- **验证点**: HTTP、数据库、缓存的集成

## 测试覆盖的功能

### HTTP服务器
- ✅ 多种HTTP方法（GET, POST, PUT, DELETE, PATCH）
- ✅ 路由注册和匹配
- ✅ 请求上下文访问
- ✅ 响应设置

### 缓存系统
- ✅ 基础操作（Set, Get, Delete, Clear）
- ✅ TTL和过期机制
- ✅ 模式匹配
- ✅ 批量操作
- ✅ 业务场景应用（限流、会话管理）

### 数据库操作
- ✅ 多数据库支持（SQLite测试）
- ✅ 表创建和管理
- ✅ CRUD操作
- ✅ 参数化查询
- ✅ 聚合查询
- ✅ JOIN查询
- ✅ 事务模拟

### 集成场景
- ✅ HTTP + 数据库
- ✅ HTTP + 缓存
- ✅ 数据库 + 缓存
- ✅ HTTP + 数据库 + 缓存

## 运行测试

```bash
# 运行所有复杂场景测试
go test ./tests/integration -v

# 运行特定场景
go test ./tests/integration -v -run TestECommerceAPI
go test ./tests/integration -v -run TestCompleteUserWorkflow
go test ./tests/integration -v -run TestDatabaseTransactionSimulation

# 运行真实世界场景
go test ./tests/integration -v -run "TestECommerceAPI|TestBlogAPI|TestSessionManagementWithCache"
```

## 注意事项

1. **数据库驱动**: 测试使用SQLite内存数据库，无需外部数据库
2. **端口冲突**: 不同测试使用不同端口（9189, 9190, 9191, 9192）避免冲突
3. **处理器执行**: 部分HTTP处理器执行功能需要完整实现，当前返回占位符响应
4. **缓存清理**: 缓存清理每1分钟运行一次，测试中可能需要等待

## 测试统计

- **总测试数**: 14个复杂场景测试
- **覆盖功能**: HTTP服务器、缓存、数据库、集成场景
- **测试类型**: 单元测试、集成测试、端到端测试
