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

func TestIsExportedFunction(t *testing.T) {
	mm := NewModuleManager()

	// Create a temporary module with exports
	tmpDir, err := os.MkdirTemp("", "shode-module-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJson := filepath.Join(tmpDir, "package.json")
	pkgData := PackageJson{
		Name: "test-utils",
		Main: "main.sh",
	}
	jsonData, _ := json.Marshal(pkgData)
	os.WriteFile(packageJson, jsonData, 0644)

	// Create main.sh with exports
	mainFile := filepath.Join(tmpDir, "main.sh")
	scriptContent := `
export_hello() { echo 'Hello'; }
export_goodbye() { echo 'Goodbye'; }
non_export_function() { echo 'Not exported'; }
`
	os.WriteFile(mainFile, []byte(scriptContent), 0644)

	// Load the module
	module, err := mm.LoadModule(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify exports are loaded
	if len(module.Exports) != 2 {
		t.Errorf("Expected 2 exports, got %d", len(module.Exports))
	}

	// Test IsExportedFunction with exported functions
	if !mm.IsExportedFunction("hello") {
		t.Error("Expected 'hello' to be recognized as exported function")
	}

	if !mm.IsExportedFunction("goodbye") {
		t.Error("Expected 'goodbye' to be recognized as exported function")
	}

	// Test with non-exported function
	if mm.IsExportedFunction("non_export_function") {
		t.Error("Expected 'non_export_function' to NOT be recognized as exported")
	}

	// Test with non-existent function
	if mm.IsExportedFunction("does_not_exist") {
		t.Error("Expected 'does_not_exist' to NOT be recognized as exported")
	}
}

func TestHybridExecutionMode(t *testing.T) {
	// This test verifies that the hybrid execution mode can distinguish between
	// stdlib functions, user-defined functions, module exports, and external commands

	// Note: This is a basic smoke test. Full integration tests would require
	// setting up an execution engine with all dependencies.

	mm := NewModuleManager()

	// Create a module
	tmpDir, err := os.MkdirTemp("", "shode-module-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create package.json
	packageJson := filepath.Join(tmpDir, "package.json")
	pkgData := PackageJson{
		Name: "math-utils",
		Main: "index.sh",
	}
	jsonData, _ := json.Marshal(pkgData)
	os.WriteFile(packageJson, jsonData, 0644)

	// Create index.sh with a math export
	indexFile := filepath.Join(tmpDir, "index.sh")
	os.WriteFile(indexFile, []byte("export_add() { echo $1 + $2; }"), 0644)

	// Load module
	_, err = mm.LoadModule(tmpDir)
	if err != nil {
		t.Fatalf("Failed to load module: %v", err)
	}

	// Verify the module export is recognized
	if !mm.IsExportedFunction("add") {
		t.Error("Expected 'add' to be recognized as exported function from math-utils module")
	}

	// Verify non-existent functions are not recognized
	if mm.IsExportedFunction("subtract") {
		t.Error("Expected 'subtract' to NOT be recognized (not exported)")
	}
}
