package web

import (
	"fmt"
	"reflect"
	"strings"
)

// Validator validates data
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a value based on rules
func (v *Validator) Validate(value interface{}, rules ...ValidationRule) error {
	for _, rule := range rules {
		if err := rule.Validate(value); err != nil {
			return err
		}
	}
	return nil
}

// ValidationRule defines a validation rule
type ValidationRule interface {
	Validate(value interface{}) error
}

// RequiredRule validates that a value is not empty
type RequiredRule struct{}

func (r *RequiredRule) Validate(value interface{}) error {
	if value == nil {
		return fmt.Errorf("value is required")
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		if strings.TrimSpace(val.String()) == "" {
			return fmt.Errorf("value is required")
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if val.Len() == 0 {
			return fmt.Errorf("value is required")
		}
	}

	return nil
}

// MinLengthRule validates minimum length
type MinLengthRule struct {
	Min int
}

func (r *MinLengthRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.String {
		return fmt.Errorf("min length rule only applies to strings")
	}

	if len(val.String()) < r.Min {
		return fmt.Errorf("value must be at least %d characters", r.Min)
	}

	return nil
}

// MaxLengthRule validates maximum length
type MaxLengthRule struct {
	Max int
}

func (r *MaxLengthRule) Validate(value interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.String {
		return fmt.Errorf("max length rule only applies to strings")
	}

	if len(val.String()) > r.Max {
		return fmt.Errorf("value must be at most %d characters", r.Max)
	}

	return nil
}

// NewRequiredRule creates a required validation rule
func NewRequiredRule() *RequiredRule {
	return &RequiredRule{}
}

// NewMinLengthRule creates a min length validation rule
func NewMinLengthRule(min int) *MinLengthRule {
	return &MinLengthRule{Min: min}
}

// NewMaxLengthRule creates a max length validation rule
func NewMaxLengthRule(max int) *MaxLengthRule {
	return &MaxLengthRule{Max: max}
}
