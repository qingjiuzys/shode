// Package encodingutil 提供编码转换工具
package encodingutil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
)

// Base64Encode Base64编码
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode Base64解码
func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

// Base64URLEncode Base64 URL安全编码
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode Base64 URL安全解码
func Base64URLDecode(data string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(data)
}

// Base64RawURLEncode Base64 Raw URL安全编码（无填充）
func Base64RawURLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

// Base64RawURLDecode Base64 Raw URL安全解码（无填充）
func Base64RawURLDecode(data string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(data)
}

// HexEncode 十六进制编码
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode 十六进制解码
func HexDecode(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

// HexToUpper 十六进制编码（大写）
func HexToUpper(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToLower 十六进制编码（小写）
func HexToLower(data []byte) string {
	return hex.EncodeToString(data)
}

// GzipEncode Gzip压缩
func GzipEncode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GzipDecode Gzip解压
func GzipDecode(data []byte) ([]byte, error) {
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return io.ReadAll(r)
}

// ZlibEncode Zlib压缩
func ZlibEncode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ZlibDecode Zlib解压
func ZlibDecode(data []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer r.Close()

	return io.ReadAll(r)
}

// XMLEncode XML编码
func XMLEncode(v any) ([]byte, error) {
	return xml.Marshal(v)
}

// XMLEncodeIndent XML编码（带缩进）
func XMLEncodeIndent(v any) ([]byte, error) {
	return xml.MarshalIndent(v, "", "  ")
}

// XMLDecode XML解码
func XMLDecode(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

// IsValidBase64 检查是否为有效Base64
func IsValidBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// IsValidHex 检查是否为有效十六进制
func IsValidHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// IsValidBase64URL 检查是否为有效Base64 URL
func IsValidBase64URL(s string) bool {
	_, err := base64.URLEncoding.DecodeString(s)
	return err == nil
}

// EncodeBytes 编码字节
func EncodeBytes(data []byte, encoding string) (string, error) {
	switch encoding {
	case "base64":
		return Base64Encode(data), nil
	case "base64url":
		return Base64URLEncode(data), nil
	case "hex":
		return HexEncode(data), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

// DecodeBytes 解码字节
func DecodeBytes(data string, encoding string) ([]byte, error) {
	switch encoding {
	case "base64":
		return Base64Decode(data)
	case "base64url":
		return Base64URLDecode(data)
	case "hex":
		return HexDecode(data)
	default:
		return nil, fmt.Errorf("unsupported encoding: %s", encoding)
	}
}

// Compress 压缩数据
func Compress(data []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "gzip":
		return GzipEncode(data)
	case "zlib":
		return ZlibEncode(data)
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

// Decompress 解压数据
func Decompress(data []byte, algorithm string) ([]byte, error) {
	switch algorithm {
	case "gzip":
		return GzipDecode(data)
	case "zlib":
		return ZlibDecode(data)
	default:
		return nil, fmt.Errorf("unsupported compression algorithm: %s", algorithm)
	}
}

// Base64EncodeString 编码字符串
func Base64EncodeString(s string) string {
	return Base64Encode([]byte(s))
}

// Base64DecodeString 解码字符串
func Base64DecodeString(s string) (string, error) {
	data, err := Base64Decode(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HexEncodeString 编码字符串
func HexEncodeString(s string) string {
	return HexEncode([]byte(s))
}

// HexDecodeString 解码字符串
func HexDecodeString(s string) (string, error) {
	data, err := HexDecode(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GzipEncodeString 压缩字符串
func GzipEncodeString(s string) ([]byte, error) {
	return GzipEncode([]byte(s))
}

// GzipDecodeString 解压字符串
func GzipDecodeString(data []byte) (string, error) {
	decoded, err := GzipDecode(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// ZlibEncodeString 压缩字符串
func ZlibEncodeString(s string) ([]byte, error) {
	return ZlibEncode([]byte(s))
}

// ZlibDecodeString 解压字符串
func ZlibDecodeString(data []byte) (string, error) {
	decoded, err := ZlibDecode(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// ChunkedEncode 分块编码
func ChunkedEncode(data []byte, chunkSize int, encoder func([]byte) string) []string {
	if chunkSize <= 0 {
		chunkSize = 1024
	}

	var chunks []string
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		chunks = append(chunks, encoder(data[i:end]))
	}

	return chunks
}

// ChunkedDecode 分块解码
func ChunkedDecode(chunks []string, decoder func(string) ([]byte, error)) ([]byte, error) {
	var result []byte

	for _, chunk := range chunks {
		data, err := decoder(chunk)
		if err != nil {
			return nil, err
		}
		result = append(result, data...)
	}

	return result, nil
}

// Base64ChunkedEncode Base64分块编码
func Base64ChunkedEncode(data []byte, chunkSize int) []string {
	return ChunkedEncode(data, chunkSize, Base64Encode)
}

// Base64ChunkedDecode Base64分块解码
func Base64ChunkedDecode(chunks []string) ([]byte, error) {
	return ChunkedDecode(chunks, Base64Decode)
}

// HexChunkedEncode 十六进制分块编码
func HexChunkedEncode(data []byte, chunkSize int) []string {
	return ChunkedEncode(data, chunkSize, HexEncode)
}

// HexChunkedDecode 十六进制分块解码
func HexChunkedDecode(chunks []string) ([]byte, error) {
	return ChunkedDecode(chunks, HexDecode)
}

// BinaryStringToBytes 二进制字符串转字节
func BinaryStringToBytes(binary string) ([]byte, error) {
	if len(binary)%8 != 0 {
		return nil, fmt.Errorf("binary string length must be multiple of 8")
	}

	result := make([]byte, len(binary)/8)
	for i := 0; i < len(binary); i += 8 {
		var b byte
		for j := 0; j < 8; j++ {
			b <<= 1
			if binary[i+j] == '1' {
				b |= 1
			}
		}
		result[i/8] = b
	}

	return result, nil
}

// BytesToBinaryString 字节转二进制字符串
func BytesToBinaryString(data []byte) string {
	var result string
	for _, b := range data {
		for i := 7; i >= 0; i-- {
			if (b>>i)&1 == 1 {
				result += "1"
			} else {
				result += "0"
			}
		}
	}
	return result
}

// Rot13Encoder Rot13编码器
func Rot13Encode(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
			// Rot13
			if b >= 'a' && b <= 'z' {
				result[i] = 'a' + (b-'a'+13)%26
			} else {
				result[i] = 'A' + (b-'A'+13)%26
			}
		} else {
			result[i] = b
		}
	}
	return result
}

// Rot13Decoder Rot13解码器（与编码相同）
func Rot13Decode(data []byte) []byte {
	return Rot13Encode(data)
}

// ReverseBits 反转位的顺序
func ReverseBits(b byte) byte {
	var result byte
	for i := 0; i < 8; i++ {
		result <<= 1
		result |= (b >> i) & 1
	}
	return result
}

// ReverseBytes 反转字节顺序
func ReverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = b
	}
	return result
}

// XOREncode XOR编码
func XOREncode(data []byte, key []byte) []byte {
	if len(key) == 0 {
		return data
	}

	result := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		result[i] = data[i] ^ key[i%len(key)]
	}

	return result
}

// XORDecode XOR解码（与编码相同）
func XORDecode(data []byte, key []byte) []byte {
	return XOREncode(data, key)
}

// ByteToBits 字节转位数组
func ByteToBits(b byte) [8]bool {
	var bits [8]bool
	for i := 0; i < 8; i++ {
		bits[7-i] = (b>>i)&1 == 1
	}
	return bits
}

// BitsToByte 位数组转字节
func BitsToByte(bits [8]bool) byte {
	var b byte
	for i := 0; i < 8; i++ {
		if bits[7-i] {
			b |= 1 << i
		}
	}
	return b
}

// PaddingPKCS7 PKCS7填充
func PaddingPKCS7(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	if padding == 0 {
		padding = blockSize
	}

	padded := make([]byte, len(data)+padding)
	copy(padded, data)

	for i := len(data); i < len(padded); i++ {
		padded[i] = byte(padding)
	}

	return padded
}

// UnpaddingPKCS7 PKCS7去填充
func UnpaddingPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data is empty")
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding > 255 {
		return nil, fmt.Errorf("invalid padding")
	}

	for i := len(data) - padding; i < len(data); i++ {
		if int(data[i]) != padding {
			return nil, fmt.Errorf("invalid padding")
		}
	}

	return data[:len(data)-padding], nil
}
