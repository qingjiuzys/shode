// Package loganalyzer 提供日志分析功能
package loganalyzer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Analyzer 日志分析器
type Analyzer struct {
	logFile    string
	patterns   []*Pattern
	mu         sync.Mutex
	stats      *Stats
	errors     []LogEntry
	warnings   []LogEntry
	byLevel    map[string][]LogEntry
	byTime     []LogEntry
	timeFormat string
}

// Pattern 日志模式
type Pattern struct {
	Name      string
	Regex     *regexp.Regexp
	Level     string
	Extractor func(match []string) map[string]interface{}
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Fields    map[string]interface{}
	LineNum   int
	Raw       string
}

// Stats 统计信息
type Stats struct {
	TotalLines   int
	ErrorCount   int
	WarningCount int
	InfoCount    int
	DebugCount   int
	ParseErrors  int
	StartTime    time.Time
	EndTime      time.Time
}

// NewAnalyzer 创建分析器
func NewAnalyzer(logFile string) *Analyzer {
	return &Analyzer{
		logFile:  logFile,
		patterns: make([]*Pattern, 0),
		byLevel:  make(map[string][]LogEntry),
		byTime:   make([]LogEntry, 0),
		stats:    &Stats{},
	}
}

// AddPattern 添加日志模式
func (a *Analyzer) AddPattern(name, pattern, level string) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile pattern %s: %w", name, err)
	}

	a.patterns = append(a.patterns, &Pattern{
		Name:  name,
		Regex: regex,
		Level: level,
	})

	return nil
}

// AddPatternWithExtractor 添加带提取器的模式
func (a *Analyzer) AddPatternWithExtractor(name, pattern, level string, extractor func([]string) map[string]interface{}) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("failed to compile pattern %s: %w", name, err)
	}

	a.patterns = append(a.patterns, &Pattern{
		Name:      name,
		Regex:     regex,
		Level:     level,
		Extractor: extractor,
	})

	return nil
}

// Parse 解析日志文件
func (a *Analyzer) Parse() error {
	file, err := os.Open(a.logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		entry, err := a.parseLine(line, lineNum)
		if err != nil {
			a.stats.ParseErrors++
			continue
		}

		if entry != nil {
			a.addEntry(entry)
		}

		a.stats.TotalLines++
	}

	return scanner.Err()
}

// parseLine 解析单行日志
func (a *Analyzer) parseLine(line string, lineNum int) (*LogEntry, error) {
	for _, pattern := range a.patterns {
		if pattern.Regex.MatchString(line) {
			matches := pattern.Regex.FindStringSubmatch(line)

			entry := &LogEntry{
				Level:   pattern.Level,
				Message: matches[0],
				LineNum: lineNum,
				Raw:     line,
			}

			// 提取字段
			if pattern.Extractor != nil {
				entry.Fields = pattern.Extractor(matches)
			}

			// 解析时间戳
			if timestamp, ok := entry.Fields["timestamp"].(time.Time); ok {
				entry.Timestamp = timestamp
			}

			return entry, nil
		}
	}

	return nil, nil
}

// addEntry 添加日志条目
func (a *Analyzer) addEntry(entry *LogEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 按级别分类
	a.byLevel[entry.Level] = append(a.byLevel[entry.Level], *entry)

	// 按时间排序
	a.byTime = append(a.byTime, *entry)

	// 统计
	switch entry.Level {
	case "ERROR", "error":
		a.errors = append(a.errors, *entry)
		a.stats.ErrorCount++
	case "WARNING", "warning", "WARN":
		a.warnings = append(a.warnings, *entry)
		a.stats.WarningCount++
	case "INFO", "info":
		a.stats.InfoCount++
	case "DEBUG", "debug":
		a.stats.DebugCount++
	}

	// 更新时间范围
	if !entry.Timestamp.IsZero() {
		if a.stats.StartTime.IsZero() || entry.Timestamp.Before(a.stats.StartTime) {
			a.stats.StartTime = entry.Timestamp
		}
		if entry.Timestamp.After(a.stats.EndTime) {
			a.stats.EndTime = entry.Timestamp
		}
	}
}

// GetStats 获取统计信息
func (a *Analyzer) GetStats() *Stats {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.stats
}

// GetErrors 获取错误日志
func (a *Analyzer) GetErrors() []LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.errors
}

// GetWarnings 获取警告日志
func (a *Analyzer) GetWarnings() []LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.warnings
}

// GetByLevel 按级别获取日志
func (a *Analyzer) GetByLevel(level string) []LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.byLevel[level]
}

// Search 搜索日志
func (a *Analyzer) Search(keyword string) []LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	var results []LogEntry

	for _, entry := range a.byTime {
		if strings.Contains(entry.Message, keyword) ||
			strings.Contains(entry.Raw, keyword) {
			results = append(results, entry)
		}
	}

	return results
}

// SearchByRegex 正则搜索
func (a *Analyzer) SearchByRegex(pattern string) ([]LogEntry, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	var results []LogEntry

	for _, entry := range a.byTime {
		if regex.MatchString(entry.Raw) {
			results = append(results, entry)
		}
	}

	return results, nil
}

// FilterByTime 按时间范围过滤
func (a *Analyzer) FilterByTime(start, end time.Time) []LogEntry {
	a.mu.Lock()
	defer a.mu.Unlock()

	var results []LogEntry

	for _, entry := range a.byTime {
		if (entry.Timestamp.IsZero() || entry.Timestamp.After(start)) &&
			(end.IsZero() || entry.Timestamp.Before(end)) {
			results = append(results, entry)
		}
	}

	return results
}

