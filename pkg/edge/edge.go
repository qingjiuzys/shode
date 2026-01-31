// Package edge 提供边缘计算功能。
package edge

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// EdgeFunction 边缘函数
type EdgeFunction struct {
	ID       string
	Name     string
	Handler  func(ctx context.Context, input *FunctionInput) (*FunctionOutput, error)
	Timeout  time.Duration
	Memory   int64
	Regions  []string
	Enabled  bool
}

// FunctionInput 函数输入
type FunctionInput struct {
	Data     interface{}                 `json:"data"`
	Headers  map[string]string          `json:"headers"`
	Query    map[string]string          `json:"query"`
	Metadata map[string]interface{}     `json:"metadata"`
}

// FunctionOutput 函数输出
type FunctionOutput struct {
	Data     interface{}            `json:"data"`
	Headers  map[string]string      `json:"headers"`
	Status   int                    `json:"status"`
	Error    string                 `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EdgeRuntime 边缘运行时
type EdgeRuntime struct {
	functions map[string]*EdgeFunction
	cache     *EdgeCache
	metrics   *EdgeMetrics
	mu        sync.RWMutex
}

// NewEdgeRuntime 创建边缘运行时
func NewEdgeRuntime() *EdgeRuntime {
	return &EdgeRuntime{
		functions: make(map[string]*EdgeFunction),
		cache:     NewEdgeCache(),
		metrics:   NewEdgeMetrics(),
	}
}

// Register 注册函数
func (er *EdgeRuntime) Register(fn *EdgeFunction) error {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.functions[fn.ID] = fn
	return nil
}

// Invoke 调用函数
func (er *EdgeRuntime) Invoke(ctx context.Context, functionID string, input *FunctionInput) (*FunctionOutput, error) {
	er.mu.RLock()
	fn, exists := er.functions[functionID]
	er.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("function not found: %s", functionID)
	}

	if !fn.Enabled {
		return nil, fmt.Errorf("function disabled: %s", functionID)
	}

	// 添加超时
	if fn.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, fn.Timeout)
		defer cancel()
	}

	start := time.Now()

	// 执行函数
	result, err := fn.Handler(ctx, input)

	duration := time.Since(start)

	// 记录指标
	er.metrics.RecordInvocation(functionID, duration, err)

	if err != nil {
		return &FunctionOutput{
			Status: 500,
			Error:   err.Error(),
		}, nil
	}

	return result, nil
}

// GetFunction 获取函数
func (er *EdgeRuntime) GetFunction(functionID string) (*EdgeFunction, bool) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	fn, exists := er.functions[functionID]
	return fn, exists
}

// ListFunctions 列出所有函数
func (er *EdgeRuntime) ListFunctions() []*EdgeFunction {
	er.mu.RLock()
	defer er.mu.RUnlock()

	functions := make([]*EdgeFunction, 0, len(er.functions))
	for _, fn := range er.functions {
		functions = append(functions, fn)
	}
	return functions
}

// EdgeCache 边缘缓存
type EdgeCache struct {
	items    map[string]*CacheItem
	mu       sync.RWMutex
	ttl      time.Duration
	maxSize  int
}

// CacheItem 缓存项
type CacheItem struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	HitCount  int
}

// NewEdgeCache 创建边缘缓存
func NewEdgeCache() *EdgeCache {
	return &EdgeCache{
		items:   make(map[string]*CacheItem),
		ttl:     5 * time.Minute,
		maxSize: 1000,
	}
}

// Get 获取
func (ec *EdgeCache) Get(key string) (interface{}, bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	item, exists := ec.items[key]
	if !exists {
		return nil, false
	}

	// 检查过期
	if time.Now().After(item.ExpiresAt) {
		delete(ec.items, key)
		return nil, false
	}

	item.HitCount++
	return item.Value, true
}

// Set 设置
func (ec *EdgeCache) Set(key string, value interface{}, ttl time.Duration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	// 如果超过大小，淘汰最少使用的项
	if len(ec.items) >= ec.maxSize {
		ec.evictLRU()
	}

	expiresAt := time.Now().Add(ttl)
	if ttl == 0 {
		expiresAt = time.Now().Add(ec.ttl)
	}

	ec.items[key] = &CacheItem{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
		HitCount:  0,
	}
}

// Delete 删除
func (ec *EdgeCache) Delete(key string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	delete(ec.items, key)
}

// Clear 清空
func (ec *EdgeCache) Clear() {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	ec.items = make(map[string]*CacheItem)
}

// evictLRU 淘汰最少使用的项
func (ec *EdgeCache) evictLRU() {
	var lruKey string
	minHits := int(^uint(0) >> 1)

	for key, item := range ec.items {
		if item.HitCount < minHits {
			minHits = item.HitCount
			lruKey = key
		}
	}

	if lruKey != "" {
		delete(ec.items, lruKey)
	}
}

// EdgeMetrics 边缘指标
type EdgeMetrics struct {
	invocations map[string]*InvocationStats
	mu          sync.RWMutex
}

// InvocationStats 调用统计
type InvocationStats struct {
	Total    int64
	Success  int64
	Errors   int64
	AvgLatency float64
	P95Latency float64
	P99Latency float64
}

// NewEdgeMetrics 创建边缘指标
func NewEdgeMetrics() *EdgeMetrics {
	return &EdgeMetrics{
		invocations: make(map[string]*InvocationStats),
	}
}

// RecordInvocation 记录调用
func (em *EdgeMetrics) RecordInvocation(functionID string, duration time.Duration, err error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if _, exists := em.invocations[functionID]; !exists {
		em.invocations[functionID] = &InvocationStats{}
	}

	stats := em.invocations[functionID]
	stats.Total++

	if err != nil {
		stats.Errors++
	} else {
		stats.Success++
	}

	// 简化的平均延迟计算
	latency := float64(duration.Milliseconds())
	stats.AvgLatency = (stats.AvgLatency*float64(stats.Total-1) + latency) / float64(stats.Total)
}

// GetStats 获取统计
func (em *EdgeMetrics) GetStats(functionID string) (*InvocationStats, bool) {
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats, exists := em.invocations[functionID]
	return stats, exists
}

// CDNManager CDN 管理器
type CDNManager struct {
	providers map[string]*CDNProvider
	cache     *EdgeCache
	mu        sync.RWMutex
}

// CDNProvider CDN 提供商
type CDNProvider struct {
	Name     string
	Endpoint string
	Enabled  bool
	Priority int
}

// NewCDNManager 创建 CDN 管理器
func NewCDNManager() *CDNManager {
	return &CDNManager{
		providers: make(map[string]*CDNProvider),
		cache:     NewEdgeCache(),
	}
}

// AddProvider 添加提供商
func (cm *CDNManager) AddProvider(provider *CDNProvider) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.providers[provider.Name] = provider
}

// GetURL 获取 URL
func (cm *CDNManager) GetURL(key string) (string, error) {
	// 检查缓存
	if url, exists := cm.cache.Get("url:" + key); exists {
		return url.(string), nil
	}

	// 简化实现，返回第一个提供商的 URL
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, provider := range cm.providers {
		if provider.Enabled {
			url := fmt.Sprintf("%s/%s", provider.Endpoint, key)

			// 缓存 URL
			cm.cache.Set("url:"+key, url, 10*time.Minute)

			return url, nil
		}
	}

	return "", fmt.Errorf("no available CDN provider")
}

// Purge 清除缓存
func (cm *CDNManager) Purge(key string) error {
	cm.cache.Delete("url:" + key)
	return nil
}

// Invalidate 失效缓存
func (cm *CDNManager) Invalidate(keys []string) error {
	for _, key := range keys {
		cm.cache.Delete("url:" + key)
	}
	return nil
}

// EdgeRouter 边缘路由器
type EdgeRouter struct {
	routes   map[string]*Route
	mu       sync.RWMutex
	cache    *EdgeCache
}

// Route 路由
type Route struct {
	Path        string
	Destination string
	Methods     []string
	CachePolicy *CachePolicy
	Middleware  []MiddlewareFunc
}

// CachePolicy 缓存策略
type CachePolicy struct {
	Enabled    bool
	TTL        time.Duration
	ByPassQuery []string
}

// MiddlewareFunc 中间件函数
type MiddlewareFunc func(ctx context.Context, input *FunctionInput) (*FunctionInput, error)

// NewEdgeRouter 创建边缘路由器
func NewEdgeRouter() *EdgeRouter {
	return &EdgeRouter{
		routes: make(map[string]*Route),
		cache:  NewEdgeCache(),
	}
}

// AddRoute 添加路由
func (er *EdgeRouter) AddRoute(route *Route) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.routes[route.Path] = route
}

// Route 路由请求
func (er *EdgeRouter) Route(ctx context.Context, path, method string, input *FunctionInput) (*Route, bool) {
	er.mu.RLock()
	defer er.mu.RUnlock()

	route, exists := er.routes[path]
	if !exists {
		return nil, false
	}

	// 检查方法
	if len(route.Methods) > 0 {
		methodAllowed := false
		for _, m := range route.Methods {
			if m == method {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			return nil, false
		}
	}

	return route, true
}

// OfflineMode 离线模式
type OfflineMode struct {
	enabled     bool
	queue       []*FunctionInput
	syncQueue   []*FunctionInput
	maxQueueSize int
	mu          sync.Mutex
}

// NewOfflineMode 创建离线模式
func NewOfflineMode() *OfflineMode {
	return &OfflineMode{
		queue:        make([]*FunctionInput, 0),
		syncQueue:    make([]*FunctionInput, 0),
		maxQueueSize: 1000,
	}
}

// Enable 启用离线模式
func (om *OfflineMode) Enable() {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.enabled = true
}

// Disable 禁用离线模式
func (om *OfflineMode) Disable() {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.enabled = false
}

// IsEnabled 检查是否启用
func (om *OfflineMode) IsEnabled() bool {
	om.mu.Lock()
	defer om.mu.Unlock()
	return om.enabled
}

// Enqueue 入队
func (om *OfflineMode) Enqueue(input *FunctionInput) error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if len(om.queue) >= om.maxQueueSize {
		return fmt.Errorf("queue is full")
	}

	om.queue = append(om.queue, input)
	return nil
}

// Dequeue 出队
func (om *OfflineMode) Dequeue() *FunctionInput {
	om.mu.Lock()
	defer om.mu.Unlock()

	if len(om.queue) == 0 {
		return nil
	}

	input := om.queue[0]
	om.queue = om.queue[1:]
	return input
}

// Sync 同步
func (om *OfflineMode) Sync(runtime *EdgeRuntime) error {
	om.mu.Lock()
	om.syncQueue = om.queue
	om.queue = make([]*FunctionInput, 0)
	om.mu.Unlock()

	// 处理排队的请求
	for _, input := range om.syncQueue {
		// 简化实现，实际应该执行函数
		_ = input
	}

	return nil
}

// Region 区域
type Region struct {
	Name      string
	Location  string
	Endpoints []string
	Latency   time.Duration
	Bandwidth int64
}

// EdgeLocation 边缘位置
type EdgeLocation struct {
	Region  string
	City    string
	Lat     float64
	Lon     float64
}

// EdgeNetwork 边缘网络
type EdgeNetwork struct {
	locations map[string]*EdgeLocation
	regions   map[string]*Region
	mu        sync.RWMutex
}

// NewEdgeNetwork 创建边缘网络
func NewEdgeNetwork() *EdgeNetwork {
	return &EdgeNetwork{
		locations: make(map[string]*EdgeLocation),
		regions:   make(map[string]*Region),
	}
}

// AddLocation 添加位置
func (en *EdgeNetwork) AddLocation(id string, location *EdgeLocation) {
	en.mu.Lock()
	defer en.mu.Unlock()

	en.locations[id] = location
}

// AddRegion 添加区域
func (en *EdgeNetwork) AddRegion(name string, region *Region) {
	en.mu.Lock()
	defer en.mu.Unlock()

	en.regions[name] = region
}

// GetNearest 获取最近的位置
func (en *EdgeNetwork) GetNearest(lat, lon float64) (*EdgeLocation, error) {
	en.mu.RLock()
	defer en.mu.RUnlock()

	if len(en.locations) == 0 {
		return nil, fmt.Errorf("no locations available")
	}

	var nearest *EdgeLocation
	minDist := 0.0

	for _, location := range en.locations {
		dist := calculateDistance(lat, lon, location.Lat, location.Lon)
		if nearest == nil || dist < minDist {
			nearest = location
			minDist = dist
		}
	}

	return nearest, nil
}

// calculateDistance 计算距离（简化）
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 简化实现，实际应该用 Haversine 公式
	dlat := lat2 - lat1
	dlon := lon2 - lon1
	return dlat*dlat + dlon*dlon
}

// Deployment 部署
type Deployment struct {
	FunctionID  string
	Version     string
	Regions     []string
	Strategy    DeploymentStrategy
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DeploymentStrategy 部署策略
type DeploymentStrategy int

const (
	StrategyAll DeploymentStrategy = iota
	StrategyNearest
	StrategyLowestLatency
	StrategyLeastLoaded
)

// EdgeDeployer 边缘部署器
type EdgeDeployer struct {
	deployments map[string]*Deployment
	runtime     *EdgeRuntime
	mu          sync.RWMutex
}

// NewEdgeDeployer 创建边缘部署器
func NewEdgeDeployer(runtime *EdgeRuntime) *EdgeDeployer {
	return &EdgeDeployer{
		deployments: make(map[string]*Deployment),
		runtime:     runtime,
	}
}

// Deploy 部署
func (ed *EdgeDeployer) Deploy(deployment *Deployment) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	// 验证函数是否存在
	fn, exists := ed.runtime.GetFunction(deployment.FunctionID)
	if !exists {
		return fmt.Errorf("function not found: %s", deployment.FunctionID)
	}

	// 更新函数的区域
	fn.Regions = deployment.Regions

	deployment.Status = "deployed"
	deployment.UpdatedAt = time.Now()

	ed.deployments[deployment.FunctionID] = deployment

	return nil
}

// Undeploy 卸载
func (ed *EdgeDeployer) Undeploy(functionID string) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if deployment, exists := ed.deployments[functionID]; exists {
		deployment.Status = "undeployed"
		deployment.UpdatedAt = time.Now()
		return nil
	}

	return fmt.Errorf("deployment not found: %s", functionID)
}

// GetDeployment 获取部署
func (ed *EdgeDeployer) GetDeployment(functionID string) (*Deployment, bool) {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	deployment, exists := ed.deployments[functionID]
	return deployment, exists
}

// ListDeployments 列出所有部署
func (ed *EdgeDeployer) ListDeployments() []*Deployment {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	deployments := make([]*Deployment, 0, len(ed.deployments))
	for _, deployment := range ed.deployments {
		deployments = append(deployments, deployment)
	}
	return deployments
}
