// Package main 演示开发者工具的使用
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gitee.com/com_818cloud/shode/pkg/devtools/apidoc"
	"gitee.com/com_818cloud/shode/pkg/devtools/codegen"
	"gitee.com/com_818cloud/shode/pkg/devtools/config"
	"gitee.com/com_818cloud/shode/pkg/devtools/depanalyzer"
	"gitee.com/com_818cloud/shode/pkg/devtools/profiler"
)

func main() {
	fmt.Println("=== Shode Developer Tools Demo ===\n")

	// 1. 代码生成示例
	fmt.Println("1. Code Generation Demo")
	fmt.Println("===========================")
	if err := demoCodeGeneration(); err != nil {
		log.Printf("Code generation error: %v\n", err)
	}

	// 2. 配置验证示例
	fmt.Println("\n2. Configuration Validation Demo")
	fmt.Println("===================================")
	demoConfigValidation()

	// 3. 依赖分析示例
	fmt.Println("\n3. Dependency Analysis Demo")
	fmt.Println("==============================")
	demoDependencyAnalysis()

	// 4. API 文档生成示例
	fmt.Println("\n4. API Documentation Generation Demo")
	fmt.Println("========================================")
	if err := demoAPIDocGeneration(); err != nil {
		log.Printf("API doc generation error: %v\n", err)
	}

	// 5. 性能分析示例
	fmt.Println("\n5. Performance Profiling Demo")
	fmt.Println("================================")
	demoProfiling()

	fmt.Println("\n=== Demo Complete ===")
}

// demoCodeGeneration 演示代码生成
func demoCodeGeneration() error {
	// 创建代码生成器
	gen := codegen.NewGenerator("model", "User")

	// 添加字段
	gen.AddField("Username", "string", `json:"username" gorm:"uniqueIndex"`)
	gen.AddField("Email", "string", `json:"email" gorm:"uniqueIndex"`)
	gen.AddField("Password", "string", `json:"-"`)
	gen.AddField("Age", "int", `json:"age"`)
	gen.AddField("IsActive", "bool", `json:"is_active"`)

	// 设置输出路径
	gen.OutputPath = "./tmp"

	// 生成 Model
	if err := gen.GenerateModel(); err != nil {
		return err
	}

	// 生成 Repository
	if err := gen.GenerateRepository(); err != nil {
		return err
	}

	// 生成 Service
	if err := gen.GenerateService(); err != nil {
		return err
	}

	// 生成 Handler
	if err := gen.GenerateHandler(); err != nil {
		return err
	}

	fmt.Println("✓ Generated User model, repository, service, and handler")
	return nil
}

// demoConfigValidation 演示配置验证
func demoConfigValidation() {
	// 定义配置结构
	type Config struct {
		Host     string `validate:"required,ip"`
		Port     int    `validate:"required,port"`
		Database string `validate:"required,min=3,max=32"`
		Email    string `validate:"required,email"`
		LogLevel string `validate:"required,oneof=debug|info|warn|error"`
	}

	// 有效配置
	validConfig := Config{
		Host:     "127.0.0.1",
		Port:     8080,
		Database: "myapp_db",
		Email:    "admin@example.com",
		LogLevel: "info",
	}

	validator := config.NewValidator()
	if err := validator.Validate(&validConfig); err != nil {
		fmt.Printf("✗ Valid config failed: %v\n", err)
	} else {
		fmt.Println("✓ Valid configuration passed")
	}

	// 无效配置
	invalidConfig := Config{
		Host:     "invalid-ip",
		Port:     99999,
		Database: "db",
		Email:    "not-an-email",
		LogLevel: "trace",
	}

	if err := validator.Validate(&invalidConfig); err != nil {
		fmt.Printf("✓ Invalid config correctly rejected:\n%v\n", err)
	} else {
		fmt.Println("✗ Invalid config incorrectly passed")
	}
}

