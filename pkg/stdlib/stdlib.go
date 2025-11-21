package stdlib

import (
	"bufio"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// StdLib provides built-in functions to replace external commands
type StdLib struct{}

// New creates a new standard library instance
func New() *StdLib {
	return &StdLib{}
}

// FileSystem functions

// ReadFile reads the contents of a file (replaces 'cat')
func (sl *StdLib) ReadFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return string(content), nil
}

// WriteFile writes content to a file (replaces echo > file)
func (sl *StdLib) WriteFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// ListFiles lists files in a directory (replaces 'ls')
func (sl *StdLib) ListFiles(dirpath string) ([]string, error) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %v", dirpath, err)
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}

// FileExists checks if a file exists (replaces test -f)
func (sl *StdLib) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// CopyFile copies a file from src to dst
func (sl *StdLib) CopyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	dest, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, source); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// Move moves a file or directory
func (sl *StdLib) Move(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		if info.IsDir() {
			return fmt.Errorf("failed to move directory across filesystems: %w", err)
		}
		// fallback copy then remove
		if err := sl.CopyFile(src, dst); err != nil {
			return err
		}
		return os.Remove(src)
	}
	return nil
}

// MkdirAll creates a directory tree
func (sl *StdLib) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

// Remove deletes a file or directory tree
func (sl *StdLib) Remove(path string) error {
	return os.RemoveAll(path)
}

// Glob expands glob patterns
func (sl *StdLib) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// TempFile creates a temporary file and returns its path
func (sl *StdLib) TempFile(prefix string) (string, error) {
	file, err := os.CreateTemp("", prefix)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer file.Close()
	return file.Name(), nil
}

// Touch updates timestamps or creates the file
func (sl *StdLib) Touch(path string) error {
	now := time.Now()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := ioutil.WriteFile(path, []byte{}, 0644); err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}
	return os.Chtimes(path, now, now)
}

// Chmod changes permissions
func (sl *StdLib) Chmod(path string, perm os.FileMode) error {
	return os.Chmod(path, perm)
}

// Chown changes ownership (Unix only, other systems may return error)
func (sl *StdLib) Chown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

// Head returns first n lines of file
func (sl *StdLib) Head(path string, n int) ([]string, error) {
	if n <= 0 {
		return []string{}, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= n {
			break
		}
	}
	return lines, scanner.Err()
}

// Tail returns last n lines of file
func (sl *StdLib) Tail(path string, n int) ([]string, error) {
	if n <= 0 {
		return []string{}, nil
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) <= n {
		return lines, nil
	}
	return lines[len(lines)-n:], nil
}

// DiskUsage sums file sizes rooted at path
func (sl *StdLib) DiskUsage(path string) (int64, error) {
	var total int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// FindFiles matches files by pattern recursively
func (sl *StdLib) FindFiles(root, pattern string) ([]string, error) {
	var matches []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if ok, _ := filepath.Match(pattern, info.Name()); ok {
			matches = append(matches, path)
		}
		return nil
	})
	return matches, err
}

// ChecksumSHA256 computes SHA256 for a file
func (sl *StdLib) ChecksumSHA256(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// String functions

// Contains checks if a string contains another string (replaces grep)
func (sl *StdLib) Contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// Replace replaces all occurrences of old with new in a string (replaces sed)
func (sl *StdLib) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// GrepLines returns lines matching the substring needle
func (sl *StdLib) GrepLines(text, needle string) []string {
	var matches []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, needle) {
			matches = append(matches, line)
		}
	}
	return matches
}

// GrepRegex returns lines matching the regex pattern
func (sl *StdLib) GrepRegex(text, pattern string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}
	var matches []string
	scanner := bufio.NewScanner(strings.NewReader(text))
	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			matches = append(matches, line)
		}
	}
	return matches, nil
}

// ToUpper converts string to uppercase (replaces tr '[:lower:]' '[:upper:]')
func (sl *StdLib) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower converts string to lowercase (replaces tr '[:upper:]' '[:lower:]')
func (sl *StdLib) ToLower(s string) string {
	return strings.ToLower(s)
}

// Trim removes leading and trailing whitespace (replaces sed trimming)
func (sl *StdLib) Trim(s string) string {
	return strings.TrimSpace(s)
}

// Split splits a string using the provided separator
func (sl *StdLib) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Join joins strings using the provided separator
func (sl *StdLib) Join(parts []string, sep string) string {
	return strings.Join(parts, sep)
}

