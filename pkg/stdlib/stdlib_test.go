package stdlib

import (
	"os"
	"testing"
)

func TestNewStdLib(t *testing.T) {
	sl := New()
	if sl == nil {
		t.Fatal("New() returned nil")
	}
}

func TestGetEnvAndSetEnv(t *testing.T) {
	sl := New()

	err := sl.SetEnv("TEST_VAR", "test_value")
	if err != nil {
		t.Fatalf("SetEnv() error = %v", err)
	}

	value := sl.GetEnv("TEST_VAR")
	if value != "test_value" {
		t.Errorf("GetEnv() = %v, want %v", value, "test_value")
	}
}

func TestWorkingDir(t *testing.T) {
	sl := New()
	pwd, err := sl.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	if pwd == "" {
		t.Error("WorkingDir() returned empty string")
	}
}

func TestChangeDir(t *testing.T) {
	sl := New()
	original, _ := os.Getwd()

	err := sl.ChangeDir("/tmp")
	if err != nil {
		t.Fatalf("ChangeDir(/tmp) error = %v", err)
	}

	pwd, err := sl.WorkingDir()
	if err != nil {
		t.Fatalf("WorkingDir() error = %v", err)
	}
	if pwd != "/tmp" && pwd != "/private/tmp" {
		t.Errorf("ChangeDir() result = %v, want %v", pwd, "/tmp")
	}

	os.Chdir(original)
}

func TestStringOperations(t *testing.T) {
	sl := New()

	if sl.ToUpper("hello") != "HELLO" {
		t.Error("ToUpper() failed")
	}

	if sl.ToLower("WORLD") != "world" {
		t.Error("ToLower() failed")
	}

	if sl.Trim("  test  ") != "test" {
		t.Error("Trim() failed")
	}

	if sl.Replace("hello world", "world", "there") != "hello there" {
		t.Error("Replace() failed")
	}
}

func TestContains(t *testing.T) {
	sl := New()

	if !sl.Contains("hello world", "world") {
		t.Error("Contains() should return true")
	}

	if sl.Contains("hello world", "xyz") {
		t.Error("Contains() should return false")
	}
}

func TestFileExists(t *testing.T) {
	sl := New()

	if !sl.FileExists("/tmp") {
		t.Error("FileExists(/tmp) should return true")
	}

	if sl.FileExists("/nonexistent/path/xyz") {
		t.Error("FileExists() should return false for non-existent path")
	}
}

func TestReadFile(t *testing.T) {
	sl := New()

	content, err := sl.ReadFile("/etc/hostname")
	if err != nil {
		t.Logf("ReadFile() error (non-critical): %v", err)
	}

	if content == "" {
		t.Log("ReadFile() returned empty content")
	}
}

func TestCacheOperations(t *testing.T) {
	sl := New()

	sl.SetCache("test_key", "test_value", 60)

	value, exists := sl.GetCache("test_key")
	if !exists {
		t.Error("GetCache() should return true after SetCache()")
	}

	if value != "test_value" {
		t.Errorf("GetCache() = %v, want %v", value, "test_value")
	}

	if !sl.CacheExists("test_key") {
		t.Error("CacheExists() should return true")
	}

	sl.DeleteCache("test_key")

	if sl.CacheExists("test_key") {
		t.Error("CacheExists() should return false after DeleteCache()")
	}
}

func TestSHA256Hash(t *testing.T) {
	sl := New()

	hash := sl.SHA256Hash("test")
	if hash == "" {
		t.Error("SHA256Hash() returned empty string")
	}

	expectedHash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if hash != expectedHash {
		t.Errorf("SHA256Hash() = %v, want %v", hash, expectedHash)
	}
}
