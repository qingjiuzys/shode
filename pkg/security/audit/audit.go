// Package audit 安全审计日志
package audit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

// Level 审计级别
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// Event 审计事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	User      string                 `json:"user"`
	IP        string                 `json:"ip"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Config 审计配置
type Config struct {
	Output   []string
	Format   string
	MinLevel Level
}

// Auditor 审计日志器
type Auditor struct {
	config  *Config
	outputs []Output
	mu      sync.Mutex
}

// Output 输出接口
type Output interface {
	Write(Event) error
	Close() error
}

// New 创建审计日志器
func New(config *Config) *Auditor {
	if config.Format == "" {
		config.Format = "json"
	}
	if config.MinLevel == 0 {
		config.MinLevel = InfoLevel
	}

	a := &Auditor{
		config:  config,
		outputs: make([]Output, 0),
	}

	// 初始化输出
	for _, output := range config.Output {
		switch output {
		case "stdout":
			a.outputs = append(a.outputs, &StdoutOutput{})
		case "stderr":
			a.outputs = append(a.outputs, &StderrOutput{})
		default:
			file, err := NewFileOutput(output)
			if err == nil {
				a.outputs = append(a.outputs, file)
			}
		}
	}

	return a
}

// Log 记录事件
func (a *Auditor) Log(event Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	event.ID = generateID()
	event.Timestamp = time.Now()

	// 写入所有输出
	for _, output := range a.outputs {
		if err := output.Write(event); err != nil {
			return err
		}
	}

	return nil
}

// Close 关闭审计日志器
func (a *Auditor) Close() error {
	for _, output := range a.outputs {
		output.Close()
	}
	return nil
}

// StdoutOutput 标准输出
type StdoutOutput struct{}

func (o *StdoutOutput) Write(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func (o *StdoutOutput) Close() error {
	return nil
}

// StderrOutput 错误输出
type StderrOutput struct{}

func (o *StderrOutput) Write(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, string(data))
	return nil
}

func (o *StderrOutput) Close() error {
	return nil
}

// FileOutput 文件输出
type FileOutput struct {
	file *os.File
	path string
	mu   sync.Mutex
}

// NewFileOutput 创建文件输出
func NewFileOutput(path string) (*FileOutput, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	return &FileOutput{file: file, path: path}, nil
}

func (o *FileOutput) Write(event Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = o.file.Write(append(data, '\n'))
	return err
}

func (o *FileOutput) Close() error {
	return o.file.Close()
}

// Middleware 审计中间件
func Middleware(auditor *Auditor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 记录请求
			event := Event{
				Type:     "http_request",
				IP:       getClientIP(r),
				Action:   r.Method,
				Resource: r.URL.Path,
				Metadata: map[string]interface{}{
					"user_agent": r.UserAgent(),
					"referer":    r.Referer(),
				},
			}

			// 获取用户信息
			if user := getUserFromContext(r); user != "" {
				event.User = user
			}

			// 执行请求
			next.ServeHTTP(w, r)

			// 记录响应
			auditor.Log(event)
		})
	}
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// getClientIP 获取客户端 IP
func getClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// getUserFromContext 从上下文获取用户
func getUserFromContext(r *http.Request) string {
	// 简化实现
	return ""
}

// Default 默认审计日志器
func Default() *Auditor {
	return New(&Config{
		Output:   []string{"stdout"},
		Format:   "json",
		MinLevel: InfoLevel,
	})
}
