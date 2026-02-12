// Package templateutil 提供模板处理工具
package templateutil

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// Template 模板接口
type Template interface {
	Execute(data any) (string, error)
	ExecuteToFile(data any, filename string) error
}

// TextTemplate 文本模板
type TextTemplate struct {
	tmpl *template.Template
}

// NewTextTemplate 创建文本模板
func NewTextTemplate(text string) (*TextTemplate, error) {
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return nil, err
	}
	return &TextTemplate{tmpl: tmpl}, nil
}

// NewTextTemplateWithFuncs 创建带函数的文本模板
func NewTextTemplateWithFuncs(text string, funcs template.FuncMap) (*TextTemplate, error) {
	tmpl, err := template.New("template").Funcs(funcs).Parse(text)
	if err != nil {
		return nil, err
	}
	return &TextTemplate{tmpl: tmpl}, nil
}

// Execute 执行模板
func (t *TextTemplate) Execute(data any) (string, error) {
	var buf bytes.Buffer
	err := t.tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ExecuteToFile 执行模板并写入文件
func (t *TextTemplate) ExecuteToFile(data any, filename string) error {
	result, err := t.Execute(data)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, []byte(result), 0644)
}

// Render 渲染模板
func Render(text string, data any) (string, error) {
	tmpl, err := template.New("render").Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderString 渲染简单字符串模板
func RenderString(format string, data map[string]any) (string, error) {
	result := format
	for key, value := range data {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result, nil
}

// RenderFile 渲染模板文件
func RenderFile(filename string, data any) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(filename).Parse(string(content))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderFileToFile 渲染模板文件到输出文件
func RenderFileToFile(templateFile, outputFile string, data any) error {
	content, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}

	tmpl, err := template.New(templateFile).Parse(string(content))
	if err != nil {
		return err
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return tmpl.Execute(outFile, data)
}

// MustRender 必须成功渲染，否则panic
func MustRender(text string, data any) string {
	result, err := Render(text, data)
	if err != nil {
		panic(err)
	}
	return result
}

// MustRenderFile 必须成功渲染文件，否则panic
func MustRenderFile(filename string, data any) string {
	result, err := RenderFile(filename, data)
	if err != nil {
		panic(err)
	}
	return result
}

// DefaultFuncs 默认模板函数
var DefaultFuncs = template.FuncMap{
	"toUpper": strings.ToUpper,
	"toLower": strings.ToLower,
	"title":   strings.Title,
	"trim":    strings.TrimSpace,
	"repeat":  strings.Repeat,
	"join":    strings.Join,
	"contains": strings.Contains,
	"hasPrefix": strings.HasPrefix,
	"hasSuffix": strings.HasSuffix,
	"replace":  strings.ReplaceAll,
	"split":    strings.Split,
	"length": func(v any) int {
		switch val := v.(type) {
		case string:
			return len(val)
		case []string:
			return len(val)
		case []any:
			return len(val)
		default:
			return 0
		}
	},
	"format": func(format string, a ...any) string {
		return fmt.Sprintf(format, a...)
	},
}

// NewTemplateWithDefaults 创建带默认函数的模板
func NewTemplateWithDefaults(text string) (*TextTemplate, error) {
	return NewTextTemplateWithFuncs(text, DefaultFuncs)
}

// TemplateBuilder 模板构建器
type TemplateBuilder struct {
	text    string
	funcs   template.FuncMap
	options []TemplateOption
}

// TemplateOption 模板选项
type TemplateOption func(*template.Template)

// NewTemplateBuilder 创建模板构建器
func NewTemplateBuilder() *TemplateBuilder {
	return &TemplateBuilder{
		funcs: make(template.FuncMap),
	}
}

// WithText 设置模板文本
func (b *TemplateBuilder) WithText(text string) *TemplateBuilder {
	b.text = text
	return b
}

// WithFunc 添加函数
func (b *TemplateBuilder) WithFunc(name string, fn any) *TemplateBuilder {
	b.funcs[name] = fn
	return b
}

// WithFuncs 添加多个函数
func (b *TemplateBuilder) WithFuncs(funcs template.FuncMap) *TemplateBuilder {
	for k, v := range funcs {
		b.funcs[k] = v
	}
	return b
}

// WithDefaults 使用默认函数
func (b *TemplateBuilder) WithDefaults() *TemplateBuilder {
	for k, v := range DefaultFuncs {
		b.funcs[k] = v
	}
	return b
}

// Build 构建模板
func (b *TemplateBuilder) Build() (*TextTemplate, error) {
	if b.text == "" {
		return nil, fmt.Errorf("template text is empty")
	}

	tmpl, err := template.New("template").Funcs(b.funcs).Parse(b.text)
	if err != nil {
		return nil, err
	}

	return &TextTemplate{tmpl: tmpl}, nil
}

// MustBuild 必须成功构建，否则panic
func (b *TemplateBuilder) MustBuild() *TextTemplate {
	tmpl, err := b.Build()
	if err != nil {
		panic(err)
	}
	return tmpl
}

// Include 包含其他模板
func Include(name string, data map[string]any) string {
	// 简化实现，实际应该支持从文件或注册表中查找
	return fmt.Sprintf("{{include %s}}", name)
}

// Block 定义块
func Block(name string, content string) string {
	return fmt.Sprintf("{{block %s}}%s{{end}}", name, content)
}

// Define 定义模板
func Define(name string, content string) string {
	return fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, content)
}

// If 条件语句
func If(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

// IfEmpty 如果为空则返回默认值
func IfEmpty(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// Select 多选一
func Select(index int, values ...string) string {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return ""
}

// Range 生成范围
func Range(start, end int) []int {
	result := make([]int, end-start)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// Loop 循环
func Loop(count int) []int {
	result := make([]int, count)
	for i := range result {
		result[i] = i
	}
	return result
}

// Add 加法
func Add(a, b int) int {
	return a + b
}

// Sub 减法
func Sub(a, b int) int {
	return a - b
}

// Mul 乘法
func Mul(a, b int) int {
	return a * b
}

// Div 除法
func Div(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

// Mod 取模
func Mod(a, b int) int {
	if b == 0 {
		return 0
	}
	return a % b
}

// Eq 等于
func Eq(a, b any) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// Ne 不等于
func Ne(a, b any) bool {
	return !Eq(a, b)
}

// Lt 小于
func Lt(a, b int) bool {
	return a < b
}

// Le 小于等于
func Le(a, b int) bool {
	return a <= b
}

// Gt 大于
func Gt(a, b int) bool {
	return a > b
}

// Ge 大于等于
func Ge(a, b int) bool {
	return a >= b
}

// And 逻辑与
func And(a, b bool) bool {
	return a && b
}

// Or 逻辑或
func Or(a, b bool) bool {
	return a || b
}

// Not 逻辑非
func Not(a bool) bool {
	return !a
}

// Coalesce 返回第一个非空值
func Coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// First 返回第一个元素
func First[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[0]
}

// Last 返回最后一个元素
func Last[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[len(slice)-1]
}

// Slice 切片
func Slice[T any](slice []T, start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(slice) {
		end = len(slice)
	}
	if start >= end {
		return []T{}
	}
	return slice[start:end]
}

// Dict 创建字典
func Dict(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, fmt.Errorf("pairs must be even")
	}

	result := make(map[string]any)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("key must be string")
		}
		result[key] = pairs[i+1]
	}

	return result, nil
}

// List 创建列表
func List(items ...any) []any {
	return items
}

// SafeString 安全字符串
type SafeString string

// HTML 安全HTML
type HTML string

// HTMLAttr 安全HTML属性
type HTMLAttr string

// CSS 安全CSS
type CSS string

// URL 安全URL
type URL string

// JS 安全JavaScript
type JS string

// Safe 安全内容
type Safe string

// MarshalJSON 实现JSON序列化
func (s SafeString) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s)), nil
}

