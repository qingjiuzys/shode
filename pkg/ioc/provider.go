package ioc

import (
	"fmt"
	"reflect"
)

// Provider provides beans to the container
type Provider interface {
	Provide() (interface{}, error)
}

// FactoryProvider is a provider that uses a factory function
type FactoryProvider struct {
	factory interface{}
	container *Container
}

// NewFactoryProvider creates a new factory provider
func NewFactoryProvider(factory interface{}, container *Container) *FactoryProvider {
	return &FactoryProvider{
		factory:   factory,
		container: container,
	}
}

// Provide creates a bean instance using the factory function
func (fp *FactoryProvider) Provide() (interface{}, error) {
	factoryValue := reflect.ValueOf(fp.factory)
	factoryType := factoryValue.Type()

	if factoryType.Kind() != reflect.Func {
		return nil, fmt.Errorf("factory must be a function")
	}

	numIn := factoryType.NumIn()
	args := make([]reflect.Value, numIn)

	// Resolve dependencies
	for i := 0; i < numIn; i++ {
		paramType := factoryType.In(i)
		dependency, err := fp.container.resolveDependency(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency: %v", err)
		}
		args[i] = reflect.ValueOf(dependency)
	}

	// Call factory
	results := factoryValue.Call(args)
	if len(results) == 0 {
		return nil, fmt.Errorf("factory must return at least one value")
	}

	return results[0].Interface(), nil
}

// ValueProvider is a provider that provides a constant value
type ValueProvider struct {
	value interface{}
}

// NewValueProvider creates a new value provider
func NewValueProvider(value interface{}) *ValueProvider {
	return &ValueProvider{value: value}
}

// Provide returns the constant value
func (vp *ValueProvider) Provide() (interface{}, error) {
	return vp.value, nil
}
