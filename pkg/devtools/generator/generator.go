// Package generator 代码生成器
package generator

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Generator 代码生成器
type Generator struct {
	config     *GeneratorConfig
	templates *TemplateRegistry
	outputDir  string
	dryRun     bool
}

// GeneratorConfig 生成器配置
type GeneratorConfig struct {
	PackageName    string
	ProjectName    string
	Author         string
	Description    string
	License        string
	Version        string
	Features       []string
}

// TemplateRegistry 模板注册表
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// NewGenerator 创建代码生成器
func NewGenerator(config *GeneratorConfig) *Generator {
	return &Generator{
		config:     config,
		templates: NewTemplateRegistry(),
		outputDir:  ".",
		dryRun:     false,
	}
}

// NewTemplateRegistry 创建模板注册表
func NewTemplateRegistry() *TemplateRegistry {
	registry := &TemplateRegistry{
		templates: make(map[string]*template.Template),
	}
	registry.RegisterBuiltinTemplates()
	return registry
}

// RegisterBuiltinTemplates 注册内置模板
func (tr *TemplateRegistry) RegisterBuiltinTemplates() {
	// 项目模板
	tr.RegisterTemplate("project", projectTemplate)
	tr.RegisterTemplate("controller", controllerTemplate)
	tr.RegisterTemplate("model", modelTemplate)
	tr.RegisterTemplate("service", serviceTemplate)
	tr.RegisterTemplate("test", testTemplate)
	tr.RegisterTemplate("config", configTemplate)
	tr.RegisterTemplate("readme", readmeTemplate)
}

// RegisterTemplate 注册模板
func (tr *TemplateRegistry) RegisterTemplate(name, content string) {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse template %s: %v", name, err))
	}
	tr.templates[name] = tmpl
}