// MatchRegex tests whether value matches the regex pattern
func (sl *StdLib) MatchRegex(pattern, value string) (bool, error) {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return false, fmt.Errorf("invalid regex pattern: %w", err)
	}
	return matched, nil
}

// ReplaceRegex replaces regex matches with replacement
func (sl *StdLib) ReplaceRegex(pattern, replacement, value string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("invalid regex pattern: %w", err)
	}
	return re.ReplaceAllString(value, replacement), nil
}

// Environment functions

// GetEnv gets an environment variable (replaces $VAR)
func (sl *StdLib) GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets an environment variable (replaces export)
func (sl *StdLib) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// WorkingDir gets the current working directory (replaces pwd)
func (sl *StdLib) WorkingDir() (string, error) {
	return os.Getwd()
}

// ChangeDir changes the current directory (replaces cd)
func (sl *StdLib) ChangeDir(dirpath string) error {
	return os.Chdir(dirpath)
}

// Hostname returns the system hostname
func (sl *StdLib) Hostname() (string, error) {
	return os.Hostname()
}

// CurrentUser returns current username
func (sl *StdLib) CurrentUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

// Data functions

// JSONEncodeMap encodes a map as JSON string
func (sl *StdLib) JSONEncodeMap(data map[string]interface{}) (string, error) {
	if data == nil {
		data = make(map[string]interface{})
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to encode json: %w", err)
	}
	return string(bytes), nil
}

// JSONDecodeToMap decodes JSON string into a map
func (sl *StdLib) JSONDecodeToMap(raw string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	return result, nil
}

// JSONPretty formats JSON with indentation
func (sl *StdLib) JSONPretty(raw string) (string, error) {
	var obj interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return "", fmt.Errorf("failed to parse json: %w", err)
	}
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format json: %w", err)
	}
	return string(bytes), nil
}

// System & utility functions

// SleepSeconds pauses execution for N seconds
func (sl *StdLib) SleepSeconds(seconds int) {
	if seconds < 0 {
		seconds = 0
	}
	time.Sleep(time.Duration(seconds) * time.Second)
}

// TimeNowRFC3339 returns current time in RFC3339 format
func (sl *StdLib) TimeNowRFC3339() string {
	return time.Now().Format(time.RFC3339)
}

// GenerateUUID returns a random RFC4122 UUID string
func (sl *StdLib) GenerateUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate uuid: %w", err)
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	hexStr := hex.EncodeToString(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexStr[0:8],
		hexStr[8:12],
		hexStr[12:16],
		hexStr[16:20],
		hexStr[20:32],
	), nil
}

// HTTPGet performs a simple HTTP GET request with timeout (seconds)
func (sl *StdLib) HTTPGet(rawURL string, timeoutSeconds int) (string, error) {
	parsed, err := validateHTTPURL(rawURL)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: httpTimeout(timeoutSeconds)}
	resp, err := client.Get(parsed.String())
	if err != nil {
		return "", fmt.Errorf("http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("http get failed with status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// HTTPPostJSON performs an HTTP POST with JSON body
func (sl *StdLib) HTTPPostJSON(rawURL string, body string, timeoutSeconds int) (string, error) {
	parsed, err := validateHTTPURL(rawURL)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: httpTimeout(timeoutSeconds)}
	resp, err := client.Post(parsed.String(), "application/json", strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("http post failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("http post failed with status %d", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(respBody), nil
}

// Utility functions

// Print outputs text to stdout (replaces echo)
func (sl *StdLib) Print(text string) {
	fmt.Print(text)
}

// Println outputs text with newline to stdout (replaces echo)
func (sl *StdLib) Println(text string) {
	fmt.Println(text)
}

// Error outputs text to stderr (replaces echo >&2)
func (sl *StdLib) Error(text string) {
	fmt.Fprint(os.Stderr, text)
}

// Errorln outputs text with newline to stderr (replaces echo >&2)
func (sl *StdLib) Errorln(text string) {
	fmt.Fprintln(os.Stderr, text)
}

func validateHTTPURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("unsupported url scheme: %s", parsed.Scheme)
	}
	return parsed, nil
}

func httpTimeout(seconds int) time.Duration {
	if seconds <= 0 {
		return 10 * time.Second
	}
	return time.Duration(seconds) * time.Second
}
