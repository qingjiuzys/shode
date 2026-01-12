package registry

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTarball(t *testing.T) {
	// Create a temporary directory for tarball
	tmpDir, err := os.MkdirTemp("", "shode-tarball-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test tarball
	tarballPath := filepath.Join(tmpDir, "test.tar.gz")
	createTestTarball(t, tarballPath)

	// Extract tarball
	targetDir := filepath.Join(tmpDir, "extracted")
	err = extractTarball(tarballPath, targetDir)
	if err != nil {
		t.Fatalf("Failed to extract tarball: %v", err)
	}

	// Verify extracted files
	testFile := filepath.Join(targetDir, "test.txt")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Extracted file not found")
	}

	// Verify file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read extracted file: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Expected content 'test content', got '%s'", string(content))
	}
}

func TestExtractTarballPathTraversal(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-tarball-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a malicious tarball with path traversal
	tarballPath := filepath.Join(tmpDir, "malicious.tar.gz")
	createMaliciousTarball(t, tarballPath)

	// Extract tarball
	targetDir := filepath.Join(tmpDir, "extracted")
	err = extractTarball(tarballPath, targetDir)
	
	// Should fail due to path traversal protection
	if err == nil {
		t.Error("Expected error for path traversal attack, got nil")
	}
}

func TestExtractTarballWithSymlink(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-tarball-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a tarball with symlink
	tarballPath := filepath.Join(tmpDir, "symlink.tar.gz")
	createSymlinkTarball(t, tarballPath)

	// Extract tarball
	targetDir := filepath.Join(tmpDir, "extracted")
	err = extractTarball(tarballPath, targetDir)
	if err != nil {
		t.Fatalf("Failed to extract tarball with symlink: %v", err)
	}

	// Verify symlink exists
	symlinkPath := filepath.Join(targetDir, "link")
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Error("Symlink not found after extraction")
	}
}

func TestExtractTarballWithDirectories(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "shode-tarball-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a tarball with nested directories
	tarballPath := filepath.Join(tmpDir, "nested.tar.gz")
	createNestedTarball(t, tarballPath)

	// Extract tarball
	targetDir := filepath.Join(tmpDir, "extracted")
	err = extractTarball(tarballPath, targetDir)
	if err != nil {
		t.Fatalf("Failed to extract nested tarball: %v", err)
	}

	// Verify nested directory structure
	nestedFile := filepath.Join(targetDir, "dir1", "dir2", "file.txt")
	if _, err := os.Stat(nestedFile); os.IsNotExist(err) {
		t.Error("Nested file not found")
	}
}

// Helper functions to create test tarballs

func createTestTarball(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create tarball file: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add a test file
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len("test content")),
		Mode: 0644,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("test content")); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}
}

func createMaliciousTarball(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create tarball file: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add a file with path traversal
	header := &tar.Header{
		Name: "../../../etc/passwd",
		Size: int64(len("malicious")),
		Mode: 0644,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("malicious")); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}
}

func createSymlinkTarball(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create tarball file: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add a regular file
	header := &tar.Header{
		Name: "target.txt",
		Size: int64(len("target")),
		Mode: 0644,
	}
	if err := tw.WriteHeader(header); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write([]byte("target")); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	// Add a symlink
	linkHeader := &tar.Header{
		Name:     "link",
		Linkname: "target.txt",
		Typeflag: tar.TypeSymlink,
		Mode:     0644,
	}
	if err := tw.WriteHeader(linkHeader); err != nil {
		t.Fatalf("Failed to write symlink header: %v", err)
	}
}

func createNestedTarball(t *testing.T, path string) {
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create tarball file: %v", err)
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Add nested directories
	dir1Header := &tar.Header{
		Name:     "dir1/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tw.WriteHeader(dir1Header); err != nil {
		t.Fatalf("Failed to write dir header: %v", err)
	}

	dir2Header := &tar.Header{
		Name:     "dir1/dir2/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
	}
	if err := tw.WriteHeader(dir2Header); err != nil {
		t.Fatalf("Failed to write nested dir header: %v", err)
	}

	// Add file in nested directory
	fileHeader := &tar.Header{
		Name: "dir1/dir2/file.txt",
		Size: int64(len("nested content")),
		Mode: 0644,
	}
	if err := tw.WriteHeader(fileHeader); err != nil {
		t.Fatalf("Failed to write file header: %v", err)
	}
	if _, err := tw.Write([]byte("nested content")); err != nil {
		t.Fatalf("Failed to write file content: %v", err)
	}
}
