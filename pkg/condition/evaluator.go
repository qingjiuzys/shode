package condition

import (
	"fmt"
)

// Condition evaluates a condition
type Condition interface {
	Evaluate(context map[string]interface{}) bool
}

// ConditionEvaluator evaluates conditions
type ConditionEvaluator struct {
	conditions map[string]Condition
}

// NewConditionEvaluator creates a new condition evaluator
func NewConditionEvaluator() *ConditionEvaluator {
	return &ConditionEvaluator{
		conditions: make(map[string]Condition),
	}
}

// RegisterCondition registers a condition
func (ce *ConditionEvaluator) RegisterCondition(name string, condition Condition) {
	ce.conditions[name] = condition
}

// Evaluate evaluates a condition by name
func (ce *ConditionEvaluator) Evaluate(name string, context map[string]interface{}) (bool, error) {
	condition, exists := ce.conditions[name]
	if !exists {
		return false, fmt.Errorf("condition %s not found", name)
	}
	return condition.Evaluate(context), nil
}

// OnClassCondition checks if a class exists
type OnClassCondition struct {
	ClassName string
}

func (c *OnClassCondition) Evaluate(context map[string]interface{}) bool {
	// Check if class exists in context
	classes, ok := context["classes"].(map[string]bool)
	if !ok {
		return false
	}
	return classes[c.ClassName]
}

// OnPropertyCondition checks if a property has a value
type OnPropertyCondition struct {
	Property string
	Value    string
}

func (c *OnPropertyCondition) Evaluate(context map[string]interface{}) bool {
	properties, ok := context["properties"].(map[string]string)
	if !ok {
		return false
	}
	value, exists := properties[c.Property]
	if !exists {
		return false
	}
	return value == c.Value
}

// OnBeanCondition checks if a bean exists
type OnBeanCondition struct {
	BeanName string
}

func (c *OnBeanCondition) Evaluate(context map[string]interface{}) bool {
	beans, ok := context["beans"].(map[string]bool)
	if !ok {
		return false
	}
	return beans[c.BeanName]
}

// NewOnClassCondition creates a new OnClass condition
func NewOnClassCondition(className string) *OnClassCondition {
	return &OnClassCondition{ClassName: className}
}

// NewOnPropertyCondition creates a new OnProperty condition
func NewOnPropertyCondition(property, value string) *OnPropertyCondition {
	return &OnPropertyCondition{Property: property, Value: value}
}

// NewOnBeanCondition creates a new OnBean condition
func NewOnBeanCondition(beanName string) *OnBeanCondition {
	return &OnBeanCondition{BeanName: beanName}
}
