package scaffold

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Template 表示一个项目模板
type Template struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Variables   []TemplateVariable `json:"variables"`
	Files       []TemplateFile    `json:"files"`
}

// TemplateVariable 表示模板变量
type TemplateVariable struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Default      string `json:"default"`
	Required     bool   `json:"required"`
}

// TemplateFile 表示模板文件
type TemplateFile struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Executable bool `json:"executable"`
}

// Engine 模板引擎
type Engine struct {
	templatesDir string
}

// NewEngine 创建新的模板引擎
func NewEngine(templatesDir string) *Engine {
	return &Engine{
		templatesDir: templatesDir,
	}
}

// LoadTemplate 加载模板
func (e *Engine) LoadTemplate(name string) (*Template, error) {
	// 首先检查内置模板
	switch name {
	case "basic":
		return e.getBasicTemplate(), nil
	case "web-service":
		return e.getWebServiceTemplate(), nil
	case "cli-tool":
		return e.getCLIToolTemplate(), nil
	default:
		// 尝试从文件系统加载
		templateDir := filepath.Join(e.templatesDir, name)

		// 检查模板目录是否存在
		if _, err := os.Stat(templateDir); os.IsNotExist(err) {
			return nil, fmt.Errorf("模板不存在: %s", name)
		}

		// TODO: 从 template.json 加载模板配置
		return nil, fmt.Errorf("从文件系统加载模板尚未实现: %s", name)
	}
}

// ListTemplates 列出所有可用模板
func (e *Engine) ListTemplates() []string {
	return []string{"basic", "web-service", "cli-tool"}
}

// Generate 生成项目
func (e *Engine) Generate(templateName string, variables map[string]string, targetDir string) error {
	// 加载模板
	tmpl, err := e.LoadTemplate(templateName)
	if err != nil {
		return err
	}

	// 验证必需变量
	for _, v := range tmpl.Variables {
		if v.Required {
			if value, exists := variables[v.Name]; !exists || value == "" {
				if v.Default != "" {
					variables[v.Name] = v.Default
				} else {
					return fmt.Errorf("缺少必需变量: %s", v.Name)
				}
			}
		}
	}

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 生成文件
	for _, file := range tmpl.Files {
		if err := e.generateFile(file, variables, targetDir); err != nil {
			return err
		}
	}

	return nil
}

// generateFile 生成单个文件
func (e *Engine) generateFile(file TemplateFile, variables map[string]string, targetDir string) error {
	// 解析文件路径
	filePath := filepath.Join(targetDir, e.parseTemplate(file.Path, variables))

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 解析文件内容
	content := e.parseTemplate(file.Content, variables)

	// 写入文件
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	// 设置可执行权限
	if file.Executable {
		if err := os.Chmod(filePath, 0755); err != nil {
			return fmt.Errorf("设置可执行权限失败: %w", err)
		}
	}

	return nil
}

// parseTemplate 解析模板内容
func (e *Engine) parseTemplate(content string, variables map[string]string) string {
	tmpl, err := template.New("content").Parse(content)
	if err != nil {
		return content // 如果解析失败，返回原始内容
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return content
	}

	return buf.String()
}

// getBasicTemplate 获取基础模板
func (e *Engine) getBasicTemplate() *Template {
	return &Template{
		Name:        "basic",
		Description: "基础 Shode 项目",
		Variables: []TemplateVariable{
			{Name: "Name", Description: "项目名称", Default: "my-shode-project", Required: true},
			{Name: "Version", Description: "版本号", Default: "1.0.0", Required: true},
			{Name: "Description", Description: "项目描述", Default: "A basic Shode project", Required: false},
		},
		Files: []TemplateFile{
			{
				Path:    "shode.json",
				Content: `{
  "name": "{{.Name}}",
  "version": "{{.Version}}",
  "description": "{{.Description}}",
  "main": "main.sh",
  "scripts": {
    "start": "shode run main.sh",
    "test": "shode test"
  },
  "dependencies": {},
  "author": "",
  "license": "MIT"
}`,
			},
			{
				Path:        "main.sh",
				Content:     `#!/bin/sh
# {{.Name}} - {{.Description}}

echo "Hello from {{.Name}} v{{.Version}}!"
`,
				Executable: true,
			},
			{
				Path:    "README.md",
				Content: "# {{.Name}}\n\n{{.Description}}\n\n## 快速开始\n\n```bash\n# 安装依赖\nshode pkg install\n\n# 运行\nshode pkg run start\n```\n\n## 许可证\n\nMIT\n",
			},
		},
	}
}

