// Package serverless 提供无服务器计算功能。
package serverless

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ServerlessPlatform 无服务器平台
type ServerlessPlatform struct {
	functions     map[string]*Function
	deployments   map[string]*Deployment
	triggers      map[string]*Trigger
	invocations   map[string]*Invocation
	scaler        *AutoScaler
	scheduler     *InvocationScheduler
	runtimeMgr    *RuntimeManager
	versionMgr    *VersionManager
	metrics       *FunctionMetrics
	coldStartOpt  *ColdStartOptimizer
	mu            sync.RWMutex
}

// NewServerlessPlatform 创建无服务器平台
func NewServerlessPlatform() *ServerlessPlatform {
	return &ServerlessPlatform{
		functions:    make(map[string]*Function),
		deployments:  make(map[string]*Deployment),
		triggers:     make(map[string]*Trigger),
		invocations:  make(map[string]*Invocation),
		scaler:       NewAutoScaler(),
		scheduler:    NewInvocationScheduler(),
		runtimeMgr:   NewRuntimeManager(),
		versionMgr:   NewVersionManager(),
		metrics:      NewFunctionMetrics(),
		coldStartOpt: NewColdStartOptimizer(),
	}
}

// DeployFunction 部署函数
func (sp *ServerlessPlatform) DeployFunction(ctx context.Context, fn *Function) (*Deployment, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	// 验证函数
	if err := sp.validateFunction(fn); err != nil {
		return nil, err
	}

	// 创建部署
	deployment := &Deployment{
		ID:          generateDeploymentID(),
		Function:    fn,
		Status:      "deploying",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	sp.deployments[deployment.ID] = deployment
	sp.functions[fn.Name] = fn

	// 启动函数
	if err := sp.runtimeMgr.StartFunction(ctx, fn); err != nil {
		return nil, err
	}

	deployment.Status = "active"

	return deployment, nil
}

// InvokeFunction 调用函数
func (sp *ServerlessPlatform) InvokeFunction(ctx context.Context, functionName string, payload interface{}) (*Invocation, error) {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	fn, exists := sp.functions[functionName]
	if !exists {
		return nil, fmt.Errorf("function not found: %s", functionName)
	}

	// 创建调用
	invocation := &Invocation{
		ID:         generateInvocationID(),
		Function:   fn,
		Payload:    payload,
		Status:     "pending",
		StartTime:  time.Now(),
	}

	sp.invocations[invocation.ID] = invocation

	// 调度执行
	if err := sp.scheduler.Schedule(ctx, fn, invocation); err != nil {
		invocation.Status = "failed"
		invocation.Error = err.Error()
		return nil, err
	}

	return invocation, nil
}

// UpdateFunction 更新函数
func (sp *ServerlessPlatform) UpdateFunction(ctx context.Context, functionName string, fn *Function) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	oldFn, exists := sp.functions[functionName]
	if !exists {
		return fmt.Errorf("function not found: %s", functionName)
	}

	// 保存旧版本
	sp.versionMgr.SaveVersion(oldFn)

	// 更新函数
	sp.functions[functionName] = fn

	return nil
}

// DeleteFunction 删除函数
func (sp *ServerlessPlatform) DeleteFunction(ctx context.Context, functionName string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	fn, exists := sp.functions[functionName]
	if !exists {
		return fmt.Errorf("function not found: %s", functionName)
	}

	// 停止函数
	if err := sp.runtimeMgr.StopFunction(ctx, fn); err != nil {
		return err
	}

	delete(sp.functions, functionName)

	return nil
}

// GetInvocation 获取调用
func (sp *ServerlessPlatform) GetInvocation(invocationID string) (*Invocation, bool) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	invocation, exists := sp.invocations[invocationID]
	return invocation, exists
}

