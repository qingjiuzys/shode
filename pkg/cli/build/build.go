// Package build 构建打包工具
package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Builder 构建器
type Builder struct {
	AppName      string
	Version      string
	OutputDir    string
	Platforms    []Platform
	Compress     bool
	Docker       bool
}

// Platform 目标平台
type Platform struct {
	OS   string
	Arch string
}

// NewBuilder 创建构建器
func NewBuilder(name, version string) *Builder {
	return &Builder{
		AppName:   name,
		Version:   version,
		OutputDir: "dist",
		Platforms: []Platform{
			{OS: runtime.GOOS, Arch: runtime.GOARCH},
		},
	}
}

// Build 构建应用
func (b *Builder) Build() error {
	fmt.Printf("Building %s v%s\n", b.AppName, b.Version)

	// 创建输出目录
	if err := os.MkdirAll(b.OutputDir, 0755); err != nil {
		return err
	}

	// 构建每个平台
	for _, platform := range b.Platforms {
		if err := b.buildPlatform(platform); err != nil {
			return fmt.Errorf("build failed for %s/%s: %w", platform.OS, platform.Arch, err)
		}
	}

	fmt.Println("✓ Build completed")
	return nil
}

// buildPlatform 构建特定平台
func (b *Builder) buildPlatform(platform Platform) error {
	fmt.Printf("Building for %s/%s...\n", platform.OS, platform.Arch)

	// 设置环境变量
	env := append(os.Environ(),
		fmt.Sprintf("GOOS=%s", platform.OS),
		fmt.Sprintf("GOARCH=%s", platform.Arch),
	)

	// 构建二进制文件
	binaryName := b.getBinaryName(platform)
	outputPath := filepath.Join(b.OutputDir, binaryName)

	cmd := exec.Command("go", "build", "-o", outputPath, "cmd/*/main.go")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// 压缩
	if b.Compress {
		if err := b.compressBinary(outputPath); err != nil {
			return err
		}
	}

	return nil
}

// getBinaryName 获取二进制文件名
func (b *Builder) getBinaryName(platform Platform) string {
	name := b.AppName

	if platform.OS == "windows" {
		name += ".exe"
	}

	if platform.OS != runtime.GOOS || platform.Arch != runtime.GOARCH {
		name = fmt.Sprintf("%s_%s_%s", b.AppName, platform.OS, platform.Arch)
		if platform.OS == "windows" {
			name += ".exe"
		}
	}

	return name
}

// compressBinary 压缩二进制文件
func (b *Builder) compressBinary(binaryPath string) error {
	fmt.Printf("Compressing %s...\n", binaryPath)

	// 使用 gzip 压缩
	cmd := exec.Command("gzip", binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// BuildAll 构建所有平台
func (b *Builder) BuildAll() error {
	b.Platforms = []Platform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
		{OS: "darwin", Arch: "amd64"},
		{OS: "darwin", Arch: "arm64"},
		{OS: "windows", Arch: "amd64"},
	}

	return b.Build()
}

// BuildDocker 构建 Docker 镜像
func (b *Builder) BuildDocker() error {
	fmt.Println("Building Docker image...")

	imageName := fmt.Sprintf("%s:%s", b.AppName, b.Version)

	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("✓ Docker image built: %s\n", imageName)
	return nil
}

// GetAppName 获取应用名
func GetAppName() (string, error) {
	// 从 go.mod 或目录名获取
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Base(dir), nil
}

// GetVersion 获取版本号
func GetVersion() (string, error) {
	// 从 git tag 或配置文件获取
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return "1.0.0", nil
	}

	return strings.TrimSpace(string(output)), nil
}

// Clean 清理构建产物
func Clean() error {
	fmt.Println("Cleaning build artifacts...")

	outputDir := "dist"
	if _, err := os.Stat(outputDir); err == nil {
		return os.RemoveAll(outputDir)
	}

	return nil
}

// BuildCurrent 构建当前平台
func BuildCurrent() error {
	name, err := GetAppName()
	if err != nil {
		return err
	}

	version, err := GetVersion()
	if err != nil {
		return err
	}

	builder := NewBuilder(name, version)
	return builder.Build()
}

// BuildMultiPlatform 构建多平台
func BuildMultiPlatform() error {
	name, err := GetAppName()
	if err != nil {
		return err
	}

	version, err := GetVersion()
	if err != nil {
		return err
	}

	builder := NewBuilder(name, version)
	builder.Compress = true
	return builder.BuildAll()
}

// Release 创建发布包
func Release() error {
	name, err := GetAppName()
	if err != nil {
		return err
	}

	version, err := GetVersion()
	if err != nil {
		return err
	}

	builder := NewBuilder(name, version)
	builder.Compress = true

	// 构建所有平台
	if err := builder.BuildAll(); err != nil {
		return err
	}

	// 构建 Docker 镜像
	if err := builder.BuildDocker(); err != nil {
		return err
	}

	fmt.Println("✓ Release package created")
	return nil
}

// GetGitStatus 获取 Git 状态
func GetGitStatus() (string, error) {
	cmd := exec.Command("git", "status", "--short")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// IsGitDirty 检查是否有未提交的更改
func IsGitDirty() (bool, error) {
	status, err := GetGitStatus()
	if err != nil {
		return false, err
	}

	return status != "", nil
}

// GenerateChecksums 生成校验和
func GenerateChecksums() error {
	fmt.Println("Generating checksums...")

	// 简化实现：需要遍历 dist 目录并生成 SHA256 校验和
	return nil
}

// CreateArchive 创建归档文件
func CreateArchive(source, dest string) error {
	fmt.Printf("Creating archive: %s\n", dest)

	// 使用 tar 创建归档
	cmd := exec.Command("tar", "-czf", dest, source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
