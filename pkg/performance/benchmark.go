package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/parser"
)

// BenchmarkSuite 性能测试套件
type BenchmarkSuite struct {
	name       string
	benchmarks []*Benchmark
	results    []*BenchmarkResult
	warmupRuns int
	iterations int
}

// Benchmark 单个性能测试
type Benchmark struct {
	Name        string
	Setup       func() (*ExecutionContext, error)
	Execute     func(*ExecutionContext) error
	Teardown   func(*ExecutionContext) error
	Iterations  int
}

// BenchmarkResult 性能测试结果
type BenchmarkResult struct {
	Name          string
	Iterations    int
	Duration      time.Duration
	AvgDuration   time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	MemoryAlloc   uint64
	MemoryFreed   uint64
	Passed        bool
	Error         error
	Timestamp     time.Time
	Stats         map[string]interface{}
}

// BenchmarkComparison 性能对比
type BenchmarkComparison struct {
	Before       *BenchmarkResult
	After        *BenchmarkResult
	Speedup      float64
	MemoryChange float64
}

// Profiler 性能分析器
type Profiler struct {
	enabled       bool
	outputWriter  io.Writer
	cpuProfile    *os.File
	memProfile    *os.File
	blockProfile  *os.File
}

// NewBenchmarkSuite 创建性能测试套件
func NewBenchmarkSuite(name string) *BenchmarkSuite {
	return &BenchmarkSuite{
		name:        name,
		benchmarks:   make([]*Benchmark, 0),
		results:     make([]*BenchmarkResult, 0),
		warmupRuns:  3,
		iterations:  100,
	}
}

// AddBenchmark 添加性能测试
func (bs *BenchmarkSuite) AddBenchmark(bench *Benchmark) {
	bs.benchmarks = append(bs.benchmarks, bench)
}

// Run 运行性能测试套件
func (bs *BenchmarkSuite) Run(ctx context.Context) error {
	fmt.Printf("Running benchmark suite: %s\n", bs.name)
	fmt.Printf("Warmup runs: %d, Iterations: %d\n\n", bs.warmupRuns, bs.iterations)

	for _, bench := range bs.benchmarks {
		result, err := bs.runBenchmark(ctx, bench)
		if err != nil {
			return fmt.Errorf("benchmark %s failed: %w", bench.Name, err)
		}

		bs.results = append(bs.results, result)
		bs.printResult(result)
	}

	return nil
}

// runBenchmark 运行单个性能测试
func (bs *BenchmarkSuite) runBenchmark(ctx context.Context, bench *Benchmark) (*BenchmarkResult, error) {
	fmt.Printf("Benchmark: %s\n", bench.Name)

	iterations := bench.Iterations
	if iterations == 0 {
		iterations = bs.iterations
	}

	// 预热
	for i := 0; i < bs.warmupRuns; i++ {
		execCtx, err := bench.Setup()
		if err != nil {
			return nil, err
		}
		bench.Execute(execCtx)
		bench.Teardown(execCtx)
	}

	// 实际测试
	var durations []time.Duration
	var totalMem uint64
	var startMem, endMem uint64

	runtime.GC()
	startMem = getMemoryUsage()

	startTime := time.Now()

	for i := 0; i < iterations; i++ {
		execCtx, err := bench.Setup()
		if err != nil {
			return nil, err
		}

		iterStart := time.Now()
		err = bench.Execute(execCtx)
		dur := time.Since(iterStart)

		bench.Teardown(execCtx)

		if err != nil {
			return &BenchmarkResult{
				Name:      bench.Name,
				Iterations: i,
				Passed:    false,
				Error:      err,
				Timestamp: time.Now(),
			}, nil
		}

		durations = append(durations, dur)
		totalMem += getMemoryUsage() - startMem
	}

	endMem = getMemoryUsage()

	// 计算统计
	totalDuration := time.Since(startTime)
	avgDuration := totalDuration / time.Duration(iterations)
	minDuration := durations[0]
	maxDuration := durations[0]

	for _, d := range durations {
		if d < minDuration {
			minDuration = d
		}
		if d > maxDuration {
			maxDuration = d
		}
	}

	result := &BenchmarkResult{
		Name:        bench.Name,
		Iterations:  iterations,
		Duration:    totalDuration,
		AvgDuration: avgDuration,
		MinDuration: minDuration,
		MaxDuration: maxDuration,
		MemoryAlloc: endMem - startMem,
		MemoryFreed: totalMem,
		Passed:      true,
		Timestamp:   time.Now(),
		Stats:       make(map[string]interface{}),
	}

	result.Stats["ops_per_sec"] = float64(iterations) / totalDuration.Seconds()
	result.Stats["ns_per_op"] = float64(avgDuration.Nanoseconds())

	return result, nil
}

