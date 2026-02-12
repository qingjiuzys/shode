// Package searchutil 提供搜索工具
package searchutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// SearchResult 搜索结果
type SearchResult struct {
	Path    string
	Line    int
	Content string
	Match   string
}

// SearchInFile 在文件中搜索
func SearchInFile(filename, pattern string, caseSensitive bool) ([]SearchResult, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return SearchInBytes(filename, content, pattern, caseSensitive)
}

// SearchInBytes 在字节中搜索
func SearchInBytes(filename string, content []byte, pattern string, caseSensitive bool) ([]SearchResult, error) {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
		content = bytes.ToLower(content)
	}

	lines := bytes.Split(content, []byte{'\n'})
	var results []SearchResult

	for lineNum, line := range lines {
		lineStr := string(line)
		if strings.Contains(lineStr, pattern) {
			results = append(results, SearchResult{
				Path:    filename,
				Line:    lineNum + 1,
				Content: strings.TrimSpace(lineStr),
				Match:   pattern,
			})
		}
	}

	return results, nil
}

// SearchRegex 正则表达式搜索
func SearchRegex(filename, pattern string) ([]SearchResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(content, []byte{'\n'})
	var results []SearchResult

	for lineNum, line := range lines {
		lineStr := string(line)
		if re.MatchString(lineStr) {
			matches := re.FindAllString(lineStr, -1)
			for _, match := range matches {
				results = append(results, SearchResult{
					Path:    filename,
					Line:    lineNum + 1,
					Content: strings.TrimSpace(lineStr),
					Match:   match,
				})
			}
		}
	}

	return results, nil
}

// SearchInDir 在目录中搜索
func SearchInDir(dir, pattern string, recursive bool) ([]SearchResult, error) {
	var results []SearchResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		// 只搜索文本文件
		if !isTextFile(path) {
			return nil
		}

		matches, err := SearchInFile(path, pattern, false)
		if err != nil {
			return err
		}

		results = append(results, matches...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// isTextFile 检查是否为文本文件
func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))

	textExts := map[string]bool{
		".txt":        true,
		".md":         true,
		".json":       true,
		".xml":        true,
		".html":       true,
		".css":        true,
		".js":         true,
		".ts":         true,
		".go":         true,
		".py":         true,
		".java":       true,
		".c":          true,
		".cpp":        true,
		".h":          true,
		".hpp":        true,
		".sh":         true,
		".yml":        true,
		".yaml":       true,
		".toml":       true,
		".ini":        true,
		".conf":       true,
		".config":     true,
		".csv":        true,
		".sql":        true,
		".log":        true,
		".rst":        true,
		".tex":        true,
		".erb":        true,
		".haml":       true,
		".sass":       true,
		".scss":       true,
		".less":       true,
		".svg":        true,
	}

	return textExts[ext]
}

