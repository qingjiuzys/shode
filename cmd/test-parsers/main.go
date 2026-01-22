package main

import (
	"fmt"
	"gitee.com/com_818cloud/shode/pkg/parser"
)

func main() {
	// Test SimpleParser
	simpleParser := parser.NewSimpleParser()
	fmt.Println("=== Testing SimpleParser ===")

	testCases := []string{
		`echo "hello" | cat`,
		`ls -la | grep "test" | wc -l`,
		`echo "output" > file.txt`,
		`count=0; while [ $count -lt 5 ]; do count=$((count+1)); done`,
	}

	for _, tc := range testCases {
		fmt.Printf("\nInput: %s\n", tc)
		script, err := simpleParser.ParseString(tc)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Parsed %d nodes\n", len(script.Nodes))
		for i, node := range script.Nodes {
			fmt.Printf("  %d: %T\n", i, node)
		}
	}

	// Test tree-sitter Parser
	fmt.Println("\n\n=== Testing tree-sitter Parser ===")

	treeParser := parser.NewParser()
	for _, tc := range testCases {
		fmt.Printf("\nInput: %s\n", tc)
		script, err := treeParser.ParseString(tc)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Parsed %d nodes\n", len(script.Nodes))
		for i, node := range script.Nodes {
			fmt.Printf("  %d: %T\n", i, node)
		}
	}
}
