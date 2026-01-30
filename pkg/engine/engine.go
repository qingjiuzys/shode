package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/errors"
	"gitee.com/com_818cloud/shode/pkg/metrics"
	"gitee.com/com_818cloud/shode/pkg/module"
	shodeparser "gitee.com/com_818cloud/shode/pkg/parser"
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
	envManager     *environment.EnvironmentManager
	stdlib         *stdlib.StdLib
	moduleMgr      *module.ModuleManager
	security       *sandbox.SecurityChecker
	processPool    *ProcessPool
	cache          *CommandCache
	functions      map[string]*types.FunctionNode // User-defined functions
	metrics        *metrics.MetricsCollector      // Performance metrics collector
	backgroundJobs map[int]*exec.Cmd              // Background jobs (PID -> Cmd)
	jobCounter     int                            // Counter for job IDs
	arrays         map[string][]string            // Array variables
}

// ExecutionResult represents the result of executing an AST
type ExecutionResult struct {
	Success      bool
	ExitCode     int
	Output       string
	Error        string
	Duration     time.Duration
	Commands     []*CommandResult
	BreakFlag    bool // Set to true if break statement was encountered
	ContinueFlag bool // Set to true if continue statement was encountered
}

// CommandResult represents the result of a single command execution
type CommandResult struct {
	Command  *types.CommandNode
	Success  bool
	ExitCode int
	Output   string
	Error    string
	Duration time.Duration
	Mode     ExecutionMode
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
		envManager:     envManager,
		stdlib:         stdlib,
		moduleMgr:      moduleMgr,
		security:       security,
		processPool:    NewProcessPool(10, 30*time.Second),
		cache:          NewCommandCache(1000),
		functions:      make(map[string]*types.FunctionNode),
		metrics:        metrics.NewMetricsCollector(),
		backgroundJobs: make(map[int]*exec.Cmd),
		jobCounter:     0,
		arrays:         make(map[string][]string),
	}
}

// Helper methods for Execute function refactoring

// executeCommandNode handles CommandNode execution including special commands
func (ee *ExecutionEngine) executeCommandNode(ctx context.Context, node *types.CommandNode, result *ExecutionResult) (*ExecutionResult, error) {
	// Check for break/continue commands
	if node.Name == "break" {
		result.BreakFlag = true
		result.Success = true
		return result, nil
	}
	if node.Name == "continue" {
		result.ContinueFlag = true
		result.Success = true
		return result, nil
	}

	// Check for Source command - load and execute another script file
	if node.Name == "Source" {
		return ee.executeSourceCommand(ctx, node, result)
	}

	cmdResult, err := ee.ExecuteCommand(ctx, node)
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, cmdResult)

	// Collect command output
	if cmdResult.Output != "" {
		result.Output += cmdResult.Output
		if !strings.HasSuffix(cmdResult.Output, "\n") {
			result.Output += "\n"
		}
	}

	if !cmdResult.Success {
		result.Success = false
		result.ExitCode = cmdResult.ExitCode
	}
	return result, nil
}

// executeSourceCommand handles the Source command to load external scripts
func (ee *ExecutionEngine) executeSourceCommand(ctx context.Context, node *types.CommandNode, result *ExecutionResult) (*ExecutionResult, error) {
	if len(node.Args) == 0 {
		result.Success = false
		result.ExitCode = 1
		result.Error = "Source requires a file path argument"
		return result, nil
	}
	// Expand the file path
	filePath := ee.expandVariables(node.Args[0])
	// Parse and execute the source file
	sourceResult, err := ee.executeSourceFile(ctx, filePath)
	if err != nil {
		result.Success = false
		result.ExitCode = 1
		result.Error = fmt.Sprintf("Source error: %v", err)
		return result, nil
	}
	result.Commands = append(result.Commands, sourceResult.Commands...)
	if !sourceResult.Success {
		result.Success = false
		result.ExitCode = sourceResult.ExitCode
		result.Error = sourceResult.Error
	}
	return result, nil
}

// executePipeNode handles PipeNode execution
func (ee *ExecutionEngine) executePipeNode(ctx context.Context, node *types.PipeNode, result *ExecutionResult) (*ExecutionResult, error) {
	pipeResult, err := ee.ExecutePipeline(ctx, node)
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, pipeResult.Results...)

	// Collect pipeline output
	if pipeResult.Output != "" {
		result.Output += pipeResult.Output
	}

	if !pipeResult.Success {
		result.Success = false
		result.ExitCode = pipeResult.ExitCode
	}
	return result, nil
}

// executeIfNode handles IfNode execution
func (ee *ExecutionEngine) executeIfNode(ctx context.Context, node *types.IfNode, result *ExecutionResult) (*ExecutionResult, error) {
	ifResult, err := ee.ExecuteIf(ctx, node)
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, ifResult.Commands...)

	// Collect if statement output
	if ifResult.Output != "" {
		result.Output += ifResult.Output
	}

	if !ifResult.Success {
		result.Success = false
		result.ExitCode = ifResult.ExitCode
	}
	return result, nil
}

// executeForNode handles ForNode execution
func (ee *ExecutionEngine) executeForNode(ctx context.Context, node *types.ForNode, result *ExecutionResult) (*ExecutionResult, error) {
	forResult, err := ee.ExecuteFor(ctx, node)
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, forResult.Commands...)

	// Collect for loop output
	if forResult.Output != "" {
		result.Output += forResult.Output
	}

	if !forResult.Success {
		result.Success = false
		result.ExitCode = forResult.ExitCode
	}
	return result, nil
}

// executeWhileNode handles WhileNode execution
func (ee *ExecutionEngine) executeWhileNode(ctx context.Context, node *types.WhileNode, result *ExecutionResult) (*ExecutionResult, error) {
	whileResult, err := ee.ExecuteWhile(ctx, node)
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, whileResult.Commands...)

	// Collect while loop output
	if whileResult.Output != "" {
		result.Output += whileResult.Output
	}

	if !whileResult.Success {
		result.Success = false
		result.ExitCode = whileResult.ExitCode
	}
	return result, nil
}

// executeAssignmentNode handles AssignmentNode execution
func (ee *ExecutionEngine) executeAssignmentNode(node *types.AssignmentNode) error {
	// First, expand variables in the value
	expandedValue := ee.expandVariables(node.Value)

	// Check if this is an array assignment: (value1 value2 ...)
	trimmedValue := strings.TrimSpace(expandedValue)
	if strings.HasPrefix(trimmedValue, "(") && strings.HasSuffix(trimmedValue, ")") {
		return ee.executeArrayAssignment(node.Name, trimmedValue)
	}

	// Handle command substitution and simple assignment
	if !strings.HasPrefix(trimmedValue, "\"") && !strings.HasPrefix(trimmedValue, "'") {
		if strings.HasPrefix(trimmedValue, "$(") || strings.HasPrefix(trimmedValue, "`") {
			expandedValue = ee.executeCommandSubstitutionInAssignment(trimmedValue)
		}
	}
	ee.envManager.SetEnv(node.Name, expandedValue)
	return nil
}

// executeArrayAssignment handles array assignment like arr=(1 2 3)
func (ee *ExecutionEngine) executeArrayAssignment(name, trimmedValue string) error {
	arrayContent := trimmedValue[1 : len(trimmedValue)-1] // Remove ( and )
	values := ee.parseArrayElements(arrayContent)

	// Store array as space-separated string
	arrayValue := strings.Join(values, " ")
	ee.envManager.SetEnv(name, arrayValue)

	// Also store individual elements as name[0], name[1], etc.
	for i, val := range values {
		key := fmt.Sprintf("%s[%d]", name, i)
		ee.envManager.SetEnv(key, val)
	}

	// Store array length
	lengthKey := fmt.Sprintf("%s[@]", name)
	ee.envManager.SetEnv(lengthKey, fmt.Sprintf("%d", len(values)))
	return nil
}

// executeCommandSubstitutionInAssignment executes command substitution for assignment values
func (ee *ExecutionEngine) executeCommandSubstitutionInAssignment(trimmedValue string) string {
	p := shodeparser.NewSimpleParser()
	script, err := p.ParseString(trimmedValue)
	if err == nil && len(script.Nodes) > 0 {
		cmdResult, execErr := ee.Execute(context.Background(), script)
		if execErr == nil && cmdResult != nil && cmdResult.Success {
			return strings.TrimSpace(cmdResult.Output)
		}
	}
	return trimmedValue
}

