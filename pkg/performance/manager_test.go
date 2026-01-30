package performance

import (
	"context"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// TestNewPerformanceManager 测试性能管理器创建
func TestNewPerformanceManager(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      true,
		EnableCache:    true,
		EnableParallel: true,
		EnableGC:       true,
		MaxWorkers:     4,
		GCThreshold:    1024 * 1024, // 1MB
	}

	pm := NewPerformanceManager(profile)
	if pm == nil {
		t.Fatal("NewPerformanceManager returned nil")
	}

	if pm.config != profile {
		t.Error("Config not set correctly")
	}
}

// TestPerformanceManagerInitialize 测试初始化
func TestPerformanceManagerInitialize(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      false, // 禁用JIT以避免文件系统依赖
		EnableCache:    false,
		EnableParallel: true,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)
	err := pm.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if !pm.initialized {
		t.Error("Manager not marked as initialized")
	}
}

// TestExecuteOptimized 测试优化执行
func TestExecuteOptimized(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: false, // 禁用并行以简化测试
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)
	err := pm.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 创建简单脚本
	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{
				Name:  "x",
				Value: "42",
			},
		},
	}

	result, err := pm.ExecuteOptimized(context.Background(), script, "test.sh")
	if err != nil {
		t.Fatalf("ExecuteOptimized failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if !result.Success {
		t.Error("Execution not successful")
	}

	if result.Duration <= 0 {
		t.Error("Invalid duration")
	}
}

// TestParallelExecution 测试并行执行
func TestParallelExecution(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: true,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)
	err := pm.Initialize()
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// 创建多语句脚本以测试并行执行
	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
			&types.AssignmentNode{Name: "y", Value: "2"},
			&types.AssignmentNode{Name: "z", Value: "3"},
		},
	}

	result, err := pm.ExecuteOptimized(context.Background(), script, "test.sh")
	if err != nil {
		t.Fatalf("ExecuteOptimized failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// 验证并行执行统计
	stats := pm.parallel.GetStats()
	if stats.TasksTotal == 0 {
		t.Error("No tasks executed")
	}
}

// TestFastExecute 测试快速执行模式
func TestFastExecute(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: false,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
		},
	}

	result, err := pm.FastExecute(context.Background(), script)
	if err != nil {
		t.Fatalf("FastExecute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if !result.Success {
		t.Error("Fast execution not successful")
	}
}

// TestGetPerformanceReport 测试性能报告
func TestGetPerformanceReport(t *testing.T) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: false,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)

	report := pm.GetPerformanceReport()
	if report == nil {
		t.Fatal("Report is nil")
	}

	if report.Config != profile {
		t.Error("Report config mismatch")
	}

	if report.Stats == nil {
		t.Error("Stats is nil")
	}

	if report.JITStats == nil {
		t.Error("JITStats is nil")
	}

	if report.ParallelStats == nil {
		t.Error("ParallelStats is nil")
	}

	if report.MemoryStats == nil {
		t.Error("MemoryStats is nil")
	}
}

// TestMemoryOptimizer 测试内存优化器
func TestMemoryOptimizer(t *testing.T) {
	mo := NewMemoryOptimizer(false, 1024*1024)
	if mo == nil {
		t.Fatal("NewMemoryOptimizer returned nil")
	}

	// 测试对象分配
	ptr, err := mo.Allocate("test", 1024)
	if err != nil {
		t.Fatalf("Allocate failed: %v", err)
	}

	if ptr == 0 {
		t.Error("Invalid pointer returned")
	}

	// 测试对象信息
	info, err := mo.allocator.GetInfo(ptr)
	if err != nil {
		t.Fatalf("GetInfo failed: %v", err)
	}

	if info.Type != "test" {
		t.Errorf("Expected type 'test', got '%s'", info.Type)
	}

	if info.Size != 1024 {
		t.Errorf("Expected size 1024, got %d", info.Size)
	}

	// 测试对象释放
	err = mo.Free(ptr)
	if err != nil {
		t.Fatalf("Free failed: %v", err)
	}
}

// TestJITCompiler 测试JIT编译器
func TestJITCompiler(t *testing.T) {
	jit := NewJITCompiler("", false, false)
	if jit == nil {
		t.Fatal("NewJITCompiler returned nil")
	}

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
		},
	}

	// 测试编译
	compiled, err := jit.Compile(context.Background(), script, "test.sh")
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if compiled == nil {
		t.Fatal("Compiled result is nil")
	}

	if len(compiled.Bytecode) == 0 {
		t.Error("Bytecode is empty")
	}

	// 测试统计
	stats := jit.GetStats()
	if stats == nil {
		t.Fatal("Stats is nil")
	}
}

// TestParallelExecutor 测试并行执行器
func TestParallelExecutor(t *testing.T) {
	pe := NewParallelExecutor(2)
	if pe == nil {
		t.Fatal("NewParallelExecutor returned nil")
	}

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
			&types.AssignmentNode{Name: "y", Value: "2"},
		},
	}

	// 测试执行
	result, err := pe.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	// 测试统计
	stats := pe.GetStats()
	if stats == nil {
		t.Fatal("Stats is nil")
	}

	if stats.TasksTotal == 0 {
		t.Error("No tasks executed")
	}
}

// BenchmarkSimpleExecution 基准测试简单执行
func BenchmarkSimpleExecution(b *testing.B) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: false,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)
	pm.Initialize()

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.ExecuteOptimized(context.Background(), script, "test.sh")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkJITCompilation 基准测试JIT编译
func BenchmarkJITCompilation(b *testing.B) {
	profile := &PerformanceProfile{
		EnableJIT:      true,
		EnableCache:    true,
		EnableParallel: false,
		EnableGC:       false,
		MaxWorkers:     2,
	}

	pm := NewPerformanceManager(profile)
	pm.Initialize()

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.ExecuteOptimized(context.Background(), script, "test.sh")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

// BenchmarkParallelExecution 基准测试并行执行
func BenchmarkParallelExecution(b *testing.B) {
	profile := &PerformanceProfile{
		EnableJIT:      false,
		EnableCache:    false,
		EnableParallel: true,
		EnableGC:       false,
		MaxWorkers:     4,
	}

	pm := NewPerformanceManager(profile)
	pm.Initialize()

	script := &types.ScriptNode{
		Nodes: []types.Node{
			&types.AssignmentNode{Name: "x", Value: "1"},
			&types.AssignmentNode{Name: "y", Value: "2"},
			&types.AssignmentNode{Name: "z", Value: "3"},
			&types.AssignmentNode{Name: "w", Value: "4"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pm.ExecuteOptimized(context.Background(), script, "test.sh")
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}
