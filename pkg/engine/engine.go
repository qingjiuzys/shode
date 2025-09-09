package engine

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// ExecutionMode represents the execution mode for commands
type ExecutionMode int

const (
	ModeInterpreted ExecutionMode = iota // Interpret built-in functions
	ModeProcess                          // Execute external processes
	ModeHybrid                           // Smart hybrid execution
)

// ExecutionEngine is the core engine for executing shell commands
type ExecutionEngine struct {
	envManager  *environment.EnvironmentManager
	stdlib      *stdlib.StdLib
	moduleMgr   *module.ModuleManager
	security    *sandbox.SecurityChecker
	processPool *ProcessPool
	cache       *CommandCache
}

// ExecutionResult represents the result of executing an AST
type ExecutionResult struct {
	Success    bool
	ExitCode   int
	Output     string
	Error      string
	Duration   time.Duration
	Commands   []*CommandResult
}

// CommandResult represents the result of a single command execution
type CommandResult struct {
	Command   *types.CommandNode
	Success   bool
	ExitCode  int
	Output    string
	Error     string
	Duration  time.Duration
	Mode      ExecutionMode
}

// PipelineResult represents the result of pipeline execution
type PipelineResult struct {
	Success  bool
	ExitCode int
	Output   string
	Error    string
	Results  []*CommandResult
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(
	envManager *environment.EnvironmentManager,
	stdlib *stdlib.StdLib,
	moduleMgr *module.ModuleManager,
	security *sandbox.SecurityChecker,
) *ExecutionEngine {
	return &ExecutionEngine{
		envManager: envManager,
		stdlib:     stdlib,
		moduleMgr:  moduleMgr,
		security:   security,
		processPool: NewProcessPool(10, 30*time.Second),
		cache:       NewCommandCache(1000),
	}
}

// Execute executes a complete script
func (ee *ExecutionEngine) Execute(ctx context.Context, script *types.ScriptNode) (*ExecutionResult, error) {
	startTime := time.Now()
	
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0, len(script.Nodes)),
	}

	for _, node := range script.Nodes {
		switch n := node.(type) {
		case *types.CommandNode:
			cmdResult, err := ee.ExecuteCommand(ctx, n)
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, cmdResult)
			
			if !cmdResult.Success {
				result.Success = false
				result.ExitCode = cmdResult.ExitCode
				break
			}

		case *types.PipeNode:
			// For now, treat PipeNode as a simple command sequence
			// TODO: Implement proper pipeline execution
			leftResult, err := ee.ExecuteCommand(ctx, n.Left.(*types.CommandNode))
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, leftResult)
			
			if !leftResult.Success {
				result.Success = false
				result.ExitCode = leftResult.ExitCode
				break
			}

			rightResult, err := ee.ExecuteCommand(ctx, n.Right.(*types.CommandNode))
			if err != nil {
				return nil, err
			}
			result.Commands = append(result.Commands, rightResult)
			
			if !rightResult.Success {
				result.Success = false
				result.ExitCode = rightResult.ExitCode
				break
			}

		// TODO: Add support for other node types (if, for, while, etc.)
		default:
			return nil, fmt.Errorf("unsupported node type: %T", n)
		}
	}

	result.Duration = time.Since(startTime)
	result.Success = true
	return result, nil
}

