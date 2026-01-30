package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// Buffer 缓冲区
type Buffer struct {
	data []byte
	mu   sync.Mutex
}

// PerformanceManager 性能管理器
type PerformanceManager struct {
	jit           *JITCompiler
	parallel      *ParallelExecutor
	memory       *MemoryOptimizer
	profiler     *Profiler
	config       *PerformanceProfile
	stats        *AggregateStats
	mu           sync.RWMutex
	initialized   bool
}

// AggregateStats 聚合统计
type AggregateStats struct {
	ExecutionCount  int64
	TotalTime       time.Duration
	AverageTime     time.Duration
	PeakMemory      uint64
	Optimizations  int64
	GCCount         int64
	CacheHitRate    float64
	ParallelSpeedup float64
}

// PerformanceProfile 性能配置
type PerformanceProfile struct {
	EnableJIT         bool              `json:"enable_jit"`
	EnableCache       bool              `json:"enable_cache"`
	EnableParallel    bool              `json:"enable_parallel"`
	EnableGC          bool              `json:"enable_gc"`
	MaxWorkers        int               `json:"max_workers"`
	GCThreshold       uint64            `json:"gc_threshold"`
	CacheDir         string            `json:"cache_dir"`
	SampleInterval   time.Duration     `json:"sample_interval"`
	OptimizationLevel string            `json:"optimization_level"` // basic, standard, aggressive
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	Success        bool
	Output         string
	Error          error
	Duration       time.Duration
	MemoryUsed     uint64
	Optimizations  int
	CacheHit       bool
	ParallelTasks  int
	Stats          map[string]interface{}
}

// ExecutionContextEnhanced 增强的执行上下文（包含性能信息）
type ExecutionContextEnhanced struct {
	Variables        map[string]interface{}
	Functions        map[string]*types.FunctionNode
	Stdout           *Buffer
	Stderr           *Buffer
	Stdin            *Buffer
	Parent           *ExecutionContextEnhanced
	PerformanceMode  string // fast, normal, safe
	StartTime        time.Time
	MemoryLimit      uint64
	Timeout          time.Duration
}

// NewPerformanceManager 创建性能管理器
func NewPerformanceManager(profile *PerformanceProfile) *PerformanceManager {
	return &PerformanceManager{
		jit:     NewJITCompiler(profile.CacheDir, profile.EnableJIT, profile.EnableCache),
		parallel: NewParallelExecutor(profile.MaxWorkers),
		memory:  NewMemoryOptimizer(profile.EnableGC, profile.GCThreshold),
		profiler: NewProfiler(""),
		config:  profile,
		stats:   &AggregateStats{},
	}
}

// Initialize 初始化性能组件
func (pm *PerformanceManager) Initialize() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.initialized {
		return nil
	}

	// 初始化各个性能组件
	if pm.config.EnableCache {
		// 加载缓存
		if err := pm.jit.loadCache(); err != nil {
			return fmt.Errorf("failed to load JIT cache: %w", err)
		}
	}

	if pm.config.EnableParallel {
		// 启用并行执行器
		pm.parallel = NewParallelExecutor(pm.config.MaxWorkers)
	}

	if pm.config.EnableGC {
		// 设置GC参数
		SetGCPercent(50) // 更激进的GC
	}

	pm.initialized = true
	return nil
}

