package stdlib

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ========== File Operations ==========

// CopyFile copies a file (replaces 'cp')
func (fm *FilesManager) CopyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer input.Close()

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
func (fm *FilesManager) CopyFileRecursive(src, dst string) error {
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

		return fm.CopyFile(path, dstPath)
	})
}

// MoveFile moves/renames a file (replaces 'mv')
func (fm *FilesManager) MoveFile(src, dst string) error {
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	return os.Rename(src, dst)
}

// DeleteFile deletes a file (replaces 'rm')
func (fm *FilesManager) DeleteFile(path string) error {
	return os.Remove(path)
}

// DeleteFileRecursive deletes a directory recursively (replaces 'rm -r')
func (fm *FilesManager) DeleteFileRecursive(path string) error {
	return os.RemoveAll(path)
}

// CreateDir creates a directory (replaces 'mkdir')
func (fm *FilesManager) CreateDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// CreateDirWithPerms creates a directory with permissions (replaces 'mkdir -m')
func (fm *FilesManager) CreateDirWithPerms(path, perms string) error {
	perm, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions: %v", err)
	}
	return os.MkdirAll(path, os.FileMode(perm))
}

// HeadFile reads first N lines of a file (replaces 'head')
func (fm *FilesManager) HeadFile(path string, n int) (string, error) {
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
func (fm *FilesManager) TailFile(path string, n int) (string, error) {
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
func (fm *FilesManager) FindFiles(dir, pattern string) ([]string, error) {
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
func (fm *FilesManager) ChangePermissions(path, perms string) error {
	perm, err := strconv.ParseUint(perms, 8, 32)
	if err != nil {
		return fmt.Errorf("invalid permissions: %v", err)
	}
	return os.Chmod(path, os.FileMode(perm))
}

// ChangePermissionsRecursive changes permissions recursively (replaces 'chmod -R')
func (fm *FilesManager) ChangePermissionsRecursive(path, perms string) error {
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

// WordCount counts lines, words, and bytes (replaces 'wc')
func (fm *FilesManager) WordCount(path string) (map[string]int, error) {
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
func (fm *FilesManager) DiffFiles(file1, file2 string) (string, error) {
	cmd := exec.Command("diff", file1, file2)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil
	}
	return string(output), nil
}

// UniqueLines removes duplicate lines (replaces 'uniq')
func (fm *FilesManager) UniqueLines(input string) string {
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
func (fm *FilesManager) SortLines(input string) string {
	lines := strings.Split(input, "\n")
	var nonEmpty []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}
	return strings.Join(nonEmpty, "\n")
}

// ========== System Operations ==========

// ListProcesses lists running processes (replaces 'ps')
func (sm *SystemManager) ListProcesses(filter string) ([]map[string]interface{}, error) {
	cmd := exec.Command("ps", "aux")
	if filter != "" {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("ps aux | grep %s", filter))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var processes []map[string]interface{}

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}

		cpu, _ := strconv.ParseFloat(fields[2], 64)
		mem, _ := strconv.ParseFloat(fields[3], 64)

		process := map[string]interface{}{
			"user":    fields[0],
			"pid":     pid,
			"cpu":     cpu,
			"memory":  mem,
			"command": strings.Join(fields[10:], " "),
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// KillProcess terminates a process (replaces 'kill')
func (sm *SystemManager) KillProcess(pid int, signal string) error {
	sig := "TERM"
	if signal != "" {
		sig = signal
	}

	cmd := exec.Command("kill", "-"+sig, strconv.Itoa(pid))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill process: %s", string(output))
	}

	return nil
}

// KillProcessByName terminates processes by name (replaces 'pkill')
func (sm *SystemManager) KillProcessByName(name string, signal string) error {
	sig := "TERM"
	if signal != "" {
		sig = signal
	}

	cmd := exec.Command("pkill", "-"+sig, name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to kill processes: %s", string(output))
	}

	return nil
}

// DiskUsage shows disk usage (replaces 'df')
func (sm *SystemManager) DiskUsage(path string) (map[string]string, error) {
	cmd := exec.Command("df", "-h", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get disk usage: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 6 {
		return nil, fmt.Errorf("unexpected df format")
	}

	return map[string]string{
		"filesystem":  fields[0],
		"size":        fields[1],
		"used":        fields[2],
		"available":   fields[3],
		"use_percent": strings.TrimSuffix(fields[4], "%"),
		"mounted":     fields[5],
	}, nil
}

// DirSize shows directory size (replaces 'du')
func (sm *SystemManager) DirSize(path string) (string, error) {
	cmd := exec.Command("du", "-sh", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get directory size: %v", err)
	}

	fields := strings.Fields(string(output))
	if len(fields) >= 2 {
		return fields[0], nil
	}

	return "", fmt.Errorf("unexpected du output")
}

// StartService starts a systemd service (replaces 'systemctl start')
func (sm *SystemManager) StartService(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %s", name, string(output))
	}

	return nil
}

// StopService stops a systemd service (replaces 'systemctl stop')
func (sm *SystemManager) StopService(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %s", name, string(output))
	}

	return nil
}

// RestartService restarts a systemd service (replaces 'systemctl restart')
func (sm *SystemManager) RestartService(name string) error {
	cmd := exec.Command("systemctl", "restart", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %s", name, string(output))
	}

	return nil
}

// ServiceStatus gets service status (replaces 'systemctl status')
func (sm *SystemManager) ServiceStatus(name string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ServiceEnabled checks if service is enabled (replaces 'systemctl is-enabled')
func (sm *SystemManager) ServiceEnabled(name string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check if service is enabled: %v", err)
	}

	status := strings.TrimSpace(string(output))
	return status == "enabled", nil
}

// GetSystemInfo gets system information (replaces 'uname -a')
func (sm *SystemManager) GetSystemInfo() (string, error) {
	cmd := exec.Command("uname", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get system info: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetHostname gets system hostname (replaces 'hostname')
func (sm *SystemManager) GetHostname() (string, error) {
	cmd := exec.Command("hostname")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentUser gets current username (replaces 'whoami')
func (sm *SystemManager) GetCurrentUser() (string, error) {
	cmd := exec.Command("whoami")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetUptime gets system uptime (replaces 'uptime')
func (sm *SystemManager) GetUptime() (string, error) {
	cmd := exec.Command("uptime")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get uptime: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetMemoryUsage gets memory usage (replaces 'free')
func (sm *SystemManager) GetMemoryUsage() (map[string]string, error) {
	cmd := exec.Command("free", "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("unexpected free output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) >= 4 {
		return map[string]string{
			"total":     fields[1],
			"used":      fields[2],
			"available": fields[3],
		}, nil
	}

	return nil, fmt.Errorf("unexpected free format")
}

// ========== Network Operations ==========

// HTTPRequest sends an HTTP request (replaces 'curl')
func (nm *NetworkManager) HTTPRequest(method, url string, headers map[string]string, body string) (string, int, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read response: %v", err)
	}

	return string(responseBody), resp.StatusCode, nil
}

// Ping sends ICMP echo requests (replaces 'ping')
func (nm *NetworkManager) Ping(host string, count int) (map[string]interface{}, error) {
	cmd := exec.Command("ping", "-c", strconv.Itoa(count), host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	result := map[string]interface{}{
		"host":         host,
		"packets_sent": count,
	}

	for _, line := range lines {
		if strings.Contains(line, "packets transmitted") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				received, _ := strconv.Atoi(fields[3])
				result["packets_received"] = received
			}
		}

		if strings.Contains(line, "packet loss") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				loss, _ := strconv.ParseFloat(strings.TrimSuffix(fields[5], "%"), 64)
				result["packet_loss"] = loss
			}
		}

		if strings.Contains(line, "avg") {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				avg, _ := strconv.ParseFloat(fields[6], 64)
				result["avg_time"] = avg
			}
		}
	}

	return result, nil
}

// DownloadFile downloads a file from URL (replaces 'wget')
func (nm *NetworkManager) DownloadFile(url, dstPath string) (string, error) {
	cmd := exec.Command("wget", "-O", dstPath, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("wget failed: %s", string(output))
	}

	return string(output), nil
}

// Netstat shows network connections (replaces 'netstat' or 'ss')
func (nm *NetworkManager) Netstat(proto string) (string, error) {
	cmd := exec.Command("ss", "-tunlp")
	if proto != "" {
		cmd = exec.Command("ss", "-tunlp", strings.ToLower(proto))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("netstat", "-tunlp")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("netstat/ss failed: %v", err)
		}
	}

	return string(output), nil
}

// GetLocalIP gets local IP address
func (nm *NetworkManager) GetLocalIP() (string, error) {
	cmd := exec.Command("hostname", "-I")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get local IP: %v", err)
	}

	ips := strings.Fields(string(output))
	if len(ips) > 0 {
		return ips[0], nil
	}

	return "", fmt.Errorf("no IP address found")
}

// ========== Archive Operations ==========

// Tar creates a tar archive (replaces 'tar -cf')
func (am *ArchiveManager) Tar(src, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create tar file: %v", err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Untar extracts a tar archive (replaces 'tar -xf')
func (am *ArchiveManager) Untar(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %v", err)
	}
	defer file.Close()

	tr := tar.NewReader(file)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %v", err)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, tr)
			if err != nil {
				return err
			}

		default:
			fmt.Printf("Unknown type: %v in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}

// Gzip compresses a file (replaces 'gzip')
func (am *ArchiveManager) Gzip(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	gz := gzip.NewWriter(dstFile)
	defer gz.Close()

	_, err = io.Copy(gz, srcFile)
	return err
}

// Gunzip decompresses a gzip file (replaces 'gunzip')
func (am *ArchiveManager) Gunzip(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	gz, err := gzip.NewReader(srcFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gz.Close()

	_, err = io.Copy(dstFile, gz)
	return err
}

// GzipDir compresses a directory into a tar.gz file
func (am *ArchiveManager) GzipDir(src, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create tar.gz file: %v", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = path

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// GunzipDir decompresses a tar.gz file
func (am *ArchiveManager) GunzipDir(src, dst string) error {
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}

	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read error: %v", err)
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, tr)
			if err != nil {
				return err
			}

		default:
			fmt.Printf("Unknown type: %v in %s\n", header.Typeflag, header.Name)
		}
	}

	return nil
}
