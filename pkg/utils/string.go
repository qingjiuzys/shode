// Package utils 提供字符串工具函数
package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"unicode"
)

// IsEmpty 检查字符串是否为空
func IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// IsNotEmpty 检查字符串是否不为空
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Truncate 截断字符串
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen] + "..."
}

// TruncateWithSuffix 截断字符串并添加后缀
func TruncateWithSuffix(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}

	suffixLen := len(suffix)
	if maxLen <= suffixLen {
		return s[:maxLen]
	}

	return s[:maxLen-suffixLen] + suffix
}

// PadLeft 左填充
func PadLeft(s string, pad string, length int) string {
	for len(s) < length {
		s = pad + s
	}
	return s
}

// PadRight 右填充
func PadRight(s string, pad string, length int) string {
	for len(s) < length {
		s = s + pad
	}
	return s
}

// CamelCase 转换为驼峰命名
func CamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	if len(words) == 0 {
		return ""
	}

	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = strings.Title(strings.ToLower(word))
		}
	}

	return strings.Join(words, "")
}

// PascalCase 转换为帕斯卡命名
func PascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// SnakeCase 转换为蛇形命名
func SnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// KebabCase 转换为短横线命名
func KebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Contains 检查字符串是否包含子串（不区分大小写）
func ContainsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// HasPrefix 检查字符串是否有指定前缀（不区分大小写）
func HasPrefixIgnoreCase(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}

// HasSuffix 检查字符串是否有指定后缀（不区分大小写）
func HasSuffixIgnoreCase(s, suffix string) bool {
	return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
}

// ReplaceAll 替换所有子串（不区分大小写）
func ReplaceAllIgnoreCase(s, old, new string) string {
	reg := regexp.MustCompile("(?i)" + regexp.QuoteMeta(old))
	return reg.ReplaceAllString(s, new)
}

// CountOccurrences 统计子串出现次数
func CountOccurrences(s, substr string) int {
	count := 0
	index := 0

	for {
		i := strings.Index(s[index:], substr)
		if i == -1 {
			break
		}
		count++
		index += i + len(substr)
	}

	return count
}

// Remove 移除所有指定的子串
func Remove(s, substr string) string {
	return strings.ReplaceAll(s, substr, "")
}

// RemoveChars 移除指定的字符
func RemoveChars(s string, chars string) string {
	filter := func(r rune) rune {
		if strings.ContainsRune(chars, r) {
			return -1
		}
		return r
	}
	return strings.Map(filter, s)
}

// KeepOnly 只保留指定的字符
func KeepOnly(s string, chars string) string {
	filter := func(r rune) rune {
		if strings.ContainsRune(chars, r) {
			return r
		}
		return -1
	}
	return strings.Map(filter, s)
}

// FirstChar 获取第一个字符
func FirstChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	return string([]rune(s)[0])
}

// LastChar 获取最后一个字符
func LastChar(s string) string {
	if len(s) == 0 {
		return ""
	}
	runes := []rune(s)
	return string(runes[len(runes)-1])
}

// Initials 获取首字母
func Initials(s string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var result string
	for _, word := range words {
		if len(word) > 0 {
			result += string([]rune(word)[0])
		}
	}

	return result
}

// Before 获取子串之前的部分
func Before(s, sep string) string {
	if i := strings.Index(s, sep); i != -1 {
		return s[:i]
	}
	return s
}

// After 获取子串之后的部分
func After(s, sep string) string {
	if i := strings.Index(s, sep); i != -1 {
		return s[i+len(sep):]
	}
	return ""
}

// BeforeLast 获取最后一个子串之前的部分
func BeforeLast(s, sep string) string {
	if i := strings.LastIndex(s, sep); i != -1 {
		return s[:i]
	}
	return s
}

// AfterLast 获取最后一个子串之后的部分
func AfterLast(s, sep string) string {
	if i := strings.LastIndex(s, sep); i != -1 {
		return s[i+len(sep):]
	}
	return ""
}

// Between 获取两个子串之间的部分
func Between(s, start, end string) string {
	startIndex := strings.Index(s, start)
	if startIndex == -1 {
		return ""
	}

	startIndex += len(start)
	endIndex := strings.Index(s[startIndex:], end)
	if endIndex == -1 {
		return ""
	}

	return s[startIndex : startIndex+endIndex]
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)

	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

// RandomAlphaString 生成随机字母字符串
func RandomAlphaString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)

	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

// RandomNumericString 生成随机数字字符串
func RandomNumericString(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)

	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