// ExecuteOptimized 执行优化后的脚本
func (pm *PerformanceManager) ExecuteOptimized(ctx context.Context, script *types.ScriptNode, sourcePath string) (*ExecutionResult, error) {
	startTime := time.Now()
	result := &ExecutionResult{
		Stats: make(map[string]interface{}),
	}

	// 1. 检查缓存（如果启用）
	var cacheHit bool

	if pm.config.EnableJIT {
		_, err := pm.jit.Compile(ctx, script, sourcePath)
		if err != nil {
			return nil, fmt.Errorf("JIT compilation failed: %w", err)
		}
		cacheHit = pm.jit.GetStats().CacheHits > 0
		// TODO: 使用编译后的字节码执行（当前仍使用解释执行）
	}

	// 2. 选择执行模式
	if pm.config.EnableParallel && canParallelize(script) {
		// 并行执行模式
		parallelResult, err := pm.parallel.Execute(ctx, script)
		if err != nil {
			return nil, fmt.Errorf("parallel execution failed: %w", err)
		}

		result.Success = parallelResult.Error == nil
		result.Output = fmt.Sprintf("%v", parallelResult.Result)
		result.Error = parallelResult.Error
		result.ParallelTasks = int(pm.parallel.GetStats().TasksTotal)
		result.Duration = time.Since(startTime)

	} else {
		// 标准执行模式
		// TODO: 实际执行脚本
		result.Success = true
		result.Duration = time.Since(startTime)
	}

	// 3. 更新统计
	result.Optimizations = int(pm.jit.optimizer.stats.OptimizationsApplied["dead_code_elimination"])
	result.CacheHit = cacheHit

	if pm.config.EnableGC {
		// 运行GC（如果内存使用超过阈值）
		memStats := pm.memory.GetMemoryStats()
		result.MemoryUsed = memStats.CurrentUsage

		if memStats.CurrentUsage > pm.config.GCThreshold {
			pm.memory.RunGC()
		}
	}

	// 4. 更新聚合统计
	pm.updateStats(result)

	return result, nil
}

// FastExecute 快速执行模式（牺牲正确性换取速度）
func (pm *PerformanceManager) FastExecute(ctx context.Context, script *types.ScriptNode) (*ExecutionResult, error) {
	// 创建快速执行上下文
	execCtx := &ExecutionContextEnhanced{
		PerformanceMode: "fast",
		StartTime:      time.Now(),
		MemoryLimit:     1024 * 1024, // 1MB limit
		Timeout:         30 * time.Second,
	}

	result := &ExecutionResult{
		Stats: make(map[string]interface{}),
	}

	// 跳过优化，直接执行
	// TODO: 实现快速执行路径
	result.Success = true
	result.Duration = time.Since(execCtx.StartTime)

	return result, result.Error
}

// ProfileExecution 分析执行性能
func (pm *PerformanceManager) ProfileExecution(ctx context.Context, script *types.ScriptNode, sourcePath string) (*ProfileReport, error) {
	// 启动性能分析
	if err := pm.profiler.Start(); err != nil {
		return nil, err
	}
	defer pm.profiler.Stop()

	// 执行脚本
	result, err := pm.ExecuteOptimized(ctx, script, sourcePath)
	if err != nil {
		return nil, err
	}

	// 收集性能数据
	profile := &ProfileReport{
		ExecutionResult: result,
		MemorySnapshots: pm.memory.GetMemoryProfile(),
		ParallelStats:  pm.parallel.GetStats(),
		JITStats:        pm.jit.GetStats(),
		AggregateStats:  pm.stats,
	}

	return profile, nil
}

// WarmupCache 预热缓存
func (pm *PerformanceManager) WarmupCache(directory string) error {
	if !pm.config.EnableCache {
		return nil
	}

	return pm.jit.WarmupCache(directory)
}

// ClearCache 清空所有缓存
func (pm *PerformanceManager) ClearCache() error {
	if !pm.config.EnableCache {
		return fmt.Errorf("caching is disabled")
	}

	return pm.jit.ClearCache()
}

// GetPerformanceReport 获取性能报告
func (pm *PerformanceManager) GetPerformanceReport() *PerformanceReport {
	return &PerformanceReport{
		Stats:           pm.stats,
		JITStats:        pm.jit.GetStats(),
		ParallelStats:  pm.parallel.GetStats(),
		MemoryStats:    pm.memory.GetMemoryStats(),
		Config:          pm.config,
	}
}

