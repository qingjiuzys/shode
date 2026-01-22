package parser

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/types"
)

// SimpleParser provides basic shell command parsing without external dependencies
type SimpleParser struct{}

// NewSimpleParser creates a new simple parser
func NewSimpleParser() *SimpleParser {
	return &SimpleParser{}
}

// ParseString parses shell commands from a string
func (p *SimpleParser) ParseString(source string) (*types.ScriptNode, error) {
	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
	}

	lines := strings.Split(source, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		// Simple command parsing
		cmd := p.parseCommand(line, lineNum+1)
		if cmd != nil {
			script.Nodes = append(script.Nodes, cmd)
		}
	}

	return script, nil
}

// ParseFile parses shell commands from a file
func (p *SimpleParser) ParseFile(filename string) (*types.ScriptNode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	script := &types.ScriptNode{
		Pos: types.Position{Line: 1, Column: 1, Offset: 0},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var pendingAnnotation *types.AnnotationNode
	var inFunction bool
	var functionName string
	var functionArgs []string
	var functionBody []string
	var functionStartLine int

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if inFunction {
				functionBody = append(functionBody, line)
			}
			continue // Skip empty lines and comments
		}

		// Check if we're inside a function body
		if inFunction {
			// Check for function end (closing brace on its own line)
			if trimmed == "}" {
				// End of function - create FunctionNode
				bodyScript := &types.ScriptNode{
					Pos: types.Position{Line: functionStartLine, Column: 1, Offset: 0},
				}
				// Parse function body
				bodyParser := NewSimpleParser()
				bodyContent := strings.Join(functionBody, "\n")
				bodyParsed, err := bodyParser.ParseString(bodyContent)
				if err == nil {
					bodyScript.Nodes = bodyParsed.Nodes
				}

				fnNode := &types.FunctionNode{
					Pos: types.Position{
						Line:   functionStartLine,
						Column: 1,
						Offset: 0,
					},
					Name: functionName,
					Body: bodyScript,
				}
				script.Nodes = append(script.Nodes, fnNode)

				// Reset function state
				inFunction = false
				functionName = ""
				functionArgs = nil
				functionBody = nil
				continue
			}
			// Add line to function body
			functionBody = append(functionBody, line)
			continue
		}

		// Check for function definition: function name(args) { or function name() {
		if strings.HasPrefix(trimmed, "function ") {
			// Parse function definition
			funcDef := strings.TrimPrefix(trimmed, "function ")
			funcDef = strings.TrimSpace(funcDef)

			// Find opening brace
			braceIdx := strings.Index(funcDef, "{")
			if braceIdx == -1 {
				// Multi-line function definition - brace on next line
				// Extract function name and args
				parenIdx := strings.Index(funcDef, "(")
				if parenIdx != -1 {
					functionName = strings.TrimSpace(funcDef[:parenIdx])
					closeParenIdx := strings.Index(funcDef, ")")
					if closeParenIdx > parenIdx {
						argsStr := funcDef[parenIdx+1 : closeParenIdx]
						if argsStr != "" {
							argsStr = strings.TrimSpace(argsStr)
							functionArgs = strings.Split(argsStr, ",")
							for i := range functionArgs {
								functionArgs[i] = strings.TrimSpace(functionArgs[i])
							}
						}
					}
				} else {
					// No args: function name {
					functionName = strings.TrimSpace(funcDef)
				}
				inFunction = true
				functionStartLine = lineNum
				functionBody = []string{}
				continue
			}

			// Single-line function definition
			funcPart := strings.TrimSpace(funcDef[:braceIdx])
			parenIdx := strings.Index(funcPart, "(")
			if parenIdx != -1 {
				functionName = strings.TrimSpace(funcPart[:parenIdx])
				closeParenIdx := strings.Index(funcPart, ")")
				if closeParenIdx > parenIdx {
					argsStr := funcPart[parenIdx+1 : closeParenIdx]
					if argsStr != "" {
						argsStr = strings.TrimSpace(argsStr)
						functionArgs = strings.Split(argsStr, ",")
						for i := range functionArgs {
							functionArgs[i] = strings.TrimSpace(functionArgs[i])
						}
					}
				}
			} else {
				functionName = strings.TrimSpace(funcPart)
			}

			// Check if closing brace is on same line
			closeBraceIdx := strings.Index(trimmed, "}")
			if closeBraceIdx != -1 && closeBraceIdx > braceIdx {
				// Single-line function: function name() { }
				bodyContent := trimmed[braceIdx+1 : closeBraceIdx]
				bodyContent = strings.TrimSpace(bodyContent)
				if bodyContent != "" {
					bodyParser := NewSimpleParser()
					bodyParsed, err := bodyParser.ParseString(bodyContent)
					if err == nil {
						fnNode := &types.FunctionNode{
							Pos: types.Position{
								Line:   lineNum,
								Column: 1,
								Offset: 0,
							},
							Name: functionName,
							Body: bodyParsed,
						}
						script.Nodes = append(script.Nodes, fnNode)
					}
				}
				continue
			}

			// Multi-line function - start collecting body
			inFunction = true
			functionStartLine = lineNum
			functionBody = []string{}
			continue
		}

		// Check for annotation
		if annotation := p.parseAnnotation(trimmed, lineNum); annotation != nil {
			pendingAnnotation = annotation
			continue
		}

		// Parse command or assignment
		node := p.parseCommand(trimmed, lineNum)
		if node != nil {
			// If we have a pending annotation, associate it with the node
			if pendingAnnotation != nil {
				script.Nodes = append(script.Nodes, pendingAnnotation)
				pendingAnnotation = nil
			}
			script.Nodes = append(script.Nodes, node)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return script, nil
}

// parseCommand parses a single line into a command node or assignment node
func (p *SimpleParser) parseCommand(line string, lineNum int) types.Node {
	// Check for pipeline first (before background job)
	if pipeIdx := p.findPipeIndex(line); pipeIdx != -1 {
		// Split into left and right parts
		leftStr := strings.TrimSpace(line[:pipeIdx])
		rightStr := strings.TrimSpace(line[pipeIdx+1:])

		// Recursively parse left and right commands
		leftNode := p.parseCommand(leftStr, lineNum)
		rightNode := p.parseCommand(rightStr, lineNum)

		// Create pipe node
		return &types.PipeNode{
			Pos: types.Position{
				Line:   lineNum,
				Column: 1,
				Offset: 0,
			},
			Left:  leftNode,
			Right: rightNode,
		}
	}

	// Check for background job (command &)
	trimmed := strings.TrimSpace(line)
	if strings.HasSuffix(trimmed, "&") {
		// Remove trailing &
		cmdStr := strings.TrimSpace(trimmed[:len(trimmed)-1])
		if cmdStr != "" {
			// Parse command part
			cmdNode := p.parseCommand(cmdStr, lineNum)
			if cmdNode != nil {
				return &types.BackgroundNode{
					Pos: types.Position{
						Line:   lineNum,
						Column: 1,
						Offset: 0,
					},
					Command: cmdNode,
				}
			}
		}
	}

	// Check if this is a variable assignment (var = value)
	if assignment := p.parseAssignment(line, lineNum); assignment != nil {
		return assignment
	}

	// Check for array assignment (array=(value1 value2 value3))
	if arrayNode := p.parseArrayAssignment(line, lineNum); arrayNode != nil {
		return arrayNode
	}

	// Simple tokenization - split by spaces, handle quotes
	tokens := p.tokenize(line)
	if len(tokens) == 0 {
		return nil
	}

	cmd := &types.CommandNode{
		Pos: types.Position{
			Line:   lineNum,
			Column: 1,
			Offset: 0,
		},
		Name: tokens[0],
		Args: tokens[1:],
	}

	return cmd
}

// parseAssignment parses a variable assignment (var = value)
func (p *SimpleParser) parseAssignment(line string, lineNum int) *types.AssignmentNode {
	// Look for = sign (not inside quotes)
	equalsIndex := -1
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		} else if char == '=' && !inQuotes {
			equalsIndex = i
			break
		}
	}

	if equalsIndex == -1 {
		return nil // Not an assignment
	}

	// Extract variable name (left side of =)
	varName := strings.TrimSpace(line[:equalsIndex])
	if varName == "" {
		return nil // Invalid assignment
	}

	// Extract value (right side of =)
	value := strings.TrimSpace(line[equalsIndex+1:])

	// Remove quotes if present
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
	}

	return &types.AssignmentNode{
		Pos: types.Position{
			Line:   lineNum,
			Column: 1,
			Offset: 0,
		},
		Name:  varName,
		Value: value,
	}
}

