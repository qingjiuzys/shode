// Package jsonutil 提供JSON处理工具
package jsonutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Marshal 序列化为JSON
func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalToString 序列化为JSON字符串
func MarshalToString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// MarshalIndent 序列化为格式化JSON
func MarshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// MarshalIndentToString 序列化为格式化JSON字符串
func MarshalIndentToString(v any) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Unmarshal 反序列化JSON
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// UnmarshalFromString 从字符串反序列化
func UnmarshalFromString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

// IsValidJSON 检查是否为有效JSON
func IsValidJSON(data string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(data), &js) == nil
}

// Beautify 美化JSON
func Beautify(data []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// BeautifyString 美化JSON字符串
func BeautifyString(data string) (string, error) {
	b, err := Beautify([]byte(data))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Minify 压缩JSON
func Minify(data []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Compact(&out, data)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// MinifyString 压缩JSON字符串
func MinifyString(data string) (string, error) {
	b, err := Minify([]byte(data))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Merge 合并JSON对象
func Merge(dest, src map[string]any) (map[string]any, error) {
	destData, err := json.Marshal(dest)
	if err != nil {
		return nil, err
	}

	srcData, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}

	var destMap map[string]any
	if err := json.Unmarshal(destData, &destMap); err != nil {
		return nil, err
	}

	srcMap := make(map[string]any)
	if err := json.Unmarshal(srcData, &srcMap); err != nil {
		return nil, err
	}

	for k, v := range srcMap {
		destMap[k] = v
	}

	return destMap, nil
}

// Get 获取JSON路径值
func Get(data any, path string) (any, error) {
	parts := strings.Split(path, ".")

	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("key not found: %s", part)
			}
		case []any:
			index := 0
			_, err := fmt.Sscanf(part, "[%d]", &index)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("index out of bounds: %d", index)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot access path: %s (not an object or array)", path)
		}
	}

	return current, nil
}

// Set 设置JSON路径值
func Set(data any, path string, value any) (any, error) {
	parts := strings.Split(path, ".")

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	current := data
	for _, part := range parts[:len(parts)-1] {
		switch v := current.(type) {
		case map[string]any:
			if next, ok := v[part]; ok {
				current = next
			} else {
				newObj := make(map[string]any)
				v[part] = newObj
				current = newObj
			}
		case []any:
			index := 0
			_, err := fmt.Sscanf(part, "[%d]", &index)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
			if index < 0 || index >= len(v) {
				return nil, fmt.Errorf("index out of bounds: %d", index)
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot set path: %s", path)
		}
	}

	// 设置最后一个路径的值
	lastPart := parts[len(parts)-1]
	switch v := current.(type) {
	case map[string]any:
		v[lastPart] = value
	case []any:
		index := 0
		_, err := fmt.Sscanf(lastPart, "[%d]", &index)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", lastPart)
		}
		if index < 0 || index >= len(v) {
			return nil, fmt.Errorf("index out of bounds: %d", index)
		}
		v[index] = value
	default:
		return nil, fmt.Errorf("cannot set path: %s", path)
	}

	return data, nil
}

// Delete 删除JSON路径值
func Delete(data any, path string) (any, error) {
	parts := strings.Split(path, ".")

	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	current := data
	for _, part := range parts[:len(parts)-1] {
		switch v := current.(type) {
		case map[string]any:
			if next, ok := v[part]; ok {
				current = next
			} else {
				return data, nil
			}
		case []any:
			index := 0
			_, err := fmt.Sscanf(part, "[%d]", &index)
			if err != nil {
				return nil, fmt.Errorf("invalid array index: %s", part)
			}
			if index < 0 || index >= len(v) {
				return data, nil
			}
			current = v[index]
		default:
			return data, nil
		}
	}

	// 删除最后一个路径的值
	lastPart := parts[len(parts)-1]
	switch v := current.(type) {
	case map[string]any:
		delete(v, lastPart)
	case []any:
		index := 0
		_, err := fmt.Sscanf(lastPart, "[%d]", &index)
		if err != nil {
			return nil, fmt.Errorf("invalid array index: %s", lastPart)
		}
		if index >= 0 && index < len(v) {
			v = append(v[:index], v[index+1:]...)
		}
	default:
		return data, nil
	}

	return data, nil
}

