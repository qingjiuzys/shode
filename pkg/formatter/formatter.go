package formatter

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var supportedExtensions = map[string]struct{}{
	".sh":    {},
	".sho":   {},
	".shode": {},
}

// FormatScript applies a simple indentation-style formatter to Shode scripts.
func FormatScript(source string) string {
	lines := strings.Split(source, "\n")
	var buf bytes.Buffer
	indent := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			if i != len(lines)-1 {
				buf.WriteString("\n")
			}
			continue
		}

		lower := strings.ToLower(trimmed)
		if startsClosingBlock(lower) {
			if indent > 0 {
				indent--
			}
		}

		buf.WriteString(strings.Repeat("  ", indent))
		buf.WriteString(normalizeComment(trimmed))

		if i != len(lines)-1 {
			buf.WriteString("\n")
		}

		if endsOpeningBlock(lower) {
			indent++
		}
	}

	if buf.Len() == 0 || !strings.HasSuffix(buf.String(), "\n") {
		buf.WriteString("\n")
	}

	return buf.String()
}

func startsClosingBlock(line string) bool {
	switch {
	case line == "fi",
		line == "done",
		line == "}",
		line == "elif",
		line == "else",
		line == "esac":
		return true
	case strings.HasPrefix(line, "elif "):
		return true
	case strings.HasPrefix(line, "}"):
		return true
	}
	return false
}

func endsOpeningBlock(line string) bool {
	if strings.HasSuffix(line, " then") ||
		strings.HasSuffix(line, " do") ||
		strings.HasSuffix(line, "{") ||
		line == "do" ||
		line == "then" ||
		line == "{" {
		return true
	}
	if line == "else" || strings.HasPrefix(line, "else ") {
		return true
	}
	if strings.HasPrefix(line, "function ") && strings.HasSuffix(line, "{") {
		return true
	}
	return false
}

func normalizeComment(line string) string {
	if !strings.HasPrefix(line, "#") {
		return line
	}
	if len(line) == 1 {
		return "#"
	}
	content := strings.TrimSpace(line[1:])
	if content == "" {
		return "#"
	}
	return "# " + content
}

// FormatFile formats a file and optionally writes the result back.
// Returns true if the file changed.
func FormatFile(path string, write bool) (bool, error) {
	original, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read file: %w", err)
	}

	formatted := FormatScript(string(original))
	if formatted == string(original) {
		return false, nil
	}

	if write {
		if err := os.WriteFile(path, []byte(formatted), 0o644); err != nil {
			return false, fmt.Errorf("write file: %w", err)
		}
	}

	return true, nil
}

// FormatPath formats all supported files under the provided paths.
func FormatPath(paths []string, write bool) ([]string, error) {
	var changed []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return changed, err
		}

		if info.IsDir() {
			err = filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					if shouldSkipDir(d.Name()) {
						return filepath.SkipDir
					}
					return nil
				}
				if !IsSupportedFile(path) {
					return nil
				}
				c, err := FormatFile(path, write)
				if err != nil {
					return err
				}
				if c {
					changed = append(changed, path)
				}
				return nil
			})
			if err != nil {
				return changed, err
			}
			continue
		}

		if !IsSupportedFile(p) {
			continue
		}
		c, err := FormatFile(p, write)
		if err != nil {
			return changed, err
		}
		if c {
			changed = append(changed, p)
		}
	}
	return changed, nil
}

// IsSupportedFile reports whether a path has a supported script extension.
func IsSupportedFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	_, ok := supportedExtensions[ext]
	return ok
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".git", "node_modules", "sh_models":
		return true
	}
	return strings.HasPrefix(name, ".")
}
