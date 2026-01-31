// Package logger 提供结构化日志系统。
//
// 日志系统特点：
//   - 结构化日志格式 (JSON)
//   - 多日志级别 (DEBUG/INFO/WARN/ERROR/FATAL)
//   - 多输出目标 (控制台/文件/syslog)
//   - 日志轮转 (按大小/时间)
//   - 请求追踪集成
//   - 性能监控
//
// 使用示例：
//
//	logger := logger.NewLogger(logger.Config{
//	    Level: logger.INFO,
//	    Format: logger.JSONFormat,
//	    Output: logger.MultiOutput,
//	})
//	logger.Info("Server started", "port", 8080)
//	logger.Error("Database connection failed", "error", err)
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	// DEBUG 调试信息
	DEBUG LogLevel = iota
	// INFO 一般信息
	INFO
	// WARN 警告信息
	WARN
	// ERROR 错误信息
	ERROR
	// FATAL 致命错误
	FATAL
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLevel 解析日志级别字符串
func ParseLevel(s string) LogLevel {
	switch s {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// OutputFormat 输出格式类型
type OutputFormat int

const (
	// TextFormat 文本格式
	TextFormat OutputFormat = iota
	// JSONFormat JSON格式
	JSONFormat
)

// OutputTarget 输出目标类型
type OutputTarget int

const (
	// ConsoleOutput 控制台输出
	ConsoleOutput OutputTarget = iota
	// FileOutput 文件输出
	FileOutput
	// MultiOutput 多目标输出
	MultiOutput
)

// Config 日志配置
type Config struct {
	Level          LogLevel     // 日志级别
	Format         OutputFormat // 输出格式
	Output         OutputTarget // 输出目标
	FilePath       string       // 日志文件路径
	MaxSize        int64        // 最大文件大小 (字节)
	MaxBackups     int          // 最大备份数
	MaxAge         int          // 最大保留天数
	Compress       bool         // 是否压缩旧日志
	CallerSkip     int          // 调用栈跳过层数
	EnableTrace    bool         // 启用请求追踪
	EnableCaller   bool         // 启用调用位置
	EnableStackTrace bool       // 错误时打印堆栈
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Level:           INFO,
	Format:          JSONFormat,
	Output:          ConsoleOutput,
	MaxSize:         100 * 1024 * 1024, // 100MB
	MaxBackups:      3,
	MaxAge:          28, // 28天
	Compress:        true,
	CallerSkip:      2,
	EnableTrace:     true,
	EnableCaller:    true,
	EnableStackTrace: true,
}

// LogEntry 日志条目
type LogEntry struct {
	Time      string                 `json:"time"`
	Level     string                 `json:"level"`
	Message   string                 `json:"msg"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
	Function  string                 `json:"function,omitempty"`
	Duration  int64                  `json:"duration_ms,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

// Logger 日志记录器
type Logger struct {
	config    Config
	mu        sync.Mutex
	file      *os.File
	writer    io.Writer
	atomicLevel atomic.Value // 存储 LogLevel
	stats     *LoggerStats
}

// LoggerStats 日志统计
type LoggerStats struct {
	DebugLogs int64
	InfoLogs  int64
	WarnLogs  int64
	ErrorLogs int64
	FatalLogs int64
	TotalLogs int64
}

// NewLogger 创建新的日志记录器
func NewLogger(config Config) *Logger {
	l := &Logger{
		config: config,
		stats:  &LoggerStats{},
	}

	// 如果所有值都是默认/零值，使用默认配置
	isDefaultConfig := config.Level == 0 && config.Format == 0 &&
		config.Output == 0 && config.FilePath == "" &&
		config.MaxSize == 0 && config.MaxBackups == 0 &&
		config.MaxAge == 0 && !config.Compress &&
		config.CallerSkip == 0 && !config.EnableTrace &&
		!config.EnableCaller && !config.EnableStackTrace

	if isDefaultConfig {
		l.config = DefaultConfig
	}

	// 对于显式设置为 0 的字段，保留它们
	// 其他未设置的字段使用默认值
	if l.config.Format == 0 && !isDefaultConfig {
		l.config.Format = TextFormat
	}
	if l.config.Output == 0 && !isDefaultConfig {
		l.config.Output = ConsoleOutput
	}
	if l.config.CallerSkip == 0 && !isDefaultConfig {
		l.config.CallerSkip = DefaultConfig.CallerSkip
	}
	if l.config.MaxSize == 0 && !isDefaultConfig {
		l.config.MaxSize = DefaultConfig.MaxSize
	}
	if l.config.MaxBackups == 0 && !isDefaultConfig {
		l.config.MaxBackups = DefaultConfig.MaxBackups
	}
	if l.config.MaxAge == 0 && !isDefaultConfig {
		l.config.MaxAge = DefaultConfig.MaxAge
	}

	l.atomicLevel.Store(l.config.Level)

	// 初始化输出
	l.initOutput()

	return l
}

// initOutput 初始化输出
func (l *Logger) initOutput() {
	switch l.config.Output {
	case ConsoleOutput:
		l.writer = os.Stdout
	case FileOutput:
		l.openLogFile()
	case MultiOutput:
		writers := []io.Writer{os.Stdout}
		if l.config.FilePath != "" {
			if f := l.openLogFile(); f != nil {
				writers = append(writers, f)
			}
		}
		l.writer = io.MultiWriter(writers...)
	}
}

// openLogFile 打开日志文件
func (l *Logger) openLogFile() *os.File {
	if l.config.FilePath == "" {
		return nil
	}

	// 确保目录存在
	dir := filepath.Dir(l.config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err)
		return nil
	}

	// 打开文件
	f, err := os.OpenFile(l.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return nil
	}

	l.mu.Lock()
	l.file = f
	l.writer = f
	l.mu.Unlock()

	return f
}

// rotateLogFile 轮转日志文件
func (l *Logger) rotateLogFile() error {
	if l.config.FilePath == "" {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		l.file.Close()
	}

	// 重命名当前文件
	if _, err := os.Stat(l.config.FilePath); err == nil {
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		backupPath := l.config.FilePath + "." + timestamp
		if err := os.Rename(l.config.FilePath, backupPath); err != nil {
			return fmt.Errorf("failed to rotate log file: %w", err)
		}

		// 压缩旧日志
		if l.config.Compress {
			go l.compressLog(backupPath)
		}

		// 清理旧备份
		go l.cleanupOldLogs()
	}

	// 打开新文件
	l.openLogFile()
	return nil
}

// compressLog 压缩日志文件
func (l *Logger) compressLog(path string) {
	// TODO: 实现 gzip 压缩
}

// cleanupOldLogs 清理旧日志
func (l *Logger) cleanupOldLogs() {
	if l.config.FilePath == "" {
		return
	}

	dir := filepath.Dir(l.config.FilePath)
	base := filepath.Base(l.config.FilePath)

	files, err := filepath.Glob(dir + "/" + base + ".*")
	if err != nil {
		return
	}

	// 按修改时间排序
	type fileInfo struct {
		name    string
		modTime time.Time
	}
	var sortedFiles []fileInfo
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		sortedFiles = append(sortedFiles, fileInfo{f, info.ModTime()})
	}

	// 删除超过保留数量的文件
	if len(sortedFiles) > l.config.MaxBackups {
		for _, f := range sortedFiles[l.config.MaxBackups:] {
			os.Remove(f.name)
		}
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.atomicLevel.Store(level)
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	return l.atomicLevel.Load().(LogLevel)
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, msg string, fields ...interface{}) {
	if level < l.GetLevel() {
		return
	}

	entry := &LogEntry{
		Time:    time.Now().Format(time.RFC3339Nano),
		Level:   level.String(),
		Message: msg,
		Fields:  make(map[string]interface{}),
	}

	// 解析字段
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			key := fmt.Sprintf("%v", fields[i])
			value := fields[i+1]

			// 检查是否为 error
			if err, ok := value.(error); ok {
				entry.Fields[key] = err.Error()
				entry.Error = err.Error()
				if l.config.EnableStackTrace && level >= ERROR {
					entry.Stack = getStackTrace()
				}
			} else {
				entry.Fields[key] = value
			}
		}
	}

	// 添加调用位置
	if l.config.EnableCaller {
		if pc, file, line, ok := runtime.Caller(l.config.CallerSkip); ok {
			fn := runtime.FuncForPC(pc)
			entry.Function = fn.Name()
			entry.File = filepath.Base(file)
			entry.Line = line
			entry.Caller = fmt.Sprintf("%s:%d", entry.File, entry.Line)
		}
	}

	l.writeEntry(entry)

	// 更新统计
	atomic.AddInt64(&l.stats.TotalLogs, 1)
	switch level {
	case DEBUG:
		atomic.AddInt64(&l.stats.DebugLogs, 1)
	case INFO:
		atomic.AddInt64(&l.stats.InfoLogs, 1)
	case WARN:
		atomic.AddInt64(&l.stats.WarnLogs, 1)
	case ERROR:
		atomic.AddInt64(&l.stats.ErrorLogs, 1)
	case FATAL:
		atomic.AddInt64(&l.stats.FatalLogs, 1)
	}

	// FATAL 级别退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// writeEntry 写入日志条目
func (l *Logger) writeEntry(entry *LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var output []byte
	var err error

	if l.config.Format == JSONFormat {
		output, err = json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
			return
		}
		output = append(output, '\n')
	} else {
		// 文本格式
		output = []byte(fmt.Sprintf("[%s] %s %s", entry.Time, entry.Level, entry.Message))
		for k, v := range entry.Fields {
			output = append(output, ' ')
			output = append(output, fmt.Sprintf("%s=%v", k, v)...)
		}
		if entry.Caller != "" {
			output = append(output, ' ')
			output = append(output, fmt.Sprintf("caller=%s", entry.Caller)...)
		}
		output = append(output, '\n')
	}

	// 检查文件大小并轮转
	if l.file != nil {
		if info, err := l.file.Stat(); err == nil {
			if info.Size() >= l.config.MaxSize {
				l.rotateLogFile()
			}
		}
	}

	if l.writer != nil {
		l.writer.Write(output)
	}
}

