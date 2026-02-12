// Package generate 代码生成工具
package generate

import (
	"fmt"
	"os"
	"strings"
)

// Generator 代码生成器
type Generator struct {
	Package string
	Type    string
	Name    string
	Fields  map[string]string
}

// NewGenerator 创建生成器
func NewGenerator(pkg, typ, name string) *Generator {
	return &Generator{
		Package: pkg,
		Type:    typ,
		Name:    name,
		Fields:  make(map[string]string),
	}
}

// Generate 生成代码
func (g *Generator) Generate() error {
	switch g.Type {
	case "model":
		return g.generateModel()
	case "crud", "handler":
		return g.generateCRUD()
	case "service":
		return g.generateService()
	default:
		return fmt.Errorf("unknown generation type: %s", g.Type)
	}
}

// generateModel 生成 Model
func (g *Generator) generateModel() error {
	filename := fmt.Sprintf("internal/model/%s.go", strings.ToLower(g.Name))

	content := "package model\n\nimport \"time\"\n\n// " + g.Name + " 数据模型\ntype " + g.Name + " struct {\n"
	content += "\tID        uint      " + "`" + "json:\"id\" gorm:\"primaryKey\"" + "`" + "\n"
	content += "\tCreatedAt time.Time " + "`" + "json:\"created_at\"" + "`" + "\n"
	content += "\tUpdatedAt time.Time " + "`" + "json:\"updated_at\"" + "`" + "\n"
	content += "}\n"

	return os.WriteFile(filename, []byte(content), 0644)
}

// generateCRUD 生成 CRUD Handler
func (g *Generator) generateCRUD() error {
	filename := fmt.Sprintf("internal/handler/%s_handler.go", strings.ToLower(g.Name))

	content := "package handler\n\n"
	content += "import (\n"
	content += "\t\"net/http\"\n"
	content += "\t\"github.com/gin-gonic/gin\"\n"
	content += ")\n\n"
	content += "// " + g.Name + "Handler " + g.Name + " 处理器\n"
	content += "type " + g.Name + "Handler struct {\n"
	content += "}\n\n"
	content += "// New" + g.Name + "Handler 创建处理器\n"
	content += "func New" + g.Name + "Handler() *" + g.Name + "Handler {\n"
	content += "\treturn &" + g.Name + "Handler{}\n"
	content += "}\n\n"
	content += "// Create 创建\n"
	content += "func (h *" + g.Name + "Handler) Create(c *gin.Context) {\n"
	content += "\tc.JSON(http.StatusOK, gin.H{\"message\": \"Create " + g.Name + "\"})\n"
	content += "}\n\n"
	content += "// Get 获取\n"
	content += "func (h *" + g.Name + "Handler) Get(c *gin.Context) {\n"
	content += "\tc.JSON(http.StatusOK, gin.H{\"message\": \"Get " + g.Name + "\"})\n"
	content += "}\n\n"
	content += "// Update 更新\n"
	content += "func (h *" + g.Name + "Handler) Update(c *gin.Context) {\n"
	content += "\tc.JSON(http.StatusOK, gin.H{\"message\": \"Update " + g.Name + "\"})\n"
	content += "}\n\n"
	content += "// Delete 删除\n"
	content += "func (h *" + g.Name + "Handler) Delete(c *gin.Context) {\n"
	content += "\tc.JSON(http.StatusOK, gin.H{\"message\": \"Delete " + g.Name + "\"})\n"
	content += "}\n\n"
	content += "// List 列表\n"
	content += "func (h *" + g.Name + "Handler) List(c *gin.Context) {\n"
	content += "\tc.JSON(http.StatusOK, gin.H{\"message\": \"List " + g.Name + "\"})\n"
	content += "}\n"

	return os.WriteFile(filename, []byte(content), 0644)
}

// generateService 生成 Service
func (g *Generator) generateService() error {
	filename := fmt.Sprintf("internal/service/%s_service.go", strings.ToLower(g.Name))

	content := "package service\n\n"
	content += "// " + g.Name + "Service " + g.Name + " 服务\n"
	content += "type " + g.Name + "Service struct {\n"
	content += "}\n\n"
	content += "// New" + g.Name + "Service 创建服务\n"
	content += "func New" + g.Name + "Service() *" + g.Name + "Service {\n"
	content += "\treturn &" + g.Name + "Service{}\n"
	content += "}\n\n"
	content += "// Create 创建\n"
	content += "func (s *" + g.Name + "Service) Create(data interface{}) error {\n"
	content += "\treturn nil\n"
	content += "}\n\n"
	content += "// Get 获取\n"
	content += "func (s *" + g.Name + "Service) Get(id uint) (interface{}, error) {\n"
	content += "\treturn nil, nil\n"
	content += "}\n\n"
	content += "// Update 更新\n"
	content += "func (s *" + g.Name + "Service) Update(id uint, data interface{}) error {\n"
	content += "\treturn nil\n"
	content += "}\n\n"
	content += "// Delete 删除\n"
	content += "func (s *" + g.Name + "Service) Delete(id uint) error {\n"
	content += "\treturn nil\n"
	content += "}\n\n"
	content += "// List 列表\n"
	content += "func (s *" + g.Name + "Service) List() ([]interface{}, error) {\n"
	content += "\treturn nil, nil\n"
	content += "}\n"

	return os.WriteFile(filename, []byte(content), 0644)
}
