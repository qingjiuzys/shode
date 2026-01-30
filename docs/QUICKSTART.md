# Shode 快速开始指南

欢迎使用 Shode - 现代化的 Shell 脚本开发和包管理平台！

## 🚀 5 分钟快速上手

### 1. 安装 Shode

```bash
# 从源码编译
git clone https://github.com/shode-lang/shode.git
cd shode
go build -o shode ./cmd/shode

# 添加到 PATH
export PATH=$PATH:$PWD/shode
```

### 2. 创建你的第一个项目

```bash
# 初始化项目
shode pkg init my-first-app

# 项目结构已创建
ls -la
# shode.json - 项目配置文件
```

### 3. 使用官方包

```bash
# 添加日志包
shode pkg add @shode/logger ^1.0.0

# 添加配置管理
shode pkg add @shode/config ^1.0.0

# 安装依赖
shode pkg install

# 依赖被安装到 sh_modules/ 目录
ls sh_modules/
# @shode/logger/
# @shode/config/
```

### 4. 编写你的脚本

创建 `main.sh`:

```bash
#!/bin/sh
# 加载依赖
. sh_modules/@shode/logger/index.sh
. sh_modules/@shode/config/index.sh

# 使用日志
LogInfo "应用启动！"

# 使用配置
ConfigLoad "config.json"
api_host=$(ConfigGet "API_HOST" "localhost")

LogInfo "连接到 $api_host"
```

### 5. 运行脚本

```bash
chmod +x main.sh
./main.sh
# [2026-01-30T10:00:00.000Z] [info] 应用启动！
# [2026-01-30T10:00:00.100Z] [info] 连接到 localhost
```

---

## 📦 包管理核心命令

### 项目初始化

```bash
shode pkg init [name] [version]    # 初始化包
shode pkg add <package> [version]    # 添加依赖
shode pkg install                   # 安装所有依赖
shode pkg list                      # 列出依赖
```

### 版本管理

```bash
shode pkg update [package]          # 更新包
shode pkg update --latest          # 忽略 semver 更新到最新
shode pkg outdated                  # 检查过期包
shode pkg info <package>            # 查看包信息
shode pkg uninstall <package>       # 卸载包
```

### 发布

```bash
shode pkg search <query>            # 搜索包
shode pkg publish                    # 发布包到注册表
```

---

## 🌟 实用示例

### 示例 1: Web 服务

```bash
shode pkg init web-service
shode pkg add @shode/http ^1.0.0
shode pkg add @shode/logger ^1.0.0
shode pkg install
```

`src/main.sh`:
```bash
#!/bin/sh
. sh_modules/@shode/http/index.sh
. sh_modules/@shode/logger/index.sh

# 启动 HTTP 服务器
LogInfo "启动服务在 8080 端口"

# 处理请求
response=$(HttpGet "http://localhost:8080/api/health")
LogInfo "API 响应: $response"
```

### 示例 2: 定时任务

```bash
shode pkg init task-scheduler
shode pkg add @shode/cron ^1.0.0
shode pkg install
```

`src/scheduler.sh`:
```bash
#!/bin/sh
. sh_modules/@shode/cron/index.sh

# 每小时备份数据库
CronSchedule "0 * * * *" "./backup.sh"

# 每天清理日志
CronSchedule "0 0 * * *" "./cleanup.sh"

# 启动调度器
CronStart &
```

### 示例 3: 数据库应用

```bash
shode pkg init db-app
shode pkg add @shode/database ^1.0.0
shode pkg add @shode/config ^1.0.0
shode pkg install
```

`src/app.sh`:
```bash
#!/bin/sh
. sh_modules/@shode/database/index.sh
. sh_modules/@shode/config/index.sh

# 加载配置
ConfigLoad "config.json"
db_url=$(ConfigGet "DATABASE_URL")

# 连接数据库
DbConnect sqlite "$db_url"

# 查询数据
results=$(DbQuery "SELECT * FROM users")
echo "$results"
```

---

## 🔧 高级用法

### 语义版本范围

```bash
# 精确版本
shode pkg add lodash 4.17.21

# 兼容更新
shode pkg add lodash ^4.17.21    # >=4.17.21 <5.0.0
shode pkg add lodash ~4.17.21    # >=4.17.21 <4.18.0

# 范围
shode pkg add lodash ">=4.17.0"  # 4.17.0 或更高
shode pkg add lodash "1.x.x"      # 1.x.x 任何版本
```

### 脚本管理

在 `shode.json` 中定义脚本：

```json
{
  "name": "my-app",
  "version": "1.0.0",
  "scripts": {
    "start": "./src/main.sh",
    "test": "./tests/test.sh",
    "build": "./scripts/build.sh",
    "deploy": "./scripts/deploy.sh"
  }
}
```

运行脚本：

```bash
shode pkg run start
shode pkg run test
shode pkg run build
```

### 依赖锁定

```bash
# 生成锁文件
shode pkg install

# 查看锁文件
cat shode-lock.json

# 验证锁文件
shode pkg verify

# 更新特定包
shode pkg update lodash
```

---

## 📚 下一步

- 📖 阅读 [完整文档](../README.md)
- 🌟 查看 [官方包示例](../shode-registry/packages/)
- 🔧 查看 [API 文档](API.md)
- 💡 查看 [最佳实践](BEST_PRACTICES.md)

---

## 🆘 需要帮助？

- 📖 [文档](https://docs.shode.io)
- 💬 [Discord 社区](https://discord.gg/shode)
- 🐛 [报告问题](https://github.com/shode/shode/issues)

---

**Happy Coding! 🎉**
