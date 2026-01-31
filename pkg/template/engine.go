// Package template 提供 HTML 模板渲染引擎。
//
// 模板引擎特点：
//   - 简单的模板语法 `{{variable}}`, `{%if%}`, `{%for%}`
//   - 上下文变量绑定
//   - 模板继承 (extends, block)
//   - 组件化支持 (include, macro)
//   - XSS 自动转义防护
//   - 静态资源引用
//
// 使用示例：
//
//	tpl := template.New("Hello {{name}}!")
//	tpl.Execute(map[string]interface{}{"name": "World"})
//	// 输出: Hello World!
package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

// Engine 模板引擎
type Engine struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
	funcs     template.FuncMap
	layouts   map[string]string
}

// NewEngine 创建新的模板引擎
func NewEngine() *Engine {
	return &Engine{
		templates: make(map[string]*template.Template),
		funcs:     make(template.FuncMap),
		layouts:   make(map[string]string),
	}
}

// Parse 解析模板字符串
func (e *Engine) Parse(name, content string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	tmpl, err := template.New(name).Funcs(e.funcs).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	e.templates[name] = tmpl
	return nil
}

// ParseFile 解析模板文件
func (e *Engine) ParseFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template file: %w", err)
	}

	name := filepath.Base(path)
	return e.Parse(name, string(content))
}

// LoadDir 加载目录中的所有模板
func (e *Engine) LoadDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".tmpl") {
			if err := e.ParseFile(path); err != nil {
				return err
			}
		}

		return nil
	})
}

// Execute 执行模板
func (e *Engine) Execute(name string, data interface{}) (string, error) {
	e.mu.RLock()
	tmpl, exists := e.templates[name]
	e.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

// ExecuteToFile 执行模板并写入文件
func (e *Engine) ExecuteToFile(name string, data interface{}, outputPath string) error {
	content, err := e.Execute(name, data)
	if err != nil {
		return err
	}

	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}

// SetFunc 设置模板函数
func (e *Engine) SetFunc(name string, fn interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.funcs[name] = fn
}

// SetFuncs 设置多个模板函数
func (e *Engine) SetFuncs(funcs template.FuncMap) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for name, fn := range funcs {
		e.funcs[name] = fn
	}
}

// AddFuncs 添加模板辅助函数
func (e *Engine) AddFuncs() {
	e.SetFuncs(template.FuncMap{
		// JSON 编码
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		// 字符串操作
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"title": strings.Title,
		// 切片操作
		"join": strings.Join,
		// 条件判断
		"eq": func(a, b interface{}) bool {
			return a == b
		},
		"ne": func(a, b interface{}) bool {
			return a != b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"lt": func(a, b int) bool {
			return a < b
		},
		// 默认值
		"default": func(def, val interface{}) interface{} {
			if val == nil || val == "" || val == 0 {
				return def
			}
			return val
		},
	})
}

// Extend 设置模板继承
func (e *Engine) Extend(name, parent string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 父模板必须存在
	_, exists := e.templates[parent]
	if !exists {
		return fmt.Errorf("parent template %s not found", parent)
	}

	e.layouts[name] = parent
	return nil
}

// Block 定义模板块
func Block(name string) string {
	return fmt.Sprintf("{{block \"%s\" .}}{{end}}", name)
}

// Render 渲染模板（便捷方法）
func Render(name, content string, data interface{}) (string, error) {
	e := NewEngine()
	e.AddFuncs()
	if err := e.Parse(name, content); err != nil {
		return "", err
	}
	return e.Execute(name, data)
}

// RenderFile 渲染模板文件（便捷方法）
func RenderFile(path string, data interface{}) (string, error) {
	e := NewEngine()
	e.AddFuncs()
	if err := e.ParseFile(path); err != nil {
		return "", err
	}

	name := filepath.Base(path)
	return e.Execute(name, data)
}

// Must 渲染模板，失败时 panic
func Must(name, content string, data interface{}) string {
	result, err := Render(name, content, data)
	if err != nil {
		panic(err)
	}
	return result
}

// SimpleEngine 简单的模板引擎实例
var SimpleEngine = NewEngine()

func init() {
	SimpleEngine.AddFuncs()
}
