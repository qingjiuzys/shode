// Package config 提供配置管理系统。
//
// 配置管理特点：
//   - YAML/JSON 配置文件解析
//   - 多环境支持 (dev/staging/prod)
//   - 配置热重载
//   - 环境变量覆盖
//   - 配置验证和默认值
//   - 类型安全的配置访问
//
// 使用示例：
//
//	cfg, _ := config.Load("config.yaml")
//	port := cfg.Int("server.port", 8080)
//	db := cfg.String("database.host", "localhost")
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 配置管理器
type Config struct {
	data     map[string]interface{}
	mu       sync.RWMutex
	watchers []watcher
	env      string
	envPrefix string
}

// watcher 配置文件监听器
type watcher struct {
	path   string
	modTime time.Time
	callback func(Config)
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	cfg := &Config{
		data: make(map[string]interface{}),
		env:  getEnv(),
		envPrefix: "APP_",
	}

	if err := cfg.parse(data, ext); err != nil {
		return nil, err
	}

	// 应用环境变量覆盖
	cfg.applyEnvOverrides()

	return cfg, nil
}

// LoadFromString 从字符串加载配置
func LoadFromString(content, format string) (*Config, error) {
	cfg := &Config{
		data: make(map[string]interface{}),
		env:  getEnv(),
		envPrefix: "APP_",
	}

	if err := cfg.parse([]byte(content), format); err != nil {
		return nil, err
	}

	cfg.applyEnvOverrides()

	return cfg, nil
}

// LoadWithEnv 加载配置并指定环境
func LoadWithEnv(path, env string) (*Config, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	cfg.env = env
	return cfg, nil
}

// parse 解析配置内容
func (c *Config) parse(data []byte, format string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch format {
	case ".yaml", ".yml", "yaml", "yml":
		var raw map[string]interface{}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("failed to parse YAML: %w", err)
		}
		c.data = raw

	case ".json", "json":
		if err := json.Unmarshal(data, &c.data); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

	default:
		return errors.New("unsupported config format")
	}

	return nil
}

// applyEnvOverrides 应用环境变量覆盖
func (c *Config) applyEnvOverrides() {
	for _, envVar := range os.Environ() {
		if !strings.HasPrefix(envVar, c.envPrefix) {
			continue
		}

		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(parts[0][len(c.envPrefix):])
		key = strings.ReplaceAll(key, "_", ".")

		// 尝试解析值
		value := parseValue(parts[1])

		c.mu.Lock()
		c.set(key, value)
		c.mu.Unlock()
	}
}

// parseValue 解析环境变量值
func parseValue(s string) interface{} {
	// 尝试解析为布尔值
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// 尝试解析为数字
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// 默认为字符串
	return s
}

// Get 获取配置值
func (c *Config) Get(key string, defaultValue interface{}) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.get(key)
	if !exists {
		return defaultValue
	}
	return value
}

// String 获取字符串配置
func (c *Config) String(key string, defaultValue string) string {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case string:
			return v
		case int, int64, float64:
			return fmt.Sprintf("%v", value)
		case bool:
			return strconv.FormatBool(v)
		}
	}
	return defaultValue
}

// Int 获取整数配置
func (c *Config) Int(key string, defaultValue int) int {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// Int64 获取64位整数配置
func (c *Config) Int64(key string, defaultValue int64) int64 {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case int:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				return i
			}
		}
	}
	return defaultValue
}

// Float64 获取浮点数配置
func (c *Config) Float64(key string, defaultValue float64) float64 {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case int:
			return float64(v)
		case float64:
			return v
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return f
			}
		}
	}
	return defaultValue
}

// Bool 获取布尔配置
func (c *Config) Bool(key string, defaultValue bool) bool {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case bool:
			return v
		case string:
			if b, err := strconv.ParseBool(v); err == nil {
				return b
			}
		}
	}
	return defaultValue
}

