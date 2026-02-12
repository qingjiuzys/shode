# Shode 实用工具集

## 📦 新增工具

### 1. API客户端生成器 (API Client Generator)

自动根据OpenAPI规范生成类型安全的HTTP客户端代码。

**位置：** `pkg/codegen/client/`

**功能特性：**
- ✅ 从OpenAPI 3.0规范生成客户端
- ✅ 自动生成类型定义
- ✅ 生成CRUD方法
- ✅ 支持路径参数、查询参数、请求体
- ✅ 自定义配置选项

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/codegen/client"

gen := clientgen.NewGenerator()
gen.LoadSpec("openapi.json")
gen.Package = "myclient"
gen.ClientName = "MyAPIClient"
gen.OutputDir = "./client"

// 生成客户端代码
err := gen.Generate()
if err != nil {
    log.Fatal(err)
}
```

**生成的代码包括：**
- `client.go` - HTTP客户端核心实现
- `api.go` - API方法
- `models.go` - 数据模型
- `config.go` - 配置管理

### 2. 日志分析器 (Log Analyzer)

强大的日志文件分析工具，支持模式匹配、搜索、过滤和统计。

**位置：** `pkg/loganalyzer/`

**功能特性：**
- ✅ 多种日志模式支持
- ✅ 正则表达式搜索
- ✅ 按级别/时间过滤
- ✅ 错误统计和趋势分析
- ✅ 实时日志监控
- ✅ 导出为JSON/CSV/文本

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/loganalyzer"

analyzer := loganalyzer.NewAnalyzer("app.log")

// 添加日志模式
analyzer.AddPattern("error", `\[ERROR\].*`, "ERROR")
analyzer.AddPattern("warning", `\[WARN\].*`, "WARNING")
analyzer.AddPattern("info", `\[INFO\].*`, "INFO")

// 解析日志文件
err := analyzer.Parse()
if err != nil {
    log.Fatal(err)
}

// 获取统计信息
stats := analyzer.GetStats()
fmt.Printf("Errors: %d, Warnings: %d\n", stats.ErrorCount, stats.WarningCount)

// 搜索错误日志
errors := analyzer.GetErrors()
for _, err := range errors {
    fmt.Printf("[%s] %s\n", err.Timestamp, err.Message)
}

// 搜索特定关键词
results := analyzer.Search("database")

// 按时间范围过滤
start := time.Now().Add(-24 * time.Hour)
end := time.Now()
filtered := analyzer.FilterByTime(start, end)

// 获取错误率
errorRate := analyzer.GetErrorRate()
fmt.Printf("Error Rate: %.2f%%\n", errorRate)

// 获取最常见的错误
topErrors := analyzer.GetTopErrors(10)
for i, err := range topErrors {
    fmt.Printf("%d. [%d] %s\n", i+1, err.Count, err.Message)
}

// 打印报告
analyzer.PrintReport()

// 导出分析结果
analyzer.Export("output.json", "json")

// 实时监控日志文件
logChan := analyzer.Watch(5 * time.Second)
for entry := range logChan {
    if entry.Level == "ERROR" {
        fmt.Printf("New error: %s\n", entry.Message)
    }
}
```

**高级用法：自定义提取器**

```go
// 使用自定义提取器提取结构化数据
analyzer.AddPatternWithExtractor(
    "api_request",
    `\[API\] (\w+) (\S+) from (\d+\.\d+\.\d+\.\d+)`,
    "INFO",
    func(matches []string) map[string]interface{} {
        return map[string]interface{}{
            "method": matches[1],
            "path":   matches[2],
            "ip":     matches[3],
        }
    },
)
```

### 3. 数据库迁移工具 (Database Migration)

完整的数据库schema版本管理和迁移工具。

**位置：** `pkg/database/migrate/`

**功能特性：**
- ✅ 版本管理
- ✅ 向上/向下迁移
- ✅ 迁移历史记录
- ✅ 支持多种数据库 (MySQL, PostgreSQL, SQLite)
- ✅ 自动迁移文件加载
- ✅ 迁移验证

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/database/migrate"

migrator := migrate.NewMigrator(&migrate.Config{
    DB:      db,
    Dialect: "sqlite3",
})

// 初始化迁移表
err := migrator.Init()
if err != nil {
    log.Fatal(err)
}

// 从目录加载迁移文件
err = migrator.LoadMigrationsFromDir("./migrations")
if err != nil {
    log.Fatal(err)
}

// 查看状态
migrator.PrintStatus()

