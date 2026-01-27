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

	tree := parser.Parse(nil, []byte(source))
	if tree == nil {
		return nil, fmt.Errorf("failed to parse script")
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

	switch node.Type() {
	case "program":
		// Process all children of program
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)

			// Check if this child is a command and next child is & (background job)
			if child.Type() == "command" && i+1 < int(node.ChildCount()) && node.Child(i+1).Type() == "&" {
				// This is a background job
				bgNode := p.parseBackgroundCommand(child, source)
				if bgNode != nil {
					script.Nodes = append(script.Nodes, bgNode)
				}
				i++ // Skip the & token
				continue
			}

			p.walkTree(child, source, script)
		}

	case "pipeline":
		// Process pipeline - create PipeNode
		pipeNode := p.parsePipeline(node, source)
		if pipeNode != nil {
			script.Nodes = append(script.Nodes, pipeNode)
		}

	case "list":
		// Process list with logical operators (&&, ||)
		listNode := p.parseList(node, source)
		if listNode != nil {
			script.Nodes = append(script.Nodes, listNode)
		}

	case "redirected_statement":
		// Process redirected statement - command with redirection
		// Check if it's a heredoc
		heredocNode := p.parseHeredocFromRedirected(node, source)
		if heredocNode != nil {
			script.Nodes = append(script.Nodes, heredocNode)
			break
		}
		// Parse command with redirection info
		cmdNode := p.parseRedirectedStatement(node, source)
		if cmdNode != nil {
			script.Nodes = append(script.Nodes, cmdNode)
		}

	case "if_statement":
		// Process if statement
		ifNode := p.parseIfStatement(node, source)
		if ifNode != nil {
			script.Nodes = append(script.Nodes, ifNode)
		}

	case "for_statement":
		// Process for loop
		forNode := p.parseForStatement(node, source)
		if forNode != nil {
			script.Nodes = append(script.Nodes, forNode)
		}

	case "while_statement":
		// Process while loop
		whileNode := p.parseWhileStatement(node, source)
		if whileNode != nil {
			script.Nodes = append(script.Nodes, whileNode)
		}

	case "command":
		// Process simple command
		// Check for background job (&)
		hasBackground := p.isBackgroundCommand(node)

		if hasBackground {
			bgNode := p.parseBackgroundCommand(node, source)
			if bgNode != nil {
				script.Nodes = append(script.Nodes, bgNode)
			}
			break
		}

		cmd := p.parseCommandNode(node, source)
		if cmd != nil {
			script.Nodes = append(script.Nodes, cmd)
		}

	case "variable_assignment":
		// Process variable assignment
		assign := p.parseVariableAssignment(node, source)
		if assign != nil {
			script.Nodes = append(script.Nodes, assign)
		}

	case "function_definition":
		// Process function definition
		funcNode := p.parseFunctionDefinition(node, source)
		if funcNode != nil {
			script.Nodes = append(script.Nodes, funcNode)
		}

	default:
		// Skip heredoc-related nodes to avoid processing EOF markers as commands
		if node.Type() == "heredoc_start" || node.Type() == "heredoc_body" || node.Type() == "heredoc_end" {
			return
		}
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

	tree := parser.Parse(nil, []byte(source))
	if tree == nil {
		fmt.Println("Parse error: failed to create tree")
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

// parsePipeline parses a pipeline node
func (p *Parser) parsePipeline(node *sitter.Node, source string) *types.PipeNode {
	var commands []types.Node

	// Collect all commands in pipeline
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "command" {
			cmd := p.parseCommandNode(child, source)
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
	}

	// Build pipeline tree from left to right
	if len(commands) < 2 {
		return nil
	}

	current := commands[0]
	for i := 1; i < len(commands); i++ {
		current = &types.PipeNode{
			Pos:   types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
			Left:  current,
			Right: commands[i],
		}
	}

	return current.(*types.PipeNode)
}

// parseCommandNode parses a command node
func (p *Parser) parseCommandNode(node *sitter.Node, source string) *types.CommandNode {
	var name string
	var args []string

	// Process children to find command name and arguments
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command_name":
			for j := 0; j < int(child.ChildCount()); j++ {
				grandchild := child.Child(j)
				if grandchild.Type() == "word" {
					name = grandchild.Content([]byte(source))
				}
			}
		case "word":
			content := child.Content([]byte(source))
			args = append(args, strings.Trim(content, `"'`))
		case "string":
			content := child.Content([]byte(source))
			// Remove surrounding quotes from strings
			args = append(args, strings.Trim(content, `"'`))
		case "number":
			content := child.Content([]byte(source))
			// Numbers don't have quotes
			args = append(args, content)
		}
	}

	if name == "" {
		return nil
	}

	return &types.CommandNode{
		Pos:  types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Name: name,
		Args: args,
	}
}

