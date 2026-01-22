package main

import (
	"fmt"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func main() {
	p := parser.NewParser()

	testCases := []string{
		`if test -f file.txt; then echo "exists"; fi`,
		`for i in 1 2 3; do echo $i; done`,
		`while [ $count -lt 5 ]; do count=$((count+1)); done`,
		`echo "output" > file.txt`,
		`echo "line1" > file.txt && echo "line2" >> file.txt`,
		`count=0`,
		`x=$((1+1))`,
	}

	for _, tc := range testCases {
		fmt.Printf("\nInput: %s\n", tc)
		script, err := p.ParseString(tc)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		fmt.Printf("Parsed %d nodes:\n", len(script.Nodes))
		for i, node := range script.Nodes {
			switch n := node.(type) {
			case *types.IfNode:
				fmt.Printf("  %d: IfNode\n", i)
			case *types.ForNode:
				fmt.Printf("  %d: ForNode (var=%s)\n", i, n.Variable)
			case *types.WhileNode:
				fmt.Printf("  %d: WhileNode\n", i)
			case *types.CommandNode:
				fmt.Printf("  %d: CommandNode (%s) redirect=%v\n", i, n.Name, n.Redirect != nil)
			case *types.PipeNode:
				fmt.Printf("  %d: PipeNode\n", i)
			case *types.AssignmentNode:
				fmt.Printf("  %d: AssignmentNode (%s=%s)\n", i, n.Name, n.Value)
			default:
				fmt.Printf("  %d: Unknown type %T\n", i, node)
			}
		}
	}
}
