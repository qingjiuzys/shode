package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageConfig(t *testing.T) {
	pm := NewPackageManager()
	
	// Test initialization
	err := pm.Init("test-package", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to initialize package: %v", err)
	}
	
	config := pm.GetConfig()
	if config.Name != "test-package" {
		t.Errorf("Expected name 'test-package', got '%s'", config.Name)
	}
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}
}

func TestAddDependency(t *testing.T) {
	pm := NewPackageManager()
	
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "shode-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Change working directory
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	// Initialize package
	err = pm.Init("test-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	// Add dependency
	err = pm.AddDependency("lodash", "4.17.21", false)
	if err != nil {
		t.Fatalf("Failed to add dependency: %v", err)
	}
	
	// Reload config to verify
	err = pm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	
	config := pm.GetConfig()
	if config.Dependencies["lodash"] != "4.17.21" {
		t.Errorf("Dependency not found or wrong version")
	}
}

func TestCreateTarball(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "shode-tarball-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	// Create package.json
	packageJson := filepath.Join(tmpDir, "shode.json")
	pkgData := PackageConfig{
		Name:    "test-package",
		Version: "1.0.0",
	}
	jsonData, _ := json.Marshal(pkgData)
	err = os.WriteFile(packageJson, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}
	
	// Create tarball
	tarballData, err := createTarball(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create tarball: %v", err)
	}
	
	if len(tarballData) == 0 {
		t.Error("Tarball data is empty")
	}
	
	// Verify checksum
	checksum := calculateChecksum(tarballData)
	if checksum == "" {
		t.Error("Checksum is empty")
	}
}

func TestPackageManagerScripts(t *testing.T) {
	pm := NewPackageManager()
	
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-script-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	// Initialize
	err = pm.Init("test-pkg", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to initialize: %v", err)
	}
	
	// Add script
	err = pm.AddScript("test", "echo 'test'")
	if err != nil {
		t.Fatalf("Failed to add script: %v", err)
	}
	
	// Reload and verify
	err = pm.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to reload: %v", err)
	}
	
	config := pm.GetConfig()
	if config.Scripts["test"] != "echo 'test'" {
		t.Error("Script not found or incorrect")
	}
}
