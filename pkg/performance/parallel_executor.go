package performance

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// ParallelExecutor 并行执行引擎
type ParallelExecutor struct {
	maxWorkers    int
	workerPool    chan *Worker
	taskQueue     chan *ParallelTask
	resultQueue   chan *TaskResult
	stats         *ParallelStats
	dependencyMap *DependencyGraph
	mu            sync.RWMutex
}

// Worker 工作线程
type Worker struct {
	id     int
	active atomic.Bool
}

// ParallelTask 并行任务
type ParallelTask struct {
	ID       string
	Node     types.Node
	Context  *ExecutionContext
	Depends  []string // 依赖的任务ID
	Priority int       // 优先级 (0=最高)
}

// TaskResult 任务结果
type TaskResult struct {
	TaskID    string
	Result    interface{}
	Error     error
	Duration  time.Duration
	Memory    uint64
}

// ExecutionContext 执行上下文（简化版，用于并行执行）
type ExecutionContext struct {
	Variables  map[string]interface{}
	Functions  map[string]*types.FunctionNode
	Stdout     *Buffer
	Stderr     *Buffer
	Stdin      *Buffer
	Parent     *ExecutionContext
}

// DependencyGraph 依赖图
type DependencyGraph struct {
	nodes map[string]*ParallelTask
	edges map[string][]string // taskID -> dependent task IDs
	mu    sync.RWMutex
}

// ParallelStats 并行执行统计
type ParallelStats struct {
	TasksTotal      int64
	TasksCompleted  int64
	TasksFailed     int64
	TasksSkipped    int64
	AvgDuration     time.Duration
	MaxDuration     time.Duration
	MinDuration     time.Duration
	PeakMemory      uint64
	ParallelismUtil float64 // 并行利用率
}

// NewParallelExecutor 创建并行执行器
func NewParallelExecutor(maxWorkers int) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 4 // 默认4个工作线程
	}

	return &ParallelExecutor{
		maxWorkers:    maxWorkers,
		workerPool:    make(chan *Worker, maxWorkers),
		taskQueue:     make(chan *ParallelTask, 100),
		resultQueue:   make(chan *TaskResult, 100),
		stats:         &ParallelStats{},
		dependencyMap: NewDependencyGraph(),
	}
}

// Execute 并行执行脚本
func (pe *ParallelExecutor) Execute(ctx context.Context, script *types.ScriptNode) (*TaskResult, error) {
	// 1. 构建任务图
	tasks, err := pe.buildTaskGraph(script)
	if err != nil {
		return nil, fmt.Errorf("failed to build task graph: %w", err)
	}

	// 2. 分析依赖关系
	pe.dependencyMap.Build(tasks)

	// 3. 启动工作线程
	var wg sync.WaitGroup
	for i := 0; i < pe.maxWorkers; i++ {
		worker := &Worker{
			id:     i,
			active: atomic.Bool{},
		}
		pe.workerPool <- worker
		wg.Add(1)

		go pe.workerLoop(ctx, worker, &wg)
	}

	// 4. 提交任务
	for _, task := range tasks {
		pe.taskQueue <- task
		atomic.AddInt64(&pe.stats.TasksTotal, 1)
	}

	// 5. 收集结果
	results := make([]*TaskResult, 0)
	close(pe.taskQueue) // 关闭任务队列，工作线程会自动退出

	// 等待所有 worker 完成
	wg.Wait()

	// 关闭结果队列
	close(pe.resultQueue)

	for result := range pe.resultQueue {
		results = append(results, result)

		if result.Error != nil {
			atomic.AddInt64(&pe.stats.TasksFailed, 1)
		} else {
			atomic.AddInt64(&pe.stats.TasksCompleted, 1)
		}
	}

	// 6. 计算统计
	pe.calculateStats(results)

	// 返回最终结果
	if len(results) > 0 {
		return results[len(results)-1], nil
	}

	return &TaskResult{
		Result: nil,
		Error:  fmt.Errorf("no tasks executed"),
	}, nil
}

