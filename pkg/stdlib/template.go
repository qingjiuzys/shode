package stdlib

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"regexp"
	"strings"
)

// TemplateEngine represents a simple template engine
type TemplateEngine struct {
	templates map[string]*template.Template
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
	}
}

// SimpleTemplate represents a simple string-based template
type SimpleTemplate struct {
	content string
}

// NewSimpleTemplate creates a new simple template
func NewSimpleTemplate(content string) *SimpleTemplate {
	return &SimpleTemplate{
		content: content,
	}
}

// Render renders the template with given data
func (t *SimpleTemplate) Render(data map[string]interface{}) (string, error) {
	result := t.content

	// Replace {{variable}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(result, -1)

	for _, match := range matches {
		if len(match) > 1 {
			varName := strings.TrimSpace(match[1])

			// Handle special case: {{#if var}}...{{/if}}
			if strings.HasPrefix(varName, "#if ") {
				condition := strings.TrimPrefix(varName, "#if ")
				condition = strings.TrimSpace(condition)
				// Find the matching {{/if}}
				ifResult := t.processIf(result, data, condition)
				result = ifResult
				continue
			}

			// Handle {{#each var}}...{{/each}}
			if strings.HasPrefix(varName, "#each ") {
				varName := strings.TrimPrefix(varName, "#each ")
				varName = strings.TrimSpace(varName)
				eachResult := t.processEach(result, data, varName)
				result = eachResult
				continue
			}

			// Simple variable replacement
			varValue := t.getVarValue(data, varName)
			varValueStr := fmt.Sprintf("%v", varValue)
			result = strings.Replace(result, match[0], varValueStr, -1)
		}
	}

	return result, nil
}

// getVarValue gets the value of a variable from data
func (t *SimpleTemplate) getVarValue(data map[string]interface{}, varName string) interface{} {
	// Handle nested variables like "user.name"
	parts := strings.Split(varName, ".")
	var current interface{} = data

	for _, part := range parts {
		if mapData, ok := current.(map[string]interface{}); ok {
			current = mapData[part]
		} else {
			return nil
		}
	}

	if current == nil {
		return ""
	}
	return current
}

// processIf handles {{#if condition}}...{{/if}} blocks
func (t *SimpleTemplate) processIf(content string, data map[string]interface{}, condition string) string {
	// Simple condition: check if variable exists and is truthy
	value := t.getVarValue(data, condition)
	isTrue := false

	switch v := value.(type) {
	case bool:
		isTrue = v
	case string:
		// Check for string "false" or "0"
		isTrue = v != "" && v != "false" && v != "0"
	case int, int64, float64:
		isTrue = v != 0
	case []interface{}:
		isTrue = len(v) > 0
	case map[string]interface{}:
		isTrue = len(v) > 0
	default:
		isTrue = value != nil
	}

	// Find and replace the if block
	ifPattern := regexp.MustCompile(`\{\{#if\s+` + regexp.QuoteMeta(condition) + `\}\}(.*?)\{\{/if\}\}`)
	matches := ifPattern.FindStringSubmatch(content)

	if len(matches) > 1 {
		if isTrue {
			// Return the content inside the if block
			return ifPattern.ReplaceAllString(content, matches[1])
		} else {
			// Remove the if block entirely
			return ifPattern.ReplaceAllString(content, "")
		}
	}

	return content
}

// processEach handles {{#each items}}...{{/each}} blocks
func (t *SimpleTemplate) processEach(content string, data map[string]interface{}, varName string) string {
	value := t.getVarValue(data, varName)
	var items []interface{}

	switch v := value.(type) {
	case []interface{}:
		items = v
	case []map[string]interface{}:
		for _, item := range v {
			items = append(items, item)
		}
	default:
		// Not an array, return empty
		eachPattern := regexp.MustCompile(`\{\{#each\s+` + regexp.QuoteMeta(varName) + `\}\}.*?\{\{/each\}\}`)
		return eachPattern.ReplaceAllString(content, "")
	}

	// Find the each block
	eachPattern := regexp.MustCompile(`\{\{#each\s+` + regexp.QuoteMeta(varName) + `\}\}(.*?)\{\{/each\}\}`)
	matches := eachPattern.FindStringSubmatch(content)

	if len(matches) > 1 {
		templateContent := matches[1]
		var result strings.Builder

		for i, item := range items {
			// Create a new data map with @index and @item
			itemData := make(map[string]interface{})
			if itemMap, ok := item.(map[string]interface{}); ok {
				for k, v := range itemMap {
					itemData[k] = v
				}
			} else {
				itemMap := map[string]interface{}{"this": item}
				for k, v := range itemMap {
					itemData[k] = v
				}
			}
			itemData["@index"] = i
			itemData["@first"] = i == 0

			// Render the template for this item
			itemTemplate := NewSimpleTemplate(templateContent)
			rendered, err := itemTemplate.Render(itemData)
			if err == nil {
				result.WriteString(rendered)
			}
		}

		return eachPattern.ReplaceAllString(content, result.String())
	}

	return content
}

// RenderTemplateFile renders a template file with given data
func (sl *StdLib) RenderTemplateFile(templatePath string, data map[string]interface{}) (string, error) {
	// Read template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %v", err)
	}

	// Create and render template
	tmpl := NewSimpleTemplate(string(content))
	result, err := tmpl.Render(data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %v", err)
	}

	return result, nil
}

// RenderTemplateString renders a template string with given data
func (sl *StdLib) RenderTemplateString(templateContent string, data map[string]interface{}) (string, error) {
	tmpl := NewSimpleTemplate(templateContent)
	result, err := tmpl.Render(data)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %v", err)
	}

	return result, nil
}

// SaveTemplateFile saves a template to a file
func (sl *StdLib) SaveTemplateFile(path, content string) error {
	return ioutil.WriteFile(path, []byte(content), 0644)
}

// SetHTTPResponseTemplate renders a template file and sets it as HTTP response
func (sl *StdLib) SetHTTPResponseTemplate(status int, templatePath string, data map[string]interface{}) error {
	result, err := sl.RenderTemplateFile(templatePath, data)
	if err != nil {
		return err
	}
	sl.SetHTTPResponse(status, result)
	return nil
}

// SetHTTPResponseTemplateString renders a template string and sets it as HTTP response
func (sl *StdLib) SetHTTPResponseTemplateString(status int, templateContent string, data map[string]interface{}) error {
	result, err := sl.RenderTemplateString(templateContent, data)
	if err != nil {
		return err
	}
	sl.SetHTTPResponse(status, result)
	return nil
}
