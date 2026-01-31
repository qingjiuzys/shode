// Package main 提供微服务架构示例。
// 这是一个简单的微服务系统，包含服务注册、发现、负载均衡等功能。
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/logger"
)

// ServiceInstance 服务实例
type ServiceInstance struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Address  string            `json:"address"`
	Port     int               `json:"port"`
	Metadata map[string]string `json:"metadata"`
	Healthy  bool              `json:"healthy"`
}

// ServiceRegistry 服务注册中心
type ServiceRegistry struct {
	services map[string][]*ServiceInstance // service name -> instances
	mu       sync.RWMutex
}

// NewServiceRegistry 创建服务注册中心
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string][]*ServiceInstance),
	}
}

// Register 注册服务
func (sr *ServiceRegistry) Register(instance *ServiceInstance) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.services[instance.Name] = append(sr.services[instance.Name], instance)
	fmt.Printf("Service registered: %s (%s:%d)\n", instance.Name, instance.Address, instance.Port)
}

// Deregister 注销服务
func (sr *ServiceRegistry) Deregister(serviceName, instanceID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return
	}

	for i, inst := range instances {
		if inst.ID == instanceID {
			sr.services[serviceName] = append(instances[:i], instances[i+1:]...)
			fmt.Printf("Service deregistered: %s (%s)\n", serviceName, instanceID)
			return
		}
	}
}

// Discover 发现服务
func (sr *ServiceRegistry) Discover(serviceName string) ([]*ServiceInstance, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	instances, exists := sr.services[serviceName]
	if !exists || len(instances) == 0 {
		return nil, fmt.Errorf("service not found: %s", serviceName)
	}

	// 返回健康实例
	healthy := make([]*ServiceInstance, 0)
	for _, inst := range instances {
		if inst.Healthy {
			healthy = append(healthy, inst)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy instances for service: %s", serviceName)
	}

	return healthy, nil
}

// GetAll 获取所有服务
func (sr *ServiceRegistry) GetAll() map[string][]*ServiceInstance {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make(map[string][]*ServiceInstance)
	for name, instances := range sr.services {
		instancesCopy := make([]*ServiceInstance, len(instances))
		copy(instancesCopy, instances)
		result[name] = instancesCopy
	}
	return result
}

// HealthCheck 健康检查
func (sr *ServiceRegistry) HealthCheck(serviceName, instanceID string, healthy bool) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	instances, exists := sr.services[serviceName]
	if !exists {
		return
	}

	for _, inst := range instances {
		if inst.ID == instanceID {
			inst.Healthy = healthy
			return
		}
	}
}

