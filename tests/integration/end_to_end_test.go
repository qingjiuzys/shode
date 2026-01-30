package integration

import (
	"context"
	"os"
	"testing"
	
	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

// TestEndToEndScriptExecution 测试完整的脚本执行流程
func TestEndToEndScriptExecution(t *testing.T) {
	// 创建执行环境
	stdLib := stdlib.New()
	modMgr := module.NewModuleManager()
	sb := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(envMgr, stdLib, modMgr, sb)
	
	// 测试脚本
	script := `
		# 变量赋值
		NAME="Alice"
		AGE=25
		
		# 输出
		echo "Name: $NAME"
		echo "Age: $AGE"
		
		# 数组
		FRUITS=(apple banana cherry)
		echo "${FRUITS[0]}"
		
		# 管道
		echo "hello" | cat
	`
	
	sp := parser.NewSimpleParser()
	parsed, err := sp.ParseString(script)
	if err != nil {
		t.Fatalf("ParseString() error = %v", err)
	}
	
	// 执行脚本
	result, err := ee.Execute(context.Background(), parsed)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	
	// 验证环境变量
	if envMgr.GetEnv("NAME") != "Alice" {
		t.Errorf("NAME = %s, want Alice", envMgr.GetEnv("NAME"))
	}
	
	if envMgr.GetEnv("AGE") != "25" {
		t.Errorf("AGE = %s, want 25", envMgr.GetEnv("AGE"))
	}
}

// TestEndToEndFileOperations 测试文件操作端到端流程
func TestEndToEndFileOperations(t *testing.T) {
	// 创建临时文件
	tmpfile := "/tmp/test_shode_" + os.Args[0] + ".txt"
	content := "Hello, Shode!"
	
	stdLib := stdlib.New()
	
	// 写入文件
	err := stdLib.WriteFile(tmpfile, content)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	
	// 验证文件存在
	if !stdLib.FileExists(tmpfile) {
		t.Error("File should exist after WriteFile()")
	}
	
	// 读取文件
	readContent, err := stdLib.ReadFile(tmpfile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	
	if readContent != content {
		t.Errorf("ReadFile() = %s, want %s", readContent, content)
	}
	
	// 清理
	os.Remove(tmpfile)
}

// TestEndToEndCacheOperations 测试缓存操作集成
func TestEndToEndCacheOperations(t *testing.T) {
	stdLib := stdlib.New()
	
	// 设置缓存
	stdLib.SetCache("user:123", `{"name":"Alice","age":25}`, 300)
	
	// 检查缓存存在
	if !stdLib.CacheExists("user:123") {
		t.Error("Cache key should exist")
	}
	
	// 获取缓存
	value, exists := stdLib.GetCache("user:123")
	if !exists {
		t.Fatal("GetCache() should return true")
	}
	
	if value != `{"name":"Alice","age":25}` {
		t.Errorf("GetCache() = %s, want expected value", value)
	}
	
	// 删除缓存
	stdLib.DeleteCache("user:123")
	
	if stdLib.CacheExists("user:123") {
		t.Error("Cache key should not exist after DeleteCache()")
	}
}

// TestEndToEndStringOperations 测试字符串操作集成
func TestEndToEndStringOperations(t *testing.T) {
	stdLib := stdlib.New()
	
	// ToUpper
	if stdLib.ToUpper("hello") != "HELLO" {
		t.Error("ToUpper() failed")
	}
	
	// ToLower
	if stdLib.ToLower("WORLD") != "world" {
		t.Error("ToLower() failed")
	}
	
	// Trim
	if stdLib.Trim("  test  ") != "test" {
		t.Error("Trim() failed")
	}
	
	// Replace
	if stdLib.Replace("hello world", "world", "there") != "hello there" {
		t.Error("Replace() failed")
	}
	
	// Contains
	if !stdLib.Contains("hello world", "world") {
		t.Error("Contains() should return true")
	}
	
	// SHA256Hash
	hash := stdLib.SHA256Hash("test")
	expectedHash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
	if hash != expectedHash {
		t.Errorf("SHA256Hash() = %s, want %s", hash, expectedHash)
	}
}
