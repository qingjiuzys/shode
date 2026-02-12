// Package clientgen 提供API客户端代码生成功能
package clientgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Generator 客户端生成器
type Generator struct {
	Spec      *OpenAPISpec
	OutputDir string
	Package   string
	ClientName string
}

// OpenAPISpec OpenAPI规范
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Servers    []Server               `json:"servers"`
	Paths      map[string]PathItem    `json:"paths"`
	Components map[string]interface{} `json:"components"`
}

// Info API信息
type Info struct {
	Title          string `json:"title"`
	Version        string `json:"version"`
	Description    string `json:"description"`
	TermsOfService string `json:"termsOfService"`
	Contact        *Contact `json:"contact"`
	License        *License `json:"license"`
}

// Contact 联系信息
type Contact struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Email string `json:"email"`
}

// License 许可证
type License struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Server 服务器
type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
	Variables   map[string]Variable `json:"variables"`
}

// Variable 服务器变量
type Variable struct {
	Default     interface{}            `json:"default"`
	Description string                 `json:"description"`
	Enum        []interface{}          `json:"enum"`
}

// PathItem 路径项
type PathItem struct {
	Ref    string                 `json:"$ref"`
	Summary string                 `json:"summary"`
	Description string              `json:"description"`
	Get    *Operation              `json:"get"`
	Put    *Operation              `json:"put"`
	Post   *Operation              `json:"post"`
	Delete *Operation              `json:"delete"`
	Options *Operation             `json:"options"`
	Head   *Operation              `json:"head"`
	Patch  *Operation              `json:"patch"`
	Parameters []Parameter         `json:"parameters"`
}

// Operation 操作
type Operation struct {
	Tags        []string            `json:"tags"`
	Summary     string              `json:"summary"`
	Description string              `json:"description"`
	OperationID string              `json:"operationId"`
	Parameters  []Parameter         `json:"parameters"`
	RequestBody *RequestBody        `json:"requestBody"`
	Responses   map[string]Response `json:"responses"`
}

// Parameter 参数
type Parameter struct {
	Name            string      `json:"name"`
	In              string      `json:"in"`
	Description     string      `json:"description"`
	Required        bool        `json:"required"`
	Schema          *Schema     `json:"schema"`
	Content         *Content    `json:"content"`
	AllowEmptyValue bool        `json:"allowEmptyValue"`
}

// RequestBody 请求体
type RequestBody struct {
	Description string             `json:"description"`
	Required    bool               `json:"required"`
	Content     map[string]Content `json:"content"`
}

// Response 响应
type Response struct {
	Description string              `json:"description"`
	Headers     map[string]Header  `json:"headers"`
	Content     map[string]Content `json:"content"`
}

// Content 内容
type Content struct {
	Schema *Schema `json:"schema"`
	Example interface{} `json:"example"`
}

// Header 头
type Header struct {
	Description string  `json:"description"`
	Schema      *Schema `json:"schema"`
	Required    bool    `json:"required"`
}

// Schema Schema
type Schema struct {
	Type                 string             `json:"type"`
	Format              string             `json:"format"`
	Description         string             `json:"description"`
	Ref                  string             `json:"$ref"`
	Properties           map[string]*Schema `json:"properties"`
	Required             []string           `json:"required"`
	Items                *Schema            `json:"items"`
	AdditionalProperties *Schema            `json:"additionalProperties"`
}

// NewGenerator 创建生成器
func NewGenerator() *Generator {
	return &Generator{
		Spec: &OpenAPISpec{},
	}
}

// LoadSpec 加载OpenAPI规范
func (g *Generator) LoadSpec(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}

	if err := json.Unmarshal(data, g.Spec); err != nil {
		return fmt.Errorf("failed to parse spec: %w", err)
	}

	return nil
}

// Generate 生成客户端代码
func (g *Generator) Generate() error {
	if g.Package == "" {
		g.Package = "client"
	}
	if g.ClientName == "" {
		g.ClientName = "APIClient"
	}

	// 生成客户端
	if err := g.generateClient(); err != nil {
		return err
	}

	// 生成API接口
	if err := g.generateAPI(); err != nil {
		return err
	}

	// 生成数据模型
	if err := g.generateModels(); err != nil {
		return err
	}

	// 生成配置
	if err := g.generateConfig(); err != nil {
		return err
	}

	return nil
}

