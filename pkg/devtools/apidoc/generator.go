// Package apidoc 提供 API 文档生成功能
package apidoc

import (
	"fmt"
	"os"
	"strings"
)

// Generator 文档生成器
type Generator struct {
	info      *APIInfo
	outputDir string
	format    string
}

// APIInfo API 信息
type APIInfo struct {
	Title          string
	Description    string
	Version        string
	BaseURL        string
	Host           string
	Schemes        []string
	Consumes       []string
	Produces       []string
	Tags           []Tag
	Paths          map[string]*Path
	Definitions    map[string]*Schema
	SecurityDefs   map[string]*SecurityScheme
}

// Tag 标签
type Tag struct {
	Name        string
	Description string
}

// Path API 路径
type Path struct {
	Method      string
	Summary     string
	Description string
	Tags        []string
	Parameters  []Parameter
	Responses   map[int]*Response
	Security    []map[string][]string
}

// Parameter 参数
type Parameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Type        string
	Schema      *Schema
}

// Response 响应
type Response struct {
	Description string
	Schema      *Schema
}

// Schema 数据模型
type Schema struct {
	Type       string
	Properties map[string]*Property
	Required   []string
	Ref        string
}

// Property 属性
type Property struct {
	Type        string
	Description string
	Format      string
	Example     interface{}
}

// SecurityScheme 安全方案
type SecurityScheme struct {
	Type             string
	Description      string
	Name             string
	In               string
	Flow             string
	AuthorizationURL string
	TokenURL         string
	Scopes           map[string]string
}

// NewGenerator 创建生成器
func NewGenerator(title, version string) *Generator {
	return &Generator{
		info: &APIInfo{
			Title:       title,
			Version:     version,
			Schemes:     []string{"http"},
			Consumes:    []string{"application/json"},
			Produces:    []string{"application/json"},
			Paths:       make(map[string]*Path),
			Definitions: make(map[string]*Schema),
			SecurityDefs: make(map[string]*SecurityScheme),
		},
		format: "openapi",
	}
}

// SetOutputDir 设置输出目录
func (g *Generator) SetOutputDir(dir string) {
	g.outputDir = dir
}

// SetFormat 设置输出格式
func (g *Generator) SetFormat(format string) {
	g.format = format
}

// AddPath 添加路径
func (g *Generator) AddPath(method, path string, pathInfo *Path) {
	key := strings.ToUpper(method) + " " + path
	g.info.Paths[key] = pathInfo
}

// AddDefinition 添加定义
func (g *Generator) AddDefinition(name string, schema *Schema) {
	g.info.Definitions[name] = schema
}

// AddTag 添加标签
func (g *Generator) AddTag(name, description string) {
	g.info.Tags = append(g.info.Tags, Tag{
		Name:        name,
		Description: description,
	})
}

// AddSecurityScheme 添加安全方案
func (g *Generator) AddSecurityScheme(name string, scheme *SecurityScheme) {
	g.info.SecurityDefs[name] = scheme
}

