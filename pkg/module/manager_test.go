package module

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPackageJson(t *testing.T) {
	mm := NewModuleManager()
	
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-module-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create package.json
	packageJson := filepath.Join(tmpDir, "package.json")
	pkgData := PackageJson{
		Name:        "test-module",
		Version:     "1.0.0",
		Description: "Test module",
		Main:        "index.sh",
	}
	
	jsonData, err := json.Marshal(pkgData)
	if err != nil {
		t.Fatalf("Failed to marshal package.json: %v", err)
	}
	
	err = os.WriteFile(packageJson, jsonData, 0644)
	if err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}
	
	// Load package.json
	pkg, err := mm.loadPackageJson(packageJson)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}
	
	if pkg.Name != "test-module" {
		t.Errorf("Expected name 'test-module', got '%s'", pkg.Name)
	}
	
	if pkg.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", pkg.Version)
	}
	
	if pkg.Main != "index.sh" {
		t.Errorf("Expected main 'index.sh', got '%s'", pkg.Main)
	}
}

func TestLoadPackageJsonDefaultMain(t *testing.T) {
	mm := NewModuleManager()
	
	tmpDir, err := os.MkdirTemp("", "shode-module-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create package.json without main field
	packageJson := filepath.Join(tmpDir, "package.json")
	pkgData := PackageJson{
		Name:    "test-module",
		Version: "1.0.0",
	}
	
	jsonData, _ := json.Marshal(pkgData)
	os.WriteFile(packageJson, jsonData, 0644)
	
	// Load package.json
	pkg, err := mm.loadPackageJson(packageJson)
	if err != nil {
		t.Fatalf("Failed to load package.json: %v", err)
	}
	
	// Should default to index.sh
	if pkg.Main != "index.sh" {
		t.Errorf("Expected default main 'index.sh', got '%s'", pkg.Main)
	}
}

func TestModuleLoadWithPackageJson(t *testing.T) {
	mm := NewModuleManager()
	
	tmpDir, err := os.MkdirTemp("", "shode-module-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create package.json
	packageJson := filepath.Join(tmpDir, "package.json")
	pkgData := PackageJson{
		Name: "test-module",
		Main:  "main.sh",
	}
	jsonData, _ := json.Marshal(pkgData)
	os.WriteFile(packageJson, jsonData, 0644)
	
	// Create main.sh file
	mainFile := filepath.Join(tmpDir, "main.sh")
	os.WriteFile(mainFile, []byte("export_hello() { echo 'Hello'; }"), 0644)
	
	// Load module
	module, err := mm.LoadModule(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}
	
	if module == nil {
		t.Fatal("Module is nil")
	}
	
	if module.Name != "test-module" {
		t.Errorf("Expected module name 'test-module', got '%s'", module.Name)
	}
}
