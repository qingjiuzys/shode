package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// LinkManager 管理本地包链接
type LinkManager struct {
	linksFile string
	links     map[string]string // packageName -> localPath
	mu        sync.RWMutex
}

// LinkEntry 表示一个链接条目
type LinkEntry struct {
	PackageName string `json:"packageName"`
	LocalPath   string `json:"localPath"`
}

// NewLinkManager 创建新的链接管理器
func NewLinkManager(projectRoot string) *LinkManager {
	linksFile := filepath.Join(projectRoot, "shode-links.json")
	lm := &LinkManager{
		linksFile: linksFile,
		links:     make(map[string]string),
	}
	lm.load()
	return lm
}

// load 从文件加载链接配置
func (lm *LinkManager) load() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	data, err := os.ReadFile(lm.linksFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在是正常情况
		}
		return fmt.Errorf("读取链接文件失败: %w", err)
	}

	var entries []LinkEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("解析链接文件失败: %w", err)
	}

	lm.links = make(map[string]string)
	for _, entry := range entries {
		lm.links[entry.PackageName] = entry.LocalPath
	}

	return nil
}

// save 保存链接配置到文件
func (lm *LinkManager) save() error {
	entries := make([]LinkEntry, 0, len(lm.links))
	for pkg, path := range lm.links {
		entries = append(entries, LinkEntry{
			PackageName: pkg,
			LocalPath:   path,
		})
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化链接配置失败: %w", err)
	}

	if err := os.WriteFile(lm.linksFile, data, 0644); err != nil {
		return fmt.Errorf("写入链接文件失败: %w", err)
	}

	return nil
}

// Link 创建本地包链接
func (lm *LinkManager) Link(packageName, localPath string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// 验证本地路径存在
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("本地路径不存在: %s", localPath)
	}

	// 验证是一个有效的 Shode 包（包含 package.json）
	pkgJsonPath := filepath.Join(localPath, "package.json")
	if _, err := os.Stat(pkgJsonPath); os.IsNotExist(err) {
		return fmt.Errorf("不是有效的 Shode 包: %s (缺少 package.json)", localPath)
	}

	// 检查包名是否匹配
	data, err := os.ReadFile(pkgJsonPath)
	if err != nil {
		return fmt.Errorf("读取 package.json 失败: %w", err)
	}

	var pkgJson struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkgJson); err != nil {
		return fmt.Errorf("解析 package.json 失败: %w", err)
	}

	if pkgJson.Name != packageName {
		return fmt.Errorf("包名不匹配: 期望 %s, 实际 %s", packageName, pkgJson.Name)
	}

	// 创建链接
	lm.links[packageName] = localPath

	if err := lm.save(); err != nil {
		return err
	}

	return nil
}

// Unlink 移除本地包链接
func (lm *LinkManager) Unlink(packageName string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if _, exists := lm.links[packageName]; !exists {
		return fmt.Errorf("链接不存在: %s", packageName)
	}

	delete(lm.links, packageName)

	if err := lm.save(); err != nil {
		return err
	}

	return nil
}

// GetLink 获取包的链接路径
func (lm *LinkManager) GetLink(packageName string) (string, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	path, exists := lm.links[packageName]
	return path, exists
}

// ListLinks 列出所有链接
func (lm *LinkManager) ListLinks() []LinkEntry {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	entries := make([]LinkEntry, 0, len(lm.links))
	for pkg, path := range lm.links {
		entries = append(entries, LinkEntry{
			PackageName: pkg,
			LocalPath:   path,
		})
	}

	return entries
}

// IsLinked 检查包是否已链接
func (lm *LinkManager) IsLinked(packageName string) bool {
	_, exists := lm.GetLink(packageName)
	return exists
}

// ResolveLink 解析包路径（优先返回链接路径）
func (lm *LinkManager) ResolveLink(packageName, modulesPath string) string {
	if linkPath, exists := lm.GetLink(packageName); exists {
		return linkPath
	}

	// 返回默认的 modules 路径
	return filepath.Join(modulesPath, packageName)
}