// GenerateOpenAPI 生成 OpenAPI 规范
func (g *Generator) GenerateOpenAPI() error {
	var sb strings.Builder

	sb.WriteString("{\n")
	sb.WriteString(fmt.Sprintf("  \"openapi\": \"3.0.0\",\n"))
	sb.WriteString(fmt.Sprintf("  \"info\": {\n"))
	sb.WriteString(fmt.Sprintf("    \"title\": \"%s\",\n", g.info.Title))
	sb.WriteString(fmt.Sprintf("    \"version\": \"%s\"\n", g.info.Version))
	sb.WriteString(fmt.Sprintf("  },\n"))

	if g.info.Description != "" {
		sb.WriteString(fmt.Sprintf("  \"description\": \"%s\",\n", g.info.Description))
	}

	if g.info.Host != "" {
		sb.WriteString(fmt.Sprintf("  \"host\": \"%s\",\n", g.info.Host))
	}

	if g.info.BaseURL != "" {
		sb.WriteString(fmt.Sprintf("  \"basePath\": \"%s\",\n", g.info.BaseURL))
	}

	if len(g.info.Schemes) > 0 {
		sb.WriteString(fmt.Sprintf("  \"schemes\": [\"%s\"],\n", strings.Join(g.info.Schemes, "\", \"")))
	}

	// Tags
	if len(g.info.Tags) > 0 {
		sb.WriteString(fmt.Sprintf("  \"tags\": [\n"))
		for i, tag := range g.info.Tags {
			sb.WriteString(fmt.Sprintf("    {\n"))
			sb.WriteString(fmt.Sprintf("      \"name\": \"%s\",\n", tag.Name))
			sb.WriteString(fmt.Sprintf("      \"description\": \"%s\"\n", tag.Description))
			if i < len(g.info.Tags)-1 {
				sb.WriteString(fmt.Sprintf("    },\n"))
			} else {
				sb.WriteString(fmt.Sprintf("    }\n"))
			}
		}
		sb.WriteString(fmt.Sprintf("  ],\n"))
	}

	// Paths
	sb.WriteString(fmt.Sprintf("  \"paths\": {\n"))
	i := 0
	for key, path := range g.info.Paths {
		parts := strings.SplitN(key, " ", 2)
		method := strings.ToLower(parts[0])
		pathStr := parts[1]

		sb.WriteString(fmt.Sprintf("    \"%s\": {\n", pathStr))
		sb.WriteString(fmt.Sprintf("      \"%s\": {\n", method))

		if path.Summary != "" {
			sb.WriteString(fmt.Sprintf("        \"summary\": \"%s\",\n", path.Summary))
		}
		if path.Description != "" {
			sb.WriteString(fmt.Sprintf("        \"description\": \"%s\",\n", path.Description))
		}

		if len(path.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("        \"tags\": [\"%s\"],\n", strings.Join(path.Tags, "\", \"")))
		}

		if len(path.Parameters) > 0 {
			sb.WriteString(fmt.Sprintf("        \"parameters\": [\n"))
			for j, param := range path.Parameters {
				sb.WriteString(fmt.Sprintf("          {\n"))
				sb.WriteString(fmt.Sprintf("            \"name\": \"%s\",\n", param.Name))
				sb.WriteString(fmt.Sprintf("            \"in\": \"%s\",\n", param.In))
				sb.WriteString(fmt.Sprintf("            \"required\": %t,\n", param.Required))
				sb.WriteString(fmt.Sprintf("            \"type\": \"%s\"\n", param.Type))
				if j < len(path.Parameters)-1 {
					sb.WriteString(fmt.Sprintf("          },\n"))
				} else {
					sb.WriteString(fmt.Sprintf("          }\n"))
				}
			}
			sb.WriteString(fmt.Sprintf("        ],\n"))
		}

		if len(path.Responses) > 0 {
			sb.WriteString(fmt.Sprintf("        \"responses\": {\n"))
			respIdx := 0
			for code, resp := range path.Responses {
				sb.WriteString(fmt.Sprintf("          \"%d\": {\n", code))
				sb.WriteString(fmt.Sprintf("            \"description\": \"%s\"\n", resp.Description))
				if resp.Schema != nil {
					sb.WriteString(fmt.Sprintf("            \"schema\": {\n"))
					if resp.Schema.Ref != "" {
						sb.WriteString(fmt.Sprintf("              \"$ref\": \"#/definitions/%s\"\n", resp.Schema.Ref))
					}
					sb.WriteString(fmt.Sprintf("            }\n"))
				}
				if respIdx < len(path.Responses)-1 {
					sb.WriteString(fmt.Sprintf("          },\n"))
				} else {
					sb.WriteString(fmt.Sprintf("          }\n"))
				}
				respIdx++
			}
			sb.WriteString(fmt.Sprintf("        }\n"))
		}

		if i < len(g.info.Paths)-1 {
			sb.WriteString(fmt.Sprintf("      }\n    },\n"))
		} else {
			sb.WriteString(fmt.Sprintf("      }\n    }\n"))
		}
		i++
	}
	sb.WriteString(fmt.Sprintf("  },\n"))

	// Definitions
	if len(g.info.Definitions) > 0 {
		sb.WriteString(fmt.Sprintf("  \"definitions\": {\n"))
		defIdx := 0
		for name, schema := range g.info.Definitions {
			sb.WriteString(fmt.Sprintf("    \"%s\": {\n", name))
			sb.WriteString(fmt.Sprintf("      \"type\": \"%s\"\n", schema.Type))

			if len(schema.Properties) > 0 {
				sb.WriteString(fmt.Sprintf("      \"properties\": {\n"))
				propIdx := 0
				for propName, prop := range schema.Properties {
					sb.WriteString(fmt.Sprintf("        \"%s\": {\n", propName))
					sb.WriteString(fmt.Sprintf("          \"type\": \"%s\"\n", prop.Type))
					if prop.Description != "" {
						sb.WriteString(fmt.Sprintf("          \"description\": \"%s\"\n", prop.Description))
					}
					if propIdx < len(schema.Properties)-1 {
						sb.WriteString(fmt.Sprintf("        },\n"))
					} else {
						sb.WriteString(fmt.Sprintf("        }\n"))
					}
					propIdx++
				}
				sb.WriteString(fmt.Sprintf("      }\n"))
			}

			if defIdx < len(g.info.Definitions)-1 {
				sb.WriteString(fmt.Sprintf("    },\n"))
			} else {
				sb.WriteString(fmt.Sprintf("    }\n"))
			}
			defIdx++
		}
		sb.WriteString(fmt.Sprintf("  }\n"))
	}

	sb.WriteString("}\n")

	// 写入文件
	if g.outputDir == "" {
		g.outputDir = "."
	}

	filename := g.outputDir + "/openapi.json"
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write OpenAPI file: %w", err)
	}

	fmt.Printf("✓ Generated OpenAPI specification: %s\n", filename)
	return nil
}

