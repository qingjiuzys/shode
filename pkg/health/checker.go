package health

import (
	"encoding/json"
	"fmt"
	"sync"
)

// HealthStatus represents health status
type HealthStatus string

const (
	StatusUp   HealthStatus = "UP"
	StatusDown HealthStatus = "DOWN"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Status  HealthStatus            `json:"status"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthIndicator checks the health of a component
type HealthIndicator interface {
	Health() HealthCheck
}

// HealthChecker manages health checks
type HealthChecker struct {
	indicators map[string]HealthIndicator
	mu         sync.RWMutex
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		indicators: make(map[string]HealthIndicator),
	}
}

// RegisterIndicator registers a health indicator
func (hc *HealthChecker) RegisterIndicator(name string, indicator HealthIndicator) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.indicators[name] = indicator
}

// Check performs all health checks
func (hc *HealthChecker) Check() map[string]HealthCheck {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]HealthCheck)
	for name, indicator := range hc.indicators {
		results[name] = indicator.Health()
	}

	return results
}

// CheckOverall performs overall health check
func (hc *HealthChecker) CheckOverall() HealthCheck {
	results := hc.Check()
	
	overallStatus := StatusUp
	details := make(map[string]interface{})
	
	for name, check := range results {
		details[name] = check
		if check.Status == StatusDown {
			overallStatus = StatusDown
		}
	}

	return HealthCheck{
		Status:  overallStatus,
		Details: details,
	}
}

// ToJSON converts health check to JSON
func (hc HealthCheck) ToJSON() (string, error) {
	data, err := json.Marshal(hc)
	if err != nil {
		return "", fmt.Errorf("failed to marshal health check: %v", err)
	}
	return string(data), nil
}
