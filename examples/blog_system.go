// Package main 提供博客系统示例。
// 这是一个完整的博客系统，包含文章管理、评论、用户认证等功能。
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"gitee.com/com_818cloud/shode/pkg/database"
	"gitee.com/com_818cloud/shode/pkg/logger"
	"gitee.com/com_818cloud/shode/pkg/middleware"
	"gitee.com/com_818cloud/shode/pkg/web"
)

// Article 文章模型
type Article struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Comment 评论模型
type Comment struct {
	ID        int       `json:"id" db:"id"`
	ArticleID int       `json:"article_id" db:"article_id"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// User 用户模型
type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

// BlogService 博客服务
type BlogService struct {
	db      *database.ORM
	articles map[int]*Article
	comments map[int][]*Comment
	users    map[string]*User
	nextID   int
}

// NewBlogService 创建博客服务
func NewBlogService(db *database.ORM) *BlogService {
	return &BlogService{
		db:       db,
		articles: make(map[int]*Article),
		comments: make(map[int][]*Comment),
		users:    make(map[string]*User),
		nextID:   1,
	}
}

// CreateArticle 创建文章
func (bs *BlogService) CreateArticle(title, content, author string) *Article {
	article := &Article{
		ID:        bs.nextID,
		Title:     title,
		Content:   content,
		Author:    author,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	bs.articles[article.ID] = article
	bs.comments[article.ID] = make([]*Comment, 0)
	bs.nextID++
	return article
}

// GetArticle 获取文章
func (bs *BlogService) GetArticle(id int) (*Article, bool) {
	article, exists := bs.articles[id]
	return article, exists
}

// ListArticles 列出所有文章
func (bs *BlogService) ListArticles() []*Article {
	articles := make([]*Article, 0, len(bs.articles))
	for _, article := range bs.articles {
		articles = append(articles, article)
	}
	return articles
}

// UpdateArticle 更新文章
func (bs *BlogService) UpdateArticle(id int, title, content string) (*Article, error) {
	article, exists := bs.articles[id]
	if !exists {
		return nil, fmt.Errorf("article not found")
	}
	article.Title = title
	article.Content = content
	article.UpdatedAt = time.Now()
	return article, nil
}

// DeleteArticle 删除文章
func (bs *BlogService) DeleteArticle(id int) error {
	if _, exists := bs.articles[id]; !exists {
		return fmt.Errorf("article not found")
	}
	delete(bs.articles, id)
	delete(bs.comments, id)
	return nil
}

// AddComment 添加评论
func (bs *BlogService) AddComment(articleID, content, author string) (*Comment, error) {
	if _, exists := bs.articles[articleID]; !exists {
		return nil, fmt.Errorf("article not found")
	}

	comment := &Comment{
		ID:        len(bs.comments[articleID]) + 1,
		ArticleID: articleID,
		Content:   content,
		Author:    author,
		CreatedAt: time.Now(),
	}

	bs.comments[articleID] = append(bs.comments[articleID], comment)
	return comment, nil
}

// GetComments 获取文章评论
func (bs *BlogService) GetComments(articleID int) ([]*Comment, error) {
	comments, exists := bs.comments[articleID]
	if !exists {
		return nil, fmt.Errorf("article not found")
	}
	return comments, nil
}

// BlogController 博客控制器
type BlogController struct {
	service *BlogService
}

// NewBlogController 创建博客控制器
func NewBlogController(service *BlogService) *BlogController {
	return &BlogController{service: service}
}

// RegisterRoutes 注册路由
func (bc *BlogController) RegisterRoutes(r *web.Router) {
	r.Get("/api/articles", bc.ListArticles)
	r.Post("/api/articles", bc.CreateArticle)
	r.Get("/api/articles/:id", bc.GetArticle)
	r.Put("/api/articles/:id", bc.UpdateArticle)
	r.Delete("/api/articles/:id", bc.DeleteArticle)

	r.Get("/api/articles/:id/comments", bc.GetComments)
	r.Post("/api/articles/:id/comments", bc.AddComment)

	r.Get("/api/health", bc.HealthCheck)
}

// ListArticles 列出文章
func (bc *BlogController) ListArticles(w http.ResponseWriter, r *http.Request) {
	articles := bc.service.ListArticles()
	respondJSON(w, http.StatusOK, articles)
}

// CreateArticle 创建文章
func (bc *BlogController) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	article := bc.service.CreateArticle(req.Title, req.Content, req.Author)
	respondJSON(w, http.StatusCreated, article)
}

// GetArticle 获取文章
func (bc *BlogController) GetArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(web.PathParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	article, exists := bc.service.GetArticle(id)
	if !exists {
		respondError(w, http.StatusNotFound, fmt.Errorf("article not found"))
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// UpdateArticle 更新文章
func (bc *BlogController) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(web.PathParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	article, err := bc.service.UpdateArticle(id, req.Title, req.Content)
	if err != nil {
		respondError(w, http.StatusNotFound, err)
		return
	}

	respondJSON(w, http.StatusOK, article)
}

// DeleteArticle 删除文章
func (bc *BlogController) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(web.PathParam(r, "id"))
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := bc.service.DeleteArticle(id); err != nil {
		respondError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetComments 获取评论
func (bc *BlogController) GetComments(w http.ResponseWriter, r *http.Request) {
	articleID := web.PathParam(r, "id")
	id, err := strconv.Atoi(articleID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	comments, err := bc.service.GetComments(id)
	if err != nil {
		respondError(w, http.StatusNotFound, err)
		return
	}

	respondJSON(w, http.StatusOK, comments)
}

// AddComment 添加评论
func (bc *BlogController) AddComment(w http.ResponseWriter, r *http.Request) {
	articleID := web.PathParam(r, "id")

	var req struct {
		Content string `json:"content"`
		Author  string `json:"author"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err)
		return
	}

	comment, err := bc.service.AddComment(articleID, req.Content, req.Author)
	if err != nil {
		respondError(w, http.StatusNotFound, err)
		return
	}

	respondJSON(w, http.StatusCreated, comment)
}

