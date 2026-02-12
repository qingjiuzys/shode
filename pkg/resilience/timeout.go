// Package resilience 提供超时控制功能
package resilience

import (
	"context"
	"fmt"
	"time"
)

// TimeoutConfig 超时配置
type TimeoutConfig struct {
	// 默认超时时间
	DefaultTimeout time.Duration
	// 最小超时时间
	MinTimeout time.Duration
	// 最大超时时间
	MaxTimeout time.Duration
}

// DefaultTimeoutConfig 默认超时配置
var DefaultTimeoutConfig = TimeoutConfig{
	DefaultTimeout: 30 * time.Second,
	MinTimeout:     1 * time.Second,
	MaxTimeout:     5 * time.Minute,
}

// Timeout 执行带超时的函数
func Timeout(fn func() error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- fn()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// TimeoutWithContext 执行带超时的上下文函数
func TimeoutWithContext(ctx context.Context, fn func(context.Context) error, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	errChan := make(chan error, 1)

	go func() {
		errChan <- fn(ctx)
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operation timed out after %v", timeout)
	}
}

// TimeoutWithResult 执行带超时的函数并返回结果
func TimeoutWithResult[T any](fn func() (T, error), timeout time.Duration) (T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		value T
		err   error
	}

	resultChan := make(chan result, 1)

	go func() {
		value, err := fn()
		resultChan <- result{value, err}
	}()

	select {
	case res := <-resultChan:
		return res.value, res.err
	case <-ctx.Done():
		var zero T
		return zero, fmt.Errorf("operation timed out after %v", timeout)
	}
}

// TimeoutController 超时控制器
type TimeoutController struct {
	cfg TimeoutConfig
}

// NewTimeoutController 创建超时控制器
func NewTimeoutController(cfg TimeoutConfig) *TimeoutController {
	if cfg.DefaultTimeout <= 0 {
		cfg.DefaultTimeout = DefaultTimeoutConfig.DefaultTimeout
	}
	if cfg.MinTimeout <= 0 {
		cfg.MinTimeout = DefaultTimeoutConfig.MinTimeout
	}
	if cfg.MaxTimeout <= 0 {
		cfg.MaxTimeout = DefaultTimeoutConfig.MaxTimeout
	}

	return &TimeoutController{
		cfg: cfg,
	}
}

// Execute 执行带默认超时的函数
func (tc *TimeoutController) Execute(fn func() error) error {
	return Timeout(fn, tc.cfg.DefaultTimeout)
}

// ExecuteWithTimeout 执行带指定超时的函数
func (tc *TimeoutController) ExecuteWithTimeout(fn func() error, timeout time.Duration) error {
	// 限制超时时间在合理范围内
	if timeout < tc.cfg.MinTimeout {
		timeout = tc.cfg.MinTimeout
	}
	if timeout > tc.cfg.MaxTimeout {
		timeout = tc.cfg.MaxTimeout
	}

	return Timeout(fn, timeout)
}

// GetTimeout 获取超时时间
func (tc *TimeoutController) GetTimeout() time.Duration {
	return tc.cfg.DefaultTimeout
}

// SetTimeout 设置超时时间
func (tc *TimeoutController) SetTimeout(timeout time.Duration) {
	if timeout < tc.cfg.MinTimeout {
		timeout = tc.cfg.MinTimeout
	}
	if timeout > tc.cfg.MaxTimeout {
		timeout = tc.cfg.MaxTimeout
	}

	tc.cfg.DefaultTimeout = timeout
}

// AdaptiveTimeout 自适应超时
type AdaptiveTimeout struct {
	minTimeout      time.Duration
	maxTimeout      time.Duration
	currentTimeout  time.Duration
	successCount    int
	failureCount    int
	adjustmentFactor float64
}

// NewAdaptiveTimeout 创建自适应超时
func NewAdaptiveTimeout(minTimeout, maxTimeout, initialTimeout time.Duration) *AdaptiveTimeout {
	if minTimeout <= 0 {
		minTimeout = 1 * time.Second
	}
	if maxTimeout <= minTimeout {
		maxTimeout = minTimeout * 10
	}
	if initialTimeout < minTimeout {
		initialTimeout = minTimeout
	}
	if initialTimeout > maxTimeout {
		initialTimeout = maxTimeout
	}

	return &AdaptiveTimeout{
		minTimeout:      minTimeout,
		maxTimeout:      maxTimeout,
		currentTimeout:  initialTimeout,
		adjustmentFactor: 0.1, // 每次调整10%
	}
}

// Execute 执行带自适应超时的函数
func (at *AdaptiveTimeout) Execute(fn func() error) error {
	start := time.Now()
	err := Timeout(fn, at.currentTimeout)
	duration := time.Since(start)

	// 根据执行结果和耗时调整超时
	at.adjustTimeout(err, duration)

	return err
}

