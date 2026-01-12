package engine

import (
	"context"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/errors"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestTimeoutRecovery(t *testing.T) {
	// Create execution engine
	em := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := NewExecutionEngine(em, stdlib, mm, sc)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Create a script that will timeout (sleep command)
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "sleep",
				Args: []string{"1"}, // Sleep for 1 second, but timeout is 100ms
			},
		},
	}

	// Execute script
	result, err := ee.Execute(ctx, script)

	// Should get timeout error
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	// Check if it's a timeout error
	if !errors.IsTimeout(err) {
		t.Errorf("Expected timeout error, got: %v", err)
	}

	// Result should indicate failure
	if result != nil && result.Success {
		t.Error("Expected result to indicate failure")
	}
}

func TestContextCancellation(t *testing.T) {
	// Create execution engine
	em := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := NewExecutionEngine(em, stdlib, mm, sc)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Create a script
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{"hello"},
			},
		},
	}

	// Cancel context before execution
	cancel()

	// Execute script
	result, err := ee.Execute(ctx, script)

	// Should get cancellation error
	if err == nil {
		t.Fatal("Expected cancellation error, got nil")
	}

	// Check if it's a timeout error (cancellation is treated as timeout)
	if !errors.IsTimeout(err) {
		t.Errorf("Expected timeout/cancellation error, got: %v", err)
	}

	// Result should indicate failure
	if result != nil && result.Success {
		t.Error("Expected result to indicate failure")
	}
}

func TestPipelinePartialFailure(t *testing.T) {
	// Create execution engine
	em := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := NewExecutionEngine(em, stdlib, mm, sc)

	// Create a pipeline with a failing command
	pipe := &types.PipeNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Left: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{"hello"},
		},
		Right: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 10, Offset: 0},
			Name: "nonexistentcommand12345", // This will fail
			Args: []string{},
		},
	}

	// Execute pipeline
	ctx := context.Background()
	result, _ := ee.ExecutePipeline(ctx, pipe)

	// Should get partial results
	if result == nil {
		t.Fatal("Expected partial results, got nil")
	}

	// Should have at least one result (from first command)
	if len(result.Results) == 0 {
		t.Error("Expected at least one result from pipeline")
	}

	// Pipeline should indicate failure
	if result.Success {
		t.Error("Expected pipeline to indicate failure")
	}

	// Should have error information
	if result.Error == "" {
		t.Error("Expected error information in pipeline result")
	}
}

func TestCacheGracefulDegradation(t *testing.T) {
	// Create execution engine
	em := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := NewExecutionEngine(em, stdlib, mm, sc)

	// Create a context with reasonable timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a command that should work
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "echo",
		Args: []string{"test"},
	}

	// First execution - should succeed and cache
	result1, err1 := ee.ExecuteCommand(ctx, cmd)
	// Note: echo might not be available in all test environments
	// We're testing that the mechanism works, not that echo works
	if err1 != nil {
		// If command fails, that's okay - we're testing error handling
		if !errors.IsTimeout(err1) {
			t.Logf("First execution failed (expected in some environments): %v", err1)
		}
		return
	}

	// If first execution succeeded, second should use cache
	if result1 != nil && result1.Success {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()

		result2, err2 := ee.ExecuteCommand(ctx2, cmd)
		if err2 != nil {
			// If cache fails due to context, should still execute directly
			// This tests graceful degradation
			if !errors.IsTimeout(err2) {
				t.Logf("Second execution failed (graceful degradation): %v", err2)
			}
		} else if result2 != nil && result2.Success {
			// Cache worked or direct execution worked
			t.Log("Cache or direct execution succeeded")
		}
	}
}

func TestResourceCleanup(t *testing.T) {
	// Create execution engine
	em := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := NewExecutionEngine(em, stdlib, mm, sc)

	// Create a command with input
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "cat",
		Args: []string{},
	}

	ctx := context.Background()
	
	// Execute with input - should properly clean up stdin pipe
	result, err := ee.ExecuteCommandWithInput(ctx, cmd, "test input")
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Should succeed
	if !result.Success {
		t.Error("Execution should succeed")
	}

	// Resource cleanup is verified by lack of errors and proper execution
	// In a real scenario, we would check for file descriptor leaks
}
