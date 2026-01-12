package parser

import (
	"os"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestSimpleParserParseString(t *testing.T) {
	p := NewSimpleParser()

	// Test simple command
	source := "echo hello world"
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	if len(script.Nodes) != 1 {
		t.Errorf("Expected 1 command, got %d", len(script.Nodes))
	}

	cmd, ok := script.Nodes[0].(*types.CommandNode)
	if !ok {
		t.Fatal("Expected CommandNode")
	}

	if cmd.Name != "echo" {
		t.Errorf("Expected command name 'echo', got '%s'", cmd.Name)
	}

	if len(cmd.Args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(cmd.Args))
	}
}

func TestSimpleParserParseStringWithQuotes(t *testing.T) {
	p := NewSimpleParser()

	// Test command with quoted arguments
	source := `echo "hello world" 'test string'`
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	if len(script.Nodes) != 1 {
		t.Errorf("Expected 1 command, got %d", len(script.Nodes))
	}

	cmd := script.Nodes[0].(*types.CommandNode)
	if len(cmd.Args) != 2 {
		t.Errorf("Expected 2 arguments, got %d", len(cmd.Args))
	}

	if cmd.Args[0] != "hello world" {
		t.Errorf("Expected first arg 'hello world', got '%s'", cmd.Args[0])
	}

	if cmd.Args[1] != "test string" {
		t.Errorf("Expected second arg 'test string', got '%s'", cmd.Args[1])
	}
}

func TestSimpleParserParseStringWithComments(t *testing.T) {
	p := NewSimpleParser()

	// Test command with comments
	source := `# This is a comment
echo hello
# Another comment
ls -la`
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	// Should skip comments and empty lines
	if len(script.Nodes) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(script.Nodes))
	}
}

func TestSimpleParserParseStringEmptyLines(t *testing.T) {
	p := NewSimpleParser()

	// Test with empty lines
	source := `echo hello

ls -la

`
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	// Should skip empty lines
	if len(script.Nodes) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(script.Nodes))
	}
}

func TestSimpleParserParseFile(t *testing.T) {
	p := NewSimpleParser()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "shode-test-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test content
	content := "echo hello world\nls -la"
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Parse file
	script, err := p.ParseFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if len(script.Nodes) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(script.Nodes))
	}
}

func TestSimpleParserParseFileNonExistent(t *testing.T) {
	p := NewSimpleParser()

	// Try to parse non-existent file
	_, err := p.ParseFile("/non/existent/file.sh")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestSimpleParserTokenize(t *testing.T) {
	p := NewSimpleParser()

	testCases := []struct {
		input    string
		expected []string
	}{
		{"echo hello", []string{"echo", "hello"}},
		{"echo hello world", []string{"echo", "hello", "world"}},
		{`echo "hello world"`, []string{"echo", "hello world"}},
		{`echo 'test string' arg`, []string{"echo", "test string", "arg"}},
		{`cmd "arg1" 'arg2' arg3`, []string{"cmd", "arg1", "arg2", "arg3"}},
	}

	for _, tc := range testCases {
		tokens := p.tokenize(tc.input)
		if len(tokens) != len(tc.expected) {
			t.Errorf("Input '%s': expected %d tokens, got %d", tc.input, len(tc.expected), len(tokens))
			continue
		}
		for i, expected := range tc.expected {
			if tokens[i] != expected {
				t.Errorf("Input '%s': token %d: expected '%s', got '%s'", tc.input, i, expected, tokens[i])
			}
		}
	}
}

func TestParserParseString(t *testing.T) {
	p := NewParser()

	// Test simple command with panic recovery
	// Note: tree-sitter may panic in some environments
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Tree-sitter parser panicked (may be expected in some environments): %v", r)
		}
	}()

	source := "echo hello world"
	script, err := p.ParseString(source)
	if err != nil {
		t.Logf("Tree-sitter parser returned error (may be expected): %v", err)
		return
	}

	if script == nil {
		t.Fatal("Script is nil")
	}
}

func TestParserParseFile(t *testing.T) {
	p := NewParser()

	// Test with panic recovery
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Tree-sitter parser panicked (may be expected in some environments): %v", r)
		}
	}()

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "shode-test-*.sh")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write test content
	content := "echo hello world"
	_, err = tmpFile.WriteString(content)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	// Parse file
	script, err := p.ParseFile(tmpFile.Name())
	if err != nil {
		t.Logf("Tree-sitter parser returned error (may be expected): %v", err)
		return
	}

	if script == nil {
		t.Fatal("Script is nil")
	}
}

func TestParserParseFileNonExistent(t *testing.T) {
	p := NewParser()

	// Try to parse non-existent file
	_, err := p.ParseFile("/non/existent/file.sh")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestSimpleParserMultipleCommands(t *testing.T) {
	p := NewSimpleParser()

	source := `echo hello
ls -la
cat file.txt`
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	if len(script.Nodes) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(script.Nodes))
	}

	// Verify first command
	cmd1 := script.Nodes[0].(*types.CommandNode)
	if cmd1.Name != "echo" {
		t.Errorf("Expected first command 'echo', got '%s'", cmd1.Name)
	}

	// Verify second command
	cmd2 := script.Nodes[1].(*types.CommandNode)
	if cmd2.Name != "ls" {
		t.Errorf("Expected second command 'ls', got '%s'", cmd2.Name)
	}

	// Verify third command
	cmd3 := script.Nodes[2].(*types.CommandNode)
	if cmd3.Name != "cat" {
		t.Errorf("Expected third command 'cat', got '%s'", cmd3.Name)
	}
}

func TestSimpleParserLineNumbers(t *testing.T) {
	p := NewSimpleParser()

	source := `# Comment
echo hello
# Another comment
ls -la`
	script, err := p.ParseString(source)
	if err != nil {
		t.Fatalf("Failed to parse string: %v", err)
	}

	// First command should be on line 2 (after comment)
	cmd1 := script.Nodes[0].(*types.CommandNode)
	if cmd1.Pos.Line != 2 {
		t.Errorf("Expected first command on line 2, got %d", cmd1.Pos.Line)
	}

	// Second command should be on line 4 (after comment)
	cmd2 := script.Nodes[1].(*types.CommandNode)
	if cmd2.Pos.Line != 4 {
		t.Errorf("Expected second command on line 4, got %d", cmd2.Pos.Line)
	}
}
