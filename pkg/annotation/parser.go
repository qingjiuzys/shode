package annotation

import (
	"fmt"
	"regexp"
	"strings"
)

// Annotation represents a parsed annotation
type Annotation struct {
	Name      string
	Arguments map[string]string // Key-value pairs from annotation
	Raw       string            // Raw annotation text
}

// Parser parses annotations from source code
type Parser struct {
	annotationRegex *regexp.Regexp
}

// NewParser creates a new annotation parser
func NewParser() *Parser {
	// Match @AnnotationName or @AnnotationName("value") or @AnnotationName(key="value")
	pattern := `@(\w+)(?:\(([^)]*)\))?`
	return &Parser{
		annotationRegex: regexp.MustCompile(pattern),
	}
}

// ParseAnnotation parses a single annotation line
// Example: @Controller("/api/users") -> Annotation{Name: "Controller", Arguments: {"": "/api/users"}}
// Example: @Service -> Annotation{Name: "Service", Arguments: {}}
func (p *Parser) ParseAnnotation(line string) (*Annotation, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "@") {
		return nil, fmt.Errorf("not an annotation line")
	}

	matches := p.annotationRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid annotation format: %s", line)
	}

	annotation := &Annotation{
		Name:      matches[1],
		Arguments: make(map[string]string),
		Raw:       line,
	}

	// Parse arguments if present
	if len(matches) > 2 && matches[2] != "" {
		argsStr := matches[2]
		p.parseArguments(argsStr, annotation)
	}

	return annotation, nil
}

// parseArguments parses annotation arguments
// Supports: "value", key="value", key1="value1", key2="value2"
func (p *Parser) parseArguments(argsStr string, annotation *Annotation) {
	argsStr = strings.TrimSpace(argsStr)

	// If it's a simple quoted string, treat as default argument
	if strings.HasPrefix(argsStr, `"`) && strings.HasSuffix(argsStr, `"`) {
		annotation.Arguments[""] = strings.Trim(argsStr, `"`)
		return
	}

	// Parse key-value pairs
	// Split by comma, but respect quoted strings
	parts := p.splitRespectingQuotes(argsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check if it's key=value format
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			value = strings.Trim(value, `"`)
			annotation.Arguments[key] = value
		} else {
			// Treat as default argument
			annotation.Arguments[""] = strings.Trim(part, `"`)
		}
	}
}

// splitRespectingQuotes splits a string by delimiter, respecting quoted strings
func (p *Parser) splitRespectingQuotes(s, delimiter string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escape := false

	for _, char := range s {
		if escape {
			current.WriteRune(char)
			escape = false
			continue
		}

		if char == '\\' {
			escape = true
			current.WriteRune(char)
			continue
		}

		if char == '"' {
			inQuotes = !inQuotes
			current.WriteRune(char)
			continue
		}

		if !inQuotes && strings.HasPrefix(s[current.Len():], delimiter) {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(char)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// ParseAnnotations parses multiple annotations from a block of code
func (p *Parser) ParseAnnotations(code string) ([]*Annotation, error) {
	lines := strings.Split(code, "\n")
	var annotations []*Annotation

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@") {
			annotation, err := p.ParseAnnotation(line)
			if err != nil {
				return nil, err
			}
			annotations = append(annotations, annotation)
		}
	}

	return annotations, nil
}

// GetValue returns the value of an annotation argument
func (a *Annotation) GetValue(key string) string {
	return a.Arguments[key]
}

// GetDefaultValue returns the default (unnamed) argument value
func (a *Annotation) GetDefaultValue() string {
	return a.Arguments[""]
}

// HasArgument checks if an argument exists
func (a *Annotation) HasArgument(key string) bool {
	_, exists := a.Arguments[key]
	return exists
}
