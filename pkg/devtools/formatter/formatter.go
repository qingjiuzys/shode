// Package formatter 代码格式化工具
package formatter

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Formatter 代码格式化器
type Formatter struct {
	config     *FormatterConfig
	diff       bool
	check      bool
	write      bool
	stdin      bool
	verbose    bool
}

// FormatterConfig 格式化器配置
type FormatterConfig struct {
	Indent       int
	TabWidth     int
	MaxLineLength int
	IndentStyle  string // "tab" or "space"
	Semicolons   bool
	Quotes       string // "double" or "single"
	TrailingComma bool
}

// FormatResult 格式化结果
type FormatResult struct {
	Path      string
	Changed    bool
	Before    string
	After     string
	Diff      string
	Error     error
}

// NewFormatter 创建格式化器
func NewFormatter(config *FormatterConfig) *Formatter {
	if config == nil {
		config = &FormatterConfig{
			Indent:        4,
			TabWidth:      4,
			MaxLineLength: 100,
			IndentStyle:   "space",
			Semicolons:    false,
			Quotes:        "double",
			TrailingComma: false,
		}
	}

	return &Formatter{
		config: config,
	}
}

// Format 格式化代码
func (f *Formatter) Format(code string) (string, error) {
	// 1. 标准化换行符
	code = strings.ReplaceAll(code, "\r\n", "\n")

	// 2. 移除行尾空格
	lines := strings.Split(code, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	code = strings.Join(lines, "\n")

	// 3. 规范化缩进
	code = f.normalizeIndent(code)

	// 4. 规范化空格
	code = f.normalizeSpaces(code)

	// 5. 规范化运算符
	code = f.normalizeOperators(code)

	// 6. 规范化引号
	if f.config.Quotes == "double" {
		code = f.normalizeQuotes(code, "double")
	} else {
		code = f.normalizeQuotes(code, "single")
	}

	// 7. 检查行长度
	if f.config.MaxLineLength > 0 {
		code = f.checkLineLength(code)
	}

	// 8. 添加末尾换行
	if !strings.HasSuffix(code, "\n") {
		code += "\n"
	}

	return code, nil
}

// FormatFile 格式化文件
func (f *Formatter) FormatFile(path string) (*FormatResult, error) {
	// 读取文件
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	before := string(content)

	// 格式化
	after, err := f.Format(before)
	if err != nil {
		return &FormatResult{
			Path:  path,
			Error: err,
		}, err
	}

	// 检查是否有变化
	changed := before != after

	result := &FormatResult{
		Path:   path,
		Changed: changed,
		Before: before,
		After:  after,
	}

	// 生成 diff
	if f.diff {
		result.Diff = f.generateDiff(before, after)
	}

	// 写入文件
	if f.write && changed {
		if err := os.WriteFile(path, []byte(after), 0644); err != nil {
			return result, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return result, nil
}

// normalizeIndent 规范化缩进
func (f *Formatter) normalizeIndent(code string) string {
	lines := strings.Split(code, "\n")
	result := make([]string, 0, len(lines))

	indentStr := strings.Repeat(" ", f.config.Indent)

	for _, line := range lines {
		// 计算当前缩进级别
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			result = append(result, "")
			continue
		}

		// 计算缩进层级
		indentLevel := f.calculateIndentLevel(line)

		// 应用缩进
		if f.config.IndentStyle == "tab" {
			result = append(result, strings.Repeat("\t", indentLevel)+trimmed)
		} else {
			result = append(result, strings.Repeat(indentStr, indentLevel)+trimmed)
		}
	}

	return strings.Join(result, "\n")
}

// calculateIndentLevel 计算缩进层级
func (f *Formatter) calculateIndentLevel(line string) int {
	// 简化实现：基于括号和关键字的缩进
	level := 0

	// 检查增加缩进的关键字
	openKeywords := []string{"func", "if", "else", "for", "while", "switch", "case", "default"}
	for _, keyword := range openKeywords {
		if strings.Contains(line, keyword+" ") {
			level++
		}
	}

	// 检查闭合括号
	closeBraces := strings.Count(line, "}")
	if closeBraces > 0 {
		level -= closeBraces
	}

	if level < 0 {
		level = 0
	}

	return level
}

// normalizeSpaces 规范化空格
func (f *Formatter) normalizeSpaces(code string) string {
	// 运算符周围添加空格
	operators := []string{"+", "-", "*", "/", "=", "==", "!=", "<", ">", "<=", ">="}

	for _, op := range operators {
		// operator前面加空格（如果还没有）
		code = regexp.MustCompile(`([^ ])`+regexp.QuoteMeta(op)).ReplaceAllString(code, "$1 "+op+" ")
	}

	// 移除多余空格
	code = regexp.MustCompile(` +`).ReplaceAllString(code, " ")

	return code
}

// normalizeOperators 规范化运算符
func (f *Formatter) normalizeOperators(code string) string {
	// 统一运算符
	replacements := map[string]string{
		"=":  " = ",
		"==": "==",
		"!=": "!=",
		"<=": "<=",
		">=": ">=",
		"&&": " && ",
		"||": " || ",
		"!":  "!",
	}

	for old, new := range replacements {
		code = strings.ReplaceAll(code, old, new)
	}

	return code
}

// normalizeQuotes 规范化引号
func (f *Formatter) normalizeQuotes(code string, style string) string {
	if style == "double" {
		// 单引号转双引号
		code = regexp.MustCompile(`'([^']*)'`).ReplaceAllString(code, `"$1"`)
	} else {
		// 双引号转单引号
		code = regexp.MustCompile(`"([^"]*)"`).ReplaceAllString(code, `'$1'`)
	}

	return code
}

// checkLineLength 检查行长度
func (f *Formatter) checkLineLength(code string) string {
	lines := strings.Split(code, "\n")

	for i, line := range lines {
		if len(line) > f.config.MaxLineLength {
			// 尝试在运算符处换行
			lines[i] = f.breakLongLine(line)
		}
	}

	return strings.Join(lines, "\n")
}

// breakLongLine 断开长行
func (f *Formatter) breakLongLine(line string) string {
	// 在运算符处换行
	breakPoints := []string{" + ", " - ", " * ", " / ", " && ", " || "}

	for _, bp := range breakPoints {
		if strings.Contains(line, bp) {
			parts := strings.SplitN(line, bp, 2)
			if len(parts) == 2 {
				indent := strings.Repeat(" ", f.config.Indent)
				return parts[0] + bp + "\n" + indent + parts[1]
			}
		}
	}

	return line
}

// generateDiff 生成差异
func (f *Formatter) generateDiff(before, after string) string {
	diff := &strings.Builder{}

	diff.WriteString("--- a/original\n")
	diff.WriteString("+++ b/formatted\n")
	diff.WriteString("@@ -1,1 +1,1 @@\n")

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	maxLines := len(beforeLines)
	if len(afterLines) > maxLines {
		maxLines = len(afterLines)
	}

	for i := 0; i < maxLines; i++ {
		var beforeLine, afterLine string

		if i < len(beforeLines) {
			beforeLine = beforeLines[i]
		}
		if i < len(afterLines) {
			afterLine = afterLines[i]
		}

		if beforeLine == afterLine {
			diff.WriteString(" " + beforeLine + "\n")
		} else {
			if beforeLine != "" {
				diff.WriteString("-" + beforeLine + "\n")
			}
			if afterLine != "" {
				diff.WriteString("+" + afterLine + "\n")
			}
		}
	}

	return diff.String()
}

// CheckFormat 检查格式
func (f *Formatter) CheckFormat(code string) bool {
	formatted, err := f.Format(code)
	if err != nil {
		return false
	}

	return code == formatted
}

// FormatFiles 格式化多个文件
func (f *Formatter) FormatFiles(paths []string) ([]*FormatResult, error) {
	results := make([]*FormatResult, 0, len(paths))

	for _, path := range paths {
		result, err := f.FormatFile(path)
		if err != nil {
			result = &FormatResult{
				Path:  path,
				Error: err,
			}
		}

		results = append(results, result)

		// 打印结果
		if f.verbose {
			if result.Error != nil {
				fmt.Printf("Error: %s: %v\n", path, result.Error)
			} else if result.Changed {
				fmt.Printf("Formatted: %s\n", path)
			}
		}
	}

	return results, nil
}

// SetDiff 设置是否输出差异
func (f *Formatter) SetDiff(diff bool) {
	f.diff = diff
}

// SetCheck 设置是否只检查
func (f *Formatter) SetCheck(check bool) {
	f.check = check
}

// SetWrite 设置是否写入文件
func (f *Formatter) SetWrite(write bool) {
	f.write = write
}

// SetVerbose 设置详细输出
func (f *Formatter) SetVerbose(verbose bool) {
	f.verbose = verbose
}

// PrintStats 打印统计信息
func (f *Formatter) PrintStats(results []*FormatResult) {
	total := len(results)
	changed := 0
	errors := 0

	for _, result := range results {
		if result.Error != nil {
			errors++
		} else if result.Changed {
			changed++
		}
	}

	fmt.Printf("\n📊 Format Statistics:\n")
	fmt.Printf("  Total files: %d\n", total)
	fmt.Printf("  Changed:    %d\n", changed)
	fmt.Printf("  Errors:     %d\n", errors)

	if f.check && changed > 0 {
		fmt.Printf("\n⚠️  %d file(s) need formatting\n", changed)
		fmt.Printf("   Run 'shode fmt -w' to fix\n")
	}
}
