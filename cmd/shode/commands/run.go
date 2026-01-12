package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
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

			// Initialize execution engine components
			envManager := environment.NewEnvironmentManager()
			stdLib := stdlib.New()
			moduleMgr := module.NewModuleManager()
			security := sandbox.NewSecurityChecker()

			// Create execution engine
			executionEngine := engine.NewExecutionEngine(envManager, stdLib, moduleMgr, security)

			// Set engine factory for HTTP handlers
			stdLib.SetEngineFactory(func() interface{} {
				// Create a new engine instance for each request
				em := environment.NewEnvironmentManager()
				mm := module.NewModuleManager()
				sc := sandbox.NewSecurityChecker()
				return engine.NewExecutionEngine(em, stdLib, mm, sc)
			})

			// Execute the script with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			fmt.Println("\n--- Execution Output ---")
			result, err := executionEngine.Execute(ctx, script)
			if err != nil {
				return fmt.Errorf("execution error: %v", err)
			}

			// Display results
			fmt.Println("\n--- Execution Summary ---")
			fmt.Printf("Success: %v\n", result.Success)
			fmt.Printf("Exit Code: %d\n", result.ExitCode)
			fmt.Printf("Duration: %v\n", result.Duration)
			fmt.Printf("Commands Executed: %d\n", len(result.Commands))

			if result.Output != "" {
				fmt.Printf("\nOutput:\n%s\n", result.Output)
			}

			if result.Error != "" {
				fmt.Printf("\nErrors:\n%s\n", result.Error)
			}

			// Return error if script failed
			if !result.Success {
				return fmt.Errorf("script execution failed with exit code %d", result.ExitCode)
			}

			// Check if HTTP server is running and keep the program alive
			if stdLib.IsHTTPServerRunning() {
				fmt.Println("\n--- HTTP Server Running ---")
				fmt.Println("Server is running in the background.")
				fmt.Println("Press Ctrl+C to stop the server and exit.")

				// Set up signal handling for graceful shutdown
				sigChan := make(chan os.Signal, 1)
				signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

				// Wait for interrupt signal
				<-sigChan
				fmt.Println("\nShutting down HTTP server...")

				// Stop the HTTP server
				if err := stdLib.StopHTTPServer(); err != nil {
					fmt.Printf("Error stopping HTTP server: %v\n", err)
				} else {
					fmt.Println("HTTP server stopped successfully.")
				}
			}

			return nil
		},
	}

	return cmd
}
