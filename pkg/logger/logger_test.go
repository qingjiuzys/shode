package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"
)

// TestNewLogger tests creating a new logger
func TestNewLogger(t *testing.T) {
	config := Config{
		Level:  INFO,
		Format: JSONFormat,
		Output: ConsoleOutput,
	}

	logger := NewLogger(config)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.GetLevel() != INFO {
		t.Errorf("Expected level INFO, got %v", logger.GetLevel())
	}
}

// TestDefaultLogger tests the default logger
func TestDefaultLogger(t *testing.T) {
	if DefaultLogger == nil {
		t.Fatal("DefaultLogger is nil")
	}

	if DefaultLogger.GetLevel() != INFO {
		t.Errorf("Expected default level INFO, got %v", DefaultLogger.GetLevel())
	}
}

// TestLogLevelString tests log level string representation
func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
		}
	}
}

// TestParseLevel tests parsing log level from string
func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"WARN", WARN},
		{"WARNING", WARN},
		{"ERROR", ERROR},
		{"FATAL", FATAL},
		{"invalid", INFO},
	}

	for _, tt := range tests {
		if got := ParseLevel(tt.input); got != tt.expected {
			t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

// TestLogLevels tests different log levels
func TestLogLevels(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  DEBUG,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	// Test all log levels
	logger.Debug("debug message", "key", "value")
	logger.Info("info message", "key", "value")
	logger.Warn("warn message", "key", "value")
	logger.Error("error message", "key", "value")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != 4 {
		t.Errorf("Expected 4 log lines, got %d: %s", len(lines), output)
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i, err)
		}
	}
}

// TestLogLevelFiltering tests log level filtering
func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  WARN,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := strings.TrimSpace(buf.String())
	lines := strings.Split(output, "\n")

	// Should only have WARN and ERROR
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines (WARN, ERROR), got %d", len(lines))
	}
}

// TestSetLevel tests changing log level
func TestSetLevel(t *testing.T) {
	logger := NewLogger(Config{Level: INFO})

	logger.SetLevel(ERROR)
	if logger.GetLevel() != ERROR {
		t.Errorf("Expected level ERROR, got %v", logger.GetLevel())
	}

	logger.SetLevel(DEBUG)
	if logger.GetLevel() != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", logger.GetLevel())
	}
}

// TestJSONFormat tests JSON log format
func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	logger.Info("test message", "key1", "value1", "key2", 123)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}

	if entry.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got '%s'", entry.Level)
	}

	if entry.Fields["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", entry.Fields["key1"])
	}

	if entry.Fields["key2"] != float64(123) { // JSON numbers are float64
		t.Errorf("Expected key2=123, got %v", entry.Fields["key2"])
	}
}

// TestTextFormat tests text log format
func TestTextFormat(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: TextFormat,
		EnableCaller: false, // Disable caller for simpler output
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	logger.Info("test message", "key", "value")

	output := buf.String()

	if !strings.Contains(output, "INFO") {
		t.Errorf("Text format should contain log level, got: %s", output)
	}

	if !strings.Contains(output, "test message") {
		t.Errorf("Text format should contain message, got: %s", output)
	}

	if !strings.Contains(output, "key=value") {
		t.Errorf("Text format should contain fields, got: %s", output)
	}
}

// TestLoggerContext tests logger context
func TestLoggerContext(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	ctx := logger.WithFields(map[string]interface{}{
		"user_id": 123,
		"request_id": "abc",
	})

	ctx.Info("test message")

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Fields["user_id"] != float64(123) {
		t.Errorf("Expected user_id=123, got %v", entry.Fields["user_id"])
	}

	if entry.Fields["request_id"] != "abc" {
		t.Errorf("Expected request_id='abc', got %v", entry.Fields["request_id"])
	}
}

// TestWithTrace tests trace ID
func TestWithTrace(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	traceID := "trace-123"
	ctx := logger.WithTrace(traceID)
	ctx.Info("test message")

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Fields["trace_id"] != traceID {
		t.Errorf("Expected trace_id='%s', got %v", traceID, entry.Fields["trace_id"])
	}
}

// TestWithDuration tests duration tracking
func TestWithDuration(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	ctx := logger.WithFields(map[string]interface{}{}).
		StartTimer()

	time.Sleep(10 * time.Millisecond)

	ctx.WithDuration().Info("operation completed")

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	duration, ok := entry.Fields["duration_ms"]
	if !ok {
		t.Fatal("Duration field not found")
	}

	if d, ok := duration.(float64); !ok || d < 10 {
		t.Errorf("Expected duration >= 10ms, got %v", duration)
	}
}

// TestErrorLogging tests error logging
func TestErrorLogging(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
		EnableStackTrace: true,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	testErr := io.ErrUnexpectedEOF
	logger.Error("operation failed", "error", testErr)

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Error == "" {
		t.Error("Expected error field to be set")
	}

	if entry.Stack == "" {
		t.Error("Expected stack trace for ERROR level")
	}
}

// TestLoggerStats tests logger statistics
func TestLoggerStats(t *testing.T) {
	logger := NewLogger(Config{Level: DEBUG, Format: JSONFormat})
	logger.ResetStats()

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	stats := logger.GetStats()

	if stats.DebugLogs != 1 {
		t.Errorf("Expected 1 debug log, got %d", stats.DebugLogs)
	}

	if stats.InfoLogs != 1 {
		t.Errorf("Expected 1 info log, got %d", stats.InfoLogs)
	}

	if stats.WarnLogs != 1 {
		t.Errorf("Expected 1 warn log, got %d", stats.WarnLogs)
	}

	if stats.ErrorLogs != 1 {
		t.Errorf("Expected 1 error log, got %d", stats.ErrorLogs)
	}

	if stats.TotalLogs != 4 {
		t.Errorf("Expected 4 total logs, got %d", stats.TotalLogs)
	}
}

// TestConcurrentLogging tests concurrent logging
func TestConcurrentLogging(t *testing.T) {
	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = io.Discard
	logger.mu.Unlock()

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(n int) {
			logger.Info("message", "number", n)
			done <- true
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	stats := logger.GetStats()
	if stats.TotalLogs != 100 {
		t.Errorf("Expected 100 total logs, got %d", stats.TotalLogs)
	}
}

// TestLoggerClose tests closing the logger
func TestLoggerClose(t *testing.T) {
	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})

	err := logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestChainedFieldAdding tests chained field adding
func TestChainedFieldAdding(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger(Config{
		Level:  INFO,
		Format: JSONFormat,
	})
	logger.mu.Lock()
	logger.writer = &buf
	logger.mu.Unlock()

	logger.WithFields(map[string]interface{}{}).
		WithField("field1", "value1").
		WithField("field2", 123).
		WithFields(map[string]interface{}{"field3": true}).
		Info("test message")

	var entry LogEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if entry.Fields["field1"] != "value1" {
		t.Errorf("Expected field1='value1', got %v", entry.Fields["field1"])
	}

	if entry.Fields["field2"] != float64(123) {
		t.Errorf("Expected field2=123, got %v", entry.Fields["field2"])
	}

	if entry.Fields["field3"] != true {
		t.Errorf("Expected field3=true, got %v", entry.Fields["field3"])
	}
}