// executeAndNode handles AndNode (&&) execution
func (ee *ExecutionEngine) executeAndNode(ctx context.Context, node *types.AndNode, result *ExecutionResult) (*ExecutionResult, error) {
	// Execute left side first
	leftResult, err := ee.ExecuteCommand(ctx, types.CastToCommandNode(node.Left))
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, leftResult)

	// Collect left output
	ee.collectOutput(result, leftResult.Output)

	// If left side succeeded, execute right side
	if leftResult.Success {
		rightResult, err := ee.ExecuteCommand(ctx, types.CastToCommandNode(node.Right))
		if err != nil {
			return nil, err
		}
		result.Commands = append(result.Commands, rightResult)

		// Collect right output
		ee.collectOutput(result, rightResult.Output)

		// Overall success depends on right side
		if !rightResult.Success {
			result.Success = false
			result.ExitCode = rightResult.ExitCode
			result.Error = rightResult.Error
		}
	} else {
		// Left side failed, skip right side
		result.ExitCode = leftResult.ExitCode
		result.Error = leftResult.Error
	}
	return result, nil
}

// executeOrNode handles OrNode (||) execution
func (ee *ExecutionEngine) executeOrNode(ctx context.Context, node *types.OrNode, result *ExecutionResult) (*ExecutionResult, error) {
	// Execute left side first
	leftResult, err := ee.ExecuteCommand(ctx, types.CastToCommandNode(node.Left))
	if err != nil {
		return nil, err
	}
	result.Commands = append(result.Commands, leftResult)

	// Collect left output
	ee.collectOutput(result, leftResult.Output)

	// If left side failed, execute right side
	if !leftResult.Success {
		rightResult, err := ee.ExecuteCommand(ctx, types.CastToCommandNode(node.Right))
		if err != nil {
			return nil, err
		}
		result.Commands = append(result.Commands, rightResult)

		// Collect right output
		ee.collectOutput(result, rightResult.Output)

		// Overall success depends on right side
		if !rightResult.Success {
			result.Success = false
			result.ExitCode = rightResult.ExitCode
			result.Error = rightResult.Error
		}
	}
	// If left succeeded, we're done (success is already true)
	return result, nil
}

// executeBackgroundNode handles BackgroundNode execution
func (ee *ExecutionEngine) executeBackgroundNode(ctx context.Context, node *types.BackgroundNode, result *ExecutionResult) (*ExecutionResult, error) {
	cmdNode := types.CastToCommandNode(node.Command)
	if cmdNode == nil {
		return nil, errors.NewExecutionError(errors.ErrExecutionFailed,
			"background command must be a CommandNode").
			WithContext("command_type", fmt.Sprintf("%T", node.Command))
	}

	bgResult, err := ee.ExecuteCommand(ctx, cmdNode)
	if err != nil {
		return nil, err
	}

	// For now, execute synchronously but mark as background
	result.Commands = append(result.Commands, bgResult)

	// Collect background command output
	ee.collectOutput(result, bgResult.Output)

	if !bgResult.Success {
		result.Success = false
		result.ExitCode = bgResult.ExitCode
		result.Error = bgResult.Error
	}
	return result, nil
}

// executeHeredocNode handles HeredocNode execution
func (ee *ExecutionEngine) executeHeredocNode(ctx context.Context, node *types.HeredocNode, result *ExecutionResult) (*ExecutionResult, error) {
	// The heredoc body from tree-sitter should NOT contain the end marker
	heredocBody := strings.TrimRight(node.Body, "\n")

	cmdNode := types.CastToCommandNode(node.Command)
	if cmdNode == nil {
		return nil, errors.NewExecutionError(errors.ErrExecutionFailed,
			"heredoc command must be a CommandNode").
			WithContext("command_type", fmt.Sprintf("%T", node.Command))
	}

	cmdResult, err := ee.ExecuteCommandWithInput(ctx, cmdNode, heredocBody+"\n")
	if err != nil {
		return nil, err
	}

	result.Commands = append(result.Commands, cmdResult)
	if !cmdResult.Success {
		result.Success = false
		result.ExitCode = cmdResult.ExitCode
		result.Error = cmdResult.Error
	} else {
		result.Output = cmdResult.Output
	}
	return result, nil
}

// executeArrayNode handles ArrayNode execution
func (ee *ExecutionEngine) executeArrayNode(node *types.ArrayNode) {
	// Store array as a space-separated string in environment
	arrayValue := strings.Join(node.Values, " ")
	ee.envManager.SetEnv(node.Name, arrayValue)

	// Also store individual elements as array_name[0], array_name[1], etc.
	for i, val := range node.Values {
		key := fmt.Sprintf("%s[%d]", node.Name, i)
		ee.envManager.SetEnv(key, val)
	}

	// Store array length
	lengthKey := fmt.Sprintf("%s[@]", node.Name)
	ee.envManager.SetEnv(lengthKey, fmt.Sprintf("%d", len(node.Values)))
}

// collectOutput adds output to result with proper newline handling
func (ee *ExecutionEngine) collectOutput(result *ExecutionResult, output string) {
	if output != "" {
		result.Output += output
		if !strings.HasSuffix(output, "\n") {
			result.Output += "\n"
		}
	}
}

// Execute executes a complete script and returns the execution result.
//
// The method processes all nodes in the script sequentially, handling commands,
// pipelines, control flow statements, and function calls. It checks for context
// cancellation to support timeout handling.
//
// Parameters:
//   - ctx: Context for cancellation and timeout support
//   - script: The script AST to execute
//
// Returns:
//   - ExecutionResult: Contains success status, exit code, output, and command results
//   - error: Returns error if execution fails or context is cancelled
//
// Example:
//
//	script := &types.ScriptNode{
//	    Nodes: []types.Node{
//	        &types.CommandNode{Name: "echo", Args: []string{"hello"}},
//	    },
//	}
//	result, err := ee.Execute(ctx, script)
func (ee *ExecutionEngine) Execute(ctx context.Context, script *types.ScriptNode) (*ExecutionResult, error) {
	startTime := time.Now()

	// Check for context cancellation before starting
	if ctx.Err() != nil {
		return nil, errors.NewTimeoutError("script execution").
			WithContext("reason", ctx.Err().Error())
	}

	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0, len(script.Nodes)),
	}

	for _, node := range script.Nodes {
		// Check for context cancellation during execution
		if ctx.Err() != nil {
			result.Success = false
			result.ExitCode = 1
			result.Error = "execution cancelled or timed out"
			return result, errors.NewTimeoutError("script execution").
				WithContext("reason", ctx.Err().Error())
		}
	// Execute node based on its type
	var err error
	var shouldBreak bool

	switch n := node.(type) {
	case *types.CommandNode:
		result, err = ee.executeCommandNode(ctx, n, result)
		if err == nil && result != nil {
			shouldBreak = !result.Success || result.BreakFlag || result.ContinueFlag
		}
	case *types.PipeNode:
		result, err = ee.executePipeNode(ctx, n, result)
		if err == nil && result != nil {
			shouldBreak = !result.Success
		}
	case *types.IfNode:
		result, err = ee.executeIfNode(ctx, n, result)
		if err == nil && result != nil {
			shouldBreak = !result.Success
		}
	case *types.ForNode:
		result, err = ee.executeForNode(ctx, n, result)
		if err == nil && result != nil {
			shouldBreak = !result.Success
		}
	case *types.WhileNode:
		result, err = ee.executeWhileNode(ctx, n, result)
		if err == nil && result != nil {
			shouldBreak = !result.Success
		}
	case *types.AssignmentNode:
		err = ee.executeAssignmentNode(n)
	case *types.AnnotationNode:
		// Process annotation (register with annotation processor)
		// For now, annotations are parsed but not processed
	case *types.FunctionNode:
		// Store function definition (not executing it)
		ee.functions[n.Name] = n
	case *types.BreakNode:
		result.BreakFlag = true
		result.Success = true
		return result, nil
	case *types.ContinueNode:
		result.ContinueFlag = true
		result.Success = true
		return result, nil
	case *types.AndNode:
		result, err = ee.executeAndNode(ctx, n, result)
	case *types.OrNode:
		result, err = ee.executeOrNode(ctx, n, result)
	case *types.BackgroundNode:
		result, err = ee.executeBackgroundNode(ctx, n, result)
	case *types.HeredocNode:
		result, err = ee.executeHeredocNode(ctx, n, result)
	case *types.ArrayNode:
		ee.executeArrayNode(n)
	default:
		return nil, errors.NewExecutionError(errors.ErrExecutionFailed,
			fmt.Sprintf("unsupported node type: %T", n)).
			WithContext("node_type", fmt.Sprintf("%T", n))
	}

	// Handle error or early exit
	if err != nil {
		return nil, err
	}
	if shouldBreak && !result.BreakFlag && !result.ContinueFlag {
		break
	}
	if result.BreakFlag || result.ContinueFlag {
		return result, nil
	}
	}

	result.Duration = time.Since(startTime)
	result.Success = true
	// Ensure ExitCode is set to 0 on success
	if result.ExitCode != 0 {
		// This shouldn't happen if all commands succeeded
		// But if it does, log it and fix it
		fmt.Fprintf(os.Stderr, "[WARNING] Execute: Success=true but ExitCode=%d, resetting to 0\n", result.ExitCode)
		result.ExitCode = 0
	}
	return result, nil
}

