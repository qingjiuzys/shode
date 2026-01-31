// Package upload 提供文件上传功能。
package upload

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// File 上传的文件信息
type File struct {
	Filename    string
	ContentType string
	Size        int64
	Path        string
}

// Handler 文件上传处理器
type Handler struct {
	UploadDir    string
	MaxSize     int64
	AllowedExts []string
}

// NewHandler 创建上传处理器
func NewHandler(dir string, maxSize int64) *Handler {
	return &Handler{
		UploadDir: dir,
		MaxSize:  maxSize,
	}
}

// ParseForm 解析 multipart 表单
func (h *Handler) ParseForm(r *http.Request) (map[string][]string, map[string][]*File, error) {
	if err := r.ParseMultipartForm(h.MaxSize); err != nil {
		return nil, nil, err
	}

	files := make(map[string][]*File)
	for key, headers := range r.MultipartForm.File {
		for _, header := range headers {
			file, err := header.Open()
			if err != nil {
				return nil, nil, err
			}
			defer file.Close()

			// 验证文件扩展名
			ext := strings.ToLower(filepath.Ext(header.Filename))
			if len(h.AllowedExts) > 0 && !h.isAllowed(ext) {
				os.Remove(header.Filename)
				continue
			}

			// 保存文件
			path := filepath.Join(h.UploadDir, header.Filename)
			dst, err := os.Create(path)
			if err != nil {
				return nil, nil, err
			}
			defer dst.Close()

			size, err := io.Copy(dst, file)
			if err != nil {
				return nil, nil, err
			}

			files[key] = append(files[key], &File{
				Filename:    header.Filename,
				ContentType: header.Header.Get("Content-Type"),
				Size:        size,
				Path:        path,
			})
		}
	}

	return r.MultipartForm.Value, files, nil
}

func (h *Handler) isAllowed(ext string) bool {
	for _, allowed := range h.AllowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}
