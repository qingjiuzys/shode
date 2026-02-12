// Package logging 结构化日志系统
package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// String 返回日志级别字符串
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	case FatalLevel:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Config 日志配置
type Config struct {
	Level      Level
	Format     string
	Output     []string
	TimeFormat string
	Color      bool
	Caller     bool
}

// Logger 日志器
type Logger struct {
	config    Config
	mu        sync.RWMutex
	out       []io.Writer
	fields    map[string]interface{}
	callerSkip int
}

// NewLogger 创建日志器
func NewLogger(config Config) *Logger {
	logger := &Logger{
		config:     config,
		out:        make([]io.Writer, 0),
		fields:     make(map[string]interface{}),
		callerSkip: 1,
	}

	// 初始化输出
	for _, output := range config.Output {
		switch output {
		case "stdout":
			logger.out = append(logger.out, os.Stdout)
		case "stderr":
			logger.out = append(logger.out, os.Stderr)
		default:
			file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err == nil {
				logger.out = append(logger.out, file)
			}
		}
	}

	// 如果没有输出，默认使用 stdout
	if len(logger.out) == 0 {
		logger.out = append(logger.out, os.Stdout)
	}

	return logger
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Level = level
}

// SetFields 设置字段
func (l *Logger) SetFields(fields map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for k, v := range fields {
		l.fields[k] = v
	}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		config:     l.config,
		out:        l.out,
		fields:     make(map[string]interface{}),
		callerSkip: l.callerSkip,
	}

	// 复制现有字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 添加新字段
	newLogger.fields[key] = value

	return newLogger
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		config:     l.config,
		out:        l.out,
		fields:     make(map[string]interface{}),
		callerSkip: l.callerSkip,
	}

	// 复制现有字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// 添加新字段
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// WithCaller 设置调用者跳过层数
func (l *Logger) WithCaller(skip int) *Logger {
	newLogger := &Logger{
		config:     l.config,
		out:        l.out,
		fields:     make(map[string]interface{}),
		callerSkip: skip + 1,
	}

	// 复制现有字段
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// Log 记录日志
func (l *Logger) Log(level Level, msg string, fields ...map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 检查日志级别
	if level < l.config.Level {
		return
	}

	// 创建日志记录
	record := make(map[string]interface{})

	// 添加时间
	timeFormat := l.config.TimeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}
	record["time"] = time.Now().Format(timeFormat)

	// 添加级别
	record["level"] = level.String()

	// 添加消息
	record["message"] = msg

	// 添加字段
	for k, v := range l.fields {
		record[k] = v
	}

	// 添加额外字段
	for _, f := range fields {
		for k, v := range f {
			record[k] = v
		}
	}

	// 添加调用者信息
	if l.config.Caller {
		_, file, line, ok := runtime.Caller(l.callerSkip)
		if ok {
			record["caller"] = fmt.Sprintf("%s:%d", file, line)
		}
	}

	// 格式化输出
	var output []byte
	switch l.config.Format {
	case "json":
		output, _ = json.Marshal(record)
	default:
		output = []byte(fmt.Sprintf("%s %s %s", record["time"], record["level"], record["message"]))
		if len(record) > 3 {
			output = append(output, ' ')
			for k, v := range record {
				if k != "time" && k != "level" && k != "message" {
					output = append(output, fmt.Sprintf("%s=%v ", k, v)...)
				}
			}
		}
	}

	// 写入输出
	for _, w := range l.out {
		w.Write(output)
		w.Write([]byte("\n"))
	}

	// Fatal 级别退出程序
	if level == FatalLevel {
		os.Exit(1)
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.Log(DebugLevel, msg, fields...)
}

// Info 记录信息日志
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.Log(InfoLevel, msg, fields...)
}

// Warn 记录警告日志
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.Log(WarnLevel, msg, fields...)
}

// Error 记录错误日志
func (l *Logger) Error(msg string, fields ...map[string]interface{}) {
	l.Log(ErrorLevel, msg, fields...)
}

// Fatal 记录致命错误日志
func (l *Logger) Fatal(msg string, fields ...map[string]interface{}) {
	l.Log(FatalLevel, msg, fields...)
}

// ContextLogger 上下文日志器
type ContextLogger struct {
	logger *Logger
	ctx    context.Context
}

// NewContextLogger 创建上下文日志器
func NewContextLogger(ctx context.Context, logger *Logger) *ContextLogger {
	return &ContextLogger{
		logger: logger,
		ctx:    ctx,
	}
}

// WithContext 创建带上下文的日志器
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return NewContextLogger(ctx, l)
}

// Debug 记录调试日志
func (cl *ContextLogger) Debug(msg string, fields ...map[string]interface{}) {
	cl.logger.Debug(msg, cl.addContextFields(fields...)...)
}

// Info 记录信息日志
func (cl *ContextLogger) Info(msg string, fields ...map[string]interface{}) {
	cl.logger.Info(msg, cl.addContextFields(fields...)...)
}