// adjustTimeout 调整超时时间
func (at *AdaptiveTimeout) adjustTimeout(err error, duration time.Duration) {
	if err != nil {
		// 执行失败，可能是超时导致的
		at.failureCount++
		at.successCount = 0

		// 增加超时时间
		newTimeout := time.Duration(float64(at.currentTimeout) * (1 + at.adjustmentFactor))
		if newTimeout <= at.maxTimeout {
			at.currentTimeout = newTimeout
		} else {
			at.currentTimeout = at.maxTimeout
		}
	} else {
		// 执行成功
		at.successCount++
		at.failureCount = 0

		// 如果执行时间远小于超时时间，适当减少超时
		if duration < at.currentTimeout/2 {
			newTimeout := time.Duration(float64(at.currentTimeout) * (1 - at.adjustmentFactor))
			if newTimeout >= at.minTimeout {
				at.currentTimeout = newTimeout
			} else {
				at.currentTimeout = at.minTimeout
			}
		}
	}
}

// GetCurrentTimeout 获取当前超时时间
func (at *AdaptiveTimeout) GetCurrentTimeout() time.Duration {
	return at.currentTimeout
}

// GetSuccessRate 获取成功率
func (at *AdaptiveTimeout) GetSuccessRate() float64 {
	total := at.successCount + at.failureCount
	if total == 0 {
		return 0
	}
	return float64(at.successCount) / float64(total)
}

// DeadlineChecker 截止时间检查器
type DeadlineChecker struct {
	deadline time.Time
}

// NewDeadlineChecker 创建截止时间检查器
func NewDeadlineChecker(deadline time.Time) *DeadlineChecker {
	return &DeadlineChecker{deadline: deadline}
}

// NewDeadlineCheckerFromTimeout 从超时时间创建截止时间检查器
func NewDeadlineCheckerFromTimeout(timeout time.Duration) *DeadlineChecker {
	return &DeadlineChecker{
		deadline: time.Now().Add(timeout),
	}
}

// IsExpired 检查是否已过期
func (dc *DeadlineChecker) IsExpired() bool {
	return time.Now().After(dc.deadline)
}

// Remaining 剩余时间
func (dc *DeadlineChecker) Remaining() time.Duration {
	remaining := time.Until(dc.deadline)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Deadline 获取截止时间
func (dc *DeadlineChecker) Deadline() time.Time {
	return dc.deadline
}

// Extend 延长截止时间
func (dc *DeadlineChecker) Extend(duration time.Duration) {
	dc.deadline = dc.deadline.Add(duration)
}

// StepTimeout 分步超时
type StepTimeout struct {
	steps   []time.Duration
	current int
}

// NewStepTimeout 创建分步超时
func NewStepTimeout(steps ...time.Duration) *StepTimeout {
	if len(steps) == 0 {
		steps = []time.Duration{30 * time.Second}
	}
	return &StepTimeout{
		steps:   steps,
		current: 0,
	}
}

// NextTimeout 下一步超时时间
func (st *StepTimeout) NextTimeout() time.Duration {
	if st.current >= len(st.steps) {
		return st.steps[len(st.steps)-1]
	}
	timeout := st.steps[st.current]
	st.current++
	return timeout
}

// Reset 重置步骤
func (st *StepTimeout) Reset() {
	st.current = 0
}

// HasMore 是否还有更多步骤
func (st *StepTimeout) HasMore() bool {
	return st.current < len(st.steps)
}

// Bulkhead 隔离仓模式（限制并发）
type Bulkhead struct {
	semaphore chan struct{}
	maxWorkers int
}

// NewBulkhead 创建隔离仓
func NewBulkhead(maxWorkers int) *Bulkhead {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}
	return &Bulkhead{
		semaphore:  make(chan struct{}, maxWorkers),
		maxWorkers: maxWorkers,
	}
}

// Execute 执行任务
func (b *Bulkhead) Execute(fn func() error) error {
	// 获取槽位
	b.semaphore <- struct{}{}
	defer func() { <-b.semaphore }()

	return fn()
}

// ExecuteWithContext 带上下文执行任务
func (b *Bulkhead) ExecuteWithContext(ctx context.Context, fn func(context.Context) error) error {
	select {
	case b.semaphore <- struct{}{}:
		defer func() { <-b.semaphore }()
		return fn(ctx)
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AvailableWorkers 可用工作线程数
func (b *Bulkhead) AvailableWorkers() int {
	return len(b.semaphore)
}

// MaxWorkers 最大工作线程数
func (b *Bulkhead) MaxWorkers() int {
	return b.maxWorkers
}

// TryExecute 尝试执行任务（非阻塞）
func (b *Bulkhead) TryExecute(fn func() error) error {
	select {
	case b.semaphore <- struct{}{}:
		defer func() { <-b.semaphore }()
		return fn()
	default:
		return fmt.Errorf("bulkhead: no available workers")
	}
}

// WaitForAvailable 等待可用工作线程
func (b *Bulkhead) WaitForAvailable(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case b.semaphore <- struct{}{}:
		// 立即释放，只是检查是否有槽位
		<-b.semaphore
		return nil
	case <-ctx.Done():
		return fmt.Errorf("bulkhead: timeout waiting for available worker")
	}
}