// workerLoop 工作线程循环
func (pe *ParallelExecutor) workerLoop(ctx context.Context, worker *Worker, wg *sync.WaitGroup) {
	defer wg.Done()
	worker.active.Store(true)

	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-pe.taskQueue:
			if !ok {
				return
			}

			// 检查依赖是否满足
			if !pe.dependencyMap.Ready(task.ID) {
				// 依赖未满足，重新入队
				time.Sleep(10 * time.Millisecond)
				pe.taskQueue <- task
				continue
			}

			// 执行任务
			start := time.Now()
			result := pe.executeTask(task)
			duration := time.Since(start)

			pe.resultQueue <- &TaskResult{
				TaskID:   task.ID,
				Result:   result.Result,
				Error:    result.Error,
				Duration: duration,
			}

			// 更新依赖图（标记任务完成）
			pe.dependencyMap.Complete(task.ID)
		}
	}
}

// executeTask 执行单个任务
func (pe *ParallelExecutor) executeTask(task *ParallelTask) *TaskResult {
	// 简化版：直接执行节点
	// 实际应该执行字节码或调用原生执行引擎

	result := &TaskResult{
		TaskID: task.ID,
	}

	switch node := task.Node.(type) {
	case *types.CommandNode:
		// 执行命令
		output, err := pe.executeCommand(node, task.Context)
		result.Result = output
		result.Error = err
		result.Memory = 1024 // 估算的内存使用

	case *types.AssignmentNode:
		// 执行赋值
		err := pe.executeAssignment(node, task.Context)
		result.Error = err
		result.Memory = 512

	case *types.IfNode:
		// 执行条件语句
		output, err := pe.executeIf(node, task.Context)
		result.Result = output
		result.Error = err
		result.Memory = 2048

	case *types.ForNode:
		// 执行循环
		output, err := pe.executeFor(node, task.Context)
		result.Result = output
		result.Error = err
		result.Memory = 4096

	case *types.WhileNode:
		// 执行while循环
		output, err := pe.executeWhile(node, task.Context)
		result.Result = output
		result.Error = err
		result.Memory = 3072

	case *types.FunctionNode:
		// 定义函数
		err := pe.executeFunction(node, task.Context)
		result.Error = err
		result.Memory = 1024

	default:
		result.Error = fmt.Errorf("unsupported node type: %T", node)
	}

	return result
}

// executeCommand 执行命令
func (pe *ParallelExecutor) executeCommand(node *types.CommandNode, ctx *ExecutionContext) (string, error) {
	// TODO: 实际的命令执行
	return fmt.Sprintf("executed: %s", node.Name), nil
}

// executeAssignment 执行赋值
func (pe *ParallelExecutor) executeAssignment(node *types.AssignmentNode, ctx *ExecutionContext) error {
	if ctx.Variables == nil {
		ctx.Variables = make(map[string]interface{})
	}
	ctx.Variables[node.Name] = node.Value
	return nil
}

// executeIf 执行if语句
func (pe *ParallelExecutor) executeIf(node *types.IfNode, ctx *ExecutionContext) (interface{}, error) {
	// TODO: 实际的条件判断和执行
	return nil, nil
}

// executeFor 执行for循环
func (pe *ParallelExecutor) executeFor(node *types.ForNode, ctx *ExecutionContext) (interface{}, error) {
	// TODO: 实际的循环执行
	return nil, nil
}

// executeWhile 执行while循环
func (pe *ParallelExecutor) executeWhile(node *types.WhileNode, ctx *ExecutionContext) (interface{}, error) {
	// TODO: 实际的循环执行
	return nil, nil
}

// executeFunction 执行函数定义
func (pe *ParallelExecutor) executeFunction(node *types.FunctionNode, ctx *ExecutionContext) error {
	if ctx.Functions == nil {
		ctx.Functions = make(map[string]*types.FunctionNode)
	}
	ctx.Functions[node.Name] = node
	return nil
}