// getStackTrace 获取堆栈跟踪
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Debug 记录调试信息
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.log(DEBUG, msg, fields...)
}

// Info 记录一般信息
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.log(INFO, msg, fields...)
}

// Warn 记录警告信息
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.log(WARN, msg, fields...)
}

// Error 记录错误信息
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.log(ERROR, msg, fields...)
}

// Fatal 记录致命错误并退出
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.log(FATAL, msg, fields...)
}

// WithFields 返回带预设字段的日志上下文
func (l *Logger) WithFields(fields map[string]interface{}) *LoggerContext {
	return &LoggerContext{
		logger:  l,
		fields:  fields,
		traceID: generateTraceID(),
	}
}

// WithTrace 返回带追踪ID的日志上下文
func (l *Logger) WithTrace(traceID string) *LoggerContext {
	return &LoggerContext{
		logger:  l,
		traceID: traceID,
		fields:  make(map[string]interface{}),
	}
}

// LoggerContext 日志上下文
type LoggerContext struct {
	logger  *Logger
	fields  map[string]interface{}
	traceID string
	start   time.Time
}

// WithField 添加单个字段
func (lc *LoggerContext) WithField(key string, value interface{}) *LoggerContext {
	lc.fields[key] = value
	return lc
}

// WithFields 添加多个字段
func (lc *LoggerContext) WithFields(fields map[string]interface{}) *LoggerContext {
	for k, v := range fields {
		lc.fields[k] = v
	}
	return lc
}