// ExecuteCommand executes a single command and returns the result.
//
// The method performs security checks, determines execution mode (interpreted,
// process, or hybrid), and executes the command accordingly. Results are cached
// for performance when appropriate.
//
// Parameters:
//   - ctx: Context for cancellation and timeout support
//   - cmd: The command node to execute
//
// Returns:
//   - CommandResult: Contains command output, exit code, and execution metadata
//   - error: Returns error if execution fails
//
// Example:
//
//	cmd := &types.CommandNode{
//	    Name: "echo",
//	    Args: []string{"hello", "world"},
//	}
//	result, err := ee.ExecuteCommand(ctx, cmd)
func (ee *ExecutionEngine) ExecuteCommand(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Check for nil command
	if cmd == nil {
		return nil, errors.NewExecutionError(errors.ErrInvalidInput,
			"cannot execute nil command")
	}

	startTime := time.Now()

	// Expand variables in command arguments
	expandedArgs := ee.expandArgs(cmd.Args)
	// Create a copy of command with expanded args
	expandedCmd := &types.CommandNode{
		Pos:      cmd.Pos,
		Name:     cmd.Name,
		Args:     expandedArgs,
		Redirect: cmd.Redirect,
	}

	// Security check
	if err := ee.security.CheckCommand(expandedCmd); err != nil {
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    fmt.Sprintf("Security violation: %v", err),
			Duration: time.Since(startTime),
		}, nil
	}

	// Decide execution mode
	mode := ee.decideExecutionMode(expandedCmd)

	var result *CommandResult
	var err error

	switch mode {
	case ModeInterpreted:
		result, err = ee.executeInterpreted(ctx, expandedCmd)
	case ModeProcess:
		result, err = ee.executeProcess(ctx, expandedCmd)
	case ModeHybrid:
		result, err = ee.executeHybrid(ctx, expandedCmd)
	default:
		return nil, errors.NewExecutionError(errors.ErrExecutionFailed,
			fmt.Sprintf("unknown execution mode: %v", mode)).
			WithContext("mode", mode).
			WithContext("command", cmd.Name)
	}

	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	result.Mode = mode
	return result, nil
}

// ExecutePipeline executes a pipeline of commands with proper data flow
func (ee *ExecutionEngine) ExecutePipeline(ctx context.Context, pipeline *types.PipeNode) (*PipelineResult, error) {
	// Collect all commands in the pipeline
	commands := ee.collectPipelineCommands(pipeline)
	results := make([]*CommandResult, 0, len(commands))

	// Execute commands with piped data flow
	var previousOutput string
	for i, cmd := range commands {
		var result *CommandResult
		var err error

		if i == 0 {
			// First command - execute normally
			result, err = ee.ExecuteCommand(ctx, cmd)
		} else {
			// Subsequent commands - use previous output as input
			result, err = ee.ExecuteCommandWithInput(ctx, cmd, previousOutput)
		}

		if err != nil {
			return nil, err
		}

		results = append(results, result)

		// If command failed, stop pipeline
		if !result.Success {
			return &PipelineResult{
				Success:  false,
				ExitCode: result.ExitCode,
				Output:   result.Output,
				Error:    result.Error,
				Results:  results,
			}, nil
		}

		// Store output for next command
		previousOutput = result.Output
	}

	// Return final result
	lastResult := results[len(results)-1]
	return &PipelineResult{
		Success:  true,
		ExitCode: 0,
		Output:   lastResult.Output,
		Error:    "",
		Results:  results,
	}, nil
}

// collectPipelineCommands collects all commands from a pipeline tree
func (ee *ExecutionEngine) collectPipelineCommands(node types.Node) []*types.CommandNode {
	var commands []*types.CommandNode

	switch n := node.(type) {
	case *types.PipeNode:
		// Recursively collect left commands
		commands = append(commands, ee.collectPipelineCommands(n.Left)...)
		// Recursively collect right commands
		commands = append(commands, ee.collectPipelineCommands(n.Right)...)
	case *types.CommandNode:
		commands = append(commands, n)
	}

	return commands
}

// ExecuteCommandWithInput executes a command with input data
func (ee *ExecutionEngine) ExecuteCommandWithInput(ctx context.Context, cmd *types.CommandNode, input string) (*CommandResult, error) {
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

	// Execute process with input
	result, err := ee.executeProcessWithInput(ctx, cmd, input)
	if err != nil {
		return nil, err
	}

	result.Duration = time.Since(startTime)
	result.Mode = ModeProcess
	return result, nil
}

// executeProcessWithInput executes a command with stdin input
func (ee *ExecutionEngine) executeProcessWithInput(ctx context.Context, cmd *types.CommandNode, input string) (*CommandResult, error) {
	// Create command with context for timeout support
	command := exec.CommandContext(ctx, cmd.Name, cmd.Args...)

	// Set environment
	envVars := make([]string, 0, len(ee.envManager.GetAllEnv()))
	for key, value := range ee.envManager.GetAllEnv() {
		envVars = append(envVars, key+"="+value)
	}
	command.Env = envVars
	command.Dir = ee.envManager.GetWorkingDir()

	// Set up pipes with resource cleanup
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, errors.WrapError(errors.ErrExecutionFailed,
			"failed to create stdin pipe", err).
			WithContext("command", cmd.Name)
	}

	// Ensure stdin is closed on exit (resource cleanup)
	defer func() {
		if stdin != nil {
			stdin.Close()
		}
	}()

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Start command
	if err := command.Start(); err != nil {
		// Clean up stdin before returning
		stdin.Close()
		return &CommandResult{
			Command:  cmd,
			Success:  false,
			ExitCode: 1,
			Error:    err.Error(),
		}, nil
	}

	// Write input to stdin
	if _, err := stdin.Write([]byte(input)); err != nil {
		// Clean up process if write fails
		if command.Process != nil {
			command.Process.Kill()
		}
		stdin.Close()
		return nil, errors.WrapError(errors.ErrExecutionFailed,
			"failed to write to stdin", err).
			WithContext("command", cmd.Name)
	}
	stdin.Close()
	stdin = nil // Mark as closed to prevent double close in defer

	// Wait for command to complete with timeout handling
	err = command.Wait()

	// Check for context cancellation (timeout)
	if ctx.Err() != nil {
		// Clean up process if still running
		if command.Process != nil {
			command.Process.Kill()
		}
		return &CommandResult{
				Command:  cmd,
				Success:  false,
				ExitCode: 124, // Standard timeout exit code
				Error:    "command execution timed out",
			}, errors.NewTimeoutError(cmd.Name).
				WithContext("command", cmd.Name).
				WithContext("args", cmd.Args)
	}

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return &CommandResult{
		Command:  cmd,
		Success:  err == nil,
		ExitCode: exitCode,
		Output:   stdout.String(),
		Error:    stderr.String(),
	}, nil
}

