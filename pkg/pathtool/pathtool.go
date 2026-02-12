// Package pathtool 提供路径处理工具
package pathtool

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Join 连接路径
func Join(elements ...string) string {
	return filepath.Join(elements...)
}

// Clean 清理路径
func Clean(p string) string {
	return filepath.Clean(p)
}

// Abs 获取绝对路径
func Abs(p string) (string, error) {
	return filepath.Abs(p)
}

// Rel 获取相对路径
func Rel(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// Dir 获取目录路径
func Dir(p string) string {
	return filepath.Dir(p)
}

// Base 获取基础名称
func Base(p string) string {
	return filepath.Base(p)
}

// Ext 获取扩展名
func Ext(p string) string {
	return filepath.Ext(p)
}

// Split 分割路径
func Split(p string) (dir, file string) {
	return filepath.Split(p)
}

// Match 匹配路径
func Match(pattern, name string) (matched bool, err error) {
	return filepath.Match(pattern, name)
}

// Glob 查找匹配的文件
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Walk 遍历目录树
func Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// WalkDir 遍历目录树（更高效）
func WalkDir(root string, walkFn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, walkFn)
}

// Exists 检查路径是否存在
func Exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// IsFile 检查是否为文件
func IsFile(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir 检查是否为目录
func IsDir(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsAbs 检查是否为绝对路径
func IsAbs(p string) bool {
	return filepath.IsAbs(p)
}

// IsReadable 检查是否可读
func IsReadable(p string) bool {
	file, err := os.Open(p)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsWritable 检查是否可写
func IsWritable(p string) bool {
	file, err := os.OpenFile(p, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// IsExecutable 检查是否可执行
func IsExecutable(p string) bool {
	info, err := os.Stat(p)
	if err != nil {
		return false
	}

	// 检查权限位
	mode := info.Mode()
	return mode.Perm()&0111 != 0
}

// GetSize 获取文件大小
func GetSize(p string) (int64, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetModTime 获取修改时间
func GetModTime(p string) (int64, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	return info.ModTime().Unix(), nil
}

// Create 创建文件
func Create(p string) (*os.File, error) {
	return os.Create(p)
}

// CreateDir 创建目录
func CreateDir(p string) error {
	return os.MkdirAll(p, 0755)
}

// Remove 删除文件或目录
func Remove(p string) error {
	return os.RemoveAll(p)
}

// Rename 重命名文件或目录
func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

// Copy 复制文件
func Copy(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// CopyDir 复制目录
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return Copy(path, dstPath)
	})
}

// Move 移动文件或目录
func Move(src, dst string) error {
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// 如果重命名失败，尝试复制后删除
	if err := Copy(src, dst); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// Normalize 规范化路径（使用正斜杠）
func Normalize(p string) string {
	if runtime.GOOS == "windows" {
		return strings.ReplaceAll(p, "\\", "/")
	}
	return p
}

// ToSlash 转换为正斜杠
func ToSlash(p string) string {
	return filepath.ToSlash(p)
}

// FromSlash 从正斜杠转换
func FromSlash(p string) string {
	return filepath.FromSlash(p)
}

// GetExtWithoutDot 获取扩展名（不带点）
func GetExtWithoutDot(p string) string {
	ext := filepath.Ext(p)
	if len(ext) > 0 {
		return ext[1:]
	}
	return ""
}

// GetNameWithoutExt 获取文件名（不带扩展名）
func GetNameWithoutExt(p string) string {
	base := filepath.Base(p)
	ext := filepath.Ext(p)
	if len(ext) > 0 {
		return base[:len(base)-len(ext)]
	}
	return base
}

// ReplaceExt 替换扩展名
func ReplaceExt(p, newExt string) string {
	base := p
	ext := filepath.Ext(p)
	if len(ext) > 0 {
		base = p[:len(p)-len(ext)]
	}
	if !strings.HasPrefix(newExt, ".") {
		newExt = "." + newExt
	}
	return base + newExt
}

// AppendToName 在文件名后添加后缀
func AppendToName(p, suffix string) string {
	dir := filepath.Dir(p)
	base := filepath.Base(p)
	ext := filepath.Ext(p)
	name := base[:len(base)-len(ext)]

	return filepath.Join(dir, name+suffix+ext)
}

// GetParent 获取父目录
func GetParent(p string) string {
	return filepath.Dir(p)
}

// GetParents 获取所有父目录
func GetParents(p string) []string {
	parents := []string{}

	for {
		parent := filepath.Dir(p)
		if parent == p || parent == "" || parent == "." {
			break
		}
		parents = append(parents, parent)
		p = parent
	}

	return parents
}

// CommonPrefix 获取公共前缀
func CommonPrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}

	if len(paths) == 1 {
		return paths[0]
	}

	prefix := paths[0]
	for _, p := range paths[1:] {
		prefix = commonPrefixTwo(prefix, p)
		if prefix == "" {
			break
		}
	}

	return prefix
}

func commonPrefixTwo(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	i := 0
	for i < minLen && a[i] == b[i] {
		i++
	}

	return a[:i]
}

// IsSubPath 检查是否为子路径
func IsSubPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// UniquePath 生成唯一路径
func UniquePath(p string) string {
	base := p
	ext := filepath.Ext(p)
	name := base[:len(base)-len(ext)]

	counter := 1
	for Exists(p) {
		p = fmt.Sprintf("%s_%d%s", name, counter, ext)
		counter++
	}

	return p
}

// EnsureDir 确保目录存在
func EnsureDir(p string) error {
	if !Exists(p) {
		return os.MkdirAll(p, 0755)
	}
	return nil
}

// EnsureFileDir 确保文件的目录存在
func EnsureFileDir(filePath string) error {
	dir := filepath.Dir(filePath)
	return EnsureDir(dir)
}

// ExpandUser 扩展用户目录（~）
func ExpandUser(p string) string {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, p[1:])
		}
	}
	return p
}

