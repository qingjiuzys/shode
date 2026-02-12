// Package config 提供配置验证功能
package config

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Validator 配置验证器
type Validator struct {
	errors []string
}

// NewValidator 创建验证器
func NewValidator() *Validator {
	return &Validator{
		errors: make([]string, 0),
	}
}

// Validate 验证配置
func (v *Validator) Validate(config interface{}) error {
	val := reflect.ValueOf(config)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct or pointer to struct")
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// 获取验证标签
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// 执行验证
		if err := v.validateField(field, fieldType.Name, validateTag); err != nil {
			v.errors = append(v.errors, err.Error())
		}
	}

	if len(v.errors) > 0 {
		return fmt.Errorf("validation failed:\n  - %s", strings.Join(v.errors, "\n  - "))
	}

	return nil
}

// validateField 验证字段
func (v *Validator) validateField(field reflect.Value, fieldName, tag string) error {
	rules := strings.Split(tag, ",")

	for _, rule := range rules {
		parts := strings.SplitN(rule, "=", 2)
		ruleName := parts[0]
		var ruleValue string
		if len(parts) > 1 {
			ruleValue = parts[1]
		}

		switch ruleName {
		case "required":
			if v.isEmpty(field) {
				return fmt.Errorf("%s is required", fieldName)
			}
		case "min":
			if err := v.validateMin(field, fieldName, ruleValue); err != nil {
				return err
			}
		case "max":
			if err := v.validateMax(field, fieldName, ruleValue); err != nil {
				return err
			}
		case "email":
			if err := v.validateEmail(field, fieldName); err != nil {
				return err
			}
		case "url":
			if err := v.validateURL(field, fieldName); err != nil {
				return err
			}
		case "port":
			if err := v.validatePort(field, fieldName); err != nil {
				return err
			}
		case "ip":
			if err := v.validateIP(field, fieldName); err != nil {
				return err
			}
		case "oneof":
			if err := v.validateOneOf(field, fieldName, ruleValue); err != nil {
				return err
			}
		case "env":
			if err := v.validateEnv(field, fieldName, ruleValue); err != nil {
				return err
			}
		case "file":
			if err := v.validateFile(field, fieldName, ruleValue); err != nil {
				return err
			}
		}
	}

	return nil
}

// isEmpty 检查是否为空
func (v *Validator) isEmpty(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Bool:
		return !field.Bool()
	case reflect.Interface, reflect.Ptr:
		return field.IsNil()
	default:
		return false
	}
}

// validateMin 验证最小值
func (v *Validator) validateMin(field reflect.Value, fieldName, value string) error {
	min, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid min value for %s: %s", fieldName, value)
	}

	switch field.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		if int64(field.Len()) < min {
			return fmt.Errorf("%s length must be at least %d", fieldName, min)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < min {
			return fmt.Errorf("%s must be at least %d", fieldName, min)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() < uint64(min) {
			return fmt.Errorf("%s must be at least %d", fieldName, min)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() < float64(min) {
			return fmt.Errorf("%s must be at least %d", fieldName, min)
		}
	}

	return nil
}

// validateMax 验证最大值
func (v *Validator) validateMax(field reflect.Value, fieldName, value string) error {
	max, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid max value for %s: %s", fieldName, value)
	}

	switch field.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		if int64(field.Len()) > max {
			return fmt.Errorf("%s length must be at most %d", fieldName, max)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > max {
			return fmt.Errorf("%s must be at most %d", fieldName, max)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() > uint64(max) {
			return fmt.Errorf("%s must be at most %d", fieldName, max)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() > float64(max) {
			return fmt.Errorf("%s must be at most %d", fieldName, max)
		}
	}

	return nil
}

// validateEmail 验证邮箱格式
func (v *Validator) validateEmail(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return nil
	}

	email := field.String()
	if !strings.Contains(email, "@") {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}

	if !strings.Contains(parts[1], ".") {
		return fmt.Errorf("%s must be a valid email address", fieldName)
	}

	return nil
}

// validateURL 验证 URL 格式
func (v *Validator) validateURL(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return nil
	}

	url := field.String()
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("%s must be a valid URL starting with http:// or https://", fieldName)
	}

	return nil
}

// validatePort 验证端口号
func (v *Validator) validatePort(field reflect.Value, fieldName string) error {
	var port int

	switch field.Kind() {
	case reflect.String:
		p, err := strconv.Atoi(field.String())
		if err != nil {
			return fmt.Errorf("%s must be a valid port number", fieldName)
		}
		port = p
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		port = int(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		port = int(field.Uint())
	default:
		return nil
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be a valid port number (1-65535)", fieldName)
	}

	return nil
}

// validateIP 验证 IP 地址
func (v *Validator) validateIP(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return nil
	}

	ip := field.String()
	if net.ParseIP(ip) == nil {
		return fmt.Errorf("%s must be a valid IP address", fieldName)
	}

	return nil
}

// validateOneOf 验证枚举值
func (v *Validator) validateOneOf(field reflect.Value, fieldName, values string) error {
	allowed := strings.Split(values, "|")

	switch field.Kind() {
	case reflect.String:
		str := field.String()
		for _, v := range allowed {
			if str == v {
				return nil
			}
		}
		return fmt.Errorf("%s must be one of: %s", fieldName, values)
	default:
		return nil
	}
}

// validateEnv 验证环境变量
func (v *Validator) validateEnv(field reflect.Value, fieldName, envName string) error {
	envValue := os.Getenv(envName)
	if envValue == "" && !v.isEmpty(field) {
		return nil
	}

	if envValue == "" {
		return fmt.Errorf("%s requires environment variable %s to be set", fieldName, envName)
	}

	return nil
}

// validateFile 验证文件存在性
func (v *Validator) validateFile(field reflect.Value, fieldName, mode string) error {
	if field.Kind() != reflect.String {
		return nil
	}

	path := field.String()
	if path == "" {
		return nil
	}

	switch mode {
	case "exists":
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("%s file does not exist: %s", fieldName, path)
		}
	case "readable":
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("%s file is not readable: %s", fieldName, path)
		}
		f.Close()
	}

	return nil
}