// GenerateMarkdown 生成 Markdown 文档
func (g *Generator) GenerateMarkdown() error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# %s\n\n", g.info.Title))
	sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", g.info.Version))

	if g.info.Description != "" {
		sb.WriteString(fmt.Sprintf("%s\n\n", g.info.Description))
	}

	// Tags
	if len(g.info.Tags) > 0 {
		sb.WriteString("## Tags\n\n")
		for _, tag := range g.info.Tags {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", tag.Name, tag.Description))
		}
		sb.WriteString("\n")
	}

	// API Endpoints
	sb.WriteString("## API Endpoints\n\n")

	// 按标签分组
	tagsMap := make(map[string][]*PathItem)
	for key, path := range g.info.Paths {
		parts := strings.SplitN(key, " ", 2)
		method := parts[0]
		pathStr := parts[1]

		for _, tag := range path.Tags {
			tagsMap[tag] = append(tagsMap[tag], &PathItem{
				Method:  method,
				Path:    pathStr,
				PathObj: path,
			})
		}
	}

	for tag, items := range tagsMap {
		sb.WriteString(fmt.Sprintf("### %s\n\n", tag))

		for _, item := range items {
			sb.WriteString(fmt.Sprintf("#### %s %s\n\n", item.Method, item.Path))

			if item.PathObj.Summary != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", item.PathObj.Summary))
			}

			if item.PathObj.Description != "" {
				sb.WriteString(fmt.Sprintf("**Description:** %s\n\n", item.PathObj.Description))
			}

			if len(item.PathObj.Parameters) > 0 {
				sb.WriteString("**Parameters:**\n\n")
				sb.WriteString("| Name | In | Type | Required | Description |\n")
				sb.WriteString("|------|-----|------|----------|-------------|\n")
				for _, param := range item.PathObj.Parameters {
					req := "false"
					if param.Required {
						req = "true"
					}
					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
						param.Name, param.In, param.Type, req, param.Description))
				}
				sb.WriteString("\n")
			}

			if len(item.PathObj.Responses) > 0 {
				sb.WriteString("**Responses:**\n\n")
				for code, resp := range item.PathObj.Responses {
					sb.WriteString(fmt.Sprintf("- **%d**: %s\n", code, resp.Description))
				}
				sb.WriteString("\n")
			}
		}
	}

	// Data Models
	if len(g.info.Definitions) > 0 {
		sb.WriteString("## Data Models\n\n")
		for name, schema := range g.info.Definitions {
			sb.WriteString(fmt.Sprintf("### %s\n\n", name))
			sb.WriteString(fmt.Sprintf("**Type:** %s\n\n", schema.Type))

			if len(schema.Properties) > 0 {
				sb.WriteString("| Field | Type | Description |\n")
				sb.WriteString("|-------|------|-------------|\n")
				for propName, prop := range schema.Properties {
					sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
						propName, prop.Type, prop.Description))
				}
				sb.WriteString("\n")
			}
		}
	}

	// 写入文件
	if g.outputDir == "" {
		g.outputDir = "."
	}

	filename := g.outputDir + "/api.md"
	if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("failed to write Markdown file: %w", err)
	}

	fmt.Printf("✓ Generated Markdown documentation: %s\n", filename)
	return nil
}

// PathItem 路径项
type PathItem struct {
	Method  string
	Path    string
	PathObj *Path
}
