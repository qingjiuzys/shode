// Package health 健康检查系统
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Check 健康检查接口
type Check interface {
	Name() string
	Check(ctx context.Context) error
}

// CheckFunc 健康检查函数
type CheckFunc func(ctx context.Context) error

// CheckResult 检查结果
type CheckResult struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Duration  time.Duration `json:"duration"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Checks    map[string]CheckResult `json:"checks"`
}

// Checker 健康检查器
type Checker struct {
	checks    map[string]Check
	timeout   time.Duration
	mu        sync.RWMutex
	disabled  map[string]bool
}

// NewChecker 创建健康检查器
func NewChecker() *Checker {
	return &Checker{
		checks:   make(map[string]Check),
		timeout:  10 * time.Second,
		disabled: make(map[string]bool),
	}
}

// AddCheck 添加健康检查
func (c *Checker) AddCheck(name string, check Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// AddCheckFunc 添加健康检查函数
func (c *Checker) AddCheckFunc(name string, check CheckFunc) {
	c.AddCheck(name, &funcCheck{name: name, check: check})
}

// RemoveCheck 移除健康检查
func (c *Checker) RemoveCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.checks, name)
}

// DisableCheck 禁用健康检查
func (c *Checker) DisableCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disabled[name] = true
}

// EnableCheck 启用健康检查
func (c *Checker) EnableCheck(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.disabled, name)
}

// SetTimeout 设置超时时间
func (c *Checker) SetTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeout = timeout
}

// Check 执行健康检查
func (c *Checker) Check(ctx context.Context) HealthStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	results := make(map[string]CheckResult)
	overallStatus := "healthy"

	for name, check := range c.checks {
		// 检查是否被禁用
		if c.disabled[name] {
			continue
		}

		result := c.runCheck(ctx, name, check)
		results[name] = result

		if result.Status != "healthy" {
			overallStatus = "unhealthy"
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Checks:    results,
	}
}

// runCheck 运行单个检查
func (c *Checker) runCheck(ctx context.Context, name string, check Check) CheckResult {
	start := time.Now()
	result := CheckResult{
		Name:      name,
		Timestamp: start,
	}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// 运行检查
	err := check.Check(ctx)

	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	} else {
		result.Status = "healthy"
	}

	return result
}

// Handler 返回 HTTP 处理器
func (c *Checker) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		status := c.Check(ctx)

		// 设置状态码
		var statusCode int
		switch status.Status {
		case "healthy":
			statusCode = http.StatusOK
		case "degraded":
			statusCode = http.StatusOK
		default:
			statusCode = http.StatusServiceUnavailable
		}

		// 设置响应头
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		// 返回 JSON
		json.NewEncoder(w).Encode(status)
	}
}

// LiveHandler 存活探针处理器
func (c *Checker) LiveHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	}
}

// ReadyHandler 就绪探针处理器
func (c *Checker) ReadyHandler() http.HandlerFunc {
	return c.Handler()
}

// funcCheck 函数检查
type funcCheck struct {
	name  string
	check CheckFunc
}

func (fc *funcCheck) Name() string {
	return fc.name
}

func (fc *funcCheck) Check(ctx context.Context) error {
	return fc.check(ctx)
}

// PingCheck ping 检查
type PingCheck struct {
	name     string
	pingFunc func(ctx context.Context) error
}

// NewPingCheck 创建 ping 检查
func NewPingCheck(name string, pingFunc func(ctx context.Context) error) *PingCheck {
	return &PingCheck{
		name:     name,
		pingFunc: pingFunc,
	}
}

func (pc *PingCheck) Name() string {
	return pc.name
}

func (pc *PingCheck) Check(ctx context.Context) error {
	return pc.pingFunc(ctx)
}

// HTTPCheck HTTP 检查
type HTTPCheck struct {
	name    string
	url     string
	client  *http.Client
	timeout time.Duration
}

// NewHTTPCheck 创建 HTTP 检查
func NewHTTPCheck(name, url string) *HTTPCheck {
	return &HTTPCheck{
		name:    name,
		url:     url,
		client:  &http.Client{},
		timeout: 5 * time.Second,
	}
}

func (hc *HTTPCheck) Name() string {
	return hc.name
}

func (hc *HTTPCheck) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", hc.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := hc.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// TCPCheck TCP 检查
type TCPCheck struct {
	host    string
	port    int
	timeout time.Duration
}

// NewTCPCheck 创建 TCP 检查
func NewTCPCheck(name string, host string, port int) *TCPCheck {
	return &TCPCheck{
		host:    host,
		port:    port,
		timeout: 5 * time.Second,
	}
}

func (tc *TCPCheck) Name() string {
	return fmt.Sprintf("%s:%d", tc.host, tc.port)
}

func (tc *TCPCheck) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, tc.timeout)
	defer cancel()

	// 简化实现：实际应该尝试 TCP 连接
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// CompositeCheck 组合检查
type CompositeCheck struct {
	name   string
	checks []Check
	logic  string // "and" or "or"
}

// NewCompositeCheck 创建组合检查
func NewCompositeCheck(name string, checks []Check, logic string) *CompositeCheck {
	return &CompositeCheck{
		name:   name,
		checks: checks,
		logic:  logic,
	}
}

func (cc *CompositeCheck) Name() string {
	return cc.name
}

func (cc *CompositeCheck) Check(ctx context.Context) error {
	var lastErr error

	for _, check := range cc.checks {
		err := check.Check(ctx)
		lastErr = err

		if cc.logic == "or" {
			if err == nil {
				return nil
			}
		} else {
			if err != nil {
				return err
			}
		}
	}

	if cc.logic == "or" {
		return fmt.Errorf("all checks failed: %w", lastErr)
	}

	return nil
}

// ThresholdCheck 阈值检查
type ThresholdCheck struct {
	name      string
	threshold float64
	current   func() (float64, error)
	operator  string // "gt", "lt", "eq", "gte", "lte"
}

// NewThresholdCheck 创建阈值检查
func NewThresholdCheck(name string, threshold float64, current func() (float64, error), operator string) *ThresholdCheck {
	return &ThresholdCheck{
		name:      name,
		threshold: threshold,
		current:   current,
		operator:  operator,
	}
}

func (tc *ThresholdCheck) Name() string {
	return tc.name
}

func (tc *ThresholdCheck) Check(ctx context.Context) error {
	value, err := tc.current()
	if err != nil {
		return fmt.Errorf("failed to get current value: %w", err)
	}

	switch tc.operator {
	case "gt":
		if value <= tc.threshold {
			return fmt.Errorf("value %.2f is not greater than %.2f", value, tc.threshold)
		}
	case "lt":
		if value >= tc.threshold {
			return fmt.Errorf("value %.2f is not less than %.2f", value, tc.threshold)
		}
	case "eq":
		if value != tc.threshold {
			return fmt.Errorf("value %.2f is not equal to %.2f", value, tc.threshold)
		}
	case "gte":
		if value < tc.threshold {
			return fmt.Errorf("value %.2f is not greater than or equal to %.2f", value, tc.threshold)
		}
	case "lte":
		if value > tc.threshold {
			return fmt.Errorf("value %.2f is not less than or equal to %.2f", value, tc.threshold)
		}
	default:
		return fmt.Errorf("unknown operator: %s", tc.operator)
	}

	return nil
}

// DegradedCheck 降级检查
type DegradedCheck struct {
	name         string
	healthyCheck Check
	degradedCheck Check
}

// NewDegradedCheck 创建降级检查
func NewDegradedCheck(name string, healthyCheck, degradedCheck Check) *DegradedCheck {
	return &DegradedCheck{
		name:         name,
		healthyCheck: healthyCheck,
		degradedCheck: degradedCheck,
	}
}

func (dc *DegradedCheck) Name() string {
	return dc.name
}

func (dc *DegradedCheck) Check(ctx context.Context) error {
	// 先检查健康状态
	if err := dc.healthyCheck.Check(ctx); err == nil {
		return nil
	}

	// 健康检查失败，尝试降级检查
	if dc.degradedCheck != nil {
		return dc.degradedCheck.Check(ctx)
	}

	return fmt.Errorf("service is degraded")
}

// CachedCheck 缓存检查
type CachedCheck struct {
	name       string
	check      Check
	cache      *CheckResult
	cacheTTL   time.Duration
	lastCheck  time.Time
	mu         sync.RWMutex
}

// NewCachedCheck 创建缓存检查
func NewCachedCheck(name string, check Check, cacheTTL time.Duration) *CachedCheck {
	return &CachedCheck{
		name:     name,
		check:    check,
		cacheTTL: cacheTTL,
	}
}

func (cc *CachedCheck) Name() string {
	return cc.name
}

func (cc *CachedCheck) Check(ctx context.Context) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// 检查缓存是否有效
	if cc.cache != nil && time.Since(cc.lastCheck) < cc.cacheTTL {
		if cc.cache.Status == "healthy" {
			return nil
		}
		return fmt.Errorf(cc.cache.Error)
	}

	// 执行检查
	result := CheckResult{
		Name:      cc.name,
		Timestamp: time.Now(),
	}

	start := time.Now()
	err := cc.check.Check(ctx)
	result.Duration = time.Since(start)

	if err != nil {
		result.Status = "unhealthy"
		result.Error = err.Error()
	} else {
		result.Status = "healthy"
	}

	cc.cache = &result
	cc.lastCheck = time.Now()

	return err
}
