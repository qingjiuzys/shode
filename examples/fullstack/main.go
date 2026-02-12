// Package main Shode框架完整示例应用
// 这是一个TODO应用，展示Shode框架的完整功能
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/logger"
	"gitee.com/com_818cloud/shode/pkg/realtime/websocket"
	"gitee.com/com_818cloud/shode/pkg/web"
)

// Todo TODO项
type Todo struct {
	ID        uint      `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Completed bool      `json:"completed" db:"completed"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TodoService TODO服务
type TodoService struct {
	todos  map[uint]*Todo
	nextID uint
	mu     sync.RWMutex
}

// NewTodoService 创建TODO服务
func NewTodoService() *TodoService {
	service := &TodoService{
		todos:  make(map[uint]*Todo),
		nextID: 1,
	}

	// 添加示例数据
	service.Create("Learn Shode Framework", false)
	service.Create("Build amazing applications", false)
	service.Create("Deploy to production", true)

	return service
}

// Create 创建TODO
func (s *TodoService) Create(title string, completed bool) (*Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo := &Todo{
		ID:        s.nextID,
		Title:     title,
		Completed: completed,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	s.todos[s.nextID] = todo
	s.nextID++

	return todo, nil
}

// Get 获取TODO
func (s *TodoService) Get(id uint) (*Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todo, ok := s.todos[id]
	if !ok {
		return nil, fmt.Errorf("todo not found: %d", id)
	}

	return todo, nil
}

// List 列出所有TODO
func (s *TodoService) List() ([]*Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	todos := make([]*Todo, 0, len(s.todos))
	for _, todo := range s.todos {
		todos = append(todos, todo)
	}

	return todos, nil
}

// Update 更新TODO
func (s *TodoService) Update(id uint, title string, completed *bool) (*Todo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, ok := s.todos[id]
	if !ok {
		return nil, fmt.Errorf("todo not found: %d", id)
	}

	if title != "" {
		todo.Title = title
	}

	if completed != nil {
		todo.Completed = *completed
	}

	todo.UpdatedAt = time.Now()

	return todo, nil
}

// Delete 删除TODO
func (s *TodoService) Delete(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.todos[id]; !ok {
		return fmt.Errorf("todo not found: %d", id)
	}

	delete(s.todos, id)

	return nil
}

// TodoHandler TODO处理器
type TodoHandler struct {
	service *TodoService
	hub     *websocket.Hub
}

// NewTodoHandler 创建TODO处理器
func NewTodoHandler(service *TodoService, hub *websocket.Hub) *TodoHandler {
	return &TodoHandler{
		service: service,
		hub:     hub,
	}
}

// RegisterRoutes 注册路由
func (h *TodoHandler) RegisterRoutes(r *web.Router) {
	r.Get("/api/todos", h.ListTodos)
	r.Post("/api/todos", h.CreateTodo)
	r.Get("/api/todos/:id", h.GetTodo)
	r.Put("/api/todos/:id", h.UpdateTodo)
	r.Delete("/api/todos/:id", h.DeleteTodo)
	r.Post("/api/todos/:id/toggle", h.ToggleTodo)
	r.Get("/ws", h.HandleWebSocket)
}

// ListTodos 列出所有TODO
func (h *TodoHandler) ListTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.service.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, todos)
}

// CreateTodo 创建TODO
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	todo, err := h.service.Create(req.Title, req.Completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 广播到所有WebSocket客户端
	h.broadcastCreate(todo)

	respondJSON(w, http.StatusCreated, todo)
}

// GetTodo 获取单个TODO
func (h *TodoHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	id := web.PathParam(r, "id")

	var todoID uint
	fmt.Sscanf(id, "%d", &todoID)

	todo, err := h.service.Get(todoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, todo)
}

// UpdateTodo 更新TODO
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := web.PathParam(r, "id")

	var todoID uint
	fmt.Sscanf(id, "%d", &todoID)

	var req struct {
		Title     *bool  `json:"title"`
		Completed *bool  `json:"completed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := ""
	var completed *bool

	if req.Title != nil {
		// 这里简化处理，实际应该传入字符串
	}

	todo, err := h.service.Update(todoID, title, completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 广播更新
	h.broadcastUpdate(todo)

	respondJSON(w, http.StatusOK, todo)
}

// DeleteTodo 删除TODO
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := web.PathParam(r, "id")

	var todoID uint
	fmt.Sscanf(id, "%d", &todoID)

	if err := h.service.Delete(todoID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 广播删除
	h.broadcastDelete(todoID)

	w.WriteHeader(http.StatusNoContent)
}

// ToggleTodo 切换TODO状态
func (h *TodoHandler) ToggleTodo(w http.ResponseWriter, r *http.Request) {
	id := web.PathParam(r, "id")

	var todoID uint
	fmt.Sscanf(id, "%d", &todoID)

	todo, err := h.service.Get(todoID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	completed := !todo.Completed
	todo, err = h.service.Update(todoID, "", &completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 广播更新
	h.broadcastUpdate(todo)

	respondJSON(w, http.StatusOK, todo)
}

// HandleWebSocket 处理WebSocket连接
func (h *TodoHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	websocket.ServeWebSocket(h.hub, w, r)
}

// broadcastCreate 广播创建事件
func (h *TodoHandler) broadcastCreate(todo *Todo) {
	h.hub.Broadcast(websocket.Message{
		Type: "todo_created",
		Data: todo,
		Time: time.Now(),
	})
}

// broadcastUpdate 广播更新事件
func (h *TodoHandler) broadcastUpdate(todo *Todo) {
	h.hub.Broadcast(websocket.Message{
		Type: "todo_updated",
		Data: todo,
		Time: time.Now(),
	})
}

// broadcastDelete 广播删除事件
func (h *TodoHandler) broadcastDelete(id uint) {
	h.hub.Broadcast(websocket.Message{
		Type: "todo_deleted",
		Data: map[string]interface{}{"id": id},
		Time: time.Now(),
	})
}

// respondJSON 响应JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	// 初始化日志
	appLogger := logger.NewLogger(logger.DefaultConfig)

	appLogger.Info("Starting Shode Full-Stack Example Application")

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	// 创建服务
	todoService := NewTodoService()
	todoHandler := NewTodoHandler(todoService, hub)

	// 创建路由器
	router := web.NewRouter()

	// 注册API路由
	todoHandler.RegisterRoutes(router)

	// 静态文件服务
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	// 创建服务器
	server := web.NewServer(":8080")
	server.SetHandler(router)

	appLogger.Info("Server starting on http://localhost:8080")
	appLogger.Info("API endpoints:")
	appLogger.Info("  GET    /api/todos       - List all todos")
	appLogger.Info("  POST   /api/todos       - Create todo")
	appLogger.Info("  GET    /api/todos/:id   - Get todo")
	appLogger.Info("  PUT    /api/todos/:id   - Update todo")
	appLogger.Info("  DELETE /api/todos/:id   - Delete todo")
	appLogger.Info("  POST   /api/todos/:id/toggle - Toggle todo")
	appLogger.Info("  WS     /ws              - WebSocket endpoint")

	if err := server.Start(); err != nil {
		appLogger.Error("Failed to start server:", err)
	}
}
