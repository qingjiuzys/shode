package main

import (
	"context"
	"fmt"
	"os"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <script.sh>\n", os.Args[0])
		os.Exit(1)
	}

	scriptPath := os.Args[1]

	// Setup components
	em := environment.NewEnvironmentManager()
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	// Parse script
	p := parser.NewSimpleParser()
	script, err := p.ParseFile(scriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Shode Debug Mode ===\n\n")
	fmt.Printf("Script: %s\n", scriptPath)
	fmt.Printf("Nodes: %d\n\n", len(script.Nodes))

	// Execute with metrics
	ctx := context.Background()
	result, err := ee.Execute(ctx, script)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}

	// Print results
	fmt.Printf("=== Execution Results ===\n")
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Exit Code: %d\n", result.ExitCode)
	fmt.Printf("Commands Executed: %d\n", len(result.Commands))

	// Print metrics
	metrics := ee.GetMetrics()
	if metrics != nil {
		fmt.Printf("\n=== Performance Metrics ===\n")
		fmt.Print(metrics.Format())
	}

	os.Exit(result.ExitCode)
}
