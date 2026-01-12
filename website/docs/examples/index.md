# 示例集合

本目录包含各种 Shode 使用示例，展示 HTTP 服务器、缓存、数据库等功能的实际应用场景。

## 基础示例

### [HTTP 服务器](./basic/http-server.md)
简单的 HTTP 服务器示例，展示如何启动服务器和注册基本路由。

### [缓存系统](./basic/cache.md)
缓存系统基础示例，展示缓存的设置、获取、删除等操作。

### [数据库操作](./basic/database.md)
数据库操作示例，展示如何连接数据库、执行查询和更新操作。

## 高级示例

### [电商 API](./advanced/ecommerce-api.md)
完整的电商 API 示例，包含产品管理、订单创建、缓存策略等。

### [博客 API](./advanced/blog-api.md)
博客 API 示例，展示文章管理、评论系统、浏览量统计等功能。

### [Spring 功能](./advanced/spring-features.md)
Spring 化功能完整示例，展示 IoC、配置管理、Web 层、事务管理等企业级特性。

## 快速开始

选择一个示例，查看详细文档并运行：

```bash
# 运行 HTTP 服务器示例
shode run examples/http_server.sh

# 运行缓存示例
shode run examples/cache_example.sh

# 运行数据库示例
shode run examples/database_example.sh
```

## 更多示例

查看 `examples/` 目录获取更多示例：
- `ecommerce_api.sh` - 电商 API
- `blog_api.sh` - 博客 API
- `spring_complete_example.sh` - Spring 完整示例
- `user_management.sh` - 用户管理系统
- `session_management.sh` - 会话管理
- `rate_limiting.sh` - 限流示例

## 相关文档

- [快速开始](../getting-started/quick-start.md)
- [用户指南](../guides/user-guide.md)
- [API 参考](../api/stdlib.md)