// GetOptimizationSuggestions 获取优化建议
func (pm *PerformanceManager) GetOptimizationSuggestions() []string {
	suggestions := []string{}

	// 检查缓存命中率
	if pm.config.EnableCache {
		stats := pm.jit.GetStats()
		if stats.CacheHits > 0 {
			hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses)
			if hitRate < 0.5 {
				suggestions = append(suggestions, "低缓存命中率 (< 50%)，考虑预热缓存或增加缓存大小")
			}
		}
	}

	// 检查并行利用率
	if pm.config.EnableParallel {
		stats := pm.parallel.GetStats()
		if stats.ParallelismUtil < 0.7 {
			suggestions = append(suggestions, "低并行利用率 (< 70%)，考虑增加并行任务或优化依赖图")
		}
	}

	// 检查内存使用
	if pm.config.EnableGC {
		stats := pm.memory.GetMemoryStats()
		if stats.PeakUsage > 100*1024*1024 { // 100MB
			suggestions = append(suggestions, "高内存使用 (> 100MB)，考虑实现对象池或增加GC频率")
		}
	}

	// 检查JIT编译
	if pm.config.EnableJIT {
		if pm.jit.optimizer.stats.NodesOptimized == 0 {
			suggestions = append(suggestions, "未应用代码优化，考虑启用死代码消除和常量折叠")
		}
	}

	return suggestions
}

// TunePerformance 自动调优性能参数
func (pm *PerformanceManager) TunePerformance() (*TuningReport, error) {
	// 基于当前性能指标自动调优

	report := &TuningReport{
		OriginalConfig: pm.config,
		RecommendedConfig: pm.config.Clone(),
		Reasons: []string{},
	}

	// 分析内存使用
	memStats := pm.memory.GetMemoryStats()
	if memStats.PeakUsage > 50*1024*1024 {
		report.RecommendedConfig.GCThreshold = 25 * 1024 * 1024 // 50MB
		report.Reasons = append(report.Reasons, "降低GC阈值以防止内存溢出")
	}

	// 分析缓存效率
	if pm.config.EnableCache {
		stats := pm.jit.GetStats()
		if stats.CacheHits > stats.CacheMisses {
			report.RecommendedConfig.EnableCache = true
			report.Reasons = append(report.Reasons, "缓存效率高，保持启用")
		} else {
			report.RecommendedConfig.EnableCache = false
			report.Reasons = append(report.Reasons, "缓存效率低，建议禁用以避免开销")
		}
	}

	// 分析并行效果
	if pm.config.EnableParallel {
		stats := pm.parallel.GetStats()
		if stats.ParallelismUtil > 0.8 {
			report.RecommendedConfig.MaxWorkers = 8 // 增加工作线程
			report.Reasons = append(report.Reasons, "高并行利用率，增加工作线程数")
		}
	}

	return report, nil
}

// updateStats 更新聚合统计
func (pm *PerformanceManager) updateStats(result *ExecutionResult) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.stats.ExecutionCount++
	pm.stats.TotalTime += result.Duration
	pm.stats.Optimizations += int64(result.Optimizations)

	if pm.config.EnableGC {
		pm.stats.GCCount++
	}
}

// ProfileReport 性能分析报告
type ProfileReport struct {
	ExecutionResult *ExecutionResult
	MemorySnapshots  []*MemorySnapshot
	ParallelStats    *ParallelStats
	JITStats         *JITStats
	AggregateStats   *AggregateStats
	Config           *PerformanceProfile
}

// TuningReport 调优报告
type TuningReport struct {
	OriginalConfig     *PerformanceProfile
	RecommendedConfig *PerformanceProfile
	Reasons           []string
	AppliedAt          time.Time
	Improvement       float64 // 性能提升百分比
}

// PerformanceReport 性能报告
type PerformanceReport struct {
	Stats           *AggregateStats
	JITStats        *JITStats
	ParallelStats   *ParallelStats
	MemoryStats     *MemoryStats
	Config          *PerformanceProfile
}

