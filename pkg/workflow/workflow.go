// Package workflow 提供工作流引擎功能。
package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// WorkflowEngine 工作流引擎
type WorkflowEngine struct {
	workflows  map[string]*Workflow
	executions map[string]*Execution
	scheduler  *DAGScheduler
	monitor    *WorkflowMonitor
	mu         sync.RWMutex
}

// NewWorkflowEngine 创建工作流引擎
func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows:  make(map[string]*Workflow),
		executions: make(map[string]*Execution),
		scheduler:  NewDAGScheduler(),
		monitor:   NewWorkflowMonitor(),
	}
}

// DefineWorkflow 定义工作流
func (we *WorkflowEngine) DefineWorkflow(workflow *Workflow) error {
	we.mu.Lock()
	defer we.mu.Unlock()

	we.workflows[workflow.ID] = workflow

	return nil
}

// Execute 执行工作流
func (we *WorkflowEngine) Execute(ctx context.Context, workflowID string, input interface{}) (*Execution, error) {
	we.mu.Lock()
	defer we.mu.Unlock()

	workflow, exists := we.workflows[workflowID]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	execution := &Execution{
		ID:        generateExecutionID(),
		Workflow:  workflowID,
		Status:    "running",
		Input:     input,
		StartedAt: time.Now(),
	}

	we.executions[execution.ID] = execution

	// 执行 DAG
	return execution, we.scheduler.Execute(ctx, workflow, execution)
}

// Workflow 工作流
type Workflow struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Tasks       []*Task      `json:"tasks"`
	DAG         *DAG         `json:"dag"`
	Timeout     time.Duration `json:"timeout"`
	RetryPolicy *RetryPolicy `json:"retry_policy"`
}

// Task 任务
type Task struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "http", "script", "container", "workflow"
	Dependencies []string               `json:"dependencies"`
	Config      map[string]interface{} `json:"config"`
	Timeout     time.Duration          `json:"timeout"`
	Retry       int                    `json:"retry"`
}

// DAG 有向无环图
type DAG struct {
	Nodes map[string]*Node
	Edges map[string][]string
}

// Node 节点
type Node struct {
	ID     string   `json:"id"`
	Task   *Task   `json:"task"`
	Deps   []string `json:"deps"`
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Backoff      time.Duration `json:"backoff"`
	Multiplier  float64       `json:"multiplier"`
}

// Execution 执行
type Execution struct {
	ID         string                 `json:"id"`
	Workflow   string                 `json:"workflow"`
	Status     string                 `json:"status"`
	Input      interface{}            `json:"input"`
	Output     interface{}            `json:"output"`
	StartedAt   time.Time              `json:"started_at"`
	FinishedAt  time.Time              `json:"finished_at"`
	Duration   time.Duration          `json:"duration"`
	Tasks      map[string]*TaskExecution `json:"tasks"`
	Error      string                 `json:"error,omitempty"`
}

// TaskExecution 任务执行
type TaskExecution struct {
	ID        string       `json:"id"`
	Task      string       `json:"task"`
	Status    string       `json:"status"`
	Input     interface{}  `json:"input"`
	Output    interface{}  `json:"output"`
	Started   time.Time    `json:"started"`
	Finished  time.Time    `json:"finished"`
	Duration  time.Duration `json:"duration"`
	Retries   int          `json:"retries"`
	Error     string       `json:"error,omitempty"`
}

// DAGScheduler DAG 调度器
type DAGScheduler struct {
	queue     chan *Node
	workers   int
	mu        sync.RWMutex
}

// NewDAGScheduler 创建 DAG 调度器
func NewDAGScheduler() *DAGScheduler {
	return &DAGScheduler{
		queue:   make(chan *Node, 100),
		workers: 10,
	}
}

// Execute 执行
func (ds *DAGScheduler) Execute(ctx context.Context, workflow *Workflow, execution *Execution) (*Execution, error) {
	// 拓扑排序
	sortedTasks, err := ds.topologicalSort(workflow)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		return execution, err
	}

	// 执行任务
	for _, taskID := range sortedTasks {
		taskExecution := &TaskExecution{
			ID:     generateTaskExecutionID(),
			Task:   taskID,
			Status: "running",
			Started: time.Now(),
		}

		execution.Tasks[taskID] = taskExecution

		// 执行任务
		result, err := ds.executeTask(ctx, workflow, taskID, execution.Input)
		taskExecution.Output = result
		taskExecution.Finished = time.Now()
		taskExecution.Duration = taskExecution.Finished.Sub(taskExecution.Started)

		if err != nil {
			taskExecution.Status = "failed"
			taskExecution.Error = err.Error()
		} else {
			taskExecution.Status = "completed"
		}
	}

	execution.Status = "completed"
	execution.FinishedAt = time.Now()
	execution.Duration = execution.FinishedAt.Sub(execution.StartedAt)

	return execution, nil
}

// topologicalSort 拓扑排序
func (ds *DAGScheduler) topologicalSort(workflow *Workflow) ([]string, error) {
	// Kahn 算法
	inDegree := make(map[string]int)
	for _, task := range workflow.Tasks {
		inDegree[task.ID] = len(task.Dependencies)
	}

	queue := make([]string, 0)
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	result := make([]string, 0)
	for len(queue) > 0 {
		taskID := queue[0]
		queue = queue[1:]
		result = append(result, taskID)

		// 减少依赖任务的入度
		for _, task := range workflow.Tasks {
			for _, dep := range task.Dependencies {
				if dep == taskID {
					inDegree[task.ID]--
					if inDegree[task.ID] == 0 {
						queue = append(queue, task.ID)
					}
				}
			}
		}
	}

	if len(result) != len(workflow.Tasks) {
		return nil, fmt.Errorf("cycle detected in workflow")
	}

	return result, nil
}

// executeTask 执行任务
func (ds *DAGScheduler) executeTask(ctx context.Context, workflow *Workflow, taskID string, input interface{}) (interface{}, error) {
	// 简化实现
	return fmt.Sprintf("executed task %s", taskID), nil
}

// WorkflowMonitor 工作流监控
type WorkflowMonitor struct {
	metrics map[string]*WorkflowMetrics
	mu      sync.RWMutex
}

// WorkflowMetrics 工作流指标
type WorkflowMetrics struct {
	TotalExecutions    int64         `json:"total_executions"`
	SuccessExecutions  int64         `json:"success_executions"`
	FailedExecutions   int64         `json:"failed_executions"`
	AvgDuration        time.Duration `json:"avg_duration"`
	LastExecution      time.Time     `json:"last_execution"`
}

// NewWorkflowMonitor 创建工作流监控
func NewWorkflowMonitor() *WorkflowMonitor {
	return &WorkflowMonitor{
		metrics: make(map[string]*WorkflowMetrics),
	}
}

// Record 记录
func (wm *WorkflowMonitor) Record(workflowID string, execution *Execution) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	metrics, exists := wm.metrics[workflowID]
	if !exists {
		metrics = &WorkflowMetrics{}
		wm.metrics[workflowID] = metrics
	}

	metrics.TotalExecutions++
	metrics.LastExecution = time.Now()

	if execution.Status == "completed" {
		metrics.SuccessExecutions++
	} else {
		metrics.FailedExecutions++
	}
}

// generateExecutionID 生成执行 ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// generateTaskExecutionID 生成任务执行 ID
func generateTaskExecutionID() string {
	return fmt.Sprintf("task_exec_%d", time.Now().UnixNano())
}