// validateFunction 验证函数
func (sp *ServerlessPlatform) validateFunction(fn *Function) error {
	if fn.Name == "" {
		return fmt.Errorf("function name is required")
	}
	if fn.Runtime == "" {
		return fmt.Errorf("runtime is required")
	}
	if fn.Handler == "" {
		return fmt.Errorf("handler is required")
	}
	return nil
}

// Function 函数
type Function struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Runtime     string                 `json:"runtime"`
	Handler     string                 `json:"handler"`
	Code        *FunctionCode          `json:"code"`
	Config      *FunctionConfig        `json:"config"`
	Environment map[string]string      `json:"environment"`
	Layers      []string               `json:"layers"`
	Metadata    map[string]string      `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// FunctionCode 函数代码
type FunctionCode struct {
	Location   string `json:"location"`   // "inline", "s3", "git"
	Repository string `json:"repository"`
	Branch     string `json:"branch"`
	Source     string `json:"source"`
	ZipFile    []byte `json:"zip_file"`
}

// FunctionConfig 函数配置
type FunctionConfig struct {
	Memory     int           `json:"memory"`
	Timeout    time.Duration `json:"timeout"`
	EphemeralStorage int      `json:"ephemeral_storage"`
	ReservedConcurrency *int  `json:"reserved_concurrency"`
}

// Deployment 部署
type Deployment struct {
	ID        string       `json:"id"`
	Function  *Function    `json:"function"`
	Status    string       `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Version   string       `json:"version"`
}

