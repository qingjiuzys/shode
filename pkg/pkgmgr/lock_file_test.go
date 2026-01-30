package pkg

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/semver"
)

func TestNewLockFileManager(t *testing.T) {
	lfm := NewLockFileManager("/path/to/shode.json")
	if lfm == nil {
		t.Fatal("NewLockFileManager returned nil")
	}

	if lfm.configPath != "/path/to/shode.json" {
		t.Errorf("Expected configPath '/path/to/shode.json', got '%s'", lfm.configPath)
	}
}

func TestLockFile_Generate(t *testing.T) {
	lfm := NewLockFileManager("")

	// Create resolved dependencies
	resolved := []*ResolvedDependency{
		{
			Name:    "@shode/logger",
			Version: semver.MustParseVersion("1.2.3"),
			Dependencies: []*ResolvedDependency{
				{
					Name:    "@shode/config",
					Version: semver.MustParseVersion("1.0.0"),
				},
			},
		},
		{
			Name:    "@shode/http",
			Version: semver.MustParseVersion("2.0.0"),
		},
	}

	lockfile, err := lfm.Generate(resolved)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify lock file structure
	if lockfile.LockfileVersion != 1 {
		t.Errorf("Expected lockfileVersion 1, got %d", lockfile.LockfileVersion)
	}

	if len(lockfile.Resolved) != 3 {
		t.Errorf("Expected 3 resolved packages, got %d", len(lockfile.Resolved))
	}

	// Check specific package
	logger, exists := lockfile.Resolved["@shode/logger"]
	if !exists {
		t.Fatal("@shode/logger not found in resolved")
	}

	if logger.Version != "1.2.3" {
		t.Errorf("Expected version 1.2.3, got %s", logger.Version)
	}

	// Check dependencies
	if len(logger.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(logger.Dependencies))
	}

	if logger.Dependencies["@shode/config"] != "1.0.0" {
		t.Errorf("Expected @shode/config@1.0.0, got %s", logger.Dependencies["@shode/config"])
	}
}

func TestLockFile_SaveAndLoad(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create lock file
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved: map[string]*LockEntry{
			"test-pkg": {
				Version:   "1.0.0",
				Integrity: "sha512-abc123",
				Resolved:  "https://registry.shode.io/test-pkg/-/test-pkg-1.0.0.tgz",
			},
		},
	}

	// Save
	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		t.Error("Lock file was not created")
	}

	// Load
	loaded, err := lfm.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify loaded content
	if loaded.LockfileVersion != lockfile.LockfileVersion {
		t.Errorf("Expected lockfileVersion %d, got %d", lockfile.LockfileVersion, loaded.LockfileVersion)
	}

	if len(loaded.Resolved) != len(lockfile.Resolved) {
		t.Errorf("Expected %d resolved packages, got %d", len(lockfile.Resolved), len(loaded.Resolved))
	}

	entry := loaded.Resolved["test-pkg"]
	if entry.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", entry.Version)
	}
}

func TestLockFile_Load_NotFound(t *testing.T) {
	lfm := NewLockFileManager("/nonexistent/shode.json")

	_, err := lfm.Load()
	if err == nil {
		t.Error("Expected error for missing lock file")
	}

	if err.Error() != "lock file not found: /nonexistent/shode-lock.json" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestLockFile_Validate(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create lock file
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved: map[string]*LockEntry{
			"test-pkg": {
				Version: "1.2.3",
			},
		},
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Validate with matching config
	config := &PackageConfig{
		Dependencies: map[string]string{
			"test-pkg": "^1.0.0",
		},
	}

	if err := lfm.Validate(config); err != nil {
		t.Errorf("Validate failed: %v", err)
	}

	// Validate with non-matching config
	config2 := &PackageConfig{
		Dependencies: map[string]string{
			"test-pkg": "^2.0.0",
		},
	}

	if err := lfm.Validate(config2); err == nil {
		t.Error("Expected error for non-matching constraint")
	}
}

func TestLockFile_Update(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create initial lock file
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved: map[string]*LockEntry{
			"test-pkg": {
				Version: "1.0.0",
			},
		},
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Update package version
	if err := lfm.Update("test-pkg", "1.2.0"); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Load and verify
	loaded, err := lfm.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Resolved["test-pkg"].Version != "1.2.0" {
		t.Errorf("Expected version 1.2.0, got %s", loaded.Resolved["test-pkg"].Version)
	}
}

func TestLockFile_Verify(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create lock file
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved: map[string]*LockEntry{
			"test-pkg": {
				Version:   "1.0.0",
				Integrity: "sha512-abc123def456",
			},
		},
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify
	if err := lfm.Verify(); err != nil {
		t.Errorf("Verify failed: %v", err)
	}
}

func TestLockFile_Verify_InvalidVersion(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create lock file with invalid version
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved: map[string]*LockEntry{
			"test-pkg": {
				Version: "invalid",
			},
		},
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify should fail
	if err := lfm.Verify(); err == nil {
		t.Error("Expected error for invalid version")
	}
}

func TestLockFile_Exists(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Should not exist initially
	if lfm.Exists() {
		t.Error("Exists() returned true for non-existent file")
	}

	// Create lock file
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved:        make(map[string]*LockEntry),
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Should exist now
	if !lfm.Exists() {
		t.Error("Exists() returned false for existing file")
	}
}

func TestLockFile_UnsupportedVersion(t *testing.T) {
	// Create temp directory
	tmpDir, err := ioutil.TempDir("", "shode-lock-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	lockfilePath := filepath.Join(tmpDir, "shode-lock.json")
	lfm := NewLockFileManager(lockfilePath)

	// Create lock file with unsupported version
	lockfile := &LockFile{
		LockfileVersion: 999,
		GeneratedAt:     time.Now(),
		Resolved:        make(map[string]*LockEntry),
	}

	if err := lfm.Save(lockfile); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load should fail
	_, err = lfm.Load()
	if err == nil {
		t.Error("Expected error for unsupported lock file version")
	}

	if err.Error() != "unsupported lock file version: 999" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestCalculateIntegrity(t *testing.T) {
	name := "test-pkg"
	version := semver.MustParseVersion("1.2.3")

	integrity := calculateIntegrity(name, version)
	if integrity == "" {
		t.Error("calculateIntegrity returned empty string")
	}

	// Integrity should contain name and version
	expected := "sha512-test-pkg-1.2.3"
	if integrity != expected {
		t.Errorf("Expected integrity '%s', got '%s'", expected, integrity)
	}
}