// FindFiles 查找文件
func FindFiles(dir string, pattern string) ([]string, error) {
	var files []string

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if re.MatchString(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// FindFilesByExt 按扩展名查找文件
func FindFilesByExt(dir string, exts ...string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fileExt := strings.ToLower(filepath.Ext(path))
		for _, ext := range exts {
			if strings.HasPrefix(ext, ".") {
				ext = ext
			} else {
				ext = "." + ext
			}

			if fileExt == ext {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// FindFilesByName 按文件名模式查找
func FindFilesByName(dir, namePattern string) ([]string, error) {
	return FindFiles(dir, namePattern)
}

// SearchContent 搜索文件内容
func SearchContent(dir, content string, recursive bool) ([]SearchResult, error) {
	return SearchInDir(dir, content, recursive)
}

// Grep grep风格搜索
func Grep(pattern, filename string) ([]SearchResult, error) {
	return SearchRegex(filename, pattern)
}

// GrepInDir grep目录搜索
func GrepInDir(pattern, dir string, recursive bool) ([]SearchResult, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextFile(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		lines := bytes.Split(content, []byte{'\n'})
		for lineNum, line := range lines {
			lineStr := string(line)
			if re.MatchString(lineStr) {
				matches := re.FindAllString(lineStr, -1)
				for _, match := range matches {
					results = append(results, SearchResult{
						Path:    path,
						Line:    lineNum + 1,
						Content: strings.TrimSpace(lineStr),
						Match:   match,
					})
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// ReplaceInFile 替换文件内容
func ReplaceInFile(filename, old, new string, caseSensitive bool) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var searchStr string
	if caseSensitive {
		searchStr = string(content)
		old = old
	} else {
		searchStr = strings.ToLower(string(content))
		old = strings.ToLower(old)
	}

	if !strings.Contains(searchStr, old) {
		return nil
	}

	var replacement string
	if caseSensitive {
		replacement = strings.ReplaceAll(string(content), old, new)
	} else {
		// 需要逐个替换保持大小写
		replacement = replaceIgnoreCase(string(content), old, new)
	}

	return os.WriteFile(filename, []byte(replacement), 0644)
}

func replaceIgnoreCase(content, old, new string) string {
	// 简化实现
	return strings.ReplaceAll(strings.ToLower(content), old, new)
}

// ReplaceInDir 替换目录中所有文件的内容
func ReplaceInDir(dir, old, new string, recursive bool) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextFile(path) {
			return nil
		}

		return ReplaceInFile(path, old, new, false)
	})
}

// FuzzyMatch 模糊匹配
func FuzzyMatch(text, pattern string) float64 {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	if text == pattern {
		return 1.0
	}

	if strings.Contains(text, pattern) {
		return 0.8
	}

	// 简化的模糊匹配算法
	patternRunes := []rune(pattern)
	textRunes := []rune(text)

	patternIdx := 0
	matches := 0

	for _, r := range textRunes {
		if patternIdx < len(patternRunes) && r == patternRunes[patternIdx] {
			matches++
			patternIdx++
		}
	}

	if patternIdx == 0 {
		return 0
	}

	return float64(matches) / float64(len(patternRunes))
}

// FilterByFilter 通过过滤器搜索
func FilterByFilter(results []SearchResult, filter func(SearchResult) bool) []SearchResult {
	var filtered []SearchResult
	for _, result := range results {
		if filter(result) {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// SortResults 排序搜索结果
func SortResults(results []SearchResult, by string) []SearchResult {
	sorted := make([]SearchResult, len(results))
	copy(sorted, results)

	switch by {
	case "path":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Path < sorted[j].Path
		})
	case "line":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Line < sorted[j].Line
		})
	default:
		// 默认按路径排序
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Path < sorted[j].Path
		})
	}

	return sorted
}

// UniqueResults 去重搜索结果
func UniqueResults(results []SearchResult) []SearchResult {
	seen := make(map[string]bool)
	unique := make([]SearchResult, 0)

	for _, result := range results {
		key := fmt.Sprintf("%s:%d:%s", result.Path, result.Line, result.Match)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, result)
		}
	}

	return unique
}

// GroupByPath 按路径分组
func GroupByPath(results []SearchResult) map[string][]SearchResult {
	groups := make(map[string][]SearchResult)

	for _, result := range results {
		groups[result.Path] = append(groups[result.Path], result)
	}

	return groups
}

// FormatResults 格式化搜索结果
func FormatResults(results []SearchResult, format string) string {
	switch format {
	case "verbose":
		return formatVerbose(results)
	case "compact":
		return formatCompact(results)
	default:
		return formatSimple(results)
	}
}

func formatVerbose(results []SearchResult) string {
	var buf strings.Builder

	for _, result := range results {
		buf.WriteString(fmt.Sprintf("%s:%d:%s\n", result.Path, result.Line, result.Content))
	}

	return buf.String()
}

func formatCompact(results []SearchResult) string {
	var buf strings.Builder

	for _, result := range results {
		buf.WriteString(fmt.Sprintf("%s:%d\n", result.Path, result.Line))
	}

	return buf.String()
}

func formatSimple(results []SearchResult) string {
	var buf strings.Builder

	for _, result := range results {
		buf.WriteString(fmt.Sprintf("%s\n", result.Path))
	}

	return buf.String()
}

// SearchString 搜索字符串
func SearchString(content, pattern string, caseSensitive bool) []Range {
	if !caseSensitive {
		content = strings.ToLower(content)
		pattern = strings.ToLower(pattern)
	}

	var ranges []Range
	start := 0

	for {
		idx := strings.Index(content[start:], pattern)
		if idx == -1 {
			break
		}

		ranges = append(ranges, Range{
			Start: start + idx,
			End:   start + idx + len(pattern),
		})

		start += idx + len(pattern)
	}

	return ranges
}

// Range 范围
type Range struct {
	Start int
	End   int
}

// SearchLines 搜索行
func SearchLines(lines []string, pattern string, caseSensitive bool) []int {
	if !caseSensitive {
		pattern = strings.ToLower(pattern)
	}

	var lineNumbers []int

	for i, line := range lines {
		lineStr := line
		if !caseSensitive {
			lineStr = strings.ToLower(line)
		}

		if strings.Contains(lineStr, pattern) {
			lineNumbers = append(lineNumbers, i+1)
		}
	}

	return lineNumbers
}

// FindCommonPrefix 查找公共前缀
func FindCommonStr(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) == 1 {
		return strs[0]
	}

	prefix := strs[0]
	for _, str := range strs[1:] {
		for len(prefix) > len(str) {
			if len(str) < len(prefix) {
				prefix, str = str, prefix
			} else {
				break
			}
		}

		minLen := len(prefix)
		if minLen > len(str) {
			minLen = len(str)
		}

		i := 0
		for i < minLen && prefix[i] == str[i] {
			i++
		}

		prefix = prefix[:i]
		if prefix == "" {
			break
		}
	}

	return prefix
}

