package ioc

import (
	"fmt"
	"reflect"
)

// Injector handles dependency injection
type Injector struct {
	container *Container
}

// NewInjector creates a new injector
func NewInjector(container *Container) *Injector {
	return &Injector{container: container}
}

// Inject injects dependencies into a struct or function
func (inj *Injector) Inject(target interface{}) error {
	targetValue := reflect.ValueOf(target)
	targetType := targetValue.Type()

	// Handle pointer to struct
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
		targetValue = targetValue.Elem()
	}

	if targetType.Kind() != reflect.Struct {
		return fmt.Errorf("target must be a struct or pointer to struct")
	}

	// Inject dependencies into struct fields
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Check if field has inject tag
		tag := field.Tag.Get("inject")
		if tag == "" {
			continue
		}

		// Get bean name from tag or use field type
		beanName := tag
		if beanName == "" {
			beanName = field.Type.Name()
		}

		// Get bean from container
		bean, err := inj.container.GetBean(beanName)
		if err != nil {
			return fmt.Errorf("failed to inject bean '%s' into field '%s': %v", beanName, field.Name, err)
		}

		// Set field value
		if !fieldValue.CanSet() {
			return fmt.Errorf("field '%s' cannot be set", field.Name)
		}

		beanValue := reflect.ValueOf(bean)
		if !beanValue.Type().AssignableTo(field.Type) {
			return fmt.Errorf("bean type %s is not assignable to field type %s", beanValue.Type(), field.Type)
		}

		fieldValue.Set(beanValue)
	}

	return nil
}

// InjectFunction injects dependencies into a function and calls it
func (inj *Injector) InjectFunction(fn interface{}) ([]interface{}, error) {
	fnValue := reflect.ValueOf(fn)
	fnType := fnValue.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("target must be a function")
	}

	numIn := fnType.NumIn()
	args := make([]reflect.Value, numIn)

	// Resolve dependencies for each parameter
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		dependency, err := inj.container.resolveDependency(paramType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve dependency for parameter %d: %v", i, err)
		}
		args[i] = reflect.ValueOf(dependency)
	}

	// Call function
	results := fnValue.Call(args)

	// Convert results to interface slice
	resultInterfaces := make([]interface{}, len(results))
	for i, result := range results {
		resultInterfaces[i] = result.Interface()
	}

	return resultInterfaces, nil
}
