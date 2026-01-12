package parser

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/types"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
)

// Parser represents a shell script parser
type Parser struct {
	language *sitter.Language
}

// NewParser creates a new shell script parser
func NewParser() *Parser {
	return &Parser{
		language: bash.GetLanguage(),
	}
}

// ParseString parses a shell script from a string
func (p *Parser) ParseString(source string) (*types.ScriptNode, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)

	tree, err := parser.ParseCtx(nil, nil, []byte(source))
	if err != nil {
		return nil, fmt.Errorf("failed to parse script: %v", err)
	}
	defer tree.Close()

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("failed to get root node")
	}

	return p.buildAST(rootNode, source), nil
}

// ParseFile parses a shell script from a file
func (p *Parser) ParseFile(filename string) (*types.ScriptNode, error) {
	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filename, err)
	}

	// Parse the content using ParseString
	return p.ParseString(string(content))
}

// buildAST converts tree-sitter nodes to our AST structure
func (p *Parser) buildAST(node *sitter.Node, source string) *types.ScriptNode {
	script := &types.ScriptNode{
		Pos: types.Position{
			Line:   int(node.StartPoint().Row) + 1,
			Column: int(node.StartPoint().Column) + 1,
			Offset: int(node.StartByte()),
		},
	}

	// Simple initial implementation - just extract commands
	p.walkTree(node, source, script)

	return script
}

// walkTree recursively walks the parse tree and builds our AST
func (p *Parser) walkTree(node *sitter.Node, source string, script *types.ScriptNode) {
	// Extract node content
	nodeText := node.Content([]byte(source))

	switch node.Type() {
	case "program":
		// Process all children of program
		for i := 0; i < int(node.ChildCount()); i++ {
			p.walkTree(node.Child(i), source, script)
		}

	case "command":
		// Simple command detection for now
		if strings.TrimSpace(nodeText) != "" {
			cmd := &types.CommandNode{
				Pos: types.Position{
					Line:   int(node.StartPoint().Row) + 1,
					Column: int(node.StartPoint().Column) + 1,
					Offset: int(node.StartByte()),
				},
				Name: "raw_command", // Placeholder
				Args: []string{nodeText},
			}
			script.Nodes = append(script.Nodes, cmd)
		}

	default:
		// Recursively process children for other node types
		for i := 0; i < int(node.ChildCount()); i++ {
			p.walkTree(node.Child(i), source, script)
		}
	}
}

// DebugPrint prints the parse tree for debugging
func (p *Parser) DebugPrint(source string) {
	parser := sitter.NewParser()
	parser.SetLanguage(p.language)

	tree, err := parser.ParseCtx(nil, nil, []byte(source))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}
	defer tree.Close()

	fmt.Println("Parse tree:")
	p.printNode(tree.RootNode(), 0, source)
}

// printNode recursively prints the parse tree structure
func (p *Parser) printNode(node *sitter.Node, depth int, source string) {
	indent := strings.Repeat("  ", depth)
	content := node.Content([]byte(source))
	fmt.Printf("%s%s: %q\n", indent, node.Type(), content)

	for i := 0; i < int(node.ChildCount()); i++ {
		p.printNode(node.Child(i), depth+1, source)
	}
}