// RandomHex 生成随机十六进制字符串
func RandomHex(length int) string {
	b := make([]byte, length/2)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// Base64Encode Base64编码
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode Base64解码
func Base64Decode(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// Base64URLEncode Base64 URL编码
func Base64URLEncode(s string) string {
	return base64.URLEncoding.EncodeToString([]byte(s))
}

// Base64URLDecode Base64 URL解码
func Base64URLDecode(s string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// Mask 掩码字符串
func Mask(s string, visibleStart, visibleEnd int) string {
	runes := []rune(s)
	length := len(runes)

	if length <= visibleStart+visibleEnd {
		return s
	}

	var result bytes.Buffer

	// 保留开头
	for i := 0; i < visibleStart && i < length; i++ {
		result.WriteRune(runes[i])
	}

	// 掩码中间
	maskLength := length - visibleStart - visibleEnd
	for i := 0; i < maskLength; i++ {
		result.WriteRune('*')
	}

	// 保留结尾
	for i := length - visibleEnd; i < length; i++ {
		result.WriteRune(runes[i])
	}

	return result.String()
}

// MaskEmail 掩码邮箱
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return email
	}

	maskedUsername := string([]rune(username)[0]) + "****" + string([]rune(username)[len(username)-1])
	return maskedUsername + "@" + domain
}

// MaskPhone 掩码手机号
func MaskPhone(phone string) string {
	runes := []rune(phone)
	length := len(runes)

	if length <= 7 {
		return phone
	}

	var result bytes.Buffer

	// 保留前3位
	for i := 0; i < 3 && i < length; i++ {
		result.WriteRune(runes[i])
	}

	// 掩码中间
	for i := 3; i < length-4; i++ {
		result.WriteRune('*')
	}

	// 保留后4位
	for i := length - 4; i < length; i++ {
		result.WriteRune(runes[i])
	}

	return result.String()
}

// LevenshteinDistance 计算编辑距离
func LevenshteinDistance(a, b string) int {
	lenA := len(a)
	lenB := len(b)

	if lenA == 0 {
		return lenB
	}
	if lenB == 0 {
		return lenA
	}

	// 使用一维数组优化空间
	matrix := make([]int, lenB+1)

	for i := 0; i <= lenB; i++ {
		matrix[i] = i
	}

	for i := 1; i <= lenA; i++ {
		prev := matrix[0]
		matrix[0] = i

		for j := 1; j <= lenB; j++ {
			temp := matrix[j]

			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}

			matrix[j] = min(
				matrix[j]+1,      // 删除
				matrix[j-1]+1,    // 插入
				prev+cost,        // 替换
			)

			prev = temp
		}
	}

	return matrix[lenB]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Similarity 计算相似度（0-1）
func Similarity(a, b string) float64 {
	distance := LevenshteinDistance(a, b)
	maxLen := max(len(a), len(b))

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(distance)/float64(maxLen)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// IsAlpha 检查是否全是字母
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphanumeric 检查是否全是字母或数字
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsNumeric 检查是否全是数字
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsLower 检查是否全是小写
func IsLower(s string) bool {
	for _, r := range s {
		if !unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// IsUpper 检查是否全是大写
func IsUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// ToTitleCase 转换为标题格式
func ToTitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// Capitalize 首字母大写
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}

// Uncapitalize 首字母小写
func Uncapitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])

	return string(runes)
}

// ToLower 转换为小写（支持Unicode）
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToUpper 转换为大写（支持Unicode）
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// SwapCase 大小写互换
func SwapCase(s string) string {
	var result []rune

	for _, r := range s {
		if unicode.IsUpper(r) {
			result = append(result, unicode.ToLower(r))
		} else if unicode.IsLower(r) {
			result = append(result, unicode.ToUpper(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}

// WordCount 统计单词数量
func WordCount(s string) int {
	words := strings.Fields(s)
	return len(words)
}

// LineCount 统计行数
func LineCount(s string) int {
	return strings.Count(s, "\n") + 1
}

// Indent 缩进字符串
func Indent(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// Dedent 去除缩进
func Dedent(s string) string {
	lines := strings.Split(s, "\n")

	// 找到最小缩进
	minIndent := -1
	for _, line := range lines {
		if line == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	// 去除最小缩进
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}

// Center 居中文本
func Center(s string, width int) string {
	length := len(s)
	if length >= width {
		return s
	}

	padding := width - length
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
}

// LAlign 左对齐
func LAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// RAlign 右对齐
func RAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// WrapText 文本换行
func WrapText(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	lineLength := 0

	for i, word := range words {
		if lineLength+len(word) > width && lineLength > 0 {
			result.WriteString("\n")
			lineLength = 0
		} else if i > 0 {
			result.WriteString(" ")
			lineLength++
		}

		result.WriteString(word)
		lineLength += len(word)
	}

	return result.String()
}

// StripTags 移除HTML标签
func StripTags(s string) string {
	reg := regexp.MustCompile(`<[^>]*>`)
	return reg.ReplaceAllString(s, "")
}

// EscapeHTML 转义HTML
func EscapeHTML(s string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// UnescapeHTML 反转义HTML
func UnescapeHTML(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	result := s
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}
