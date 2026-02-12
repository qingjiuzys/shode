// Package validation 提供数据验证功能
package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

// ValidationErrors 验证错误集合
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validator 验证器接口
type Validator interface {
	Validate(value any) error
}

// RuleFunc 验证规则函数
type RuleFunc func(value any) error

// ValidatorImpl 验证器实现
type ValidatorImpl struct {
	rules []RuleFunc
}

// NewValidator 创建验证器
func NewValidator() *ValidatorImpl {
	return &ValidatorImpl{
		rules: make([]RuleFunc, 0),
	}
}

// AddRule 添加验证规则
func (v *ValidatorImpl) AddRule(rule RuleFunc) *ValidatorImpl {
	v.rules = append(v.rules, rule)
	return v
}

// Validate 执行验证
func (v *ValidatorImpl) Validate(value any) error {
	var errors ValidationErrors

	for _, rule := range v.rules {
		if err := rule(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				errors = append(errors, ValidationError{
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// Required 必填验证
func Required(field string) RuleFunc {
	return func(value any) error {
		if value == nil {
			return ValidationError{
				Field:   field,
				Message: "is required",
			}
		}

		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.String:
			if strings.TrimSpace(rv.String()) == "" {
				return ValidationError{
					Field:   field,
					Message: "is required",
				}
			}
		case reflect.Array, reflect.Slice, reflect.Map:
			if rv.Len() == 0 {
				return ValidationError{
					Field:   field,
					Message: "is required",
				}
			}
		}

		return nil
	}
}

// Min 最小值验证
func Min(field string, min int) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.String:
			if len(rv.String()) < min {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at least %d characters", min),
					Value:   value,
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if rv.Int() < int64(min) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at least %d", min),
					Value:   value,
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if rv.Uint() < uint64(min) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at least %d", min),
					Value:   value,
				}
			}
		case reflect.Float32, reflect.Float64:
			if rv.Float() < float64(min) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at least %d", min),
					Value:   value,
				}
			}
		case reflect.Array, reflect.Slice:
			if rv.Len() < min {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must contain at least %d items", min),
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Max 最大值验证
func Max(field string, max int) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.String:
			if len(rv.String()) > max {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at most %d characters", max),
					Value:   value,
				}
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if rv.Int() > int64(max) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at most %d", max),
					Value:   value,
				}
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if rv.Uint() > uint64(max) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at most %d", max),
					Value:   value,
				}
			}
		case reflect.Float32, reflect.Float64:
			if rv.Float() > float64(max) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must be at most %d", max),
					Value:   value,
				}
			}
		case reflect.Array, reflect.Slice:
			if rv.Len() > max {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must contain at most %d items", max),
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Length 长度验证
func Length(field string, min, max int) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		rv := reflect.ValueOf(value)
		var length int

		switch rv.Kind() {
		case reflect.String:
			length = len(rv.String())
		case reflect.Array, reflect.Slice, reflect.Map:
			length = rv.Len()
		default:
			return nil
		}

		if length < min {
			return ValidationError{
				Field:   field,
				Message: fmt.Sprintf("must be at least %d characters", min),
				Value:   value,
			}
		}

		if length > max {
			return ValidationError{
				Field:   field,
				Message: fmt.Sprintf("must be at most %d characters", max),
				Value:   value,
			}
		}

		return nil
	}
}

// Email 邮箱验证
func Email(field string) RuleFunc {
	pattern := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		if str == "" {
			return nil
		}

		if !pattern.MatchString(str) {
			return ValidationError{
				Field:   field,
				Message: "must be a valid email address",
				Value:   value,
			}
		}

		return nil
	}
}

// URL URL验证
func URL(field string) RuleFunc {
	pattern := regexp.MustCompile(`^https?://[a-zA-Z0-9\-._~:/?#\[\]@!$&'()*+,;=]+$`)

	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		if str == "" {
			return nil
		}

		if !pattern.MatchString(str) {
			return ValidationError{
				Field:   field,
				Message: "must be a valid URL",
				Value:   value,
			}
		}

		return nil
	}
}

// Pattern 正则表达式验证
func Pattern(field string, regex string) RuleFunc {
	pattern := regexp.MustCompile(regex)

	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		if str == "" {
			return nil
		}

		if !pattern.MatchString(str) {
			return ValidationError{
				Field:   field,
				Message: fmt.Sprintf("must match pattern: %s", regex),
				Value:   value,
			}
		}

		return nil
	}
}

// In 包含验证
func In(field string, values []any) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		for _, v := range values {
			if reflect.DeepEqual(value, v) {
				return nil
			}
		}

		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("must be one of: %v", values),
			Value:   value,
		}
	}
}

// NotIn 不包含验证
func NotIn(field string, values []any) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		for _, v := range values {
			if reflect.DeepEqual(value, v) {
				return ValidationError{
					Field:   field,
					Message: fmt.Sprintf("must not be one of: %v", values),
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Alpha 字母验证
func Alpha(field string) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		for _, r := range str {
			if !unicode.IsLetter(r) {
				return ValidationError{
					Field:   field,
					Message: "must contain only letters",
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Alphanumeric 字母数字验证
func Alphanumeric(field string) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		for _, r := range str {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				return ValidationError{
					Field:   field,
					Message: "must contain only letters and numbers",
					Value:   value,
				}
			}
		}

		return nil
	}
}

// Numeric 数字验证
func Numeric(field string) RuleFunc {
	return func(value any) error {
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return ValidationError{
				Field:   field,
				Message: "must be a string",
			}
		}

		for _, r := range str {
			if !unicode.IsDigit(r) {
				return ValidationError{
					Field:   field,
					Message: "must contain only numbers",
					Value:   value,
				}
			}
		}

		return nil
	}
}

// ValidateStruct 验证结构体
func ValidateStruct(s any) error {
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %s", rv.Kind())
	}

	var errors ValidationErrors
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// 获取验证标签
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// 解析验证规则
		rules := strings.Split(tag, ",")

		for _, rule := range rules {
			parts := strings.Split(strings.TrimSpace(rule), "=")
			ruleName := parts[0]

			var ruleFunc RuleFunc
			fieldName := strings.ToLower(fieldType.Name)

			switch ruleName {
			case "required":
				ruleFunc = Required(fieldName)
			case "email":
				ruleFunc = Email(fieldName)
			case "url":
				ruleFunc = URL(fieldName)
			case "alpha":
				ruleFunc = Alpha(fieldName)
			case "alphanumeric":
				ruleFunc = Alphanumeric(fieldName)
			case "numeric":
				ruleFunc = Numeric(fieldName)
			case "min":
				if len(parts) > 1 {
					min := 0
					fmt.Sscanf(parts[1], "%d", &min)
					ruleFunc = Min(fieldName, min)
				}
			case "max":
				if len(parts) > 1 {
					max := 0
					fmt.Sscanf(parts[1], "%d", &max)
					ruleFunc = Max(fieldName, max)
				}
			}

			if ruleFunc != nil {
				if err := ruleFunc(field.Interface()); err != nil {
					if ve, ok := err.(ValidationError); ok {
						errors = append(errors, ve)
					}
				}
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// ValidateMap 验证Map
func ValidateMap(data map[string]any, rules map[string]RuleFunc) error {
	var errors ValidationErrors

	for field, ruleFunc := range rules {
		value := data[field]
		if err := ruleFunc(value); err != nil {
			if ve, ok := err.(ValidationError); ok {
				errors = append(errors, ve)
			} else {
				errors = append(errors, ValidationError{
					Field:   field,
					Message: err.Error(),
				})
			}
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
