package annotation

import (
	"fmt"
	"strings"
)

// ProcessorChain processes annotations in sequence
type ProcessorChain struct {
	processors []Processor
}

// NewProcessorChain creates a new processor chain
func NewProcessorChain(processors ...Processor) *ProcessorChain {
	return &ProcessorChain{
		processors: processors,
	}
}

// Process processes annotations through the chain
func (pc *ProcessorChain) Process(annotation *Annotation, target interface{}) error {
	for _, processor := range pc.processors {
		if processor.Supports(annotation.Name) {
			if err := processor.Process(annotation, target); err != nil {
				return fmt.Errorf("processor failed for annotation '%s': %v", annotation.Name, err)
			}
		}
	}
	return nil
}

// Supports checks if any processor in the chain supports the annotation
func (pc *ProcessorChain) Supports(annotationName string) bool {
	for _, processor := range pc.processors {
		if processor.Supports(annotationName) {
			return true
		}
	}
	return false
}

// Scanner scans code for annotations
type Scanner struct {
	parser *Parser
}

// NewScanner creates a new annotation scanner
func NewScanner() *Scanner {
	return &Scanner{
		parser: NewParser(),
	}
}

// Scan scans code for annotations and returns them with their targets
type AnnotationTarget struct {
	Annotation *Annotation
	Target     string // Function name, struct name, etc.
	Line       int
}

// Scan scans code for annotations
func (s *Scanner) Scan(code string) ([]*AnnotationTarget, error) {
	lines := strings.Split(code, "\n")
	var targets []*AnnotationTarget

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "@") {
			annotation, err := s.parser.ParseAnnotation(line)
			if err != nil {
				return nil, fmt.Errorf("failed to parse annotation at line %d: %v", i+1, err)
			}

			// Try to find the target (next non-empty, non-comment line)
			target := s.findTarget(lines, i+1)
			targets = append(targets, &AnnotationTarget{
				Annotation: annotation,
				Target:     target,
				Line:       i + 1,
			})
		}
	}

	return targets, nil
}

// findTarget finds the target of an annotation (function, struct, etc.)
func (s *Scanner) findTarget(lines []string, startLine int) string {
	for i := startLine; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for function definition
		if strings.HasPrefix(line, "function ") {
			// Extract function name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return strings.TrimSuffix(parts[1], "()")
			}
		}

		return line
	}

	return ""
}
