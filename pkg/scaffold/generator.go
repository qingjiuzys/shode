package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generator 项目生成器
type Generator struct {
	engine *Engine
}

// NewGenerator 创建新的生成器
func NewGenerator() *Generator {
	return &Generator{
		engine: NewEngine(""),
	}
}

// Generate 生成项目
func (g *Generator) Generate(projectName, templateType string, options map[string]string) error {
	// 格式化项目名称
	projectName = FormatProjectName(projectName)

	// 验证模板类型
	if !ValidateTemplateName(templateType) {
		return fmt.Errorf("无效的模板类型: %s (可用: basic, web-service, cli-tool)", templateType)
	}

	// 检查目标目录
	targetDir := filepath.Join(".", projectName)
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("目录已存在: %s", targetDir)
	}

	// 准备模板变量
	variables := g.prepareVariables(projectName, options)

	// 生成项目
	fmt.Printf("🚀 创建项目: %s\n", projectName)
	fmt.Printf("📦 模板类型: %s\n", templateType)
	fmt.Println()

	if err := g.engine.Generate(templateType, variables, targetDir); err != nil {
		return fmt.Errorf("生成项目失败: %w", err)
	}

	fmt.Println("✅ 项目创建成功！")
	fmt.Println()
	fmt.Println("下一步:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  shode pkg install")
	fmt.Println("  shode pkg run start")

	return nil
}

// prepareVariables 准备模板变量
func (g *Generator) prepareVariables(projectName string, options map[string]string) map[string]string {
	variables := make(map[string]string)

	// 基础变量
	variables["Name"] = projectName
	variables["Version"] = options["version"]
	if variables["Version"] == "" {
		variables["Version"] = "1.0.0"
	}

	variables["Description"] = options["description"]
	if variables["Description"] == "" {
		variables["Description"] = "A Shode project"
	}

	variables["Port"] = options["port"]
	if variables["Port"] == "" {
		variables["Port"] = "8080"
	}

	// 添加自定义选项
	for key, value := range options {
		if _, exists := variables[key]; !exists {
			variables[key] = value
		}
	}

	return variables
}

// ListTemplates 列出所有可用模板
func (g *Generator) ListTemplates() []TemplateInfo {
	templates := g.engine.ListTemplates()

	infos := make([]TemplateInfo, 0, len(templates))
	descriptions := map[string]string{
		"basic":       "基础 Shode 项目 - 适合简单的脚本工具",
		"web-service": "Web 服务项目 - 包含 HTTP 服务和配置管理",
		"cli-tool":    "CLI 工具项目 - 适合命令行工具开发",
	}

	for _, tmpl := range templates {
		infos = append(infos, TemplateInfo{
			Name:        tmpl,
			Description: descriptions[tmpl],
		})
	}

	return infos
}

// TemplateInfo 模板信息
type TemplateInfo struct {
	Name        string
	Description string
}

// FormatProjectName 格式化项目名称
func FormatProjectName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, "_", "-")
	return name
}

// ValidateProjectName 验证项目名称
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("项目名称不能为空")
	}

	if strings.ContainsAny(name, "/\\<>:\"|?*") {
		return fmt.Errorf("项目名称包含非法字符")
	}

	return nil
}
