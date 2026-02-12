// Package codegen 提供代码生成功能
package codegen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
)

// Generator 代码生成器
type Generator struct {
	Package    string
	Name       string
	Fields     []Field
	OutputPath string
}

// Field 字段定义
type Field struct {
	Name string
	Type string
	Tag  string
}

// NewGenerator 创建生成器
func NewGenerator(pkg, name string) *Generator {
	return &Generator{
		Package: pkg,
		Name:    name,
		Fields:  make([]Field, 0),
	}
}

// AddField 添加字段
func (g *Generator) AddField(name, fieldType, tag string) {
	g.Fields = append(g.Fields, Field{
		Name: name,
		Type: fieldType,
		Tag:  tag,
	})
}

// GenerateModel 生成 Model 结构体
func (g *Generator) GenerateModel() error {
	tmpl := `package {{.Package}}

import "time"

// {{.Name}} 数据模型
type {{.Name}} struct {
	ID        uint      ` + "`json:\"id\" gorm:\"primaryKey\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
{{range .Fields}}
	{{.Name}} {{.Type}} ` + "`{{.Tag}}`" + `
{{end}}
}

// TableName 指定表名
func ({{.Name}}) TableName() string {
	return "{{.Name | ToSnake}}"
}
`

	return g.generateFromTemplate(tmpl, g.OutputPath+"/"+strings.ToLower(g.Name)+".go")
}

// GenerateRepository 生成 Repository 接口和实现
func (g *Generator) GenerateRepository() error {
	interfaceTmpl := `package {{.Package}}

import (
	"context"
)

// {{.Name}}Repository {{.Name}} 仓储接口
type {{.Name}}Repository interface {
	Create(ctx context.Context, entity *{{.Name}}) error
	GetByID(ctx context.Context, id uint) (*{{.Name}}, error)
	Update(ctx context.Context, entity *{{.Name}}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*{{.Name}}, error)
}
`

	implTmpl := `package {{.Package}}

import (
	"context"
	"gorm.io/gorm"
)

type {{.Name}}RepositoryImpl struct {
	db *gorm.DB
}

// New{{.Name}}Repository 创建仓储
func New{{.Name}}Repository(db *gorm.DB) {{.Name}}Repository {
	return &{{.Name}}RepositoryImpl{db: db}
}

func (r *{{.Name}}RepositoryImpl) Create(ctx context.Context, entity *{{.Name}}) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

func (r *{{.Name}}RepositoryImpl) GetByID(ctx context.Context, id uint) (*{{.Name}}, error) {
	var entity {{.Name}}
	err := r.db.WithContext(ctx).First(&entity, id).Error
	return &entity, err
}

func (r *{{.Name}}RepositoryImpl) Update(ctx context.Context, entity *{{.Name}}) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

func (r *{{.Name}}RepositoryImpl) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&{{.Name}}{}, id).Error
}

func (r *{{.Name}}RepositoryImpl) List(ctx context.Context, limit, offset int) ([]*{{.Name}}, error) {
	var entities []*{{.Name}}
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}
`

	// 生成接口
	if err := g.generateFromTemplate(interfaceTmpl, g.OutputPath+"/"+strings.ToLower(g.Name)+"_repository.go"); err != nil {
		return err
	}

	// 生成实现
	return g.generateFromTemplate(implTmpl, g.OutputPath+"/"+strings.ToLower(g.Name)+"_repository_impl.go")
}

// GenerateService 生成 Service 层
func (g *Generator) GenerateService() error {
	tmpl := `package {{.Package}}

import (
	"context"
	"errors"
)

var (
	Err{{.Name}}NotFound = errors.New("{{.Name}} not found")
)

// {{.Name}}Service {{.Name}} 服务接口
type {{.Name}}Service interface {
	Create(ctx context.Context, entity *{{.Name}}) (*{{.Name}}, error)
	GetByID(ctx context.Context, id uint) (*{{.Name}}, error)
	Update(ctx context.Context, entity *{{.Name}}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, limit, offset int) ([]*{{.Name}}, error)
}

type {{.Name}}ServiceImpl struct {
	repo {{.Name}}Repository
}

// New{{.Name}}Service 创建服务
func New{{.Name}}Service(repo {{.Name}}Repository) {{.Name}}Service {
	return &{{.Name}}ServiceImpl{repo: repo}
}

func (s *{{.Name}}ServiceImpl) Create(ctx context.Context, entity *{{.Name}}) (*{{.Name}}, error) {
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return entity, nil
}

func (s *{{.Name}}ServiceImpl) GetByID(ctx context.Context, id uint) (*{{.Name}}, error) {
	entity, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, Err{{.Name}}NotFound
	}
	return entity, nil
}

func (s *{{.Name}}ServiceImpl) Update(ctx context.Context, entity *{{.Name}}) error {
	return s.repo.Update(ctx, entity)
}

func (s *{{.Name}}ServiceImpl) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *{{.Name}}ServiceImpl) List(ctx context.Context, limit, offset int) ([]*{{.Name}}, error) {
	return s.repo.List(ctx, limit, offset)
}
`

	return g.generateFromTemplate(tmpl, g.OutputPath+"/"+strings.ToLower(g.Name)+"_service.go")
}

// GenerateHandler 生成 HTTP Handler
func (g *Generator) GenerateHandler() error {
	tmpl := `package {{.Package}}

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type {{.Name}}Handler struct {
	service {{.Name}}Service
}

func New{{.Name}}Handler(service {{.Name}}Service) *{{.Name}}Handler {
	return &{{.Name}}Handler{service: service}
}

func (h *{{.Name}}Handler) Create(c *gin.Context) {
	var entity {{.Name}}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Create(c.Request.Context(), &entity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *{{.Name}}Handler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	entity, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *{{.Name}}Handler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var entity {{.Name}}
	if err := c.ShouldBindJSON(&entity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entity.ID = uint(id)
	if err := h.service.Update(c.Request.Context(), &entity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entity)
}

func (h *{{.Name}}Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *{{.Name}}Handler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	entities, err := h.service.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, entities)
}

// RegisterRoutes 注册路由
func (h *{{.Name}}Handler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/{{.Name | ToSnake}}s")
	{
		group.POST("", h.Create)
		group.GET("/:id", h.GetByID)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
		group.GET("", h.List)
	}
}
`

	return g.generateFromTemplate(tmpl, g.OutputPath+"/"+strings.ToLower(g.Name)+"_handler.go")
}

// generateFromTemplate 从模板生成代码
func (g *Generator) generateFromTemplate(tmplStr, outputPath string) error {
	// 创建自定义函数
	funcMap := template.FuncMap{
		"ToSnake": toSnake,
	}

	// 解析模板
	tmpl, err := template.New("codegen").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// 执行模板
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 写入文件
	if err := os.MkdirAll(g.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✓ Generated: %s\n", outputPath)
	return nil
}

// toSnake 转换为 snake_case
func toSnake(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}
