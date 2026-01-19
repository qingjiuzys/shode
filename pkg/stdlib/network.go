package stdlib

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// PingResult represents ping result
type PingResult struct {
	Host            string
	PacketsSent     int
	PacketsReceived int
	PacketLoss      float64
	AvgTime         float64
}

// HTTPRequest sends an HTTP request (replaces 'curl')
func (sl *StdLib) HTTPRequest(method, url string, headers map[string]string, body string) (string, int, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	var reqBody io.Reader
	if body != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
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
func (sl *StdLib) Ping(host string, count int) (*PingResult, error) {
	cmd := exec.Command("ping", "-c", strconv.Itoa(count), host)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ping failed: %v", err)
	}

	// Parse ping output
	lines := strings.Split(string(output), "\n")
	result := &PingResult{
		Host:        host,
		PacketsSent: count,
	}

	// Extract statistics
	for _, line := range lines {
		if strings.Contains(line, "packets transmitted") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				received, _ := strconv.Atoi(fields[3])
				result.PacketsReceived = received
			}
		}

		if strings.Contains(line, "packet loss") {
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				loss, _ := strconv.ParseFloat(strings.TrimSuffix(fields[5], "%"), 64)
				result.PacketLoss = loss
			}
		}

		if strings.Contains(line, "avg") {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				avg, _ := strconv.ParseFloat(fields[6], 64)
				result.AvgTime = avg
			}
		}
	}

	return result, nil
}

// DownloadFile downloads a file from URL (replaces 'wget')
func (sl *StdLib) DownloadFile(url, dstPath string) (string, error) {
	cmd := exec.Command("wget", "-O", dstPath, url)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("wget failed: %s", string(output))
	}

	return string(output), nil
}

// Netstat shows network connections (replaces 'netstat' or 'ss')
func (sl *StdLib) Netstat(proto string) (string, error) {
	// Try ss first (more modern)
	cmd := exec.Command("ss", "-tunlp")
	if proto != "" {
		cmd = exec.Command("ss", "-tunlp", strings.ToLower(proto))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to netstat
		cmd = exec.Command("netstat", "-tunlp")
		output, err = cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("netstat/ss failed: %v", err)
		}
	}

	return string(output), nil
}

// GetLocalIP gets local IP address
func (sl *StdLib) GetLocalIP() (string, error) {
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
