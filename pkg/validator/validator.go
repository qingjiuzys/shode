// Package validator 提供数据验证工具
package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"unicode"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error 实现error接口
func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

// Error 实现error接口
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	msg := "validation failed:"
	for _, e := range ve {
		msg += "\n  - " + e.Error()
	}
	return msg
}

// Validator 验证器接口
type Validator interface {
	Validate(value any) error
}

// FuncValidator 函数验证器
type FuncValidator struct {
	Name string
	Fn   func(any) error
}

// Validate 验证
func (fv *FuncValidator) Validate(value any) error {
	return fv.Fn(value)
}

// Rule 验证规则
type Rule struct {
	name   string
	check  func(any) bool
	message func(any) string
}

// Common rules
var (
	// Required 必填
	Required = &Rule{
		name: "required",
		check: func(v any) bool {
			if v == nil {
				return false
			}
			rv := reflect.ValueOf(v)
			switch rv.Kind() {
			case reflect.String:
				return rv.String() != ""
			case reflect.Slice, reflect.Array, reflect.Map:
				return rv.Len() > 0
			default:
				return !reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
			}
		},
		message: func(v any) string {
			return "is required"
		},
	}

	// Email 邮箱
	Email = &Rule{
		name: "email",
		check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
			return pattern.MatchString(s)
		},
		message: func(v any) string {
			return "must be a valid email"
		},
	}

	// URL URL
	URL = &Rule{
		name: "url",
		check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			pattern := regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`)
			return pattern.MatchString(s)
		},
		message: func(v any) string {
			return "must be a valid URL"
		},
	}

	// MinLength 最小长度
	MinLength = func(min int) *Rule {
		return &Rule{
			name: fmt.Sprintf("min_length(%d)", min),
			check: func(v any) bool {
				s, ok := v.(string)
				if !ok {
					return false
				}
				return len(s) >= min
			},
			message: func(v any) string {
				return fmt.Sprintf("must be at least %d characters", min)
			},
		}
	}

	// MaxLength 最大长度
	MaxLength = func(max int) *Rule {
		return &Rule{
			name: fmt.Sprintf("max_length(%d)", max),
			check: func(v any) bool {
				s, ok := v.(string)
				if !ok {
					return false
				}
				return len(s) <= max
			},
			message: func(v any) string {
				return fmt.Sprintf("must be at most %d characters", max)
			},
		}
	}

	// Length 长度范围
	Length = func(min, max int) *Rule {
		return &Rule{
			name: fmt.Sprintf("length(%d,%d)", min, max),
			check: func(v any) bool {
				s, ok := v.(string)
				if !ok {
					return false
				}
				l := len(s)
				return l >= min && l <= max
			},
			message: func(v any) string {
				return fmt.Sprintf("must be between %d and %d characters", min, max)
			},
		}
	}

	// Min 最小值
	Min = func(min int) *Rule {
		return &Rule{
			name: fmt.Sprintf("min(%d)", min),
			check: func(v any) bool {
				var n int
				switch val := v.(type) {
				case int:
					n = val
				case int64:
					n = int(val)
				case float64:
					n = int(val)
				default:
					return false
				}
				return n >= min
			},
			message: func(v any) string {
				return fmt.Sprintf("must be at least %d", min)
			},
		}
	}

	// Max 最大值
	Max = func(max int) *Rule {
		return &Rule{
			name: fmt.Sprintf("max(%d)", max),
			check: func(v any) bool {
				var n int
				switch val := v.(type) {
				case int:
					n = val
				case int64:
					n = int(val)
				case float64:
					n = int(val)
				default:
					return false
				}
				return n <= max
			},
			message: func(v any) string {
				return fmt.Sprintf("must be at most %d", max)
			},
		}
	}

	// Range 范围
	Range = func(min, max int) *Rule {
		return &Rule{
			name: fmt.Sprintf("range(%d,%d)", min, max),
			check: func(v any) bool {
				var n int
				switch val := v.(type) {
				case int:
					n = val
				case int64:
					n = int(val)
				case float64:
					n = int(val)
				default:
					return false
				}
				return n >= min && n <= max
			},
			message: func(v any) string {
				return fmt.Sprintf("must be between %d and %d", min, max)
			},
		}
	}

	// Pattern 正则表达式
	Pattern = func(pattern string) *Rule {
		re := regexp.MustCompile(pattern)
		return &Rule{
			name: fmt.Sprintf("pattern(%s)", pattern),
			check: func(v any) bool {
				s, ok := v.(string)
				if !ok {
					return false
				}
				return re.MatchString(s)
			},
			message: func(v any) string {
				return fmt.Sprintf("must match pattern: %s", pattern)
			},
		}
	}

	// AlphaOnly 只包含字母
	AlphaOnly = &Rule{
		name: "alpha",
		check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			for _, r := range s {
				if !unicode.IsLetter(r) {
					return false
				}
			}
			return true
		},
		message: func(v any) string {
			return "must contain only letters"
		},
	}

	// Alphanumeric 只包含字母数字
	Alphanumeric = &Rule{
		name: "alphanumeric",
		check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			for _, r := range s {
				if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
					return false
				}
			}
			return true
		},
		message: func(v any) string {
			return "must contain only letters and numbers"
		},
	}

	// Numeric 只包含数字
	Numeric = &Rule{
		name: "numeric",
		check: func(v any) bool {
			s, ok := v.(string)
			if !ok {
				return false
			}
			for _, r := range s {
				if !unicode.IsDigit(r) {
					return false
				}
			}
			return true
		},
		message: func(v any) string {
			return "must contain only numbers"
		},
	}

	// Positive 正数
	Positive = &Rule{
		name: "positive",
		check: func(v any) bool {
			var n float64
			switch val := v.(type) {
			case int:
				n = float64(val)
			case int64:
				n = float64(val)
			case float64:
				n = val
			default:
				return false
			}
			return n > 0
		},
		message: func(v any) string {
			return "must be positive"
		},
	}

	// Negative 负数
	Negative = &Rule{
		name: "negative",
		check: func(v any) bool {
			var n float64
			switch val := v.(type) {
			case int:
				n = float64(val)
			case int64:
				n = float64(val)
			case float64:
				n = val
			default:
				return false
			}
			return n < 0
		},
		message: func(v any) string {
			return "must be negative"
		},
	}

	// NonZero 非零
	NonZero = &Rule{
		name: "non_zero",
		check: func(v any) bool {
			var n float64
			switch val := v.(type) {
			case int:
				n = float64(val)
			case int64:
				n = float64(val)
			case float64:
				n = val
			default:
				return false
			}
			return n != 0
		},
		message: func(v any) string {
			return "must not be zero"
		},
	}

	// In 在列表中
	In = func(values ...any) *Rule {
		return &Rule{
			name: fmt.Sprintf("in(%v)", values),
			check: func(v any) bool {
				for _, val := range values {
					if reflect.DeepEqual(v, val) {
						return true
					}
				}
				return false
			},
			message: func(v any) string {
				return fmt.Sprintf("must be one of: %v", values)
			},
		}
	}

	// NotIn 不在列表中
	NotIn = func(values ...any) *Rule {
		return &Rule{
			name: fmt.Sprintf("not_in(%v)", values),
			check: func(v any) bool {
				for _, val := range values {
					if reflect.DeepEqual(v, val) {
						return false
					}
				}
				return true
			},
			message: func(v any) string {
				return fmt.Sprintf("must not be one of: %v", values)
			},
		}
	}
)

// Validate 验证值
func Validate(value any, rules ...*Rule) error {
	var errors ValidationErrors

	for _, rule := range rules {
		if !rule.check(value) {
			errors = append(errors, ValidationError{
				Message: rule.message(value),
				Value:   value,
			})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateStruct 验证结构体
func ValidateStruct(s any) error {
	v := reflect.ValueOf(s)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", v.Kind())
	}

	var errors ValidationErrors
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i).Interface()

		// 获取验证规则标签
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// 解析规则
		rules, err := parseRules(tag)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   field.Name,
				Message: err.Error(),
			})
			continue
		}

		// 验证字段
		for _, rule := range rules {
			if !rule.check(fieldValue) {
				errors = append(errors, ValidationError{
					Field:   field.Name,
					Message: rule.message(fieldValue),
					Value:   fieldValue,
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// parseRules 解析规则字符串
func parseRules(tag string) ([]*Rule, error) {
	var rules []*Rule

	// 简化实现，实际应该解析更复杂的规则
	if tag == "required" {
		rules = append(rules, Required)
	} else if tag == "email" {
		rules = append(rules, Email)
	}

	return rules, nil
}

// IsEmail 验证邮箱
func IsEmail(email string) bool {
	return Email.check(email)
}

// IsURL 验证URL
func IsURL(url string) bool {
	return URL.check(url)
}

// IsAlpha 验证字母
func IsAlpha(s string) bool {
	return AlphaOnly.check(s)
}

// IsAlphanumeric 验证字母数字
func IsAlphanumeric(s string) bool {
	return Alphanumeric.check(s)
}

// IsNumeric 验证数字
func IsNumeric(s string) bool {
	return Numeric.check(s)
}

// IsPositive 验证正数
func IsPositive(n any) bool {
	return Positive.check(n)
}

// IsNegative 验证负数
func IsNegative(n any) bool {
	return Negative.check(n)
}

// IsEmpty 验证是否为空
func IsEmpty(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	default:
		return false
	}
}

// HasMinLength 验证最小长度
func HasMinLength(s string, min int) bool {
	return len(s) >= min
}

// HasMaxLength 验证最大长度
func HasMaxLength(s string, max int) bool {
	return len(s) <= max
}

// IsInRange 验证范围
func IsInRange(n, min, max int) bool {
	return n >= min && n <= max
}

// MatchesPattern 验证正则表达式
func MatchesPattern(s, pattern string) bool {
	re := regexp.MustCompile(pattern)
	return re.MatchString(s)
}

// IsInSlice 验证是否在切片中
func IsInSlice[T comparable](item T, slice []T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// IsUnique 验证切片元素是否唯一
func IsUnique[T comparable](slice []T) bool {
	seen := make(map[T]struct{})
	for _, v := range slice {
		if _, exists := seen[v]; exists {
			return false
		}
		seen[v] = struct{}{}
	}
	return true
}

// ValidateMap 验证map
func ValidateMap(m map[string]any, rules map[string][]*Rule) error {
	var errors ValidationErrors

	for field, value := range m {
		if fieldRules, ok := rules[field]; ok {
			for _, rule := range fieldRules {
				if !rule.check(value) {
					errors = append(errors, ValidationError{
						Field:   field,
						Message: rule.message(value),
						Value:   value,
					})
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
