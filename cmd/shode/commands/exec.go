package commands

import (
	"fmt"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
	"github.com/spf13/cobra"
)

// NewExecCommand creates the 'exec' command for executing inline commands
func NewExecCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [command]",
		Short: "Execute an inline shell command",
		Long: `Execute runs a single shell command with Shode's security features.
The command will be parsed, analyzed for security risks, and executed in a sandboxed environment.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := args[0]
			if len(args) > 1 {
				// Join all arguments to form the complete command
				for _, arg := range args[1:] {
					command += " " + arg
				}
			}

			fmt.Printf("Executing command: %s\n", command)
			
			// Parse the command
			parser := parser.NewSimpleParser()
			script, err := parser.ParseString(command)
			if err != nil {
				return fmt.Errorf("failed to parse command: %v", err)
			}
			
			if len(script.Nodes) == 0 {
				return fmt.Errorf("no valid commands found")
			}
			
			fmt.Printf("Parsed %d command(s) successfully\n", len(script.Nodes))
			
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
