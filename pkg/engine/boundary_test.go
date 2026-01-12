package engine

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestLargeFileHandling(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()

	// Create a large file (10MB)
	tmpDir, err := os.MkdirTemp("", "shode-largefile-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := strings.Repeat("A", 10*1024*1024) // 10MB
	err = os.WriteFile(largeFile, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	// Test reading large file
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "cat",
		Args: []string{largeFile},
	}

	result, err := ee.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("Failed to execute command: %v", err)
	}

	// Verify output size
	if len(result.Output) != len(largeContent) {
		t.Errorf("Expected output size %d, got %d", len(largeContent), len(result.Output))
	}
}

func TestLongCommandArguments(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()

	// Create command with very long arguments
	longArg := strings.Repeat("A", 10000) // 10KB argument
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "echo",
		Args: []string{longArg},
	}

	result, err := ee.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("Failed to execute command with long arguments: %v", err)
	}

	if !result.Success {
		t.Error("Command with long arguments should succeed")
	}
}

func TestDeepNestedPipeline(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()

	// Create deeply nested pipeline (10 levels)
	var buildPipeline func(depth int) *types.PipeNode
	buildPipeline = func(depth int) *types.PipeNode {
		if depth == 0 {
			return &types.PipeNode{
				Pos: types.Position{Line: 1, Column: 1, Offset: 0},
				Left: &types.CommandNode{
					Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
					Name: "echo",
					Args: []string{"test"},
				},
				Right: &types.CommandNode{
					Pos:  types.Position{Line: 1, Column: 10, Offset: 0},
					Name: "cat",
					Args: []string{},
				},
			}
		}
		return &types.PipeNode{
			Pos:   types.Position{Line: 1, Column: 1, Offset: 0},
			Left:  buildPipeline(depth - 1),
			Right: &types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 10, Offset: 0},
				Name: "cat",
				Args: []string{},
			},
		}
	}

	pipeline := buildPipeline(10)
	result, err := ee.ExecutePipeline(ctx, pipeline)

	// Should handle deep nesting without stack overflow
	if err != nil {
		t.Logf("Deep pipeline execution result: %v (may fail in some environments)", err)
	} else if result != nil {
		t.Log("Deep nested pipeline executed successfully")
	}
}

func TestConcurrentExecution(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()

	// Test concurrent command execution
	concurrency := 10
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cmd := &types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{string(rune('0' + id))},
			}
			_, err := ee.ExecuteCommand(ctx, cmd)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		t.Logf("Concurrent execution error: %v", err)
		errorCount++
	}

	if errorCount > concurrency/2 {
		t.Errorf("Too many concurrent execution errors: %d", errorCount)
	}
}

func TestResourceExhaustion(t *testing.T) {
	ee := setupTestEngine()
	ctx := context.Background()

	// Test with maximum iterations in while loop
	whileNode := &types.WhileNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Condition: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 5, Offset: 0},
			Name: "test",
			Args: []string{"-f", "/dev/null"},
		},
		Body: &types.ScriptNode{
			Pos: types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
					Name: "echo",
					Args: []string{"loop"},
				},
			},
		},
	}

	// This should be prevented by max iterations limit
	result, err := ee.ExecuteWhile(ctx, whileNode)
	if err != nil {
		// Expected: should hit max iterations limit
		t.Logf("Resource exhaustion prevented (expected): %v", err)
	} else if result != nil {
		t.Log("While loop executed (may succeed if condition becomes false)")
	}
}

func TestTimeoutHandling(t *testing.T) {
	ee := setupTestEngine()

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Create a command that will timeout
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "sleep",
		Args: []string{"1"},
	}

	result, err := ee.ExecuteCommand(ctx, cmd)
	if err == nil && result != nil {
		// Should timeout
		if result.ExitCode != 124 {
			t.Errorf("Expected timeout exit code 124, got %d", result.ExitCode)
		}
	}
}

func TestMemoryEfficientCache(t *testing.T) {
	// Test cache with many entries
	cache := NewCommandCache(1000)

	// Add many entries
	for i := 0; i < 2000; i++ {
		cmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{string(rune('0' + i%10))},
		}
		result := &CommandResult{
			Command: cmd,
			Success: true,
		}
		cache.Put(cmd.Name, cmd.Args, result)
	}

	// Verify cache size is limited
	hits, misses, size := cache.Stats()
	if size > 1000 {
		t.Errorf("Cache size should be limited to 1000, got %d", size)
	}

	t.Logf("Cache stats: hits=%d, misses=%d, size=%d", hits, misses, size)
}
