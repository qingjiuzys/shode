package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngine_LoadTemplate(t *testing.T) {
	engine := NewEngine("")

	// 测试加载基础模板
	tmpl, err := engine.LoadTemplate("basic")
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	if tmpl.Name != "basic" {
		t.Errorf("模板名称不正确: got %s, want basic", tmpl.Name)
	}

	if len(tmpl.Files) == 0 {
		t.Error("模板应该包含文件")
	}
}

func TestEngine_LoadTemplate_NotExists(t *testing.T) {
	engine := NewEngine("")

	_, err := engine.LoadTemplate("nonexistent")
	if err == nil {
		t.Error("应该返回错误：模板不存在")
	}
}

func TestEngine_ListTemplates(t *testing.T) {
	engine := NewEngine("")

	templates := engine.ListTemplates()

	if len(templates) != 3 {
		t.Errorf("应该有 3 个模板，got %d", len(templates))
	}

	expected := []string{"basic", "web-service", "cli-tool"}
	for i, tmpl := range templates {
		if tmpl != expected[i] {
			t.Errorf("模板 %d 不正确: got %s, want %s", i, tmpl, expected[i])
		}
	}
}

func TestEngine_Generate(t *testing.T) {
	engine := NewEngine("")
	tmpDir := t.TempDir()

	variables := map[string]string{
		"Name":        "test-project",
		"Version":     "1.0.0",
		"Description": "Test project",
		"Port":        "8080",
	}

	err := engine.Generate("basic", variables, tmpDir)
	if err != nil {
		t.Fatalf("生成项目失败: %v", err)
	}

	// 验证文件已创建
	files := []string{
		"shode.json",
		"main.sh",
		"README.md",
	}

	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("文件未创建: %s", file)
		}
	}

	// 验证内容
	shodeJsonPath := filepath.Join(tmpDir, "shode.json")
	content, err := os.ReadFile(shodeJsonPath)
	if err != nil {
		t.Fatalf("读取 shode.json 失败: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "test-project") {
		t.Error("shode.json 应该包含项目名称")
	}
}

func TestEngine_Generate_WebService(t *testing.T) {
	engine := NewEngine("")
	tmpDir := t.TempDir()

	variables := map[string]string{
		"Name":    "test-service",
		"Version": "1.0.0",
		"Port":    "3000",
	}

	err := engine.Generate("web-service", variables, tmpDir)
	if err != nil {
		t.Fatalf("生成项目失败: %v", err)
	}

	// 验证 src/main.sh
	mainShPath := filepath.Join(tmpDir, "src", "main.sh")
	if _, err := os.Stat(mainShPath); os.IsNotExist(err) {
		t.Error("src/main.sh 未创建")
	}

	// 验证 config/app.json
	configPath := filepath.Join(tmpDir, "config", "app.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config/app.json 未创建")
	}
}

func TestEngine_parseTemplate(t *testing.T) {
	engine := NewEngine("")

	tests := []struct {
		name      string
		template  string
		variables map[string]string
		expected  string
	}{
		{
			name:     "简单替换",
			template: "Hello {{.Name}}",
			variables: map[string]string{
				"Name": "World",
			},
			expected: "Hello World",
		},
		{
			name:     "多个变量",
			template: "{{.Name}} v{{.Version}}",
			variables: map[string]string{
				"Name":    "Test",
				"Version": "1.0.0",
			},
			expected: "Test v1.0.0",
		},
		{
			name:      "无变量",
			template:  "Static text",
			variables: map[string]string{},
			expected:  "Static text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.parseTemplate(tt.template, tt.variables)
			if result != tt.expected {
				t.Errorf("结果不正确: got %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGetTemplateByName(t *testing.T) {
	tmpl, err := GetTemplateByName("basic")
	if err != nil {
		t.Fatalf("获取模板失败: %v", err)
	}

	if tmpl.Name != "basic" {
		t.Errorf("模板名称不正确: got %s, want basic", tmpl.Name)
	}
}

func TestValidateTemplateName(t *testing.T) {
	tests := []struct {
		name     string
		template string
		valid    bool
	}{
		{"basic", "basic", true},
		{"web-service", "web-service", true},
		{"cli-tool", "cli-tool", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateTemplateName(tt.template)
			if result != tt.valid {
				t.Errorf("ValidateTemplateName(%s) = %v, want %v", tt.template, result, tt.valid)
			}
		})
	}
}

func TestFormatTemplateName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Basic", "basic"},
		{"Web Service", "web-service"},
		{"CLI_Tool", "cli-tool"},
		{"  spaced  ", "spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatTemplateName(tt.input)
			if result != tt.expected {
				t.Errorf("FormatTemplateName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()
	tmpDir := t.TempDir()

	// 修改工作目录到临时目录
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	options := map[string]string{
		"version":     "2.0.0",
		"description": "Generated project",
	}

	err := gen.Generate("test-project", "basic", options)
	if err != nil {
		t.Fatalf("生成项目失败: %v", err)
	}

	// 验证目录已创建
	projectDir := filepath.Join(tmpDir, "test-project")
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Error("项目目录未创建")
	}

	// 验证文件
	shodeJsonPath := filepath.Join(projectDir, "shode.json")
	if _, err := os.Stat(shodeJsonPath); os.IsNotExist(err) {
		t.Error("shode.json 未创建")
	}
}

func TestGenerator_ListTemplates(t *testing.T) {
	gen := NewGenerator()

	templates := gen.ListTemplates()

	if len(templates) != 3 {
		t.Errorf("应该有 3 个模板，got %d", len(templates))
	}

	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("模板名称不应为空")
		}
		if tmpl.Description == "" {
			t.Error("模板描述不应为空")
		}
	}
}

func TestFormatProjectName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"My Project", "my-project"},
		{"test_project", "test-project"},
		{"  spaced  ", "spaced"},
		{"Mixed_Case", "mixed-case"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatProjectName(tt.input)
			if result != tt.expected {
				t.Errorf("FormatProjectName(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name    string
		project string
		valid   bool
	}{
		{"valid", "my-project", true},
		{"empty", "", false},
		{"invalid chars", "my/project", false},
		{"path separator", "my\\project", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.project)
			valid := err == nil
			if valid != tt.valid {
				t.Errorf("ValidateProjectName(%s) = %v, want %v", tt.project, valid, tt.valid)
			}
		})
	}
}