// decideExecutionMode determines the best execution mode for a command
func (ee *ExecutionEngine) decideExecutionMode(cmd *types.CommandNode) ExecutionMode {
	// Check if it's a standard library function
	if ee.isStdLibFunction(cmd.Name) {
		return ModeInterpreted
	}

	// Check if it's a user-defined function
	if ee.isUserDefinedFunction(cmd.Name) {
		return ModeInterpreted
	}

	// Check if it's a module export
	if ee.moduleMgr.IsExportedFunction(cmd.Name) {
		return ModeInterpreted
	}

	// Check if external command exists
	if ee.isExternalCommandAvailable(cmd.Name) {
		return ModeProcess
	}

	// Default to process execution
	return ModeProcess
}

// isStdLibFunction checks if a function exists in standard library
func (ee *ExecutionEngine) isStdLibFunction(funcName string) bool {
	// Map of standard library functions
	stdlibFunctions := map[string]bool{
		"Print":                     true,
		"Println":                   true,
		"Error":                     true,
		"Errorln":                   true,
		"ReadFile":                  true,
		"WriteFile":                 true,
		"ListFiles":                 true,
		"FileExists":                true,
		"Contains":                  true,
		"Replace":                   true,
		"ToUpper":                   true,
		"ToLower":                   true,
		"Trim":                      true,
		"GetEnv":                    true,
		"SetEnv":                    true,
		"WorkingDir":                true,
		"ChangeDir":                 true,
		"StartHTTPServer":           true,
		"RegisterRoute":             true,
		"RegisterHTTPRoute":         true,
		"RegisterRouteWithResponse": true,
		"RegisterStaticRoute":         true,
		"RegisterStaticRouteAdvanced": true,
		"RegisterHTTPRouteAdvanced":   true,
		"EnableRequestLog":            true,
		"RegisterWebSocketRoute":      true,
		"BroadcastWebSocketMessage":   true,
		"BroadcastWebSocketMessageToRoom": true,
		"SendWebSocketMessage":        true,
		"JoinRoom":                   true,
		"LeaveRoom":                  true,
		"GetWebSocketConnectionCount": true,
		"GetWebSocketRoomCount":       true,
		"ListWebSocketRooms":         true,
		"StopHTTPServer":              true,
		"IsHTTPServerRunning":       true,
		"GetHTTPMethod":             true,
		"GetHTTPPath":               true,
		"GetHTTPQuery":              true,
		"GetHTTPHeader":             true,
		"GetHTTPBody":               true,
		"SetHTTPResponse":           true,
		"SetHTTPHeader":             true,
		"SetCache":                  true,
		"GetCache":                  true,
		"DeleteCache":               true,
		"ClearCache":                true,
		"CacheExists":               true,
		"GetCacheTTL":               true,
		"SetCacheBatch":             true,
		"GetCacheKeys":              true,
		"ConnectDB":                 true,
		"CloseDB":                   true,
		"IsDBConnected":             true,
		"QueryDB":                   true,
		"QueryRowDB":                true,
		"ExecDB":                    true,
		"GetQueryResult":            true,
		// IoC functions
		"RegisterBean": true,
		"GetBean":      true,
		"ContainsBean": true,
		// Config functions
		"LoadConfig":        true,
		"LoadConfigWithEnv": true,
		"GetConfig":         true,
		"GetConfigString":   true,
		"GetConfigInt":      true,
		"GetConfigBool":     true,
		"SetConfig":         true,
		"Source":            true,
		// Template functions
		"RenderTemplateFile":              true,
		"RenderTemplateString":            true,
		"SaveTemplateFile":                true,
		"SetHTTPResponseTemplate":         true,
		"SetHTTPResponseTemplateString":   true,
		// Error page functions
		"SetErrorPage":                    true,
	}
	return stdlibFunctions[funcName]
}

// executeInterpreted executes a command using the interpreter (built-in functions)
func (ee *ExecutionEngine) executeInterpreted(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Check if it's a user-defined function
	if fn, exists := ee.functions[cmd.Name]; exists {
		return ee.executeUserFunction(ctx, fn, cmd.Args)
	}

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

	output := result
	if output == "" {
		// If result is empty, try to get from command execution
		// This handles cases where the function returns empty string but we need the output
		output = result
	}
	return &CommandResult{
		Command:  cmd,
		Success:  true,
		ExitCode: 0,
		Output:   output,
	}, nil
}