// Warn 记录警告日志
func (cl *ContextLogger) Warn(msg string, fields ...map[string]interface{}) {
	cl.logger.Warn(msg, cl.addContextFields(fields...)...)
}

// Error 记录错误日志
func (cl *ContextLogger) Error(msg string, fields ...map[string]interface{}) {
	cl.logger.Error(msg, cl.addContextFields(fields...)...)
}

// Fatal 记录致命错误日志
func (cl *ContextLogger) Fatal(msg string, fields ...map[string]interface{}) {
	cl.logger.Fatal(msg, cl.addContextFields(fields...)...)
}

// addContextFields 添加上下文字段
func (cl *ContextLogger) addContextFields(fields ...map[string]interface{}) []map[string]interface{} {
	// 从上下文中提取信息
	// 例如：请求 ID、用户 ID 等
	contextFields := make(map[string]interface{})

	// 添加请求 ID
	if requestID := cl.ctx.Value("request_id"); requestID != nil {
		contextFields["request_id"] = requestID
	}

	// 添加用户 ID
	if userID := cl.ctx.Value("user_id"); userID != nil {
		contextFields["user_id"] = userID
	}

	// 添加追踪 ID
	if traceID := cl.ctx.Value("trace_id"); traceID != nil {
		contextFields["trace_id"] = traceID
	}

	// 合并字段
	if len(contextFields) > 0 {
		result := make([]map[string]interface{}, len(fields)+1)
		result[0] = contextFields
		copy(result[1:], fields)
		return result
	}

	return fields
}

// LogMiddleware 日志中间件
type LogMiddleware struct {
	logger *Logger
}

// NewLogMiddleware 创建日志中间件
func NewLogMiddleware(logger *Logger) *LogMiddleware {
	return &LogMiddleware{logger: logger}
}

// Wrap 包装函数
func (lm *LogMiddleware) Wrap(name string, fn func() error) func() error {
	return func() error {
		start := time.Now()
		err := fn()

		fields := map[string]interface{}{
			"duration": time.Since(start).String(),
			"function": name,
		}

		if err != nil {
			fields["error"] = err.Error()
			lm.logger.Error("Function failed", fields)
		} else {
			lm.logger.Info("Function completed", fields)
		}

		return err
	}
}

// PerformanceLogger 性能日志器
type PerformanceLogger struct {
	logger *Logger
}

// NewPerformanceLogger 创建性能日志器
func NewPerformanceLogger(logger *Logger) *PerformanceLogger {
	return &PerformanceLogger{logger: logger}
}

// Time 计时
func (pl *PerformanceLogger) Time(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		pl.logger.Info("Performance",
			map[string]interface{}{
				"operation": name,
				"duration":  duration.String(),
			})
	}
}

// AuditLogger 审计日志器
type AuditLogger struct {
	logger *Logger
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger(logger *Logger) *AuditLogger {
	return &AuditLogger{logger: logger}
}

// Log 记录审计日志
func (al *AuditLogger) Log(action, user, resource string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"action":   action,
		"user":     user,
		"resource": resource,
	}

	for k, v := range details {
		fields[k] = v
	}

	al.logger.Info("Audit", fields)
}

// ErrorLogger 错误日志器
type ErrorLogger struct {
	logger *Logger
}

// NewErrorLogger 创建错误日志器
func NewErrorLogger(logger *Logger) *ErrorLogger {
	return &ErrorLogger{logger: logger}
}

// Log 记录错误日志
func (el *ErrorLogger) Log(err error, fields ...map[string]interface{}) {
	allFields := make([]map[string]interface{}, len(fields)+1)
	allFields[0] = map[string]interface{}{
		"error": err.Error(),
		"type":  fmt.Sprintf("%T", err),
	}
	copy(allFields[1:], fields)

	el.logger.Error("Error occurred", allFields...)
}

// Global logger instance
var globalLogger *Logger

// Init 初始化全局日志器
func Init(config Config) {
	globalLogger = NewLogger(config)
}

// SetLevel 设置全局日志级别
func SetLevel(level Level) {
	if globalLogger != nil {
		globalLogger.SetLevel(level)
	}
}

// Debug 记录调试日志
func Debug(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(msg, fields...)
	}
}

// Info 记录信息日志
func Info(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(msg, fields...)
	}
}

// Warn 记录警告日志
func Warn(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(msg, fields...)
	}
}

// Error 记录错误日志
func Error(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误日志
func Fatal(msg string, fields ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(msg, fields...)
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *Logger {
	if globalLogger != nil {
		return globalLogger.WithField(key, value)
	}
	return nil
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) *Logger {
	if globalLogger != nil {
		return globalLogger.WithFields(fields)
	}
	return nil
}

// WithContext 创建带上下文的日志器
func WithContext(ctx context.Context) *ContextLogger {
	if globalLogger != nil {
		return globalLogger.WithContext(ctx)
	}
	return nil
}