// LevenshteinDistance 编辑距离
func LevenshteinDistance(a, b string) int {
	runes1 := []rune(a)
	runes2 := []rune(b)

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

			min := matrix[i-1] + 1
			if temp := matrix[i] + 1; temp < min {
				min = temp
			}
			if temp := prev + cost; temp < min {
				min = temp
			}

			prev = matrix[i]
			matrix[i] = min
		}
	}

	return matrix[len1]
}

// Similarity 相似度
func Similarity(a, b string) float64 {
	dist := LevenshteinDistance(a, b)
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}

	if maxLen == 0 {
		return 1.0
	}

	return 1.0 - float64(dist)/float64(maxLen)
}

// FindSimilar 查找相似字符串
func FindSimilar(target string, candidates []string, threshold float64) []string {
	var similar []string

	for _, candidate := range candidates {
		sim := Similarity(target, candidate)
		if sim >= threshold {
			similar = append(similar, candidate)
		}
	}

	return similar
}

// MatchWildcard 通配符匹配
func MatchWildcard(text, pattern string) bool {
	// 简化实现，只支持 * 和 ?
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, "\\*", ".*")
	pattern = strings.ReplaceAll(pattern, "\\?", ".")

	re, err := regexp.Compile("^" + pattern + "$")
	if err != nil {
		return false
	}

	return re.MatchString(text)
}

// MatchWildcards 通配符匹配（多个模式）
func MatchWildcards(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchWildcard(text, pattern) {
			return true
		}
	}
	return false
}

// SearchStream 流式搜索
func SearchStream(reader io.Reader, pattern string, caseSensitive bool) ([]SearchResult, error) {
	scanner := bufio.NewScanner(reader)
	var results []SearchResult
	lineNum := 0

	if !caseSensitive {
		pattern = strings.ToLower(pattern)
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		searchLine := line
		if !caseSensitive {
			searchLine = strings.ToLower(line)
		}

		if strings.Contains(searchLine, pattern) {
			results = append(results, SearchResult{
				Line:    lineNum,
				Content: line,
				Match:   pattern,
			})
		}
	}

	return results, scanner.Err()
}

// CountMatches 统计匹配次数
func CountMatches(content, pattern string, caseSensitive bool) int {
	if !caseSensitive {
		content = strings.ToLower(content)
		pattern = strings.ToLower(pattern)
	}

	return strings.Count(content, pattern)
}

// ExtractLines 提取包含模式的行
func ExtractLines(content string, pattern string) []string {
	lines := strings.Split(content, "\n")
	var extracted []string

	for _, line := range lines {
		if strings.Contains(line, pattern) {
			extracted = append(extracted, line)
		}
	}

	return extracted
}

// ContextLines 获取上下文行
func ContextLines(content string, lineNum, context int) []string {
	lines := strings.Split(content, "\n")

	start := lineNum - context - 1
	if start < 0 {
		start = 0
	}

	end := lineNum + context
	if end > len(lines) {
		end = len(lines)
	}

	return lines[start:end]
}

// Highlight 高亮匹配文本
func Highlight(content, pattern string) string {
	if pattern == "" {
		return content
	}

	// 简化实现：使用ANSI颜色
	reset := "\033[0m"
	bold := "\033[1m"
	color := "\033[31m" // 红色

	return strings.ReplaceAll(content, pattern, bold+color+pattern+reset)
}

