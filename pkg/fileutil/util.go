// Package fileutil 提供文件操作工具
package fileutil

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Exists 检查文件或目录是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile 检查是否为文件
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir 检查是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// Size 获取文件大小
func Size(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ModTime 获取文件修改时间
func ModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// ReadFile 读取文件内容
func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadFileBytes 读取文件字节内容
func ReadFileBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile 写入文件内容
func WriteFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

// WriteFileBytes 写入文件字节内容
func WriteFileBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// AppendFile 追加文件内容
func AppendFile(path string, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// CreateFile 创建文件
func CreateFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	return file.Close()
}

// DeleteFile 删除文件
func DeleteFile(path string) error {
	return os.Remove(path)
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// Rename 重命名文件
func Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// CreateDir 创建目录
func CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// DeleteDir 删除目录
func DeleteDir(path string) error {
	return os.RemoveAll(path)
}

// ListFiles 列出目录下的所有文件
func ListFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// ListDirs 列出目录下的所有子目录
func ListDirs(dir string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(dir, entry.Name()))
		}
	}

	return dirs, nil
}

// ListEntries 列出目录下的所有条目（文件和目录）
func ListEntries(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		result = append(result, filepath.Join(dir, entry.Name()))
	}

	return result, nil
}

// CleanDir 清空目录（删除目录下所有内容但保留目录本身）
func CleanDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

// DirSize 计算目录大小
func DirSize(dir string) (int64, error) {
	var size int64

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FileCount 统计目录下的文件数量
func FileCount(dir string) (int, error) {
	count := 0

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	return count, err
}

// TempDir 创建临时目录
func TempDir() (string, error) {
	return os.MkdirTemp("", "temp")
}

// TempFile 创建临时文件
func TempFile() (*os.File, error) {
	return os.CreateTemp("", "temp")
}

// TempFileName 生成临时文件名
func TempFileName() string {
	return fmt.Sprintf("temp_%d", time.Now().UnixNano())
}

// ReadLines 读取文件所有行
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

// WriteLines 写入行到文件
func WriteLines(path string, lines []string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// AppendLine 追加一行到文件
func AppendLine(path string, line string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(line + "\n")
	return err
}

// ReadCSV 读取CSV文件
func ReadCSV(path string) ([][]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}

// WriteCSV 写入CSV文件
func WriteCSV(path string, records [][]string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	return writer.WriteAll(records)
}

// JoinPath 连接路径
func JoinPath(paths ...string) string {
	return filepath.Join(paths...)
}

// AbsPath 获取绝对路径
func AbsPath(path string) (string, error) {
	return filepath.Abs(path)
}

// RelPath 获取相对路径
func RelPath(base, target string) (string, error) {
	return filepath.Rel(base, target)
}

// BaseName 获取文件基名（不含扩展名）
func BaseName(path string) string {
	return strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
}

// Ext 获取文件扩展名
func Ext(path string) string {
	return filepath.Ext(path)
}

// Dir 获取目录路径
func Dir(path string) string {
	return filepath.Dir(path)
}

// Clean 清理路径
func Clean(path string) string {
	return filepath.Clean(path)
}

// Split 分割路径为目录和文件
func Split(path string) (dir, file string) {
	return filepath.Split(path)
}

// Match 检查文件名是否匹配模式
func Match(pattern, name string) (bool, error) {
	return filepath.Match(pattern, name)
}

// Glob 查找匹配模式的所有文件
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Walk 遍历目录树
func Walk(root string, walkFn filepath.WalkFunc) error {
	return filepath.Walk(root, walkFn)
}

// Watch 监控文件变化
func Watch(ctx context.Context, path string, interval time.Duration, callback func()) error {
	lastModTime := time.Time{}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.ModTime().After(lastModTime) {
				lastModTime = info.ModTime()
				callback()
			}
		}
	}
}

// FileBuffer 文件缓冲区
type FileBuffer struct {
	buffer *bytes.Buffer
	path   string
}

// NewFileBuffer 创建文件缓冲区
func NewFileBuffer() *FileBuffer {
	return &FileBuffer{
		buffer: &bytes.Buffer{},
	}
}

// Write 写入数据
func (fb *FileBuffer) Write(data []byte) (int, error) {
	return fb.buffer.Write(data)
}

// WriteString 写入字符串
func (fb *FileBuffer) WriteString(s string) (int, error) {
	return fb.buffer.WriteString(s)
}

// Bytes 获取字节数据
func (fb *FileBuffer) Bytes() []byte {
	return fb.buffer.Bytes()
}

// String 获取字符串数据
func (fb *FileBuffer) String() string {
	return fb.buffer.String()
}

// Save 保存到文件
func (fb *FileBuffer) Save(path string) error {
	fb.path = path
	return os.WriteFile(path, fb.buffer.Bytes(), 0644)
}

// SaveWithPermission 保存到文件（指定权限）
func (fb *FileBuffer) SaveWithPermission(path string, perm os.FileMode) error {
	fb.path = path
	return os.WriteFile(path, fb.buffer.Bytes(), perm)
}

