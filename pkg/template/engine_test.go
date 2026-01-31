package template

import (
	"testing"
)

// TestNewEngine 测试创建引擎
func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
}

// TestParse 测试解析模板
func TestParse(t *testing.T) {
	engine := NewEngine()
	engine.AddFuncs()

	content := "Hello {{.name}}!"
	if err := engine.Parse("test", content); err != nil {
		t.Fatalf("Failed to parse template: %v", err)
	}

	result, err := engine.Execute("test", map[string]interface{}{"name": "World"})
	if err != nil {
		t.Fatalf("Failed to execute template: %v", err)
	}

	if result != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", result)
	}
}

// TestRender 测试便捷渲染方法
func TestRender(t *testing.T) {
	result, err := Render("test", "Hello {{.name}}!", map[string]interface{}{"name": "Go"})
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	if result != "Hello Go!" {
		t.Errorf("Expected 'Hello Go!', got '%s'", result)
	}
}

// TestVariables 测试变量处理
func TestVariables(t *testing.T) {
	engine := NewEngine()
	engine.AddFuncs()

	content := `{{.greeting}} {{.name}}!`
	engine.Parse("vars", content)

	data := map[string]interface{}{
		"greeting": "Hello",
		"name":     "Test",
	}

	result, err := engine.Execute("vars", data)
	if err != nil {
		t.Fatalf("Failed to execute: %v", err)
	}

	expected := "Hello Test!"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestConditionals 测试条件语句
func TestConditionals(t *testing.T) {
	engine := NewEngine()
	engine.AddFuncs()

	content := `{{if .show}}Visible{{end}}`
	engine.Parse("cond", content)

	data := map[string]interface{}{"show": true}
	// 测试 true
	result, _ := engine.Execute("cond", data)
	if !strings.Contains(result, "Visible") {
		t.Error("Expected 'Visible' when show=true")
	}

	data2 := map[string]interface{}{"show": false}
	// 测试 false
	result, _ = engine.Execute("cond", data2)
	if strings.Contains(result, "Visible") {
		t.Error("Did not expect 'Visible' when show=false")
	}
}

// TestLoops 测试循环
func TestLoops(t *testing.T) {
	engine := NewEngine()
	engine.AddFuncs()

	content := `{{range .items}}{{.}} {{end}}`
	engine.Parse("loop", content)

	result, _ := engine.Execute("loop", map[string]interface{}{
		"items": []string{"a", "b", "c"},
	})

	expected := "a b c "
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestTemplateFuncs 测试模板函数
func TestTemplateFuncs(t *testing.T) {
	engine := NewEngine()
	engine.AddFuncs()

	content := `{{upper .name}}`
	engine.Parse("funcs", content)

	result, _ := engine.Execute("funcs", map[string]interface{}{"name": "hello"})
	if result != "HELLO" {
		t.Errorf("Expected 'HELLO', got '%s'", result)
	}
}

// TestSetFunc 测试设置自定义函数
func TestSetFunc(t *testing.T) {
	engine := NewEngine()

	// 添加自定义函数
	engine.SetFunc("reverse", func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	})

	content := `{{reverse .text}}`
	engine.Parse("reverse", content)

	result, _ := engine.Execute("reverse", map[string]interface{}{"text": "abc"})
	if result != "cba" {
		t.Errorf("Expected 'cba', got '%s'", result)
	}
}
