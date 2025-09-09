package engine

import (
	"context"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestExecutionEngine_ExecuteCommand(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name     string
		command  *types.CommandNode
		wantMode ExecutionMode
		wantErr  bool
	}{
		{
			name: "standard library function",
			command: &types.CommandNode{
				Name: "upper",
				Args: []string{"hello"},
			},
			wantMode: ModeInterpreted,
			wantErr:  false,
		},
		{
			name: "external command",
			command: &types.CommandNode{
				Name: "echo",
				Args: []string{"test"},
			},
			wantMode: ModeProcess,
			wantErr:  false,
		},
		{
			name: "nonexistent command",
			command: &types.CommandNode{
				Name: "nonexistent_command_123",
				Args: []string{},
			},
			wantMode: ModeProcess,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := engine.ExecuteCommand(ctx, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if result.Mode != tt.wantMode {
					t.Errorf("ExecuteCommand() mode = %v, want %v", result.Mode, tt.wantMode)
				}

				// For successful commands, verify basic properties
				if result.Command != tt.command {
					t.Error("ExecuteCommand() should return the original command")
				}

				if tt.wantMode == ModeInterpreted && result.Output == "" {
					t.Error("Interpreted commands should have output")
				}
			}
		})
	}
}

func TestExecutionEngine_decideExecutionMode(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name     string
		command  string
		wantMode ExecutionMode
	}{
		{
			name:     "standard library function",
			command:  "upper",
			wantMode: ModeInterpreted,
		},
		{
			name:     "external command that is not in stdlib",
			command:  "date",  // Use date instead of ls since ls is in stdlib
			wantMode: ModeProcess,
		},
		{
			name:     "nonexistent command",
			command:  "nonexistent_command_xyz",
			wantMode: ModeProcess, // Should default to process mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmdNode := &types.CommandNode{Name: tt.command}
			mode := engine.decideExecutionMode(cmdNode)

			if mode != tt.wantMode {
				t.Errorf("decideExecutionMode() = %v, want %v", mode, tt.wantMode)
			}
		})
	}
}

func TestExecutionEngine_isStdLibFunction(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name     string
		function string
		want     bool
	}{
		{
			name:     "existing stdlib function",
			function: "upper",
			want:     true,
		},
		{
			name:     "another stdlib function",
			function: "contains",
			want:     true,
		},
		{
			name:     "nonexistent function",
			function: "nonexistent_func",
			want:     false,
		},
		{
			name:     "external command not in stdlib",
			function: "date",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.isStdLibFunction(tt.function)
			if result != tt.want {
				t.Errorf("isStdLibFunction(%s) = %v, want %v", tt.function, result, tt.want)
			}
		})
	}
}

func TestExecutionEngine_isExternalCommandAvailable(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{
			name:    "existing command",
			command: "ls",
			want:    true,
		},
		{
			name:    "another existing command",
			command: "echo",
			want:    true,
		},
		{
			name:    "nonexistent command",
			command: "nonexistent_command_abc123",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.isExternalCommandAvailable(tt.command)
			if result != tt.want {
				t.Errorf("isExternalCommandAvailable(%s) = %v, want %v", tt.command, result, tt.want)
			}
		})
	}
}

func TestExecutionEngine_ExecuteStdLibFunction(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name     string
		function string
		args     []string
		wantErr  bool
	}{
		{
			name:     "upper function",
			function: "upper",
			args:     []string{"hello"},
			wantErr:  false,
		},
		{
			name:     "contains function",
			function: "contains",
			args:     []string{"hello world", "world"},
			wantErr:  false,
		},
		{
			name:     "nonexistent function",
			function: "nonexistent_func",
			args:     []string{},
			wantErr:  true,
		},
		{
			name:     "invalid arguments",
			function: "contains",
			args:     []string{"only_one_arg"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.executeStdLibFunction(tt.function, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("executeStdLibFunction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if result == "" && tt.function != "trim" {
					t.Errorf("executeStdLibFunction() returned empty result for %s", tt.function)
				}
			}
		})
	}
}