// Load 从文件加载
func (fb *FileBuffer) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	fb.path = path
	_, err = fb.buffer.Write(data)
	return err
}

// Clear 清空缓冲区
func (fb *FileBuffer) Clear() {
	fb.buffer.Reset()
}

// Size 获取缓冲区大小
func (fb *FileBuffer) Size() int {
	return fb.buffer.Len()
}

// Path 获取关联的文件路径
func (fb *FileBuffer) Path() string {
	return fb.path
}

// FileFilter 文件过滤器
type FileFilter struct {
	extensions []string
	patterns   []string
	minSize    int64
	maxSize    int64
}

// NewFileFilter 创建文件过滤器
func NewFileFilter() *FileFilter {
	return &FileFilter{}
}

// AddExtension 添加扩展名过滤
func (ff *FileFilter) AddExtension(ext string) *FileFilter {
	ff.extensions = append(ff.extensions, ext)
	return ff
}

// AddPattern 添加模式过滤
func (ff *FileFilter) AddPattern(pattern string) *FileFilter {
	ff.patterns = append(ff.patterns, pattern)
	return ff
}

// SetMinSize 设置最小文件大小
func (ff *FileFilter) SetMinSize(size int64) *FileFilter {
	ff.minSize = size
	return ff
}

// SetMaxSize 设置最大文件大小
func (ff *FileFilter) SetMaxSize(size int64) *FileFilter {
	ff.maxSize = size
	return ff
}

// Match 检查文件是否匹配
func (ff *FileFilter) Match(path string) bool {
	// 检查扩展名
	if len(ff.extensions) > 0 {
		ext := filepath.Ext(path)
		matched := false
		for _, e := range ff.extensions {
			if ext == e {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查模式
	if len(ff.patterns) > 0 {
		matched := false
		for _, pattern := range ff.patterns {
			if m, _ := filepath.Match(pattern, filepath.Base(path)); m {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// 检查文件大小
	if ff.minSize > 0 || ff.maxSize > 0 {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}

		if ff.minSize > 0 && info.Size() < ff.minSize {
			return false
		}

		if ff.maxSize > 0 && info.Size() > ff.maxSize {
			return false
		}
	}

	return true
}

// Filter 过滤文件列表
func (ff *FileFilter) Filter(files []string) []string {
	result := make([]string, 0)
	for _, file := range files {
		if ff.Match(file) {
			result = append(result, file)
		}
	}
	return result
}

// FindFiles 查找匹配的文件
func FindFiles(dir string, filter *FileFilter) ([]string, error) {
	files, err := ListFiles(dir)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		files = filter.Filter(files)
	}

	return files, nil
}

// FileCopier 文件复制器
type FileCopier struct {
	src      string
	dst      string
	overwrite bool
}

// NewFileCopier 创建文件复制器
func NewFileCopier(src, dst string) *FileCopier {
	return &FileCopier{
		src: src,
		dst: dst,
		overwrite: false,
	}
}

// SetOverwrite 设置是否覆盖
func (fc *FileCopier) SetOverwrite(overwrite bool) *FileCopier {
	fc.overwrite = overwrite
	return fc
}

// Copy 执行复制
func (fc *FileCopier) Copy() error {
	// 检查源文件
	if !Exists(fc.src) {
		return errors.New("source file does not exist")
	}

	// 检查目标文件
	if Exists(fc.dst) && !fc.overwrite {
		return errors.New("destination file already exists")
	}

	// 复制文件
	return CopyFile(fc.src, fc.dst)
}

// FileMover 文件移动器
type FileMover struct {
	src      string
	dst      string
	overwrite bool
}

// NewFileMover 创建文件移动器
func NewFileMover(src, dst string) *FileMover {
	return &FileMover{
		src: src,
		dst: dst,
		overwrite: false,
	}
}

// SetOverwrite 设置是否覆盖
func (fm *FileMover) SetOverwrite(overwrite bool) *FileMover {
	fm.overwrite = overwrite
	return fm
}

// Move 执行移动
func (fm *FileMover) Move() error {
	// 检查源文件
	if !Exists(fm.src) {
		return errors.New("source file does not exist")
	}

	// 检查目标文件
	if Exists(fm.dst) && !fm.overwrite {
		return errors.New("destination file already exists")
	}

	// 移动文件
	return MoveFile(fm.src, fm.dst)
}

// StreamReader 流读取器
type StreamReader struct {
	reader io.Reader
	buffer []byte
}

// NewStreamReader 创建流读取器
func NewStreamReader(r io.Reader, bufferSize int) *StreamReader {
	if bufferSize <= 0 {
		bufferSize = 4096
	}
	return &StreamReader{
		reader: r,
		buffer: make([]byte, bufferSize),
	}
}

// ReadLine 读取一行
func (sr *StreamReader) ReadLine() (string, error) {
	bufReader := bufio.NewReader(sr.reader)
	line, err := bufReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

// ReadAll 读取全部
func (sr *StreamReader) ReadAll() ([]byte, error) {
	return io.ReadAll(sr.reader)
}

// ReadString 读取字符串
func (sr *StreamReader) ReadString() (string, error) {
	data, err := io.ReadAll(sr.reader)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
