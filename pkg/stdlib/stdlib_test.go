package stdlib

import (
	"os"
	"testing"
	"time"
)

func TestStdLib_ExecuteFunction(t *testing.T) {
	sl := New()

	tests := []struct {
		name     string
		function string
		args     []interface{}
		wantErr  bool
	}{
		{
			name:     "upper function",
			function: "upper",
			args:     []interface{}{"hello"},
			wantErr:  false,
		},
		{
			name:     "contains function",
			function: "contains",
			args:     []interface{}{"hello world", "world"},
			wantErr:  false,
		},
		{
			name:     "trim function",
			function: "trim",
			args:     []interface{}{"   hello   "},
			wantErr:  false,
		},
		{
			name:     "nonexistent function",
			function: "nonexistent",
			args:     []interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sl.ExecuteFunction(tt.function, tt.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStdLib_FileOperations(t *testing.T) {
	sl := New()

	// Test file creation and reading
	testContent := "Hello, Shode Test!"
	testFile := "test_file.txt"

	// Clean up before test
	os.Remove(testFile)

	t.Run("write and read file", func(t *testing.T) {
		// Write file
		_, err := sl.ExecuteFunction("write", testFile, testContent)
		if err != nil {
			t.Fatalf("Write file failed: %v", err)
		}

		// Read file
		result, err := sl.ExecuteFunction("readfile", testFile)
		if err != nil {
			t.Fatalf("Read file failed: %v", err)
		}

		if result != testContent {
			t.Errorf("Read content = %v, want %v", result, testContent)
		}

		// Check file exists
		exists, err := sl.ExecuteFunction("exists", testFile)
		if err != nil {
			t.Fatalf("Exists check failed: %v", err)
		}

		// The exists function returns a boolean, but ExecuteFunction converts it to string
		if exists != true && exists != "true" {
			t.Errorf("File should exist, got: %v", exists)
		}
	})

	// Clean up
	os.Remove(testFile)
}

func TestStdLib_StringOperations(t *testing.T) {
	sl := New()

	tests := []struct {
		name     string
		function string
		args     []interface{}
		want     interface{}
	}{
		{
			name:     "upper case",
			function: "upper",
			args:     []interface{}{"hello"},
			want:     "HELLO",
		},
		{
			name:     "lower case",
			function: "lower",
			args:     []interface{}{"HELLO"},
			want:     "hello",
		},
		{
			name:     "trim spaces",
			function: "trim",
			args:     []interface{}{"   hello   "},
			want:     "hello",
		},
		{
			name:     "contains true",
			function: "contains",
			args:     []interface{}{"hello world", "world"},
			want:     true,
		},
		{
			name:     "contains false",
			function: "contains",
			args:     []interface{}{"hello world", "missing"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sl.ExecuteFunction(tt.function, tt.args...)
			if err != nil {
				t.Fatalf("ExecuteFunction failed: %v", err)
			}

			if result != tt.want {
				t.Errorf("ExecuteFunction() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestStdLib_HashOperations(t *testing.T) {
	sl := New()

	testString := "hello world"

	tests := []struct {
		name     string
		function string
		wantLen  int // Expected hash length
	}{
		{
			name:     "md5 hash",
			function: "md5",
			wantLen:  32, // MD5 produces 32-character hex string
		},
		{
			name:     "sha1 hash",
			function: "sha1",
			wantLen:  40, // SHA1 produces 40-character hex string
		},
		{
			name:     "sha256 hash",
			function: "sha256",
			wantLen:  64, // SHA256 produces 64-character hex string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sl.ExecuteFunction(tt.function, testString)
			if err != nil {
				t.Fatalf("%s failed: %v", tt.function, err)
			}

			strResult, ok := result.(string)
			if !ok {
				t.Fatalf("Expected string result, got %T", result)
			}

			if len(strResult) != tt.wantLen {
				t.Errorf("%s length = %d, want %d", tt.function, len(strResult), tt.wantLen)
			}
		})
	}
}

func TestStdLib_SystemInfo(t *testing.T) {
	sl := New()

	t.Run("hostname", func(t *testing.T) {
		result, err := sl.ExecuteFunction("hostname")
		if err != nil {
			t.Fatalf("hostname failed: %v", err)
		}

		if result == "" {
			t.Error("hostname should not be empty")
		}
	})

	t.Run("username", func(t *testing.T) {
		result, err := sl.ExecuteFunction("whoami")
		if err != nil {
			t.Fatalf("whoami failed: %v", err)
		}

		if result == "" {
			t.Error("username should not be empty")
		}
	})

	t.Run("process id", func(t *testing.T) {
		result, err := sl.ExecuteFunction("pid")
		if err != nil {
			t.Fatalf("pid failed: %v", err)
		}

		// PID should be a positive integer
		if result.(int) <= 0 {
			t.Errorf("pid should be positive, got %v", result)
		}
	})
}

func TestStdLib_ListFunctions(t *testing.T) {
	sl := New()

	t.Run("list functions", func(t *testing.T) {
		functions := sl.ListFunctions()
		
		// Should have at least the basic functions
		if len(functions) < 20 {
			t.Errorf("Expected at least 20 functions, got %d", len(functions))
		}

		// Check for some essential functions
		essentialFuncs := []string{"upper", "lower", "trim", "contains", "readfile", "write"}
		for _, funcName := range essentialFuncs {
			found := false
			for _, availableFunc := range functions {
				if availableFunc == funcName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Essential function %s not found", funcName)
			}
		}
	})
}

func TestStdLib_Performance(t *testing.T) {
	sl := New()

	// Test performance of string operations
	start := time.Now()
	
	for i := 0; i < 1000; i++ {
		_, err := sl.ExecuteFunction("upper", "test string")
		if err != nil {
			t.Fatalf("Performance test failed: %v", err)
		}
	}
	
	duration := time.Since(start)
	t.Logf("1000 upper operations took: %v", duration)

	// Should be very fast (less than 100ms for 1000 operations)
	if duration > 100*time.Millisecond {
		t.Errorf("Performance test took too long: %v", duration)
	}
}

func TestStdLib_ErrorHandling(t *testing.T) {
	sl := New()

	tests := []struct {
		name     string
		function string
		args     []interface{}
	}{
		{
			name:     "readfile nonexistent",
			function: "readfile",
			args:     []interface{}{"nonexistent_file.txt"},
		},
		{
			name:     "writefile invalid args",
			function: "writefile",
			args:     []interface{}{}, // Missing arguments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sl.ExecuteFunction(tt.function, tt.args...)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

func TestStdLib_ExecuteFunctionSafe(t *testing.T) {
	sl := New()

	t.Run("safe execution with panic", func(t *testing.T) {
		// This should not panic even if function doesn't exist
		result, err := sl.ExecuteFunctionSafe("nonexistent_function")
		if err == nil {
			t.Error("Expected error for nonexistent function")
		}
		if result != nil {
			t.Error("Expected nil result for nonexistent function")
		}
	})

	t.Run("safe execution valid function", func(t *testing.T) {
		result, err := sl.ExecuteFunctionSafe("upper", "hello")
		if err != nil {
			t.Errorf("Valid function should not error: %v", err)
		}
		if result != "HELLO" {
			t.Errorf("Expected HELLO, got %v", result)
		}
	})
}
