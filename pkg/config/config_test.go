package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadYAML 测试加载 YAML 配置
func TestLoadYAML(t *testing.T) {
	content := `
server:
  port: 8080
  host: "0.0.0.0"
database:
  host: "localhost"
  port: 5432
  name: "mydb"
debug: true
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 测试嵌套键访问
	if cfg.Int("server.port", 0) != 8080 {
		t.Errorf("Expected server.port=8080, got %d", cfg.Int("server.port", 0))
	}

	if cfg.String("server.host", "") != "0.0.0.0" {
		t.Errorf("Expected server.host=0.0.0.0, got %s", cfg.String("server.host", ""))
	}

	if cfg.String("database.host", "") != "localhost" {
		t.Errorf("Expected database.host=localhost, got %s", cfg.String("database.host", ""))
	}

	if cfg.Bool("debug", false) != true {
		t.Error("Expected debug=true")
	}
}

// TestLoadJSON 测试加载 JSON 配置
func TestLoadJSON(t *testing.T) {
	content := `{
	"server": {
		"port": 9090,
		"host": "127.0.0.1"
	},
	"debug": false
}`

	cfg, err := LoadFromString(content, "json")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Int("server.port", 0) != 9090 {
		t.Errorf("Expected server.port=9090, got %d", cfg.Int("server.port", 0))
	}

	if cfg.Bool("debug", true) != false {
		t.Error("Expected debug=false")
	}
}

// TestLoadFromString 测试从字符串加载配置
func TestLoadFromString(t *testing.T) {
	yamlContent := `
key1: "value1"
key2: 42
key3: true
`

	cfg, err := LoadFromString(yamlContent, "yaml")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.String("key1", "") != "value1" {
		t.Errorf("Expected key1=value1, got %s", cfg.String("key1", ""))
	}

	if cfg.Int("key2", 0) != 42 {
		t.Errorf("Expected key2=42, got %d", cfg.Int("key2", 0))
	}

	if cfg.Bool("key3", false) != true {
		t.Error("Expected key3=true")
	}
}

// TestGetWithDefault 测试使用默认值
func TestGetWithDefault(t *testing.T) {
	cfg, _ := LoadFromString("existing: value", "yaml")

	// 存在的键
	if v := cfg.Get("existing", "default"); v != "value" {
		t.Errorf("Expected 'value', got %v", v)
	}

	// 不存在的键
	if v := cfg.Get("nonexistent", "default"); v != "default" {
		t.Errorf("Expected 'default', got %v", v)
	}
}

// TestString 测试字符串类型获取
func TestString(t *testing.T) {
	cfg, _ := LoadFromString(`
string_key: "hello"
number_key: 42
bool_key: true
`, "yaml")

	if cfg.String("string_key", "default") != "hello" {
		t.Errorf("Expected 'hello', got %s", cfg.String("string_key", "default"))
	}

	// 类型转换
	if cfg.String("number_key", "default") != "42" {
		t.Errorf("Expected '42', got %s", cfg.String("number_key", "default"))
	}
}

// TestInt 测试整数类型获取
func TestInt(t *testing.T) {
	cfg, _ := LoadFromString(`
int_key: 100
float_key: 3.14
string_int: "42"
`, "yaml")

	if cfg.Int("int_key", 0) != 100 {
		t.Errorf("Expected 100, got %d", cfg.Int("int_key", 0))
	}

	// float 转换为 int
	if cfg.Int("float_key", 0) != 3 {
		t.Errorf("Expected 3, got %d", cfg.Int("float_key", 0))
	}

	// 字符串转 int
	if cfg.Int("string_int", 0) != 42 {
		t.Errorf("Expected 42, got %d", cfg.Int("string_int", 0))
	}
}

// TestFloat64 测试浮点数获取
func TestFloat64(t *testing.T) {
	cfg, _ := LoadFromString(`
float_key: 3.14159
int_key: 42
`, "yaml")

	if cfg.Float64("float_key", 0) != 3.14159 {
		t.Errorf("Expected 3.14159, got %f", cfg.Float64("float_key", 0))
	}

	// int 转换为 float
	if cfg.Float64("int_key", 0) != 42.0 {
		t.Errorf("Expected 42.0, got %f", cfg.Float64("int_key", 0))
	}
}

// TestBool 测试布尔值获取
func TestBool(t *testing.T) {
	cfg, _ := LoadFromString(`
bool_true: true
bool_false: false
string_true: "true"
string_false: "false"
int_one: 1
int_zero: 0
`, "yaml")

	if !cfg.Bool("bool_true", false) {
		t.Error("Expected bool_true=true")
	}

	if cfg.Bool("bool_false", true) {
		t.Error("Expected bool_false=false")
	}

	if !cfg.Bool("string_true", false) {
		t.Error("Expected string_true=true")
	}

	if cfg.Bool("string_false", true) {
		t.Error("Expected string_false=false")
	}
}

// TestDuration 测试时间间隔获取
func TestDuration(t *testing.T) {
	cfg, _ := LoadFromString(`
duration_string: "5s"
duration_int: 1000
`, "yaml")

	if d := cfg.Duration("duration_string", 0); d != 5*time.Second {
		t.Errorf("Expected 5s, got %v", d)
	}

	if d := cfg.Duration("duration_int", 0); d != 1000*time.Nanosecond {
		t.Errorf("Expected 1000ns, got %v", d)
	}
}

// TestStringSlice 测试字符串数组获取
func TestStringSlice(t *testing.T) {
	cfg, _ := LoadFromString(`
array: ["a", "b", "c"]
comma_separated: "x,y,z"
`, "yaml")

	arr := cfg.StringSlice("array", nil)
	if len(arr) != 3 || arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
		t.Errorf("Unexpected array: %v", arr)
	}

	// 逗号分隔的字符串
	slice := cfg.StringSlice("comma_separated", nil)
	if len(slice) != 3 || slice[0] != "x" || slice[1] != "y" || slice[2] != "z" {
		t.Errorf("Unexpected slice: %v", slice)
	}
}

// TestHas 测试键存在性检查
func TestHas(t *testing.T) {
	cfg, _ := LoadFromString(`
existing: "value"
nested:
  key: "value"
`, "yaml")

	if !cfg.Has("existing") {
		t.Error("Expected 'existing' key to exist")
	}

	if !cfg.Has("nested.key") {
		t.Error("Expected 'nested.key' to exist")
	}

	if cfg.Has("nonexistent") {
		t.Error("Expected 'nonexistent' key to not exist")
	}
}

// TestSet 测试设置配置值
func TestSet(t *testing.T) {
	cfg, _ := LoadFromString("key: value", "yaml")

	cfg.Set("key", "new_value")
	if cfg.String("key", "") != "new_value" {
		t.Errorf("Expected 'new_value', got %s", cfg.String("key", ""))
	}

	// 设置嵌套键
	cfg.Set("new.nested.key", "nested_value")
	if cfg.String("new.nested.key", "") != "nested_value" {
		t.Errorf("Expected 'nested_value', got %s", cfg.String("new.nested.key", ""))
	}
}

// TestEnv 测试环境设置
func TestEnv(t *testing.T) {
	cfg, _ := LoadFromString("key: value", "yaml")

	if cfg.Env() != "development" {
		t.Errorf("Expected env='development', got %s", cfg.Env())
	}

	cfg.SetEnv("production")
	if cfg.Env() != "production" {
		t.Errorf("Expected env='production', got %s", cfg.Env())
	}
}

// TestMerge 测试配置合并
func TestMerge(t *testing.T) {
	cfg1, _ := LoadFromString(`
key1: "value1"
key2: "value2"
nested:
  key: "original"
`, "yaml")

	cfg2, _ := LoadFromString(`
key2: "new_value2"
key3: "value3"
nested:
  key: "updated"
  new_key: "new"
`, "yaml")

	cfg1.Merge(cfg2)

	if cfg1.String("key1", "") != "value1" {
		t.Errorf("Expected key1='value1', got %s", cfg1.String("key1", ""))
	}

	if cfg1.String("key2", "") != "new_value2" {
		t.Errorf("Expected key2='new_value2', got %s", cfg1.String("key2", ""))
	}

	if cfg1.String("key3", "") != "value3" {
		t.Errorf("Expected key3='value3', got %s", cfg1.String("key3", ""))
	}

	if cfg1.String("nested.key", "") != "updated" {
		t.Errorf("Expected nested.key='updated', got %s", cfg1.String("nested.key", ""))
	}

	if cfg1.String("nested.new_key", "") != "new" {
		t.Errorf("Expected nested.new_key='new', got %s", cfg1.String("nested.new_key", ""))
	}
}

// TestBind 测试绑定到结构体
func TestBind(t *testing.T) {
	cfg, _ := LoadFromString(`
server_port: 8080
server_host: "localhost"
debug_mode: true
timeout: "30s"
`, "yaml")

	type ServerConfig struct {
		Port    int           `config:"server_port"`
		Host    string        `config:"server_host"`
		Debug   bool          `config:"debug_mode"`
		Timeout time.Duration `config:"timeout"`
	}

	var server ServerConfig
	if err := cfg.Bind(&server); err != nil {
		t.Fatalf("Failed to bind config: %v", err)
	}

	if server.Port != 8080 {
		t.Errorf("Expected Port=8080, got %d", server.Port)
	}

	if server.Host != "localhost" {
		t.Errorf("Expected Host=localhost, got %s", server.Host)
	}

	if !server.Debug {
		t.Error("Expected Debug=true")
	}

	if server.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout=30s, got %v", server.Timeout)
	}
}

// TestValidate 测试配置验证
func TestValidate(t *testing.T) {
	cfg, _ := LoadFromString(`
required_key: "value"
port: 8080
`, "yaml")

	rules := map[string]ValidationRule{
		"required_key": {
			Required: true,
		},
		"port": {
			Required: true,
			Validator: func(v interface{}) error {
				if port, ok := v.(int); !ok || port < 1024 {
					return fmt.Errorf("port must be >= 1024")
				}
				return nil
			},
		},
	}

	if err := cfg.Validate(rules); err != nil {
		t.Errorf("Validation failed: %v", err)
	}
}

// TestGetAll 测试获取所有配置
func TestGetAll(t *testing.T) {
	cfg, _ := LoadFromString(`
key1: "value1"
nested:
  key2: "value2"
`, "yaml")

	all := cfg.GetAll()

	if all["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", all["key1"])
	}

	if all["nested.key2"] != "value2" {
		t.Errorf("Expected nested.key2='value2', got %v", all["nested.key2"])
	}
}

// TestEnvOverrides 测试环境变量覆盖
func TestEnvOverrides(t *testing.T) {
	// 设置环境变量
	os.Setenv("APP_SERVER_PORT", "9090")
	os.Setenv("APP_DATABASE_HOST", "remote.db")
	defer os.Unsetenv("APP_SERVER_PORT")
	defer os.Unsetenv("APP_DATABASE_HOST")

	cfg, _ := LoadFromString(`
server:
  port: 8080
database:
  host: "localhost"
`, "yaml")

	// 环境变量应该覆盖配置文件
	if cfg.Int("server.port", 0) != 9090 {
		t.Errorf("Expected server.port=9090 (from env), got %d", cfg.Int("server.port", 0))
	}

	if cfg.String("database.host", "") != "remote.db" {
		t.Errorf("Expected database.host='remote.db' (from env), got %s", cfg.String("database.host", ""))
	}
}