// parseIfStatement parses an if statement
func (p *Parser) parseIfStatement(node *sitter.Node, source string) *types.IfNode {
	var condition types.Node
	var thenScript, elseScript *types.ScriptNode
	var foundThen bool

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		nodeType := child.Type()

		if nodeType == "if" || nodeType == "fi" {
			// Skip keywords
			continue
		}

		if nodeType == "then" {
			foundThen = true
			// Initialize then script
			if thenScript == nil {
				thenScript = &types.ScriptNode{
					Pos: types.Position{
						Line:   int(child.StartPoint().Row) + 1,
						Column: int(child.StartPoint().Column) + 1,
						Offset: int(child.StartByte()),
					},
				}
			}
			continue
		}

		if nodeType == "test_command" {
			if !foundThen {
				// For test_command, always use parseTestCommand
				condition = p.parseTestCommand(child, source)
			}
			continue
		}

		if nodeType == "command" {
			if !foundThen {
				condition = p.parseCommandNode(child, source)
			} else if foundThen && thenScript != nil {
				// Command in then block
				cmd := p.parseCommandNode(child, source)
				if cmd != nil {
					thenScript.Nodes = append(thenScript.Nodes, cmd)
				}
			}
			continue
		}

		if nodeType == "command_substitution" && !foundThen {
			condition = p.parseCommandSubstitution(child, source)
			continue
		}

		if nodeType == "do_group" {
			thenScript = p.parseDoGroup(child, source)
			continue
		}

		if nodeType == "else_clause" {
			elseScript = p.parseElseClause(child, source)
			continue
		}

		if nodeType == "pipeline" && foundThen && thenScript != nil {
			pipe := p.parsePipeline(child, source)
			if pipe != nil {
				thenScript.Nodes = append(thenScript.Nodes, pipe)
			}
			continue
		}
	}

	if condition == nil || thenScript == nil {
		return nil
	}

	return &types.IfNode{
		Pos:       types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Condition: condition,
		Then:      thenScript,
		Else:      elseScript,
	}
}

// parseForStatement parses a for loop
func (p *Parser) parseForStatement(node *sitter.Node, source string) *types.ForNode {
	var variable string
	var list []string
	var bodyScript *types.ScriptNode

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "variable_name":
			variable = child.Content([]byte(source))
		case "word", "string", "number":
			list = append(list, child.Content([]byte(source)))
		case "simple_expansion":
			list = append(list, child.Content([]byte(source)))
		case "do_group":
			bodyScript = p.parseDoGroup(child, source)
		}
	}

	if variable == "" || bodyScript == nil {
		return nil
	}

	return &types.ForNode{
		Pos:      types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Variable: variable,
		List:     list,
		Body:     bodyScript,
	}
}

// parseWhileStatement parses a while loop
func (p *Parser) parseWhileStatement(node *sitter.Node, source string) *types.WhileNode {
	var condition types.Node
	var bodyScript *types.ScriptNode

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command", "test_command":
			condition = p.parseCommandNode(child, source)
			if condition == nil {
				condition = p.parseTestCommand(child, source)
			}
		case "do_group":
			bodyScript = p.parseDoGroup(child, source)
		}
	}

	if condition == nil || bodyScript == nil {
		return nil
	}

	return &types.WhileNode{
		Pos:       types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Condition: condition,
		Body:      bodyScript,
	}
}

