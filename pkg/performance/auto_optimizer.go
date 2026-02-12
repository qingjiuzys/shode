package performance

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"
)

// AutoOptimizer 自动优化器
type AutoOptimizer struct {
	profile      *PerformanceProfile
	metrics      *PerformanceMetrics
	baseline     *PerformanceBaseline
	thresholds   *PerformanceThresholds
	history      []*PerformanceSnapshot
	analyzer     *PerformanceAnalyzer
	optimizer    *OptimizationEngine
	monitor      *RegressionMonitor
	mu           sync.RWMutex
	enabled      bool
	checkInterval time.Duration
}

// AutoPerformanceMetrics 自动优化器的性能指标
type AutoPerformanceMetrics struct {
	ExecutionTime    time.Duration
	MemoryUsage      uint64
	CPUUsage         float64
	GCCount          uint32
	GCPauseTime      time.Duration
	CacheHitRate     float64
	Throughput       float64
	Latency          time.Duration
	ErrorRate        float64
	LastUpdated      time.Time
}

// PerformanceBaseline 性能基线
type PerformanceBaseline struct {
	Name              string
	EstablishedAt     time.Time
	SampleSize        int
	MeanExecutionTime time.Duration
	MeanMemoryUsage   uint64
	Percentile50      time.Duration
	Percentile95      time.Duration
	Percentile99      time.Duration
	StdDev            time.Duration
	MinExecutionTime  time.Duration
	MaxExecutionTime  time.Duration
}

// PerformanceThresholds 性能阈值
type PerformanceThresholds struct {
	MaxExecutionTime  time.Duration
	MaxMemoryUsage    uint64
	MaxGCPause        time.Duration
	MinCacheHitRate   float64
	MinThroughput     float64
	MaxLatency        time.Duration
	MaxErrorRate      float64
	WarningThreshold  float64 // 0.0 - 1.0 (e.g., 0.8 = 80% of threshold)
	CriticalThreshold float64 // 0.0 - 1.0 (e.g., 0.95 = 95% of threshold)
}

// PerformanceSnapshot 性能快照
type PerformanceSnapshot struct {
	Timestamp     time.Time
	Metrics       *PerformanceMetrics
	Config        *PerformanceProfile
	Optimizations []string
	Alerts        []PerformanceAlert
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	Level       string // "info", "warning", "critical"
	Metric      string
	CurrentValue float64
	Threshold    float64
	Message      string
	Timestamp    time.Time
}

// PerformanceAnalyzer 性能分析器
type PerformanceAnalyzer struct {
	patterns     []*PerformancePattern
	anomalies    []*PerformanceAnomaly
	trends       map[string]*PerformanceTrend
	mu           sync.RWMutex
}

// PerformancePattern 性能模式
type PerformancePattern struct {
	Name        string
	Description string
	DetectedAt  time.Time
	Confidence  float64
	Metrics     map[string]interface{}
}

// PerformanceAnomaly 性能异常
type PerformanceAnomaly struct {
	Type        string
	Description string
	Severity    string
	DetectedAt  time.Time
	Metrics     map[string]interface{}
	SuggestedActions []string
}

// PerformanceTrend 性能趋势
type PerformanceTrend struct {
	Metric      string
	Direction   string // "improving", "degrading", "stable"
	ChangeRate  float64
	Confidence  float64
	StartedAt   time.Time
}

// OptimizationEngine 优化引擎
type OptimizationEngine struct {
	strategies   map[string]*OptimizationStrategy
	registry     *OptimizationRegistry
	applied      []string
	mu           sync.RWMutex
}

// OptimizationStrategy 优化策略
type OptimizationStrategy struct {
	Name          string
	Description   string
	Category      string // "memory", "cpu", "io", "cache", "parallel"
	Applicability func(*PerformanceMetrics) float64
	Apply         func(context.Context, *PerformanceProfile) (*OptimizationResult, error)
	Revert        func(context.Context, *PerformanceProfile) error
	Cost          int // 1-10, higher = more expensive
	Effectiveness float64 // 0.0 - 1.0
	SideEffects   []string
}

