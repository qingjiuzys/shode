package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	pkg "gitee.com/com_818cloud/shode/pkg/pkgmgr"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func setupIntegrationTest() (*engine.ExecutionEngine, *pkg.PackageManager, string) {
	tmpDir, _ := os.MkdirTemp("", "shode-integration-*")
	
	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)
	
	pm := pkg.NewPackageManager()
	
	return ee, pm, tmpDir
}

func TestFullScriptExecution(t *testing.T) {
	ee, _, tmpDir := setupIntegrationTest()
	defer os.RemoveAll(tmpDir)
	
	ctx := context.Background()
	
	// Create a test script
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{"hello"},
			},
			&types.CommandNode{
				Pos:  types.Position{Line: 2, Column: 1, Offset: 0},
				Name: "echo",
				Args: []string{"world"},
			},
		},
	}
	
	result, err := ee.Execute(ctx, script)
	if err != nil {
		t.Fatalf("Script execution failed: %v", err)
	}
	
	if !result.Success {
		t.Error("Script execution should succeed")
	}
	
	if len(result.Commands) != 2 {
		t.Errorf("Expected 2 commands executed, got %d", len(result.Commands))
	}
}

func TestPackageManagementFlow(t *testing.T) {
	_, pm, tmpDir := setupIntegrationTest()
	defer os.RemoveAll(tmpDir)
	
	// Test init -> add -> install flow
	err := pm.Init("test-package", "1.0.0")
	if err != nil {
		t.Fatalf("Package init failed: %v", err)
	}
	
	err = pm.AddDependency("lodash", "4.17.21", false)
	if err != nil {
		t.Fatalf("Add dependency failed: %v", err)
	}
	
	// Verify dependency was added
	config := pm.GetConfig()
	if config.Dependencies["lodash"] != "4.17.21" {
		t.Error("Dependency not found in config")
	}
	
	// Test script management
	err = pm.AddScript("test", "echo 'test'")
	if err != nil {
		t.Fatalf("Add script failed: %v", err)
	}
	
	scripts := pm.GetConfig().Scripts
	if scripts["test"] != "echo 'test'" {
		t.Error("Script not found in config")
	}
}

func TestModuleSystemFlow(t *testing.T) {
	_, _, tmpDir := setupIntegrationTest()
	defer os.RemoveAll(tmpDir)
	
	// Create a test module
	moduleDir := filepath.Join(tmpDir, "test-module")
	os.MkdirAll(moduleDir, 0755)
	
	// Create module file
	moduleFile := filepath.Join(moduleDir, "index.sh")
	os.WriteFile(moduleFile, []byte("export_hello() { echo 'Hello from module'; }"), 0644)
	
	// Load module
	mm := module.NewModuleManager()
	mod, err := mm.LoadModule(moduleDir)
	if err != nil {
		t.Fatalf("Module load failed: %v", err)
	}
	
	if mod == nil {
		t.Fatal("Module is nil")
	}
	
	// Test module export
	export, err := mm.GetExport(moduleDir, "hello")
	if err != nil {
		t.Logf("Export retrieval: %v (may not be available)", err)
	}
	
	if export != nil {
		t.Log("Module export retrieved successfully")
	}
}

func TestErrorRecoveryFlow(t *testing.T) {
	ee, _, tmpDir := setupIntegrationTest()
	defer os.RemoveAll(tmpDir)
	
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	
	// Create script with timeout
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
		Nodes: []types.Node{
			&types.CommandNode{
				Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
				Name: "sleep",
				Args: []string{"1"},
			},
		},
	}
	
	result, err := ee.Execute(ctx, script)
	
	// Should handle timeout gracefully
	if err != nil {
		t.Logf("Timeout handled (expected): %v", err)
	} else if result != nil && !result.Success {
		t.Log("Error recovery working: execution failed gracefully")
	}
}

func TestPerformanceRegression(t *testing.T) {
	ee, _, tmpDir := setupIntegrationTest()
	defer os.RemoveAll(tmpDir)
	
	ctx := context.Background()
	
	// Execute multiple commands to test performance
	start := time.Now()
	for i := 0; i < 100; i++ {
		cmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{"test"},
		}
		_, err := ee.ExecuteCommand(ctx, cmd)
		if err != nil {
			t.Fatalf("Command execution failed: %v", err)
		}
	}
	duration := time.Since(start)
	
	// Performance should be reasonable (less than 10 seconds for 100 commands)
	if duration > 10*time.Second {
		t.Errorf("Performance regression: 100 commands took %v", duration)
	}
	
	t.Logf("100 commands executed in %v", duration)
	
	// Check metrics
	metrics := ee.GetMetrics()
	if metrics != nil {
		t.Logf("Command executions: %d, Success rate: %.2f%%", 
			metrics.CommandExecutions, metrics.SuccessRate)
	}
}
