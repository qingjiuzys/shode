// Package strutil 提供字符串处理工具
package strutil

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Empty 检查是否为空
func Empty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// NotEmpty 检查是否非空
func NotEmpty(s string) bool {
	return !Empty(s)
}

// Equal 忽略大小写比较
func Equal(a, b string) bool {
	return strings.EqualFold(a, b)
}

// Contains 包含子串（忽略大小写）
func Contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// HasPrefix 前缀匹配
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix 后缀匹配
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Trim 去除首尾空白
func Trim(s string) string {
	return strings.TrimSpace(s)
}

// TrimLeft 去除左边空白
func TrimLeft(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

// TrimRight 去除右边空白
func TrimRight(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// TrimPrefix 去除前缀
func TrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}

// TrimSuffix 去除后缀
func TrimSuffix(s, suffix string) string {
	return strings.TrimSuffix(s, suffix)
}

// Split 分割字符串
func Split(s, sep string) []string {
	if sep == "" {
		return strings.Fields(s)
	}
	return strings.Split(s, sep)
}

// SplitLines 按行分割
func SplitLines(s string) []string {
	return strings.Split(s, "\n")
}

// Join 连接字符串
func Join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}

// Repeat 重复字符串
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// Replace 替换字符串
func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// ReplaceAll 替换所有
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Index 查找子串位置
func Index(s, substr string) int {
	return strings.Index(s, substr)
}

// LastIndex 最后出现位置
func LastIndex(s, substr string) int {
	return strings.LastIndex(s, substr)
}

// Count 统计子串出现次数
func Count(s, substr string) int {
	count := 0
	pos := 0
	for {
		idx := strings.Index(s[pos:], substr)
		if idx == -1 {
			break
		}
		count++
		pos += idx + len(substr)
	}
	return count
}

// Lower 转小写
func Lower(s string) string {
	return strings.ToLower(s)
}

// Upper 转大写
func Upper(s string) string {
	return strings.ToUpper(s)
}

// Title 标题格式
func Title(s string) string {
	return strings.Title(s)
}

// SnakeCase 转蛇形命名
func SnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result = append(result, '_')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// KebabCase 转短横线命名
func KebabCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result = append(result, '-')
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// CamelCase 转驼峰命名
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

// PascalCase 转帕斯卡命名
func PascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// Reverse 反转字符串
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Truncate 截断字符串
func Truncate(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}

	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	return string(runes[:maxLen]) + "..."
}

// PadLeft 左填充
func PadLeft(s, pad string, length int) string {
	for utf8.RuneCountInString(s) < length {
		s = pad + s
	}
	return s
}

// PadRight 右填充
func PadRight(s, pad string, length int) string {
	for utf8.RuneCountInString(s) < length {
		s = s + pad
	}
	return s
}

// Center 居中
func Center(s, pad string, length int) string {
	if utf8.RuneCountInString(s) >= length {
		return s
	}

	totalPad := length - utf8.RuneCountInString(s)
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	return strings.Repeat(pad, leftPad) + s + strings.Repeat(pad, rightPad)
}

// Substring 截取子串
func Substring(s string, start, end int) string {
	runes := []rune(s)

	if start < 0 {
		start = len(runes) + start
	}
	if end < 0 {
		end = len(runes) + end
	}

	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}

	if start >= end {
		return ""
	}

	return string(runes[start:end])
}

// Left 获取左边n个字符
func Left(s string, n int) string {
	runes := []rune(s)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}

// Right 获取右边n个字符
func Right(s string, n int) string {
	runes := []rune(s)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[len(runes)-n:])
}

// Before 获取子串之前的部分
func Before(s, sep string) string {
	if idx := strings.Index(s, sep); idx != -1 {
		return s[:idx]
	}
	return s
}

// After 获取子串之后的部分
func After(s, sep string) string {
	if idx := strings.Index(s, sep); idx != -1 {
		return s[idx+len(sep):]
	}
	return ""
}

// Between 获取两个子串之间的部分
func Between(s, start, end string) string {
	startIdx := strings.Index(s, start)
	if startIdx == -1 {
		return ""
	}

	endIdx := strings.Index(s[startIdx+len(start):], end)
	if endIdx == -1 {
		return ""
	}

	return s[startIdx+len(start) : startIdx+len(start)+endIdx]
}

// Words 提取单词
func Words(s string) []string {
	return strings.Fields(s)
}

// WordCount 单词计数
func WordCount(s string) int {
	return len(strings.Fields(s))
}

// LineCount 行数
func LineCount(s string) int {
	return strings.Count(s, "\n") + 1
}

// CharCount 字符数
func CharCount(s string) int {
	return utf8.RuneCountInString(s)
}

// ByteCount 字节数
func ByteCount(s string) int {
	return len(s)
}

// Remove 移除子串
func Remove(s, remove string) string {
	return strings.ReplaceAll(s, remove, "")
}

