// Package sqli SQL 注入防护
package sqli

import (
	"database/sql"
	"regexp"
	"strings"
)

// Patterns SQL 注入模式
var Patterns = struct {
	Comment      string
	Union        string
	Conditional  string
	Sequential   string
	Quotes       string
}{
	Comment:     `(--|#|\/\*|\*\/)`,
	Union:       `(?i)\bunion\s+(all\s+)?select\b`,
	Conditional: `(?i)\b(and|or)\s+\d+\s*=\s*\d+`,
	Sequential:  `;.*(?:drop|delete|insert|update|create|alter)`,
	Quotes:      `'.*('|")|".*"(")`,
}

// Validator 输入验证器
type Validator struct {
	patterns []*regexp.Regexp
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	v := &Validator{
		patterns: make([]*regexp.Regexp, 0),
	}

	// 编译所有模式
	for _, pattern := range []string{
		Patterns.Comment,
		Patterns.Union,
		Patterns.Conditional,
		Patterns.Sequential,
		Patterns.Quotes,
	} {
		re, err := regexp.Compile(pattern)
		if err == nil {
			v.patterns = append(v.patterns, re)
		}
	}

	return v
}

// IsValidInput 验证输入是否安全
func (v *Validator) IsValidInput(input string) bool {
	for _, pattern := range v.patterns {
		if pattern.MatchString(input) {
			return false
		}
	}
	return true
}

// Sanitize 清理输入
func (v *Validator) Sanitize(input string) string {
	// 移除危险字符
	input = strings.ReplaceAll(input, "'", "''")
	input = strings.ReplaceAll(input, "--", "")
	input = strings.ReplaceAll(input, ";", "")
	input = strings.ReplaceAll(input, "/*", "")
	input = strings.ReplaceAll(input, "*/", "")

	return strings.TrimSpace(input)
}

// Query 安全查询
func Query(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

// QueryRow 安全查询单行
func QueryRow(db *sql.DB, query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// Exec 安全执行
func Exec(db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	return db.Exec(query, args...)
}

// Quote 转义字符串
func Quote(input string) string {
	return strings.ReplaceAll(input, "'", "''")
}

// EscapeString 转义字符串
func EscapeString(input string) string {
	replacer := strings.NewReplacer(
		"\x00", "\\0",
		"\n", "\\n",
		"\r", "\\r",
		"\\", "\\\\",
		"'", "\\'",
		"\"", "\\\"",
		"\x1a", "\\Z",
	)
	return replacer.Replace(input)
}

// IsSQLInjection 检测是否是 SQL 注入
func IsSQLInjection(input string) bool {
	validator := NewValidator()
	return !validator.IsValidInput(input)
}

// ValidateUsernamePattern 用户名验证模式
const UsernamePattern = `^[a-zA-Z0-9_]{3,20}$`

// ValidateEmailPattern 邮箱验证模式
const EmailPattern = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

// ValidatePattern 验证输入是否符合模式
func ValidatePattern(input, pattern string) bool {
	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		return false
	}
	return matched
}

// ValidateUsername 验证用户名
func ValidateUsername(username string) bool {
	return ValidatePattern(username, UsernamePattern)
}

// ValidateEmail 验证邮箱
func ValidateEmail(email string) bool {
	return ValidatePattern(email, EmailPattern)
}

// Default 默认验证器
var DefaultValidator = NewValidator()

// IsValidInput 验证输入（使用默认验证器）
func IsValidInput(input string) bool {
	return DefaultValidator.IsValidInput(input)
}

// Sanitize 清理输入（使用默认验证器）
func Sanitize(input string) string {
	return DefaultValidator.Sanitize(input)
}
