package metrics

import (
	"fmt"
	"sync"
	"time"
)

// MetricsCollector collects and tracks execution metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Command execution metrics
	commandExecutions   int64
	commandSuccesses   int64
	commandFailures    int64
	commandTimeouts    int64
	commandDurations   []time.Duration
	totalCommandTime   time.Duration

	// Cache metrics
	cacheHits   int64
	cacheMisses int64

	// Process pool metrics
	processPoolSize     int
	processPoolActive   int
	processPoolIdle    int

	// Error metrics
	errorCounts map[string]int64

	// Pipeline metrics
	pipelineExecutions int64
	pipelineSuccesses  int64
	pipelineFailures   int64

	// Loop metrics
	loopExecutions int64
	loopIterations int64

	// Start time for uptime tracking
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		commandDurations: make([]time.Duration, 0, 1000),
		errorCounts:      make(map[string]int64),
		startTime:         time.Now(),
	}
}

// RecordCommandExecution records a command execution
func (mc *MetricsCollector) RecordCommandExecution(duration time.Duration, success bool, timeout bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.commandExecutions++
	mc.totalCommandTime += duration

	// Keep only last 1000 durations for percentile calculation
	if len(mc.commandDurations) < 1000 {
		mc.commandDurations = append(mc.commandDurations, duration)
	} else {
		// Replace oldest entry
		mc.commandDurations = append(mc.commandDurations[1:], duration)
	}

	if timeout {
		mc.commandTimeouts++
		mc.commandFailures++
	} else if success {
		mc.commandSuccesses++
	} else {
		mc.commandFailures++
	}
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheHits++
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.cacheMisses++
}

// RecordError records an error occurrence
func (mc *MetricsCollector) RecordError(errorType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.errorCounts[errorType]++
}

// RecordPipelineExecution records a pipeline execution
func (mc *MetricsCollector) RecordPipelineExecution(success bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pipelineExecutions++
	if success {
		mc.pipelineSuccesses++
	} else {
		mc.pipelineFailures++
	}
}

// RecordLoopExecution records a loop execution
func (mc *MetricsCollector) RecordLoopExecution(iterations int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.loopExecutions++
	mc.loopIterations += iterations
}

// UpdateProcessPoolMetrics updates process pool metrics
func (mc *MetricsCollector) UpdateProcessPoolMetrics(size, active, idle int) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.processPoolSize = size
	mc.processPoolActive = active
	mc.processPoolIdle = idle
}

// GetMetrics returns current metrics snapshot
func (mc *MetricsCollector) GetMetrics() *Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	avgDuration := time.Duration(0)
	if mc.commandExecutions > 0 {
		avgDuration = mc.totalCommandTime / time.Duration(mc.commandExecutions)
	}

	// Calculate percentiles
	p50 := mc.calculatePercentile(50)
	p95 := mc.calculatePercentile(95)
	p99 := mc.calculatePercentile(99)

	cacheHitRate := float64(0)
	totalCacheOps := mc.cacheHits + mc.cacheMisses
	if totalCacheOps > 0 {
		cacheHitRate = float64(mc.cacheHits) / float64(totalCacheOps) * 100
	}

	successRate := float64(0)
	if mc.commandExecutions > 0 {
		successRate = float64(mc.commandSuccesses) / float64(mc.commandExecutions) * 100
	}

	avgIterations := float64(0)
	if mc.loopExecutions > 0 {
		avgIterations = float64(mc.loopIterations) / float64(mc.loopExecutions)
	}

	// Copy error counts
	errorCounts := make(map[string]int64)
	for k, v := range mc.errorCounts {
		errorCounts[k] = v
	}

	return &Metrics{
		CommandExecutions:   mc.commandExecutions,
		CommandSuccesses:    mc.commandSuccesses,
		CommandFailures:     mc.commandFailures,
		CommandTimeouts:     mc.commandTimeouts,
		AverageDuration:     avgDuration,
		P50Duration:         p50,
		P95Duration:         p95,
		P99Duration:         p99,
		CacheHits:           mc.cacheHits,
		CacheMisses:         mc.cacheMisses,
		CacheHitRate:        cacheHitRate,
		SuccessRate:         successRate,
		ProcessPoolSize:     mc.processPoolSize,
		ProcessPoolActive:   mc.processPoolActive,
		ProcessPoolIdle:     mc.processPoolIdle,
		ErrorCounts:         errorCounts,
		PipelineExecutions:  mc.pipelineExecutions,
		PipelineSuccesses:   mc.pipelineSuccesses,
		PipelineFailures:    mc.pipelineFailures,
		LoopExecutions:      mc.loopExecutions,
		AverageIterations:   avgIterations,
		Uptime:              time.Since(mc.startTime),
	}
}