// Duration 获取时间间隔配置
func (c *Config) Duration(key string, defaultValue time.Duration) time.Duration {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case time.Duration:
			return v
		case int:
			return time.Duration(v)
		case string:
			if d, err := time.ParseDuration(v); err == nil {
				return d
			}
		}
	}
	return defaultValue
}

// StringSlice 获取字符串数组配置
func (c *Config) StringSlice(key string, defaultValue []string) []string {
	if value := c.Get(key, nil); value != nil {
		switch v := value.(type) {
		case []string:
			return v
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result
		case string:
			return strings.Split(v, ",")
		}
	}
	return defaultValue
}

// get 内部方法：获取嵌套配置值
func (c *Config) get(key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	var current interface{} = c.data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		case map[interface{}]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	return current, true
}

// set 内部方法：设置嵌套配置值
func (c *Config) set(key string, value interface{}) {
	parts := strings.Split(key, ".")

	if len(parts) == 1 {
		c.data[parts[0]] = value
		return
	}

	// 创建嵌套结构
	current := c.data
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			// 类型不匹配，创建新 map
			newMap := make(map[string]interface{})
			current[part] = newMap
			current = newMap
		}
	}

	current[parts[len(parts)-1]] = value
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.set(key, value)
}

// Has 检查配置键是否存在
func (c *Config) Has(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.get(key)
	return exists
}

// GetAll 获取所有配置
func (c *Config) GetAll() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	c.flatten(result, c.data, "")
	return result
}

// flatten 扁平化嵌套配置
func (c *Config) flatten(result map[string]interface{}, data interface{}, prefix string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newKey := key
			if prefix != "" {
				newKey = prefix + "." + key
			}
			c.flatten(result, value, newKey)
		}
	case map[interface{}]interface{}:
		for key, value := range v {
			newKey := fmt.Sprintf("%v", key)
			if prefix != "" {
				newKey = prefix + "." + newKey
			}
			c.flatten(result, value, newKey)
		}
	default:
		result[prefix] = data
	}
}

// Env 获取当前环境
func (c *Config) Env() string {
	return c.env
}

// SetEnv 设置环境
func (c *Config) SetEnv(env string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.env = env
}

// Merge 合并另一个配置
func (c *Config) Merge(other *Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.merge(c.data, other.data)
}

// merge 内部方法：合并配置
func (c *Config) merge(target, source map[string]interface{}) {
	for key, value := range source {
		if existingValue, exists := target[key]; exists {
			// 如果两边都是 map，递归合并
			if existingMap, ok := existingValue.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					c.merge(existingMap, sourceMap)
					continue
				}
			}
		}
		target[key] = value
	}
}

// ToYAML 转换为 YAML 格式
func (c *Config) ToYAML() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return yaml.Marshal(c.data)
}

// ToJSON 转换为 JSON 格式
func (c *Config) ToJSON() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return json.MarshalIndent(c.data, "", "  ")
}

// Bind 绑定配置到结构体
func (c *Config) Bind(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("target must point to a struct")
	}

	return c.bindStruct(val, "")
}

// bindStruct 绑定配置到结构体
func (c *Config) bindStruct(val reflect.Value, prefix string) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 跳过不可导出的字段
		if !field.CanSet() {
			continue
		}

		// 获取配置键
		key := strings.ToLower(fieldType.Name)
		tag := fieldType.Tag.Get("config")
		if tag != "" {
			if tag == "-" {
				continue
			}
			key = tag
		}

		if prefix != "" {
			key = prefix + "." + key
		}

		configValue, exists := c.get(key)
		if !exists {
			continue
		}

		// 根据字段类型设置值
		if err := c.setFieldValue(field, configValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", key, err)
		}
	}

	return nil
}

