package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

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

	// Memory profiling
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Execute
	start := time.Now()
	ctx := context.Background()
	result, err := ee.Execute(ctx, script)
	duration := time.Since(start)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Print profile
	fmt.Printf("=== Shode Performance Profile ===\n\n")
	fmt.Printf("Script: %s\n", scriptPath)
	fmt.Printf("Execution Time: %v\n", duration)
	fmt.Printf("Success: %v\n", result.Success)
	fmt.Printf("Exit Code: %d\n", result.ExitCode)
	
	fmt.Printf("\n=== Memory Usage ===\n")
	fmt.Printf("Allocated: %d KB\n", (m2.Alloc-m1.Alloc)/1024)
	fmt.Printf("Total Allocated: %d KB\n", m2.TotalAlloc/1024)
	fmt.Printf("Heap Objects: %d\n", m2.HeapObjects)

	// Print metrics
	metrics := ee.GetMetrics()
	if metrics != nil {
		fmt.Printf("\n=== Execution Metrics ===\n")
		fmt.Print(metrics.Format())
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(result.ExitCode)
}