// tokenize splits a command line into tokens, handling quotes
func (p *SimpleParser) tokenize(line string) []string {
	var tokens []string
	var currentToken strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				currentToken.WriteByte(char)
			}

		case char == ' ' && !inQuotes:
			if currentToken.Len() > 0 {
				tokens = append(tokens, currentToken.String())
				currentToken.Reset()
			}

		default:
			currentToken.WriteByte(char)
		}
	}

	if currentToken.Len() > 0 {
		tokens = append(tokens, currentToken.String())
	}

	return tokens
}

// parseArrayAssignment parses an array assignment (array=(value1 value2 value3))
func (p *SimpleParser) parseArrayAssignment(line string, lineNum int) *types.ArrayNode {
	// Look for pattern: name=(values...)
	idx := strings.Index(line, "=(")
	if idx == -1 {
		return nil
	}

	varName := strings.TrimSpace(line[:idx])
	if varName == "" {
		return nil
	}

	// Find closing )
	closeIdx := strings.LastIndex(line, ")")
	if closeIdx == -1 || closeIdx <= idx {
		return nil
	}

	// Extract values between =(
	valuesStr := line[idx+2 : closeIdx]
	values := p.tokenize(valuesStr)

	return &types.ArrayNode{
		Pos: types.Position{
			Line:   lineNum,
			Column: 1,
			Offset: 0,
		},
		Name:   varName,
		Values: values,
	}
}