// setFieldValue 设置字段值
func (c *Config) setFieldValue(field reflect.Value, value interface{}) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		s, ok := value.(string)
		if !ok {
			s = fmt.Sprintf("%v", value)
		}
		field.SetString(s)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 特殊处理 time.Duration（它是 int64 的别名）
		if field.Type().PkgPath() == "time" && field.Type().Name() == "Duration" {
			var d time.Duration
			switch v := value.(type) {
			case time.Duration:
				d = v
			case int:
				d = time.Duration(v)
			case int64:
				d = time.Duration(v)
			case float64:
				d = time.Duration(v)
			case string:
				var err error
				d, err = time.ParseDuration(v)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("cannot convert %T to duration", value)
			}
			field.Set(reflect.ValueOf(d))
			return nil
		}

		// 常规整数处理
		var i int64
		switch v := value.(type) {
		case int:
			i = int64(v)
		case int64:
			i = v
		case float64:
			i = int64(v)
		case string:
			var err error
			i, err = strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot convert %T to int", value)
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var i uint64
		switch v := value.(type) {
		case int:
			i = uint64(v)
		case uint64:
			i = v
		case float64:
			i = uint64(v)
		case string:
			var err error
			i, err = strconv.ParseUint(v, 10, 64)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot convert %T to uint", value)
		}
		field.SetUint(i)

	case reflect.Float32, reflect.Float64:
		var f float64
		switch v := value.(type) {
		case float64:
			f = v
		case int:
			f = float64(v)
		case string:
			var err error
			f, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot convert %T to float", value)
		}
		field.SetFloat(f)

	case reflect.Bool:
		var b bool
		switch v := value.(type) {
		case bool:
			b = v
		case string:
			var err error
			b, err = strconv.ParseBool(v)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("cannot convert %T to bool", value)
		}
		field.SetBool(b)

	case reflect.Slice:
		// 处理数组/切片
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice {
			field.Set(val)
		}

	case reflect.Struct:
		// 特殊处理 time.Duration
		if field.Type().String() == "time.Duration" {
			var d time.Duration
			switch v := value.(type) {
			case time.Duration:
				d = v
			case int:
				d = time.Duration(v)
			case float64:
				d = time.Duration(v)
			case string:
				var err error
				d, err = time.ParseDuration(v)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("cannot convert %T to duration", value)
			}
			field.Set(reflect.ValueOf(d))
		}

	case reflect.Interface:
		field.Set(reflect.ValueOf(value))

	case reflect.Pointer:
		if value != nil {
			ptr := reflect.New(field.Type().Elem())
			if err := c.setFieldValue(ptr.Elem(), value); err != nil {
				return err
			}
			field.Set(ptr)
		}

	default:
		return fmt.Errorf("unsupported field kind: %s", field.Kind())
	}

	return nil
}

// Validate 验证配置
func (c *Config) Validate(rules map[string]ValidationRule) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for key, rule := range rules {
		value, exists := c.get(key)
		if !exists {
			if rule.Required {
				return fmt.Errorf("required config key '%s' is missing", key)
			}
			continue
		}

		if rule.Validator != nil {
			if err := rule.Validator(value); err != nil {
				return fmt.Errorf("validation failed for '%s': %w", key, err)
			}
		}
	}

	return nil
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required   bool
	Validator func(interface{}) error
}

// getEnv 获取当前环境
func getEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = os.Getenv("ENV")
	}
	if env == "" {
		env = "development"
	}
	return env
}

// Watch 监听配置文件变化
func (c *Config) Watch(path string, callback func(Config)) error {
	file, err := os.Stat(path)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.watchers = append(c.watchers, watcher{
		path:    path,
		modTime: file.ModTime(),
		callback: callback,
	})
	c.mu.Unlock()

	return nil
}

// CheckUpdates 检查配置文件更新
func (c *Config) CheckUpdates() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, w := range c.watchers {
		file, err := os.Stat(w.path)
		if err != nil {
			continue
		}

		if file.ModTime().After(w.modTime) {
			// 重新加载配置
			data, err := os.ReadFile(w.path)
			if err != nil {
				return err
			}

			ext := strings.ToLower(filepath.Ext(w.path))
			if err := c.parse(data, ext); err != nil {
				return err
			}

			w.modTime = file.ModTime()
			c.watchers[i] = w

			// 触发回调
			if w.callback != nil {
				w.callback(*c)
			}
		}
	}

	return nil
}
