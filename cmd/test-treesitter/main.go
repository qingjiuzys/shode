package main

import (
	"fmt"
	"github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/bash"
)

func main() {
	language := bash.GetLanguage()

	testCases := []string{
		`echo "hello" | cat`,
		`ls -la | grep "test" | wc -l`,
		`echo "output" > file.txt`,
		`if test -f file.txt; then echo "exists"; fi`,
		`for i in 1 2 3; do echo $i; done`,
		`count=0; while [ $count -lt 5 ]; do count=$((count+1)); done`,
	}

	for _, testCase := range testCases {
		fmt.Println("\n========================================")
		fmt.Printf("Testing: %s\n", testCase)
		fmt.Println("========================================")

		parser := sitter.NewParser()
		parser.SetLanguage(language)

		tree := parser.Parse(nil, []byte(testCase))
		if tree == nil {
			fmt.Printf("Error parsing: failed to create tree\n")
			continue
		}
		defer tree.Close()

		printNode(tree.RootNode(), 0, testCase)
	}
}

func printNode(node *sitter.Node, depth int, source string) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	content := node.Content([]byte(source))
	if len(content) > 50 {
		content = content[:50] + "..."
	}

	fmt.Printf("%s%s: %q\n", indent, node.Type(), content)

	for i := 0; i < int(node.ChildCount()); i++ {
		printNode(node.Child(i), depth+1, source)
	}
}
