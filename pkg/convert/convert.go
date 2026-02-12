// Package convert 提供类型转换功能
package convert

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ToString 转换为字符串
func ToString(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case []byte:
		return string(val)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case time.Time:
		return val.Format(time.RFC3339)
	case time.Duration:
		return val.String()
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToInt 转换为整数
func ToInt(v any) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int8:
		return int(val), nil
	case int16:
		return int(val), nil
	case int32:
		return int(val), nil
	case int64:
		return int(val), nil
	case uint:
		return int(val), nil
	case uint8:
		return int(val), nil
	case uint16:
		return int(val), nil
	case uint32:
		return int(val), nil
	case uint64:
		return int(val), nil
	case float32:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		i, err := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
		return int(i), err
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// ToInt64 转换为int64
func ToInt64(v any) (int64, error) {
	switch val := v.(type) {
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return int64(val), nil
	case float64:
		return int64(val), nil
	case string:
		return strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	case bool:
		if val {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", v)
	}
}

// ToFloat64 转换为float64
func ToFloat64(v any) (float64, error) {
	switch val := v.(type) {
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(strings.TrimSpace(val), 64)
	case bool:
		if val {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

// ToBool 转换为布尔值
func ToBool(v any) (bool, error) {
	switch val := v.(type) {
	case bool:
		return val, nil
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(val).Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(val).Uint() != 0, nil
	case float32:
		return val != 0, nil
	case float64:
		return val != 0, nil
	case string:
		s := strings.ToLower(strings.TrimSpace(val))
		return s == "true" || s == "1" || s == "yes" || s == "on", nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", v)
	}
}

// ToIntE 转换为整数（带错误）
func ToIntE(v any) (int, error) {
	return ToInt(v)
}

// MustInt 转换为整数（panic on error）
func MustInt(v any) int {
	i, err := ToInt(v)
	if err != nil {
		panic(err)
	}
	return i
}

// ToSlice 转换为切片
func ToSlice(v any) []any {
	if v == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}

	result := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		result[i] = val.Index(i).Interface()
	}

	return result
}

// ToMap 转换为map
func ToMap(v any) (map[string]any, error) {
	if v == nil {
		return nil, nil
	}

	switch val := v.(type) {
	case map[string]any:
		return val, nil
	default:
		// 尝试JSON转换
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}

		var result map[string]any
		err = json.Unmarshal(data, &result)
		return result, err
	}
}

// StructToMap 结构体转map
func StructToMap(obj any) (map[string]any, error) {
	result := make(map[string]any)

	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", obj)
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		// 获取JSON标签
		tag := field.Tag.Get("json")
		if tag == "-" {
			continue
		}

		// 解析标签名称
		name := field.Name
		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				name = parts[0]
			}
		}

		result[name] = val.Field(i).Interface()
	}

	return result, nil
}

// MapToStruct map转结构体
func MapToStruct(m map[string]any, obj any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, obj)
}

// CopyValue 复制值
func CopyValue(src, dst any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}

// Clone 克隆对象
func Clone(v any) (any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	// 创建新对象
	newV := reflect.New(reflect.TypeOf(v))
	err = json.Unmarshal(data, newV.Interface())
	if err != nil {
		return nil, err
	}

	return newV.Elem().Interface(), nil
}

// ConvertSlice 转换切片类型
func ConvertSlice[T, U any](slice []T, converter func(T) U) []U {
	if slice == nil {
		return nil
	}

	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = converter(item)
	}

	return result
}

// MapKeys 获取map的所有键
func MapKeys(m any) ([]string, error) {
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected map, got %T", m)
	}

	keys := make([]string, 0, val.Len())
	for _, key := range val.MapKeys() {
		keys = append(keys, ToString(key.Interface()))
	}

	return keys, nil
}

// MapValues 获取map的所有值
func MapValues(m any) ([]any, error) {
	val := reflect.ValueOf(m)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("expected map, got %T", m)
	}

	values := make([]any, 0, val.Len())
	for _, key := range val.MapKeys() {
		values = append(values, val.MapIndex(key).Interface())
	}

	return values, nil
}

// MergeMaps 合并多个map
func MergeMaps(maps ...map[string]any) map[string]any {
	result := make(map[string]any)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

// DeepEqual 深度比较
func DeepEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}

// IsZero 检查是否为零值
func IsZero(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	case reflect.String:
		return val.String() == ""
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return val.IsNil()
	default:
		return false
	}
}

// GetType 获取类型名称
func GetType(v any) string {
	if v == nil {
		return "nil"
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.String()
}

// IsNil 检查是否为nil
func IsNil(v any) bool {
	if v == nil {
		return true
	}

	val := reflect.ValueOf(v)
	switch val.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return val.IsNil()
	default:
		return false
	}
}

// ToDuration 转换为时间间隔
func ToDuration(v any) (time.Duration, error) {
	switch val := v.(type) {
	case time.Duration:
		return val, nil
	case int, int8, int16, int32, int64:
		return time.Duration(reflect.ValueOf(val).Int()), nil
	case uint, uint8, uint16, uint32, uint64:
		return time.Duration(reflect.ValueOf(val).Uint()), nil
	case float32, float64:
		return time.Duration(reflect.ValueOf(val).Float()), nil
	case string:
		return time.ParseDuration(val)
	default:
		return 0, fmt.Errorf("cannot convert %T to duration", v)
	}
}

// ToTime 转换为时间
func ToTime(v any) (time.Time, error) {
	switch val := v.(type) {
	case time.Time:
		return val, nil
	case int, int8, int16, int32, int64:
		return time.Unix(reflect.ValueOf(val).Int(), 0), nil
	case string:
		// 尝试多种格式
		formats := []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02 15:04:05",
			"2006-01-02",
			time.RFC1123,
			time.RFC1123Z,
		}

		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t, nil
			}
		}

		return time.Time{}, fmt.Errorf("cannot parse time string: %s", val)
	default:
		return time.Time{}, fmt.Errorf("cannot convert %T to time", v)
	}
}

// PointerTo 获取指针
func PointerTo[T any](v T) *T {
	return &v
}

// Deref 解引用
func Deref[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}

// Coalesce 返回第一个非nil值
func Coalesce[T any](values ...*T) *T {
	for _, v := range values {
		if v != nil {
			return v
		}
	}
	return nil
}

// CoalesceValue 返回第一个非零值
func CoalesceValue[T comparable](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}