// OptimizationRegistry 优化注册表
type OptimizationRegistry struct {
	strategies   map[string]*OptimizationStrategy
	history      []*OptimizationRecord
	mu           sync.RWMutex
}

// OptimizationResult 优化结果
type OptimizationResult struct {
	StrategyName string
	AppliedAt    time.Time
	Improvement  float64 // percentage
	MetricsBefore *PerformanceMetrics
	MetricsAfter  *PerformanceMetrics
	Success      bool
	Error        error
	SideEffects  []string
}

// OptimizationRecord 优化记录
type OptimizationRecord struct {
	Strategy     string
	AppliedAt    time.Time
	RevertedAt   time.Time
	Improvement  float64
	Success      bool
	Duration     time.Duration
}

// RegressionMonitor 回归监控器
type RegressionMonitor struct {
	baselines    map[string]*PerformanceBaseline
	detectors    []*RegressionDetector
	alerts       []*RegressionAlert
	mu           sync.RWMutex
	enabled      bool
}

// RegressionDetector 回归检测器
type RegressionDetector struct {
	Name         string
	Metric       string
	Threshold    float64
	WindowSize   int
	Algorithm    string // "z-score", "iqr", "ewma"
	LastCheck    time.Time
}

// RegressionAlert 回归告警
type RegressionAlert struct {
	ID          string
	Metric      string
	Severity    string
	Baseline    float64
	CurrentValue float64
	ChangePercent float64
	DetectedAt  time.Time
	ConfirmedAt time.Time
	Resolved    bool
}

// NewAutoOptimizer 创建自动优化器
func NewAutoOptimizer(profile *PerformanceProfile) *AutoOptimizer {
	return &AutoOptimizer{
		profile:       profile,
		metrics:       &PerformanceMetrics{LastUpdated: time.Now()},
		baseline:      &PerformanceBaseline{},
		thresholds:    DefaultThresholds(),
		history:       make([]*PerformanceSnapshot, 0),
		analyzer:      NewPerformanceAnalyzer(),
		optimizer:     NewOptimizationEngine(),
		monitor:       NewRegressionMonitor(),
		enabled:       true,
		checkInterval: 30 * time.Second,
	}
}

// NewPerformanceAnalyzer 创建性能分析器
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{
		patterns:   make([]*PerformancePattern, 0),
		anomalies:  make([]*PerformanceAnomaly, 0),
		trends:     make(map[string]*PerformanceTrend),
	}
}

// NewOptimizationEngine 创建优化引擎
func NewOptimizationEngine() *OptimizationEngine {
	engine := &OptimizationEngine{
		strategies: make(map[string]*OptimizationStrategy),
		registry: &OptimizationRegistry{
			strategies: make(map[string]*OptimizationStrategy),
			history:   make([]*OptimizationRecord, 0),
		},
		applied: make([]string, 0),
	}

	// 注册内置优化策略
	engine.registerBuiltinStrategies()

	return engine
}

// NewRegressionMonitor 创建回归监控器
func NewRegressionMonitor() *RegressionMonitor {
	monitor := &RegressionMonitor{
		baselines: make(map[string]*PerformanceBaseline),
		detectors: make([]*RegressionDetector, 0),
		alerts:    make([]*RegressionAlert, 0),
		enabled:   true,
	}

	// 注册默认检测器
monitor.registerDetector(&RegressionDetector{
		Name:       "execution_time_regression",
		Metric:     "execution_time",
		Threshold:  0.2, // 20% increase triggers alert
		WindowSize: 10,
		Algorithm:  "ewma",
	})

	monitor.registerDetector(&RegressionDetector{
		Name:       "memory_leak_detector",
		Metric:     "memory_usage",
		Threshold:  0.5, // 50% increase triggers alert
		WindowSize: 20,
		Algorithm:  "iqr",
	})

	monitor.registerDetector(&RegressionDetector{
		Name:       "gc_pause_regression",
		Metric:     "gc_pause_time",
		Threshold:  0.3, // 30% increase triggers alert
		WindowSize: 15,
		Algorithm:  "z-score",
	})

	return monitor
}