// parseVariableAssignment parses a variable assignment
func (p *Parser) parseVariableAssignment(node *sitter.Node, source string) *types.AssignmentNode {
	var name, value string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "variable_name":
			name = child.Content([]byte(source))
		case "number":
			value = child.Content([]byte(source))
		case "string":
			value = child.Content([]byte(source))
		case "word":
			value = child.Content([]byte(source))
		case "arithmetic_expansion":
			value = child.Content([]byte(source))
		case "simple_expansion":
			value = child.Content([]byte(source))
		case "array":
			// Parse array assignment: name=(value1 value2 ...)
			return p.parseArrayInAssignment(node, source)
		}
	}

	if name == "" {
		return nil
	}

	return &types.AssignmentNode{
		Pos:   types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Name:  name,
		Value: value,
	}
}

// parseArrayInAssignment parses an array assignment within a variable_assignment node
func (p *Parser) parseArrayInAssignment(node *sitter.Node, source string) *types.AssignmentNode {
	var name string
	var values []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_name" {
			name = child.Content([]byte(source))
		} else if child.Type() == "array" {
			// Extract array elements
			for j := 0; j < int(child.ChildCount()); j++ {
				arrayChild := child.Child(j)
				if arrayChild.Type() == "word" || arrayChild.Type() == "string" {
					content := arrayChild.Content([]byte(source))
					// Remove quotes if present
					if strings.HasPrefix(content, "\"") || strings.HasPrefix(content, "'") {
						content = content[1 : len(content)-1]
					}
					values = append(values, content)
				}
			}
		}
	}

	if name == "" {
		return nil
	}

	// For now, return AssignmentNode with array notation
	// The engine will need to handle this specially
	arrayNotation := "(" + strings.Join(values, " ") + ")"
	return &types.AssignmentNode{
		Pos:   types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Name:  name,
		Value: arrayNotation,
	}
}

// parseRedirectedStatement parses a command with redirection
func (p *Parser) parseRedirectedStatement(node *sitter.Node, source string) *types.CommandNode {
	var cmd *types.CommandNode
	var redirect *types.RedirectNode

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command":
			cmd = p.parseCommandNode(child, source)
		case "file_redirect", "heredoc_redirect", "appended_file_redirect":
			redirect = p.parseFileRedirect(child, source)
		}
	}

	if cmd == nil {
		return nil
	}

	cmd.Redirect = redirect
	return cmd
}

// parseDoGroup parses a do group (used in for/while loops)
func (p *Parser) parseDoGroup(node *sitter.Node, source string) *types.ScriptNode {
	script := &types.ScriptNode{
		Pos: types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command":
			cmd := p.parseCommandNode(child, source)
			if cmd != nil {
				script.Nodes = append(script.Nodes, cmd)
			}
		case "pipeline":
			pipe := p.parsePipeline(child, source)
			if pipe != nil {
				script.Nodes = append(script.Nodes, pipe)
			}
		}
	}

	return script
}

// parseElseClause parses an else clause
func (p *Parser) parseElseClause(node *sitter.Node, source string) *types.ScriptNode {
	script := &types.ScriptNode{
		Pos: types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command":
			cmd := p.parseCommandNode(child, source)
			if cmd != nil {
				script.Nodes = append(script.Nodes, cmd)
			}
		case "pipeline":
			pipe := p.parsePipeline(child, source)
			if pipe != nil {
				script.Nodes = append(script.Nodes, pipe)
			}
		}
	}

	return script
}

// parseTestCommand parses a test command [ ... ]
// Test commands have special syntax: [ expression ]
// The '[' is the command name and the rest are arguments
func (p *Parser) parseTestCommand(node *sitter.Node, source string) *types.CommandNode {
	var name string
	var args []string

	// Debug: Log what we're parsing
	// fmt.Printf("[DEBUG] parseTestCommand called with %d children\n", node.ChildCount())

	// Process children to find command name and arguments
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}

		nodeType := child.Type()
		content := child.Content([]byte(source))

		// Debug: Log each child
		// fmt.Printf("[DEBUG]   Child %d: type=%s, content=%q\n", i, nodeType, content)

		switch nodeType {
		case "[":
			// '[' is the command name
			name = "["
		case "]":
			// Skip closing bracket
			continue
		case "word", "string", "number":
			args = append(args, content)
		case "binary_expression", "unary_expression", "logical_expression":
			// For complex expressions, just add the content as a single argument
			// The shell will evaluate this expression
			if content != "" {
				args = append(args, content)
			}
		default:
			// For other node types, try to add content if it's not empty
			if content != "" && nodeType != ";" && nodeType != "test_operator" {
				args = append(args, content)
			}
		}
	}

	// Debug: Log result
	// fmt.Printf("[DEBUG] parseTestCommand result: name=%q, args=%v\n", name, args)

	if name == "" {
		return nil
	}

	return &types.CommandNode{
		Pos:  types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Name: name,
		Args: args,
	}
}