func TestExecutionEngine_CacheIntegration(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	// Test that cache is working for external commands
	cmd := &types.CommandNode{
		Name: "echo",
		Args: []string{"cached_test"},
	}

	ctx := context.Background()

	// First execution - should not be cached
	start1 := time.Now()
	result1, err := engine.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}
	duration1 := time.Since(start1)

	// Second execution - should be cached and faster
	start2 := time.Now()
	result2, err := engine.ExecuteCommand(ctx, cmd)
	if err != nil {
		t.Fatalf("Second execution failed: %v", err)
	}
	duration2 := time.Since(start2)

	// Cached execution should be significantly faster
	if duration2 >= duration1 {
		t.Errorf("Cached execution should be faster. First: %v, Second: %v", duration1, duration2)
	}

	// Results should be identical
	if result1.Output != result2.Output {
		t.Errorf("Cached results should be identical. First: %s, Second: %s", result1.Output, result2.Output)
	}

	if result1.Success != result2.Success {
		t.Errorf("Success status should match. First: %v, Second: %v", result1.Success, result2.Success)
	}
}

func TestExecutionEngine_ExecuteInterpreted(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name    string
		command *types.CommandNode
		want    string
		wantErr bool
	}{
		{
			name: "upper function",
			command: &types.CommandNode{
				Name: "upper",
				Args: []string{"hello"},
			},
			want:    "HELLO",
			wantErr: false,
		},
		{
			name: "trim function",
			command: &types.CommandNode{
				Name: "trim",
				Args: []string{"   hello   "},
			},
			want:    "hello",
			wantErr: false,
		},
		{
			name: "nonexistent function",
			command: &types.CommandNode{
				Name: "nonexistent_func",
				Args: []string{},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := engine.executeInterpreted(ctx, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("executeInterpreted() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if result.Output != tt.want {
					t.Errorf("executeInterpreted() output = %v, want %v", result.Output, tt.want)
				}

				if !result.Success {
					t.Error("executeInterpreted() should be successful")
				}

				if result.Mode != ModeInterpreted {
					t.Errorf("executeInterpreted() mode = %v, want %v", result.Mode, ModeInterpreted)
				}
			}
		})
	}
}

func TestExecutionEngine_ExecuteProcess(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	tests := []struct {
		name    string
		command *types.CommandNode
		wantErr bool
	}{
		{
			name: "echo command",
			command: &types.CommandNode{
				Name: "echo",
				Args: []string{"test"},
			},
			wantErr: false,
		},
		{
			name: "nonexistent command",
			command: &types.CommandNode{
				Name: "nonexistent_command_xyz",
				Args: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := engine.executeProcess(ctx, tt.command)

			if (err != nil) != tt.wantErr {
				t.Errorf("executeProcess() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if result.Mode != ModeProcess {
					t.Errorf("executeProcess() mode = %v, want %v", result.Mode, ModeProcess)
				}

				if tt.command.Name == "echo" && result.Output == "" {
					t.Error("echo command should have output")
				}
			}
		})
	}
}

func TestExecutionEngine_Performance(t *testing.T) {
	envManager := environment.NewEnvironmentManager()
	stdlib := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()

	engine := NewExecutionEngine(envManager, stdlib, moduleMgr, security)

	// Test performance of interpreted vs process execution
	interpretedCmd := &types.CommandNode{
		Name: "upper",
		Args: []string{"test"},
	}

	processCmd := &types.CommandNode{
		Name: "echo",
		Args: []string{"test"},
	}

	ctx := context.Background()

	// Test interpreted performance
	startInterpreted := time.Now()
	for i := 0; i < 100; i++ {
		_, err := engine.ExecuteCommand(ctx, interpretedCmd)
		if err != nil {
			t.Fatalf("Interpreted performance test failed: %v", err)
		}
	}
	durationInterpreted := time.Since(startInterpreted)

	// Test process performance
	startProcess := time.Now()
	for i := 0; i < 100; i++ {
		_, err := engine.ExecuteCommand(ctx, processCmd)
		if err != nil {
			t.Fatalf("Process performance test failed: %v", err)
		}
	}
	durationProcess := time.Since(startProcess)

	t.Logf("100 interpreted operations took: %v", durationInterpreted)
	t.Logf("100 process operations took: %v", durationProcess)

	// Interpreted should be significantly faster than process execution
	if durationInterpreted >= durationProcess {
		t.Errorf("Interpreted execution should be faster than process. Interpreted: %v, Process: %v",
			durationInterpreted, durationProcess)
	}

	// Both should complete in reasonable time
	maxDuration := 5 * time.Second
	if durationInterpreted > maxDuration || durationProcess > maxDuration {
		t.Errorf("Performance test took too long. Interpreted: %v, Process: %v",
			durationInterpreted, durationProcess)
	}
}
