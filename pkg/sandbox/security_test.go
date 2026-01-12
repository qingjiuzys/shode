package sandbox

import (
	"testing"

	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestDangerousCommandBlocking(t *testing.T) {
	sc := NewSecurityChecker()

	// Test dangerous commands
	dangerousCommands := []string{"rm", "dd", "mkfs", "shutdown", "reboot"}
	for _, cmd := range dangerousCommands {
		testCmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: cmd,
			Args: []string{},
		}
		err := sc.CheckCommand(testCmd)
		if err == nil {
			t.Errorf("Expected error for dangerous command '%s', got nil", cmd)
		}
	}
}

func TestNetworkCommandBlocking(t *testing.T) {
	sc := NewSecurityChecker()

	// Test network commands
	networkCommands := []string{"iptables", "ufw", "route", "ifconfig", "ip"}
	for _, cmd := range networkCommands {
		testCmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: cmd,
			Args: []string{},
		}
		err := sc.CheckCommand(testCmd)
		if err == nil {
			t.Errorf("Expected error for network command '%s', got nil", cmd)
		}
	}
}

func TestSensitiveFileProtection(t *testing.T) {
	sc := NewSecurityChecker()

	// Test sensitive files
	sensitiveFiles := []string{
		"/etc/passwd",
		"/etc/shadow",
		"/etc/sudoers",
		"/root/secret.txt",
		"/boot/config",
	}

	for _, file := range sensitiveFiles {
		testCmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "cat",
			Args: []string{file},
		}
		err := sc.CheckCommand(testCmd)
		if err == nil {
			t.Errorf("Expected error for sensitive file '%s', got nil", file)
		}
	}
}

func TestRecursiveDeletePattern(t *testing.T) {
	sc := NewSecurityChecker()

	// Test recursive delete pattern
	testCmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "rm",
		Args: []string{"-rf", "/"},
	}
	err := sc.CheckCommand(testCmd)
	if err == nil {
		t.Error("Expected error for recursive delete of root, got nil")
	}
}

func TestPasswordInCommandLine(t *testing.T) {
	sc := NewSecurityChecker()

	// Test password in command line
	testCmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "mysql",
		Args: []string{"-p", "secretpassword"},
	}
	err := sc.CheckCommand(testCmd)
	if err == nil {
		t.Error("Expected error for password in command line, got nil")
	}
}

func TestShellInjectionPattern(t *testing.T) {
	sc := NewSecurityChecker()

	// Test shell injection patterns
	injectionPatterns := []string{
		"echo hello; rm -rf /",
		"echo hello | cat",
		"echo hello `rm -rf /`",
		"echo hello $(rm -rf /)",
	}

	for _, pattern := range injectionPatterns {
		testCmd := &types.CommandNode{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "sh",
			Args: []string{"-c", pattern},
		}
		err := sc.CheckCommand(testCmd)
		if err == nil {
			t.Errorf("Expected error for shell injection pattern '%s', got nil", pattern)
		}
	}
}

func TestSafeCommand(t *testing.T) {
	sc := NewSecurityChecker()

	// Test safe commands
	safeCommands := []*types.CommandNode{
		{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "echo",
			Args: []string{"hello"},
		},
		{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "ls",
			Args: []string{"-la"},
		},
		{
			Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
			Name: "cat",
			Args: []string{"/tmp/test.txt"},
		},
	}

	for _, cmd := range safeCommands {
		err := sc.CheckCommand(cmd)
		if err != nil {
			t.Errorf("Expected no error for safe command '%s', got: %v", cmd.Name, err)
		}
	}
}

func TestAddRemoveDangerousCommand(t *testing.T) {
	sc := NewSecurityChecker()

	// Test adding custom dangerous command
	sc.AddDangerousCommand("custom-dangerous")
	testCmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "custom-dangerous",
		Args: []string{},
	}
	err := sc.CheckCommand(testCmd)
	if err == nil {
		t.Error("Expected error for custom dangerous command, got nil")
	}

	// Test removing dangerous command
	sc.RemoveDangerousCommand("custom-dangerous")
	err = sc.CheckCommand(testCmd)
	if err != nil {
		t.Errorf("Expected no error after removing command, got: %v", err)
	}
}

func TestAddRemoveSensitiveFile(t *testing.T) {
	sc := NewSecurityChecker()

	// Test adding custom sensitive file
	sc.AddSensitiveFile("/custom/sensitive")
	testCmd := &types.CommandNode{
		Pos:  types.Position{Line: 1, Column: 1, Offset: 0},
		Name: "cat",
		Args: []string{"/custom/sensitive"},
	}
	err := sc.CheckCommand(testCmd)
	if err == nil {
		t.Error("Expected error for custom sensitive file, got nil")
	}

	// Test removing sensitive file
	sc.RemoveSensitiveFile("/custom/sensitive")
	err = sc.CheckCommand(testCmd)
	if err != nil {
		t.Errorf("Expected no error after removing file, got: %v", err)
	}
}

func TestGetSecurityReport(t *testing.T) {
	sc := NewSecurityChecker()

	testCmd := &types.CommandNode{
		Pos:  types.Position{Line: 10, Column: 5, Offset: 100},
		Name: "rm",
		Args: []string{"-rf", "/tmp/test"},
	}

	report := sc.GetSecurityReport(testCmd)
	if report == nil {
		t.Fatal("Security report is nil")
	}

	if report["command"] != "rm" {
		t.Errorf("Expected command 'rm', got '%v'", report["command"])
	}

	if report["line_number"] != 10 {
		t.Errorf("Expected line number 10, got %v", report["line_number"])
	}

	isDangerous, ok := report["is_dangerous_command"].(bool)
	if !ok || !isDangerous {
		t.Error("Expected is_dangerous_command to be true")
	}
}