// parseCommandSubstitution parses a command substitution $(...)
func (p *Parser) parseCommandSubstitution(node *sitter.Node, source string) *types.CommandSubstitutionNode {
	script := &types.ScriptNode{
		Pos: types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "command" {
			cmd := p.parseCommandNode(child, source)
			if cmd != nil {
				script.Nodes = append(script.Nodes, cmd)
			}
		}
	}

	return &types.CommandSubstitutionNode{
		Pos:     types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Command: script,
	}
}

// parseFileRedirect parses a file redirection
func (p *Parser) parseFileRedirect(node *sitter.Node, source string) *types.RedirectNode {
	redirect := &types.RedirectNode{
		Pos: types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Fd:  1, // Default to stdout
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case ">", ">>", "<", ">&", "<&":
			redirect.Op = child.Type()
		case "file_descriptor":
			// Parse file descriptor (e.g., "2" in "2>&1")
			fdStr := child.Content([]byte(source))
			if fdStr != "" {
				// Simple fd parsing
				if fdStr == "0" {
					redirect.Fd = 0
				} else {
					if fdStr == "1" {
						redirect.Fd = 1
					} else {
						if fdStr == "2" {
							redirect.Fd = 2
						}
					}
				}
			}
		case "word", "string":
			redirect.File = child.Content([]byte(source))
		}
	}

	return redirect
}

// parseList parses a list with logical operators (&&, ||)
func (p *Parser) parseList(node *sitter.Node, source string) types.Node {
	var left, right types.Node
	var operator string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "&&":
			operator = "&&"
		case "||":
			operator = "||"
		case "command":
			if left == nil {
				left = p.parseCommandNode(child, source)
			} else if right == nil {
				right = p.parseCommandNode(child, source)
			}
		case "pipeline":
			if left == nil {
				left = p.parsePipeline(child, source)
			} else if right == nil {
				right = p.parsePipeline(child, source)
			}
		}
	}

	if left == nil || right == nil {
		return nil
	}

	pos := types.Position{
		Line:   int(node.StartPoint().Row) + 1,
		Column: int(node.StartPoint().Column) + 1,
		Offset: int(node.StartByte()),
	}

	if operator == "&&" {
		return &types.AndNode{
			Pos:   pos,
			Left:  left,
			Right: right,
		}
	} else if operator == "||" {
		return &types.OrNode{
			Pos:   pos,
			Left:  left,
			Right: right,
		}
	}

	return left
}

// parseListWithBackground parses a command with background job &
func (p *Parser) parseListWithBackground(node *sitter.Node, source string) types.Node {
	var childNode types.Node
	var hasBackground bool

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "&":
			hasBackground = true
		case "command":
			childNode = p.parseCommandNode(child, source)
		case "pipeline":
			childNode = p.parsePipeline(child, source)
		}
	}

	if childNode == nil {
		return nil
	}

	if hasBackground {
		return &types.BackgroundNode{
			Pos: types.Position{
				Line:   int(node.StartPoint().Row) + 1,
				Column: int(node.StartPoint().Column) + 1,
				Offset: int(node.StartByte()),
			},
			Command: childNode,
		}
	}

	return childNode
}

// parseHeredoc parses a heredoc
func (p *Parser) parseHeredoc(node *sitter.Node, source string) *types.HeredocNode {
	var commandNode types.Node
	var start, body, end string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command":
			commandNode = p.parseCommandNode(child, source)
		case "heredoc_redirect":
			for j := 0; j < int(child.ChildCount()); j++ {
				heredocChild := child.Child(j)
				switch heredocChild.Type() {
				case "heredoc_start":
					start = heredocChild.Content([]byte(source))
				case "heredoc_body":
					body = heredocChild.Content([]byte(source))
				case "heredoc_end":
					end = heredocChild.Content([]byte(source))
				}
			}
		}
	}

	if commandNode == nil {
		return nil
	}

	return &types.HeredocNode{
		Pos:     types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Command: commandNode,
		Start:   start,
		Body:    body,
		End:     end,
	}
}

