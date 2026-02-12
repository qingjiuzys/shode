package environment

import (
	"os"
	"strings"
	"sync"
)

// EnvironmentManager 管理环境变量
type EnvironmentManager struct {
	vars map[string]string
	mu   sync.RWMutex
}

// NewEnvironmentManager 创建环境管理器
func NewEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{
		vars: make(map[string]string),
	}
}

// Get 获取环境变量
func (em *EnvironmentManager) Get(key string) string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if val, ok := em.vars[key]; ok {
		return val
	}
	return os.Getenv(key)
}

// Set 设置环境变量
func (em *EnvironmentManager) Set(key, value string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.vars[key] = value
	os.Setenv(key, value)
}

// Export 批量设置环境变量
func (em *EnvironmentManager) Export(envs map[string]string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	for k, v := range envs {
		em.vars[k] = v
		os.Setenv(k, v)
	}
}

// GetAll 获取所有环境变量
func (em *EnvironmentManager) GetAll() map[string]string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	result := make(map[string]string)
	for k, v := range em.vars {
		result[k] = v
	}
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			if _, exists := result[pair[0]]; !exists {
				result[pair[0]] = pair[1]
			}
		}
	}
	return result
}

// Clear 清除自定义环境变量
func (em *EnvironmentManager) Clear() {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.vars = make(map[string]string)
}

// Clone 克隆环境管理器
func (em *EnvironmentManager) Clone() *EnvironmentManager {
	em.mu.RLock()
	defer em.mu.RUnlock()

	newEnv := NewEnvironmentManager()
	for k, v := range em.vars {
		newEnv.vars[k] = v
	}
	return newEnv
}

// GetWorkingDir 获取工作目录
func (em *EnvironmentManager) GetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "/"
	}
	return dir
}

// SetWorkingDir 设置工作目录
func (em *EnvironmentManager) SetWorkingDir(dir string) error {
	return os.Chdir(dir)
}

// GetHome 获取用户主目录
func (em *EnvironmentManager) GetHome() string {
	return os.Getenv("HOME")
}

// GetUser 获取用户名
func (em *EnvironmentManager) GetUser() string {
	return os.Getenv("USER")
}

// GetPath 获取 PATH 环境变量
func (em *EnvironmentManager) GetPath() []string {
	path := em.Get("PATH")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ":")
}

// AddToPath 添加到 PATH
func (em *EnvironmentManager) AddToPath(dir string) {
	paths := em.GetPath()
	paths = append([]string{dir}, paths...)
	em.Set("PATH", strings.Join(paths, ":"))
}

// SetEnv 批量设置环境变量
func (em *EnvironmentManager) SetEnv(key, value string) error {
	em.Set(key, value)
	return nil
}

// GetAllEnv 获取所有环境变量（别名）
func (em *EnvironmentManager) GetAllEnv() map[string]string {
	return em.GetAll()
}

// GetHomeDir 获取主目录（别名）
func (em *EnvironmentManager) GetHomeDir() string {
	return em.GetHome()
}

// ChangeDir 切换目录（别名）
func (em *EnvironmentManager) ChangeDir(dir string) error {
	return em.SetWorkingDir(dir)
}

// GetEnv 获取环境变量（别名）
func (em *EnvironmentManager) GetEnv(key string) string {
	return em.Get(key)
}

// UnsetEnv 删除环境变量
func (em *EnvironmentManager) UnsetEnv(key string) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	delete(em.vars, key)
	return nil
}
