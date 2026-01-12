# Spring化功能示例

本目录包含Shode框架Spring化功能的使用示例。

## 示例列表

### 1. `spring_ioc_example.sh` - IoC容器示例
演示依赖注入和Bean管理的基本概念。

**运行方式：**
```bash
shode run examples/spring_ioc_example.sh
```

**功能：**
- Bean注册
- 依赖注入概念
- Bean生命周期管理

### 2. `spring_config_example.sh` - 配置管理示例
演示配置文件的加载和访问。

**运行方式：**
```bash
shode run examples/spring_config_example.sh
```

**功能：**
- 配置文件加载
- 嵌套配置访问
- 环境特定配置
- 配置值获取（String, Int, Bool）
- 程序化配置设置

**示例配置：**
```json
{
  "server": {
    "port": 9188,
    "host": "localhost"
  },
  "database": {
    "url": "sqlite:app.db",
    "pool": {
      "max": 10,
      "min": 2
    }
  }
}
```

### 3. `spring_web_example.sh` - Web应用示例
演示Spring风格的Web应用架构。

**运行方式：**
```bash
shode run examples/spring_web_example.sh
```

**功能：**
- 配置驱动的服务器启动
- Service层模式
- Controller处理器
- HTTP路由注册
- RESTful API

**测试API：**
```bash
# 获取所有用户
curl http://localhost:9188/api/users

# 获取单个用户
curl http://localhost:9188/api/users/1

# 创建用户
curl -X POST http://localhost:9188/api/users -d '{"name":"John","email":"john@example.com"}'

# 健康检查
curl http://localhost:9188/health
```

### 4. `spring_transaction_example.sh` - 事务管理示例
演示事务性操作。

**运行方式：**
```bash
shode run examples/spring_transaction_example.sh
```

**功能：**
- 数据库连接
- 事务性转账操作
- 数据一致性验证

**注意：** 完整的事务管理（包括@Transactional注解和自动回滚）正在开发中。

### 5. `spring_complete_example.sh` - 完整应用示例 ⭐
演示完整的Spring风格应用，包含所有层。

**运行方式：**
```bash
shode run examples/spring_complete_example.sh
```

**架构层次：**
1. **配置层** - 配置管理
2. **数据访问层** - Repository模式
3. **服务层** - 业务逻辑
4. **控制器层** - HTTP处理
5. **缓存层** - 性能优化

**功能特性：**
- ✅ 配置管理
- ✅ 数据库操作
- ✅ Repository模式
- ✅ Service层
- ✅ Controller层
- ✅ 缓存策略
- ✅ HTTP路由
- ✅ RESTful API

**测试API：**
```bash
# 获取所有用户（带缓存）
curl http://localhost:9188/api/users

# 获取单个用户（带缓存）
curl http://localhost:9188/api/users/1

# 创建用户（自动失效缓存）
curl -X POST http://localhost:9188/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

## Spring化功能对比

| 功能 | Spring | Shode | 状态 |
|------|--------|-------|------|
| IoC容器 | ✅ | ✅ | 基础实现 |
| 依赖注入 | ✅ | ✅ | 基础实现 |
| 配置管理 | ✅ | ✅ | 完整实现 |
| 注解系统 | ✅ | ✅ | 基础实现 |
| 中间件 | ✅ | ✅ | 已实现 |
| 控制器 | ✅ | ✅ | 已实现 |
| 参数绑定 | ✅ | ✅ | 已实现 |
| 事务管理 | ✅ | ⚠️ | 基础实现 |
| Repository | ✅ | ✅ | 已实现 |
| AOP | ✅ | ✅ | 基础实现 |
| 条件注解 | ✅ | ✅ | 已实现 |
| 事件机制 | ✅ | ✅ | 已实现 |
| 健康检查 | ✅ | ✅ | 已实现 |

## 使用建议

1. **从简单开始**：先运行 `spring_config_example.sh` 了解配置管理
2. **逐步深入**：然后运行 `spring_web_example.sh` 了解Web层
3. **完整应用**：最后运行 `spring_complete_example.sh` 查看完整架构

## 注意事项

1. **IoC容器**：当前版本需要函数引用支持，完整实现待完善
2. **事务管理**：基础实现已完成，@Transactional注解支持开发中
3. **注解系统**：注解解析已完成，AST集成开发中
4. **中间件**：已实现，需要集成到HTTP服务器

## 下一步

- 完善IoC容器的函数引用支持
- 集成注解系统到AST解析器
- 实现@Transactional注解
- 完善中间件集成
- 添加更多Spring特性
