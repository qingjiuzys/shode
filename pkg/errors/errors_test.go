package errors

import (
	"errors"
	"testing"
)

func TestNewExecutionError(t *testing.T) {
	err := NewExecutionError(ErrSecurityViolation, "test error")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Type != ErrSecurityViolation {
		t.Errorf("Expected type ErrSecurityViolation, got %v", err.Type)
	}

	if err.Message != "test error" {
		t.Errorf("Expected message 'test error', got '%s'", err.Message)
	}

	if err.Stack == nil || len(err.Stack) == 0 {
		t.Error("Expected stack trace, got empty")
	}
}

func TestErrorString(t *testing.T) {
	err := NewExecutionError(ErrCommandNotFound, "command not found")
	errStr := err.Error()
	
	if !contains(errStr, "CommandNotFound") {
		t.Errorf("Expected error string to contain 'CommandNotFound', got '%s'", errStr)
	}
	
	if !contains(errStr, "command not found") {
		t.Errorf("Expected error string to contain 'command not found', got '%s'", errStr)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrapped := WrapError(ErrExecutionFailed, "execution failed", originalErr)

	if wrapped.Cause != originalErr {
		t.Error("Wrapped error cause mismatch")
	}

	if wrapped.Type != ErrExecutionFailed {
		t.Errorf("Expected type ErrExecutionFailed, got %v", wrapped.Type)
	}
}

func TestWithContext(t *testing.T) {
	err := NewExecutionError(ErrInvalidInput, "invalid input")
	err.WithContext("field", "username").WithContext("value", "test")

	if err.Context["field"] != "username" {
		t.Errorf("Expected context field='username', got '%v'", err.Context["field"])
	}

	if err.Context["value"] != "test" {
		t.Errorf("Expected context value='test', got '%v'", err.Context["value"])
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original")
	wrapped := WrapError(ErrExecutionFailed, "failed", originalErr)

	unwrapped := wrapped.Unwrap()
	if unwrapped != originalErr {
		t.Error("Unwrap returned wrong error")
	}
}

func TestIsSecurityViolation(t *testing.T) {
	err := NewSecurityViolation("security violation")
	if !IsSecurityViolation(err) {
		t.Error("Expected IsSecurityViolation to return true")
	}

	regularErr := errors.New("regular error")
	if IsSecurityViolation(regularErr) {
		t.Error("Expected IsSecurityViolation to return false for regular error")
	}
}

func TestIsCommandNotFound(t *testing.T) {
	err := NewCommandNotFound("nonexistent")
	if !IsCommandNotFound(err) {
		t.Error("Expected IsCommandNotFound to return true")
	}
}

func TestIsTimeout(t *testing.T) {
	err := NewTimeoutError("operation")
	if !IsTimeout(err) {
		t.Error("Expected IsTimeout to return true")
	}
}

func TestGetErrorType(t *testing.T) {
	err := NewExecutionError(ErrParseError, "parse error")
	errType := GetErrorType(err)
	if errType != ErrParseError {
		t.Errorf("Expected ErrParseError, got %v", errType)
	}

	regularErr := errors.New("regular")
	errType = GetErrorType(regularErr)
	if errType != ErrUnknown {
		t.Errorf("Expected ErrUnknown for regular error, got %v", errType)
	}
}

func TestGetErrorContext(t *testing.T) {
	err := NewExecutionError(ErrInvalidInput, "invalid")
	err.WithContext("key", "value")

	ctx := GetErrorContext(err)
	if ctx == nil {
		t.Fatal("Expected context, got nil")
	}

	if ctx["key"] != "value" {
		t.Errorf("Expected context key='value', got '%v'", ctx["key"])
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test NewSecurityViolation
	secErr := NewSecurityViolation("violation")
	if secErr.Type != ErrSecurityViolation {
		t.Error("NewSecurityViolation returned wrong type")
	}

	// Test NewCommandNotFound
	cmdErr := NewCommandNotFound("cmd")
	if cmdErr.Type != ErrCommandNotFound {
		t.Error("NewCommandNotFound returned wrong type")
	}
	if cmdErr.Context["command"] != "cmd" {
		t.Error("NewCommandNotFound missing command context")
	}

	// Test NewExecutionFailed
	origErr := errors.New("original")
	execErr := NewExecutionFailed("failed", origErr)
	if execErr.Type != ErrExecutionFailed {
		t.Error("NewExecutionFailed returned wrong type")
	}
	if execErr.Cause != origErr {
		t.Error("NewExecutionFailed missing cause")
	}

	// Test NewParseError
	parseErr := NewParseError("parse failed", origErr)
	if parseErr.Type != ErrParseError {
		t.Error("NewParseError returned wrong type")
	}

	// Test NewTimeoutError
	timeoutErr := NewTimeoutError("test")
	if timeoutErr.Type != ErrTimeout {
		t.Error("NewTimeoutError returned wrong type")
	}
	if timeoutErr.Context["operation"] != "test" {
		t.Error("NewTimeoutError missing operation context")
	}

	// Test NewFileNotFoundError
	fileErr := NewFileNotFoundError("/path/to/file")
	if fileErr.Type != ErrFileNotFound {
		t.Error("NewFileNotFoundError returned wrong type")
	}
	if fileErr.Context["path"] != "/path/to/file" {
		t.Error("NewFileNotFoundError missing path context")
	}

	// Test NewPermissionDeniedError
	permErr := NewPermissionDeniedError("/restricted")
	if permErr.Type != ErrPermissionDenied {
		t.Error("NewPermissionDeniedError returned wrong type")
	}
	if permErr.Context["resource"] != "/restricted" {
		t.Error("NewPermissionDeniedError missing resource context")
	}
}

func TestErrorTypeString(t *testing.T) {
	testCases := []struct {
		errType ErrorType
		expected string
	}{
		{ErrSecurityViolation, "SecurityViolation"},
		{ErrCommandNotFound, "CommandNotFound"},
		{ErrExecutionFailed, "ExecutionFailed"},
		{ErrParseError, "ParseError"},
		{ErrTimeout, "Timeout"},
		{ErrFileNotFound, "FileNotFound"},
		{ErrPermissionDenied, "PermissionDenied"},
		{ErrInvalidInput, "InvalidInput"},
		{ErrResourceExhausted, "ResourceExhausted"},
		{ErrNetworkError, "NetworkError"},
		{ErrUnknown, "Unknown"},
	}

	for _, tc := range testCases {
		result := tc.errType.String()
		if result != tc.expected {
			t.Errorf("ErrorType.String() for %v: expected '%s', got '%s'", tc.errType, tc.expected, result)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
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