// RemoveChars 移除指定字符
func RemoveChars(s string, chars string) string {
	var result []rune
	for _, r := range s {
		if !strings.ContainsRune(chars, r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// KeepChars 只保留指定字符
func KeepChars(s string, chars string) string {
	var result []rune
	for _, r := range s {
		if strings.ContainsRune(chars, r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// CollapseSpaces 压缩空白
func CollapseSpaces(s string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(s, " "))
}

// RemoveWhitespace 移除所有空白
func RemoveWhitespace(s string) string {
	var result []rune
	for _, r := range s {
		if !unicode.IsSpace(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// RemoveNewlines 移除换行符
func RemoveNewlines(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\n", ""), "\r", "")
}

// RemoveDuplicates 移除重复行
func RemoveDuplicates(s string) string {
	lines := strings.Split(s, "\n")
	seen := make(map[string]bool)
	var result []string

	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// Shuffle 打乱字符顺序
func Shuffle(s string) string {
	runes := []rune(s)

	// Fisher-Yates shuffle
	for i := len(runes) - 1; i > 0; i-- {
		j := 0 // 简化实现，实际应该用随机数
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// SortChars 排序字符
func SortChars(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}

// UniqueChars 去重字符
func UniqueChars(s string) string {
	seen := make(map[rune]bool)
	var result []rune

	for _, r := range s {
		if !seen[r] {
			seen[r] = true
			result = append(result, r)
		}
	}

	return string(result)
}

// Chunk 分块
func Chunk(s string, size int) []string {
	if size <= 0 {
		return []string{s}
	}

	runes := []rune(s)
	var chunks []string

	for i := 0; i < len(runes); i += size {
		end := i + size
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
	}

	return chunks
}

// Wrap 文本换行
func Wrap(s string, width int) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var result strings.Builder
	lineLen := 0

	for i, word := range words {
		if lineLen > 0 && lineLen+len(word)+1 > width {
			result.WriteString("\n")
			lineLen = 0
		} else if i > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}

// Indent 缩进
func Indent(s, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

// Dedent 去除缩进
func Dedent(s string) string {
	lines := strings.Split(s, "\n")

	// 找最小缩进
	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		indent := len(line) - len(trimmed)
		if indent < minIndent || minIndent == -1 {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	// 去除缩进
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
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

// EscapeHTML 转义HTML
func EscapeHTML(s string) string {
	return html.EscapeString(s)
}

// UnescapeHTML 反转义HTML
func UnescapeHTML(s string) string {
	return html.UnescapeString(s)
}

// Match 正则匹配
func Match(s, pattern string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// FindString 查找匹配的字符串
func FindString(s, pattern string) string {
	re := regexp.MustCompile(pattern)
	return re.FindString(s)
}

// FindAllString 查找所有匹配的字符串
func FindAllString(s, pattern string) []string {
	re := regexp.MustCompile(pattern)
	return re.FindAllString(s, -1)
}

// ReplaceStringRegex 正则替换
func ReplaceStringRegex(s, pattern, repl string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(s, repl)
}

// IsAlpha 检查是否只包含字母
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsNumeric 检查是否只包含数字
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsAlphanumeric 检查是否只包含字母数字
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ToCamel 转驼峰
func ToCamel(s string) string {
	return CamelCase(s)
}

// ToSnake 转蛇形
func ToSnake(s string) string {
	return SnakeCase(s)
}

// ToKebab 转短横线
func ToKebab(s string) string {
	return KebabCase(s)
}

// ToPascal 转帕斯卡
func ToPascal(s string) string {
	return PascalCase(s)
}

// Mask 掩码字符串
func Mask(s string, visible int) string {
	runes := []rune(s)
	length := len(runes)

	if length <= visible*2 {
		return strings.Repeat("*", length)
	}

	result := make([]rune, length)

	// 保留开头
	for i := 0; i < visible && i < length; i++ {
		result[i] = runes[i]
	}

	// 掩码中间
	for i := visible; i < length-visible; i++ {
		result[i] = '*'
	}

	// 保留结尾
	for i := length - visible; i < length; i++ {
		result[i] = runes[i]
	}

	return string(result)
}

// LevenshteinDistance 编辑距离
func LevenshteinDistance(s1, s2 string) int {
	runes1 := []rune(s1)
	runes2 := []rune(s2)

	len1 := len(runes1)
	len2 := len(runes2)

	matrix := make([]int, len1+1)
	for i := 0; i <= len1; i++ {
		matrix[i] = i
	}

	for j := 1; j <= len2; j++ {
		matrix[0] = j
		prev := j
		for i := 1; i <= len1; i++ {
			cost := 0
			if runes1[i-1] != runes2[j-1] {
				cost = 1
			}

			min := matrix[i-1] + 1 // 删除
			if temp := matrix[i] + 1; temp < min {
				min = temp // 插入
			}
			if temp := prev + cost; temp < min {
				min = temp // 替换
			}

			prev = matrix[i]
			matrix[i] = min
		}
	}

	return matrix[len1]
}

// Similarity 相似度（基于编辑距离）
func Similarity(s1, s2 string) float64 {
	dist := LevenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(dist)/float64(maxLen)
}

// Diff 简单的diff
func Diff(a, b string) []string {
	linesA := strings.Split(a, "\n")
	linesB := strings.Split(b, "\n")

	var diffs []string

	maxLen := len(linesA)
	if len(linesB) > maxLen {
		maxLen = len(linesB)
	}

	for i := 0; i < maxLen; i++ {
		var lineA, lineB string

		if i < len(linesA) {
			lineA = linesA[i]
		}
		if i < len(linesB) {
			lineB = linesB[i]
		}

		if lineA != lineB {
			diffs = append(diffs, fmt.Sprintf("Line %d: '%s' -> '%s'", i+1, lineA, lineB))
		}
	}

	return diffs
}

// Concat 连接字符串
func Concat(strs ...string) string {
	var buf bytes.Buffer
	for _, s := range strs {
		buf.WriteString(s)
	}
	return buf.String()
}

// FirstChar 获取首字符
func FirstChar(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		return string(runes[0])
	}
	return ""
}

// LastChar 获取尾字符
func LastChar(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		return string(runes[len(runes)-1])
	}
	return ""
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

// Capitalize 首字母大写
func Capitalize(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

// Uncapitalize 首字母小写
func Uncapitalize(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}