// Invocation 调用
type Invocation struct {
	ID           string                 `json:"id"`
	Function     *Function              `json:"function"`
	Payload      interface{}            `json:"payload"`
	Result       interface{}            `json:"result"`
	Status       string                 `json:"status"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	MemoryUsed   int                    `json:"memory_used"`
	Error        string                 `json:"error,omitempty"`
	Logs         []string               `json:"logs"`
	RequestID    string                 `json:"request_id"`
	TraceID      string                 `json:"trace_id"`
}

// Trigger 触发器
type Trigger struct {
	ID          string                 `json:"id"`
	Function    string                 `json:"function"`
	Type        string                 `json:"type"` // "http", "timer", "queue", "storage", "stream"
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AutoScaler 自动扩缩容
type AutoScaler struct {
	policies   map[string]*ScalingPolicy
	instances  map[string]int
	metrics    *ScalingMetrics
	mu         sync.RWMutex
}

// ScalingPolicy 扩缩容策略
type ScalingPolicy struct {
	FunctionName   string        `json:"function_name"`
	MinInstances   int           `json:"min_instances"`
	MaxInstances   int           `json:"max_instances"`
	TargetCPU      float64       `json:"target_cpu"`
	TargetMemory   float64       `json:"target_memory"`
	ScaleUpCooldown time.Duration `json:"scale_up_cooldown"`
	ScaleDownCooldown time.Duration `json:"scale_down_cooldown"`
}

// ScalingMetrics 扩缩容指标
type ScalingMetrics struct {
	CurrentInstances int     `json:"current_instances"`
	CurrentCPU       float64 `json:"current_cpu"`
	CurrentMemory    float64 `json:"current_memory"`
	RequestRate      float64 `json:"request_rate"`
	AvgDuration      time.Duration `json:"avg_duration"`
}

// NewAutoScaler 创建自动扩缩容
func NewAutoScaler() *AutoScaler {
	return &AutoScaler{
		policies:  make(map[string]*ScalingPolicy),
		instances: make(map[string]int),
		metrics:   &ScalingMetrics{},
	}
}

// SetPolicy 设置策略
func (as *AutoScaler) SetPolicy(policy *ScalingPolicy) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.policies[policy.FunctionName] = policy
	as.instances[policy.FunctionName] = policy.MinInstances
}

// ScaleUp 扩容
func (as *AutoScaler) ScaleUp(functionName string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	policy, exists := as.policies[functionName]
	if !exists {
		return fmt.Errorf("policy not found: %s", functionName)
	}

	current := as.instances[functionName]
	if current >= policy.MaxInstances {
		return fmt.Errorf("already at max instances")
	}

	as.instances[functionName]++
	as.metrics.CurrentInstances++

	return nil
}

// ScaleDown 缩容
func (as *AutoScaler) ScaleDown(functionName string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	policy, exists := as.policies[functionName]
	if !exists {
		return fmt.Errorf("policy not found: %s", functionName)
	}

	current := as.instances[functionName]
	if current <= policy.MinInstances {
		return fmt.Errorf("already at min instances")
	}

	as.instances[functionName]--
	as.metrics.CurrentInstances--

	return nil
}

// GetInstances 获取实例数
func (as *AutoScaler) GetInstances(functionName string) (int, bool) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	instances, exists := as.instances[functionName]
	return instances, exists
}

// InvocationScheduler 调用调度器
type InvocationScheduler struct {
	queues      map[string][]*Invocation
	workers     map[string]*WorkerPool
	strategy    string // "round-robin", "least-connections", "priority"
	mu          sync.RWMutex
}

// WorkerPool 工作池
type WorkerPool struct {
	Function   string
	Workers    []*Worker
	MaxWorkers int
	mu         sync.RWMutex
}

// Worker 工作线程
type Worker struct {
	ID         string
	Busy       bool
	CurrentInv *Invocation
}

// NewInvocationScheduler 创建调用调度器
func NewInvocationScheduler() *InvocationScheduler {
	return &InvocationScheduler{
		queues:   make(map[string][]*Invocation),
		workers:  make(map[string]*WorkerPool),
		strategy: "round-robin",
	}
}

// Schedule 调度
func (is *InvocationScheduler) Schedule(ctx context.Context, fn *Function, invocation *Invocation) error {
	is.mu.Lock()
	defer is.mu.Unlock()

	// 获取工作池
	pool, exists := is.workers[fn.Name]
	if !exists {
		pool = &WorkerPool{
			Function:   fn.Name,
			Workers:    make([]*Worker, 0),
			MaxWorkers: 10,
		}
		is.workers[fn.Name] = pool
	}

	// 查找空闲工作线程
	for _, worker := range pool.Workers {
		if !worker.Busy {
			worker.Busy = true
			worker.CurrentInv = invocation
			invocation.Status = "running"

			// 异步执行
			go is.execute(ctx, fn, invocation, worker)

			return nil
		}
	}

	// 没有空闲工作线程，加入队列
	is.queues[fn.Name] = append(is.queues[fn.Name], invocation)

	return nil
}

// execute 执行
func (is *InvocationScheduler) execute(ctx context.Context, fn *Function, invocation *Invocation, worker *Worker) {
	defer func() {
		worker.Busy = false
		worker.CurrentInv = nil
		invocation.EndTime = time.Now()
		invocation.Duration = invocation.EndTime.Sub(invocation.StartTime)
	}()

	// 简化实现，直接返回成功
	invocation.Status = "completed"
	invocation.Result = fmt.Sprintf("executed %s", fn.Name)
}

// RuntimeManager 运行时管理器
type RuntimeManager struct {
	runtimes   map[string]*Runtime
	containers map[string]*Container
	mu         sync.RWMutex
}

// Runtime 运行时
type Runtime struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Type    string `json:"type"`
	Image   string `json:"image"`
}

// Container 容器
type Container struct {
	ID         string                 `json:"id"`
	Function   string                 `json:"function"`
	Runtime    string                 `json:"runtime"`
	Status     string                 `json:"status"`
	Port       int                    `json:"port"`
	Memory     int                    `json:"memory"`
	StartTime  time.Time              `json:"start_time"`
	Metadata   map[string]string      `json:"metadata"`
}

// NewRuntimeManager 创建运行时管理器
func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{
		runtimes:   make(map[string]*Runtime),
		containers: make(map[string]*Container),
	}
}

// RegisterRuntime 注册运行时
func (rm *RuntimeManager) RegisterRuntime(runtime *Runtime) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.runtimes[runtime.Name] = runtime
}

// StartFunction 启动函数
func (rm *RuntimeManager) StartFunction(ctx context.Context, fn *Function) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	container := &Container{
		ID:        generateContainerID(),
		Function:  fn.Name,
		Runtime:   fn.Runtime,
		Status:    "running",
		Memory:    fn.Config.Memory,
		StartTime: time.Now(),
		Metadata:  make(map[string]string),
	}

	rm.containers[container.ID] = container

	return nil
}

// StopFunction 停止函数
func (rm *RuntimeManager) StopFunction(ctx context.Context, fn *Function) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// 查找并停止所有相关容器
	for _, container := range rm.containers {
		if container.Function == fn.Name {
			container.Status = "stopped"
		}
	}

	return nil
}

// GetContainer 获取容器
func (rm *RuntimeManager) GetContainer(containerID string) (*Container, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	container, exists := rm.containers[containerID]
	return container, exists
}

// VersionManager 版本管理器
type VersionManager struct {
	versions   map[string][]*FunctionVersion
	aliases    map[string]string // version -> alias
	mu         sync.RWMutex
}

// FunctionVersion 函数版本
type FunctionVersion struct {
	Version     string       `json:"version"`
	Function    *Function    `json:"function"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
	Size        int64        `json:"size"`
	Checksum    string       `json:"checksum"`
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: make(map[string][]*FunctionVersion),
		aliases:  make(map[string]string),
	}
}

