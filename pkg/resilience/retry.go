// Package resilience 提供重试机制
package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// RetryPolicy 重试策略
type RetryPolicy struct {
	// 最大重试次数
	MaxRetries int
	// 初始延迟
	InitialDelay time.Duration
	// 最大延迟
	MaxDelay time.Duration
	// 延迟乘数
	BackoffFactor float64
	// 是否添加随机抖动
	Jitter bool
	// 重试条件判断
	ShouldRetry func(error) bool
}

// DefaultRetryPolicy 默认重试策略
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries:    3,
	InitialDelay:  100 * time.Millisecond,
	MaxDelay:      10 * time.Second,
	BackoffFactor: 2.0,
	Jitter:        true,
	ShouldRetry:   DefaultShouldRetry,
}

// DefaultShouldRetry 默认重试判断
func DefaultShouldRetry(err error) bool {
	return err != nil
}

// Retry 执行重试
func Retry(ctx context.Context, policy RetryPolicy, fn func() error) error {
	var lastErr error
	delay := policy.InitialDelay

	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}
	if policy.InitialDelay < 0 {
		policy.InitialDelay = 0
	}
	if policy.MaxDelay < 0 {
		policy.MaxDelay = 30 * time.Second
	}
	if policy.BackoffFactor <= 0 {
		policy.BackoffFactor = 1.0
	}
	if policy.ShouldRetry == nil {
		policy.ShouldRetry = DefaultShouldRetry
	}

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// 执行函数
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否应该重试
		if !policy.ShouldRetry(err) {
			return err
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// 最后一次尝试失败，不再延迟
		if attempt == policy.MaxRetries {
			break
		}

		// 计算延迟时间
		waitTime := delay
		if policy.Jitter {
			// 添加随机抖动，避免惊群效应
			waitTime = addJitter(delay)
		}

		// 等待延迟时间
		select {
		case <-time.After(waitTime):
			// 继续下一次尝试
		case <-ctx.Done():
			return ctx.Err()
		}

		// 计算下一次延迟（指数退避）
		delay = time.Duration(float64(delay) * policy.BackoffFactor)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", policy.MaxRetries+1, lastErr)
}

// RetryWithContext 带上下文的重试
func RetryWithContext[T any](ctx context.Context, policy RetryPolicy, fn func(context.Context) (T, error)) (T, error) {
	var result T
	var lastErr error
	delay := policy.InitialDelay

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		// 执行函数
		res, err := fn(ctx)
		if err == nil {
			return res, nil
		}

		lastErr = err
		result = res

		// 检查是否应该重试
		if !policy.ShouldRetry(err) {
			return result, err
		}

		// 检查上下文是否已取消
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		// 最后一次尝试失败，不再延迟
		if attempt == policy.MaxRetries {
			break
		}

		// 计算延迟时间
		waitTime := delay
		if policy.Jitter {
			waitTime = addJitter(delay)
		}

		// 等待延迟时间
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return result, ctx.Err()
		}

		// 计算下一次延迟
		delay = time.Duration(float64(delay) * policy.BackoffFactor)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
	}

	return result, fmt.Errorf("retry failed after %d attempts: %w", policy.MaxRetries+1, lastErr)
}

// addJitter 添加随机抖动
func addJitter(delay time.Duration) time.Duration {
	// 抖动范围：±25%
	jitter := time.Duration(float64(delay) * 0.25 * (rand.Float64()*2 - 1))
	return delay + jitter
}

// ExponentialBackoff 指数退避策略
func ExponentialBackoff(initialDelay time.Duration, maxDelay time.Duration, maxRetries int) RetryPolicy {
	return RetryPolicy{
		MaxRetries:    maxRetries,
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: 2.0,
		Jitter:        true,
		ShouldRetry:   DefaultShouldRetry,
	}
}

// LinearBackoff 线性退避策略
func LinearBackoff(initialDelay time.Duration, maxDelay time.Duration, maxRetries int) RetryPolicy {
	return RetryPolicy{
		MaxRetries:    maxRetries,
		InitialDelay:  initialDelay,
		MaxDelay:      maxDelay,
		BackoffFactor: 1.0,
		Jitter:        true,
		ShouldRetry:   DefaultShouldRetry,
	}
}

// RetryableError 可重试的错误
type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NewRetryableError 创建可重试错误
func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	var retryable *RetryableError
	return errors.As(err, &retryable)
}

// RetryIf 根据错误类型决定是否重试
func RetryIf(shouldRetry func(error) bool) RetryPolicy {
	policy := DefaultRetryPolicy
	policy.ShouldRetry = shouldRetry
	return policy
}

// WithRetryMetrics 带指标的重试
type RetryMetrics struct {
	Attempts      int
	TotalDuration time.Duration
	LastError     error
	Success       bool
}

// RetryWithMetrics 带指标收集的重试
func RetryWithMetrics(ctx context.Context, policy RetryPolicy, fn func() error) (*RetryMetrics, error) {
	start := time.Now()
	metrics := &RetryMetrics{}

	err := Retry(ctx, policy, func() error {
		metrics.Attempts++
		return fn()
	})

	metrics.TotalDuration = time.Since(start)
	metrics.LastError = err
	metrics.Success = err == nil

	return metrics, err
}

// FixedInterval 固定间隔重试
func FixedInterval(interval time.Duration, maxRetries int) RetryPolicy {
	return RetryPolicy{
		MaxRetries:    maxRetries,
		InitialDelay:  interval,
		MaxDelay:      interval,
		BackoffFactor: 1.0,
		Jitter:        false,
		ShouldRetry:   DefaultShouldRetry,
	}
}

// FibonacciBackoff 斐波那契退避策略
type FibonacciBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	MaxRetries   int
	Jitter       bool
}

// Next 计算下一次延迟
func (fb *FibonacciBackoff) Next(attempt int) time.Duration {
	if attempt <= 0 {
		return fb.InitialDelay
	}

	// 斐波那契数列
	fib := fibonacci(attempt)
	delay := time.Duration(fib) * fb.InitialDelay

	if delay > fb.MaxDelay {
		delay = fb.MaxDelay
	}

	if fb.Jitter {
		delay = addJitter(delay)
	}

	return delay
}

// fibonacci 计算斐波那契数
func fibonacci(n int) uint64 {
	if n <= 1 {
		return uint64(n)
	}

	var a, b uint64 = 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// RetryWithFibonacci 使用斐波那契退避重试
func RetryWithFibonacci(ctx context.Context, fb FibonacciBackoff, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= fb.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		if attempt == fb.MaxRetries {
			break
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		delay := fb.Next(attempt)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return fmt.Errorf("fibonacci retry failed after %d attempts: %w", fb.MaxRetries+1, lastErr)
}

// ContinuousRetry 持续重试（无上限）
func ContinuousRetry(ctx context.Context, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay

	for {
		err := fn()
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		// 指数退避，有上限
		delay = time.Duration(math.Min(float64(delay)*2.0, float64(30*time.Second)))
	}
}

// RetryOnSpecificErrors 只在特定错误时重试
func RetryOnSpecificErrors(errTypes []error) func(error) bool {
	return func(err error) bool {
		for _, errType := range errTypes {
			if errors.Is(err, errType) {
				return true
			}
		}
		return false
	}
}
