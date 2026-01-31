// Package scheduler 提供任务调度功能。
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Task 任务
type Task struct {
	ID        string
	Name      string
	Handler   func(ctx context.Context) error
	Schedule  Schedule
	NextRun   time.Time
	LastRun   time.Time
	Running   bool
	Enabled   bool
	Metadata  map[string]interface{}
}

// Schedule 调度计划
type Schedule interface {
	Next(lastRun time.Time) time.Time
}

// IntervalSchedule 间隔调度
type IntervalSchedule struct {
	Interval time.Duration
}

// Next 下次运行时间
func (is *IntervalSchedule) Next(lastRun time.Time) time.Time {
	return lastRun.Add(is.Interval)
}

// CronSchedule Cron 表达式调度（简化实现）
type CronSchedule struct {
	Expression string
	// 简化实现，实际应该解析 cron 表达式
	Hour   int
	Minute int
}

// Next 下次运行时间
func (cs *CronSchedule) Next(lastRun time.Time) time.Time {
	next := lastRun.Add(1 * time.Minute)
	// 简化实现，只设置小时和分钟
	if cs.Hour >= 0 && cs.Hour < 24 {
		next = time.Date(next.Year(), next.Month(), next.Day(), cs.Hour, cs.Minute, 0, 0, next.Location())
		// 如果已过今天，设置到明天
		if next.Before(time.Now()) {
			next = next.Add(24 * time.Hour)
		}
	}
	return next
}

// Scheduler 调度器
type Scheduler struct {
	tasks    map[string]*Task
	mu       sync.RWMutex
	running  bool
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewScheduler 创建调度器
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks:    make(map[string]*Task),
		stopChan: make(chan struct{}),
	}
}

// AddTask 添加任务
func (s *Scheduler) AddTask(id, name string, handler func(ctx context.Context) error, schedule Schedule) *Task {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		ID:       id,
		Name:     name,
		Handler:  handler,
		Schedule: schedule,
		NextRun:  schedule.Next(time.Now()),
		Enabled:  true,
		Metadata: make(map[string]interface{}),
	}

	s.tasks[id] = task
	return task
}

// RemoveTask 移除任务
func (s *Scheduler) RemoveTask(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, id)
}

// GetTask 获取任务
func (s *Scheduler) GetTask(id string) (*Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	return task, exists
}

// ListTasks 列出所有任务
func (s *Scheduler) ListTasks() []*Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// Start 启动调度器
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.runLoop()
}

// Stop 停止调度器
func (s *Scheduler) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	s.stopChan = make(chan struct{})
}

// runLoop 运行循环
func (s *Scheduler) runLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRunTasks()
		}
	}
}

// checkAndRunTasks 检查并运行任务
func (s *Scheduler) checkAndRunTasks() {
	now := time.Now()

	s.mu.RLock()
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if task.Enabled && !task.Running && now.After(task.NextRun) {
			tasks = append(tasks, task)
		}
	}
	s.mu.RUnlock()

	for _, task := range tasks {
		s.runTask(task)
	}
}

// runTask 运行任务
func (s *Scheduler) runTask(task *Task) {
	s.mu.Lock()
	task.Running = true
	task.LastRun = time.Now()
	task.NextRun = task.Schedule.Next(task.LastRun)
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			task.Running = false
			s.mu.Unlock()
		}()

		ctx := context.Background()
		if err := task.Handler(ctx); err != nil {
			fmt.Printf("Task %s failed: %v\n", task.ID, err)
		}
	}()
}

// TaskDependency 任务依赖
type TaskDependency struct {
	TaskID      string
	DependsOn   []string
	WaitForAll  bool
}

// DependencyScheduler 依赖调度器
type DependencyScheduler struct {
	scheduler   *Scheduler
	dependencies map[string]*TaskDependency
	mu          sync.RWMutex
}

// NewDependencyScheduler 创建依赖调度器
func NewDependencyScheduler(scheduler *Scheduler) *DependencyScheduler {
	return &DependencyScheduler{
		scheduler:    scheduler,
		dependencies: make(map[string]*TaskDependency),
	}
}

// AddDependency 添加依赖
func (ds *DependencyScheduler) AddDependency(taskID string, dependsOn []string, waitForAll bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.dependencies[taskID] = &TaskDependency{
		TaskID:     taskID,
		DependsOn:  dependsOn,
		WaitForAll: waitForAll,
	}
}

// CheckDependencies 检查依赖是否满足
func (ds *DependencyScheduler) CheckDependencies(taskID string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	dep, exists := ds.dependencies[taskID]
	if !exists || len(dep.DependsOn) == 0 {
		return true
	}

	for _, depID := range dep.DependsOn {
		task, exists := ds.scheduler.GetTask(depID)
		if !exists {
			continue
		}

		// 如果依赖任务从未运行过，等待
		if task.LastRun.IsZero() {
			return false
		}

		// 如果是等待所有，检查是否所有依赖都已完成
		if dep.WaitForAll && task.Running {
			return false
		}
	}

	return true
}

