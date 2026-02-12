// Package server 开发服务器
package server

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// Server 开发服务器
type Server struct {
	Port     string
	Host     string
	HotReload bool
	Proxy    string
	Env      string
	Command  []string
	mu       sync.RWMutex
	running  bool
	process  *exec.Cmd
}

// NewServer 创建服务器
func NewServer() *Server {
	return &Server{
		Port:     "8080",
		Host:     "localhost",
		HotReload: true,
		Env:      "development",
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("server already running")
	}
	s.running = true
	s.mu.Unlock()

	if s.HotReload {
		return s.startWithHotReload()
	}

	return s.startSimple()
}

// startSimple 简单启动
func (s *Server) startSimple() error {
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	fmt.Printf("Starting server on http://%s\n", addr)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from Shode development server!")
	})

	return http.ListenAndServe(addr, nil)
}

// startWithHotReload 带热重载的启动
func (s *Server) startWithHotReload() error {
	// 启动文件监听器
	go s.watchFiles()

	// 启动应用
	return s.buildAndRun()
}

// watchFiles 监听文件变化
func (s *Server) watchFiles() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.hasChanges() {
				fmt.Println("Changes detected, rebuilding...")
				s.restart()
			}
		}
	}
}

// hasChanges 检查是否有变化
func (s *Server) hasChanges() bool {
	// 简化实现：实际应该使用文件监听
	return false
}

// buildAndRun 构建并运行
func (s *Server) buildAndRun() error {
	fmt.Println("Building...")

	// 构建命令
	buildCmd := exec.Command("go", "build", "-o", "tmp/main", "cmd/*/main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		return err
	}

	// 运行命令
	s.mu.Lock()
	s.process = exec.Command("./tmp/main")
	s.process.Stdout = os.Stdout
	s.process.Stderr = os.Stderr
	s.mu.Unlock()

	return s.process.Run()
}

// restart 重启服务器
func (s *Server) restart() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 停止当前进程
	if s.process != nil && s.process.Process != nil {
		s.process.Process.Kill()
	}

	// 重新启动
	s.buildAndRun()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("server not running")
	}

	s.running = false

	if s.process != nil && s.process.Process != nil {
		return s.process.Process.Kill()
	}

	return nil
}

// SetPort 设置端口
func (s *Server) SetPort(port string) {
	s.Port = port
}

// SetHost 设置主机
func (s *Server) SetHost(host string) {
	s.Host = host
}

// EnableHotReload 启用热重载
func (s *Server) EnableHotReload(enable bool) {
	s.HotReload = enable
}

// RunDevServer 运行开发服务器
func RunDevServer() error {
	server := NewServer()
	return server.Start()
}

// RunDevServerWithPort 在指定端口运行开发服务器
func RunDevServerWithPort(port string) error {
	server := NewServer()
	server.SetPort(port)
	return server.Start()
}

// FindMainFile 查找 main 文件
func FindMainFile() (string, error) {
	// 搜索 cmd 目录
	cmdDir := "cmd"
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			mainFile := filepath.Join(cmdDir, entry.Name(), "main.go")
			if _, err := os.Stat(mainFile); err == nil {
				return mainFile, nil
			}
		}
	}

	return "", fmt.Errorf("main.go not found in cmd directory")
}

// IsProjectValid 检查项目是否有效
func IsProjectValid() bool {
	// 检查 go.mod
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return false
	}

	// 检查 cmd 目录
	if _, err := os.Stat("cmd"); os.IsNotExist(err) {
		return false
	}

	return true
}

// GetProjectInfo 获取项目信息
func GetProjectInfo() map[string]string {
	info := make(map[string]string)

	if name, err := getProjectName(); err == nil {
		info["name"] = name
	}

	if version, err := getProjectVersion(); err == nil {
		info["version"] = version
	}

	return info
}

// getProjectName 获取项目名
func getProjectName() (string, error) {
	// 从 go.mod 读取
	return "shode-project", nil
}

// getProjectVersion 获取版本号
func getProjectVersion() (string, error) {
	// 从配置或 main.go 读取
	return "1.0.0", nil
}

// BuildBinary 构建二进制文件
func BuildBinary(output string) error {
	fmt.Printf("Building binary: %s\n", output)

	cmd := exec.Command("go", "build", "-o", output, "cmd/*/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// RunBinary 运行二进制文件
func RunBinary(binPath string) error {
	fmt.Printf("Running binary: %s\n", binPath)

	cmd := exec.Command(binPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
