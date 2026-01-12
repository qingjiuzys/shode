package engine

import (
	"gitee.com/com_818cloud/shode/pkg/metrics"
)

// GetMetrics returns a snapshot of the current execution metrics.
//
// The metrics include command execution statistics, cache performance,
// process pool usage, error counts, and performance percentiles.
//
// Returns:
//   - *metrics.Metrics: Current metrics snapshot, or nil if metrics collection is disabled
//
// Example:
//
//	m := ee.GetMetrics()
//	if m != nil {
//	    fmt.Printf("Commands executed: %d\n", m.CommandExecutions)
//	    fmt.Printf("Success rate: %.2f%%\n", m.SuccessRate)
//	    fmt.Printf("Cache hit rate: %.2f%%\n", m.CacheHitRate)
//	}
func (ee *ExecutionEngine) GetMetrics() *metrics.Metrics {
	if ee.metrics == nil {
		return nil
	}
	return ee.metrics.GetMetrics()
}

// ResetMetrics resets all collected metrics to their initial state.
//
// This is useful for starting a new metrics collection period or
// clearing metrics after a test run.
//
// Example:
//
//	ee.ResetMetrics()
//	// Execute commands...
//	metrics := ee.GetMetrics()
func (ee *ExecutionEngine) ResetMetrics() {
	if ee.metrics != nil {
		ee.metrics.Reset()
	}
}
