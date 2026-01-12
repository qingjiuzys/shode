# Shode Examples

本目录包含各种Shode使用示例，展示HTTP服务器、缓存、数据库等功能的实际应用场景。

## 基础示例

### `http_server.sh`
简单的HTTP服务器示例，展示如何启动服务器和注册基本路由。

```bash
shode run examples/http_server.sh
```

### `cache_example.sh`
缓存系统基础示例，展示缓存的设置、获取、删除等操作。

```bash
shode run examples/cache_example.sh
```

### `database_example.sh`
数据库操作示例，展示如何连接数据库、执行查询和更新操作。

```bash
shode run examples/database_example.sh
```

## 复杂场景示例

### `ecommerce_api.sh` - 电商API
完整的电商API示例，包含：
- 产品管理（查询、缓存）
- 订单创建
- 缓存策略
- HTTP路由注册

**运行方式：**
```bash
shode run examples/ecommerce_api.sh
```

**测试API：**
```bash
# 获取产品列表
curl http://localhost:9188/api/products

# 获取单个产品
curl http://localhost:9188/api/product?id=1

# 创建订单
curl -X POST http://localhost:9188/api/orders

# 获取订单
curl http://localhost:9188/api/orders?user_id=1
```

### `blog_api.sh` - 博客API
博客系统API示例，包含：
- 文章管理
- 评论系统
- 浏览量统计
- JOIN查询
- 缓存策略

**运行方式：**
```bash
shode run examples/blog_api.sh
```

**测试API：**
```bash
# 获取所有文章
curl http://localhost:9188/api/posts

# 获取单篇文章（自动增加浏览量）
curl http://localhost:9188/api/post?id=1

# 创建文章
curl -X POST "http://localhost:9188/api/posts?title=My%20Post&content=Content&author_id=1"

# 添加评论
curl -X POST "http://localhost:9188/api/comments?post_id=1&author=Alice&content=Great!"

# 获取评论
curl http://localhost:9188/api/comments?post_id=1
```

### `user_management.sh` - 用户管理系统
完整的用户管理示例，展示：
- 用户CRUD操作
- 缓存管理
- 缓存失效策略
- 数据查询优化

**运行方式：**
```bash
shode run examples/user_management.sh
```

### `session_management.sh` - 会话管理
使用缓存实现会话管理，包含：
- 会话创建和存储
- TTL管理
- 会话检索
- 模式匹配查找
- 会话失效

**运行方式：**
```bash
shode run examples/session_management.sh
```

### `rate_limiting.sh` - API限流
使用缓存实现API限流，展示：
- 请求计数
- 限流检查
- TTL窗口管理

**运行方式：**
```bash
shode run examples/rate_limiting.sh
```

### `data_aggregation.sh` - 数据聚合
数据聚合和缓存示例，包含：
- SQL聚合查询
- 结果缓存
- 多维度聚合

**运行方式：**
```bash
shode run examples/data_aggregation.sh
```

### `account_transfer.sh` - 账户转账
数据库事务模拟示例，展示：
- 账户余额管理
- 转账操作
- 数据一致性验证

**运行方式：**
```bash
shode run examples/account_transfer.sh
```

### `http_api_complex.sh` - 复杂HTTP API
综合示例，展示HTTP、数据库、缓存的集成使用。

**运行方式：**
```bash
shode run examples/http_api_complex.sh
```

## 示例场景说明

### 电商系统场景
- **产品管理**: 查询产品列表，使用缓存提高性能
- **订单处理**: 创建订单，自动使产品缓存失效
- **数据查询**: 支持分页、筛选等查询操作

### 博客系统场景
- **内容管理**: 文章和评论的CRUD操作
- **统计功能**: 浏览量统计、评论数统计
- **缓存策略**: 文章列表缓存、单篇文章缓存

### 用户管理场景
- **完整CRUD**: 创建、读取、更新、删除用户
- **缓存优化**: 用户列表缓存，更新时自动失效
- **查询优化**: 单用户查询、批量查询

### 会话管理场景
- **会话存储**: 使用缓存存储会话数据
- **TTL管理**: 自动过期机制
- **批量操作**: 模式匹配查找所有会话

### API限流场景
- **请求计数**: 使用缓存跟踪用户请求次数
- **限流策略**: 固定窗口限流
- **自动清理**: TTL自动清理过期计数

### 数据聚合场景
- **SQL聚合**: SUM、COUNT、GROUP BY等聚合操作
- **结果缓存**: 缓存聚合结果，减少数据库负载
- **多维度**: 支持按产品、日期等多维度聚合

### 金融交易场景
- **账户管理**: 账户余额管理
- **转账操作**: 多步骤数据库操作
- **一致性**: 确保数据一致性

## 注意事项

1. **数据库文件**: 示例使用SQLite，会在当前目录创建`.db`文件
2. **端口冲突**: HTTP服务器示例使用9188端口，确保端口未被占用
3. **缓存清理**: 缓存清理每1分钟运行一次
4. **语法说明**: 示例使用Shode语法，不是标准bash语法

## 运行所有示例

```bash
# 运行所有示例（非HTTP服务器示例）
for example in user_management.sh session_management.sh rate_limiting.sh data_aggregation.sh account_transfer.sh; do
    echo "Running $example..."
    shode run examples/$example
    echo ""
done
```

## 贡献

欢迎添加更多示例场景！示例应该：
- 展示实际使用场景
- 包含清晰的注释
- 提供运行说明
- 展示最佳实践
