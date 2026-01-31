// Package storage 提供文件存储功能。
package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStorage 文件存储接口
type FileStorage interface {
	Upload(ctx context.Context, key string, reader io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GetURL(ctx context.Context, key string) (string, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

// LocalStorage 本地文件存储
type LocalStorage struct {
	baseDir string
	mu      sync.RWMutex
}

// NewLocalStorage 创建本地存储
func NewLocalStorage(baseDir string) (*LocalStorage, error) {
	// 确保目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalStorage{
		baseDir: baseDir,
	}, nil
}

// Upload 上传文件
func (ls *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// 构建文件路径
	filePath := filepath.Join(ls.baseDir, key)

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// 复制数据
	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// Download 下载文件
func (ls *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	filePath := filepath.Join(ls.baseDir, key)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete 删除文件
func (ls *LocalStorage) Delete(ctx context.Context, key string) error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	filePath := filepath.Join(ls.baseDir, key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在
func (ls *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	filePath := filepath.Join(ls.baseDir, key)
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetURL 获取文件 URL
func (ls *LocalStorage) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("/files/%s", key), nil
}

// List 列出文件
func (ls *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	dir := filepath.Join(ls.baseDir, prefix)
	files := make([]string, 0)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			rel, err := filepath.Rel(ls.baseDir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return files, nil
}

// ChunkedUpload 分块上传
type ChunkedUpload struct {
	storage    FileStorage
	uploadID   string
	chunkSize  int64
	chunks     map[int]bool
	totalSize  int64
	mu         sync.Mutex
}

// NewChunkedUpload 创建分块上传
func NewChunkedUpload(storage FileStorage, uploadID string, chunkSize int64) *ChunkedUpload {
	return &ChunkedUpload{
		storage:   storage,
		uploadID:  uploadID,
		chunkSize: chunkSize,
		chunks:    make(map[int]bool),
	}
}

// UploadChunk 上传分块
func (cu *ChunkedUpload) UploadChunk(ctx context.Context, chunkNumber int, reader io.Reader) error {
	cu.mu.Lock()
	defer cu.mu.Unlock()

	// 构建分块键
	chunkKey := fmt.Sprintf("%s/chunks/%d", cu.uploadID, chunkNumber)

	// 上传分块
	if err := cu.storage.Upload(ctx, chunkKey, reader); err != nil {
		return fmt.Errorf("failed to upload chunk: %w", err)
	}

	cu.chunks[chunkNumber] = true
	return nil
}

// Complete 合并分块
func (cu *ChunkedUpload) Complete(ctx context.Context, key string) error {
	cu.mu.Lock()
	defer cu.mu.Unlock()

	// 创建临时文件
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.tmp", cu.uploadID))
	tmp, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmp.Close()
	defer os.Remove(tmpFile)

	// 合并分块
	for i := 1; i <= len(cu.chunks); i++ {
		chunkKey := fmt.Sprintf("%s/chunks/%d", cu.uploadID, i)

		reader, err := cu.storage.Download(ctx, chunkKey)
		if err != nil {
			return fmt.Errorf("failed to download chunk %d: %w", i, err)
		}
		defer reader.Close()

		if _, err := io.Copy(tmp, reader); err != nil {
			return fmt.Errorf("failed to merge chunk %d: %w", i, err)
		}
	}

	// 重置文件指针
	if _, err := tmp.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek temp file: %w", err)
	}

	// 上传合并后的文件
	if err := cu.storage.Upload(ctx, key, tmp); err != nil {
		return fmt.Errorf("failed to upload merged file: %w", err)
	}

	// 删除分块
	for i := 1; i <= len(cu.chunks); i++ {
		chunkKey := fmt.Sprintf("%s/chunks/%d", cu.uploadID, i)
		_ = cu.storage.Delete(ctx, chunkKey)
	}

	return nil
}

// FileInfo 文件信息
type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

// GetFileInfo 获取文件信息
func (ls *LocalStorage) GetFileInfo(ctx context.Context, key string) (*FileInfo, error) {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	filePath := filepath.Join(ls.baseDir, key)
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// 计算 MD5
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, fmt.Errorf("failed to calculate md5: %w", err)
	}

	return &FileInfo{
		Key:          key,
		Size:         info.Size(),
		LastModified: info.ModTime(),
		ETag:         hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

// ImageProcessor 图片处理器
type ImageProcessor struct {
	storage FileStorage
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(storage FileStorage) *ImageProcessor {
	return &ImageProcessor{storage: storage}
}

// Thumbnail 生成缩略图
func (ip *ImageProcessor) Thumbnail(ctx context.Context, key string, width, height int) error {
	// 下载原始文件
	reader, err := ip.storage.Download(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 生成缩略图键
	thumbKey := fmt.Sprintf("thumbnails/%s_%dx%d", key, width, height)

	// 简化实现，实际应该用图片处理库
	// 这里只是原样保存
	return ip.storage.Upload(ctx, thumbKey, reader)
}

// Crop 裁剪图片
func (ip *ImageProcessor) Crop(ctx context.Context, key string, x, y, width, height int) error {
	// 简化实现
	reader, err := ip.storage.Download(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	cropKey := fmt.Sprintf("cropped/%s_%d_%d_%d_%d", key, x, y, width, height)
	return ip.storage.Upload(ctx, cropKey, reader)
}

// Resize 调整图片大小
func (ip *ImageProcessor) Resize(ctx context.Context, key string, width, height int) error {
	// 简化实现
	reader, err := ip.storage.Download(ctx, key)
	if err != nil {
		return err
	}
	defer reader.Close()

	resizeKey := fmt.Sprintf("resized/%s_%dx%d", key, width, height)
	return ip.storage.Upload(ctx, resizeKey, reader)
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
}

// OSSStorage 对象存储（简化实现）
type OSSStorage struct {
	config OSSConfig
}

// NewOSSStorage 创建 OSS 存储
func NewOSSStorage(config OSSConfig) *OSSStorage {
	return &OSSStorage{config: config}
}

// Upload 上传文件
func (oss *OSSStorage) Upload(ctx context.Context, key string, reader io.Reader) error {
	// 简化实现，实际应该调用 OSS SDK
	fmt.Printf("Uploading to OSS: %s\n", key)
	return nil
}

// Download 下载文件
func (oss *OSSStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

// Delete 删除文件
func (oss *OSSStorage) Delete(ctx context.Context, key string) error {
	// 简化实现
	return nil
}

// Exists 检查文件是否存在
func (oss *OSSStorage) Exists(ctx context.Context, key string) (bool, error) {
	// 简化实现
	return false, nil
}

// GetURL 获取文件 URL
func (oss *OSSStorage) GetURL(ctx context.Context, key string) (string, error) {
	return fmt.Sprintf("https://%s.%s/%s", oss.config.BucketName, oss.config.Endpoint, key), nil
}

// List 列出文件
func (oss *OSSStorage) List(ctx context.Context, prefix string) ([]string, error) {
	// 简化实现
	return []string{}, nil
}