// printResult 打印测试结果
func (bs *BenchmarkSuite) printResult(result *BenchmarkResult) {
	fmt.Printf("  Results: %s\n", result.Name)
	fmt.Printf("    Iterations: %d\n", result.Iterations)
	fmt.Printf("    Total: %v\n", result.Duration)
	fmt.Printf("    Average: %v\n", result.AvgDuration)
	fmt.Printf("    Min: %v\n", result.MinDuration)
	fmt.Printf("    Max: %v\n", result.MaxDuration)
	fmt.Printf("    Memory: %d bytes\n", result.MemoryAlloc)
	fmt.Printf("    Ops/sec: %.0f\n\n", result.Stats["ops_per_sec"])
}

// CompareResults 对比性能测试结果
func (bs *BenchmarkSuite) CompareResults(beforeName, afterName string) (*BenchmarkComparison, error) {
	var beforeResult, afterResult *BenchmarkResult

	// 查找"before"结果
	for _, r := range bs.results {
		if r.Name == beforeName {
			beforeResult = r
			break
		}
	}

	// 查找"after"结果
	for _, r := range bs.results {
		if r.Name == afterName {
			afterResult = r
			break
		}
	}

	if beforeResult == nil {
		return nil, fmt.Errorf("before result not found: %s", beforeName)
	}

	if afterResult == nil {
		return nil, fmt.Errorf("after result not found: %s", afterName)
	}

	// 计算性能提升
	speedup := float64(beforeResult.AvgDuration) / float64(afterResult.AvgDuration)
	memoryChange := float64(afterResult.MemoryAlloc) / float64(beforeResult.MemoryAlloc)

	comparison := &BenchmarkComparison{
		Before:       beforeResult,
		After:        afterResult,
		Speedup:      speedup,
		MemoryChange: memoryChange,
	}

	bs.printComparison(comparison)

	return comparison, nil
}

// printComparison 打印对比结果
func (bs *BenchmarkSuite) printComparison(comp *BenchmarkComparison) {
	fmt.Printf("\n=== Performance Comparison ===\n")
	fmt.Printf("Before: %s\n", comp.Before.Name)
	fmt.Printf("  Avg: %v\n", comp.Before.AvgDuration)
	fmt.Printf("  Memory: %d bytes\n\n", comp.Before.MemoryAlloc)

	fmt.Printf("After: %s\n", comp.After.Name)
	fmt.Printf("  Avg: %v\n", comp.After.AvgDuration)
	fmt.Printf("  Memory: %d bytes\n\n", comp.After.MemoryAlloc)

	fmt.Printf("Speedup: %.2fx faster\n", comp.Speedup)

	if comp.MemoryChange < 1.0 {
		fmt.Printf("Memory: %.2f%% less\n", (1-comp.MemoryChange)*100)
	} else {
		fmt.Printf("Memory: %.2f%% more\n", (comp.MemoryChange-1)*100)
	}
}