// SaveVersion 保存版本
func (vm *VersionManager) SaveVersion(fn *Function) *FunctionVersion {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	version := &FunctionVersion{
		Version:  fmt.Sprintf("v%d", time.Now().Unix()),
		Function: fn,
		CreatedAt: time.Now(),
	}

	vm.versions[fn.Name] = append(vm.versions[fn.Name], version)

	return version
}

// GetVersion 获取版本
func (vm *VersionManager) GetVersion(functionName, version string) (*FunctionVersion, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[functionName]
	if !exists {
		return nil, false
	}

	for _, v := range versions {
		if v.Version == version {
			return v, true
		}
	}

	return nil, false
}

// ListVersions 列出版本
func (vm *VersionManager) ListVersions(functionName string) ([]*FunctionVersion, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[functionName]
	return versions, exists
}

// SetAlias 设置别名
func (vm *VersionManager) SetAlias(version, alias string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vm.aliases[alias] = version
}

// GetAlias 获取别名
func (vm *VersionManager) GetAlias(alias string) (string, bool) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	version, exists := vm.aliases[alias]
	return version, exists
}

// FunctionMetrics 函数指标
type FunctionMetrics struct {
	metrics map[string]*MetricData
	mu      sync.RWMutex
}

// MetricData 指标数据
type MetricData struct {
	Invocations    int64         `json:"invocations"`
	Errors         int64         `json:"errors"`
	Duration       time.Duration `json:"duration"`
	AvgDuration    time.Duration `json:"avg_duration"`
	P95Duration    time.Duration `json:"p95_duration"`
	P99Duration    time.Duration `json:"p99_duration"`
	Throttles      int64         `json:"throttles"`
	ColdStarts     int64         `json:"cold_starts"`
	LastUpdated    time.Time     `json:"last_updated"`
}

// NewFunctionMetrics 创建函数指标
func NewFunctionMetrics() *FunctionMetrics {
	return &FunctionMetrics{
		metrics: make(map[string]*MetricData),
	}
}

// RecordInvocation 记录调用
func (fm *FunctionMetrics) RecordInvocation(functionName string, duration time.Duration, err error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	data, exists := fm.metrics[functionName]
	if !exists {
		data = &MetricData{}
		fm.metrics[functionName] = data
	}

	data.Invocations++
	data.Duration += duration
	data.AvgDuration = time.Duration(int64(data.Duration) / data.Invocations)
	data.LastUpdated = time.Now()

	if err != nil {
		data.Errors++
	}
}