// MarshalJSON 实现JSON序列化
func (h HTML) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, h)), nil
}

// EscapeHTML 转义HTML
func EscapeHTML(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// UnescapeHTML 反转义HTML
func UnescapeHTML(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// TruncateHTML 截断HTML（保留标签完整性）
func TruncateHTML(html string, maxLen int) string {
	if len(html) <= maxLen {
		return html
	}

	// 简化实现，实际应该解析HTML并保留标签完整性
	return html[:maxLen] + "..."
}

// StripHTMLTags 移除HTML标签
func StripHTMLTags(html string) string {
	// 简化实现，实际应该使用HTML解析器
	result := html
	inTag := false
	var buf strings.Builder

	for _, r := range result {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}

	return buf.String()
}

// MinifyHTML 压缩HTML
func MinifyHTML(html string) string {
	// 移除多余空白
	lines := strings.Split(html, "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return strings.Join(result, "")
}

// FormatJSON 格式化JSON
func FormatJSON(json string) (string, error) {
	// 简化实现，实际应该解析JSON并格式化
	return json, nil
}

// MinifyJSON 压缩JSON
func MinifyJSON(json string) (string, error) {
	// 简化实现，实际应该解析JSON并压缩
	return strings.ReplaceAll(json, " ", ""), nil
}

// TemplateSet 模板集合
type TemplateSet struct {
	templates map[string]*TextTemplate
	funcs     template.FuncMap
}

// NewTemplateSet 创建模板集合
func NewTemplateSet() *TemplateSet {
	return &TemplateSet{
		templates: make(map[string]*TextTemplate),
		funcs:     make(template.FuncMap),
	}
}

// Add 添加模板
func (ts *TemplateSet) Add(name, text string) error {
	tmpl, err := NewTextTemplateWithFuncs(text, ts.funcs)
	if err != nil {
		return err
	}
	ts.templates[name] = tmpl
	return nil
}

// Get 获取模板
func (ts *TemplateSet) Get(name string) (*TextTemplate, bool) {
	tmpl, exists := ts.templates[name]
	return tmpl, exists
}

// Execute 执行指定模板
func (ts *TemplateSet) Execute(name string, data any) (string, error) {
	tmpl, exists := ts.Get(name)
	if !exists {
		return "", fmt.Errorf("template not found: %s", name)
	}
	return tmpl.Execute(data)
}

// AddFunc 添加全局函数
func (ts *TemplateSet) AddFunc(name string, fn any) {
	ts.funcs[name] = fn
}

// AddFuncs 添加多个全局函数
func (ts *TemplateSet) AddFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		ts.funcs[k] = v
	}
}

// LoadFromDir 从目录加载模板
func (ts *TemplateSet) LoadFromDir(dir string, ext string) error {
	// 简化实现，实际应该遍历目录并加载所有文件
	return nil
}

// Inheritance 模板继承
type Inheritance struct {
	layout   string
	templates map[string]*template.Template
}

// NewInheritance 创建模板继承
func NewInheritance(layout string) *Inheritance {
	return &Inheritance{
		layout:    layout,
		templates: make(map[string]*template.Template),
	}
}

// Extend 扩展模板
func (inh *Inheritance) Extend(name string) (string, error) {
	// 简化实现，实际应该支持模板继承
	return name, nil
}

// Block 定义块
func (inh *Inheritance) Block(name string, content string) string {
	return fmt.Sprintf("{{block \"%s\"}}%s{{end}}", name, content)
}

// Super 调用父模板块
func (inh *Inheritance) Super(name string) string {
	return fmt.Sprintf("{{template \"%s\"}}", name)
}
