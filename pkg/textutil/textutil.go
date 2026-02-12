// Package textutil 提供文本处理工具
package textutil

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// WordCount 统计单词数
func WordCount(text string) int {
	words := strings.Fields(text)
	return len(words)
}

// LineCount 统计行数
func LineCount(text string) int {
	return strings.Count(text, "\n") + 1
}

// CharCount 统计字符数
func CharCount(text string) int {
	return utf8.RuneCountInString(text)
}

// ByteCount 统计字节数
func ByteCount(text string) int {
	return len(text)
}

// SentenceCount 统计句子数
func SentenceCount(text string) int {
	// 简单实现：按.!?分割
	sentences := strings.FieldsFunc(text, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})
	return len(sentences)
}

// ParagraphCount 统计段落数
func ParagraphCount(text string) int {
	paragraphs := strings.Split(text, "\n\n")
	count := 0
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	return count
}

// ExtractWords 提取单词
func ExtractWords(text string) []string {
	return strings.Fields(text)
}

// ExtractSentences 提取句子
func ExtractSentences(text string) []string {
	// 简化实现
	re := regexp.MustCompile(`[.!?]+\s+`)
	sentences := re.Split(text, -1)

	result := make([]string, 0, len(sentences))
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}

// ExtractParagraphs 提取段落
func ExtractParagraphs(text string) []string {
	paragraphs := strings.Split(text, "\n\n")
	result := make([]string, 0, len(paragraphs))

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

// MostCommonWords 获取最常见的单词
func MostCommonWords(text string, n int) []WordFreq {
	words := ExtractWords(strings.ToLower(text))
	freq := make(map[string]int)

	for _, word := range words {
		// 清洗单词
		word = CleanWord(word)
		if word != "" {
			freq[word]++
		}
	}

	// 转换为数组并排序
	wfs := make([]WordFreq, 0, len(freq))
	for word, count := range freq {
		wfs = append(wfs, WordFreq{Word: word, Count: count})
	}

	// 简单排序
	for i := 0; i < len(wfs)-1; i++ {
		for j := i + 1; j < len(wfs); j++ {
			if wfs[j].Count > wfs[i].Count {
				wfs[i], wfs[j] = wfs[j], wfs[i]
			}
		}
	}

	if n > len(wfs) {
		n = len(wfs)
	}

	return wfs[:n]
}

// WordFreq 词频
type WordFreq struct {
	Word  string
	Count int
}

// CleanWord 清洗单词
func CleanWord(word string) string {
	// 移除标点符号
	var result []rune
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// RemovePunctuation 移除标点符号
func RemovePunctuation(text string) string {
	var result []rune
	for _, r := range text {
		if !unicode.IsPunct(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// RemoveNumbers 移除数字
func RemoveNumbers(text string) string {
	var result []rune
	for _, r := range text {
		if !unicode.IsDigit(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// RemoveSpecialChars 移除特殊字符
func RemoveSpecialChars(text string) string {
	var result []rune
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// RemoveWhitespace 移除空白
func RemoveWhitespace(text string) string {
	var result []rune
	for _, r := range text {
		if !unicode.IsSpace(r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// CollapseWhitespace 压缩空白
func CollapseWhitespace(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// ToSnakeCase 转为蛇形命名
func ToSnakeCase(text string) string {
	var result []rune
	for i, r := range text {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// ToKebabCase 转为短横线命名
func ToKebabCase(text string) string {
	var result []rune
	for i, r := range text {
		if i > 0 && (r >= 'A' && r <= 'Z') {
			result = append(result, '-')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// ToCamelCase 转为驼峰命名
func ToCamelCase(text string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
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

// ToPascalCase 转为帕斯卡命名
func ToPascalCase(text string) string {
	words := strings.FieldsFunc(text, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// ToTitleCase 转为标题格式
func ToTitleCase(text string) string {
	return strings.Title(strings.ToLower(text))
}

// Reverse 反转字符串
func Reverse(text string) string {
	runes := []rune(text)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Contains 检查是否包含子串（不区分大小写）
func ContainsIgnoreCase(text, substr string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substr))
}

// IndexOf 查找子串位置
func IndexOf(text, substr string) int {
	return strings.Index(text, substr)
}

// LastIndexOf 查找子串最后位置
func LastIndexOf(text, substr string) int {
	return strings.LastIndex(text, substr)
}

// CountOccurrences 统计子串出现次数
func CountOccurrences(text, substr string) int {
	count := 0
	index := 0

	for {
		i := strings.Index(text[index:], substr)
		if i == -1 {
			break
		}
		count++
		index += i + len(substr)
	}

	return count
}

// ReplaceAll 替换所有子串
func ReplaceAll(text, old, new string) string {
	return strings.ReplaceAll(text, old, new)
}

// ReplaceAllIgnoreCase 替换所有子串（不区分大小写）
func ReplaceAllIgnoreCase(text, old, new string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(old))
	return re.ReplaceAllString(text, new)
}

// Substring 截取子串
func Substring(text string, start, end int) string {
	runes := []rune(text)

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

// Left 获取左边N个字符
func Left(text string, n int) string {
	runes := []rune(text)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[:n])
}

// Right 获取右边N个字符
func Right(text string, n int) string {
	runes := []rune(text)
	if n > len(runes) {
		n = len(runes)
	}
	return string(runes[len(runes)-n:])
}

// TrimLeft 移除左边空白
func TrimLeft(text string) string {
	return strings.TrimLeftFunc(text, unicode.IsSpace)
}

// TrimRight 移除右边空白
func TrimRight(text string) string {
	return strings.TrimRightFunc(text, unicode.IsSpace)
}

// Trim 移除两边空白
func Trim(text string) string {
	return strings.TrimSpace(text)
}

// TrimPrefix 移除前缀
func TrimPrefix(text, prefix string) string {
	return strings.TrimPrefix(text, prefix)
}

// TrimSuffix 移除后缀
func TrimSuffix(text, suffix string) string {
	return strings.TrimSuffix(text, suffix)
}

// PadLeft 左填充
func PadLeft(text string, pad string, length int) string {
	for len(text) < length {
		text = pad + text
	}
	if len(text) > length {
		return text[len(text)-length:]
	}
	return text
}

// PadRight 右填充
func PadRight(text string, pad string, length int) string {
	for len(text) < length {
		text = text + pad
	}
	return text[:length]
}

// Center 居中文本
func Center(text string, pad string, length int) string {
	if len(text) >= length {
		return text
	}

	totalPad := length - len(text)
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	return strings.Repeat(pad, leftPad) + text + strings.Repeat(pad, rightPad)
}

// SplitLines 按行分割
func SplitLines(text string) []string {
	return strings.Split(text, "\n")
}

// JoinLines 连接行
func JoinLines(lines []string) string {
	return strings.Join(lines, "\n")
}

// Indent 缩进
func Indent(text string, indent string) string {
	lines := SplitLines(text)
	for i, line := range lines {
		lines[i] = indent + line
	}
	return JoinLines(lines)
}

// Dedent 去除缩进
func Dedent(text string) string {
	lines := SplitLines(text)

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
		return text
	}

	// 去除缩进
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return JoinLines(lines)
}

// Truncate 截断
func Truncate(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

// Abbreviate 缩写
func Abbreviate(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return Left(text, maxLen/2-1) + "..." + Right(text, maxLen/2-1)
}

// IsBlank 检查是否空白
func IsBlank(text string) bool {
	return len(strings.TrimSpace(text)) == 0
}

// IsNumeric 检查是否为数字
func IsNumeric(text string) bool {
	if len(text) == 0 {
		return false
	}

	for _, r := range text {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// IsAlpha 检查是否为字母
func IsAlpha(text string) bool {
	for _, r := range text {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// IsAlphanumeric 检查是否为字母数字
func IsAlphanumeric(text string) bool {
	for _, r := range text {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// Lowercase 转小写
func Lowercase(text string) string {
	return strings.ToLower(text)
}

// Uppercase 转大写
func Uppercase(text string) string {
	return strings.ToUpper(text)
}

// SwapCase 大小写互换
func SwapCase(text string) string {
	var result []rune
	for _, r := range text {
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
func Capitalize(text string) string {
	runes := []rune(text)
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

// Uncapitalize 首字母小写
func Uncapitalize(text string) string {
	runes := []rune(text)
	if len(runes) > 0 {
		runes[0] = unicode.ToLower(runes[0])
	}
	return string(runes)
}

// Titleize 标题化
func Titleize(text string) string {
	words := strings.Fields(text)
	for i, word := range words {
		words[i] = Capitalize(strings.ToLower(word))
	}
	return strings.Join(words, " ")
}

// SplitWords 分割单词
func SplitWords(text string) []string {
	return strings.Fields(text)
}

// JoinWords 连接单词
func JoinWords(words []string, sep string) string {
	return strings.Join(words, sep)
}

// FindSimilarWords 查找相似单词
func FindSimilarWords(word string, words []string, threshold float64) []string {
	var similar []string

	word = strings.ToLower(word)

	for _, w := range words {
		w = strings.ToLower(w)
		similarity := Similarity(word, w)
		if similarity >= threshold {
			similar = append(similar, w)
		}
	}

	return similar
}

// Similarity 计算相似度
func Similarity(s1, s2 string) float64 {
	// 简单实现：编辑距离
距离 := LevenshteinDistance(s1, s2)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(距离)/float64(maxLen)
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

			min := matrix[i-1] + 1     // 删除
			if temp := matrix[i] + 1; temp < min {
				min = temp  // 插入
			}
			if temp := prev + cost; temp < min {
				min = temp  // 替换
			}

			prev = matrix[i]
			matrix[i] = min
		}
	}

	return matrix[len1]
}

// Diff 生成差异
func Diff(text1, text2 string) []DiffLine {
	lines1 := strings.Split(text1, "\n")
	lines2 := strings.Split(text2, "\n")

	var diffs []DiffLine
	maxLen := len(lines1)
	if len(lines2) > maxLen {
		maxLen = len(lines2)
	}

	for i := 0; i < maxLen; i++ {
		var line1, line2 string

		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			diffs = append(diffs, DiffLine{
				Line:  i + 1,
				Old:   line1,
				New:   line2,
				Type:  DiffModified,
			})
		}
	}

	return diffs
}

// DiffLine 差异行
type DiffLine struct {
	Line int
	Old  string
	New  string
	Type DiffType
}

// DiffType 差异类型
type DiffType int

const (
	DiffAdded    DiffType = iota
	DiffRemoved
	DiffModified
)

// WrapText 文本换行
func WrapText(text string, width int) string {
	words := strings.Fields(text)
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

// Repeat 重复字符串
func Repeat(text string, count int) string {
	return strings.Repeat(text, count)
}

// Chunk 分块
func Chunk(text string, size int) []string {
	if size <= 0 {
		return []string{text}
	}

	runes := []rune(text)
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

// NormalizeWhitespace 标准化空白
func NormalizeWhitespace(text string) string {
	re := regexp.MustCompile(`\s+`)
	return strings.TrimSpace(re.ReplaceAllString(text, " "))
}

// StripTags 移除HTML标签
func StripTags(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, "")
}

// EscapeHTML 转义HTML
func EscapeHTML(text string) string {
	replacements := map[string]string{
		"&":  "&amp;",
		"<":  "&lt;",
		">":  "&gt;",
		"\"": "&quot;",
		"'":  "&#39;",
	}

	result := text
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}

// UnescapeHTML 反转义HTML
func UnescapeHTML(html string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
	}

	result := html
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	return result
}