// executeStdLibFunction executes a standard library function
func (ee *ExecutionEngine) executeStdLibFunction(funcName string, args []string) (string, error) {
	switch funcName {
	case "Print":
		if len(args) > 0 {
			// Expand variables in the argument
			expanded := ee.expandVariables(args[0])
			ee.stdlib.Print(expanded)
			return expanded, nil
		}
		return "", nil
	case "Println":
		if len(args) > 0 {
			// Expand variables in the argument
			expanded := ee.expandVariables(args[0])
			ee.stdlib.Println(expanded)
			return expanded, nil
		}
		ee.stdlib.Println("")
		return "", nil
	case "Error":
		if len(args) > 0 {
			ee.stdlib.Error(args[0])
			return args[0], nil
		}
		return "", nil
	case "Errorln":
		if len(args) > 0 {
			ee.stdlib.Errorln(args[0])
			return args[0], nil
		}
		ee.stdlib.Errorln("")
		return "", nil
	case "ReadFile":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"ReadFile requires filename argument").
				WithContext("function", "ReadFile")
		}
		return ee.stdlib.ReadFile(args[0])
	case "WriteFile":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"WriteFile requires filename and content arguments").
				WithContext("function", "WriteFile")
		}
		// Expand variables in arguments
		filename := ee.expandVariables(args[0])
		content := ee.expandVariables(args[1])
		err := ee.stdlib.WriteFile(filename, content)
		return "File written", err
	case "ListFiles":
		if len(args) == 0 {
			files, err := ee.stdlib.ListFiles(".")
			if err != nil {
				return "", err
			}
			return strings.Join(files, "\n"), nil
		}
		files, err := ee.stdlib.ListFiles(args[0])
		if err != nil {
			return "", err
		}
		return strings.Join(files, "\n"), nil
	case "FileExists":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"FileExists requires filename argument").
				WithContext("function", "FileExists")
		}
		exists := ee.stdlib.FileExists(args[0])
		return fmt.Sprintf("%v", exists), nil
	case "Contains":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"Contains requires haystack and needle arguments").
				WithContext("function", "Contains")
		}
		contains := ee.stdlib.Contains(args[0], args[1])
		return fmt.Sprintf("%v", contains), nil
	case "Replace":
		if len(args) < 3 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"Replace requires string, old, and new arguments").
				WithContext("function", "Replace")
		}
		return ee.stdlib.Replace(args[0], args[1], args[2]), nil
	case "ToUpper":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.ToUpper(args[0]), nil
	case "ToLower":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.ToLower(args[0]), nil
	case "Trim":
		if len(args) == 0 {
			return "", nil
		}
		return ee.stdlib.Trim(args[0]), nil
	case "GetEnv":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetEnv requires environment variable name").
				WithContext("function", "GetEnv")
		}
		return ee.stdlib.GetEnv(args[0]), nil
	case "SetEnv":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetEnv requires key and value arguments").
				WithContext("function", "SetEnv")
		}
		err := ee.stdlib.SetEnv(args[0], args[1])
		return "Environment variable set", err
	case "WorkingDir":
		wd, err := ee.stdlib.WorkingDir()
		if err != nil {
			return "", err
		}
		return wd, nil
	case "ChangeDir":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"ChangeDir requires directory path").
				WithContext("function", "ChangeDir")
		}
		err := ee.stdlib.ChangeDir(args[0])
		return "Directory changed", err
	case "StartHTTPServer":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"StartHTTPServer requires port argument").
				WithContext("function", "StartHTTPServer")
		}
		err := ee.stdlib.StartHTTPServer(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("HTTP server started on port %s", args[0]), nil
	case "RegisterRoute":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterRoute requires path and handler arguments").
				WithContext("function", "RegisterRoute")
		}
		err := ee.stdlib.RegisterRoute(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Route registered: %s -> %s", args[0], args[1]), nil
	case "RegisterHTTPRoute":
		if len(args) < 4 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterHTTPRoute requires method, path, handlerType, and handler arguments").
				WithContext("function", "RegisterHTTPRoute")
		}
		err := ee.stdlib.RegisterHTTPRoute(args[0], args[1], args[2], args[3])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("HTTP route registered: %s %s -> %s (%s)", args[0], args[1], args[3], args[2]), nil
	case "RegisterStaticRoute":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterStaticRoute requires path and directory arguments").
				WithContext("function", "RegisterStaticRoute")
		}
		err := ee.stdlib.RegisterStaticRoute(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Static route registered: %s -> %s", args[0], args[1]), nil
	case "RegisterStaticRouteAdvanced":
		if len(args) < 7 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterStaticRouteAdvanced requires path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, and spaFallback arguments").
				WithContext("function", "RegisterStaticRouteAdvanced")
		}
		err := ee.stdlib.RegisterStaticRouteAdvanced(args[0], args[1], args[2], args[3], args[4], args[5], args[6])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Advanced static route registered: %s -> %s (browse:%s, cache:%s, gzip:%s)", args[0], args[1], args[3], args[4], args[5]), nil
	case "RegisterHTTPRouteAdvanced":
		if len(args) < 8 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterHTTPRouteAdvanced requires method, path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, and spaFallback arguments").
				WithContext("function", "RegisterHTTPRouteAdvanced")
		}
		err := ee.stdlib.RegisterHTTPRouteAdvanced(args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Advanced HTTP route registered: %s %s -> %s", args[0], args[1], args[2]), nil
	case "RegisterRouteWithResponse":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterRouteWithResponse requires path and response arguments").
				WithContext("function", "RegisterRouteWithResponse")
		}
		err := ee.stdlib.RegisterRouteWithResponse(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Route registered: %s", args[0]), nil
	case "RegisterWebSocketRoute":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterWebSocketRoute requires path argument").
				WithContext("function", "RegisterWebSocketRoute")
		}
		handlerFunc := ""
		if len(args) >= 2 {
			handlerFunc = args[1]
		}
		err := ee.stdlib.RegisterWebSocketRoute(args[0], handlerFunc)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("WebSocket route registered: %s", args[0]), nil
	case "BroadcastWebSocketMessage":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"BroadcastWebSocketMessage requires message argument").
				WithContext("function", "BroadcastWebSocketMessage")
		}
		err := ee.stdlib.BroadcastWebSocketMessage(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Broadcasted: %s", args[0]), nil
	case "BroadcastWebSocketMessageToRoom":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"BroadcastWebSocketMessageToRoom requires room and message arguments").
				WithContext("function", "BroadcastWebSocketMessageToRoom")
		}
		err := ee.stdlib.BroadcastWebSocketMessageToRoom(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Broadcasted to room %s: %s", args[0], args[1]), nil
	case "SendWebSocketMessage":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SendWebSocketMessage requires connection ID and message arguments").
				WithContext("function", "SendWebSocketMessage")
		}
		err := ee.stdlib.SendWebSocketMessage(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Sent to %s: %s", args[0], args[1]), nil
	case "JoinRoom":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"JoinRoom requires connection ID and room name arguments").
				WithContext("function", "JoinRoom")
		}
		err := ee.stdlib.JoinRoom(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Joined room %s", args[1]), nil
	case "LeaveRoom":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"LeaveRoom requires connection ID argument").
				WithContext("function", "LeaveRoom")
		}
		err := ee.stdlib.LeaveRoom(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Left room"), nil
	case "GetWebSocketConnectionCount":
		count := ee.stdlib.GetWebSocketConnectionCount()
		return fmt.Sprintf("%d", count), nil
	case "GetWebSocketRoomCount":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetWebSocketRoomCount requires room name argument").
				WithContext("function", "GetWebSocketRoomCount")
		}
		count := ee.stdlib.GetWebSocketRoomCount(args[0])
		return fmt.Sprintf("%d", count), nil
	case "ListWebSocketRooms":
		rooms := ee.stdlib.ListWebSocketRooms()
		if len(rooms) == 0 {
			return "No active rooms", nil
		}
		return fmt.Sprintf("Active rooms: %s", strings.Join(rooms, ", ")), nil
	case "StopHTTPServer":
		err := ee.stdlib.StopHTTPServer()
		if err != nil {
			return "", err
		}
		return "HTTP server stopped", nil
	case "IsHTTPServerRunning":
		running := ee.stdlib.IsHTTPServerRunning()
		return fmt.Sprintf("%v", running), nil
	case "GetHTTPMethod":
		return ee.stdlib.GetHTTPMethod(), nil
	case "GetHTTPPath":
		return ee.stdlib.GetHTTPPath(), nil
	case "GetHTTPQuery":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetHTTPQuery requires key argument").
				WithContext("function", "GetHTTPQuery")
		}
		return ee.stdlib.GetHTTPQuery(args[0]), nil
	case "GetHTTPHeader":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetHTTPHeader requires name argument").
				WithContext("function", "GetHTTPHeader")
		}
		return ee.stdlib.GetHTTPHeader(args[0]), nil
	case "GetHTTPBody":
		return ee.stdlib.GetHTTPBody(), nil
	case "SetHTTPResponse":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetHTTPResponse requires status and body arguments").
				WithContext("function", "SetHTTPResponse")
		}
		status, err := strconv.Atoi(args[0])
		if err != nil {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				fmt.Sprintf("invalid status code: %s", args[0])).
				WithContext("function", "SetHTTPResponse")
		}
		ee.stdlib.SetHTTPResponse(status, args[1])
		return "Response set", nil
	case "SetHTTPHeader":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetHTTPHeader requires name and value arguments").
				WithContext("function", "SetHTTPHeader")
		}
		ee.stdlib.SetHTTPHeader(args[0], args[1])
		return "Header set", nil
	case "EnableRequestLog":
		level := ""
		if len(args) > 0 {
			level = args[0]
		}
		err := ee.stdlib.EnableRequestLog(level)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Request logging enabled (level: %s)", level), nil
	case "RenderTemplateFile":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RenderTemplateFile requires template path argument").
				WithContext("function", "RenderTemplateFile")
		}
		// Parse data from remaining args (key=value pairs)
		data := make(map[string]interface{})
		for i := 1; i < len(args); i++ {
			parts := strings.SplitN(args[i], "=", 2)
			if len(parts) == 2 {
				data[parts[0]] = parts[1]
			}
		}
		result, err := ee.stdlib.RenderTemplateFile(args[0], data)
		if err != nil {
			return "", err
		}
		return result, nil
	case "RenderTemplateString":
		if len(args) < 1 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RenderTemplateString requires template content argument").
				WithContext("function", "RenderTemplateString")
		}
		// Parse data from remaining args (key=value pairs)
		data := make(map[string]interface{})
		for i := 1; i < len(args); i++ {
			parts := strings.SplitN(args[i], "=", 2)
			if len(parts) == 2 {
				data[parts[0]] = parts[1]
			}
		}
		result, err := ee.stdlib.RenderTemplateString(args[0], data)
		if err != nil {
			return "", err
		}
		return result, nil
	case "SaveTemplateFile":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SaveTemplateFile requires path and content arguments").
				WithContext("function", "SaveTemplateFile")
		}
		err := ee.stdlib.SaveTemplateFile(args[0], args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Template saved to %s", args[0]), nil
	case "SetHTTPResponseTemplate":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetHTTPResponseTemplate requires status and template path arguments").
				WithContext("function", "SetHTTPResponseTemplate")
		}
		status, err := strconv.Atoi(args[0])
		if err != nil {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				fmt.Sprintf("invalid status code: %s", args[0])).
				WithContext("function", "SetHTTPResponseTemplate")
		}
		// Parse data from remaining args (key=value pairs)
		data := make(map[string]interface{})
		for i := 2; i < len(args); i++ {
			parts := strings.SplitN(args[i], "=", 2)
			if len(parts) == 2 {
				data[parts[0]] = parts[1]
			}
		}
		err = ee.stdlib.SetHTTPResponseTemplate(status, args[1], data)
		if err != nil {
			return "", err
		}
		return "Response set from template", nil
	case "SetHTTPResponseTemplateString":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetHTTPResponseTemplateString requires status and template content arguments").
				WithContext("function", "SetHTTPResponseTemplateString")
		}
		status, err := strconv.Atoi(args[0])
		if err != nil {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				fmt.Sprintf("invalid status code: %s", args[0])).
				WithContext("function", "SetHTTPResponseTemplateString")
		}
		// Parse data from remaining args (key=value pairs)
		data := make(map[string]interface{})
		for i := 2; i < len(args); i++ {
			parts := strings.SplitN(args[i], "=", 2)
			if len(parts) == 2 {
				data[parts[0]] = parts[1]
			}
		}
		err = ee.stdlib.SetHTTPResponseTemplateString(status, args[1], data)
		if err != nil {
			return "", err
		}
		return "Response set from template string", nil
	case "SetErrorPage":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetErrorPage requires status code and file path arguments").
				WithContext("function", "SetErrorPage")
		}
		statusCode, err := strconv.Atoi(args[0])
		if err != nil {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				fmt.Sprintf("invalid status code: %s", args[0])).
				WithContext("function", "SetErrorPage")
		}
		err = ee.stdlib.SetErrorPage(statusCode, args[1])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Error page set for status %d: %s", statusCode, args[1]), nil
	case "SetCache":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetCache requires key and value arguments").
				WithContext("function", "SetCache")
		}
		ttl := 0
		if len(args) >= 3 {
			var err error
			ttl, err = strconv.Atoi(args[2])
			if err != nil {
				return "", errors.NewExecutionError(errors.ErrInvalidInput,
					fmt.Sprintf("invalid TTL: %s", args[2])).
					WithContext("function", "SetCache")
			}
		}
		ee.stdlib.SetCache(args[0], args[1], ttl)
		return "Cache set", nil
	case "GetCache":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetCache requires key argument").
				WithContext("function", "GetCache")
		}
		value, exists := ee.stdlib.GetCache(args[0])
		if !exists {
			return "", nil // Return empty string if key not found (not an error)
		}
		return value, nil
	case "DeleteCache":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"DeleteCache requires key argument").
				WithContext("function", "DeleteCache")
		}
		ee.stdlib.DeleteCache(args[0])
		return "Cache deleted", nil
	case "ClearCache":
		ee.stdlib.ClearCache()
		return "Cache cleared", nil
	case "CacheExists":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"CacheExists requires key argument").
				WithContext("function", "CacheExists")
		}
		exists := ee.stdlib.CacheExists(args[0])
		return fmt.Sprintf("%v", exists), nil
	case "GetCacheTTL":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetCacheTTL requires key argument").
				WithContext("function", "GetCacheTTL")
		}
		ttl := ee.stdlib.GetCacheTTL(args[0])
		return fmt.Sprintf("%d", ttl), nil
	case "GetCacheKeys":
		pattern := "*"
		if len(args) > 0 {
			pattern = args[0]
		}
		keys := ee.stdlib.GetCacheKeys(pattern)
		return strings.Join(keys, "\n"), nil
	case "ConnectDB":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"ConnectDB requires dbType and dsn arguments").
				WithContext("function", "ConnectDB")
		}
		err := ee.stdlib.ConnectDB(args[0], args[1])
		if err != nil {
			return "", err
		}
		return "Database connected", nil
	case "CloseDB":
		err := ee.stdlib.CloseDB()
		if err != nil {
			return "", err
		}
		return "Database closed", nil
	case "IsDBConnected":
		connected := ee.stdlib.IsDBConnected()
		return fmt.Sprintf("%v", connected), nil
	case "QueryDB":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"QueryDB requires sql argument").
				WithContext("function", "QueryDB")
		}
		result, err := ee.stdlib.QueryDB(args[0], args[1:]...)
		if err != nil {
			return "", err
		}
		jsonResult, err := result.ToJSON()
		if err != nil {
			return "", err
		}
		return jsonResult, nil
	case "QueryRowDB":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"QueryRowDB requires sql argument").
				WithContext("function", "QueryRowDB")
		}
		result, err := ee.stdlib.QueryRowDB(args[0], args[1:]...)
		if err != nil {
			return "", err
		}
		jsonResult, err := result.ToJSON()
		if err != nil {
			return "", err
		}
		return jsonResult, nil
	case "ExecDB":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"ExecDB requires sql argument").
				WithContext("function", "ExecDB")
		}
		result, err := ee.stdlib.ExecDB(args[0], args[1:]...)
		if err != nil {
			return "", err
		}
		jsonResult, err := result.ToJSON()
		if err != nil {
			return "", err
		}
		return jsonResult, nil
	// IoC functions
	case "RegisterBean":
		if len(args) < 3 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"RegisterBean requires name, scope, and factory arguments").
				WithContext("function", "RegisterBean")
		}
		// Note: Factory function needs to be passed as a function reference
		// For now, this is a placeholder - full implementation requires function references
		return "Bean registration requires function reference (not yet fully implemented)", nil
	case "GetBean":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetBean requires name argument").
				WithContext("function", "GetBean")
		}
		bean, err := ee.stdlib.GetBean(args[0])
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", bean), nil
	case "ContainsBean":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"ContainsBean requires name argument").
				WithContext("function", "ContainsBean")
		}
		exists := ee.stdlib.ContainsBean(args[0])
		return fmt.Sprintf("%v", exists), nil
	// Config functions
	case "LoadConfig":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"LoadConfig requires path argument").
				WithContext("function", "LoadConfig")
		}
		err := ee.stdlib.LoadConfig(args[0])
		if err != nil {
			return "", err
		}
		return "Config loaded", nil
	case "LoadConfigWithEnv":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"LoadConfigWithEnv requires path and env arguments").
				WithContext("function", "LoadConfigWithEnv")
		}
		err := ee.stdlib.LoadConfigWithEnv(args[0], args[1])
		if err != nil {
			return "", err
		}
		return "Config loaded", nil
	case "GetConfig":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetConfig requires key argument").
				WithContext("function", "GetConfig")
		}
		value, err := ee.stdlib.GetConfig(args[0])
		if err != nil {
			return "", err
		}
		return value, nil
	case "GetConfigString":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetConfigString requires key and defaultValue arguments").
				WithContext("function", "GetConfigString")
		}
		return ee.stdlib.GetConfigString(args[0], args[1]), nil
	case "GetConfigInt":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetConfigInt requires key and defaultValue arguments").
				WithContext("function", "GetConfigInt")
		}
		defaultValue, err := strconv.Atoi(args[1])
		if err != nil {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				fmt.Sprintf("invalid default value: %s", args[1])).
				WithContext("function", "GetConfigInt")
		}
		return fmt.Sprintf("%d", ee.stdlib.GetConfigInt(args[0], defaultValue)), nil
	case "GetConfigBool":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"GetConfigBool requires key and defaultValue arguments").
				WithContext("function", "GetConfigBool")
		}
		defaultValue := args[1] == "true"
		return fmt.Sprintf("%v", ee.stdlib.GetConfigBool(args[0], defaultValue)), nil
	case "SetConfig":
		if len(args) < 2 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SetConfig requires key and value arguments").
				WithContext("function", "SetConfig")
		}
		ee.stdlib.SetConfig(args[0], args[1])
		return "Config set", nil
	case "Source":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"Source requires filepath argument").
				WithContext("function", "Source")
		}
		filePath := ee.expandVariables(args[0])
		result, err := ee.executeSourceFile(context.Background(), filePath)
		if err != nil {
			return "", err
		}
		if !result.Success {
			return "", errors.NewExecutionError(errors.ErrExecutionFailed,
				fmt.Sprintf("Source file execution failed: %s", result.Error)).
				WithContext("file", filePath)
		}
		return fmt.Sprintf("Source file loaded: %s", filePath), nil
	case "AddMiddleware":
		// Note: Middleware registration requires function reference
		// For now, this is a placeholder
		return "Middleware registration requires function reference (not yet fully implemented)", nil
	case "ClearMiddlewares":
		ee.stdlib.ClearMiddlewares()
		return "Middlewares cleared", nil
	case "SHA256Hash":
		if len(args) == 0 {
			return "", errors.NewExecutionError(errors.ErrInvalidInput,
				"SHA256Hash requires text argument").
				WithContext("function", "SHA256Hash")
		}
		return ee.stdlib.SHA256Hash(args[0]), nil
	default:
		return "", errors.NewExecutionError(errors.ErrExecutionFailed,
			fmt.Sprintf("unknown standard library function: %s", funcName)).
			WithContext("function", funcName)
	}
}

