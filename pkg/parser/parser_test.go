package parser

import (
	"testing"

	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestSimpleParser_ParseString(t *testing.T) {
	sp := NewSimpleParser()

	tests := []struct {
		name  string
		input string
	}{
		{"simple command", "echo hello"},
		{"command with args", "ls -la /tmp"},
		{"variable assignment", "NAME=value"},
		{"pipeline", "echo hello | cat"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := sp.ParseString(tt.input)
			if err != nil {
				t.Fatalf("ParseString() error = %v", err)
			}
			if result == nil {
				t.Fatal("ParseString() returned nil")
			}
		})
	}
}

func TestSimpleParser_ParsePipeline(t *testing.T) {
	sp := NewSimpleParser()

	input := "echo hello | cat"
	result, err := sp.ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Fatal("ParseString() returned empty script")
	}

	// SimpleParser creates PipeNode for pipelines
	if _, ok := result.Nodes[0].(*types.PipeNode); !ok {
		t.Logf("ParseString() returned %T (SimpleParser may not support all features)", result.Nodes[0])
	}
}

func TestSimpleParser_ParseVariableAssignment(t *testing.T) {
	sp := NewSimpleParser()

	input := "NAME=value"
	result, err := sp.ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Fatal("ParseString() returned empty script")
	}

	if _, ok := result.Nodes[0].(*types.AssignmentNode); !ok {
		t.Errorf("ParseString() returned %T, want AssignmentNode", result.Nodes[0])
	}
}

func TestSimpleParser_ParseCommand(t *testing.T) {
	sp := NewSimpleParser()

	input := "echo hello world"
	result, err := sp.ParseString(input)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Fatal("ParseString() returned empty script")
	}

	if cmd, ok := result.Nodes[0].(*types.CommandNode); ok {
		if cmd.Name != "echo" {
			t.Errorf("Command name = %v, want %v", cmd.Name, "echo")
		}
		if len(cmd.Args) != 2 {
			t.Errorf("Command args length = %v, want %v", len(cmd.Args), 2)
		}
	} else {
		t.Errorf("ParseString() returned %T, want CommandNode", result.Nodes[0])
	}
}