// ExecuteCommand executes a single command
func (ee *ExecutionEngine) ExecuteCommand(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	startTime := time.Now()

	// Security check
	if err := ee.security.CheckCommand(cmd); err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    fmt.Sprintf("Security violation: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// Decide execution mode
	mode := ee.decideExecutionMode(cmd)

	var result *CommandResult
	var err error

	switch mode {
	case ModeInterpreted:
		result, err = ee.executeInterpreted(ctx, cmd)
	case ModeProcess:
		result, err = ee.executeProcess(ctx, cmd)
	case ModeHybrid:
		result, err = ee.executeHybrid(ctx, cmd)
	default:
		return nil, fmt.Errorf("unknown execution mode: %v", mode)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	result.Mode = mode
	return result, nil
}

// ExecutePipeline executes a pipeline of commands (placeholder)
func (ee *ExecutionEngine) ExecutePipeline(ctx context.Context, pipeline *types.PipeNode) (*PipelineResult, error) {
	// For now, treat pipeline as sequential execution
	// TODO: Implement proper pipeline with stream processing
	
	// Execute left command
	leftResult, err := ee.ExecuteCommand(ctx, pipeline.Left.(*types.CommandNode))
	if err != nil {
		return nil, err
	}

	// Execute right command
	rightResult, err := ee.ExecuteCommand(ctx, pipeline.Right.(*types.CommandNode))
	if err != nil {
		return nil, err
	}

	results := []*CommandResult{leftResult, rightResult}
	success := leftResult.Success && rightResult.Success
	exitCode := 0
	if !success {
		exitCode = 1
	}

	// Combine output (simple concatenation for now)
	output := leftResult.Output + rightResult.Output

	return &PipelineResult{
		Success:  success,
		ExitCode: exitCode,
		Output:   output,
		Error:    "",
		Results:  results,
	}, nil
}

// decideExecutionMode determines the best execution mode for a command
func (ee *ExecutionEngine) decideExecutionMode(cmd *types.CommandNode) ExecutionMode {
	// Check if it's a standard library function
	if ee.isStdLibFunction(cmd.Name) {
		return ModeInterpreted
	}

	// Check if it's a module export (TODO: implement module export check)
	// if ee.moduleMgr.IsExportedFunction(cmd.Name) {
	//     return ModeInterpreted
	// }

	// Check if external command exists
	if ee.isExternalCommandAvailable(cmd.Name) {
		return ModeProcess
	}

	// Default to process execution
	return ModeProcess
}

// isStdLibFunction checks if a function exists in standard library
func (ee *ExecutionEngine) isStdLibFunction(funcName string) bool {
	return ee.stdlib.HasFunction(funcName)
}

// executeInterpreted executes a command using the interpreter (built-in functions)
func (ee *ExecutionEngine) executeInterpreted(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Execute using standard library
	result, err := ee.executeStdLibFunction(cmd.Name, cmd.Args)
	if err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    err.Error(),
		}, nil
	}

	return &CommandResult{
		Command:  cmd,
		Success:  true,
		ExitCode: 0,
		Output:   result,
	}, nil
}

// executeStdLibFunction executes a standard library function
func (ee *ExecutionEngine) executeStdLibFunction(funcName string, args []string) (string, error) {
	// Convert string arguments to interface{} for the executor
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg
	}

	// Use the standard library's ExecuteFunction method
	result, err := ee.stdlib.ExecuteFunction(funcName, interfaceArgs...)
	if err != nil {
		return "", err
	}

	// Convert result to string
	switch v := result.(type) {
	case string:
		return v, nil
	case []string:
		return strings.Join(v, "\n"), nil
	case bool:
		return fmt.Sprintf("%v", v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case int64:
		return fmt.Sprintf("%d", v), nil
	case time.Time:
		return v.Format(time.RFC3339), nil
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// executeProcess executes a command as an external process
func (ee *ExecutionEngine) executeProcess(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Check cache first
	if cached, ok := ee.cache.Get(cmd.Name, cmd.Args); ok {
		return cached, nil
	}

	// Create command with context
	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	// Set environment - convert map[string]string to []string
	envVars := make([]string, 0, len(ee.envManager.GetAllEnv()))
	for key, value := range ee.envManager.GetAllEnv() {
		envVars = append(envVars, key+"="+value)
	}
	command.Env = envVars

	// Set working directory
	command.Dir = ee.envManager.GetWorkingDir()

	// Capture output
	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Execute command
	startTime := time.Now()
	err := command.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	result := &CommandResult{
		Command:  cmd,
		Success:  err == nil,
		ExitCode: exitCode,
		Output:   stdout.String(),
		Error:    stderr.String(),
		Duration: duration,
	}

	// Cache successful results
	if err == nil {
		ee.cache.Put(cmd.Name, cmd.Args, result)
	}

	return result, nil
}

// executeHybrid executes a command using hybrid approach (future enhancement)
func (ee *ExecutionEngine) executeHybrid(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// For now, default to process execution
	// TODO: Implement smart hybrid execution logic
	return ee.executeProcess(ctx, cmd)
}

// isExternalCommandAvailable checks if an external command exists
func (ee *ExecutionEngine) isExternalCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Helper function to convert error to string
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