// parseAnnotation parses an annotation (@AnnotationName or @AnnotationName(value))
func (p *SimpleParser) parseAnnotation(line string, lineNum int) *types.AnnotationNode {
	if !strings.HasPrefix(line, "@") {
		return nil
	}

	// Remove @
	content := strings.TrimPrefix(line, "@")

	// Check for value in parentheses
	var name, value string
	if idx := strings.Index(content, "("); idx != -1 {
		name = strings.TrimSpace(content[:idx])
		if strings.HasSuffix(content, ")") {
			value = strings.TrimSpace(content[idx+1 : len(content)-1])
		}
	} else {
		name = strings.TrimSpace(content)
	}

	if name == "" {
		return nil
	}

	return &types.AnnotationNode{
		Pos: types.Position{
			Line:   lineNum,
			Column: 1,
			Offset: 0,
		},
		Name:  name,
		Value: value,
	}
}

// findPipeIndex finds the index of a pipe symbol that is not inside quotes
func (p *SimpleParser) findPipeIndex(line string) int {
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(line); i++ {
		char := line[i]

		// Handle quote state
		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
		} else if char == '|' && !inQuotes {
			// Found a pipe outside quotes
			return i
		}
	}

	return -1
}

// DebugPrint prints debug information about parsing
func (p *SimpleParser) DebugPrint(source string) {
	fmt.Println("Simple parser debug output:")
	fmt.Println("Input:", source)

	script, err := p.ParseString(source)
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d nodes:\n", len(script.Nodes))
	for i, node := range script.Nodes {
		switch n := node.(type) {
		case *types.CommandNode:
			fmt.Printf("  %d: %s %v (line %d)\n", i+1, n.Name, n.Args, n.Pos.Line)
		case *types.AssignmentNode:
			fmt.Printf("  %d: %s = %s (line %d)\n", i+1, n.Name, n.Value, n.Pos.Line)
		case *types.AnnotationNode:
			fmt.Printf("  %d: @%s(%s) (line %d)\n", i+1, n.Name, n.Value, n.Pos.Line)
		case *types.PipeNode:
			fmt.Printf("  %d: Pipeline (line %d)\n", i+1, n.Pos.Line)
		}
	}
}