// parseHeredocFromRedirected parses a heredoc from a redirected_statement node
func (p *Parser) parseHeredocFromRedirected(node *sitter.Node, source string) *types.HeredocNode {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "heredoc_redirect" {
			var commandNode types.Node
			var start, body, end string

			// Find the command (previous sibling)
			if i > 0 {
				prevChild := node.Child(i - 1)
				if prevChild.Type() == "command" {
					commandNode = p.parseCommandNode(prevChild, source)
				}
			}

			// Parse heredoc parts
			for j := 0; j < int(child.ChildCount()); j++ {
				heredocChild := child.Child(j)
				switch heredocChild.Type() {
				case "heredoc_start":
					start = heredocChild.Content([]byte(source))
				case "heredoc_body":
					body = heredocChild.Content([]byte(source))
				case "heredoc_end":
					end = heredocChild.Content([]byte(source))
				}
			}

			if commandNode != nil {
				return &types.HeredocNode{
					Pos:     types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
					Command: commandNode,
					Start:   start,
					Body:    body,
					End:     end,
				}
			}
		}
	}
	return nil
}

// parseBackgroundCommand parses a command with background job (&)
func (p *Parser) parseBackgroundCommand(node *sitter.Node, source string) *types.BackgroundNode {
	// node is a command node, parse it directly
	commandNode := p.parseCommandNode(node, source)

	if commandNode == nil {
		return nil
	}

	return &types.BackgroundNode{
		Pos: types.Position{
			Line:   int(node.StartPoint().Row) + 1,
			Column: int(node.StartPoint().Column) + 1,
			Offset: int(node.StartByte()),
		},
		Command: commandNode,
	}
}

// isBackgroundCommand checks if a command node has a background job (&)
func (p *Parser) isBackgroundCommand(node *sitter.Node) bool {
	for i := 0; i < int(node.ChildCount()); i++ {
		if node.Child(i).Type() == "&" {
			return true
		}
	}
	return false
}

// parseFunctionDefinition parses a function definition
func (p *Parser) parseFunctionDefinition(node *sitter.Node, source string) *types.FunctionNode {
	var name string
	var bodyScript *types.ScriptNode

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "word":
			name = child.Content([]byte(source))
		case "compound_statement", "subshell":
			bodyScript = p.parseCompoundStatement(child, source)
		}
	}

	if name == "" {
		return nil
	}

	if bodyScript == nil {
		bodyScript = &types.ScriptNode{
			Pos: types.Position{
				Line:   int(node.StartPoint().Row) + 1,
				Column: int(node.StartPoint().Column) + 1,
				Offset: int(node.StartByte()),
			},
			Nodes: []types.Node{},
		}
	}

	return &types.FunctionNode{
		Pos:  types.Position{Line: int(node.StartPoint().Row) + 1, Column: int(node.StartPoint().Column) + 1, Offset: int(node.StartByte())},
		Name: name,
		Body: bodyScript,
	}
}

// parseCompoundStatement parses a compound statement block
func (p *Parser) parseCompoundStatement(node *sitter.Node, source string) *types.ScriptNode {
	script := &types.ScriptNode{
		Pos: types.Position{
			Line:   int(node.StartPoint().Row) + 1,
			Column: int(node.StartPoint().Column) + 1,
			Offset: int(node.StartByte()),
		},
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "command":
			cmd := p.parseCommandNode(child, source)
			if cmd != nil {
				script.Nodes = append(script.Nodes, cmd)
			}
		case "pipeline":
			pipe := p.parsePipeline(child, source)
			if pipe != nil {
				script.Nodes = append(script.Nodes, pipe)
			}
		case "list":
			listNode := p.parseList(child, source)
			if listNode != nil {
				script.Nodes = append(script.Nodes, listNode)
			}
		}
	}

	return script
}