// DefaultThresholds 默认阈值
func DefaultThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		MaxExecutionTime:  5 * time.Second,
		MaxMemoryUsage:    500 * 1024 * 1024, // 500MB
		MaxGCPause:        100 * time.Millisecond,
		MinCacheHitRate:   0.7, // 70%
		MinThroughput:     1000, // ops/sec
		MaxLatency:        100 * time.Millisecond,
		MaxErrorRate:      0.01, // 1%
		WarningThreshold:  0.8,
		CriticalThreshold: 0.95,
	}
}

// Start 启动自动优化器
func (ao *AutoOptimizer) Start(ctx context.Context) error {
	if !ao.enabled {
		return nil
	}

	// 启动定期检查
	go ao.runPeriodicChecks(ctx)

	return nil
}

// Stop 停止自动优化器
func (ao *AutoOptimizer) Stop() {
	ao.enabled = false
}

// runPeriodicChecks 运行定期检查
func (ao *AutoOptimizer) runPeriodicChecks(ctx context.Context) {
	ticker := time.NewTicker(ao.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ao.checkAndOptimize(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// checkAndOptimize 检查并优化
func (ao *AutoOptimizer) checkAndOptimize(ctx context.Context) {
	// 1. 采集当前指标
	metrics := ao.collectMetrics()

	// 2. 创建快照
	snapshot := &PerformanceSnapshot{
		Timestamp: time.Now(),
		Metrics:   metrics,
		Config:    ao.profile,
	}

	// 3. 检查阈值
	alerts := ao.checkThresholds(metrics)
	snapshot.Alerts = alerts

	// 4. 性能分析
	ao.analyzer.Analyze(metrics, ao.history)

	// 5. 回归检测
	ao.monitor.CheckForRegression(metrics)

	// 6. 应用优化
	if ao.shouldOptimize(alerts) {
		optimizations := ao.optimize(ctx, metrics)
		snapshot.Optimizations = optimizations
	}

	// 7. 更新基线（如果需要）
	ao.updateBaselineIfNeeded(metrics)

	// 8. 保存快照
	ao.mu.Lock()
	ao.history = append(ao.history, snapshot)
	// 保持历史记录在合理范围内
	if len(ao.history) > 1000 {
		ao.history = ao.history[len(ao.history)-1000:]
	}
	ao.mu.Unlock()
}

// collectMetrics 采集性能指标
func (ao *AutoOptimizer) collectMetrics() *PerformanceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := &PerformanceMetrics{LastUpdated: time.Now()}
	metrics.GCCount = m.NumGC
	metrics.MemoryUsage = int64(m.HeapAlloc)

	// 获取CPU使用率
	if cpuUsage, err := getCPUUsage(); err == nil {
		metrics.CPUUsage = cpuUsage
	}

	// 获取其他指标
	// 这些指标通常从 PerformanceManager 获取
	if ao.metrics.ExecutionTime > 0 {
		metrics.ExecutionTime = ao.metrics.ExecutionTime
	}
	if ao.metrics.CacheHitRate > 0 {
		metrics.CacheHitRate = ao.metrics.CacheHitRate
	}
	if ao.metrics.Throughput > 0 {
		metrics.Throughput = ao.metrics.Throughput
	}
	if ao.metrics.Latency > 0 {
		metrics.Latency = ao.metrics.Latency
	}
	if ao.metrics.ErrorRate > 0 {
		metrics.ErrorRate = ao.metrics.ErrorRate
	}

	return metrics
}

// checkThresholds 检查阈值
func (ao *AutoOptimizer) checkThresholds(metrics *PerformanceMetrics) []PerformanceAlert {
	alerts := []PerformanceAlert{}

	// 检查执行时间
	if metrics.ExecutionTime > ao.thresholds.MaxExecutionTime {
		alerts = append(alerts, PerformanceAlert{
			Level:        "critical",
			Metric:       "execution_time",
			CurrentValue: float64(metrics.ExecutionTime.Milliseconds()),
			Threshold:    float64(ao.thresholds.MaxExecutionTime.Milliseconds()),
			Message:      fmt.Sprintf("执行时间超过阈值: %v > %v", metrics.ExecutionTime, ao.thresholds.MaxExecutionTime),
			Timestamp:    time.Now(),
		})
	} else if metrics.ExecutionTime > time.Duration(float64(ao.thresholds.MaxExecutionTime)*ao.thresholds.WarningThreshold) {
		alerts = append(alerts, PerformanceAlert{
			Level:        "warning",
			Metric:       "execution_time",
			CurrentValue: float64(metrics.ExecutionTime.Milliseconds()),
			Threshold:    float64(ao.thresholds.MaxExecutionTime.Milliseconds()),
			Message:      fmt.Sprintf("执行时间接近阈值: %v", metrics.ExecutionTime),
			Timestamp:    time.Now(),
		})
	}

	// 检查内存使用
	if uint64(metrics.MemoryUsage) > ao.thresholds.MaxMemoryUsage {
		alerts = append(alerts, PerformanceAlert{
			Level:        "critical",
			Metric:       "memory_usage",
			CurrentValue: float64(metrics.MemoryUsage),
			Threshold:    float64(ao.thresholds.MaxMemoryUsage),
			Message:      fmt.Sprintf("内存使用超过阈值: %d > %d", metrics.MemoryUsage, ao.thresholds.MaxMemoryUsage),
			Timestamp:    time.Now(),
		})
	}

	// 检查缓存命中率
	if metrics.CacheHitRate > 0 && metrics.CacheHitRate < ao.thresholds.MinCacheHitRate {
		alerts = append(alerts, PerformanceAlert{
			Level:        "warning",
			Metric:       "cache_hit_rate",
			CurrentValue: metrics.CacheHitRate,
			Threshold:    ao.thresholds.MinCacheHitRate,
			Message:      fmt.Sprintf("缓存命中率低于阈值: %.2f%% < %.2f%%", metrics.CacheHitRate*100, ao.thresholds.MinCacheHitRate*100),
			Timestamp:    time.Now(),
		})
	}

	// 检查错误率
	if metrics.ErrorRate > ao.thresholds.MaxErrorRate {
		alerts = append(alerts, PerformanceAlert{
			Level:        "critical",
			Metric:       "error_rate",
			CurrentValue: metrics.ErrorRate,
			Threshold:    ao.thresholds.MaxErrorRate,
			Message:      fmt.Sprintf("错误率超过阈值: %.2f%% > %.2f%%", metrics.ErrorRate*100, ao.thresholds.MaxErrorRate*100),
			Timestamp:    time.Now(),
		})
	}

	return alerts
}

// shouldOptimize 判断是否需要优化
func (ao *AutoOptimizer) shouldOptimize(alerts []PerformanceAlert) bool {
	for _, alert := range alerts {
		if alert.Level == "critical" || alert.Level == "warning" {
			return true
		}
	}
	return false
}

// optimize 执行优化
func (ao *AutoOptimizer) optimize(ctx context.Context, metrics *PerformanceMetrics) []string {
	applied := []string{}

	// 获取适用的优化策略
	strategies := ao.optimizer.GetApplicableStrategies(metrics)

	// 按成本和效果排序
	strategies = ao.sortStrategiesByPriority(strategies)

	// 应用优化策略
	for _, strategy := range strategies {
		// 检查是否已经应用过
		if ao.isStrategyApplied(strategy.Name) {
			continue
		}

		// 应用优化
		result, err := ao.optimizer.ApplyStrategy(ctx, strategy, ao.profile)
		if err != nil {
			continue
		}

		if result.Success {
			applied = append(applied, strategy.Name)
			ao.markStrategyApplied(strategy.Name)

			// 如果改善明显，可以提前停止
			if result.Improvement > 0.2 { // 20% improvement
				break
			}
		}
	}

	return applied
}

// sortStrategiesByPriority 按优先级排序策略
func (ao *AutoOptimizer) sortStrategiesByPriority(strategies []*OptimizationStrategy) []*OptimizationStrategy {
	// 简单排序：按效果/成本比例
	sorted := make([]*OptimizationStrategy, len(strategies))
	copy(sorted, strategies)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			scoreI := sorted[i].Effectiveness / float64(sorted[i].Cost)
			scoreJ := sorted[j].Effectiveness / float64(sorted[j].Cost)
			if scoreJ > scoreI {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// isStrategyApplied 检查策略是否已应用
func (ao *AutoOptimizer) isStrategyApplied(name string) bool {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	for _, applied := range ao.optimizer.applied {
		if applied == name {
			return true
		}
	}
	return false
}

// markStrategyApplied 标记策略已应用
func (ao *AutoOptimizer) markStrategyApplied(name string) {
	ao.mu.Lock()
	defer ao.mu.Unlock()

	ao.optimizer.applied = append(ao.optimizer.applied, name)
}

// updateBaselineIfNeeded 更新基线
func (ao *AutoOptimizer) updateBaselineIfNeeded(metrics *PerformanceMetrics) {
	// 如果没有基线，创建一个
	if ao.baseline.SampleSize == 0 {
		ao.establishBaseline(metrics)
		return
	}

	// 如果基线样本不足，添加样本
	if ao.baseline.SampleSize < 100 {
		ao.baseline.SampleSize++
		// 更新统计
		ao.updateBaselineStats(metrics)
	}
}

// establishBaseline 建立基线
func (ao *AutoOptimizer) establishBaseline(metrics *PerformanceMetrics) {
	ao.baseline = &PerformanceBaseline{
		Name:             fmt.Sprintf("baseline_%d", time.Now().Unix()),
		EstablishedAt:    time.Now(),
		SampleSize:       1,
		MeanExecutionTime: metrics.ExecutionTime,
		MeanMemoryUsage:  uint64(metrics.MemoryUsage),
		MinExecutionTime: metrics.ExecutionTime,
		MaxExecutionTime: metrics.ExecutionTime,
	}
}

// updateBaselineStats 更新基线统计
func (ao *AutoOptimizer) updateBaselineStats(metrics *PerformanceMetrics) {
	// 更新平均值（简化版，应该用更精确的在线算法）
	oldMean := float64(ao.baseline.MeanExecutionTime)
	newMean := oldMean + (float64(metrics.ExecutionTime)-oldMean)/float64(ao.baseline.SampleSize)
	ao.baseline.MeanExecutionTime = time.Duration(newMean)

	// 更新最小/最大值
	if metrics.ExecutionTime < ao.baseline.MinExecutionTime {
		ao.baseline.MinExecutionTime = metrics.ExecutionTime
	}
	if metrics.ExecutionTime > ao.baseline.MaxExecutionTime {
		ao.baseline.MaxExecutionTime = metrics.ExecutionTime
	}
}

// registerBuiltinStrategies 注册内置优化策略
func (oe *OptimizationEngine) registerBuiltinStrategies() {
	// 内存优化策略
	oe.RegisterStrategy(&OptimizationStrategy{
		Name:        "aggressive_gc",
		Description: "启用激进的垃圾回收",
		Category:    "memory",
		Applicability: func(metrics *PerformanceMetrics) float64 {
			if metrics.MemoryUsage > 100*1024*1024 { // > 100MB
				return 0.9
			}
			return 0.3
		},
		Apply: func(ctx context.Context, profile *PerformanceProfile) (*OptimizationResult, error) {
			// 启用激进GC
			runtime.GC()

			return &OptimizationResult{
				StrategyName: "aggressive_gc",
				AppliedAt:    time.Now(),
				Success:      true,
			}, nil
		},
		Revert: func(ctx context.Context, profile *PerformanceProfile) error {
			return nil
		},
		Cost:          3,
		Effectiveness: 0.7,
	})

	// 缓存优化策略
	oe.RegisterStrategy(&OptimizationStrategy{
		Name:        "enable_cache",
		Description: "启用缓存优化",
		Category:    "cache",
		Applicability: func(metrics *PerformanceMetrics) float64 {
			if metrics.CacheHitRate < 0.5 {
				return 0.95
			}
			return 0.2
		},
		Apply: func(ctx context.Context, profile *PerformanceProfile) (*OptimizationResult, error) {
			profile.EnableCache = true
			return &OptimizationResult{
				StrategyName: "enable_cache",
				AppliedAt:    time.Now(),
				Success:      true,
			}, nil
		},
		Revert: func(ctx context.Context, profile *PerformanceProfile) error {
			profile.EnableCache = false
			return nil
		},
		Cost:          2,
		Effectiveness: 0.85,
	})

	// 并行优化策略
	oe.RegisterStrategy(&OptimizationStrategy{
		Name:        "increase_parallelism",
		Description: "增加并行度",
		Category:    "parallel",
		Applicability: func(metrics *PerformanceMetrics) float64 {
			numCPU := runtime.NumCPU()
			// 简化实现，总是返回0.8
			_ = numCPU
			return 0.8
		},
		Apply: func(ctx context.Context, profile *PerformanceProfile) (*OptimizationResult, error) {
			numCPU := runtime.NumCPU()
			profile.MaxWorkers = numCPU
			return &OptimizationResult{
				StrategyName: "increase_parallelism",
				AppliedAt:    time.Now(),
				Success:      true,
			}, nil
		},
		Revert: func(ctx context.Context, profile *PerformanceProfile) error {
			profile.MaxWorkers = 4
			return nil
		},
		Cost:          4,
		Effectiveness: 0.75,
	})

	// JIT优化策略
	oe.RegisterStrategy(&OptimizationStrategy{
		Name:        "enable_jit",
		Description: "启用JIT编译",
		Category:    "cpu",
		Applicability: func(metrics *PerformanceMetrics) float64 {
			if metrics.ExecutionTime > 100*time.Millisecond {
				return 0.9
			}
			return 0.4
		},
		Apply: func(ctx context.Context, profile *PerformanceProfile) (*OptimizationResult, error) {
			profile.EnableJIT = true
			return &OptimizationResult{
				StrategyName: "enable_jit",
				AppliedAt:    time.Now(),
				Success:      true,
			}, nil
		},
		Revert: func(ctx context.Context, profile *PerformanceProfile) error {
			profile.EnableJIT = false
			return nil
		},
		Cost:          5,
		Effectiveness: 0.8,
	})
}

// RegisterStrategy 注册优化策略
func (oe *OptimizationEngine) RegisterStrategy(strategy *OptimizationStrategy) {
	oe.mu.Lock()
	defer oe.mu.Unlock()

	oe.strategies[strategy.Name] = strategy
	oe.registry.strategies[strategy.Name] = strategy
}

// GetApplicableStrategies 获取适用的优化策略
func (oe *OptimizationEngine) GetApplicableStrategies(metrics *PerformanceMetrics) []*OptimizationStrategy {
	oe.mu.RLock()
	defer oe.mu.RUnlock()

	applicable := make([]*OptimizationStrategy, 0)
	for _, strategy := range oe.strategies {
		score := strategy.Applicability(metrics)
		if score > 0.5 {
			applicable = append(applicable, strategy)
		}
	}

	return applicable
}

// ApplyStrategy 应用优化策略
func (oe *OptimizationEngine) ApplyStrategy(ctx context.Context, strategy *OptimizationStrategy, profile *PerformanceProfile) (*OptimizationResult, error) {
	oe.mu.Lock()
	defer oe.mu.Unlock()

	// 应用策略
	result, err := strategy.Apply(ctx, profile)
	if err != nil {
		return &OptimizationResult{
			StrategyName: strategy.Name,
			AppliedAt:    time.Now(),
			Success:      false,
			Error:        err,
		}, err
	}

	// 记录到历史
	record := &OptimizationRecord{
		Strategy:    strategy.Name,
		AppliedAt:   time.Now(),
		Success:     result.Success,
	}
	oe.registry.history = append(oe.registry.history, record)

	return result, nil
}

// Analyze 分析性能
func (pa *PerformanceAnalyzer) Analyze(metrics *PerformanceMetrics, history []*PerformanceSnapshot) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	// 检测异常
	if len(history) > 0 {
		pa.detectAnomalies(metrics, history)
	}

	// 检测模式
	pa.detectPatterns(metrics, history)

	// 分析趋势
	pa.analyzeTrends(metrics, history)
}

// detectAnomalies 检测异常
func (pa *PerformanceAnalyzer) detectAnomalies(metrics *PerformanceMetrics, history []*PerformanceSnapshot) {
	if len(history) < 10 {
		return
	}

	// 计算执行时间的平均值和标准差
	times := make([]time.Duration, len(history))
	for i, snapshot := range history {
		times[i] = snapshot.Metrics.ExecutionTime
	}

	mean := calculateMean(times)
	stddev := calculateStdDev(times, mean)

	// 检查当前指标是否异常（使用3-sigma规则）
	zScore := float64(metrics.ExecutionTime-mean) / float64(stddev)
	if math.Abs(zScore) > 3.0 {
		pa.anomalies = append(pa.anomalies, &PerformanceAnomaly{
			Type:        "execution_time_anomaly",
			Description: fmt.Sprintf("执行时间异常: z-score = %.2f", zScore),
			Severity:    "high",
			DetectedAt:  time.Now(),
			Metrics: map[string]interface{}{
				"z_score": zScore,
				"current": metrics.ExecutionTime,
				"mean":    mean,
				"stddev":  stddev,
			},
			SuggestedActions: []string{
				"检查是否有配置变更",
				"检查系统资源使用情况",
				"检查是否有外部依赖问题",
			},
		})
	}
}

// detectPatterns 检测模式
func (pa *PerformanceAnalyzer) detectPatterns(metrics *PerformanceMetrics, history []*PerformanceSnapshot) {
	// 检测内存泄漏模式
	if len(history) >= 20 {
		memoryTrend := pa.calculateMemoryTrend(history)
		if memoryTrend > 0.1 { // 持续增长
			pa.patterns = append(pa.patterns, &PerformancePattern{
				Name:        "memory_leak_pattern",
				Description: "检测到内存持续增长模式",
				DetectedAt:  time.Now(),
				Confidence:  0.8,
				Metrics: map[string]interface{}{
					"trend": memoryTrend,
				},
			})
		}
	}
}

// analyzeTrends 分析趋势
func (pa *PerformanceAnalyzer) analyzeTrends(metrics *PerformanceMetrics, history []*PerformanceSnapshot) {
	if len(history) < 10 {
		return
	}

	// 分析执行时间趋势
	execTimes := make([]time.Duration, len(history))
	for i, snapshot := range history {
		execTimes[i] = snapshot.Metrics.ExecutionTime
	}

	// 计算趋势（简单线性回归）
	trend := pa.calculateTrend(execTimes)

	direction := "stable"
	if trend > 0.05 {
		direction = "degrading"
	} else if trend < -0.05 {
		direction = "improving"
	}

	pa.trends["execution_time"] = &PerformanceTrend{
		Metric:     "execution_time",
		Direction:  direction,
		ChangeRate: trend,
		Confidence: 0.7,
		StartedAt:  history[0].Timestamp,
	}
}

// calculateMemoryTrend 计算内存趋势
func (pa *PerformanceAnalyzer) calculateMemoryTrend(history []*PerformanceSnapshot) float64 {
	if len(history) < 2 {
		return 0
	}

	// 简单线性回归
	n := float64(len(history))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, snapshot := range history {
		x := float64(i)
		y := float64(snapshot.Metrics.MemoryUsage)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	meanY := sumY / n

	// 归一化斜率
	if meanY > 0 {
		return slope / meanY
	}
	return slope
}

// calculateTrend 计算趋势
func (pa *PerformanceAnalyzer) calculateTrend(values []time.Duration) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, v := range values {
		x := float64(i)
		y := float64(v)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	meanY := sumY / n

	if meanY > 0 {
		return slope / meanY
	}
	return slope
}

// registerDetector 注册检测器
func (rm *RegressionMonitor) registerDetector(detector *RegressionDetector) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.detectors = append(rm.detectors, detector)
}

// CheckForRegression 检查回归
func (rm *RegressionMonitor) CheckForRegression(metrics *PerformanceMetrics) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for _, detector := range rm.detectors {
		alert := rm.checkDetector(detector, metrics)
		if alert != nil {
			rm.alerts = append(rm.alerts, alert)
		}
	}
}

// checkDetector 检查单个检测器
func (rm *RegressionMonitor) checkDetector(detector *RegressionDetector, metrics *PerformanceMetrics) *RegressionAlert {
	// 简化实现，实际应该基于历史数据
	var currentValue, baseline float64

	switch detector.Metric {
	case "execution_time":
		currentValue = float64(metrics.ExecutionTime)
		if base, exists := rm.baselines["execution"]; exists {
			baseline = float64(base.MeanExecutionTime)
		}
	case "memory_usage":
		currentValue = float64(metrics.MemoryUsage)
		if base, exists := rm.baselines["memory"]; exists {
			baseline = float64(base.MeanMemoryUsage)
		}
	}

	if baseline > 0 {
		changePercent := (currentValue - baseline) / baseline
		if math.Abs(changePercent) > detector.Threshold {
			return &RegressionAlert{
				ID:            fmt.Sprintf("alert_%d", time.Now().UnixNano()),
				Metric:        detector.Metric,
				Severity:      "warning",
				Baseline:      baseline,
				CurrentValue: currentValue,
				ChangePercent: changePercent,
				DetectedAt:    time.Now(),
				Resolved:      false,
			}
		}
	}

	return nil
}

// SetBaseline 设置基线
func (rm *RegressionMonitor) SetBaseline(name string, baseline *PerformanceBaseline) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.baselines[name] = baseline
}

// GetAlerts 获取告警
func (rm *RegressionMonitor) GetAlerts() []*RegressionAlert {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	alerts := make([]*RegressionAlert, len(rm.alerts))
	copy(alerts, rm.alerts)
	return alerts
}

// calculateMean 计算平均值
func calculateMean(values []time.Duration) time.Duration {
	if len(values) == 0 {
		return 0
	}

	sum := time.Duration(0)
	for _, v := range values {
		sum += v
	}
	return sum / time.Duration(len(values))
}

// calculateStdDev 计算标准差
func calculateStdDev(values []time.Duration, mean time.Duration) time.Duration {
	if len(values) < 2 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		diff := float64(v - mean)
		sum += diff * diff
	}

	variance := sum / float64(len(values))
	return time.Duration(math.Sqrt(variance))
}

