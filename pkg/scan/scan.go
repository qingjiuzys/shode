// Package scan 提供扫描和匹配功能
package scan

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Scanner 扫描器
type Scanner struct {
	patterns []*regexp.Regexp
	ignore   []*regexp.Regexp
}

// NewScanner 创建扫描器
func NewScanner() *Scanner {
	return &Scanner{
		patterns: make([]*regexp.Regexp, 0),
		ignore:   make([]*regexp.Regexp, 0),
	}
}

// AddPattern 添加匹配模式
func (s *Scanner) AddPattern(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	s.patterns = append(s.patterns, re)
	return nil
}

// AddIgnore 添加忽略模式
func (s *Scanner) AddIgnore(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	s.ignore = append(s.ignore, re)
	return nil
}

// Match 检查文本是否匹配
func (s *Scanner) Match(text string) bool {
	// 检查是否应该忽略
	for _, re := range s.ignore {
		if re.MatchString(text) {
			return false
		}
	}

	// 检查是否匹配
	if len(s.patterns) == 0 {
		return true
	}

	for _, re := range s.patterns {
		if re.MatchString(text) {
			return true
		}
	}

	return false
}

// FindInFile 在文件中查找匹配项
func (s *Scanner) FindInFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return s.FindInReader(file)
}

// FindInReader 在reader中查找匹配项
func (s *Scanner) FindInReader(r io.Reader) ([]string, error) {
	var matches []string

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if s.Match(line) {
			matches = append(matches, line)
		}
	}

	return matches, scanner.Err()
}

// FileScanner 文件扫描器
type FileScanner struct {
	root      string
	extensions []string
	recursive bool
	ignore    []string
}

// NewFileScanner 创建文件扫描器
func NewFileScanner(root string) *FileScanner {
	return &FileScanner{
		root:      root,
		recursive: true,
		extensions: make([]string, 0),
		ignore:    make([]string, 0),
	}
}

// SetRecursive 设置是否递归扫描
func (fs *FileScanner) SetRecursive(recursive bool) *FileScanner {
	fs.recursive = recursive
	return fs
}

// AddExtension 添加文件扩展名
func (fs *FileScanner) AddExtension(ext string) *FileScanner {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	fs.extensions = append(fs.extensions, ext)
	return fs
}

// AddExtensions 添加多个文件扩展名
func (fs *FileScanner) AddExtensions(exts ...string) *FileScanner {
	for _, ext := range exts {
		fs.AddExtension(ext)
	}
	return fs
}

// AddIgnore 添加忽略模式
func (fs *FileScanner) AddIgnore(pattern string) *FileScanner {
	fs.ignore = append(fs.ignore, pattern)
	return fs
}