// GetMetrics 获取指标
func (fm *FunctionMetrics) GetMetrics(functionName string) (*MetricData, bool) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	data, exists := fm.metrics[functionName]
	return data, exists
}

// Reset 重置指标
func (fm *FunctionMetrics) Reset(functionName string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.metrics, functionName)
}

// ColdStartOptimizer 冷启动优化器
type ColdStartOptimizer struct {
	strategies  map[string]*OptimizationStrategy
	pool        map[string]*WarmPool
	prewarm     map[string]bool
	mu          sync.RWMutex
}

// OptimizationStrategy 优化策略
type OptimizationStrategy struct {
	Name        string        `json:"name"`
	Type        string        `json:"type"` // "prewarm", "snapshots", "reuse"
	Config      map[string]interface{} `json:"config"`
	Enabled     bool          `json:"enabled"`
}

// WarmPool 预热池
type WarmPool struct {
	Function   string        `json:"function"`
	Size       int           `json:"size"`
	Instances  []*Container  `json:"instances"`
	LastUsed   time.Time     `json:"last_used"`
}

// NewColdStartOptimizer 创建冷启动优化器
func NewColdStartOptimizer() *ColdStartOptimizer {
	return &ColdStartOptimizer{
		strategies: make(map[string]*OptimizationStrategy),
		pool:       make(map[string]*WarmPool),
		prewarm:    make(map[string]bool),
	}
}

// AddStrategy 添加策略
func (cso *ColdStartOptimizer) AddStrategy(strategy *OptimizationStrategy) {
	cso.mu.Lock()
	defer cso.mu.Unlock()

	cso.strategies[strategy.Name] = strategy
}

// Prewarm 预热
func (cso *ColdStartOptimizer) Prewarm(functionName string, poolSize int) {
	cso.mu.Lock()
	defer cso.mu.Unlock()

	cso.prewarm[functionName] = true
	cso.pool[functionName] = &WarmPool{
		Function:  functionName,
		Size:      poolSize,
		Instances: make([]*Container, 0),
		LastUsed:  time.Now(),
	}
}

// GetWarmPool 获取预热池
func (cso *ColdStartOptimizer) GetWarmPool(functionName string) (*WarmPool, bool) {
	cso.mu.RLock()
	defer cso.mu.RUnlock()

	pool, exists := cso.pool[functionName]
	return pool, exists
}

// EventManager 事件管理器
type EventManager struct {
	sources   map[string]*EventSource
	handlers  map[string][]*EventHandler
	filters   map[string]*EventFilter
	mu        sync.RWMutex
}

// EventSource 事件源
type EventSource struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "http", "timer", "queue", "storage", "stream"
	Config      map[string]interface{} `json:"config"`
	Enabled     bool                   `json:"enabled"`
}

// EventHandler 事件处理器
type EventHandler struct {
	ID         string       `json:"id"`
	Function   string       `json:"function"`
	Source     string       `json:"source"`
	Filter     *EventFilter `json:"filter"`
	Retry      int          `json:"retry"`
	Timeout    time.Duration `json:"timeout"`
}

// EventFilter 事件过滤器
type EventFilter struct {
	Rules  map[string]string `json:"rules"`
	Logic  string            `json:"logic"` // "and", "or"
}

// NewEventManager 创建事件管理器
func NewEventManager() *EventManager {
	return &EventManager{
		sources:  make(map[string]*EventSource),
		handlers: make(map[string][]*EventHandler),
		filters:  make(map[string]*EventFilter),
	}
}

// AddSource 添加事件源
func (em *EventManager) AddSource(source *EventSource) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.sources[source.ID] = source
}

// AddHandler 添加处理器
func (em *EventManager) AddHandler(handler *EventHandler) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.handlers[handler.Source] = append(em.handlers[handler.Source], handler)
}

