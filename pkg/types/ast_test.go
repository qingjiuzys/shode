package types

import (
	"testing"
)

func TestCommandNode(t *testing.T) {
	cmd := &CommandNode{
		Name: "echo",
		Args: []string{"hello", "world"},
	}

	if cmd.Name != "echo" {
		t.Errorf("CommandNode.Name = %v, want %v", cmd.Name, "echo")
	}

	if len(cmd.Args) != 2 {
		t.Errorf("len(CommandNode.Args) = %v, want %v", len(cmd.Args), 2)
	}
}

func TestAssignmentNode(t *testing.T) {
	assign := &AssignmentNode{
		Name:  "TEST",
		Value: "value",
	}

	if assign.Name != "TEST" {
		t.Errorf("AssignmentNode.Name = %v, want %v", assign.Name, "TEST")
	}

	if assign.Value != "value" {
		t.Errorf("AssignmentNode.Value = %v, want %v", assign.Value, "value")
	}
}

func TestArrayNode(t *testing.T) {
	arr := &ArrayNode{
		Name:   "myarray",
		Values: []string{"a", "b", "c"},
	}

	if arr.Name != "myarray" {
		t.Errorf("ArrayNode.Name = %v, want %v", arr.Name, "myarray")
	}

	if len(arr.Values) != 3 {
		t.Errorf("len(ArrayNode.Values) = %v, want %v", len(arr.Values), 3)
	}
}

func TestFunctionNode(t *testing.T) {
	fn := &FunctionNode{
		Name: "testfunc",
		Body: &ScriptNode{
			Nodes: []Node{
				&CommandNode{Name: "echo", Args: []string{"test"}},
			},
		},
	}

	if fn.Name != "testfunc" {
		t.Errorf("FunctionNode.Name = %v, want %v", fn.Name, "testfunc")
	}

	if fn.Body == nil {
		t.Fatal("FunctionNode.Body is nil")
	}
}

func TestPipeNode(t *testing.T) {
	pipe := &PipeNode{
		Left: &CommandNode{Name: "echo", Args: []string{"hello"}},
		Right: &CommandNode{Name: "cat", Args: []string{}},
	}

	if pipe.Left == nil {
		t.Error("PipeNode.Left is nil")
	}

	if pipe.Right == nil {
		t.Error("PipeNode.Right is nil")
	}
}

func TestScriptNode(t *testing.T) {
	script := &ScriptNode{
		Nodes: []Node{
			&CommandNode{Name: "echo", Args: []string{"test"}},
		},
	}

	if len(script.Nodes) != 1 {
		t.Errorf("len(ScriptNode.Nodes) = %v, want %v", len(script.Nodes), 1)
	}
}

func TestRedirectNode(t *testing.T) {
	redirect := &RedirectNode{
		Op:   ">",
		File: "/tmp/output.txt",
		Fd:   1,
	}

	if redirect.Op != ">" {
		t.Errorf("RedirectNode.Op = %v, want %v", redirect.Op, ">")
	}

	if redirect.File != "/tmp/output.txt" {
		t.Errorf("RedirectNode.File = %v, want %v", redirect.File, "/tmp/output.txt")
	}
}

func TestForNode(t *testing.T) {
	forNode := &ForNode{
		Variable: "i",
		List:     []string{"1", "2", "3"},
		Body:     &ScriptNode{Nodes: []Node{}},
	}

	if forNode.Variable != "i" {
		t.Errorf("ForNode.Variable = %v, want %v", forNode.Variable, "i")
	}

	if len(forNode.List) != 3 {
		t.Errorf("len(ForNode.List) = %v, want %v", len(forNode.List), 3)
	}
}
