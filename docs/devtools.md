# Shode 开发者工具

Shode 提供了一套强大的开发者工具，帮助您更高效地开发和维护应用程序。

## 📦 工具列表

### 1. 代码生成工具 (Code Generator)

自动生成常用代码模板，包括 Model、Repository、Service 和 Handler。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/devtools/codegen"

// 创建生成器
gen := codegen.NewGenerator("model", "User")

// 添加字段
gen.AddField("Username", "string", `json:"username" gorm:"uniqueIndex"`)
gen.AddField("Email", "string", `json:"email" gorm:"uniqueIndex"`)
gen.AddField("Password", "string", `json:"-"`)

// 设置输出路径
gen.OutputPath = "./internal/model"

// 生成代码
gen.GenerateModel()       // 生成 Model
gen.GenerateRepository()  // 生成 Repository 接口和实现
gen.GenerateService()     // 生成 Service 层
gen.GenerateHandler()     // 生成 HTTP Handler
```

**特性：**
- ✅ 自动生成标准的 CRUD 操作
- ✅ 支持 GORM 标签
- ✅ 遵循最佳实践的代码结构
- ✅ 自动转换为 snake_case 表名

### 2. 性能分析工具 (Profiler)

全面的性能分析工具，帮助您找出性能瓶颈。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/devtools/profiler"

// 创建性能分析器
p := profiler.NewProfiler(&profiler.Config{
    CPUProfile:     "./cpu.prof",
    MemProfile:     "./mem.prof",
    BlockProfile:   "./block.prof",
    MutexProfile:   "./mutex.prof",
    RecordMemStats: true,
})
defer p.Stop()

// 启动内存监控
p.StartMemStatsMonitor(5 * time.Second)

// 打印内存统计
p.PrintMemStats()

// 基准测试
bench := profiler.NewBenchmark("my_operation")
duration := bench.RunMultiple(1000, func() {
    // 您的代码
})

// 性能对比
profiler.Comparison("method1", "method2", func() {
    // 方法 1
}, func() {
    // 方法 2
})

// 获取内存快照
p.Snapshot("./mem_snapshot.prof")
```

**特性：**
- ✅ CPU 性能分析
- ✅ 内存使用分析
- ✅ Goroutine 阻塞分析
- ✅ Mutex 锁竞争分析
- ✅ 实时内存监控
- ✅ 基准测试辅助
- ✅ 函数性能对比

### 3. 配置验证工具 (Config Validator)

声明式配置验证，确保配置的正确性。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/devtools/config"

type Config struct {
    Host     string `validate:"required,ip"`
    Port     int    `validate:"required,port"`
    Database string `validate:"required,min=3,max=32"`
    Email    string `validate:"required,email"`
    URL      string `validate:"required,url"`
    LogLevel string `validate:"required,oneof=debug|info|warn|error"`
}

validator := config.NewValidator()
err := validator.Validate(&config)
if err != nil {
    log.Fatal("Invalid config:", err)
}
```

**验证规则：**
- `required` - 必填字段
- `min` - 最小值/长度
- `max` - 最大值/长度
- `email` - 邮箱格式
- `url` - URL 格式
- `port` - 端口号范围 (1-65535)
- `ip` - IP 地址格式
- `oneof` - 枚举值
- `env` - 环境变量
- `file` - 文件存在性

### 4. 依赖分析工具 (Dependency Analyzer)

分析项目的依赖关系，发现潜在问题。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/devtools/depanalyzer"

analyzer := depanalyzer.NewAnalyzer()

// 忽略某些包
analyzer.IgnorePackage("vendor")
analyzer.IgnorePackage("C")

// 分析项目
err := analyzer.Analyze("./...")
if err != nil {
    log.Fatal(err)
}

// 打印报告
analyzer.PrintReport()

// 获取统计信息
stats := analyzer.GetPackageStatistics()

// 查找未使用的包
unused := analyzer.FindUnusedPackages()

// 获取导入树
tree := analyzer.GetImportTree("main", 0)
```

**特性：**
- ✅ 依赖关系可视化
- ✅ 循环依赖检测
- ✅ 未使用包检测
- ✅ 依赖分类（内部/外部/标准库）
- ✅ 导入树生成

### 5. API 文档生成工具 (API Doc Generator)

自动生成 OpenAPI 规范和 Markdown 文档。

**使用示例：**

```go
import "gitee.com/com_818cloud/shode/pkg/devtools/apidoc"

gen := apidoc.NewGenerator("My API", "1.0.0")
gen.SetOutputDir("./docs")

// 添加标签
gen.AddTag("users", "User management")

// 定义数据模型
gen.AddDefinition("User", &apidoc.Schema{
    Type: "object",
    Properties: map[string]*apidoc.Property{
        "id": {Type: "integer", Description: "User ID"},
        "name": {Type: "string", Description: "User name"},
    },
    Required: []string{"id", "name"},
})

// 添加 API 端点
gen.AddPath("GET", "/api/users", &apidoc.Path{
    Summary:     "List users",
    Description: "Get all users",
    Tags:        []string{"users"},
    Responses: map[int]*apidoc.Response{
        200: {
            Description: "Success",
            Schema:      &apidoc.Schema{Ref: "#/definitions/User"},
        },
    },
})

// 生成文档
gen.GenerateOpenAPI()   // OpenAPI 3.0 JSON
gen.GenerateMarkdown()  // Markdown 文档
```

**特性：**
- ✅ OpenAPI 3.0 规范生成
- ✅ Markdown 文档生成
- ✅ 支持标签分组
- ✅ 数据模型定义
- ✅ 参数和响应定义

## 🚀 快速开始

### 安装

```bash
go get gitee.com/com_818cloud/shode/pkg/devtools/...
```

### 示例项目

```bash
# 运行开发者工具演示
cd examples/devtools
go run main.go
```

### 代码生成示例

```bash
# 创建一个新的 Model
cat > main.go << 'EOF'
package main

import (
    "gitee.com/com_818cloud/shode/pkg/devtools/codegen"
)

func main() {
    gen := codegen.NewGenerator("model", "Product")
    gen.AddField("Name", "string", `json:"name"`)
    gen.AddField("Price", "float64", `json:"price"`)
    gen.AddField("Stock", "int", `json:"stock"`)
    gen.OutputPath = "./internal"

    gen.GenerateModel()
    gen.GenerateRepository()
    gen.GenerateService()
    gen.GenerateHandler()
}
EOF

go run main.go
```

## 📚 文档

- [API 文档](./api.md)
- [配置指南](./config.md)
- [最佳实践](./best-practices.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
