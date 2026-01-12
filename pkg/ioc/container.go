package ioc

import (
	"fmt"
	"reflect"
	"sync"
)

// BeanScope represents the scope of a bean
type BeanScope string

const (
	ScopeSingleton BeanScope = "singleton"
	ScopePrototype BeanScope = "prototype"
)

// BeanDefinition represents a bean definition
type BeanDefinition struct {
	Name         string
	Scope        BeanScope
	Factory      interface{} // Function that creates the bean
	Instance     interface{} // Cached instance for singleton
	Initialized  bool
	mu           sync.RWMutex
}

// Container is the IoC container
type Container struct {
	beans        map[string]*BeanDefinition
	dependencies map[string][]string // Bean name -> list of dependency names
	creating     map[string]bool     // Track beans currently being created (for circular dependency detection)
	mu           sync.RWMutex
	initialized  bool
}

// NewContainer creates a new IoC container
func NewContainer() *Container {
	return &Container{
		beans:        make(map[string]*BeanDefinition),
		dependencies: make(map[string][]string),
		creating:     make(map[string]bool),
		initialized:  false,
	}
}

// RegisterBean registers a bean in the container
// name: bean name
// scope: "singleton" or "prototype"
// factory: function that creates the bean instance
func (c *Container) RegisterBean(name string, scope BeanScope, factory interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.beans[name]; exists {
		return fmt.Errorf("bean '%s' already registered", name)
	}

	// Validate factory is a function
	factoryType := reflect.TypeOf(factory)
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory for bean '%s' must be a function", name)
	}

	c.beans[name] = &BeanDefinition{
		Name:        name,
		Scope:       scope,
		Factory:     factory,
		Instance:    nil,
		Initialized: false,
	}

	return nil
}

// GetBean retrieves a bean from the container
func (c *Container) GetBean(name string) (interface{}, error) {
	c.mu.RLock()
	beanDef, exists := c.beans[name]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("bean '%s' not found", name)
	}

	// Handle singleton scope
	if beanDef.Scope == ScopeSingleton {
		beanDef.mu.RLock()
		if beanDef.Initialized && beanDef.Instance != nil {
			instance := beanDef.Instance
			beanDef.mu.RUnlock()
			return instance, nil
		}
		beanDef.mu.RUnlock()

		// Check for circular dependency
		c.mu.Lock()
		if c.creating[name] {
			c.mu.Unlock()
			return nil, fmt.Errorf("circular dependency detected: bean '%s' is being created", name)
		}
		c.creating[name] = true
		c.mu.Unlock()

		// Create instance for singleton
		beanDef.mu.Lock()

		// Double-check after acquiring write lock
		if beanDef.Initialized && beanDef.Instance != nil {
			beanDef.mu.Unlock()
			c.mu.Lock()
			delete(c.creating, name)
			c.mu.Unlock()
			return beanDef.Instance, nil
		}

		instance, err := c.createBeanInstance(beanDef)
		if err != nil {
			beanDef.mu.Unlock()
			c.mu.Lock()
			delete(c.creating, name)
			c.mu.Unlock()
			return nil, err
		}

		beanDef.Instance = instance
		beanDef.Initialized = true
		beanDef.mu.Unlock()

		c.mu.Lock()
		delete(c.creating, name)
		c.mu.Unlock()

		return instance, nil
	}

	// Handle prototype scope - always create new instance
	// Check for circular dependency
	c.mu.Lock()
	if c.creating[name] {
		c.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected: bean '%s' is being created", name)
	}
	c.creating[name] = true
	c.mu.Unlock()

	instance, err := c.createBeanInstance(beanDef)

	c.mu.Lock()
	delete(c.creating, name)
	c.mu.Unlock()

	return instance, err
}

// createBeanInstance creates a bean instance using the factory function
func (c *Container) createBeanInstance(beanDef *BeanDefinition) (interface{}, error) {
	factoryValue := reflect.ValueOf(beanDef.Factory)
	factoryType := factoryValue.Type()

	// Prepare arguments for factory function
	numIn := factoryType.NumIn()
	args := make([]reflect.Value, numIn)

	for i := 0; i < numIn; i++ {
		paramType := factoryType.In(i)
		
		// Try to resolve dependency from container
		dependency, err := c.resolveDependency(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency for bean '%s': %v", beanDef.Name, err)
		}
		
		args[i] = reflect.ValueOf(dependency)
	}

	// Call factory function
	results := factoryValue.Call(args)
	
	if len(results) == 0 {
		return nil, fmt.Errorf("factory function for bean '%s' must return at least one value", beanDef.Name)
	}

	// Return first result (the bean instance)
	return results[0].Interface(), nil
}

// resolveDependency resolves a dependency by type
func (c *Container) resolveDependency(paramType reflect.Type) (interface{}, error) {
	// Check if we're in a circular dependency situation
	c.mu.RLock()
	currentCreating := make([]string, 0, len(c.creating))
	for name := range c.creating {
		currentCreating = append(currentCreating, name)
	}
	c.mu.RUnlock()

	// Find bean name first (without holding lock during GetBean call)
	var beanName string
	
	c.mu.RLock()
	for name, beanDef := range c.beans {
		if beanDef.Factory != nil {
			factoryType := reflect.TypeOf(beanDef.Factory)
			if factoryType.NumOut() > 0 {
				returnType := factoryType.Out(0)
				if returnType.AssignableTo(paramType) {
					// Check if this would create a circular dependency
					for _, creatingName := range currentCreating {
						if creatingName == name {
							c.mu.RUnlock()
							return nil, fmt.Errorf("circular dependency detected: bean '%s' is being created", name)
						}
					}
					beanName = name
					break
				}
			}
		}
	}
	c.mu.RUnlock()

	if beanName != "" {
		return c.GetBean(beanName)
	}

	// Try to find by exact type match
	c.mu.RLock()
	for name, beanDef := range c.beans {
		if beanDef.Instance != nil {
			if reflect.TypeOf(beanDef.Instance).AssignableTo(paramType) {
				// Check if this would create a circular dependency
				for _, creatingName := range currentCreating {
					if creatingName == name {
						c.mu.RUnlock()
						return nil, fmt.Errorf("circular dependency detected: bean '%s' is being created", name)
					}
				}
				beanName = name
				break
			}
		}
	}
	c.mu.RUnlock()

	if beanName != "" {
		return c.GetBean(beanName)
	}

	return nil, fmt.Errorf("no bean found for type %s", paramType)
}

// checkCircularDependency checks for circular dependencies
func (c *Container) checkCircularDependency(beanName string, visited map[string]bool) error {
	if visited[beanName] {
		return fmt.Errorf("circular dependency detected involving bean '%s'", beanName)
	}

	visited[beanName] = true
	defer delete(visited, beanName)

	// Check dependencies
	if deps, exists := c.dependencies[beanName]; exists {
		for _, dep := range deps {
			if err := c.checkCircularDependency(dep, visited); err != nil {
				return err
			}
		}
	}

	return nil
}

// RegisterDependency registers a dependency relationship
func (c *Container) RegisterDependency(beanName string, dependencies []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[beanName] = dependencies
}

// ContainsBean checks if a bean is registered
func (c *Container) ContainsBean(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.beans[name]
	return exists
}

// GetBeanNames returns all registered bean names
func (c *Container) GetBeanNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	names := make([]string, 0, len(c.beans))
	for name := range c.beans {
		names = append(names, name)
	}
	return names
}

// Clear removes all beans from the container
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beans = make(map[string]*BeanDefinition)
	c.dependencies = make(map[string][]string)
	c.creating = make(map[string]bool)
	c.initialized = false
}