// Scan 扫描文件
func (fs *FileScanner) Scan() ([]string, error) {
	var files []string

	err := filepath.Walk(fs.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			// 检查是否应该忽略该目录
			for _, pattern := range fs.ignore {
				if strings.Contains(path, pattern) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// 检查扩展名
		if len(fs.extensions) > 0 {
			ext := filepath.Ext(path)
			found := false
			for _, e := range fs.extensions {
				if ext == e {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// 检查是否应该忽略
		for _, pattern := range fs.ignore {
			if strings.Contains(path, pattern) {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

// ScanFiles 扫描文件并返回文件信息
func (fs *FileScanner) ScanFiles() ([]*FileInfo, error) {
	paths, err := fs.Scan()
	if err != nil {
		return nil, err
	}

	infos := make([]*FileInfo, 0, len(paths))
	for _, path := range paths {
		info, err := NewFileInfo(path)
		if err != nil {
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// FileInfo 文件信息
type FileInfo struct {
	Path     string
	Name     string
	Ext      string
	Size     int64
	Mode     os.FileMode
	ModTime  string
	IsDir    bool
}

// NewFileInfo 创建文件信息
func NewFileInfo(path string) (*FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Path:    path,
		Name:    filepath.Base(path),
		Ext:     filepath.Ext(path),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime().String(),
		IsDir:   info.IsDir(),
	}, nil
}

// TextScanner 文本扫描器
type TextScanner struct {
	scanner *Scanner
}

// NewTextScanner 创建文本扫描器
func NewTextScanner() *TextScanner {
	return &TextScanner{
		scanner: NewScanner(),
	}
}

// AddPattern 添加匹配模式
func (ts *TextScanner) AddPattern(pattern string) error {
	return ts.scanner.AddPattern(pattern)
}

// AddIgnore 添加忽略模式
func (ts *TextScanner) AddIgnore(pattern string) error {
	return ts.scanner.AddIgnore(pattern)
}

// ScanString 扫描字符串
func (ts *TextScanner) ScanString(text string) []string {
	var matches []string

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if ts.scanner.Match(line) {
			matches = append(matches, line)
		}
	}

	return matches
}

// ScanFile 扫描文件
func (ts *TextScanner) ScanFile(filename string) ([]string, error) {
	return ts.scanner.FindInFile(filename)
}

// ScanDir 扫描目录中的所有文件
func (ts *TextScanner) ScanDir(dir string) (map[string][]string, error) {
	results := make(map[string][]string)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		matches, err := ts.ScanFile(path)
		if err != nil {
			return nil
		}

		if len(matches) > 0 {
			results[path] = matches
		}

		return nil
	})

	return results, err
}

// GrepScanner Grep扫描器
type GrepScanner struct {
	pattern string
	context int
	ignore  []string
}

// NewGrepScanner 创建Grep扫描器
func NewGrepScanner(pattern string) *GrepScanner {
	return &GrepScanner{
		pattern: pattern,
		context: 0,
		ignore:  make([]string, 0),
	}
}

// SetContext 设置上下文行数
func (gs *GrepScanner) SetContext(lines int) *GrepScanner {
	gs.context = lines
	return gs
}

// AddIgnore 添加忽略模式
func (gs *GrepScanner) AddIgnore(pattern string) *GrepScanner {
	gs.ignore = append(gs.ignore, pattern)
	return gs
}

// Search 在文件中搜索
func (gs *GrepScanner) Search(filename string) ([]*Match, error) {
	re, err := regexp.Compile(gs.pattern)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []*Match
	scanner := bufio.NewScanner(file)
	var lines []string
	lineNum := 0

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		lineNum++

		if re.MatchString(lines[len(lines)-1]) {
			match := &Match{
				File:    filename,
				Line:    lineNum,
				Text:    lines[len(lines)-1],
				Context: make([]string, 0),
			}

			// 添加上下文
			if gs.context > 0 {
				start := len(lines) - gs.context - 1
				if start < 0 {
					start = 0
				}
				match.Context = lines[start:]
			}

			matches = append(matches, match)
		}
	}

	return matches, scanner.Err()
}

// SearchDir 在目录中搜索
func (gs *GrepScanner) SearchDir(dir string) (map[string][]*Match, error) {
	results := make(map[string][]*Match)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过忽略的文件
		for _, pattern := range gs.ignore {
			if strings.Contains(path, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		matches, err := gs.Search(path)
		if err != nil {
			return nil
		}

		if len(matches) > 0 {
			results[path] = matches
		}

		return nil
	})

	return results, err
}

// Match 匹配结果
type Match struct {
	File    string
	Line    int
	Text    string
	Context []string
}

// String 字符串表示
func (m *Match) String() string {
	if len(m.Context) > 0 {
		return fmt.Sprintf("%s:%d: %s\nContext:\n%s", m.File, m.Line, m.Text, strings.Join(m.Context, "\n"))
	}
	return fmt.Sprintf("%s:%d: %s", m.File, m.Line, m.Text)
}

// ReplaceScanner 替换扫描器
type ReplaceScanner struct {
	search   string
	replace  string
	regex    bool
	ignore   []string
}

// NewReplaceScanner 创建替换扫描器
func NewReplaceScanner(search, replace string) *ReplaceScanner {
	return &ReplaceScanner{
		search:  search,
		replace: replace,
		regex:   false,
		ignore:  make([]string, 0),
	}
}

// SetRegex 设置是否使用正则表达式
func (rs *ReplaceScanner) SetRegex(regex bool) *ReplaceScanner {
	rs.regex = regex
	return rs
}

// AddIgnore 添加忽略模式
func (rs *ReplaceScanner) AddIgnore(pattern string) *ReplaceScanner {
	rs.ignore = append(rs.ignore, pattern)
	return rs
}

// ReplaceInFile 在文件中替换
func (rs *ReplaceScanner) ReplaceInFile(filename string, inPlace bool) (string, int, error) {
	// 读取文件
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", 0, err
	}

	text := string(content)
	var result string
	var count int

	if rs.regex {
		// 正则表达式替换
		re, err := regexp.Compile(rs.search)
		if err != nil {
			return "", 0, err
		}
		result = re.ReplaceAllString(text, rs.replace)
		count = len(re.FindAllStringIndex(text, -1))
	} else {
		// 普通字符串替换
		result = strings.ReplaceAll(text, rs.search, rs.replace)
		count = strings.Count(text, rs.search)
	}

	if inPlace {
		err = os.WriteFile(filename, []byte(result), 0644)
		if err != nil {
			return "", 0, err
		}
	}

	return result, count, nil
}

// ReplaceInDir 在目录中替换
func (rs *ReplaceScanner) ReplaceInDir(dir string, inPlace bool) (map[string]int, error) {
	results := make(map[string]int)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过忽略的文件
		for _, pattern := range rs.ignore {
			if strings.Contains(path, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if info.IsDir() {
			return nil
		}

		_, count, err := rs.ReplaceInFile(path, inPlace)
		if err != nil {
			return err
		}

		if count > 0 {
			results[path] = count
		}

		return nil
	})

	return results, err
}

// DiffScanner 差异扫描器
type DiffScanner struct {
	lines1 []string
	lines2 []string
}

// NewDiffScanner 创建差异扫描器
func NewDiffScanner() *DiffScanner {
	return &DiffScanner{
		lines1: make([]string, 0),
		lines2: make([]string, 0),
	}
}

// SetFile1 设置第一个文件
func (ds *DiffScanner) SetFile1(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	ds.lines1 = strings.Split(string(content), "\n")
	return nil
}

// SetFile2 设置第二个文件
func (ds *DiffScanner) SetFile2(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	ds.lines2 = strings.Split(string(content), "\n")
	return nil
}

// SetText1 设置第一个文本
func (ds *DiffScanner) SetText1(text string) {
	ds.lines1 = strings.Split(text, "\n")
}

// SetText2 设置第二个文本
func (ds *DiffScanner) SetText2(text string) {
	ds.lines2 = strings.Split(text, "\n")
}

// Compare 比较差异
func (ds *DiffScanner) Compare() []*Diff {
	var diffs []*Diff

	maxLen := len(ds.lines1)
	if len(ds.lines2) > maxLen {
		maxLen = len(ds.lines2)
	}

	for i := 0; i < maxLen; i++ {
		var line1, line2 string

		if i < len(ds.lines1) {
			line1 = ds.lines1[i]
		}
		if i < len(ds.lines2) {
			line2 = ds.lines2[i]
		}

		if line1 != line2 {
			diff := &Diff{
				Line:  i + 1,
				Old:   line1,
				New:   line2,
				Type:  DiffModified,
			}

			if line1 == "" {
				diff.Type = DiffAdded
			} else if line2 == "" {
				diff.Type = DiffRemoved
			}

			diffs = append(diffs, diff)
		}
	}

	return diffs
}

// DiffType 差异类型
type DiffType int

const (
	DiffAdded    DiffType = iota
	DiffRemoved
	DiffModified
)

// Diff 差异
type Diff struct {
	Line int
	Old  string
	New  string
	Type DiffType
}

func (d *Diff) String() string {
	switch d.Type {
	case DiffAdded:
		return fmt.Sprintf("+ %d: %s", d.Line, d.New)
	case DiffRemoved:
		return fmt.Sprintf("- %d: %s", d.Line, d.Old)
	case DiffModified:
		return fmt.Sprintf("~ %d: %s -> %s", d.Line, d.Old, d.New)
	}
	return ""
}

// WordCounter 词频统计器
type WordCounter struct {
	words map[string]int
	ignore []*regexp.Regexp
}

// NewWordCounter 创建词频统计器
func NewWordCounter() *WordCounter {
	return &WordCounter{
		words:  make(map[string]int),
		ignore: make([]*regexp.Regexp, 0),
	}
}

// AddIgnore 添加忽略模式
func (wc *WordCounter) AddIgnore(pattern string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	wc.ignore = append(wc.ignore, re)
	return nil
}

// CountString 统计字符串
func (wc *WordCounter) CountString(text string) error {
	words := strings.Fields(text)

	for _, word := range words {
		// 检查是否应该忽略
		ignored := false
		for _, re := range wc.ignore {
			if re.MatchString(word) {
				ignored = true
				break
			}
		}

		if !ignored {
			wc.words[word]++
		}
	}

	return nil
}

// CountFile 统计文件
func (wc *WordCounter) CountFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return wc.CountString(string(content))
}

// Top 获取前N个高频词
func (wc *WordCounter) Top(n int) []WordFreq {
	type kv struct {
		word  string
		count int
	}

	var sorted []kv
	for word, count := range wc.words {
		sorted = append(sorted, kv{word, count})
	}

	// 简单排序（实际可以用更高效的算法）
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	result := make([]WordFreq, 0, n)
	for i := 0; i < n && i < len(sorted); i++ {
		result = append(result, WordFreq{
			Word:  sorted[i].word,
			Count: sorted[i].count,
		})
	}

	return result
}

// WordFreq 词频
type WordFreq struct {
	Word  string
	Count int
}

// LineCounter 行数统计器
type LineCounter struct {
	total    int
	blank    int
	code     int
	comment  int
}

// NewLineCounter 创建行数统计器
func NewLineCounter() *LineCounter {
	return &LineCounter{}
}

// CountFile 统计文件
func (lc *LineCounter) CountFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return lc.CountReader(file)
}

// CountReader 统计reader
func (lc *LineCounter) CountReader(r io.Reader) error {
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		lc.total++

		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			lc.blank++
		} else if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") {
			lc.comment++
		} else {
			lc.code++
		}
	}

	return scanner.Err()
}

// Total 总行数
func (lc *LineCounter) Total() int {
	return lc.total
}

// Blank 空行数
func (lc *LineCounter) Blank() int {
	return lc.blank
}

// Code 代码行数
func (lc *LineCounter) Code() int {
	return lc.code
}

// Comment 注释行数
func (lc *LineCounter) Comment() int {
	return lc.comment
}
