package engine

import (
	"context"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

// setupTestEngine 创建测试用的执行引擎
func setupTestEngine(t *testing.T) *ExecutionEngine {
	envMgr := environment.NewEnvironmentManager()
	stdLib := stdlib.New()
	modMgr := module.NewModuleManager()
	sb := sandbox.NewSecurityChecker()
	
	return NewExecutionEngine(envMgr, stdLib, modMgr, sb)
}

// TestNewExecutionEngine 测试创建执行引擎
func TestNewExecutionEngine(t *testing.T) {
	ee := setupTestEngine(t)
	if ee == nil {
		t.Fatal("NewExecutionEngine() returned nil")
	}
}

// TestExecute_SimpleCommand 测试简单命令执行
func TestExecute_SimpleCommand(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, err := sp.ParseString("echo hello")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	ee := setupTestEngine(t)
	result, err := ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result == nil {
		t.Error("Execute() returned nil result")
	}
}

// TestExecute_Pipeline 测试管道执行
func TestExecute_Pipeline(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, err := sp.ParseString("echo hello | cat")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	ee := setupTestEngine(t)
	result, err := ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute() exitCode = %v, want 0", result.ExitCode)
	}
}

// TestExecute_VariableAssignment 测试变量赋值
func TestExecute_VariableAssignment(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, err := sp.ParseString("NAME=value")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	envMgr := environment.NewEnvironmentManager()
	ee := setupTestEngine(t)
	
	_, err = ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	// 验证变量已设置
	value := envMgr.GetEnv("NAME")
	if value != "value" {
		t.Errorf("NAME = %v, want %v", value, "value")
	}
}

// TestExecute_MultiCommand 测试多命令执行
func TestExecute_MultiCommand(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, err := sp.ParseString("echo hello && echo world")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	ee := setupTestEngine(t)
	result, err := ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("Execute() exitCode = %v, want 0", result.ExitCode)
	}
}

// TestExecute_ArrayAssignment 测试数组赋值
func TestExecute_ArrayAssignment(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, err := sp.ParseString("arr=(a b c)")
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	ee := setupTestEngine(t)
	_, err = ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

// TestExecute_ExitCode 测试命令退出码
func TestExecute_ExitCode(t *testing.T) {
	sp := parser.NewSimpleParser()
	script, _ := sp.ParseString("true")
	
	ee := setupTestEngine(t)
	result, err := ee.Execute(context.Background(), script)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("true exitCode = %v, want 0", result.ExitCode)
	}
}