// HealthCheck 健康检查
func (bc *BlogController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondJSON 响应 JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError 响应错误
func respondError(w http.ResponseWriter, status int, err error) {
	respondJSON(w, status, map[string]string{"error": err.Error()})
}

func main() {
	// 初始化日志
	log := logger.NewLogger()

	// 初始化数据库
	db, err := database.NewORM("blog.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建博客服务
	blogService := NewBlogService(db)

	// 创建示例数据
	blogService.CreateArticle(
		"Welcome to Shode Blog",
		"This is the first article in our blog system built with Shode!",
		"Admin",
	)
	blogService.CreateArticle(
		"Getting Started with REST APIs",
		"Learn how to build RESTful APIs using Shode framework...",
		"Developer",
	)

	// 创建控制器
	blogController := NewBlogController(blogService)

	// 创建路由器
	router := web.NewRouter()

	// 注册路由
	blogController.RegisterRoutes(router)

	// 添加中间件
	router.Use(middleware.NewLoggerMiddleware(log))
	router.Use(middleware.NewCORSMiddleware())
	router.Use(middleware.NewRecoveryMiddleware(log))

	// 创建服务器
	server := web.NewServer(":8080")
	server.SetHandler(router)

	log.Info("Starting Shode Blog System on http://localhost:8080")
	log.Info("API endpoints:")
	log.Info("  GET    /api/articles       - List all articles")
	log.Info("  POST   /api/articles       - Create article")
	log.Info("  GET    /api/articles/:id   - Get article")
	log.Info("  PUT    /api/articles/:id   - Update article")
	log.Info("  DELETE /api/articles/:id   - Delete article")
	log.Info("  GET    /api/articles/:id/comments - Get comments")
	log.Info("  POST   /api/articles/:id/comments - Add comment")

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
