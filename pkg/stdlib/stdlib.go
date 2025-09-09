package stdlib

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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

// String functions

// Contains checks if a string contains another string (replaces grep)
func (sl *StdLib) Contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// Replace replaces all occurrences of old with new in a string (replaces sed)
func (sl *StdLib) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
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

// ==================== 扩展文件系统操作 ====================

// CopyFile copies a file from source to destination (replaces cp)
func (sl *StdLib) CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %v", src, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %v", dst, err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %v", err)
	}

	return nil
}

// MoveFile moves/renames a file (replaces mv)
func (sl *StdLib) MoveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// DeleteFile deletes a file (replaces rm)
func (sl *StdLib) DeleteFile(filename string) error {
	return os.Remove(filename)
}

// DeleteDir deletes a directory recursively (replaces rm -rf)
func (sl *StdLib) DeleteDir(dirpath string) error {
	return os.RemoveAll(dirpath)
}

// MakeDir creates a directory (replaces mkdir)
func (sl *StdLib) MakeDir(dirpath string) error {
	return os.MkdirAll(dirpath, 0755)
}

// FileSize gets the size of a file in bytes (replaces wc -c)
func (sl *StdLib) FileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FileModTime gets the modification time of a file
func (sl *StdLib) FileModTime(filename string) (time.Time, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// IsDir checks if a path is a directory (replaces test -d)
func (sl *StdLib) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// IsFile checks if a path is a regular file (replaces test -f)
func (sl *StdLib) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// Chmod changes file mode (replaces chmod)
func (sl *StdLib) Chmod(filename string, mode os.FileMode) error {
	return os.Chmod(filename, mode)
}

// Chown changes file owner (replaces chown)
func (sl *StdLib) Chown(filename string, uid, gid int) error {
	return os.Chown(filename, uid, gid)
}

// Glob finds files matching a pattern (replaces find with pattern)
func (sl *StdLib) Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

// Walk walks a directory tree (replaces find)
func (sl *StdLib) Walk(root string, walkFn func(path string, info os.FileInfo, err error) error) error {
	return filepath.Walk(root, walkFn)
}

// ==================== 扩展字符串操作 ====================

// Split splits a string by separator (replaces awk/split)
func (sl *StdLib) Split(s, sep string) []string {
	return strings.Split(s, sep)
}

// Join joins strings with separator (replaces paste)
func (sl *StdLib) Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}

// HasPrefix checks if string has prefix (replaces grep ^)
func (sl *StdLib) HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix checks if string has suffix (replaces grep $)
func (sl *StdLib) HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Index returns the index of substring (replaces grep -n)
func (sl *StdLib) Index(s, substr string) int {
	return strings.Index(s, substr)
}

// LastIndex returns the last index of substring
func (sl *StdLib) LastIndex(s, substr string) int {
	return strings.LastIndex(s, substr)
}

// Count counts occurrences of substring (replaces grep -c)
func (sl *StdLib) Count(s, substr string) int {
	return strings.Count(s, substr)
}

// Repeat repeats a string multiple times
func (sl *StdLib) Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// Compare compares two strings lexicographically
func (sl *StdLib) Compare(a, b string) int {
	return strings.Compare(a, b)
}

// ==================== 正则表达式操作 ====================

// RegexMatch checks if string matches regex pattern (replaces grep -E)
func (sl *StdLib) RegexMatch(pattern, s string) (bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.MatchString(s), nil
}

// RegexFind finds first match of regex pattern
func (sl *StdLib) RegexFind(pattern, s string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.FindString(s), nil
}

// RegexFindAll finds all matches of regex pattern
func (sl *StdLib) RegexFindAll(pattern, s string) ([]string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return re.FindAllString(s, -1), nil
}

// RegexReplace replaces regex matches (replaces sed -E)
func (sl *StdLib) RegexReplace(pattern, replacement, s string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(s, replacement), nil
}

// ==================== 系统信息操作 ====================

// Hostname returns the hostname (replaces hostname)
func (sl *StdLib) Hostname() (string, error) {
	return os.Hostname()
}

// GetUsername returns the current username (replaces whoami)
func (sl *StdLib) GetUsername() string {
	return os.Getenv("USER")
}

// GetPID returns the current process ID
func (sl *StdLib) GetPID() int {
	return os.Getpid()
}

// GetPPID returns the parent process ID
func (sl *StdLib) GetPPID() int {
	return os.Getppid()
}

// Sleep pauses execution for specified duration (replaces sleep)
func (sl *StdLib) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// Now returns current time
func (sl *StdLib) Now() time.Time {
	return time.Now()
}

// ==================== 网络操作 ====================

// HTTPGet performs HTTP GET request (replaces curl)
func (sl *StdLib) HTTPGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// HTTPPost performs HTTP POST request
func (sl *StdLib) HTTPPost(url, contentType, data string) (string, error) {
	resp, err := http.Post(url, contentType, strings.NewReader(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// ==================== 加密哈希操作 ====================

// MD5Hash computes MD5 hash of a string
func (sl *StdLib) MD5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// SHA1Hash computes SHA1 hash of a string
func (sl *StdLib) SHA1Hash(s string) string {
	hash := sha1.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash computes SHA256 hash of a string
func (sl *StdLib) SHA256Hash(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

// Base64Encode encodes string to base64
func (sl *StdLib) Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64Decode decodes base64 string
func (sl *StdLib) Base64Decode(s string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// ==================== 数据处理 ====================

// JSONStringify converts object to JSON string
func (sl *StdLib) JSONStringify(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// JSONParse parses JSON string to object
func (sl *StdLib) JSONParse(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}

// ==================== 进程执行 ====================

// Exec executes an external command (replaces system call)
func (sl *StdLib) Exec(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}

// ExecWithTimeout executes command with timeout
func (sl *StdLib) ExecWithTimeout(timeout time.Duration, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	
	// Set up timeout
	timer := time.AfterFunc(timeout, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}
	return string(output), nil
}
