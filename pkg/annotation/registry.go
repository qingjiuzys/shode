package annotation

import (
	"fmt"
	"sync"
)

// Processor processes annotations and performs actions
type Processor interface {
	Process(annotation *Annotation, target interface{}) error
	Supports(annotationName string) bool
}

// Registry manages annotation processors
type Registry struct {
	processors map[string]Processor
	mu         sync.RWMutex
}

// NewRegistry creates a new annotation registry
func NewRegistry() *Registry {
	return &Registry{
		processors: make(map[string]Processor),
	}
}

// RegisterProcessor registers an annotation processor
func (r *Registry) RegisterProcessor(name string, processor Processor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[name] = processor
}

// GetProcessor retrieves a processor for an annotation
func (r *Registry) GetProcessor(annotationName string) (Processor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	processor, exists := r.processors[annotationName]
	if !exists {
		return nil, fmt.Errorf("no processor registered for annotation '%s'", annotationName)
	}

	return processor, nil
}

// ProcessAnnotation processes an annotation
func (r *Registry) ProcessAnnotation(annotation *Annotation, target interface{}) error {
	processor, err := r.GetProcessor(annotation.Name)
	if err != nil {
		return err
	}

	return processor.Process(annotation, target)
}

// HasProcessor checks if a processor is registered
func (r *Registry) HasProcessor(annotationName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.processors[annotationName]
	return exists
}

// GetProcessorNames returns all registered processor names
func (r *Registry) GetProcessorNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.processors))
	for name := range r.processors {
		names = append(names, name)
	}
	return names
}
