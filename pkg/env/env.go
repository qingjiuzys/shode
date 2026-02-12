// Package env 提供环境变量管理功能
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get 获取环境变量
func Get(key string) string {
	return os.Getenv(key)
}

// GetWithDefault 获取环境变量（带默认值）
func GetWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Set 设置环境变量
func Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset 取消环境变量
func Unset(key string) error {
	return os.Unsetenv(key)
}

// Exists 检查环境变量是否存在
func Exists(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// GetAll 获取所有环境变量
func GetAll() map[string]string {
	env := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			env[pair[0]] = pair[1]
		}
	}
	return env
}

// Clear 清空所有环境变量
func Clear() {
	os.Clearenv()
}

// GetInt 获取整数环境变量
func GetInt(key string) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("environment variable %s not set", key)
	}
	return strconv.Atoi(value)
}

// GetIntWithDefault 获取整数环境变量（带默认值）
func GetIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetBool 获取布尔环境变量
func GetBool(key string) (bool, error) {
	value := strings.ToLower(os.Getenv(key))
	switch value {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", value)
	}
}

// GetBoolWithDefault 获取布尔环境变量（带默认值）
func GetBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if b, err := GetBool(key); err == nil {
			return b
		}
	}
	return defaultValue
}

// GetFloat64 获取浮点数环境变量
func GetFloat64(key string) (float64, error) {
	value := os.Getenv(key)
	if value == "" {
		return 0, fmt.Errorf("environment variable %s not set", key)
	}
	return strconv.ParseFloat(value, 64)
}

// GetFloat64WithDefault 获取浮点数环境变量（带默认值）
func GetFloat64WithDefault(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// GetSlice 获取切片环境变量（逗号分隔）
func GetSlice(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

// GetSliceWithDefault 获取切片环境变量（带默认值）
func GetSliceWithDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// MustGet 获取环境变量（panic on error）
func MustGet(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("required environment variable %s not set", key))
}

// MustGetInt 获取整数环境变量（panic on error）
func MustGetInt(key string) int {
	value, err := GetInt(key)
	if err != nil {
		panic(err)
	}
	return value
}

// MustGetBool 获取布尔环境变量（panic on error）
func MustGetBool(key string) bool {
	value, err := GetBool(key)
	if err != nil {
		panic(err)
	}
	return value
}

// MustGetFloat64 获取浮点数环境变量（panic on error）
func MustGetFloat64(key string) float64 {
	value, err := GetFloat64(key)
	if err != nil {
		panic(err)
	}
	return value
}

// LoadFromFile 从文件加载环境变量
func LoadFromFile(filename string) error {
	// 简化实现，实际应该解析.env文件格式
	return nil
}

// Lookup 查找环境变量
func Lookup(key string) (string, bool) {
	return os.LookupEnv(key)
}

// ExpandEnv 展开环境变量
func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

// Environ 获取所有环境变量（切片格式）
func Environ() []string {
	return os.Environ()
}

// Environment 环境管理器
type Environment struct {
	prefix string
}

// NewEnvironment 创建环境管理器
func NewEnvironment(prefix string) *Environment {
	return &Environment{prefix: prefix}
}

// Get 获取环境变量（带前缀）
func (e *Environment) Get(key string) string {
	return os.Getenv(e.prefix + key)
}

// GetWithDefault 获取环境变量（带前缀和默认值）
func (e *Environment) GetWithDefault(key, defaultValue string) string {
	if value := os.Getenv(e.prefix + key); value != "" {
		return value
	}
	return defaultValue
}

// GetInt 获取整数环境变量（带前缀）
func (e *Environment) GetInt(key string) (int, error) {
	return GetInt(e.prefix + key)
}

// GetIntWithDefault 获取整数环境变量（带前缀和默认值）
func (e *Environment) GetIntWithDefault(key string, defaultValue int) int {
	return GetIntWithDefault(e.prefix+key, defaultValue)
}

// GetBool 获取布尔环境变量（带前缀）
func (e *Environment) GetBool(key string) (bool, error) {
	return GetBool(e.prefix + key)
}

// GetBoolWithDefault 获取布尔环境变量（带前缀和默认值）
func (e *Environment) GetBoolWithDefault(key string, defaultValue bool) bool {
	return GetBoolWithDefault(e.prefix+key, defaultValue)
}

// GetFloat64 获取浮点数环境变量（带前缀）
func (e *Environment) GetFloat64(key string) (float64, error) {
	return GetFloat64(e.prefix + key)
}

// GetFloat64WithDefault 获取浮点数环境变量（带前缀和默认值）
func (e *Environment) GetFloat64WithDefault(key string, defaultValue float64) float64 {
	return GetFloat64WithDefault(e.prefix+key, defaultValue)
}

// Set 设置环境变量（带前缀）
func (e *Environment) Set(key, value string) error {
	return os.Setenv(e.prefix+key, value)
}

// Prefix 获取前缀
func (e *Environment) Prefix() string {
	return e.prefix
}

// SetPrefix 设置前缀
func (e *Environment) SetPrefix(prefix string) {
	e.prefix = prefix
}

// Map 环境变量映射
type Map struct {
	data map[string]string
}

// NewMap 创建环境变量映射
func NewMap() *Map {
	return &Map{
		data: make(map[string]string),
	}
}

// Set 设置值
func (m *Map) Set(key, value string) {
	m.data[key] = value
}

// Get 获取值
func (m *Map) Get(key string) string {
	return m.data[key]
}

// GetWithDefault 获取值（带默认值）
func (m *Map) GetWithDefault(key, defaultValue string) string {
	if value, ok := m.data[key]; ok {
		return value
	}
	return defaultValue
}

// Has 检查是否存在
func (m *Map) Has(key string) bool {
	_, ok := m.data[key]
	return ok
}

// Delete 删除值
func (m *Map) Delete(key string) {
	delete(m.data, key)
}

// Keys 获取所有键
func (m *Map) Keys() []string {
	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取所有值
func (m *Map) Values() []string {
	values := make([]string, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// Size 获取大小
func (m *Map) Size() int {
	return len(m.data)
}

// Clear 清空
func (m *Map) Clear() {
	m.data = make(map[string]string)
}

// ToOS 导出到系统环境变量
func (m *Map) ToOS() error {
	for k, v := range m.data {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

// FromOS 从系统环境变量导入
func (m *Map) FromOS() {
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			m.data[pair[0]] = pair[1]
		}
	}
}

// Merge 合并环境变量映射
func (m *Map) Merge(other *Map) {
	for k, v := range other.data {
		m.data[k] = v
	}
}

// Filter 过滤环境变量
func (m *Map) Filter(predicate func(key, value string) bool) *Map {
	result := NewMap()
	for k, v := range m.data {
		if predicate(k, v) {
			result.data[k] = v
		}
	}
	return result
}

// ByPrefix 按前缀过滤
func (m *Map) ByPrefix(prefix string) *Map {
	return m.Filter(func(k, _ string) bool {
		return strings.HasPrefix(k, prefix)
	})
}

// ToMap 转换为普通map
func (m *Map) ToMap() map[string]string {
	result := make(map[string]string, len(m.data))
	for k, v := range m.data {
		result[k] = v
	}
	return result
}
