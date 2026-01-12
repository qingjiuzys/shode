package environment

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvironmentManagerCreation(t *testing.T) {
	em := NewEnvironmentManager()
	if em == nil {
		t.Fatal("EnvironmentManager is nil")
	}

	// Check that original environment is stored
	wd := em.GetWorkingDir()
	if wd == "" {
		t.Error("Working directory should not be empty")
	}
}

func TestGetSetEnv(t *testing.T) {
	em := NewEnvironmentManager()

	// Set environment variable
	em.SetEnv("TEST_VAR", "test_value")

	// Get environment variable
	value := em.GetEnv("TEST_VAR")
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}

	// Get non-existent variable
	value = em.GetEnv("NON_EXISTENT")
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}
}

func TestUnsetEnv(t *testing.T) {
	em := NewEnvironmentManager()

	// Set and then unset
	em.SetEnv("TEST_VAR", "test_value")
	em.UnsetEnv("TEST_VAR")

	value := em.GetEnv("TEST_VAR")
	if value != "" {
		t.Errorf("Expected empty string after unset, got '%s'", value)
	}
}

func TestGetAllEnv(t *testing.T) {
	em := NewEnvironmentManager()

	// Set some variables
	em.SetEnv("VAR1", "value1")
	em.SetEnv("VAR2", "value2")

	allEnv := em.GetAllEnv()
	if len(allEnv) == 0 {
		t.Error("GetAllEnv returned empty map")
	}

	if allEnv["VAR1"] != "value1" {
		t.Errorf("Expected VAR1='value1', got '%s'", allEnv["VAR1"])
	}

	if allEnv["VAR2"] != "value2" {
		t.Errorf("Expected VAR2='value2', got '%s'", allEnv["VAR2"])
	}
}

func TestChangeDir(t *testing.T) {
	em := NewEnvironmentManager()

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	err = em.ChangeDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Verify working directory
	wd := em.GetWorkingDir()
	if wd != tmpDir {
		t.Errorf("Expected working directory '%s', got '%s'", tmpDir, wd)
	}
}

func TestChangeDirRelative(t *testing.T) {
	em := NewEnvironmentManager()

	// Get current working directory
	currentWd := em.GetWorkingDir()

	// Create a subdirectory
	subDir := filepath.Join(currentWd, "test_subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	defer os.RemoveAll(subDir)

	// Change to relative path
	err = em.ChangeDir("test_subdir")
	if err != nil {
		t.Fatalf("Failed to change to relative directory: %v", err)
	}

	// Verify working directory
	wd := em.GetWorkingDir()
	if wd != subDir {
		t.Errorf("Expected working directory '%s', got '%s'", subDir, wd)
	}
}

func TestChangeDirNonExistent(t *testing.T) {
	em := NewEnvironmentManager()

	// Try to change to non-existent directory
	err := em.ChangeDir("/non/existent/directory")
	if err == nil {
		t.Error("Expected error for non-existent directory, got nil")
	}
}

func TestGetPath(t *testing.T) {
	em := NewEnvironmentManager()

	// Set PATH
	em.SetPath("/usr/bin:/usr/local/bin")

	path := em.GetPath()
	if path != "/usr/bin:/usr/local/bin" {
		t.Errorf("Expected PATH='/usr/bin:/usr/local/bin', got '%s'", path)
	}
}

func TestAppendToPath(t *testing.T) {
	em := NewEnvironmentManager()

	// Set initial PATH
	em.SetPath("/usr/bin")

	// Append to PATH
	em.AppendToPath("/usr/local/bin")

	path := em.GetPath()
	if path != "/usr/bin:/usr/local/bin" {
		t.Errorf("Expected PATH='/usr/bin:/usr/local/bin', got '%s'", path)
	}
}

func TestAppendToPathEmpty(t *testing.T) {
	em := NewEnvironmentManager()

	// Clear PATH first
	em.SetPath("")

	// Append to empty PATH
	em.AppendToPath("/usr/bin")

	path := em.GetPath()
	if path != "/usr/bin" {
		t.Errorf("Expected PATH='/usr/bin', got '%s'", path)
	}
}

func TestPrependToPath(t *testing.T) {
	em := NewEnvironmentManager()

	// Set initial PATH
	em.SetPath("/usr/bin")

	// Prepend to PATH
	em.PrependToPath("/usr/local/bin")

	path := em.GetPath()
	if path != "/usr/local/bin:/usr/bin" {
		t.Errorf("Expected PATH='/usr/local/bin:/usr/bin', got '%s'", path)
	}
}

func TestCreateSession(t *testing.T) {
	em := NewEnvironmentManager()

	// Set some environment variables
	em.SetEnv("SESSION_VAR", "session_value")

	// Create session
	session := em.CreateSession()
	if session == nil {
		t.Fatal("Session is nil")
	}

	// Verify session has environment
	value := session.GetEnv("SESSION_VAR")
	if value != "session_value" {
		t.Errorf("Expected 'session_value', got '%s'", value)
	}
}

func TestApplySession(t *testing.T) {
	em := NewEnvironmentManager()

	// Create session and modify it
	session := em.CreateSession()
	session.SetEnv("SESSION_VAR", "modified_value")

	// Apply session
	em.ApplySession(session)

	// Verify environment was updated
	value := em.GetEnv("SESSION_VAR")
	if value != "modified_value" {
		t.Errorf("Expected 'modified_value', got '%s'", value)
	}
}

func TestRestoreOriginalEnvironment(t *testing.T) {
	em := NewEnvironmentManager()

	// Store original PATH
	originalPath := em.GetPath()

	// Modify environment
	em.SetEnv("TEST_VAR", "test")
	em.SetPath("/modified/path")

	// Restore original environment
	em.RestoreOriginalEnvironment()

	// Verify PATH was restored
	path := em.GetPath()
	if path != originalPath {
		t.Errorf("Expected PATH to be restored to '%s', got '%s'", originalPath, path)
	}

	// Verify custom variable was removed
	value := em.GetEnv("TEST_VAR")
	if value != "" {
		t.Errorf("Expected TEST_VAR to be removed, got '%s'", value)
	}
}

func TestConcurrentAccess(t *testing.T) {
	em := NewEnvironmentManager()

	// Test concurrent reads and writes
	done := make(chan bool)
	
	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			em.SetEnv("CONCURRENT_VAR", "value")
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_ = em.GetEnv("CONCURRENT_VAR")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