// demoDependencyAnalysis 演示依赖分析
func demoDependencyAnalysis() {
	analyzer := depanalyzer.NewAnalyzer()

	// 忽略标准库
	analyzer.IgnorePackage("C")

	// 分析当前目录
	if err := analyzer.Analyze("."); err != nil {
		fmt.Printf("Analysis error: %v\n", err)
		return
	}

	// 打印报告
	analyzer.PrintReport()

	// 获取包统计
	stats := analyzer.GetPackageStatistics()
	fmt.Printf("Statistics: total=%d, files=%d, imports=%d\n",
		stats["total"], stats["files"], stats["imports"])

	// 查找未使用的包
	unused := analyzer.FindUnusedPackages()
	if len(unused) > 0 {
		fmt.Printf("Unused packages (%d):\n", len(unused))
		for _, pkg := range unused {
			fmt.Printf("  - %s\n", pkg)
		}
	}
}

// demoAPIDocGeneration 演示 API 文档生成
func demoAPIDocGeneration() error {
	gen := apidoc.NewGenerator("My API", "1.0.0")
	gen.SetOutputDir("./tmp")

	// 添加标签
	gen.AddTag("users", "User management operations")
	gen.AddTag("products", "Product catalog operations")

	// 定义 User 模型
	gen.AddDefinition("User", &apidoc.Schema{
		Type: "object",
		Properties: map[string]*apidoc.Property{
			"id": {
				Type:        "integer",
				Description: "User ID",
				Format:      "int64",
			},
			"username": {
				Type:        "string",
				Description: "Username",
			},
			"email": {
				Type:        "string",
				Description: "Email address",
				Format:      "email",
			},
		},
		Required: []string{"id", "username", "email"},
	})

	// 添加路径
	gen.AddPath("GET", "/api/users", &apidoc.Path{
		Method:      "GET",
		Summary:     "List users",
		Description: "Get a paginated list of users",
		Tags:        []string{"users"},
		Responses: map[int]*apidoc.Response{
			200: {
				Description: "Success",
				Schema: &apidoc.Schema{Ref: "#/definitions/User"},
			},
		},
	})

	gen.AddPath("POST", "/api/users", &apidoc.Path{
		Method:      "POST",
		Summary:     "Create user",
		Description: "Create a new user",
		Tags:        []string{"users"},
		Parameters: []apidoc.Parameter{
			{
				Name:     "body",
				In:       "body",
				Required: true,
				Schema:   &apidoc.Schema{Ref: "#/definitions/User"},
			},
		},
		Responses: map[int]*apidoc.Response{
			201: {
				Description: "User created",
				Schema:      &apidoc.Schema{Ref: "#/definitions/User"},
			},
		},
	})

	// 生成 OpenAPI 规范
	if err := gen.GenerateOpenAPI(); err != nil {
		return err
	}

	// 生成 Markdown 文档
	if err := gen.GenerateMarkdown(); err != nil {
		return err
	}

	return nil
}

// demoProfiling 演示性能分析
func demoProfiling() {
	// 创建性能分析器
	p := profiler.NewProfiler(&profiler.Config{
		CPUProfile:     "./tmp/cpu.prof",
		MemProfile:     "./tmp/mem.prof",
		BlockProfile:   "./tmp/block.prof",
		MutexProfile:   "./tmp/mutex.prof",
		RecordMemStats: true,
	})
	defer p.Stop()

	// 启动内存监控
	p.StartMemStatsMonitor(5 * time.Second)

	// 打印内存统计
	p.PrintMemStats()

	// 基准测试
	bench := profiler.NewBenchmark("test_operation")
	bench.RunMultiple(1000, func() {
		// 模拟一些工作
		sum := 0
		for i := 0; i < 100; i++ {
			sum += i
		}
		_ = sum
	})

	// 比较两个函数
	fmt.Println("\nPerformance Comparison:")
	profiler.Comparison("slice", "array", func() {
		data := make([]int, 1000)
		for i := range data {
			data[i] = i
		}
		_ = data
	}, func() {
		data := [1000]int{}
		for i := range data {
			data[i] = i
		}
		_ = data
	})

	// 获取内存快照
	if err := p.Snapshot("./tmp/mem_snapshot.prof"); err != nil {
		fmt.Printf("Snapshot error: %v\n", err)
	}
}

func init() {
	// 创建临时目录
	os.MkdirAll("./tmp", 0755)
}
