# Shode 完整应用示例

这是一个完整的 Shode 应用示例，展示了如何使用所有官方包。

## 功能特性

- ✅ 结构化日志 (@shode/logger)
- ✅ 配置管理 (@shode/config)
- ✅ HTTP 客户端 (@shode/http)
- ✅ 错误处理和日志记录
- ✅ 配置文件管理

## 快速开始

### 1. 安装依赖

```bash
shode pkg install
```

### 2. 运行应用

```bash
shode pkg run start
```

### 3. 查看输出

应用会显示详细的日志信息，包括：
- 应用启动信息
- 配置加载详情
- HTTP 客户端测试
- 配置管理演示

## 应用结构

```
complete-app/
├── shode.json           # 项目配置
├── src/
│   └── main.sh         # 主应用入口
├── config/
│   └── app.json        # 应用配置
└── README.md           # 本文件
```

## 技术栈

- **Shell 脚本**: 应用主逻辑
- **@shode/logger**: 结构化日志
- **@shode/config**: 配置管理
- **@shode/http**: HTTP 客户端

## 扩展功能

### 添加数据库支持

```bash
shode pkg add @shode/database ^1.0.0
```

然后在代码中使用：

```bash
. sh_modules/@shode/database/index.sh

DbConnect sqlite "./data/app.db"
results=$(DbQuery "SELECT * FROM users")
```

### 添加定时任务

```bash
shode pkg add @shode/cron ^1.0.0
```

```bash
. sh_modules/@shode/cron/index.sh

CronSchedule "0 * * * *" "hourly_task.sh"
CronStart &
```

## 学习资源

- [快速开始](../../docs/QUICKSTART.md)
- [最佳实践](../../docs/BEST_PRACTICES.md)
- [官方包文档](../../shode-registry/README.md)

## 许可证

MIT
