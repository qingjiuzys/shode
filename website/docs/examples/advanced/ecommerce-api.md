# 电商 API 示例

## 简介

这是一个完整的电商 API 示例，展示了产品管理、订单创建、缓存策略等实际应用场景。

## 功能特性

- 产品管理（查询、缓存）
- 订单创建和管理
- 缓存策略优化
- HTTP 路由注册
- 数据库操作

## 代码

查看完整代码：`examples/ecommerce_api.sh`

主要功能：

```shode
# 启动 HTTP 服务器
StartHTTPServer "9188"

# 连接数据库
ConnectDB "sqlite" "ecommerce.db"

# 创建表结构
ExecDB "CREATE TABLE IF NOT EXISTS products (...)"
ExecDB "CREATE TABLE IF NOT EXISTS orders (...)"

# 定义处理函数
function handleGetProducts() {
    # 检查缓存
    cached = GetCache "products:list"
    if cached != "" {
        SetHTTPResponse 200 cached
        return
    }
    
    # 查询数据库
    QueryDB "SELECT * FROM products WHERE stock > 0"
    result = GetQueryResult
    
    # 缓存结果
    SetCache "products:list" result 300
    SetHTTPResponse 200 result
}

# 注册路由
RegisterHTTPRoute "GET" "/api/products" "function" "handleGetProducts"
RegisterHTTPRoute "POST" "/api/orders" "function" "handleCreateOrder"
```

## 运行方式

```bash
shode run examples/ecommerce_api.sh
```

## API 端点

- `GET /api/products` - 获取产品列表（带缓存）
- `GET /api/product?id=1` - 获取单个产品（带缓存）
- `POST /api/orders` - 创建订单
- `GET /api/orders?user_id=1` - 获取用户订单
- `GET /api/health` - 健康检查

## 测试

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

## 关键特性

### 缓存策略

- 产品列表缓存 5 分钟
- 单个产品缓存 10 分钟
- 订单创建时自动失效相关缓存

### 数据库设计

- 产品表：id, name, price, stock
- 订单表：id, user_id, total, status

## 相关文档

- [用户指南 - HTTP 服务器](../../guides/user-guide.md#6-http-服务器)
- [用户指南 - 缓存系统](../../guides/user-guide.md#7-缓存系统)
- [用户指南 - 数据库操作](../../guides/user-guide.md#8-数据库操作)