// Has 检查JSON路径是否存在
func Has(data any, path string) (bool, error) {
	_, err := Get(data, path)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetKeys 获取所有键
func GetKeys(data any) ([]string, error) {
	switch v := data.(type) {
	case map[string]any:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		return keys, nil
	default:
		return nil, fmt.Errorf("not an object")
	}
}

// GetValues 获取所有值
func GetValues(data any) ([]any, error) {
	switch v := data.(type) {
	case map[string]any:
		values := make([]any, 0, len(v))
		for _, val := range v {
			values = append(values, val)
		}
		return values, nil
	case []any:
		return v, nil
	default:
		return nil, fmt.Errorf("not an object or array")
	}
}

// Flatten 扁平化嵌套对象
func Flatten(data map[string]any) map[string]any {
	result := make(map[string]any)

	flattenHelper(data, "", result)

	return result
}

func flattenHelper(data any, prefix string, result map[string]any) {
	switch v := data.(type) {
	case map[string]any:
		for k, val := range v {
			newKey := k
			if prefix != "" {
				newKey = prefix + "." + k
			}
			flattenHelper(val, newKey, result)
		}
	case []any:
		for i, val := range v {
			newKey := fmt.Sprintf("%s[%d]", prefix, i)
			flattenHelper(val, newKey, result)
		}
	default:
		result[prefix] = data
	}
}

// Expand 扩展扁平化对象
func Expand(data map[string]any) (map[string]any, error) {
	result := make(map[string]any)

	for key, value := range data {
		parts := strings.Split(key, ".")
		current := result

		for _, part := range parts[:len(parts)-1] {
			if existing, ok := current[part]; ok {
				if obj, ok := existing.(map[string]any); ok {
					current = obj
				} else {
					newObj := make(map[string]any)
					current[part] = newObj
					current = newObj
				}
			} else {
				newObj := make(map[string]any)
				current[part] = newObj
				current = newObj
			}
		}

		lastPart := parts[len(parts)-1]
		current[lastPart] = value
	}

	return result, nil
}

// Clone 克隆JSON对象
func Clone(data any) (any, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result any
	err = json.Unmarshal(jsonData, &result)
	return result, err
}

// Convert 转换JSON类型
func Convert(data any, targetType string) (any, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	switch targetType {
	case "map[string]any":
		var result map[string]any
		err = json.Unmarshal(jsonData, &result)
		return result, err
	case "[]any":
		var result []any
		err = json.Unmarshal(jsonData, &result)
		return result, err
	case "string":
		return string(jsonData), nil
	default:
		return nil, fmt.Errorf("unsupported target type: %s", targetType)
	}
}

// PrettyPrint 美化打印JSON
func PrettyPrint(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// ToString 转换为JSON字符串
func ToString(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

// ToMap 转换为map
func ToMap(v any) (map[string]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	return result, err
}

// ToSlice 转换为切片
func ToSlice(v any) ([]any, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var result []any
	err = json.Unmarshal(data, &result)
	return result, err
}

// Select 选择字段
func Select(data map[string]any, fields ...string) map[string]any {
	result := make(map[string]any)

	for _, field := range fields {
		if value, ok := data[field]; ok {
			result[field] = value
		}
	}

	return result
}

// Omit 排除字段
func Omit(data map[string]any, fields ...string) map[string]any {
	result := make(map[string]any)

	for key, value := range data {
		shouldOmit := false
		for _, field := range fields {
			if key == field {
				shouldOmit = true
				break
			}
		}

		if !shouldOmit {
			result[key] = value
		}
	}

	return result
}

// Rename 重命名字段
func Rename(data map[string]any, oldKey, newKey string) map[string]any {
	if value, ok := data[oldKey]; ok {
		delete(data, oldKey)
		data[newKey] = value
	}
	return data
}

// MergeDeep 深度合并
func MergeDeep(dest, src map[string]any) error {
	for key, srcValue := range src {
		destValue, exists := dest[key]

		if !exists {
			dest[key] = srcValue
			continue
		}

		// 如果两边都是map，递归合并
		destMap, destOk := destValue.(map[string]any)
		srcMap, srcOk := srcValue.(map[string]any)

		if destOk && srcOk {
			if err := MergeDeep(destMap, srcMap); err != nil {
				return err
			}
			dest[key] = destMap
		} else {
			dest[key] = srcValue
		}
	}

	return nil
}

// Patch JSON Patch操作
func Patch(data map[string]any, patches map[string]any) map[string]any {
	for key, value := range patches {
		if value == nil {
			delete(data, key)
		} else {
			data[key] = value
		}
	}
	return data
}

// Diff 比较两个JSON对象的差异
func Diff(data1, data2 map[string]any) []string {
	var diffs []string

	// 检查data1中有但data2中没有的键
	for key := range data1 {
		if _, ok := data2[key]; !ok {
			diffs = append(diffs, fmt.Sprintf("removed: %s", key))
		}
	}

	// 检查data2中有但data1中没有的键
	for key := range data2 {
		if _, ok := data1[key]; !ok {
			diffs = append(diffs, fmt.Sprintf("added: %s", key))
		}
	}

	// 检查值变化
	for key := range data1 {
		if val2, ok := data2[key]; ok {
			val1 := data1[key]
			if fmt.Sprintf("%v", val1) != fmt.Sprintf("%v", val2) {
				diffs = append(diffs, fmt.Sprintf("changed: %s", key))
			}
		}
	}

	return diffs
}

// Schema JSON Schema生成器
type Schema struct {
	Type       string            `json:"type"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Required   []string          `json:"required,omitempty"`
	Items      *Schema           `json:"items,omitempty"`
}

// GenerateSchema 从对象生成Schema
func GenerateSchema(data any) (*Schema, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return nil, err
	}

	schema := &Schema{
		Type: "object",
	}

	properties := make(map[string]Schema)
	required := make([]string, 0)

	for key, value := range raw {
		var js any
		if err := json.Unmarshal(value, &js); err != nil {
			return nil, err
		}

		propSchema := Schema{}
		switch js.(type) {
		case string:
			propSchema.Type = "string"
		case float64:
			propSchema.Type = "number"
		case bool:
			propSchema.Type = "boolean"
		case map[string]any:
			propSchema.Type = "object"
		case []any:
			propSchema.Type = "array"
		default:
			propSchema.Type = "any"
		}

		properties[key] = propSchema
	}

	schema.Properties = properties
	schema.Required = required

	return schema, nil
}
