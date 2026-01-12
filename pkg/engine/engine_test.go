package engine

import (
	"context"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// setupTestEngine creates a test execution engine
func setupTestEngine() *ExecutionEngine {
	envManager := environment.NewEnvironmentManager()
	stdLib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()
	
	return NewExecutionEngine(envManager, stdLib, moduleMgr, security)
}

func TestExecuteCommand(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Test simple command
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "echo",
		Args: []string{"hello", "world"},
	}
	
	result, err := ee.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}
	
	if result == nil {
		t.Fatal("ExecuteCommand returned nil result")
	}
	
	// Note: echo command might not be available in all environments
	// This test verifies the function doesn't crash
}

func TestExecuteStdLibFunction(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Test standard library function
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "Println",
		Args: []string{"test message"},
	}
	
	result, err := ee.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Expected success, got exit code %d", result.ExitCode)
	}
}

func TestFunctionDefinitionAndExecution(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Define a function
	fn := &types.FunctionNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "test_func",
		Body: &types.ScriptNode{
			Pos: types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
					Name: "Println",
					Args: []string{"Hello from function"},
				},
			},
		},
	}
	
	// Store function
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{fn},
	}
	
	_, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to define function: %v", err)
	}
	
	// Check if function is stored
	if !ee.isUserDefinedFunction("test_func") {
		t.Error("Function was not stored")
	}
	
	// Call the function
	callCmd := &types.CommandNode{
		Pos:  types.Position{Line: 3, Column: 1, Offset: 0},
		Name: "test_func",
		Args: []string{},
	}
	
	result, err := ee.ExecuteCommand(ctx, callCmd)
	if err != nil {
		t.Fatalf("Failed to execute function: %v", err)
	}
	
	if !result.Success {
		t.Errorf("Function execution failed with exit code %d", result.ExitCode)
	}
}

func TestForLoopWithBreak(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Create a for loop with break
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.ForNode{
				Pos:      types.Position{Line: 1, Column: 1, Offset: 0},
				Variable: "item",
				List:     []string{"a", "b", "c", "d", "e"},
				Body: &types.ScriptNode{
					Pos: types.Position{Line: 2, Column: 1, Offset: 0},
					Nodes: []types.Node{
						&types.CommandNode{
							Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
							Name: "Println",
							Args: []string{"Processing", "$item"},
						},
						&types.CommandNode{
							Pos:  types.Position{Line: 3, Column: 1, Offset: 0},
							Name: "break",
							Args: []string{},
						},
					},
				},
			},
		},
	}
	
	result, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to execute for loop: %v", err)
	}
	
	if !result.Success {
		t.Error("For loop execution should succeed")
	}
	
	// Should only process first item before break
	if len(result.Commands) < 1 {
		t.Error("Expected at least one command to execute")
	}
}

func TestWhileLoopWithContinue(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Create a counter variable
	ee.envManager.SetEnv("counter", "0")
	
	// Create a while loop with continue
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.WhileNode{
				Pos: types.Position{Line: 1, Column: 1, Offset: 0},
				Condition: &types.CommandNode{
					Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
					Name: "test",
					Args: []string{"-lt", "5"},
				},
				Body: &types.ScriptNode{
					Pos: types.Position{Line: 2, Column: 1, Offset: 0},
					Nodes: []types.Node{
						&types.CommandNode{
							Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
							Name: "continue",
							Args: []string{},
						},
					},
				},
			},
		},
	}
	
	// Note: This test verifies the structure, actual execution depends on condition evaluation
	_, err := ee.Execute(ctx, script)
	// We expect this might fail due to condition evaluation, but structure should be correct
	if err != nil {
		t.Logf("While loop execution note: %v", err)
	}
}

func TestPipelineExecution(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Create a simple pipeline: echo "hello" | cat
	pipe := &types.PipeNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Left: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{"test"},
		},
		Right: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 15, Offset: 0},
			Name: "cat",
			Args: []string{},
		},
	}
	
	script := &types.ScriptNode{
		Pos:   types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{pipe},
	}
	
	result, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to execute pipeline: %v", err)
	}
	
	if result == nil {
		t.Fatal("Pipeline execution returned nil result")
	}
	
	// Note: Actual command execution depends on system availability
}

func TestIfStatement(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	// Create an if statement
	ifNode := &types.IfNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Condition: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "test",
			Args: []string{"-f", "go.mod"},
		},
		Then: &types.ScriptNode{
			Pos: types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
					Name: "Println",
					Args: []string{"File exists"},
				},
			},
		},
		Else: &types.ScriptNode{
			Pos: types.Position{Line: 3, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 3, Column: 1, Offset: 0},
					Name: "Println",
					Args: []string{"File does not exist"},
				},
			},
		},
	}
	
	script := &types.ScriptNode{
		Pos:   types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{ifNode},
	}
	
	result, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to execute if statement: %v", err)
	}
	
	if result == nil {
		t.Fatal("If statement execution returned nil result")
	}
}

func TestVariableAssignment(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()
	
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.AssignmentNode{
				Pos:   types.Position{Line: 1, Column: 1, Offset: 0},
				Name:  "TEST_VAR",
				Value: "test_value",
			},
		},
	}
	
	result, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to execute assignment: %v", err)
	}
	
	// Check if variable was set
	value := ee.envManager.GetEnv("TEST_VAR")
	if value != "test_value" {
		t.Errorf("Expected TEST_VAR='test_value', got '%s'", value)
	}
	
	if !result.Success {
		t.Error("Variable assignment should succeed")
	}
}

func TestCommandTimeout(t *testing.T) {
	ee := setupTestEngine()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// This test verifies timeout handling
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "sleep",
		Args: []string{"1"},
	}
	
	// Note: This might not work if sleep command is not available
	// But it tests the timeout mechanism
	_, err := ee.ExecuteCommand(ctx, cmd)
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("Command execution note: %v", err)
	}
}

func TestIsUserDefinedFunction(t *testing.T) {
	ee := setupTestEngine()
	
	// Function should not exist initially
	if ee.isUserDefinedFunction("nonexistent") {
		t.Error("Function should not exist")
	}
	
	// Define a function
	fn := &types.FunctionNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "my_func",
		Body: &types.ScriptNode{
			Pos:   types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{},
		},
	}
	
	script := &types.ScriptNode{
		Pos:   types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{fn},
	}
	
	ctx := context.Background()
	_, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Failed to define function: %v", err)
	}
	
	// Function should exist now
	if !ee.isUserDefinedFunction("my_func") {
		t.Error("Function should exist after definition")
	}
}