// 执行所有待执行的迁移
err = migrator.Up()
if err != nil {
    log.Fatal(err)
}

// 迁移到指定版本
err = migrator.UpTo(5)
if err != nil {
    log.Fatal(err)
}

// 回滚最近的迁移
err = migrator.Down()
if err != nil {
    log.Fatal(err)
}

// 回滚到指定版本
err = migrator.DownTo(3)
if err != nil {
    log.Fatal(err)
}

// 重做最后一次迁移
err = migrator.Redo()
if err != nil {
    log.Fatal(err)
}

// 重置所有迁移
err = migrator.Reset()
if err != nil {
    log.Fatal(err)
}

// 创建新的迁移文件
err = migrator.Create("./migrations", "Add users table")
if err != nil {
    log.Fatal(err)
}
```

**迁移文件命名规范：**

```
001_create_users_table.up.sql
001_create_users_table.down.sql
002_add_email_index.up.sql
002_add_email_index.down.sql
...
```

**迁移文件示例：**

```sql
-- 001_create_users_table.up.sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 001_create_users_table.down.sql
DROP TABLE users;
```

## 🎯 工具对比

| 工具 | 用途 | 输入 | 输出 |
|------|------|------|------|
| API客户端生成器 | 生成API客户端 | OpenAPI规范 | Go代码 |
| 日志分析器 | 分析日志文件 | 日志文件 | 统计、过滤结果 |
| 数据库迁移 | Schema版本管理 | SQL迁移文件 | 数据库变更 |

## 📚 使用场景

### 场景1: 快速集成REST API

```bash
# 1. 下载API的OpenAPI规范
wget https://api.example.com/openapi.json

# 2. 生成客户端代码
go run tools/gen-client.go -spec openapi.json -output ./client

# 3. 在代码中使用
import "myclient"

client := myclient.NewClient(nil)
users, _ := client.GetUsers(context.Background())
```

### 场景2: 生产环境日志分析

```go
// 分析过去24小时的错误
analyzer := loganalyzer.NewAnalyzer("production.log")
analyzer.Parse()

start := time.Now().Add(-24 * time.Hour)
errors := analyzer.FilterByTime(start, time.Now())

// 找出Top错误
topErrors := analyzer.GetTopErrors(10)
for _, err := range topErrors {
    fmt.Printf("Fix error: %s (occurred %d times)\n", err.Message, err.Count)
}
```

### 场景3: 数据库Schema管理

```bash
# 创建新迁移
migrate create AddIndexField

# 生成迁移文件:
# - 006_addindexfield.up.sql
# - 006_addindexfield.down.sql

# 编辑迁移文件后执行
migrate up

# 查看状态
migrate status

# 如果需要回滚
migrate down
```

## 🔄 工作流集成

### Git Hooks集成

```bash
# .git/hooks/pre-commit
#!/bin/bash
# 分析日志错误
go run ./scripts/analyze-logs.go
if [ $? -ne 0 ]; then
    echo "Logs contain errors"
    exit 1
fi
```

### CI/CD集成

```yaml
# .github/workflows/test.yml
- name: Run migrations
  run: |
    go run ./cmd/migrate/main.go up

- name: Analyze logs
  run: |
    go run ./scripts/analyze-logs.go > report.json

- name: Upload report
  uses: actions/upload-artifact@v2
  with:
    name: log-analysis
    path: report.json
```

## 📖 最佳实践

### 日志分析

1. **使用结构化日志**
```go
logger.WithFields(log.Fields{
    "user_id": userID,
    "action": "login",
    "ip": clientIP,
}).Info("User logged in")
```

2. **添加自定义模式**
```go
analyzer.AddPattern("slow_query",
    `Slow query: (\d+)ms for (.+)`,
    "WARNING")
```

### API客户端生成

1. **保持API规范更新**
2. **定期重新生成客户端**
3. **使用版本控制**
4. **自定义生成模板**

### 数据库迁移

1. **原子性操作** - 每个迁移应该是原子的
2. **可逆性** - 始终编写down迁移
3. **测试** - 在开发环境测试迁移
4. **备份** - 执行迁移前备份数据库

## 🚀 快速开始

### 安装依赖

```bash
go get gitee.com/com_818cloud/shode/pkg/codegen/client
go get gitee.com/com_818cloud/shode/pkg/loganalyzer
go get gitee.com/com_818cloud/shode/pkg/database/migrate
```

### 示例代码

查看 `examples/` 目录获取更多使用示例。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