// ExpandEnv 扩展环境变量
func ExpandEnv(p string) string {
	return os.ExpandEnv(p)
}

// ShortenPath 缩短路径（用于显示）
func ShortenPath(p string, maxLen int) string {
	if len(p) <= maxLen {
		return p
	}

	// 尝试保留开头和结尾
	if maxLen < 10 {
		return p[:maxLen]
	}

	parts := strings.Split(p, string(filepath.Separator))
	if len(parts) <= 2 {
		if len(p) > maxLen {
			return "..." + p[len(p)-maxLen+3:]
		}
		return p
	}

	// 保留第一个和最后一个部分
	first := parts[0]
	last := parts[len(parts)-1]

	remaining := maxLen - len(first) - len(last) - 4 // 4 for "..."
	if remaining < 0 {
		remaining = 0
	}

	return first + string(filepath.Separator) + strings.Repeat(".", remaining) + string(filepath.Separator) + last
}

// Resolve 解析路径（处理符号链接）
func Resolve(p string) (string, error) {
	return filepath.EvalSymlinks(p)
}

// GetWorkingDir 获取当前工作目录
func GetWorkingDir() (string, error) {
	return os.Getwd()
}

// SetWorkingDir 设置当前工作目录
func SetWorkingDir(dir string) error {
	return os.Chdir(dir)
}

// GetTempDir 获取临时目录
func GetTempDir() string {
	return os.TempDir()
}

// GetTempFile 创建临时文件
func GetTempFile() (*os.File, error) {
	return os.CreateTemp("", "")
}

// GetTempDir 创建临时目录
func GetTempDirPath() (string, error) {
	return os.MkdirTemp("", "")
}

// GetHomeDir 获取用户主目录
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetConfigDir 获取配置目录
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Roaming"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support"), nil
	default: // linux, etc.
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			return xdgConfig, nil
		}
		return filepath.Join(home, ".config"), nil
	}
}

// GetCacheDir 获取缓存目录
func GetCacheDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Caches"), nil
	default: // linux, etc.
		if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
			return xdgCache, nil
		}
		return filepath.Join(home, ".cache"), nil
	}
}

// GetDataDir 获取数据目录
func GetDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		return filepath.Join(home, "AppData", "Local"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support"), nil
	default: // linux, etc.
		if xdgData := os.Getenv("XDG_DATA_HOME"); xdgData != "" {
			return xdgData, nil
		}
		return filepath.Join(home, ".local", "share"), nil
	}
}

// SplitList 分割路径列表
func SplitList(pathList string) []string {
	return filepath.SplitList(pathList)
}

// JoinList 连接路径列表
func JoinList(paths []string) string {
	return strings.Join(paths, string(os.PathListSeparator))
}

// HasExtension 检查是否有指定扩展名
func HasExtension(p, ext string) bool {
	return filepath.Ext(p) == ext
}

// HasAnyExtension 检查是否有任一扩展名
func HasAnyExtension(p string, exts []string) bool {
	ext := filepath.Ext(p)
	for _, e := range exts {
		if ext == e {
			return true
		}
	}
	return false
}

// IsHidden 检查是否为隐藏文件/目录
func IsHidden(p string) bool {
	base := filepath.Base(p)

	// Windows: 检查隐藏属性
	if runtime.GOOS == "windows" {
		// 这里需要调用Windows API来检查文件属性
		// 简化实现，只检查文件名
	}

	// Unix-like: 以点开头
	return strings.HasPrefix(base, ".")
}

// GetRelativePath 获取相对于基准路径的相对路径
func GetRelativePath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// MakeRelativeTo 使路径相对于基准路径
func MakeRelativeTo(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
