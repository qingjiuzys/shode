// Package resilience 提供弹性模式实现，包括熔断器、重试、超时等
package resilience

import (
	"context"
	"errors"
	"sync"
	"time"
)

// State 熔断器状态
type State int

const (
	StateClosed State = iota // 关闭状态（正常工作）
	StateHalfOpen            // 半开状态（试探性恢复）
	StateOpen                // 开启状态（熔断）
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// Config 熔断器配置
type Config struct {
	// 最大失败次数
	MaxFailures uint32
	// 超时时间（开启到半开的等待时间）
	Timeout time.Duration
	// 半开状态下的最大请求数
	MaxRequests uint32
	// 统计时间窗口
	Interval time.Duration
	// 成功次数阈值（半开到关闭）
	SuccessThreshold uint32
	// 是否准备就绪
	ReadyToTrip ReadyToTripFunc
}

// ReadyToTripFunc 判断是否应该熔断的函数
type ReadyToTripFunc func(counts Counts) bool

// DefaultConfig 默认配置
var DefaultConfig = Config{
	MaxFailures:     5,
	Timeout:         60 * time.Second,
	MaxRequests:     3,
	Interval:        10 * time.Second,
	SuccessThreshold: 2,
	ReadyToTrip: func(counts Counts) bool {
		failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
		return counts.Requests >= 5 && failureRatio >= 0.6
	},
}

// Counts 统计计数
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name      string
	cfg       Config
	state     State
	mu        sync.RWMutex
	counts    Counts
	expiry    time.Time
	generation uint64
	onStateChange StateChangeCallback
}

// StateChangeCallback 状态变更回调
type StateChangeCallback func(name string, from State, to State)

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(name string, cfg Config) *CircuitBreaker {
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 1
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 10 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.SuccessThreshold == 0 {
		cfg.SuccessThreshold = 1
	}
	if cfg.ReadyToTrip == nil {
		cfg.ReadyToTrip = DefaultConfig.ReadyToTrip
	}

	return &CircuitBreaker{
		name:   name,
		cfg:    cfg,
		state:  StateClosed,
	}
}

// Name 获取熔断器名称
func (cb *CircuitBreaker) Name() string {
	return cb.name
}

// State 获取当前状态
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// 检查是否应该转换状态
	if cb.state == StateOpen && time.Now().After(cb.expiry) {
		return StateHalfOpen
	}
	return cb.state
}

// Execute 执行命令
func (cb *CircuitBreaker) Execute(ctx context.Context, req func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = req()
	cb.afterRequest(generation, err == nil)
	return err
}

// ExecuteWithContext 带上下文执行
func (cb *CircuitBreaker) ExecuteWithContext(ctx context.Context, req func(context.Context) error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		e := recover()
		if e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = req(ctx)
	cb.afterRequest(generation, err == nil)
	return err
}

// beforeRequest 请求前处理
func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)
	generation := cb.nextGeneration()

	// 检查是否需要重置计数器
	if cb.expiry.Before(now) {
		cb.resetCounts()
	}

	switch state {
	case StateClosed:
		cb.counts.Requests++
		return generation, nil

	case StateHalfOpen:
		// 限制半开状态下的并发请求数
		if cb.counts.Requests >= cb.cfg.MaxRequests {
			return generation, errors.New("circuit breaker: too many requests in half-open state")
		}
		cb.counts.Requests++
		return generation, nil

	default:
		return generation, errors.New("circuit breaker: open")
	}
}

// afterRequest 请求后处理
func (cb *CircuitBreaker) afterRequest(before uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)
	current := cb.nextGeneration()

	if before != current {
		return
	}

	if success {
		cb.onSuccess(state)
	} else {
		cb.onFailure(state)
	}
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess(state State) {
	cb.counts.TotalSuccesses++
	cb.counts.ConsecutiveSuccesses++
	cb.counts.ConsecutiveFailures = 0

	switch state {
	case StateClosed:
		// 保持关闭状态

	case StateHalfOpen:
		// 半开状态下连续成功达到阈值，转换到关闭状态
		if cb.counts.ConsecutiveSuccesses >= cb.cfg.SuccessThreshold {
			cb.setState(StateClosed)
		}
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure(state State) {
	cb.counts.TotalFailures++
	cb.counts.ConsecutiveFailures++
	cb.counts.ConsecutiveSuccesses = 0

	switch state {
	case StateClosed:
		// 检查是否应该熔断
		if cb.cfg.ReadyToTrip(cb.counts) || cb.counts.ConsecutiveFailures >= cb.cfg.MaxFailures {
			cb.setState(StateOpen)
		}

	case StateHalfOpen:
		// 半开状态下失败，重新进入开启状态
		cb.setState(StateOpen)
	}
}

// currentState 获取当前状态
func (cb *CircuitBreaker) currentState(now time.Time) State {
	// 如果超时，从开启转换到半开
	if cb.state == StateOpen && now.After(cb.expiry) {
		return StateHalfOpen
	}
	return cb.state
}

// nextGeneration 获取下一代计数
func (cb *CircuitBreaker) nextGeneration() uint64 {
	return cb.generation
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state State) {
	if cb.state == state {
		return
	}

	prev := cb.state
	cb.state = state
	cb.resetCounts()

	// 设置下次状态转换时间
	switch state {
	case StateOpen:
		cb.expiry = time.Now().Add(cb.cfg.Timeout)
	case StateHalfOpen:
		cb.expiry = time.Now().Add(cb.cfg.Interval)
	default:
		cb.expiry = time.Time{}
	}

	if cb.onStateChange != nil {
		cb.onStateChange(cb.name, prev, state)
	}
}

// resetCounts 重置计数器
func (cb *CircuitBreaker) resetCounts() {
	cb.counts = Counts{}
	cb.generation++
}

// SetStateChangeCallback 设置状态变更回调
func (cb *CircuitBreaker) SetStateChangeCallback(callback StateChangeCallback) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.onStateChange = callback
}

// Counts 获取统计信息
func (cb *CircuitBreaker) Counts() Counts {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.counts
}

// Reset 重置熔断器
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = StateClosed
	cb.resetCounts()
	cb.expiry = time.Time{}
	cb.generation++
}