// Trigger 触发事件
func (em *EventManager) Trigger(ctx context.Context, sourceID string, event interface{}) error {
	em.mu.RLock()
	defer em.mu.RUnlock()

	handlers, exists := em.handlers[sourceID]
	if !exists {
		return fmt.Errorf("no handlers for source: %s", sourceID)
	}

	for _, handler := range handlers {
		// 简化实现，直接调用
		_ = handler
	}

	return nil
}

// LayerManager 层管理器
type LayerManager struct {
	layers  map[string]*Layer
	aliases map[string]string
	mu      sync.RWMutex
}

// Layer 层
type Layer struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Runtimes    []string     `json:"runtimes"`
	Code        *LayerCode   `json:"code"`
	Size        int64        `json:"size"`
	CreatedAt   time.Time    `json:"created_at"`
}

// LayerCode 层代码
type LayerCode struct {
	Location string `json:"location"`
	URI      string `json:"uri"`
	ZipFile  []byte `json:"zip_file"`
}

// NewLayerManager 创建层管理器
func NewLayerManager() *LayerManager {
	return &LayerManager{
		layers:  make(map[string]*Layer),
		aliases: make(map[string]string),
	}
}

// CreateLayer 创建层
func (lm *LayerManager) CreateLayer(layer *Layer) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.layers[layer.Name] = layer
	layer.CreatedAt = time.Now()

	return nil
}

// GetLayer 获取层
func (lm *LayerManager) GetLayer(name string) (*Layer, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	layer, exists := lm.layers[name]
	return layer, exists
}

// AliasLayer 别名层
func (lm *LayerManager) AliasLayer(name, alias string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.aliases[alias] = name
}

// LogManager 日志管理器
type LogManager struct {
	logs      map[string][]*FunctionLog
	streams   map[string]*LogStream
	retention time.Duration
	mu        sync.RWMutex
}

// FunctionLog 函数日志
type FunctionLog struct {
	ID         string                 `json:"id"`
	Function   string                 `json:"function"`
	Invocation string                 `json:"invocation"`
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// LogStream 日志流
type LogStream struct {
	Function   string    `json:"function"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	Filter     *LogFilter `json:"filter"`
}

// LogFilter 日志过滤器
type LogFilter struct {
	Level   string `json:"level"`
	Keyword string `json:"keyword"`
}

// NewLogManager 创建日志管理器
func NewLogManager() *LogManager {
	return &LogManager{
		logs:      make(map[string][]*FunctionLog),
		streams:   make(map[string]*LogStream),
		retention: 7 * 24 * time.Hour,
	}
}

// Write 写入日志
func (lm *LogManager) Write(log *FunctionLog) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.logs[log.Function] = append(lm.logs[log.Function], log)
}

// Read 读取日志
func (lm *LogManager) Read(functionName string, filter *LogFilter) ([]*FunctionLog, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	logs, exists := lm.logs[functionName]
	if !exists {
		return nil, fmt.Errorf("no logs for function: %s", functionName)
	}

	// 简化实现，返回所有日志
	return logs, nil
}

// CreateStream 创建流
func (lm *LogManager) CreateStream(functionName string, filter *LogFilter) *LogStream {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	stream := &LogStream{
		Function: functionName,
		Start:    time.Now(),
		Filter:   filter,
	}

	lm.streams[stream.Function] = stream

	return stream
}

// generateDeploymentID 生成部署 ID
func generateDeploymentID() string {
	return fmt.Sprintf("deploy_%d", time.Now().UnixNano())
}

// generateInvocationID 生成调用 ID
func generateInvocationID() string {
	return fmt.Sprintf("inv_%d", time.Now().UnixNano())
}

// generateContainerID 生成容器 ID
func generateContainerID() string {
	return fmt.Sprintf("container_%d", time.Now().UnixNano())
}
