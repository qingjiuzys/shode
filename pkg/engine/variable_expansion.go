package engine

import (
	"regexp"
	"strings"
)

// expandVariables expands environment variables in a string
// Supports ${VAR}, $VAR syntax, and direct variable names (for string concatenation)
func (ee *ExecutionEngine) expandVariables(s string) string {
	// First, handle string concatenation (var1 + " " + var2)
	parts := splitStringConcat(s)
	if len(parts) > 1 {
		// This is a concatenation
		result := ""
		for _, part := range parts {
			part = strings.TrimSpace(part)
			// Remove quotes if present
			if len(part) >= 2 && ((part[0] == '"' && part[len(part)-1] == '"') ||
				(part[0] == '\'' && part[len(part)-1] == '\'')) {
				result += part[1 : len(part)-1]
			} else {
				// Expand as variable (check if it's a variable name)
				value := ee.envManager.GetEnv(part)
				if value != "" {
					result += value
				} else {
					// Not a variable, use as-is
					result += part
				}
			}
		}
		return result
	}

	// Expand ${VAR} syntax
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // Extract variable name
		value := ee.envManager.GetEnv(varName)
		if value == "" {
			return match // Return original if not found
		}
		return value
	})

	// Expand $VAR syntax (but not ${VAR} which we already handled)
	re2 := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	s = re2.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[1:] // Extract variable name (skip $)
		value := ee.envManager.GetEnv(varName)
		if value == "" {
			return match // Return original if not found
		}
		return value
	})

	// Check if the entire string is a variable name (for direct variable usage)
	trimmed := strings.TrimSpace(s)
	if trimmed != "" && !strings.Contains(trimmed, " ") && !strings.Contains(trimmed, "$") {
		// Might be a variable name
		value := ee.envManager.GetEnv(trimmed)
		if value != "" {
			return value
		}
	}

	return s
}

// splitStringConcat splits a string by + operator, handling quoted strings
func splitStringConcat(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		char := s[i]

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			}
			current.WriteByte(char)
		} else if char == '+' && !inQuotes {
			// Found + outside quotes - split point
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if len(parts) == 0 {
		return []string{s}
	}

	return parts
}

// expandArgs expands variables in command arguments
func (ee *ExecutionEngine) expandArgs(args []string) []string {
	expanded := make([]string, len(args))
	for i, arg := range args {
		expanded[i] = ee.expandVariables(arg)
	}
	return expanded
}
