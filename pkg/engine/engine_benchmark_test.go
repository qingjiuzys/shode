package engine

import (
	"context"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

func setupBenchmarkEngine(tb testing.TB) (*ExecutionEngine, *parser.SimpleParser) {
	tb.Helper()
	envManager := environment.NewEnvironmentManager()
	std := stdlib.New()
	moduleMgr := module.NewModuleManager()
	security := sandbox.NewSecurityChecker()
	return NewExecutionEngine(envManager, std, moduleMgr, security), parser.NewSimpleParser()
}

func BenchmarkPipelineExecution(b *testing.B) {
	engine, parser := setupBenchmarkEngine(b)
	script, err := parser.ParseString(`
		echo "alpha
beta
gamma" | grep "a" | wc -l
	`)
	if err != nil {
		b.Fatalf("failed to parse benchmark script: %v", err)
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := engine.Execute(ctx, script); err != nil {
			b.Fatalf("execution failed: %v", err)
		}
	}
}

func BenchmarkControlFlowExecution(b *testing.B) {
	engine, parser := setupBenchmarkEngine(b)
	script, err := parser.ParseString(`
		counter=0
		while [ $counter -lt 50 ]; do
			echo "item-$counter" | tr a-z A-Z > /dev/null
			counter=$((counter + 1))
		done
	`)
	if err != nil {
		b.Fatalf("failed to parse control flow script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		if _, err := engine.Execute(ctx, script); err != nil {
			cancel()
			b.Fatalf("execution failed: %v", err)
		}
		cancel()
	}
}
