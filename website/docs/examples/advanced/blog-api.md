# 博客 API 示例

## 简介

这是一个博客 API 示例，展示了文章管理、评论系统、浏览量统计等功能。

## 功能特性

- 文章管理（创建、查询、列表）
- 评论系统
- 浏览量统计
- 缓存优化
- 数据库操作

## 代码

查看完整代码：`examples/blog_api.sh`

主要功能：

```shode
# 启动 HTTP 服务器
StartHTTPServer "9188"

# 连接数据库
ConnectDB "sqlite" "blog.db"

# 创建表结构
ExecDB "CREATE TABLE IF NOT EXISTS posts (...)"
ExecDB "CREATE TABLE IF NOT EXISTS comments (...)"

# 定义处理函数
function handleGetPosts() {
    # 检查缓存
    cached = GetCache "posts:list"
    if cached != "" {
        SetHTTPResponse 200 cached
        return
    }
    
    # 查询文章和评论数
    QueryDB "SELECT p.*, COUNT(c.id) as comment_count FROM posts p LEFT JOIN comments c ON p.id = c.post_id GROUP BY p.id"
    result = GetQueryResult
    
    # 缓存结果
    SetCache "posts:list" result 300
    SetHTTPResponse 200 result
}

function handleGetPost() {
    postId = GetHTTPQuery "id"
    
    # 查询文章
    QueryDB "SELECT * FROM posts WHERE id = ?" postId
    result = GetQueryResult
    
    # 增加浏览量
    ExecDB "UPDATE posts SET views = views + 1 WHERE id = ?" postId
    
    # 缓存结果
    SetCache "post:" + postId result 600
    SetHTTPResponse 200 result
}
```

## 运行方式

```bash
shode run examples/blog_api.sh
```

## API 端点

- `GET /api/posts` - 获取文章列表（带评论数，缓存）
- `GET /api/post?id=1` - 获取单篇文章（自动增加浏览量，缓存）
- `POST /api/posts` - 创建文章
- `POST /api/comments` - 添加评论
- `GET /api/comments?post_id=1` - 获取文章评论

## 测试

```bash
# 获取文章列表
curl http://localhost:9188/api/posts

# 获取单篇文章（会增加浏览量）
curl http://localhost:9188/api/post?id=1

# 创建文章
curl -X POST "http://localhost:9188/api/posts?title=Title&content=Content&author_id=1"

# 添加评论
curl -X POST "http://localhost:9188/api/comments?post_id=1&author=Alice&content=Great!"

# 获取评论
curl http://localhost:9188/api/comments?post_id=1
```

## 关键特性

### 浏览量统计

每次查看文章时自动增加浏览量：

```shode
ExecDB "UPDATE posts SET views = views + 1 WHERE id = ?" postId
```

### 缓存策略

- 文章列表缓存 5 分钟
- 单篇文章缓存 10 分钟
- 创建/更新文章时自动失效缓存

### 数据库设计

- 文章表：id, title, content, author_id, views, created_at
- 评论表：id, post_id, author, content, created_at

## 相关文档

- [用户指南 - HTTP 服务器](../../guides/user-guide.md#6-http-服务器)
- [用户指南 - 数据库操作](../../guides/user-guide.md#8-数据库操作)
- [电商 API 示例](./ecommerce-api.md)
