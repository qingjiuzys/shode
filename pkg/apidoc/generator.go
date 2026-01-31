// Package apidoc 提供 API 文档生成功能。
package apidoc

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// DocGenerator API 文档生成器
type DocGenerator struct {
	PackageName string
	Version     string
	BaseURL     string
	Endpoints   []Endpoint
}

// Endpoint API 端点信息
type Endpoint struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Parameters  []Parameter
	Responses   map[int]Response
	Tags        []string
}

// Parameter 参数信息
type Parameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Type        string
}

// Response 响应信息
type Response struct {
	Description string
	Schema       interface{}
}

// NewGenerator 创建文档生成器
func NewGenerator(name, version, baseURL string) *DocGenerator {
	return &DocGenerator{
		PackageName: name,
		Version:     version,
		BaseURL:     baseURL,
		Endpoints:   make([]Endpoint, 0),
	}
}

// ParseFromSource 从 Go 源码解析 API 注释
func (g *DocGenerator) ParseFromSource(dir string) error {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		ast.Inspect(pkg, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				g.extractEndpoint(fn)
			}
			return true
		})
	}

	return nil
}

// extractEndpoint 从函数声明提取端点信息
func (g *DocGenerator) extractEndpoint(fn *ast.FuncDecl) {
	if fn.Doc == nil {
		return
	}

	doc := fn.Doc.Text()
	if !strings.Contains(doc, "@http") {
		return
		// TODO: 解析注释生成 API 文档
	}
}

// ToOpenAPI 生成 OpenAPI 规范
func (g *DocGenerator) ToOpenAPI() ([]byte, error) {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       g.PackageName,
			"version":      g.Version,
			"description": "API Documentation",
		},
		"servers": []map[string]interface{}{
			{
				"url": g.BaseURL,
			},
		},
		"paths": g.buildPaths(),
	}

	return json.MarshalIndent(spec, "", "  ")
}

func (g *DocGenerator) buildPaths() map[string]interface{} {
	paths := make(map[string]interface{})

	for _, ep := range g.Endpoints {
		pathItem := map[string]interface{}{}

		operation := map[string]interface{}{
			"summary":     ep.Summary,
			"description": ep.Description,
			"responses":   ep.Responses,
			"tags":        ep.Tags,
		}

		if len(ep.Parameters) > 0 {
			operation["parameters"] = ep.Parameters
		}

		pathItem[strings.ToLower(ep.Method)] = operation
		paths[ep.Path] = pathItem
	}

	return paths
}

// AddEndpoint 手动添加端点
func (g *DocGenerator) AddEndpoint(ep Endpoint) {
	g.Endpoints = append(g.Endpoints, ep)
}

// SaveToFile 保存文档到文件
func (g *DocGenerator) SaveToFile(path string) error {
	data, err := g.ToOpenAPI()
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
