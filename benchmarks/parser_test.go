package benchmarks

import (
	"os"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/parser"
)

func BenchmarkSimpleParser_SimpleCommand(b *testing.B) {
	sp := parser.NewSimpleParser()
	script := "echo hello world"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sp.ParseString(script)
	}
}

func BenchmarkSimpleParser_Pipeline(b *testing.B) {
	sp := parser.NewSimpleParser()
	script := "echo hello | cat | wc -l"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sp.ParseString(script)
	}
}

func BenchmarkSimpleParser_VariableAssignment(b *testing.B) {
	sp := parser.NewSimpleParser()
	script := "NAME=value && echo $NAME"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sp.ParseString(script)
	}
}

func BenchmarkSimpleParser_ComplexScript(b *testing.B) {
	sp := parser.NewSimpleParser()
	script := `
		# Complex script
		NAME=value
		echo $NAME | cat
		if true; then
			echo yes
		fi
		for i in 1 2 3; do
			echo $i
		done
	`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sp.ParseString(script)
	}
}

func BenchmarkSimpleParser_FileParsing(b *testing.B) {
	sp := parser.NewSimpleParser()
	
	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "benchmark*.sh")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	content := `
		#!/usr/bin/env shode
		# Benchmark script
		echo "Starting..."
		for i in $(seq 1 100); do
			echo "Line $i"
		done
		echo "Done"
	`
	if _, err := tmpfile.WriteString(content); err != nil {
		b.Fatal(err)
	}
	tmpfile.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = sp.ParseFile(tmpfile.Name())
	}
}