// BuildIndex 构建搜索索引
func BuildIndex(dir string, recursive bool) (*Index, error) {
	index := &Index{
		Files:    make(map[string]*FileIndex),
			Words:    make(map[string][]string),
			recursive: recursive,
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			return nil
		}

		if !isTextFile(path) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		fileIndex := &FileIndex{
			Path:    path,
			Words:   tokenize(content),
			Lines:   strings.Split(string(content), "\n"),
		}

		index.Files[path] = fileIndex

		// 构建倒排索引
		for _, word := range fileIndex.Words {
			if len(word) > 2 { // 只索引长度大于2的词
				index.Words[word] = append(index.Words[word], path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return index, nil
}

// Index 搜索索引
type Index struct {
	Files     map[string]*FileIndex
	Words     map[string][]string
	recursive bool
}

// FileIndex 文件索引
type FileIndex struct {
	Path  string
	Words []string
	Lines []string
}

// Search 搜索索引
func (idx *Index) Search(query string) []SearchResult {
	query = strings.ToLower(query)
	terms := tokenize([]byte(query))

	var results []SearchResult

	// 简单的搜索逻辑：匹配文件中包含所有搜索词
	for _, fileIdx := range idx.Files {
		fileWords := make(map[string]bool)
		for _, word := range fileIdx.Words {
			fileWords[word] = true
		}

		matchesAll := true
		for _, term := range terms {
			if !fileWords[term] {
				matchesAll = false
				break
			}
		}

		if matchesAll {
			for lineNum, line := range fileIdx.Lines {
				lineLower := strings.ToLower(line)
				allTermsInLine := true

				for _, term := range terms {
					if !strings.Contains(lineLower, term) {
						allTermsInLine = false
						break
					}
				}

				if allTermsInLine {
					results = append(results, SearchResult{
						Path:    fileIdx.Path,
						Line:    lineNum + 1,
						Content: strings.TrimSpace(line),
					})
				}
			}
		}
	}

	return results
}

// tokenize 分词
func tokenize(content []byte) []string {
	// 简化实现：按空白和标点符号分割
	var words []string
	var word []rune

	for _, r := range string(content) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word = append(word, r)
		} else {
			if len(word) > 0 {
				words = append(words, strings.ToLower(string(word)))
				word = word[:0]
			}
		}
	}

	if len(word) > 0 {
		words = append(words, strings.ToLower(string(word)))
	}

	return words
}

// StringDistance 字符串距离（多种算法）
func StringDistance(a, b string, algorithm string) float64 {
	switch algorithm {
	case "levenshtein":
		dist := LevenshteinDistance(a, b)
		maxLen := len(a)
		if len(b) > maxLen {
			maxLen = len(b)
		}
		if maxLen == 0 {
			return 1.0
		}
		return 1.0 - float64(dist)/float64(maxLen)
	case "jaccard":
		return jaccardSimilarity(a, b)
	case "cosine":
		return cosineSimilarity(a, b)
	default:
		return Similarity(a, b)
	}
}

func jaccardSimilarity(a, b string) float64 {
	setA := make(map[rune]bool)
	setB := make(map[rune]bool)

	for _, r := range a {
		setA[r] = true
	}

	for _, r := range b {
		setB[r] = true
	}

	intersection := 0
	for r := range setA {
		if setB[r] {
			intersection++
		}
	}

	union := len(setA) + len(setB) - intersection

	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
}

func cosineSimilarity(a, b string) float64 {
	// 简化实现
	return 0.0
}

// AdvancedSearch 高级搜索
func AdvancedSearch(dir, query string, opts SearchOptions) ([]SearchResult, error) {
	var allResults []SearchResult

	// 执行搜索
	matches, err := SearchInDir(dir, query, opts.Recursive)
	if err != nil {
		return nil, err
	}

	// 应用过滤器
	for _, match := range matches {
			if opts.Filter != nil && !opts.Filter(match) {
			continue
		}

		allResults = append(allResults, match)
	}

	// 排序
	if opts.SortBy != "" {
		allResults = SortResults(allResults, opts.SortBy)
	}

	// 限制结果数量
	if opts.MaxResults > 0 && len(allResults) > opts.MaxResults {
		allResults = allResults[:opts.MaxResults]
	}

	return allResults, nil
}

// SearchOptions 搜索选项
type SearchOptions struct {
	Recursive  bool
	Filter     func(SearchResult) bool
	SortBy     string
	MaxResults int
}

// BatchSearch 批量搜索
func BatchSearch(dir string, patterns []string) (map[string][]SearchResult, error) {
	results := make(map[string][]SearchResult)

	for _, pattern := range patterns {
		matches, err := SearchInDir(dir, pattern, true)
		if err != nil {
			continue
		}
		results[pattern] = matches
	}

	return results, nil
}