// Allow 检查是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state := cb.currentState(now)

	if state == StateOpen {
		return false
	}

	if cb.expiry.Before(now) {
		cb.resetCounts()
	}

	return true
}

// Manager 熔断器管理器
type Manager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

// NewManager 创建熔断器管理器
func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// Get 获取熔断器
func (m *Manager) Get(name string) (*CircuitBreaker, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cb, ok := m.breakers[name]
	return cb, ok
}

// Add 添加熔断器
func (m *Manager) Add(cb *CircuitBreaker) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.breakers[cb.name] = cb
}

// Remove 移除熔断器
func (m *Manager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.breakers, name)
}

// GetOrCreate 获取或创建熔断器
func (m *Manager) GetOrCreate(name string, cfg Config) *CircuitBreaker {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cb, ok := m.breakers[name]; ok {
		return cb
	}

	cb := NewCircuitBreaker(name, cfg)
	m.breakers[name] = cb
	return cb
}

// List 列出所有熔断器
func (m *Manager) List() []*CircuitBreaker {
	m.mu.RLock()
	defer m.mu.RUnlock()

	breakers := make([]*CircuitBreaker, 0, len(m.breakers))
	for _, cb := range m.breakers {
		breakers = append(breakers, cb)
	}
	return breakers
}

// ResetAll 重置所有熔断器
func (m *Manager) ResetAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cb := range m.breakers {
		cb.Reset()
	}
}

// Metrics 熔断器指标
type Metrics struct {
	Name               string  `json:"name"`
	State              string  `json:"state"`
	Requests           uint32  `json:"requests"`
	TotalSuccesses     uint32  `json:"total_successes"`
	TotalFailures      uint32  `json:"total_failures"`
	ConsecutiveSuccesses uint32 `json:"consecutive_successes"`
	ConsecutiveFailures   uint32 `json:"consecutive_failures"`
	SuccessRate        float64 `json:"success_rate"`
}

// GetMetrics 获取指标
func (cb *CircuitBreaker) GetMetrics() Metrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	counts := cb.counts
	successRate := 0.0
	if counts.Requests > 0 {
		successRate = float64(counts.TotalSuccesses) / float64(counts.Requests)
	}

	return Metrics{
		Name:                 cb.name,
		State:                cb.state.String(),
		Requests:             counts.Requests,
		TotalSuccesses:       counts.TotalSuccesses,
		TotalFailures:        counts.TotalFailures,
		ConsecutiveSuccesses: counts.ConsecutiveSuccesses,
		ConsecutiveFailures:  counts.ConsecutiveFailures,
		SuccessRate:          successRate,
	}
}

// GetAllMetrics 获取所有熔断器指标
func (m *Manager) GetAllMetrics() []Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	metrics := make([]Metrics, 0, len(m.breakers))
	for _, cb := range m.breakers {
		metrics = append(metrics, cb.GetMetrics())
	}
	return metrics
}

// ThresholdFunc 阈值函数
type ThresholdFunc func(counts Counts) bool

// ConsecutiveFailuresThreshold 连续失败阈值函数
func ConsecutiveFailuresThreshold(threshold uint32) ThresholdFunc {
	return func(counts Counts) bool {
		return counts.ConsecutiveFailures >= threshold
	}
}

// FailureRatioThreshold 失败率阈值函数
func FailureRatioThreshold(minRequests uint32, failureRatio float64) ThresholdFunc {
	return func(counts Counts) bool {
		if counts.Requests < minRequests {
			return false
		}
		ratio := float64(counts.TotalFailures) / float64(counts.Requests)
		return ratio >= failureRatio
	}
}

// CombineThresholds 组合多个阈值函数
func CombineThresholdFuncs(funcs ...ThresholdFunc) ThresholdFunc {
	return func(counts Counts) bool {
		for _, f := range funcs {
			if f(counts) {
				return true
			}
		}
		return false
	}
}