// getWebServiceTemplate 获取 Web 服务模板
func (e *Engine) getWebServiceTemplate() *Template {
	return &Template{
		Name:        "web-service",
		Description: "Web 服务项目",
		Variables: []TemplateVariable{
			{Name: "Name", Description: "项目名称", Default: "my-web-service", Required: true},
			{Name: "Version", Description: "版本号", Default: "1.0.0", Required: true},
			{Name: "Port", Description: "服务端口", Default: "8080", Required: false},
		},
		Files: []TemplateFile{
			{
				Path:    "shode.json",
				Content: `{
  "name": "{{.Name}}",
  "version": "{{.Version}}",
  "description": "A Shode web service",
  "main": "src/main.sh",
  "scripts": {
    "start": "shode run src/main.sh",
    "dev": "shode run src/main.sh --watch",
    "test": "shode test"
  },
  "dependencies": {
    "@shode/logger": "^1.2.0",
    "@shode/http": "^1.0.0",
    "@shode/config": "^1.0.0"
  },
  "author": "",
  "license": "MIT"
}`,
			},
			{
				Path:        "src/main.sh",
				Content:     `#!/bin/sh
# {{.Name}} Web Service
# 端口: {{.Port}}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 加载依赖
. "$PROJECT_ROOT/sh_modules/@shode/logger/index.sh"
. "$PROJECT_ROOT/sh_modules/@shode/config/index.sh"

# 初始化日志
SetLogLevel "info"

# 启动服务
LogInfo "启动 {{.Name}} v{{.Version}}..."
LogInfo "监听端口: {{.Port}}"

# TODO: 实现 HTTP 服务

LogInfo "服务已启动"
`,
				Executable: true,
			},
			{
				Path:    "config/app.json",
				Content: `{
  "server": {
    "port": {{.Port}},
    "host": "localhost"
  },
  "logging": {
    "level": "info"
  }
}`,
			},
			{
				Path:    "README.md",
				Content: "# {{.Name}}\n\nWeb service built with Shode.\n\n## 快速开始\n\n```bash\n# 安装依赖\nshode pkg install\n\n# 运行服务\nshode pkg run start\n```\n\n## 配置\n\n编辑 `config/app.json` 修改配置。\n\n## 许可证\n\nMIT\n",
			},
		},
	}
}

// getCLIToolTemplate 获取 CLI 工具模板
func (e *Engine) getCLIToolTemplate() *Template {
	return &Template{
		Name:        "cli-tool",
		Description: "CLI 工具项目",
		Variables: []TemplateVariable{
			{Name: "Name", Description: "工具名称", Default: "my-cli-tool", Required: true},
			{Name: "Version", Description: "版本号", Default: "1.0.0", Required: true},
		},
		Files: []TemplateFile{
			{
				Path:    "shode.json",
				Content: `{
  "name": "{{.Name}}",
  "version": "{{.Version}}",
  "description": "A CLI tool built with Shode",
  "main": "bin/{{.Name}}",
  "scripts": {
    "install": "cp src/main.sh bin/{{.Name}}",
    "start": "bin/{{.Name}}",
    "test": "shode test"
  },
  "dependencies": {
    "@shode/logger": "^1.2.0"
  },
  "bin": {
    "{{.Name}}": "bin/{{.Name}}"
  },
  "author": "",
  "license": "MIT"
}`,
			},
			{
				Path:        "src/main.sh",
				Content:     `#!/bin/sh
# {{.Name}} v{{.Version}}

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 加载依赖
. "$PROJECT_ROOT/sh_modules/@shode/logger/index.sh"

# 显示帮助
show_help() {
    echo "Usage: {{.Name}} [command] [options]"
    echo ""
    echo "Commands:"
    echo "  help     显示帮助信息"
    echo "  version  显示版本号"
    echo ""
}

# 主函数
main() {
    case "${1:-help}" in
        help)
            show_help
            ;;
        version)
            echo "{{.Name}} v{{.Version}}"
            ;;
        *)
            echo "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
`,
				Executable: true,
			},
			{
				Path:    "README.md",
				Content: "# {{.Name}}\n\nCLI tool built with Shode.\n\n## 安装\n\n```bash\n# 安装依赖\nshode pkg install\n\n# 创建可执行文件\nshode pkg run install\n```\n\n## 使用\n\n```bash\n# 运行工具\n./bin/{{.Name}} help\n```\n\n## 许可证\n\nMIT\n",
			},
		},
	}
}

// GetTemplateByName 根据名称获取模板（便捷方法）
func GetTemplateByName(name string) (*Template, error) {
	engine := NewEngine("")
	return engine.LoadTemplate(name)
}

// ValidateTemplateName 验证模板名称
func ValidateTemplateName(name string) bool {
	validNames := []string{"basic", "web-service", "cli-tool"}
	for _, valid := range validNames {
		if name == valid {
			return true
		}
	}
	return false
}

// FormatTemplateName 格式化模板名称
func FormatTemplateName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}
