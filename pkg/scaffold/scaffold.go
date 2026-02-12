package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Scaffold 脚手架
type Scaffold struct {
	name       string
	templates  map[string]*template.Template
	embedFS    embed.FS
}

// NewScaffold 创建脚手架
func NewScaffold(name string) *Scaffold {
	return &Scaffold{
		name:      name,
		templates: make(map[string]*template.Template),
	}
}

// AddTemplate 添加模板
func (s *Scaffold) AddTemplate(name string, tmpl *template.Template) {
	s.templates[name] = tmpl
}

// Generate 生成项目
func (s *Scaffold) Generate(destDir string, data interface{}) error {
	// 创建目标目录
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	// 生成文件
	for name, tmpl := range s.templates {
		destPath := filepath.Join(destDir, name)

		// 确保父目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		// 执行模板
		file, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer file.Close()

		if err := tmpl.Execute(file, data); err != nil {
			return err
		}
	}

	return nil
}

// GenerateFromEmbed 从嵌入的文件系统生成
func (s *Scaffold) GenerateFromEmbed(embedFS embed.FS, destDir string, data interface{}) error {
	return fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 读取模板文件
		content, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}

		// 处理模板
		tmpl, err := template.New(path).Parse(string(content))
		if err != nil {
			return err
		}

		// 创建目标文件
		destPath := filepath.Join(destDir, path)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		file, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer file.Close()

		return tmpl.Execute(file, data)
	})
}

// ProjectData 项目数据
type ProjectData struct {
	Name        string
	Description string
	Author      string
	Version     string
	License     string
}

// CreateProject 创建项目脚手架
func CreateProject(projectType, name, destDir string) error {
	s := NewScaffold(projectType)
	data := ProjectData{
		Name:        name,
		Description: "",
		Author:      "",
		Version:     "1.0.0",
		License:     "MIT",
	}

	return s.Generate(destDir, data)
}

// IsValidName 检查项目名是否有效
func IsValidName(name string) bool {
	if name == "" {
		return false
	}

	// 检查非法字符
	for _, c := range name {
		if !isAlphaNumeric(c) && c != '-' && c != '_' {
			return false
		}
	}

	return true
}

func isAlphaNumeric(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9')
}

// ToPackageName 将项目名转换为 Go 包名
func ToPackageName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}

// ValidateProjectName 验证项目名
func ValidateProjectName(name string) error {
	if !IsValidName(name) {
		return fmt.Errorf("invalid project name: %s", name)
	}
	return nil
}

// Generator 代码生成器
type Generator struct {
	scaffold *Scaffold
}

// NewGenerator 创建生成器
func NewGenerator(projectType string) *Generator {
	return &Generator{
		scaffold: NewScaffold(projectType),
	}
}

// Generate 生成项目
func (g *Generator) Generate(destDir string, data interface{}) error {
	return g.scaffold.Generate(destDir, data)
}

// ListTemplates 列出模板
func (g *Generator) ListTemplates() []string {
	names := make([]string, 0)
	for name := range g.scaffold.templates {
		names = append(names, name)
	}
	return names
}
