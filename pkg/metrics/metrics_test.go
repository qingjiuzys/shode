package metrics

import (
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()

	// Record some command executions
	mc.RecordCommandExecution(100*time.Millisecond, true, false)
	mc.RecordCommandExecution(200*time.Millisecond, true, false)
	mc.RecordCommandExecution(150*time.Millisecond, false, false)
	mc.RecordCommandExecution(50*time.Millisecond, true, true)

	// Record cache operations
	mc.RecordCacheHit()
	mc.RecordCacheHit()
	mc.RecordCacheMiss()

	// Record errors
	mc.RecordError("Timeout")
	mc.RecordError("Timeout")
	mc.RecordError("SecurityViolation")

	// Record pipeline
	mc.RecordPipelineExecution(true)
	mc.RecordPipelineExecution(false)

	// Record loop
	mc.RecordLoopExecution(10)
	mc.RecordLoopExecution(5)

	// Update process pool
	mc.UpdateProcessPoolMetrics(10, 3, 7)

	// Get metrics
	metrics := mc.GetMetrics()

	// Verify command metrics
	if metrics.CommandExecutions != 4 {
		t.Errorf("Expected 4 command executions, got %d", metrics.CommandExecutions)
	}
	if metrics.CommandSuccesses != 2 {
		t.Errorf("Expected 2 successes, got %d", metrics.CommandSuccesses)
	}
	if metrics.CommandFailures != 2 {
		t.Errorf("Expected 2 failures, got %d", metrics.CommandFailures)
	}
	if metrics.CommandTimeouts != 1 {
		t.Errorf("Expected 1 timeout, got %d", metrics.CommandTimeouts)
	}

	// Verify cache metrics
	if metrics.CacheHits != 2 {
		t.Errorf("Expected 2 cache hits, got %d", metrics.CacheHits)
	}
	if metrics.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", metrics.CacheMisses)
	}
	if metrics.CacheHitRate < 66.0 || metrics.CacheHitRate > 67.0 {
		t.Errorf("Expected cache hit rate ~66.67%%, got %.2f%%", metrics.CacheHitRate)
	}

	// Verify error counts
	if metrics.ErrorCounts["Timeout"] != 2 {
		t.Errorf("Expected 2 timeout errors, got %d", metrics.ErrorCounts["Timeout"])
	}
	if metrics.ErrorCounts["SecurityViolation"] != 1 {
		t.Errorf("Expected 1 security violation, got %d", metrics.ErrorCounts["SecurityViolation"])
	}

	// Verify pipeline metrics
	if metrics.PipelineExecutions != 2 {
		t.Errorf("Expected 2 pipeline executions, got %d", metrics.PipelineExecutions)
	}
	if metrics.PipelineSuccesses != 1 {
		t.Errorf("Expected 1 pipeline success, got %d", metrics.PipelineSuccesses)
	}
	if metrics.PipelineFailures != 1 {
		t.Errorf("Expected 1 pipeline failure, got %d", metrics.PipelineFailures)
	}

	// Verify loop metrics
	if metrics.LoopExecutions != 2 {
		t.Errorf("Expected 2 loop executions, got %d", metrics.LoopExecutions)
	}
	if metrics.AverageIterations != 7.5 {
		t.Errorf("Expected average iterations 7.5, got %.2f", metrics.AverageIterations)
	}

	// Verify process pool metrics
	if metrics.ProcessPoolSize != 10 {
		t.Errorf("Expected process pool size 10, got %d", metrics.ProcessPoolSize)
	}
	if metrics.ProcessPoolActive != 3 {
		t.Errorf("Expected process pool active 3, got %d", metrics.ProcessPoolActive)
	}
	if metrics.ProcessPoolIdle != 7 {
		t.Errorf("Expected process pool idle 7, got %d", metrics.ProcessPoolIdle)
	}
}

func TestMetricsReset(t *testing.T) {
	mc := NewMetricsCollector()

	// Record some metrics
	mc.RecordCommandExecution(100*time.Millisecond, true, false)
	mc.RecordCacheHit()

	// Reset
	mc.Reset()

	// Verify reset
	metrics := mc.GetMetrics()
	if metrics.CommandExecutions != 0 {
		t.Errorf("Expected 0 command executions after reset, got %d", metrics.CommandExecutions)
	}
	if metrics.CacheHits != 0 {
		t.Errorf("Expected 0 cache hits after reset, got %d", metrics.CacheHits)
	}
}

func TestMetricsPercentiles(t *testing.T) {
	mc := NewMetricsCollector()

	// Record durations in ascending order
	for i := 1; i <= 100; i++ {
		mc.RecordCommandExecution(time.Duration(i)*time.Millisecond, true, false)
	}

	metrics := mc.GetMetrics()

	// P50 should be around 50ms
	if metrics.P50Duration < 45*time.Millisecond || metrics.P50Duration > 55*time.Millisecond {
		t.Errorf("Expected P50 around 50ms, got %v", metrics.P50Duration)
	}

	// P95 should be around 95ms
	if metrics.P95Duration < 90*time.Millisecond || metrics.P95Duration > 100*time.Millisecond {
		t.Errorf("Expected P95 around 95ms, got %v", metrics.P95Duration)
	}

	// P99 should be around 99ms
	if metrics.P99Duration < 95*time.Millisecond || metrics.P99Duration > 100*time.Millisecond {
		t.Errorf("Expected P99 around 99ms, got %v", metrics.P99Duration)
	}
}

func TestMetricsFormat(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordCommandExecution(100*time.Millisecond, true, false)
	mc.RecordCacheHit()

	metrics := mc.GetMetrics()
	formatted := metrics.Format()

	if formatted == "" {
		t.Error("Expected formatted metrics string, got empty")
	}

	// Verify it contains expected sections
	if !contains(formatted, "Command Execution") {
		t.Error("Formatted string should contain 'Command Execution'")
	}
	if !contains(formatted, "Cache") {
		t.Error("Formatted string should contain 'Cache'")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
