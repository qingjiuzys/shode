package stdlib

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Process represents a system process
type Process struct {
	PID     int
	User    string
	CPU     float64
	Memory  float64
	Command string
}

// ListProcesses lists running processes (replaces 'ps')
func (sl *StdLib) ListProcesses(filter string) ([]Process, error) {
	cmd := exec.Command("ps", "aux")
	if filter != "" {
		cmd = exec.Command("sh", "-c", fmt.Sprintf("ps aux | grep %s", filter))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	var processes []Process

	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
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

		process := Process{
			PID:     pid,
			User:    fields[0],
			CPU:     cpu,
			Memory:  mem,
			Command: strings.Join(fields[10:], " "),
		}

		processes = append(processes, process)
	}

	return processes, nil
}

// KillProcess terminates a process (replaces 'kill')
func (sl *StdLib) KillProcess(pid int, signal string) error {
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
func (sl *StdLib) KillProcessByName(name string, signal string) error {
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
func (sl *StdLib) DiskUsage(path string) (map[string]interface{}, error) {
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

	return map[string]interface{}{
		"filesystem":  fields[0],
		"size":        fields[1],
		"used":        fields[2],
		"available":   fields[3],
		"use_percent": strings.TrimSuffix(fields[4], "%"),
		"mounted":     fields[5],
	}, nil
}

// DirSize shows directory size (replaces 'du')
func (sl *StdLib) DirSize(path string) (string, error) {
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
func (sl *StdLib) StartService(name string) error {
	cmd := exec.Command("systemctl", "start", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %s", name, string(output))
	}

	return nil
}

// StopService stops a systemd service (replaces 'systemctl stop')
func (sl *StdLib) StopService(name string) error {
	cmd := exec.Command("systemctl", "stop", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %s", name, string(output))
	}

	return nil
}

// RestartService restarts a systemd service (replaces 'systemctl restart')
func (sl *StdLib) RestartService(name string) error {
	cmd := exec.Command("systemctl", "restart", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to restart service %s: %s", name, string(output))
	}

	return nil
}

// ServiceStatus gets service status (replaces 'systemctl status')
func (sl *StdLib) ServiceStatus(name string) (string, error) {
	cmd := exec.Command("systemctl", "is-active", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get service status: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// ServiceEnabled checks if service is enabled (replaces 'systemctl is-enabled')
func (sl *StdLib) ServiceEnabled(name string) (bool, error) {
	cmd := exec.Command("systemctl", "is-enabled", name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check if service is enabled: %v", err)
	}

	status := strings.TrimSpace(string(output))
	return status == "enabled", nil
}

// GetSystemInfo gets system information (replaces 'uname -a')
func (sl *StdLib) GetSystemInfo() (map[string]string, error) {
	cmd := exec.Command("uname", "-a")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get system info: %v", err)
	}

	return map[string]string{
		"uname": strings.TrimSpace(string(output)),
	}, nil
}

// GetHostname gets system hostname (replaces 'hostname')
func (sl *StdLib) GetHostname() (string, error) {
	cmd := exec.Command("hostname")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get hostname: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentUser gets current username (replaces 'whoami')
func (sl *StdLib) GetCurrentUser() (string, error) {
	cmd := exec.Command("whoami")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetUptime gets system uptime (replaces 'uptime')
func (sl *StdLib) GetUptime() (string, error) {
	cmd := exec.Command("uptime")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get uptime: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetMemoryUsage gets memory usage (replaces 'free')
func (sl *StdLib) GetMemoryUsage() (map[string]interface{}, error) {
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
		total := fields[1]
		used := fields[2]
		available := fields[3]

		return map[string]interface{}{
			"total":     total,
			"used":      used,
			"available": available,
		}, nil
	}

	return nil, fmt.Errorf("unexpected free format")
}