// buildTaskGraph 构建任务图
func (pe *ParallelExecutor) buildTaskGraph(script *types.ScriptNode) ([]*ParallelTask, error) {
	tasks := make([]*ParallelTask, 0)

	for i, node := range script.Nodes {
		task := &ParallelTask{
			ID:       fmt.Sprintf("task_%d", i),
			Node:     node,
			Context:  &ExecutionContext{},
			Priority: 0,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// calculateStats 计算统计信息
func (pe *ParallelExecutor) calculateStats(results []*TaskResult) {
	if len(results) == 0 {
		return
	}

	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = results[0].Duration
	var totalMemory uint64

	for _, result := range results {
		totalDuration += result.Duration
		totalMemory += result.Memory

		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
	}

	pe.stats.AvgDuration = totalDuration / time.Duration(len(results))
	pe.stats.MaxDuration = maxDuration
	pe.stats.MinDuration = minDuration
	pe.stats.PeakMemory = totalMemory

	// 计算并行利用率
	if pe.stats.MaxDuration > 0 {
		maxDurationNs := pe.stats.MaxDuration.Nanoseconds()
		totalDurationNs := totalDuration.Nanoseconds()
		expectedTotalNs := maxDurationNs * int64(len(results))
		if expectedTotalNs > 0 {
			pe.stats.ParallelismUtil = float64(totalDurationNs) / float64(expectedTotalNs)
		}
	}
}

// NewDependencyGraph 创建依赖图
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[string]*ParallelTask),
		edges: make(map[string][]string),
	}
}

// Build 构建依赖图
func (dg *DependencyGraph) Build(tasks []*ParallelTask) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	// 添加节点
	for _, task := range tasks {
		dg.nodes[task.ID] = task
		dg.edges[task.ID] = task.Depends
	}
}

// Ready 检查任务是否就绪（所有依赖都已完成）
func (dg *DependencyGraph) Ready(taskID string) bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	dependencies, exists := dg.edges[taskID]
	if !exists || len(dependencies) == 0 {
		return true // 无依赖，立即就绪
	}

	// 检查所有依赖是否都已完成
	for _, depID := range dependencies {
		if depTask, exists := dg.nodes[depID]; exists {
			if depTask != nil {
				return false // 依赖未完成
			}
		}
	}

	return true
}

// Complete 标记任务完成
func (dg *DependencyGraph) Complete(taskID string) {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if _, exists := dg.nodes[taskID]; exists {
		// 将节点设为nil表示已完成
		dg.nodes[taskID] = nil
	}
}

// HasCycle 检测循环依赖
func (dg *DependencyGraph) HasCycle() bool {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	var dfs func(taskID string) bool
	dfs = func(taskID string) bool {
		visited[taskID] = true
		recursionStack[taskID] = true

		dependencies, _ := dg.edges[taskID]
		for _, depID := range dependencies {
			if recursionStack[depID] {
				return true // 发现循环
			}
			if !visited[depID] {
				if dfs(depID) {
					return true
				}
			}
		}

		delete(recursionStack, taskID)
		return false
	}

	for taskID := range dg.nodes {
		if dg.nodes[taskID] != nil { // 未完成的任务
			if dfs(taskID) {
				return true
			}
		}
	}

	return false
}

// NewExecutionContext 创建执行上下文
func NewExecutionContext(parent *ExecutionContext) *ExecutionContext {
	ctx := &ExecutionContext{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]*types.FunctionNode),
		Stdout:    &Buffer{},
		Stderr:    &Buffer{},
		Stdin:     &Buffer{},
		Parent:    parent,
	}

	// 继承父上下文的变量
	if parent != nil {
		for k, v := range parent.Variables {
			ctx.Variables[k] = v
		}
		for k, v := range parent.Functions {
			ctx.Functions[k] = v
		}
	}

	return ctx
}

// GetStats 获取并行统计
func (pe *ParallelExecutor) GetStats() *ParallelStats {
	return pe.stats
}