// executeProcess executes a command as an external process
func (ee *ExecutionEngine) executeProcess(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Handle special case for [ command (test command)
	// [ is a shell built-in that's equivalent to test
	var actualCmd *types.CommandNode = cmd
	if cmd.Name == "[" {
		// Convert [ args ] to test args
		// Remove the trailing ] if present
		args := cmd.Args
		if len(args) > 0 && args[len(args)-1] == "]" {
			args = args[:len(args)-1]
		}
		actualCmd = &types.CommandNode{
			Pos:      cmd.Pos,
			Name:     "test",
			Args:     args,
			Redirect: cmd.Redirect,
		}
	}

	// Check cache first (only if no redirects)
	if actualCmd.Redirect == nil {
		if cached, ok := ee.cache.Get(actualCmd.Name, actualCmd.Args); ok {
			return cached, nil
		}
	}

	// Create command with context
	command := exec.CommandContext(ctx, actualCmd.Name, actualCmd.Args...)

	// Set environment - convert map[string]string to []string
	envVars := make([]string, 0, len(ee.envManager.GetAllEnv()))
	for key, value := range ee.envManager.GetAllEnv() {
		envVars = append(envVars, key+"="+value)
	}

	// Debug: log environment for problematic commands
	// if len(envVars) == 0 {
	// }

	command.Env = envVars

	// Set working directory
	command.Dir = ee.envManager.GetWorkingDir()

	// Handle redirections
	var stdout, stderr strings.Builder
	if cmd.Redirect != nil {
		if err := ee.setupRedirect(command, cmd.Redirect, &stdout, &stderr); err != nil {
			return &CommandResult{
				Command:  cmd,
				Success:  false,
				ExitCode: 1,
				Error:    fmt.Sprintf("redirect error: %v", err),
			}, nil
		}
	} else {
		// No redirect - capture output
		command.Stdout = &stdout
		command.Stderr = &stderr
	}

	// Execute command with timeout handling
	startTime := time.Now()
	err := command.Run()
	duration := time.Since(startTime)

	// Check for context cancellation (timeout)
	if ctx.Err() != nil {
		// Context was cancelled - likely timeout
		return &CommandResult{
				Command:  cmd,
				Success:  false,
				ExitCode: 124, // Standard timeout exit code
				Error:    "command execution timed out",
				Duration: duration,
			}, errors.NewTimeoutError(cmd.Name).
				WithContext("command", cmd.Name).
				WithContext("args", cmd.Args)
	}

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

	// Cache successful results (only if no redirects)
	if err == nil && cmd.Redirect == nil {
		ee.cache.Put(cmd.Name, cmd.Args, result)
	}

	return result, nil
}

// setupRedirect sets up input/output redirection for a command
func (ee *ExecutionEngine) setupRedirect(cmd *exec.Cmd, redirect *types.RedirectNode, stdout, stderr *strings.Builder) error {
	switch redirect.Op {
	case ">": // Output redirection (overwrite)
		file, err := os.Create(redirect.File)
		if err != nil {
			return errors.WrapError(errors.ErrFileNotFound,
				fmt.Sprintf("failed to create file %s", redirect.File), err).
				WithContext("file", redirect.File).
				WithContext("operation", "create")
		}
		defer file.Close()

		if redirect.Fd == 1 || redirect.Fd == 0 { // stdout
			cmd.Stdout = file
		} else if redirect.Fd == 2 { // stderr
			cmd.Stderr = file
		}

	case ">>": // Output redirection (append)
		file, err := os.OpenFile(redirect.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return errors.WrapError(errors.ErrFileNotFound,
				fmt.Sprintf("failed to open file %s", redirect.File), err).
				WithContext("file", redirect.File).
				WithContext("operation", "append")
		}
		defer file.Close()

		if redirect.Fd == 1 || redirect.Fd == 0 {
			cmd.Stdout = file
		} else if redirect.Fd == 2 {
			cmd.Stderr = file
		}

	case "<": // Input redirection
		file, err := os.Open(redirect.File)
		if err != nil {
			return errors.NewFileNotFoundError(redirect.File).
				WithContext("operation", "read")
		}
		defer file.Close()
		cmd.Stdin = file

	case "2>&1": // Redirect stderr to stdout
		cmd.Stderr = cmd.Stdout

	case "&>": // Redirect both stdout and stderr to file
		file, err := os.Create(redirect.File)
		if err != nil {
			return errors.WrapError(errors.ErrFileNotFound,
				fmt.Sprintf("failed to create file %s", redirect.File), err).
				WithContext("file", redirect.File).
				WithContext("operation", "create")
		}
		defer file.Close()
		cmd.Stdout = file
		cmd.Stderr = file

	default:
		return errors.NewExecutionError(errors.ErrInvalidInput,
			fmt.Sprintf("unsupported redirect operator: %s", redirect.Op)).
			WithContext("operator", redirect.Op)
	}

	return nil
}

// executeHybrid executes a command using hybrid approach
//
// The hybrid execution mode attempts to intelligently choose the best execution method:
// 1. First tries interpreted execution (fast, for built-ins and user functions)
// 2. Falls back to process execution if interpreted fails or is unavailable
//
// This provides the best of both worlds - fast execution for known functions
// with graceful fallback to external process execution when needed.
func (ee *ExecutionEngine) executeHybrid(ctx context.Context, cmd *types.CommandNode) (*CommandResult, error) {
	// Try interpreted execution first
	if ee.isStdLibFunction(cmd.Name) || ee.isUserDefinedFunction(cmd.Name) || ee.moduleMgr.IsExportedFunction(cmd.Name) {
		result, err := ee.executeInterpreted(ctx, cmd)
		if err == nil && result.Success {
			return result, nil
		}
		// If interpreted execution failed, fall through to process execution
	}

	// Fall back to process execution
	return ee.executeProcess(ctx, cmd)
}

// isExternalCommandAvailable checks if an external command exists
func (ee *ExecutionEngine) isExternalCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// ExecuteIf executes an if-then-else statement
func (ee *ExecutionEngine) ExecuteIf(ctx context.Context, ifNode *types.IfNode) (*ExecutionResult, error) {
	// Evaluate condition
	conditionResult, err := ee.evaluateCondition(ctx, ifNode.Condition)
	if err != nil {
		return nil, err
	}

	// Execute appropriate branch
	if conditionResult {
		return ee.Execute(ctx, ifNode.Then)
	} else if ifNode.Else != nil {
		return ee.Execute(ctx, ifNode.Else)
	}

	// No else branch and condition was false
	return &ExecutionResult{
		Success:  true,
		ExitCode: 0,
		Commands: []*CommandResult{},
	}, nil
}

// ExecuteFor executes a for loop
func (ee *ExecutionEngine) ExecuteFor(ctx context.Context, forNode *types.ForNode) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0),
	}

	// Iterate over the list
	for _, item := range forNode.List {
		// Set loop variable
		ee.envManager.SetEnv(forNode.Variable, item)

		// Execute loop body
		loopResult, err := ee.Execute(ctx, forNode.Body)
		if err != nil {
			return nil, err
		}

		result.Commands = append(result.Commands, loopResult.Commands...)

		// Check for break statement
		if loopResult.BreakFlag {
			// Break out of loop
			result.Success = true
			result.ExitCode = 0
			return result, nil
		}

		// Check for continue statement
		if loopResult.ContinueFlag {
			// Continue to next iteration
			continue
		}

		// Check for errors
		if !loopResult.Success {
			result.Success = false
			result.ExitCode = loopResult.ExitCode
			return result, nil
		}
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}

