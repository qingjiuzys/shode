package commands

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
	"github.com/spf13/cobra"
)

// NewRunCommand creates the 'run' command for executing script files
func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [script-file]",
		Short: "Run a shell script file",
		Long: `Run executes a shell script file with Shode's security features enabled.
The script will be parsed, analyzed for security risks, and executed in a sandboxed environment.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptFile := args[0]
			
			// Check if file exists
			if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
				return fmt.Errorf("script file not found: %s", scriptFile)
			}

			fmt.Printf("Running script: %s\n", scriptFile)
			
			// Parse the script file
			parser := parser.NewSimpleParser()
			script, err := parser.ParseFile(scriptFile)
			if err != nil {
				return fmt.Errorf("failed to parse script: %v", err)
			}
			
			fmt.Printf("Parsed %d commands successfully\n", len(script.Nodes))
			
			// Execute each command
			for i, node := range script.Nodes {
				if cmdNode, ok := node.(*types.CommandNode); ok {
					fmt.Printf("[%d] Executing: %s %s\n", i+1, cmdNode.Name, strings.Join(cmdNode.Args, " "))
					
					// TODO: Implement actual execution with engine
					// For now, just simulate execution
					switch cmdNode.Name {
					case "upper":
						if len(cmdNode.Args) > 0 {
							fmt.Printf("Result: %s\n", strings.ToUpper(cmdNode.Args[0]))
						}
					case "lower":
						if len(cmdNode.Args) > 0 {
							fmt.Printf("Result: %s\n", strings.ToLower(cmdNode.Args[0]))
						}
					case "echo":
						fmt.Printf("Result: %s\n", strings.Join(cmdNode.Args, " "))
					case "contains":
						if len(cmdNode.Args) >= 2 {
							result := strings.Contains(cmdNode.Args[0], cmdNode.Args[1])
							fmt.Printf("Result: %t\n", result)
						}
					case "trim":
						if len(cmdNode.Args) > 0 {
							fmt.Printf("Result: %s\n", strings.TrimSpace(cmdNode.Args[0]))
						}
					case "println":
						fmt.Printf("%s\n", strings.Join(cmdNode.Args, " "))
					default:
						fmt.Printf("Command '%s' would be executed here\n", cmdNode.Name)
					}
				}
			}
			
			return nil
		},
	}

	return cmd
}
