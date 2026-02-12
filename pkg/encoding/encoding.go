// Package encoding 提供编码解码工具
package encoding

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// JSONEncode JSON编码
func JSONEncode(v any) ([]byte, error) {
	return json.Marshal(v)
}

// JSONEncodePretty JSON美化编码
func JSONEncodePretty(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// JSONDecode JSON解码
func JSONDecode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// JSONDecodeString JSON解码字符串
func JSONDecodeString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

// JSONEncodeToString JSON编码为字符串
func JSONEncodeToString(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// JSONDecodeFromString JSON从字符串解码
func JSONDecodeFromString(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

// IsValidJSON 检查是否为有效JSON
func IsValidJSON(data string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(data), &js) == nil
}

// JSONToMap JSON转换为map
func JSONToMap(data []byte) (map[string]any, error) {
	var result map[string]any
	err := json.Unmarshal(data, &result)
	return result, err
}

// MapToJSON map转换为JSON
func MapToJSON(m map[string]any) ([]byte, error) {
	return json.Marshal(m)
}

// JSONPrettyPrint JSON美化打印
func JSONPrettyPrint(data []byte) (string, error) {
	var buf bytes.Buffer
	err := json.Indent(&buf, data, "", "  ")
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// JSONCompact JSON压缩
func JSONCompact(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := json.Compact(&buf, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobEncode Gob编码
func GobEncode(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GobDecode Gob解码
func GobDecode(data []byte, v any) error {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	return decoder.Decode(v)
}

// RegisterGobType 注册Gob类型
func RegisterGobType(v any) {
	gob.Register(v)
}

// ToString 将任意值转换为字符串
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
	case float32, float64:
		return fmt.Sprintf("%f", val)
	case bool:
		return fmt.Sprintf("%t", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToBytes 将任意值转换为字节
func ToBytes(v any) ([]byte, error) {
	if v == nil {
		return []byte{}, nil
	}

	switch val := v.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return []byte(ToString(val)), nil
	default:
		return json.Marshal(val)
	}
}

// Base64Encode Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64解码
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// Base64URLEncode Base64 URL编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode Base64 URL解码
func Base64URLDecode(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

// HexEncode 十六进制编码
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode 十六进制解码
func HexDecode(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// IsValidHex 检查是否为有效十六进制
func IsValidHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// IsValidBase64 检查是否为有效Base64
func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// Convert 转换编码
func Convert(data []byte, fromEncoding, toEncoding string) ([]byte, error) {
	// 简化实现，实际应该使用golang.org/x/text/encoding
	return data, nil
}

// Clone 克隆对象
func Clone[T any](src T) (T, error) {
	var dst T
	data, err := GobEncode(src)
	if err != nil {
		return dst, err
	}
	err = GobDecode(data, &dst)
	return dst, err
}

// DeepCopy 深度复制
func DeepCopy(src, dst any) error {
	data, err := GobEncode(src)
	if err != nil {
		return err
	}
	return GobDecode(data, dst)
}