// ExecuteWhile executes a while loop
func (ee *ExecutionEngine) ExecuteWhile(ctx context.Context, whileNode *types.WhileNode) (*ExecutionResult, error) {
	result := &ExecutionResult{
		Commands: make([]*CommandResult, 0),
	}

	maxIterations := 10000 // Safety limit to prevent infinite loops
	iterations := 0

	for {
		// Check iteration limit
		if iterations >= maxIterations {
			return nil, errors.NewExecutionError(errors.ErrResourceExhausted,
				fmt.Sprintf("while loop exceeded maximum iterations (%d)", maxIterations)).
				WithContext("max_iterations", maxIterations).
				WithContext("iterations", iterations)
		}
		iterations++

		// Evaluate condition
		conditionResult, err := ee.evaluateCondition(ctx, whileNode.Condition)
		if err != nil {
			return nil, err
		}

		// Exit loop if condition is false
		if !conditionResult {
			break
		}

		// Execute loop body
		loopResult, err := ee.Execute(ctx, whileNode.Body)
		if err != nil {
			return nil, err
		}

		result.Commands = append(result.Commands, loopResult.Commands...)

		// Check for break statement
		if loopResult.BreakFlag {
			// Break out of loop
			result.Success = true
			result.ExitCode = 0
			return result, nil
		}

		// Check for continue statement
		if loopResult.ContinueFlag {
			// Continue to next iteration
			continue
		}

		// Check for errors
		if !loopResult.Success {
			result.Success = false
			result.ExitCode = loopResult.ExitCode
			return result, nil
		}
	}

	result.Success = true
	result.ExitCode = 0
	return result, nil
}