// LoadBalancer 负载均衡器
type LoadBalancer struct {
	registry *ServiceRegistry
	strategy string // round-robin, random, least-connection
	counters map[string]int
	mu       sync.Mutex
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(registry *ServiceRegistry, strategy string) *LoadBalancer {
	return &LoadBalancer{
		registry: registry,
		strategy: strategy,
		counters: make(map[string]int),
	}
}

// NextInstance 选择下一个实例
func (lb *LoadBalancer) NextInstance(serviceName string) (*ServiceInstance, error) {
	instances, err := lb.registry.Discover(serviceName)
	if err != nil {
		return nil, err
	}

	switch lb.strategy {
	case "round-robin":
		return lb.roundRobin(serviceName, instances), nil
	case "random":
		return lb.random(instances), nil
	default:
		return lb.roundRobin(serviceName, instances), nil
	}
}

// roundRobin 轮询策略
func (lb *LoadBalancer) roundRobin(serviceName string, instances []*ServiceInstance) *ServiceInstance {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	count := lb.counters[serviceName]
	instance := instances[count%len(instances)]
	lb.counters[serviceName] = count + 1
	return instance
}

// random 随机策略
func (lb *LoadBalancer) random(instances []*ServiceInstance) *ServiceInstance {
	return instances[time.Now().UnixNano()%int64(len(instances))]
}

// APIGateway API 网关
type APIGateway struct {
	registry      *ServiceRegistry
	loadBalancer  *LoadBalancer
	routes        map[string]string // path -> service name
	middlewares   []MiddlewareFunc
	logger        *logger.Logger
}

// MiddlewareFunc 中间件函数
type MiddlewareFunc func(http.Handler) http.Handler

// NewAPIGateway 创建 API 网关
func NewAPIGateway(registry *ServiceRegistry, loadBalancer *LoadBalancer, log *logger.Logger) *APIGateway {
	return &APIGateway{
		registry:     registry,
		loadBalancer: loadBalancer,
		routes:       make(map[string]string),
		middlewares:  make([]MiddlewareFunc, 0),
		logger:       log,
	}
}

// RegisterRoute 注册路由
func (gw *APIGateway) RegisterRoute(path, serviceName string) {
	gw.routes[path] = serviceName
	gw.logger.Info(fmt.Sprintf("Route registered: %s -> %s", path, serviceName))
}

// Use 添加中间件
func (gw *APIGateway) Use(middleware MiddlewareFunc) {
	gw.middlewares = append(gw.middlewares, middleware)
}

// ServeHTTP 处理 HTTP 请求
func (gw *APIGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 查找服务
	serviceName, exists := gw.routes[r.URL.Path]
	if !exists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// 获取服务实例
	instance, err := gw.loadBalancer.NextInstance(serviceName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// 转发请求
	gw.forwardRequest(w, r, instance)
}

// forwardRequest 转发请求
func (gw *APIGateway) forwardRequest(w http.ResponseWriter, r *http.Request, instance *ServiceInstance) {
	targetURL := fmt.Sprintf("http://%s:%d%s", instance.Address, instance.Port, r.URL.Path)

	// 创建代理请求
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 添加转发头
	proxyReq.Header.Add("X-Forwarded-For", r.RemoteAddr)
	proxyReq.Header.Add("X-Forwarded-Host", r.Host)

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to forward request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// UserService 用户服务
type UserService struct {
	instance *ServiceInstance
	registry *ServiceRegistry
	users    map[string]*User
	mu       sync.RWMutex
}

// User 用户
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// NewUserService 创建用户服务
func NewUserService(address string, port int, registry *ServiceRegistry) *UserService {
	instance := &ServiceInstance{
		ID:      fmt.Sprintf("user-%d", time.Now().UnixNano()),
		Name:    "user-service",
		Address: address,
		Port:    port,
		Metadata: map[string]string{
			"version": "1.0.0",
		},
		Healthy: true,
	}

	service := &UserService{
		instance: instance,
		registry: registry,
		users:    make(map[string]*User),
	}

	// 注册服务
	registry.Register(instance)

	return service
}

// Start 启动服务
func (us *UserService) Start() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/users/", us.handleGetUser)
	mux.HandleFunc("/users", us.handleListUsers)
	mux.HandleFunc("/health", us.handleHealthCheck)

	addr := fmt.Sprintf("%s:%d", us.instance.Address, us.instance.Port)
	us.registry.logger.Info(fmt.Sprintf("User service starting on %s", addr))

	return http.ListenAndServe(addr, mux)
}

// handleGetUser 处理获取用户
func (us *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// 简化实现
	respondJSON(w, http.StatusOK, &User{
		ID:    "1",
		Name:  "John Doe",
		Email: "john@example.com",
	})
}

// handleListUsers 处理列出用户
func (us *UserService) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users := []*User{
		{ID: "1", Name: "John Doe", Email: "john@example.com"},
		{ID: "2", Name: "Jane Smith", Email: "jane@example.com"},
	}
	respondJSON(w, http.StatusOK, users)
}

// handleHealthCheck 处理健康检查
func (us *UserService) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// respondJSON 响应 JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	// 初始化日志
	log := logger.NewLogger()

	// 创建服务注册中心
	registry := NewServiceRegistry()

	// 创建负载均衡器
	loadBalancer := NewLoadBalancer(registry, "round-robin")

	// 创建 API 网关
	gateway := NewAPIGateway(registry, loadBalancer, log)

	// 启动用户服务
	userService1 := NewUserService("localhost", 8001, registry)
	userService2 := NewUserService("localhost", 8002, registry)

	go func() {
		if err := userService1.Start(); err != nil {
			log.Fatal("User service 1 failed:", err)
		}
	}()

	go func() {
		if err := userService2.Start(); err != nil {
			log.Fatal("User service 2 failed:", err)
		}
	}()

	// 等待服务启动
	time.Sleep(1 * time.Second)

	// 注册路由
	gateway.RegisterRoute("/users", "user-service")
	gateway.RegisterRoute("/users/", "user-service")

	// 添加日志中间件
	gateway.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Info(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
			next.ServeHTTP(w, r)
			log.Info(fmt.Sprintf("Completed in %v", time.Since(start)))
		})
	})

	// 启动 API 网关
	log.Info("Starting Shode Microservices API Gateway on http://localhost:8080")
	log.Info("Registered services:")
	log.Info("  - user-service (2 instances)")

	if err := http.ListenAndServe(":8080", gateway); err != nil {
		log.Fatal("Failed to start gateway:", err)
	}
}
