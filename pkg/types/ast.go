package types

// Node represents a generic AST node
type Node interface {
	Position() Position
	String() string
}

// Position represents the source code position of a node
type Position struct {
	Line   int
	Column int
	Offset int
}

// CommandNode represents a shell command
type CommandNode struct {
	Pos      Position
	Name     string
	Args     []string
	Redirect *RedirectNode
}

func (n *CommandNode) Position() Position { return n.Pos }
func (n *CommandNode) String() string     { return n.Name }

// PipeNode represents a pipe between commands
type PipeNode struct {
	Pos  Position
	Left Node
	Right Node
}

func (n *PipeNode) Position() Position { return n.Pos }
func (n *PipeNode) String() string     { return "|" }

// RedirectNode represents input/output redirection
type RedirectNode struct {
	Pos   Position
	Op    string // >, >>, <, etc.
	File  string
	Fd    int // file descriptor (0, 1, 2)
}

func (n *RedirectNode) Position() Position { return n.Pos }
func (n *RedirectNode) String() string     { return n.Op }

// ScriptNode represents a complete shell script
type ScriptNode struct {
	Pos   Position
	Nodes []Node
}

func (n *ScriptNode) Position() Position { return n.Pos }
func (n *ScriptNode) String() string     { return "script" }

// IfNode represents an if-then-else statement
type IfNode struct {
	Pos       Position
	Condition Node
	Then      *ScriptNode
	Else      *ScriptNode // optional
}

func (n *IfNode) Position() Position { return n.Pos }
func (n *IfNode) String() string     { return "if" }

// ForNode represents a for loop
type ForNode struct {
	Pos      Position
	Variable string
	List     []string
	Body     *ScriptNode
}

func (n *ForNode) Position() Position { return n.Pos }
func (n *ForNode) String() string     { return "for" }

// WhileNode represents a while loop
type WhileNode struct {
	Pos       Position
	Condition Node
	Body      *ScriptNode
}

func (n *WhileNode) Position() Position { return n.Pos }
func (n *WhileNode) String() string     { return "while" }

// FunctionNode represents a function definition
type FunctionNode struct {
	Pos  Position
	Name string
	Body *ScriptNode
}

func (n *FunctionNode) Position() Position { return n.Pos }
func (n *FunctionNode) String() string     { return "function" }

// AssignmentNode represents a variable assignment
type AssignmentNode struct {
	Pos   Position
	Name  string
	Value string
}

func (n *AssignmentNode) Position() Position { return n.Pos }
func (n *AssignmentNode) String() string     { return "assignment" }

// BreakNode represents a break statement
type BreakNode struct {
	Pos Position
}

func (n *BreakNode) Position() Position { return n.Pos }
func (n *BreakNode) String() string     { return "break" }

// ContinueNode represents a continue statement
type ContinueNode struct {
	Pos Position
}

func (n *ContinueNode) Position() Position { return n.Pos }
func (n *ContinueNode) String() string     { return "continue" }

// AnnotationNode represents an annotation
type AnnotationNode struct {
	Pos   Position
	Name  string
	Value string
}

func (n *AnnotationNode) Position() Position { return n.Pos }
func (n *AnnotationNode) String() string     { return "annotation" }
