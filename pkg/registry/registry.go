package registry

import (
	"context"
	"sync"
)

// Registry 通用注册表
type Registry struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		items: make(map[string]interface{}),
	}
}

// Register 注册项目
func (r *Registry) Register(name string, item interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items[name] = item
}

// Unregister 注销项目
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.items, name)
}

// Get 获取项目
func (r *Registry) Get(name string) (interface{}, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	item, ok := r.items[name]
	return item, ok
}

// List 列出所有项目名称
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.items))
	for name := range r.items {
		names = append(names, name)
	}
	return names
}

// Clear 清空注册表
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.items = make(map[string]interface{})
}

// Count 返回项目数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.items)
}

// ForEach 遍历所有项目
func (r *Registry) ForEach(fn func(name string, item interface{}) error) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for name, item := range r.items {
		if err := fn(name, item); err != nil {
			return err
		}
	}
	return nil
}

// ServiceRegistry 服务注册表
type ServiceRegistry struct {
	*Registry
}

// NewServiceRegistry 创建服务注册表
func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		Registry: NewRegistry(),
	}
}

// RegisterService 注册服务
func (sr *ServiceRegistry) RegisterService(ctx context.Context, name string, service interface{}) error {
	sr.Register(name, service)
	return nil
}

// GetService 获取服务
func (sr *ServiceRegistry) GetService(ctx context.Context, name string) (interface{}, error) {
	item, ok := sr.Get(name)
	if !ok {
		return nil, ErrServiceNotFound
	}
	return item, nil
}

// Client 注册表客户端
type Client struct {
	registry *Registry
	url      string
}

// NewClient 创建客户端
func NewClient(url string) (*Client, error) {
	return &Client{
		registry: NewRegistry(),
		url:      url,
	}, nil
}

// Get 获取包
func (c *Client) Get(ctx context.Context, name string) (interface{}, error) {
	item, ok := c.registry.Get(name)
	if !ok {
		return nil, ErrServiceNotFound
	}
	return item, nil
}

// Search 搜索包
func (c *Client) Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	return []*SearchResult{}, nil
}

// SearchResult 搜索结果
type SearchResult struct {
	Name        string
	Version     string
	Description string
	Author      string
	Keywords    []string
	Downloads   int
	Verified    bool
}

// Package 包信息
type Package struct {
	Name        string
	Version     string
	Description string
	Author      string
	Files       map[string][]byte
	Scripts     map[string]string
	Dependencies map[string]string
	DevDependencies map[string]string
	Main        string
	Versions    map[string]*Package
	LatestVersion string
	License     string
	Homepage    string
	Repository  string
	Downloads   int
	Verified    bool
}

// PublishRequest 发布请求
type PublishRequest struct {
	Name        string
	Version     string
	Description string
	Author      string
	Files       map[string][]byte
	Package     *Package
	Tarball     []byte
	Checksum    string
}

// Install 安装包
func (c *Client) Install(ctx context.Context, name, version, targetDir string) error {
	return nil
}

// GetPackage 获取包信息
func (c *Client) GetPackage(ctx context.Context, name string) (*Package, error) {
	return &Package{}, nil
}

// Publish 发布包
func (c *Client) Publish(ctx context.Context, req *PublishRequest) error {
	return nil
}

// SearchQuery 搜索查询
type SearchQuery struct {
	Query  string
	Limit  int
	Offset int
}

// Errors
var (
	ErrServiceNotFound = &RegistryError{Code: "NOT_FOUND", Message: "service not found"}
	ErrAlreadyExists  = &RegistryError{Code: "EXISTS", Message: "item already exists"}
)

// RegistryError 注册表错误
type RegistryError struct {
	Code    string
	Message string
}

func (e *RegistryError) Error() string {
	return e.Message
}