// DistributedTask 分布式任务
type DistributedTask struct {
	ID       string
	Name     string
	Handler  func(ctx context.Context) error
	Nodes    []string
	Assigned string
	Status   string
}

// DistributedScheduler 分布式调度器
type DistributedScheduler struct {
	scheduler *Scheduler
	tasks     map[string]*DistributedTask
	mu        sync.RWMutex
	nodeID    string
}

// NewDistributedScheduler 创建分布式调度器
func NewDistributedScheduler(scheduler *Scheduler, nodeID string) *DistributedScheduler {
	return &DistributedScheduler{
		scheduler: scheduler,
		tasks:     make(map[string]*DistributedTask),
		nodeID:    nodeID,
	}
}

// AssignTask 分配任务
func (ds *DistributedScheduler) AssignTask(task *DistributedTask) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// 简化的负载均衡：轮询分配
	if len(task.Nodes) == 0 {
		return fmt.Errorf("no available nodes")
	}

	// 选择第一个节点（实际应该用更复杂的策略）
	task.Assigned = task.Nodes[0]
	task.Status = "assigned"

	ds.tasks[task.ID] = task
	return nil
}

// ExecuteTask 执行任务
func (ds *DistributedScheduler) ExecuteTask(taskID string) error {
	ds.mu.Lock()
	task, exists := ds.tasks[taskID]
	if !exists {
		ds.mu.Unlock()
		return fmt.Errorf("task not found")
	}
	ds.mu.Unlock()

	// 检查是否分配给当前节点
	if task.Assigned != ds.nodeID {
		return nil
	}

	// 执行任务
	ctx := context.Background()
	if err := task.Handler(ctx); err != nil {
		ds.mu.Lock()
		task.Status = "failed"
		ds.mu.Unlock()
		return err
	}

	ds.mu.Lock()
	task.Status = "completed"
	ds.mu.Unlock()

	return nil
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

// DefaultRetryPolicy 默认重试策略
var DefaultRetryPolicy = RetryPolicy{
	MaxRetries:    3,
	InitialDelay:  1 * time.Second,
	MaxDelay:      60 * time.Second,
	BackoffFactor: 2.0,
}

// RetryScheduler 重试调度器
type RetryScheduler struct {
	scheduler    *Scheduler
	retryPolicies map[string]*RetryPolicy
	mu           sync.RWMutex
}

// NewRetryScheduler 创建重试调度器
func NewRetryScheduler(scheduler *Scheduler) *RetryScheduler {
	return &RetryScheduler{
		scheduler:      scheduler,
		retryPolicies: make(map[string]*RetryPolicy),
	}
}

// SetRetryPolicy 设置重试策略
func (rs *RetryScheduler) SetRetryPolicy(taskID string, policy *RetryPolicy) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.retryPolicies[taskID] = policy
}

// ExecuteWithRetry 执行任务（带重试）
func (rs *RetryScheduler) ExecuteWithRetry(taskID string, handler func(ctx context.Context) error) error {
	rs.mu.RLock()
	policy, exists := rs.retryPolicies[taskID]
	if !exists {
		policy = &DefaultRetryPolicy
	}
	rs.mu.RUnlock()

	var lastErr error
	delay := policy.InitialDelay

	for attempt := 0; attempt <= policy.MaxRetries; attempt++ {
		ctx := context.Background()
		if err := handler(ctx); err != nil {
			lastErr = err

			if attempt < policy.MaxRetries {
				fmt.Printf("Task %s failed (attempt %d), retrying in %v: %v\n", taskID, attempt+1, delay, err)
				time.Sleep(delay)

				// 指数退避
				delay = time.Duration(float64(delay) * policy.BackoffFactor)
				if delay > policy.MaxDelay {
					delay = policy.MaxDelay
				}
				continue
			}
		} else {
			return nil
		}
	}

	return fmt.Errorf("task %s failed after %d attempts: %w", taskID, policy.MaxRetries, lastErr)
}

// OnceScheduler 一次性调度器
type OnceScheduler struct {
	scheduler *Scheduler
	onceTasks map[string]bool
	mu        sync.RWMutex
}

// NewOnceScheduler 创建一次性调度器
func NewOnceScheduler(scheduler *Scheduler) *OnceScheduler {
	return &OnceScheduler{
		scheduler: scheduler,
		onceTasks: make(map[string]bool),
	}
}

// RunOnce 运行一次性任务
func (os *OnceScheduler) RunOnce(id, name string, handler func(ctx context.Context) error) error {
	os.mu.Lock()
	defer os.mu.Unlock()

	// 检查是否已运行
	if os.onceTasks[id] {
		return fmt.Errorf("task already executed: %s", id)
	}

	os.onceTasks[id] = true

	// 执行任务
	ctx := context.Background()
	return handler(ctx)
}

// HasRun 检查任务是否已运行
func (os *OnceScheduler) HasRun(id string) bool {
	os.mu.RLock()
	defer os.mu.RUnlock()

	return os.onceTasks[id]
}