// generateClient 生成客户端
func (g *Generator) generateClient() error {
	tmpl := `package {{.Package}}

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Config 客户端配置
type Config struct {
	BaseURL    string
	HTTPClient *http.Client
	APIKey     string
	Debug      bool
}

// Client {{.Spec.Info.Title}} API客户端
type Client struct {
	config *Config
}

// NewClient 创建客户端
func NewClient(config *Config) *Client {
	if config == nil {
		config = &Config{
			BaseURL: "{{.ServerURL}}",
			HTTPClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		}
	}

	return &Client{
		config: config,
	}
}

// doRequest 执行HTTP请求
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, headers map[string]string) (*http.Response, error) {
	// 构建URL
	u, err := url.Parse(c.config.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// 准备请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer " + c.config.APIKey)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// 调试输出
	if c.config.Debug {
		fmt.Printf("[DEBUG] %s %s\n", method, u.String())
	}

	// 发送请求
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}
`

	return g.executeTemplate("client.go", tmpl, nil)
}

// generateAPI 生成API接口
func (g *Generator) generateAPI() error {
	var methods strings.Builder

	for path, pathItem := range g.Spec.Paths {
		operations := []*struct {
			Method    string
			Operation *Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
		}

		for _, op := range operations {
			if op.Operation == nil {
				continue
			}

			methodName := g.generateMethodName(op.Operation.OperationID, op.Method, path)
			methods.WriteString(g.generateMethod(methodName, op.Method, path, op.Operation))
		}
	}

	tmpl := `package {{.Package}}

// API方法
{{.Methods}}
`

	data := map[string]interface{}{
		"Package": g.Package,
		"Methods": methods.String(),
	}

	return g.executeTemplate("api.go", tmpl, data)
}

// generateMethod 生成方法代码
func (g *Generator) generateMethod(name, method, path string, op *Operation) string {
	var builder strings.Builder

	// 方法签名
	builder.WriteString(fmt.Sprintf("// %s %s\n", name, op.Summary))
	builder.WriteString(fmt.Sprintf("func (c *Client) %s(ctx context.Context", name))

	// 参数
	params := make([]string, 0)
	pathParams := make([]string, 0)
	queryParams := make([]string, 0)
	hasBody := method == "POST" || method == "PUT" || method == "PATCH"

	for _, param := range op.Parameters {
		if param.Required {
			switch param.In {
			case "path":
				pathParams = append(pathParams, param.Name)
				params = append(params, fmt.Sprintf("%s %s", g.toGoType(param.Schema), param.Name))
			case "query":
				queryParams = append(queryParams, param.Name)
				params = append(params, fmt.Sprintf("%s %s", g.toGoType(param.Schema), param.Name))
			case "header":
				params = append(params, fmt.Sprintf("%s string", param.Name))
			}
		}
	}

	if hasBody && op.RequestBody != nil {
		params = append(params, "body interface{}")
	}

	if len(params) > 0 {
		builder.WriteString(", ")
		builder.WriteString(strings.Join(params, ", "))
	}

	builder.WriteString(") (*http.Response, error) {\n")

	// 构建路径
 fullPath := path
	for _, param := range pathParams {
		fullPath = strings.ReplaceAll(fullPath, "{"+param+"}", fmt.Sprintf("%s", param))
	}

	// 查询参数
	if len(queryParams) > 0 {
		builder.WriteString("\tqueryParams := url.Values{}\n")
		for _, param := range queryParams {
			builder.WriteString(fmt.Sprintf("\tqueryParams.Add(\"%s\", fmt.Sprintf(\"%%v\", %s))\n", param, param))
		}
		builder.WriteString(fmt.Sprintf("\tpath = %q + \"?\" + queryParams.Encode()\n", path))
	} else {
		builder.WriteString(fmt.Sprintf("\tpath := %q\n", fullPath))
	}

	// 调用doRequest
	callArgs := []string{"ctx", fmt.Sprintf("%q", method), "path"}
	if hasBody {
		callArgs = append(callArgs, "body")
	}
	if len(op.Parameters) > 0 {
		callArgs = append(callArgs, "map[string]string{}")
	}

	builder.WriteString(fmt.Sprintf("\treturn c.doRequest(%s)\n", strings.Join(callArgs, ", ")))
	builder.WriteString("}\n\n")

	return builder.String()
}