// getCPUUsage 获取CPU使用率（简化实现）
func getCPUUsage() (float64, error) {
	// 简化实现，实际应该使用更精确的方法
	return 0.0, nil
}

// GetPerformanceSnapshot 获取性能快照
func (ao *AutoOptimizer) GetPerformanceSnapshot() *PerformanceSnapshot {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	if len(ao.history) == 0 {
		return nil
	}

	return ao.history[len(ao.history)-1]
}

// GetHistory 获取历史快照
func (ao *AutoOptimizer) GetHistory() []*PerformanceSnapshot {
	ao.mu.RLock()
	defer ao.mu.RUnlock()

	history := make([]*PerformanceSnapshot, len(ao.history))
	copy(history, ao.history)
	return history
}

// GetBaseline 获取基线
func (ao *AutoOptimizer) GetBaseline() *PerformanceBaseline {
	return ao.baseline
}

// GetThresholds 获取阈值
func (ao *AutoOptimizer) GetThresholds() *PerformanceThresholds {
	return ao.thresholds
}

// SetThresholds 设置阈值
func (ao *AutoOptimizer) SetThresholds(thresholds *PerformanceThresholds) {
	ao.thresholds = thresholds
}

// GetAlerts 获取告警
func (ao *AutoOptimizer) GetAlerts() []PerformanceAlert {
	snapshot := ao.GetPerformanceSnapshot()
	if snapshot == nil {
		return []PerformanceAlert{}
	}

	return snapshot.Alerts
}

// GetTrends 获取趋势
func (ao *AutoOptimizer) GetTrends() map[string]*PerformanceTrend {
	ao.analyzer.mu.RLock()
	defer ao.analyzer.mu.RUnlock()

	return ao.analyzer.trends
}

// GetAnomalies 获取异常
func (ao *AutoOptimizer) GetAnomalies() []*PerformanceAnomaly {
	ao.analyzer.mu.RLock()
	defer ao.analyzer.mu.RUnlock()

	return ao.analyzer.anomalies
}

// GetRegressionAlerts 获取回归告警
func (ao *AutoOptimizer) GetRegressionAlerts() []*RegressionAlert {
	return ao.monitor.GetAlerts()
}