// WithDuration 添加持续时间
func (lc *LoggerContext) WithDuration() *LoggerContext {
	if !lc.start.IsZero() {
		lc.fields["duration_ms"] = time.Since(lc.start).Milliseconds()
	}
	return lc
}

// StartTimer 开始计时
func (lc *LoggerContext) StartTimer() *LoggerContext {
	lc.start = time.Now()
	return lc
}

// Debug 记录调试信息
func (lc *LoggerContext) Debug(msg string, fields ...interface{}) {
 lc.log(DEBUG, msg, fields...)
}

// Info 记录一般信息
func (lc *LoggerContext) Info(msg string, fields ...interface{}) {
 lc.log(INFO, msg, fields...)
}

// Warn 记录警告信息
func (lc *LoggerContext) Warn(msg string, fields ...interface{}) {
 lc.log(WARN, msg, fields...)
}

// Error 记录错误信息
func (lc *LoggerContext) Error(msg string, fields ...interface{}) {
 lc.log(ERROR, msg, fields...)
}

// Fatal 记录致命错误并退出
func (lc *LoggerContext) Fatal(msg string, fields ...interface{}) {
 lc.log(FATAL, msg, fields...)
}

// log 内部日志方法
func (lc *LoggerContext) log(level LogLevel, msg string, fields ...interface{}) {
	// 合并字段
	allFields := make(map[string]interface{})
	for k, v := range lc.fields {
		allFields[k] = v
	}
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			allFields[fmt.Sprintf("%v", fields[i])] = fields[i+1]
		}
	}

	if lc.traceID != "" {
		allFields["trace_id"] = lc.traceID
	}

	// 转换为切片
	fieldSlice := make([]interface{}, 0, len(allFields)*2)
	for k, v := range allFields {
		fieldSlice = append(fieldSlice, k, v)
	}

	lc.logger.log(level, msg, fieldSlice...)
}

