package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorType represents the type of error
type ErrorType int

const (
	ErrSecurityViolation ErrorType = iota
	ErrCommandNotFound
	ErrExecutionFailed
	ErrParseError
	ErrTimeout
	ErrFileNotFound
	ErrPermissionDenied
	ErrInvalidInput
	ErrResourceExhausted
	ErrNetworkError
	ErrUnknown
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrSecurityViolation:
		return "SecurityViolation"
	case ErrCommandNotFound:
		return "CommandNotFound"
	case ErrExecutionFailed:
		return "ExecutionFailed"
	case ErrParseError:
		return "ParseError"
	case ErrTimeout:
		return "Timeout"
	case ErrFileNotFound:
		return "FileNotFound"
	case ErrPermissionDenied:
		return "PermissionDenied"
	case ErrInvalidInput:
		return "InvalidInput"
	case ErrResourceExhausted:
		return "ResourceExhausted"
	case ErrNetworkError:
		return "NetworkError"
	default:
		return "Unknown"
	}
}

// ExecutionError represents a structured error in Shode with context and stack trace.
//
// ExecutionError provides rich error information including error type, message,
// underlying cause, contextual information, and stack trace for debugging.
//
// Example:
//
//	err := errors.NewExecutionError(errors.ErrSecurityViolation, "dangerous command blocked")
//	err.WithContext("command", "rm").WithContext("args", []string{"-rf", "/"})
type ExecutionError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
	Stack   []string
}

// Error implements the error interface
func (e *ExecutionError) Error() string {
	var parts []string
	
	parts = append(parts, fmt.Sprintf("[%s]", e.Type.String()))
	parts = append(parts, e.Message)
	
	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("(caused by: %v)", e.Cause))
	}
	
	if len(e.Context) > 0 {
		var ctxParts []string
		for k, v := range e.Context {
			ctxParts = append(ctxParts, fmt.Sprintf("%s=%v", k, v))
		}
		parts = append(parts, fmt.Sprintf("[%s]", strings.Join(ctxParts, ", ")))
	}
	
	return strings.Join(parts, " ")
}

// Unwrap returns the underlying error
func (e *ExecutionError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *ExecutionError) WithContext(key string, value interface{}) *ExecutionError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewExecutionError creates a new ExecutionError
func NewExecutionError(errType ErrorType, message string) *ExecutionError {
	return &ExecutionError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
		Stack:   captureStack(2),
	}
}

// WrapError wraps an existing error as an ExecutionError
func WrapError(errType ErrorType, message string, cause error) *ExecutionError {
	return &ExecutionError{
		Type:    errType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
		Stack:   captureStack(2),
	}
}

// captureStack captures the stack trace
func captureStack(skip int) []string {
	var stack []string
	for i := skip; i < skip+10; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, fn.Name()))
		}
	}
	return stack
}

// IsSecurityViolation checks if error is a security violation
func IsSecurityViolation(err error) bool {
	if e, ok := err.(*ExecutionError); ok {
		return e.Type == ErrSecurityViolation
	}
	return false
}

// IsCommandNotFound checks if error is command not found
func IsCommandNotFound(err error) bool {
	if e, ok := err.(*ExecutionError); ok {
		return e.Type == ErrCommandNotFound
	}
	return false
}

// IsTimeout checks if error is a timeout
func IsTimeout(err error) bool {
	if e, ok := err.(*ExecutionError); ok {
		return e.Type == ErrTimeout
	}
	return false
}

// GetErrorType returns the error type
func GetErrorType(err error) ErrorType {
	if e, ok := err.(*ExecutionError); ok {
		return e.Type
	}
	return ErrUnknown
}

// GetErrorContext returns the error context
func GetErrorContext(err error) map[string]interface{} {
	if e, ok := err.(*ExecutionError); ok {
		return e.Context
	}
	return nil
}

// Helper functions for common error types

// NewSecurityViolation creates a security violation error
func NewSecurityViolation(message string) *ExecutionError {
	return NewExecutionError(ErrSecurityViolation, message)
}

// NewCommandNotFound creates a command not found error
func NewCommandNotFound(command string) *ExecutionError {
	return NewExecutionError(ErrCommandNotFound, fmt.Sprintf("command not found: %s", command)).
		WithContext("command", command)
}

// NewExecutionFailed creates an execution failed error
func NewExecutionFailed(message string, cause error) *ExecutionError {
	return WrapError(ErrExecutionFailed, message, cause)
}

// NewParseError creates a parse error
func NewParseError(message string, cause error) *ExecutionError {
	return WrapError(ErrParseError, message, cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) *ExecutionError {
	return NewExecutionError(ErrTimeout, fmt.Sprintf("operation timed out: %s", operation)).
		WithContext("operation", operation)
}

// NewFileNotFoundError creates a file not found error
func NewFileNotFoundError(path string) *ExecutionError {
	return NewExecutionError(ErrFileNotFound, fmt.Sprintf("file not found: %s", path)).
		WithContext("path", path)
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(resource string) *ExecutionError {
	return NewExecutionError(ErrPermissionDenied, fmt.Sprintf("permission denied: %s", resource)).
		WithContext("resource", resource)
}