// GenerateProject 生成项目
func (g *Generator) GenerateProject(ctx context.Context, projectName string) error {
	fmt.Printf("🚀 Generating project: %s\n", projectName)

	// 创建项目目录
	projectDir := filepath.Join(g.outputDir, projectName)
	if !g.dryRun {
		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}
	}

	// 准备模板数据
	data := map[string]interface{}{
		"ProjectName": projectName,
		"PackageName": strings.ToLower(projectName),
		"Author":      g.config.Author,
		"Description": g.config.Description,
		"License":     g.config.License,
		"Version":     g.config.Version,
		"Date":        time.Now().Format("2006-01-02"),
	}

	// 生成项目文件
	files := []struct {
		name     string
		template string
		path     string
	}{
		{"main.shode", "project", filepath.Join(projectDir, "main.shode")},
		{"config.shode", "config", filepath.Join(projectDir, "config.shode")},
		{"README.md", "readme", filepath.Join(projectDir, "README.md")},
	}

	for _, file := range files {
		if err := g.generateFile(file.template, file.path, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", file.name, err)
		}
		fmt.Printf("  ✓ Generated %s\n", file.name)
	}

	// 生成目录结构
	dirs := []string{"src", "tests", "docs", "config"}
	for _, dir := range dirs {
		dirPath := filepath.Join(projectDir, dir)
		if !g.dryRun {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
		fmt.Printf("  ✓ Created directory: %s/\n", dir)
	}

	fmt.Printf("\n✅ Project %s generated successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Printf("  shode run main.shode\n")

	return nil
}

// GenerateController 生成控制器
func (g *Generator) GenerateController(ctx context.Context, name string) error {
	fmt.Printf("🔧 Generating controller: %s\n", name)

	data := map[string]interface{}{
		"ControllerName": name,
		"VariableName":   strings.ToLower(name),
		"Date":           time.Now().Format("2006-01-02"),
	}

	outputPath := fmt.Sprintf("src/controllers/%s_controller.shode", strings.ToLower(name))
	if err := g.generateFile("controller", outputPath, data); err != nil {
		return err
	}

	fmt.Printf("  ✓ Generated controller: %s\n", name)
	return nil
}

// GenerateModel 生成模型
func (g *Generator) GenerateModel(ctx context.Context, name string, fields []string) error {
	fmt.Printf("📦 Generating model: %s\n", name)

	data := map[string]interface{}{
		"ModelName": name,
		"TableName": strings.ToLower(name) + "s",
		"Fields":    parseFields(fields),
		"Date":      time.Now().Format("2006-01-02"),
	}

	outputPath := fmt.Sprintf("src/models/%s.shode", strings.ToLower(name))
	if err := g.generateFile("model", outputPath, data); err != nil {
		return err
	}

	fmt.Printf("  ✓ Generated model: %s\n", name)
	return nil
}

// GenerateService 生成服务
func (g *Generator) GenerateService(ctx context.Context, name string) error {
	fmt.Printf("⚙️  Generating service: %s\n", name)

	data := map[string]interface{}{
		"ServiceName": name,
		"VariableName": strings.ToLower(name),
		"Date":        time.Now().Format("2006-01-02"),
	}

	outputPath := fmt.Sprintf("src/services/%s_service.shode", strings.ToLower(name))
	if err := g.generateFile("service", outputPath, data); err != nil {
		return err
	}

	fmt.Printf("  ✓ Generated service: %s\n", name)
	return nil
}

// GenerateTest 生成测试
func (g *Generator) GenerateTest(ctx context.Context, targetFile string) error {
	fmt.Printf("🧪 Generating test for: %s\n", targetFile)

	baseName := strings.TrimSuffix(targetFile, ".shode")
	testName := baseName + "_test.shode"

	data := map[string]interface{}{
		"TargetFile": targetFile,
		"TestName":   testName,
		"Date":       time.Now().Format("2006-01-02"),
	}

	outputPath := filepath.Join("tests", testName)
	if err := g.generateFile("test", outputPath, data); err != nil {
		return err
	}

	fmt.Printf("  ✓ Generated test: %s\n", testName)
	return nil
}

// generateFile 生成文件
func (g *Generator) generateFile(templateName, outputPath string, data map[string]interface{}) error {
	tmpl, exists := g.templates.templates[templateName]
	if !exists {
		return fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if g.dryRun {
		fmt.Printf("  [DRY RUN] Would create: %s\n", outputPath)
		fmt.Printf("  Content:\n%s\n", buf.String())
		return nil
	}

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// parseFields 解析字段定义
func parseFields(fields []string) []Field {
	result := make([]Field, 0, len(fields))

	for _, field := range fields {
		parts := strings.Split(field, ":")
		if len(parts) >= 2 {
			result = append(result, Field{
				Name: parts[0],
				Type: parts[1],
				Tag:  parseTag(parts),
			})
		}
	}

	return result
}

// parseTag 解析标签
func parseTag(parts []string) string {
	if len(parts) > 2 {
		return strings.Join(parts[2:], " ")
	}
	return ""
}

// Field 字段定义
type Field struct {
	Name string
	Type string
	Tag  string
}

// SetDryRun 设置是否为模拟运行
func (g *Generator) SetDryRun(dryRun bool) {
	g.dryRun = dryRun
}

// SetOutputDir 设置输出目录
func (g *Generator) SetOutputDir(dir string) {
	g.outputDir = dir
}

// === 内置模板 ===

const projectTemplate = `# {{.ProjectName}}
# Generated by Shode Generator
# Date: {{.Date}}

import { http, server, logger } from "std"

// 初始化日志
log = logger.new({
    level: "info",
    format: "text"
})

// 主路由
server.get("/", func(req, res) {
    res.json({
        message: "Welcome to {{.ProjectName}}!",
        version: "{{.Version}}"
    })
})

server.get("/health", func(req, res) {
    res.json({
        status: "ok",
        timestamp: timestamp()
    })
})

// 启动服务器
log.info("Starting {{.ProjectName}}...")
log.info("Server listening on http://localhost:8080")

server.listen(8080)
`

const controllerTemplate = `# {{.ControllerName}} Controller
# Generated by Shode Generator
# Date: {{.Date}}

import { http, validator, database } from "std"

// {{.ControllerName}}Controller 控制器
controller {{.ControllerName}}Controller {
    // 索引
    index = func(req, res) {
        // 实现列表查询
        items = database.query("SELECT * FROM {{.VariableName}}s")
        res.json(items)
    }

    // 显示
    show = func(req, res) {
        id = req.params.id
        item = database.query_one("SELECT * FROM {{.VariableName}}s WHERE id = $1", [id])

        if item == null {
            res.status(404).json({"error": "Not found"})
            return
        }

        res.json(item)
    }

    // 创建
    create = func(req, res) {
        data = req.body

        // 验证数据
        errors = validator.validate(data)
        if len(errors) > 0 {
            res.status(400).json({"errors": errors})
            return
        }

        // 创建记录
        id = database.insert(
            "INSERT INTO {{.VariableName}}s (name, created_at) VALUES ($1, $2) RETURNING id",
            [data.name, timestamp()]
        )

        res.status(201).json({
            id: id,
            message: "Created successfully"
        })
    }

    // 更新
    update = func(req, res) {
        id = req.params.id
        data = req.body

        // 更新记录
        database.execute(
            "UPDATE {{.VariableName}}s SET name = $1, updated_at = $2 WHERE id = $3",
            [data.name, timestamp(), id]
        )

        res.json({"message": "Updated successfully"})
    }

    // 删除
    delete = func(req, res) {
        id = req.params.id

        // 删除记录
        database.execute("DELETE FROM {{.VariableName}}s WHERE id = $1", [id])

        res.json({"message": "Deleted successfully"})
    }
}

// 注册路由
http.get("/{{.VariableName}}s", {{.ControllerName}}Controller.index)
http.get("/{{.VariableName}}s/:id", {{.ControllerName}}Controller.show)
http.post("/{{.VariableName}}s", {{.ControllerName}}Controller.create)
http.put("/{{.VariableName}}s/:id", {{.ControllerName}}Controller.update)
http.delete("/{{.VariableName}}s/:id", {{.ControllerName}}Controller.delete)
`

const modelTemplate = `# {{.ModelName}} Model
# Generated by Shode Generator
# Date: {{.Date}}

import { database, validator } from "std"

// {{.ModelName}} 模型定义
model {{.ModelName}} {
    // 表名
    table_name = "{{.TableName}}"

    // 字段定义
    {{range .Fields}}field_{{.Name}} = "{{.Type}}"
    {{end}}

    // 查找所有记录
    find_all = func() {
        return database.query("SELECT * FROM {{.TableName}}")
    }

    // 根据 ID 查找
    find_by_id = func(id) {
        return database.query_one("SELECT * FROM {{.TableName}} WHERE id = $1", [id])
    }

    // 创建记录
    create = func(data) {
        {{range .Fields}}
        if data.{{.Name}} == null {
            throw "{{.Name}} is required"
        }
        {{end}}

        return database.insert(
            "INSERT INTO {{.TableName}} ({{range $i, $f := .Fields}}{{if $i}}, {{end}}{{.Name}}{{end}}) VALUES ({{range $i, $f := .Fields}}{{if $i}}, {{end}}${{{.Name}}}{{end}}) RETURNING *",
            [{{range $i, $f := .Fields}}{{if $i}}, {{end}}data.{{.Name}}{{end}}]
        )
    }

    // 更新记录
    update = func(id, data) {
        return database.execute(
            "UPDATE {{.TableName}} SET {{range $i, $f := .Fields}}{{if $i}}, {{end}}{{.Name}} = ${{add (len .Fields) $i}}{{end}} WHERE id = $1",
            [{{range $i, $f := .Fields}}{{if $i}}, {{end}}data.{{.Name}}{{end}}, id]
        )
    }

    // 删除记录
    delete = func(id) {
        return database.execute("DELETE FROM {{.TableName}} WHERE id = $1", [id])
    }
}

// 创建模型实例
{{.VariableName}} = {{.ModelName}}()
`

const serviceTemplate = `# {{.ServiceName}} Service
# Generated by Shode Generator
# Date: {{.Date}}

import { cache, database, logger } from "std"

// {{.ServiceName}}Service 服务定义
service {{.ServiceName}}Service {
    log = logger.new("{{.ServiceName}}")

    // 获取数据（带缓存）
    get = func(key) {
        // 尝试从缓存获取
        cached = cache.get(key)
        if cached != null {
            return cached
        }

        // 从数据库查询
        data = database.query_one("SELECT * FROM data WHERE key = $1", [key])

        // 缓存结果
        if data != null {
            cache.set(key, data, 3600)
        }

        return data
    }

    // 设置数据
    set = func(key, value) {
        // 更新数据库
        database.execute(
            "INSERT INTO data (key, value, updated_at) VALUES ($1, $2, $3) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = $3",
            [key, value, timestamp()]
        )

        // 清除缓存
        cache.delete(key)

        log.info("Data updated: ${key}")
    }

    // 删除数据
    delete = func(key) {
        // 删除数据库记录
        database.execute("DELETE FROM data WHERE key = $1", [key])

        // 清除缓存
        cache.delete(key)

        log.info("Data deleted: ${key}")
    }

    // 批量操作
    batch = func(items) {
        transaction = database.begin()

        try {
            for item in items {
                transaction.execute(
                    "INSERT INTO data (key, value, created_at) VALUES ($1, $2, $3)",
                    [item.key, item.value, timestamp()]
                )
            }

            transaction.commit()
            log.info("Batch operation completed: ${len(items)} items")
        } catch e {
            transaction.rollback()
            log.error("Batch operation failed: ${e.message}")
            throw e
        }
    }
}

// 创建服务实例
{{.VariableName}}Service = {{.ServiceName}}Service()
`

const testTemplate = `# {{.TestName}}
# Generated by Shode Generator
# Date: {{.Date}}

import { assert, assertEquals, assertContains } from "testing"

// 测试设置
setup = func() {
    // 初始化测试环境
    print("Setting up test environment...")
}

// 测试清理
teardown = func() {
    // 清理测试环境
    print("Cleaning up test environment...")
}

// 测试用例
test("example test", func() {
    setup()

    result = 1 + 1
    assertEquals(result, 2)

    teardown()
})

test("async operation", func() {
    setup()

    // 测试异步操作
    promise = async_operation()
    result = await promise

    assert(result != null)

    teardown()
})

// 测试助手
async_operation = func() {
    return new Promise(func(resolve) {
        timeout(func() {
            resolve("done")
        }, 100)
    })
}
`

const configTemplate = `# {{.ProjectName}} Configuration
# Generated by Shode Generator

// 服务器配置
server {
    host: "0.0.0.0"
    port: 8080
    mode: "development"  // "development" or "production"
}

// 日志配置
logging {
    level: "info"  // "debug", "info", "warn", "error"
    format: "text"  // "text" or "json"
}

// 数据库配置
database {
    driver: "sqlite"  // "sqlite", "postgres", "mysql"
    source: "{{.PackageName}}.db"
}

// 缓存配置
cache {
    enabled: true
    driver: "memory"  // "memory", "redis"
    ttl: 3600  // 秒
}

// 安全配置
security {
    jwt_secret = "change-this-in-production"
    jwt_expire = 24  // 小时
}
`

const readmeTemplate = "# {{.ProjectName}}\n\n" +
	"{{.Description}}\n\n" +
	"## 功能特性\n\n" +
	"- ✅ 功能列表 1\n" +
	"- ✅ 功能列表 2\n" +
	"- ✅ 功能列表 3\n\n" +
	"## 快速开始\n\n" +
	"### 安装\n\n" +
	"```bash\n" +
	"# 克隆仓库\n" +
	"git clone https://github.com/user/{{.ProjectName}}.git\n" +
	"cd {{.ProjectName}}\n\n" +
	"# 安装依赖\n" +
	"shode install\n" +
	"```\n\n" +
	"### 运行\n\n" +
	"```bash\n" +
	"# 开发模式\n" +
	"shode run main.shode\n\n" +
	"# 生产模式\n" +
	"shode build\n" +
	"shode start\n" +
	"```\n\n" +
	"### 测试\n\n" +
	"```bash\n" +
	"# 运行测试\n" +
	"shode test\n\n" +
	"# 覆盖率报告\n" +
	"shode test --cover\n" +
	"```\n\n" +
	"## 项目结构\n\n" +
	"```\n" +
	".\n" +
	"├── main.shode       # 主程序\n" +
	"├── config.shode     # 配置文件\n" +
	"├── src/             # 源代码\n" +
	"│   ├── controllers/ # 控制器\n" +
	"│   ├── models/      # 模型\n" +
	"│   └── services/    # 服务\n" +
	"├── tests/           # 测试\n" +
	"├── docs/            # 文档\n" +
	"└── README.md        # 说明\n" +
	"```\n\n" +
	"## API 文档\n\n" +
	"### 端点列表\n\n" +
	"| 方法 | 端点 | 描述 |\n" +
	"|------|------|------|\n" +
	"| GET | / | 首页 |\n" +
	"| GET | /health | 健康检查 |\n\n" +
	"## 配置说明\n\n" +
	"详细配置说明请参考 [配置文档](docs/config.md)。\n\n" +
	"## 开发指南\n\n" +
	"详细开发指南请参考 [开发文档](docs/development.md)。\n\n" +
	"## 贡献指南\n\n" +
	"欢迎贡献代码！请阅读 [贡献指南](CONTRIBUTING.md)。\n\n" +
	"## 许可证\n\n" +
	"{{.License}}\n\n" +
	"## 作者\n\n" +
	"{{.Author}}\n\n" +
	"---\n\n" +
	"*Generated by Shode Generator on {{.Date}}*\n"

