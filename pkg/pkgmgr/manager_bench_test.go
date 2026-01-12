package pkg

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/environment"
)

func setupBenchmarkPackageManager() *PackageManager {
	em := environment.NewEnvironmentManager()
	pm := NewPackageManager()
	pm.envManager = em
	return pm
}

// BenchmarkInit benchmarks package initialization
func BenchmarkInit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmpDir, _ := os.MkdirTemp("", "shode-bench-*")
		defer os.RemoveAll(tmpDir)
		
		oldWd, _ := os.Getwd()
		os.Chdir(tmpDir)
		defer os.Chdir(oldWd)
		
		pm := setupBenchmarkPackageManager()
		_ = pm.Init("test-package", "1.0.0")
	}
}

// BenchmarkLoadConfig benchmarks config loading
func BenchmarkLoadConfig(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "shode-bench-*")
	defer os.RemoveAll(tmpDir)
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	pm := setupBenchmarkPackageManager()
	pm.Init("test-package", "1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.LoadConfig()
	}
}

// BenchmarkSaveConfig benchmarks config saving
func BenchmarkSaveConfig(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "shode-bench-*")
	defer os.RemoveAll(tmpDir)
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	pm := setupBenchmarkPackageManager()
	pm.Init("test-package", "1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.SaveConfig()
	}
}

// BenchmarkAddDependency benchmarks adding dependencies
func BenchmarkAddDependency(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "shode-bench-*")
	defer os.RemoveAll(tmpDir)
	
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)
	
	pm := setupBenchmarkPackageManager()
	pm.Init("test-package", "1.0.0")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pm.AddDependency("test-dep", "1.0.0", false)
	}
}

// BenchmarkCreateTarball benchmarks tarball creation
func BenchmarkCreateTarball(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "shode-bench-*")
	defer os.RemoveAll(tmpDir)
	
	// Create test files
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)
	
	packageJson := filepath.Join(tmpDir, "shode.json")
	pkgData := PackageConfig{
		Name:    "test-package",
		Version: "1.0.0",
	}
	jsonData, _ := json.Marshal(pkgData)
	os.WriteFile(packageJson, jsonData, 0644)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = createTarball(tmpDir)
	}
}
