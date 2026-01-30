package benchmarks

import (
	"context"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
)

func BenchmarkEngine_SimpleCommand(b *testing.B) {
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString("echo hello")
	
	envMgr := environment.NewManager()
	sb := sandbox.NewSandbox()
	ee := engine.NewExecutionEngine(envMgr, sb, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(context.Background(), script)
	}
}

func BenchmarkEngine_Pipeline(b *testing.B) {
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString("echo hello | cat")
	
	envMgr := environment.NewManager()
	sb := sandbox.NewSandbox()
	ee := engine.NewExecutionEngine(envMgr, sb, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(context.Background(), script)
	}
}

func BenchmarkEngine_VariableAssignment(b *testing.B) {
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString("NAME=value && echo $NAME")
	
	envMgr := environment.NewManager()
	sb := sandbox.NewSandbox()
	ee := engine.NewExecutionEngine(envMgr, sb, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(context.Background(), script)
	}
}

func BenchmarkEngine_Loop(b *testing.B) {
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString("for i in 1 2 3 4 5; do echo $i; done")
	
	envMgr := environment.NewManager()
	sb := sandbox.NewSandbox()
	ee := engine.NewExecutionEngine(envMgr, sb, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(context.Background(), script)
	}
}

func BenchmarkEngine_FunctionCall(b *testing.B) {
	scriptStr := `
		function test() {
			echo "hello"
		}
		test
	`
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString(scriptStr)
	
	envMgr := environment.NewManager()
	sb := sandbox.NewSandbox()
	ee := engine.NewExecutionEngine(envMgr, sb, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ee.Execute(context.Background(), script)
	}
}