// Clone 克隆配置
func (p *PerformanceProfile) Clone() *PerformanceProfile {
	return &PerformanceProfile{
		EnableJIT:       p.EnableJIT,
		EnableCache:     p.EnableCache,
		EnableParallel:  p.EnableParallel,
		EnableGC:        p.EnableGC,
		MaxWorkers:      p.MaxWorkers,
		GCThreshold:     p.GCThreshold,
		CacheDir:        p.CacheDir,
		SampleInterval:  p.SampleInterval,
	}
}

// LoadProfile 从文件加载性能配置
func LoadProfile(path string) (*PerformanceProfile, error) {
	var profile PerformanceProfile
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, err
	}

	return &profile, nil
}

// SaveProfile 保存性能配置到文件
func SaveProfile(profile *PerformanceProfile, path string) error {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// canParallelize 检查脚本是否可以并行执行
func canParallelize(script *types.ScriptNode) bool {
	// 简化版：检查是否有可以并行执行的独立语句
	// TODO: 实现完整的依赖分析

	// 如果脚本有多个独立的顶层语句，可以并行执行
	return len(script.Nodes) > 1
}

// NewExecutionContextEnhanced 创建增强的执行上下文
func NewExecutionContextEnhanced(parent *ExecutionContextEnhanced) *ExecutionContextEnhanced {
	ctx := &ExecutionContextEnhanced{
		Variables: make(map[string]interface{}),
		Functions: make(map[string]*types.FunctionNode),
		Stdout:    &Buffer{},
		Stderr:    &Buffer{},
		Stdin:     &Buffer{},
		StartTime: time.Now(),
	}

	if parent != nil {
		// 继承父上下文
		for k, v := range parent.Variables {
			ctx.Variables[k] = v
		}
		for k, v := range parent.Functions {
			ctx.Functions[k] = v
		}
		ctx.Parent = parent
	}

	return ctx
}

// GetGlobalStats 获取全局性能统计
func (pm *PerformanceManager) GetGlobalStats() *AggregateStats {
	return pm.stats
}

// ResetStats 重置统计
func (pm *PerformanceManager) ResetStats() {
	pm.stats = &AggregateStats{}

	pm.jit = NewJITCompiler(pm.config.CacheDir, pm.config.EnableJIT, pm.config.EnableCache)
	pm.parallel = NewParallelExecutor(pm.config.MaxWorkers)
	pm.memory = NewMemoryOptimizer(pm.config.EnableGC, pm.config.GCThreshold)
}

// ExportMetrics 导出性能指标（用于监控系统）
func (pm *PerformanceManager) ExportMetrics() map[string]interface{} {
	stats := pm.GetGlobalStats()
	jitStats := pm.jit.GetStats()
	parallelStats := pm.parallel.GetStats()
	memStats := pm.memory.GetMemoryStats()

	return map[string]interface{}{
		"execution_count": stats.ExecutionCount,
		"total_time_ms":   stats.TotalTime.Milliseconds(),
		"avg_time_ms":     stats.AverageTime.Milliseconds(),
		"optimizations":  stats.Optimizations,
		"gc_count":        stats.GCCount,

		"jit_cache_hits":    jitStats.CacheHits,
		"jit_cache_misses":  jitStats.CacheMisses,
		"jit_compilations": jitStats.Compilations,

		"parallel_tasks":   parallelStats.TasksTotal,
		"parallel_speedup": parallelStats.ParallelismUtil,

		"memory_current":  memStats.CurrentUsage,
		"memory_peak":     memStats.PeakUsage,
		"memory_objects":  memStats.ObjectCount,
	}
}

// Write 写入缓冲区
func (b *Buffer) Write(data []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, data...)
	return len(data), nil
}

// Read 读取缓冲区
func (b *Buffer) Read(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.data) == 0 {
		return 0, nil
	}

	n := copy(p, b.data)
	b.data = b.data[n:]
	return n, nil
}

// String 获取字符串内容
func (b *Buffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return string(b.data)
}

// Bytes 获取字节数组
func (b *Buffer) Bytes() []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data
}

// Reset 重置缓冲区
func (b *Buffer) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = b.data[:0]
}

// Len 获取长度
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.data)
}