// generateModels 生成数据模型
func (g *Generator) generateModels() error {
	var models strings.Builder

	if schemas, ok := g.Spec.Components["schemas"].(map[string]interface{}); ok {
		for name, schema := range schemas {
			models.WriteString(g.generateModel(name, schema))
		}
	}

	tmpl := `package {{.Package}}

// 数据模型
{{.Models}}
`

	data := map[string]interface{}{
		"Package": g.Package,
		"Models":  models.String(),
	}

	return g.executeTemplate("models.go", tmpl, data)
}

// generateModel 生成模型代码
func (g *Generator) generateModel(name string, schema interface{}) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("// %s 数据模型\n", name))
	builder.WriteString(fmt.Sprintf("type %s struct {\n", name))

	// 简化实现，实际应该解析schema
	builder.WriteString("\t// TODO: 生成字段\n")
	builder.WriteString("}\n\n")

	return builder.String()
}

// generateConfig 生成配置
func (g *Generator) generateConfig() error {
	tmpl := `package {{.Package}}

import "time"

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	BaseURL: "{{.ServerURL}}",
	HTTPClient: &http.Client{
		Timeout: 30 * time.Second,
	},
}
`

	return g.executeTemplate("config.go", tmpl, nil)
}

// executeTemplate 执行模板
func (g *Generator) executeTemplate(filename, tmplStr string, data interface{}) error {
	if data == nil {
		data = g
	}

	// 创建模板
	funcMap := template.FuncMap{
		"toGoType": g.toGoType,
	}

	tmpl, err := template.New("clientgen").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 创建输出目录
	if g.OutputDir != "" {
		if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
			return err
		}
	}

	// 写入文件
	outputPath := filepath.Join(g.OutputDir, filename)
	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Generated: %s\n", outputPath)
	return nil
}

// toGoType 转换为Go类型
func (g *Generator) toGoType(schema *Schema) string {
	if schema == nil {
		return "interface{}"
	}

	if schema.Ref != "" {
		return strings.TrimPrefix(schema.Ref, "#/components/schemas/")
	}

	switch schema.Type {
	case "string":
		if schema.Format == "date-time" || schema.Format == "date" {
			return "time.Time"
		}
		return "string"
	case "integer":
		if schema.Format == "int64" {
			return "int64"
		}
		return "int"
	case "number":
		if schema.Format == "double" {
			return "float64"
		}
		return "float32"
	case "boolean":
		return "bool"
	case "array":
		if schema.Items != nil {
			return "[]" + g.toGoType(schema.Items)
		}
		return "[]interface{}"
	default:
		return "interface{}"
	}
}

// generateMethodName 生成方法名
func (g *Generator) generateMethodName(operationID, method, path string) string {
	if operationID != "" {
		return toCamelCase(operationID)
	}

	// 从路径生成方法名
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var name string

	if len(parts) > 0 {
		name = parts[len(parts)-1]
	} else {
		name = "Root"
	}

	switch method {
	case "GET":
		return "Get" + toCamelCase(name)
	case "POST":
		return "Create" + toCamelCase(name)
	case "PUT", "PATCH":
		return "Update" + toCamelCase(name)
	case "DELETE":
		return "Delete" + toCamelCase(name)
	default:
		return toCamelCase(method + "_" + name)
	}
}

// toCamelCase 转换为驼峰命名
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	var result string

	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}

	return result
}

// GetServerURL 获取服务器URL
func (g *Generator) GetServerURL() string {
	if len(g.Spec.Servers) > 0 {
		return g.Spec.Servers[0].URL
	}
	return "http://localhost:8080"
}