// GetErrorRate 获取错误率
func (a *Analyzer) GetErrorRate() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.stats.TotalLines == 0 {
		return 0
	}

	return float64(a.stats.ErrorCount) / float64(a.stats.TotalLines) * 100
}

// GetTopErrors 获取最常见的错误
func (a *Analyzer) GetTopErrors(n int) []ErrorCount {
	errorCounts := make(map[string]int)

	for _, entry := range a.errors {
		// 简化错误消息作为key
		key := entry.Message
		if len(key) > 100 {
			key = key[:100]
		}
		errorCounts[key]++
	}

	// 排序
	topErrors := make([]ErrorCount, 0, len(errorCounts))
	for msg, count := range errorCounts {
		topErrors = append(topErrors, ErrorCount{Message: msg, Count: count})
	}

	// 简单排序
	for i := 0; i < len(topErrors)-1; i++ {
		for j := i + 1; j < len(topErrors); j++ {
			if topErrors[j].Count > topErrors[i].Count {
				topErrors[i], topErrors[j] = topErrors[j], topErrors[i]
			}
		}
	}

	if n > len(topErrors) {
		n = len(topErrors)
	}

	return topErrors[:n]
}

// ErrorCount 错误统计
type ErrorCount struct {
	Message string
	Count   int
}

// AnalyzeTrends 分析趋势
func (a *Analyzer) AnalyzeTrends() *Trends {
	a.mu.Lock()
	defer a.mu.Unlock()

	return &Trends{
		TotalErrors:   a.stats.ErrorCount,
		TotalWarnings: a.stats.WarningCount,
		ErrorRate:     a.GetErrorRate(),
	}
}

// Trends 趋势
type Trends struct {
	TotalErrors   int
	TotalWarnings int
	ErrorRate     float64
}

// Export 导出日志
func (a *Analyzer) Export(filename string, format string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	switch format {
	case "json":
		return a.exportJSON(file)
	case "csv":
		return a.exportCSV(file)
	default:
		return a.exportText(file)
	}
}

// exportJSON 导出为JSON
func (a *Analyzer) exportJSON(w io.Writer) error {
	// 简化实现
	for _, entry := range a.byTime {
		w.Write([]byte(fmt.Sprintf("%s\n", entry.Raw)))
	}
	return nil
}

// exportCSV 导出为CSV
func (a *Analyzer) exportCSV(w io.Writer) error {
	// 写入表头
	w.Write([]byte("Timestamp,Level,Message\n"))

	// 写入数据
	for _, entry := range a.byTime {
		w.Write([]byte(fmt.Sprintf("%s,%s,%s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Level,
			strings.ReplaceAll(entry.Message, ",", ";"))))
	}

	return nil
}

// exportText 导出为文本
func (a *Analyzer) exportText(w io.Writer) error {
	for _, entry := range a.byTime {
		w.Write([]byte(entry.Raw + "\n"))
	}
	return nil
}

// PrintReport 打印报告
func (a *Analyzer) PrintReport() {
	stats := a.GetStats()

	fmt.Println("\n=== Log Analysis Report ===")
	fmt.Printf("Log File: %s\n", a.logFile)
	fmt.Printf("Total Lines: %d\n", stats.TotalLines)
	fmt.Printf("Errors: %d\n", stats.ErrorCount)
	fmt.Printf("Warnings: %d\n", stats.WarningCount)
	fmt.Printf("Info: %d\n", stats.InfoCount)
	fmt.Printf("Debug: %d\n", stats.DebugCount)
	fmt.Printf("Parse Errors: %d\n", stats.ParseErrors)

	if !stats.StartTime.IsZero() && !stats.EndTime.IsZero() {
		fmt.Printf("Time Range: %s to %s\n",
			stats.StartTime.Format(time.RFC3339),
			stats.EndTime.Format(time.RFC3339))
	}

	fmt.Printf("Error Rate: %.2f%%\n", a.GetErrorRate())

	// Top Errors
	fmt.Println("\n=== Top Errors ===")
	topErrors := a.GetTopErrors(5)
	for i, err := range topErrors {
		fmt.Printf("%d. [%d occurrences] %s\n", i+1, err.Count, err.Message)
	}

	fmt.Println("========================\n")
}

// Clear 清空数据
func (a *Analyzer) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.errors = make([]LogEntry, 0)
	a.warnings = make([]LogEntry, 0)
	a.byLevel = make(map[string][]LogEntry)
	a.byTime = make([]LogEntry, 0)
	a.stats = &Stats{}
}

// Watch 监控日志文件变化
func (a *Analyzer) Watch(interval time.Duration) <-chan *LogEntry {
	ch := make(chan *LogEntry)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(ch)

		lastPos := int64(0)

		for {
			select {
			case <-ticker.C:
				// 检查文件变化并读取新行
				file, err := os.Open(a.logFile)
				if err != nil {
					continue
				}

				file.Seek(lastPos, io.SeekStart)
				scanner := bufio.NewScanner(file)
				lineNum := 0

				for scanner.Scan() {
					line := scanner.Text()
					lineNum++

					entry, _ := a.parseLine(line, lineNum)
					if entry != nil {
						ch <- entry
						a.addEntry(entry)
					}
				}

				lastPos, _ = file.Seek(0, io.SeekCurrent)
				file.Close()
			}
		}
	}()

	return ch
}
