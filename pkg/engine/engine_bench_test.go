package engine

import (
	"context"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func setupBenchmarkEngine() *ExecutionEngine {
	em := environment.NewEnvironmentManager()
	stdLib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()
	return NewExecutionEngine(em, stdLib, moduleMgr, security)
}

// BenchmarkExecuteCommand benchmarks single command execution
func BenchmarkExecuteCommand(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "echo",
		Args: []string{"test"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.ExecuteCommand(ctx, cmd)
	}
}

// BenchmarkExecutePipeline benchmarks pipeline execution
func BenchmarkExecutePipeline(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	pipe := &types.PipeNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Left: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{"hello"},
		},
		Right: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 10, Offset: 0},
			Name: "cat",
			Args: []string{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.ExecutePipeline(ctx, pipe)
	}
}

// BenchmarkCommandCache benchmarks cache hit performance
func BenchmarkCommandCache(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	cmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "echo",
		Args: []string{"cached"},
	}

	// Warm up cache
	_, _ = ee.ExecuteCommand(ctx, cmd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.ExecuteCommand(ctx, cmd)
	}
}

// BenchmarkExecuteForLoop benchmarks for loop execution
func BenchmarkExecuteForLoop(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	// Create a simple for loop with list items
	forNode := &types.ForNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Variable: "item",
		Body: &types.ScriptNode{
			Pos: types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
					Name: "echo",
					Args: []string{"test"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.ExecuteFor(ctx, forNode)
	}
}

// BenchmarkExecuteIf benchmarks if statement execution
func BenchmarkExecuteIf(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	ifNode := &types.IfNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Condition: &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 4, Offset: 0},
			Name: "test",
			Args: []string{"-f", "/dev/null"},
		},
		Then: &types.ScriptNode{
			Pos: types.Position{Line: 2, Column: 1, Offset: 0},
			Nodes: []types.Node{
				&types.CommandNode{
					Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
					Name: "echo",
					Args: []string{"true"},
				},
			},
		},
		Else: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.ExecuteIf(ctx, ifNode)
	}
}

// BenchmarkExecuteScript benchmarks full script execution
func BenchmarkExecuteScript(b *testing.B) {
	ee := setupBenchmarkEngine()
	ctx := context.Background()
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{"hello"},
			},
			&types.CommandNode{
				Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{"world"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(ctx, script)
	}
}