// calculatePercentile calculates the nth percentile of command durations
func (mc *MetricsCollector) calculatePercentile(n int) time.Duration {
	if len(mc.commandDurations) == 0 {
		return 0
	}

	// Create a copy and sort
	durations := make([]time.Duration, len(mc.commandDurations))
	copy(durations, mc.commandDurations)

	// Simple insertion sort (for small arrays)
	for i := 1; i < len(durations); i++ {
		key := durations[i]
		j := i - 1
		for j >= 0 && durations[j] > key {
			durations[j+1] = durations[j]
			j--
		}
		durations[j+1] = key
	}

	index := (n * len(durations)) / 100
	if index >= len(durations) {
		index = len(durations) - 1
	}
	return durations[index]
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.commandExecutions = 0
	mc.commandSuccesses = 0
	mc.commandFailures = 0
	mc.commandTimeouts = 0
	mc.commandDurations = make([]time.Duration, 0, 1000)
	mc.totalCommandTime = 0
	mc.cacheHits = 0
	mc.cacheMisses = 0
	mc.errorCounts = make(map[string]int64)
	mc.pipelineExecutions = 0
	mc.pipelineSuccesses = 0
	mc.pipelineFailures = 0
	mc.loopExecutions = 0
	mc.loopIterations = 0
	mc.startTime = time.Now()
}

// Metrics represents a snapshot of collected metrics
type Metrics struct {
	CommandExecutions  int64
	CommandSuccesses   int64
	CommandFailures    int64
	CommandTimeouts    int64
	AverageDuration    time.Duration
	P50Duration        time.Duration
	P95Duration        time.Duration
	P99Duration        time.Duration
	CacheHits          int64
	CacheMisses        int64
	CacheHitRate       float64
	SuccessRate        float64
	ProcessPoolSize    int
	ProcessPoolActive  int
	ProcessPoolIdle    int
	ErrorCounts        map[string]int64
	PipelineExecutions int64
	PipelineSuccesses  int64
	PipelineFailures   int64
	LoopExecutions     int64
	AverageIterations  float64
	Uptime             time.Duration
}

// String returns a formatted string representation of metrics
func (m *Metrics) String() string {
	return m.Format()
}

// Format returns a formatted string representation of metrics
func (m *Metrics) Format() string {
	result := "=== Shode Execution Metrics ===\n\n"
	result += "Command Execution:\n"
	result += fmt.Sprintf("  Total: %d\n", m.CommandExecutions)
	result += fmt.Sprintf("  Successes: %d (%.2f%%)\n", m.CommandSuccesses, m.SuccessRate)
	result += fmt.Sprintf("  Failures: %d\n", m.CommandFailures)
	result += fmt.Sprintf("  Timeouts: %d\n", m.CommandTimeouts)
	result += fmt.Sprintf("  Average Duration: %v\n", m.AverageDuration)
	result += fmt.Sprintf("  P50 Duration: %v\n", m.P50Duration)
	result += fmt.Sprintf("  P95 Duration: %v\n", m.P95Duration)
	result += fmt.Sprintf("  P99 Duration: %v\n", m.P99Duration)
	result += "\n"
	result += "Cache:\n"
	result += fmt.Sprintf("  Hits: %d\n", m.CacheHits)
	result += fmt.Sprintf("  Misses: %d\n", m.CacheMisses)
	result += fmt.Sprintf("  Hit Rate: %.2f%%\n", m.CacheHitRate)
	result += "\n"
	result += "Process Pool:\n"
	result += fmt.Sprintf("  Size: %d\n", m.ProcessPoolSize)
	result += fmt.Sprintf("  Active: %d\n", m.ProcessPoolActive)
	result += fmt.Sprintf("  Idle: %d\n", m.ProcessPoolIdle)
	result += "\n"
	result += "Pipelines:\n"
	result += fmt.Sprintf("  Executions: %d\n", m.PipelineExecutions)
	result += fmt.Sprintf("  Successes: %d\n", m.PipelineSuccesses)
	result += fmt.Sprintf("  Failures: %d\n", m.PipelineFailures)
	result += "\n"
	result += "Loops:\n"
	result += fmt.Sprintf("  Executions: %d\n", m.LoopExecutions)
	result += fmt.Sprintf("  Average Iterations: %.2f\n", m.AverageIterations)
	result += "\n"
	result += fmt.Sprintf("Uptime: %v\n", m.Uptime)
	return result
}