// WriteResults 写入测试结果到文件
func (bs *BenchmarkSuite) WriteResults(path string) error {
	data, err := json.MarshalIndent(bs.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadResults 从文件加载测试结果
func (bs *BenchmarkSuite) LoadResults(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &bs.results)
}

// GenerateReport 生成性能测试报告
func (bs *BenchmarkSuite) GenerateReport() string {
	report := fmt.Sprintf("# Performance Benchmark Report\n\n")
	report += fmt.Sprintf("Suite: %s\n", bs.name)
	report += fmt.Sprintf("Benchmarks: %d\n\n", len(bs.benchmarks))
	report += fmt.Sprintf("Results: %d\n\n", len(bs.results))

	for _, result := range bs.results {
		report += fmt.Sprintf("## %s\n", result.Name)
		report += fmt.Sprintf("- Iterations: %d\n", result.Iterations)
		report += fmt.Sprintf("- Total Time: %v\n", result.Duration)
		report += fmt.Sprintf("- Avg Time: %v\n", result.AvgDuration)
		report += fmt.Sprintf("- Min Time: %v\n", result.MinDuration)
		report += fmt.Sprintf("- Max Time: %v\n", result.MaxDuration)
		report += fmt.Sprintf("- Memory: %d bytes\n", result.MemoryAlloc)
		report += fmt.Sprintf("- Ops/sec: %.0f\n\n", result.Stats["ops_per_sec"])
	}

	return report
}

// NewProfiler 创建性能分析器
func NewProfiler(outputPath string) *Profiler {
	profiler := &Profiler{
		enabled:      true,
		outputWriter: nil,
	}

	if outputPath != "" {
		file, err := os.Create(outputPath + ".cpu")
		if err == nil {
			profiler.cpuProfile = file
		}

		memFile, err := os.Create(outputPath + ".mem")
		if err == nil {
			profiler.memProfile = memFile
		}
	}

	return profiler
}

// Start 开始性能分析
func (p *Profiler) Start() error {
	if !p.enabled {
		return fmt.Errorf("profiler is disabled")
	}

	// CPU profiling
	if p.cpuProfile != nil {
		if err := p.startCPUProfile(); err != nil {
			return err
		}
	}

	// Memory profiling
	if p.memProfile != nil {
		if err := p.startMemoryProfile(); err != nil {
			return err
		}
	}

	return nil
}

// Stop 停止性能分析
func (p *Profiler) Stop() error {
	// Stop CPU profiling
	if p.cpuProfile != nil {
		if err := p.stopCPUProfile(); err != nil {
			return err
		}
	}

	// Stop memory profiling
	if p.memProfile != nil {
		if err := p.stopMemoryProfile(); err != nil {
			return err
		}
	}

	return nil
}

// startCPUProfile 启动CPU性能分析
func (p *Profiler) startCPUProfile() error {
	runtime.SetCPUProfileRate(100)
	return nil
}

// stopCPUProfile 停止CPU性能分析
func (p *Profiler) stopCPUProfile() error {
	runtime.SetCPUProfileRate(0)
	return p.cpuProfile.Close()
}

// startMemoryProfile 启动内存性能分析
func (p *Profiler) startMemoryProfile() error {
	runtime.MemProfileRate = 1 // 每次GC都记录
	return nil
}

// stopMemoryProfile 停止内存性能分析
func (p *Profiler) stopMemoryProfile() error {
	runtime.MemProfileRate = 0 // 关闭
	return p.memProfile.Close()
}

// WriteBenchmarkToFile 将基准测试写入文件
func WriteBenchmarkToFile(name string, fn func(*testing.B)) {
	testing.Benchmark(fn)
}

// getMemoryUsage 获取内存使用量
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// BenchmarkSimpleScript 基准测试简单脚本
func BenchmarkSimpleScript(b *testing.B) {
	// 设置脚本
	script := `
# Simple script
x=1
y=2
z=$((x + y))
`

	parser := parser.NewSimpleParser()
	_, err := parser.ParseString(script)
	if err != nil {
		b.Fatalf("Failed to parse script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 执行脚本
		// TODO: 实际执行逻辑
	}
}

// BenchmarkWithJIT 带JIT的基准测试
func BenchmarkWithJIT(b *testing.B) {
	// TODO: 实现JIT基准测试
}

// BenchmarkParallel 并行执行基准测试
func BenchmarkParallel(b *testing.B) {
	// TODO: 实现并行执行基准测试
}

// BenchmarkCache 缓存基准测试
func BenchmarkCache(b *testing.B) {
	// TODO: 实现缓存基准测试
}

// BenchmarkOptimized 优化后代码基准测试
func BenchmarkOptimized(b *testing.B) {
	// TODO: 实现优化代码基准测试
}