// evaluateCondition evaluates a condition node and returns true/false
func (ee *ExecutionEngine) evaluateCondition(ctx context.Context, condition types.Node) (bool, error) {
	switch n := condition.(type) {
	case *types.CommandNode:
		// Execute command and check exit code
		cmdResult, err := ee.ExecuteCommand(ctx, n)
		if err != nil {
			return false, err
		}
		return cmdResult.Success && cmdResult.ExitCode == 0, nil

	default:
		return false, errors.NewExecutionError(errors.ErrExecutionFailed,
			fmt.Sprintf("unsupported condition node type: %T", n)).
			WithContext("node_type", fmt.Sprintf("%T", n))
	}
}

// isUserDefinedFunction checks if a function is user-defined
func (ee *ExecutionEngine) isUserDefinedFunction(funcName string) bool {
	_, exists := ee.functions[funcName]
	return exists
}

// getFunctionNames returns a sorted list of user-defined function names
func (ee *ExecutionEngine) getFunctionNames() []string {
	names := make([]string, 0, len(ee.functions))
	for name := range ee.functions {
		names = append(names, name)
	}
	return names
}

// executeUserFunction executes a user-defined function
func (ee *ExecutionEngine) executeUserFunction(ctx context.Context, fn *types.FunctionNode, args []string) (*CommandResult, error) {
	startTime := time.Now()

	// Save current environment state for function scope
	originalEnv := make(map[string]string)
	for k, v := range ee.envManager.GetAllEnv() {
		originalEnv[k] = v
	}

	// Set function arguments as environment variables ($1, $2, etc.)
	// Also support $0 for function name, $@ for all arguments, $# for argument count
	ee.envManager.SetEnv("0", fn.Name)
	ee.envManager.SetEnv("#", fmt.Sprintf("%d", len(args)))
	ee.envManager.SetEnv("@", strings.Join(args, " "))

	for i, arg := range args {
		ee.envManager.SetEnv(fmt.Sprintf("%d", i+1), arg)
	}

	// Execute function body
	result, err := ee.Execute(ctx, fn.Body)
	if err != nil {
		// Restore environment
		ee.restoreEnvironment(originalEnv)
		return &CommandResult{
			Command: &types.CommandNode{
				Name: fn.Name,
				Args: args,
			},
			Success:  false,
			ExitCode: 1,
			Error:    err.Error(),
			Duration: time.Since(startTime),
		}, nil
	}

	// Restore environment (function scope isolation)
	ee.restoreEnvironment(originalEnv)

	// Collect output from all commands
	var output strings.Builder
	for _, cmdResult := range result.Commands {
		if cmdResult.Output != "" {
			output.WriteString(cmdResult.Output)
			if !strings.HasSuffix(cmdResult.Output, "\n") {
				output.WriteString("\n")
			}
		}
	}

	return &CommandResult{
		Command: &types.CommandNode{
			Name: fn.Name,
			Args: args,
		},
		Success:  result.Success,
		ExitCode: result.ExitCode,
		Output:   strings.TrimSuffix(output.String(), "\n"),
		Duration: time.Since(startTime),
	}, nil
}

// restoreEnvironment restores the environment to a previous state
func (ee *ExecutionEngine) restoreEnvironment(env map[string]string) {
	// Clear current environment
	currentEnv := ee.envManager.GetAllEnv()
	for k := range currentEnv {
		ee.envManager.UnsetEnv(k)
	}

	// Restore original environment
	for k, v := range env {
		ee.envManager.SetEnv(k, v)
	}
}

// ExecuteBackground executes a command in the background
func (ee *ExecutionEngine) ExecuteBackground(ctx context.Context, bgNode *types.BackgroundNode) (*CommandResult, error) {
	// Create a new context that won't be cancelled when parent context is cancelled
	bgCtx := context.Background()

	// Execute the command
	var cmdResult *CommandResult
	var err error

	switch cmd := bgNode.Command.(type) {
	case *types.CommandNode:
		cmdResult, err = ee.ExecuteCommand(bgCtx, cmd)
	case *types.ScriptNode:
		result, execErr := ee.Execute(bgCtx, cmd)
		if execErr != nil {
			err = execErr
		} else if len(result.Commands) > 0 {
			cmdResult = result.Commands[0]
		}
	default:
		return nil, errors.NewExecutionError(errors.ErrInvalidInput,
			"unsupported command type for background execution")
	}

	if err != nil {
		return nil, err
	}

	// Increment job counter and store job
	ee.jobCounter++
	jobID := ee.jobCounter

	// Store job info (we can't store exec.Cmd directly for background jobs,
	// but we can track them by job ID)
	// For now, just return the result immediately
	// In a full implementation, we would start the process and return immediately

	return &CommandResult{
		Command: &types.CommandNode{
			Name: "background",
			Args: []string{fmt.Sprintf("[%d]", jobID)},
		},
		Success:  true,
		ExitCode: 0,
		Output:   fmt.Sprintf("[%d] %d", jobID, jobID),
		Duration: cmdResult.Duration,
		Mode:     ModeProcess,
	}, nil
}

// Helper function to convert error to string
func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// executeSourceFile loads and executes a Shode script file in the current context
// This allows modular code organization - functions defined in the source file
// will be available in the current execution context
func (ee *ExecutionEngine) executeSourceFile(ctx context.Context, filePath string) (*ExecutionResult, error) {
	// Resolve relative paths
	if !filepath.IsAbs(filePath) {
		// Make path relative to current working directory
		wd := ee.envManager.GetWorkingDir()
		filePath = filepath.Join(wd, filePath)
	}

	// Use SimpleParser (supports function definitions and is more reliable)
	// Tree-sitter parser may panic in some environments
	simpleP := shodeparser.NewSimpleParser()
	script, err := simpleP.ParseFile(filePath)
	if err != nil {
		return &ExecutionResult{
			Success:  false,
			ExitCode: 1,
			Error:    fmt.Sprintf("failed to parse source file %s: %v", filePath, err),
		}, nil
	}

	// First pass: extract function definitions (so they're available immediately)
	for _, node := range script.Nodes {
		if fnNode, ok := node.(*types.FunctionNode); ok {
			ee.functions[fnNode.Name] = fnNode
		}
	}

	// Second pass: execute the script (this will execute any non-function code)
	// But skip function definitions as they're already registered
	execScript := &types.ScriptNode{
		Pos: script.Pos,
	}
	for _, node := range script.Nodes {
		if _, ok := node.(*types.FunctionNode); !ok {
			// Not a function definition, execute it
			execScript.Nodes = append(execScript.Nodes, node)
		}
	}

	// Execute non-function code
	if len(execScript.Nodes) > 0 {
		return ee.Execute(ctx, execScript)
	}

	// Only function definitions, no execution needed
	return &ExecutionResult{
		Success:  true,
		ExitCode: 0,
		Commands: []*CommandResult{},
	}, nil
}

// parseArrayElements parses array elements from a space-separated string
// Handles quoted strings: "hello world" foo 'bar baz'
func (ee *ExecutionEngine) parseArrayElements(input string) []string {
	var result []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i, r := range input {
		switch r {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			} else {
				current.WriteRune(r)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			} else {
				current.WriteRune(r)
			}
		case ' ', '\t':
			if !inSingleQuote && !inDoubleQuote {
				if current.Len() > 0 {
					result = append(result, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}

		// Handle escaped characters
		if i > 0 && input[i-1] == '\\' && (r == '\'' || r == '"' || r == '\\') {
			// Remove the backslash from the current builder
			str := current.String()
			if len(str) >= 2 {
				current.Reset()
				current.WriteString(str[:len(str)-2])
				current.WriteRune(r)
			}
		}
	}

	// Add the last element
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