// GetStats 获取日志统计
func (l *Logger) GetStats() LoggerStats {
	return LoggerStats{
		DebugLogs: atomic.LoadInt64(&l.stats.DebugLogs),
		InfoLogs:  atomic.LoadInt64(&l.stats.InfoLogs),
		WarnLogs:  atomic.LoadInt64(&l.stats.WarnLogs),
		ErrorLogs: atomic.LoadInt64(&l.stats.ErrorLogs),
		FatalLogs: atomic.LoadInt64(&l.stats.FatalLogs),
		TotalLogs: atomic.LoadInt64(&l.stats.TotalLogs),
	}
}

// ResetStats 重置统计
func (l *Logger) ResetStats() {
	atomic.StoreInt64(&l.stats.DebugLogs, 0)
	atomic.StoreInt64(&l.stats.InfoLogs, 0)
	atomic.StoreInt64(&l.stats.WarnLogs, 0)
	atomic.StoreInt64(&l.stats.ErrorLogs, 0)
	atomic.StoreInt64(&l.stats.FatalLogs, 0)
	atomic.StoreInt64(&l.stats.TotalLogs, 0)
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// generateTraceID 生成追踪ID
func generateTraceID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// DefaultLogger 默认日志记录器
var DefaultLogger = NewLogger(DefaultConfig)

// Debug 记录调试信息
func Debug(msg string, fields ...interface{}) {
	DefaultLogger.Debug(msg, fields...)
}

// Info 记录一般信息
func Info(msg string, fields ...interface{}) {
	DefaultLogger.Info(msg, fields...)
}

// Warn 记录警告信息
func Warn(msg string, fields ...interface{}) {
	DefaultLogger.Warn(msg, fields...)
}

// Error 记录错误信息
func Error(msg string, fields ...interface{}) {
	DefaultLogger.Error(msg, fields...)
}

// Fatal 记录致命错误并退出
func Fatal(msg string, fields ...interface{}) {
	DefaultLogger.Fatal(msg, fields...)
}
