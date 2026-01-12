# 图书管理系统 (Library Management System)

这是一个使用 Shode 实现的完整图书管理系统，展示了如何使用 Shode 构建 Web API 应用。

## 功能特性

### 1. 用户认证
- ✅ 用户登录功能
- ✅ 密码使用 SHA256 加密存储
- ✅ 基于 Token 的会话管理
- ✅ 认证中间件保护 API 端点

### 2. 图书分类管理
- ✅ 列出所有分类
- ✅ 创建新分类
- ✅ 更新分类信息
- ✅ 删除分类（带安全检查）

### 3. 图书管理（CRUD）
- ✅ 列出所有图书
- ✅ 按分类查询图书
- ✅ 获取单本图书详情
- ✅ 创建新图书
- ✅ 更新图书信息
- ✅ 删除图书

## 技术栈

- **运行时**: Shode Shell Script Runtime
- **数据库**: SQLite
- **HTTP 服务器**: 内置 HTTP 服务器
- **缓存**: 内存缓存（用于会话管理）
- **加密**: SHA256 哈希算法

## 快速开始

### 1. 启动服务

```bash
# 方式1: 使用主文件（推荐）
./shode run examples/library_management.sh

# 方式2: 直接使用合并的模块文件
./shode run examples/library/all_modules.sh
```

服务将在 `http://localhost:9188` 启动。

### 2. 默认账户

- **用户名**: `admin`
- **密码**: `admin123`

### 3. API 使用示例

#### 登录获取 Token

```bash
curl 'http://localhost:9188/api/login?username=admin&password=admin123'
```

响应示例：
```json
{"token":"<session_token>","username":"admin"}
```

#### 使用 Token 访问 API

```bash
# 设置 token 变量
TOKEN="<your_token_from_login_response>"

# 列出所有分类
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/categories'

# 创建新分类
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/categories?name=Technology&description=Technology books'

# 列出所有图书
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/books'

# 创建新图书
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/books?title=Shode Guide&author=Shode Team&category_id=1&price=29.99&stock=10'

# 更新图书
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/books?id=1&title=New Title'

# 删除图书
curl -H "Authorization: $TOKEN" 'http://localhost:9188/api/books?id=1'
```

## API 端点

### 认证

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/login` | 用户登录，返回 token |

### 分类管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/categories` | 列出所有分类 |
| POST | `/api/categories` | 创建新分类 |
| PUT | `/api/categories` | 更新分类 |
| DELETE | `/api/categories` | 删除分类 |

### 图书管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/books` | 列出所有图书 |
| GET | `/api/books?category_id=1` | 按分类查询图书 |
| GET | `/api/books/:id` | 获取单本图书 |
| POST | `/api/books` | 创建新图书 |
| PUT | `/api/books` | 更新图书 |
| DELETE | `/api/books` | 删除图书 |

### 健康检查

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/health` | 健康检查 |

## 数据库结构

### users 表
- `id`: 主键
- `username`: 用户名（唯一）
- `password_hash`: 密码哈希（SHA256）
- `created_at`: 创建时间

### categories 表
- `id`: 主键
- `name`: 分类名称（唯一）
- `description`: 分类描述
- `created_at`: 创建时间

### books 表
- `id`: 主键
- `title`: 书名
- `author`: 作者
- `isbn`: ISBN 号
- `category_id`: 分类ID（外键）
- `price`: 价格
- `stock`: 库存
- `created_at`: 创建时间

## 安全特性

1. **密码加密**: 使用 SHA256 哈希算法加密密码
2. **Token 认证**: 基于缓存的会话 Token 管理
3. **API 保护**: 所有业务 API 都需要认证
4. **数据验证**: 输入参数验证和错误处理

## 代码结构

项目采用模块化设计，代码拆分为多个文件：

```
examples/
├── library_management.sh    # 主入口文件
└── library/                 # 模块目录
    ├── database.sh          # 数据库初始化模块
    │   └── initDatabase()   # 初始化数据库和默认数据
    ├── auth.sh              # 认证模块
    │   ├── login()          # 登录函数
    │   └── checkAuth()      # 认证中间件
    ├── categories.sh        # 分类管理模块
    │   ├── listCategories()    # 列出分类
    │   ├── createCategory()    # 创建分类
    │   ├── updateCategory()    # 更新分类
    │   └── deleteCategory()    # 删除分类
    ├── books.sh             # 图书管理模块
    │   ├── listBooks()      # 列出图书
    │   ├── getBook()        # 获取图书
    │   ├── createBook()     # 创建图书
    │   ├── updateBook()     # 更新图书
    │   └── deleteBook()     # 删除图书
    ├── handlers.sh          # HTTP路由处理器模块
    │   ├── handleLogin()    # 登录处理器
    │   ├── handleListCategories()  # 分类列表处理器
    │   ├── handleCreateCategory()  # 创建分类处理器
    │   ├── handleUpdateCategory()  # 更新分类处理器
    │   ├── handleDeleteCategory()  # 删除分类处理器
    │   ├── handleListBooks()       # 图书列表处理器
    │   ├── handleGetBook()         # 获取图书处理器
    │   ├── handleCreateBook()      # 创建图书处理器
    │   ├── handleUpdateBook()      # 更新图书处理器
    │   └── handleDeleteBook()      # 删除图书处理器
    └── all_modules.sh       # 合并所有模块（自动生成）
```

### 模块说明

- **database.sh**: 负责数据库连接、表结构创建和默认数据初始化
- **auth.sh**: 提供用户认证功能，包括登录和Token验证
- **categories.sh**: 实现图书分类的CRUD操作
- **books.sh**: 实现图书的CRUD操作
- **handlers.sh**: 将所有业务函数封装为HTTP路由处理器，并添加认证中间件
- **library_management.sh**: 主入口文件，加载所有模块并启动HTTP服务器

## 注意事项

1. **数据库文件**: 数据库文件存储在 `test/tmp/library.db`
2. **会话过期**: Token 会话有效期为 3600 秒（1小时）
3. **分类删除**: 如果分类下有图书，无法删除分类
4. **参数传递**: 当前使用查询参数传递数据，生产环境建议使用 JSON Body

## 扩展建议

1. **JSON 解析**: 实现 JSON 解析功能，支持 JSON Body
2. **用户注册**: 添加用户注册功能
3. **权限管理**: 实现基于角色的访问控制（RBAC）
4. **数据分页**: 为列表接口添加分页功能
5. **搜索功能**: 添加图书搜索功能
6. **文件上传**: 支持图书封面图片上传

## 相关文档

- [Shode Shell 特性指南](../docs/guides/SHELL_FEATURES.md)
- [Linux 命令支持情况](../docs/guides/LINUX_COMMANDS_SUPPORT.md)
- [执行引擎文档](../docs/EXECUTION_ENGINE.md)
