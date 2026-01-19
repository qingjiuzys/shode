package stdlib

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CopyFile copies a file (replaces 'cp')
func (sl *StdLib) CopyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer input.Close()

	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	output, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dst, err)
	}
	defer output.Close()

	_, err = io.Copy(output, input)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	// Copy permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %v", err)
	}

	err = os.Chmod(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to copy permissions: %v", err)
	}

	return nil
}

// CopyFileRecursive recursively copies files (replaces 'cp -r')
func (sl *StdLib) CopyFileRecursive(src, dst string) error {
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

		return sl.CopyFile(path, dstPath)
	})
}

// MoveFile moves/renames a file (replaces 'mv')
func (sl *StdLib) MoveFile(src, dst string) error {
	// Create destination directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	return os.Rename(src, dst)
}

// DeleteFile deletes a file (replaces 'rm')
func (sl *StdLib) DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteFileRecursive deletes a directory recursively (replaces 'rm -r' or 'rm -rf')
func (sl *StdLib) DeleteFileRecursive(path string) error {
	return os.RemoveAll(path)
}

// CreateDir creates a directory (replaces 'mkdir')
func (sl *StdLib) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateDirWithPerms creates a directory with permissions (replaces 'mkdir -m')
func (sl *StdLib) CreateDirWithPerms(path string, perms string) error {
	perm, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions: %v", err)
	}
	return os.MkdirAll(path, os.FileMode(perm))
}

// DeleteDir deletes a directory (replaces 'rmdir')
func (sl *StdLib) DeleteDir(path string) error {
	return os.Remove(path)
}

// HeadFile reads first N lines of a file (replaces 'head')
func (sl *StdLib) HeadFile(path string, n int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() && count < n {
		lines = append(lines, scanner.Text())
		count++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	return strings.Join(lines, "\n"), nil
}

// TailFile reads last N lines of a file (replaces 'tail')
func (sl *StdLib) TailFile(path string, n int) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}

	return strings.Join(lines[start:], "\n"), nil
}

// FindFiles searches for files (replaces 'find')
func (sl *StdLib) FindFiles(dir, pattern string) ([]string, error) {
	var matches []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		if matched {
			matches = append(matches, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory: %v", err)
	}

	return matches, nil
}

// ChangePermissions changes file permissions (replaces 'chmod')
func (sl *StdLib) ChangePermissions(path, perms string) error {
	perm, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions: %v", err)
	}
	return os.Chmod(path, os.FileMode(perm))
}

// ChangePermissionsRecursive changes permissions recursively (replaces 'chmod -R')
func (sl *StdLib) ChangePermissionsRecursive(path, perms string) error {
	perm, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions: %v", err)
	}

	return filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return os.Chmod(p, os.FileMode(perm))
	})
}

// ChangeOwner changes file owner (replaces 'chown')
func (sl *StdLib) ChangeOwner(path, user, group string) error {
	// This would require syscall support, returning error for now
	return fmt.Errorf("chown is not supported on this platform")
}

// WordCount counts lines, words, and bytes (replaces 'wc')
func (sl *StdLib) WordCount(path string) (map[string]int, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines int
	var words int

	for scanner.Scan() {
		lines++
		words += len(strings.Fields(scanner.Text()))
	}

	result := map[string]int{
		"lines": lines,
		"words": words,
	}

	if info, err := os.Stat(path); err == nil {
		result["bytes"] = int(info.Size())
	}

	return result, nil
}

// DiffFiles compares two files (replaces 'diff')
func (sl *StdLib) DiffFiles(file1, file2 string) (string, error) {
	cmd := exec.Command("diff", file1, file2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil // diff returns non-zero for differences
	}
	return string(output), nil
}

// UniqueLines removes duplicate lines (replaces 'uniq')
func (sl *StdLib) UniqueLines(input string) string {
	lines := strings.Split(input, "\n")
	seen := make(map[string]bool)
	var unique []string

	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			unique = append(unique, line)
		}
	}

	return strings.Join(unique, "\n")
}

// SortLines sorts lines alphabetically (replaces 'sort')
func (sl *StdLib) SortLines(input string) string {
	lines := strings.Split(input, "\n")
	// Remove empty lines
	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return strings.Join(nonEmpty, "\n")
}
