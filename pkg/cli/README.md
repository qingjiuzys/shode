# CLI 工具增强 (CLI Tools Enhancement)

Shode 框架提供强大的命令行工具集，提升开发效率。

## 🛠️ 功能特性

### 1. 项目脚手架 (scaffold/)
- ✅ 快速项目生成
- ✅ 模块化结构
- ✅ 配置文件生成
- ✅ Docker 配置
- ✅ README 生成

### 2. 代码生成工具 (generate/)
- ✅ Model 生成
- ✅ CRUD 接口生成
- ✅ API 文档生成
- ✅ 测试代码生成
- ✅ 类型定义生成

### 3. 数据库迁移 (migrate/)
- ✅ 迁移文件生成
- ✅ 向上/向下迁移
- ✅ 迁移状态查看
- ✅ 数据库版本管理

### 4. 开发服务器 (server/)
- ✅ 热重载
- ✅ 环境变量管理
- ✅ 端口配置
- ✅ 代理设置

### 5. 构建打包 (build/)
- ✅ 多平台编译
- ✅ 版本管理
- ✅ Docker 镜像
- ✅ 安装包生成

## 🚀 快速开始

### 创建新项目

```bash
# 创建新项目
shode new myproject

# 或使用完整选项
shode new myproject \
  --template=rest-api \
  --db=postgres \
  --cache=redis \
  --docker
```

### 生成代码

```bash
# 生成 Model
shode generate model User \
  --fields="name:string,age:int,email:string"

# 生成 CRUD 接口
shode generate crud User

# 生成 API 文档
shode generate docs
```

### 数据库迁移

```bash
# 创建迁移
shode migrate create add_users_table

# 运行迁移
shode migrate up

# 回滚迁移
shode migrate down

# 查看状态
shode migrate status
```

### 开发服务器

```bash
# 启动开发服务器
shode dev

# 指定端口
shode dev --port=3000

# 启用热重载
shode dev --hot-reload
```

### 构建打包

```bash
# 构建当前平台
shode build

# 构建多平台
shode build --all

# 构建 Docker 镜像
shode build --docker
```

## 📋 命令参考

### shode new

创建新项目。

```bash
shode new <project-name> [options]
```

选项:
- `--template` 项目模板 (rest-api, grpc, microservice)
- `--db` 数据库类型 (postgres, mysql, sqlite, mongodb)
- `--cache` 缓存类型 (redis, memcached)
- `--docker` 包含 Docker 配置
- `--git` 初始化 Git 仓库

### shode generate

生成代码。

```bash
shode generate <type> <name> [options]
```

类型:
- `model` - 数据模型
- `crud` - CRUD 接口
- `handler` - HTTP 处理器
- `service` - 服务层
- `repository` - 数据访问层
- `docs` - API 文档
- `test` - 测试代码

### shode migrate

数据库迁移。

```bash
shode migrate <command> [options]
```

命令:
- `create` - 创建迁移文件
- `up` - 执行迁移
- `down` - 回滚迁移
- `status` - 查看状态
- `reset` - 重置数据库

### shode dev

启动开发服务器。

```bash
shode dev [options]
```

选项:
- `--port` 端口号 (默认: 8080)
- `--host` 主机地址 (默认: localhost)
- `--hot-reload` 启用热重载
- `--proxy` 代理设置
- `--env` 环境文件

### shode build

构建应用。

```bash
shode build [options]
```

选项:
- `--output` 输出路径
- `--os` 目标操作系统
- `--arch` 目标架构
- `--all` 构建所有平台
- `--docker` 构建 Docker 镜像
- `--compress` 压缩二进制文件

## 🔧 配置文件

### .shoderc

项目配置文件。

```yaml
# .shoderc
project:
  name: myproject
  version: 1.0.0

database:
  type: postgres
  host: localhost
  port: 5432
  name: myproject
  user: postgres
  password: password

cache:
  type: redis
  host: localhost
  port: 6379

server:
  port: 8080
  host: localhost

features:
  - auth
  - logging
  - metrics
```

## 📚 模板

### REST API 模板

创建 RESTful API 项目。

```bash
shode new myapi --template=rest-api
```

生成结构:
```
myapi/
├── cmd/
│   └── myapi/
│       └── main.go
├── internal/
│   ├── handler/
│   ├── service/
│   ├── repository/
│   └── model/
├── api/
│   └── openapi.yaml
├── config/
│   └── config.yaml
├── migrations/
├── Dockerfile
└── go.mod
```

### gRPC 模板

创建 gRPC 服务项目。

```bash
shode new mygrpc --template=grpc
```

## 🔌 插件扩展

### 自定义生成器

创建自定义代码生成器。

```go
package main

import (
    "github.com/myuser/myproject/generator"
)

func init() {
    generator.Register("mytype", MyGenerator)
}

func MyGenerator(params map[string]string) error {
    // 自定义生成逻辑
    return nil
}
```

### 自定义命令

添加自定义命令。

```bash
# 在 cmd/shode/main.go 中添加
cmd.AddCommand(&cobra.Command{
    Use:   "mycommand",
    Short: "My custom command",
    Run: func(cmd *cobra.Command, args []string) {
        // 命令逻辑
    },
})
```

## 🎯 最佳实践

1. **使用版本控制**: 始终使用 Git 管理项目
2. **环境隔离**: 使用不同的环境配置
3. **数据库迁移**: 随代码变更提交迁移文件
4. **测试驱动**: 生成代码后立即编写测试
5. **代码审查**: 定期审查生成的代码
6. **文档更新**: 保持文档与代码同步

## 🤝 贡献

欢迎贡献新的 CLI 工具功能！

## 📄 许可证

MIT License
