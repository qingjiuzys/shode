package aop

import (
	"fmt"
	"reflect"
)

// Proxy creates a proxy for a target object
type Proxy struct {
	target   interface{}
	aspects  []Aspect
}

// Aspect represents an aspect (before/after/around advice)
type Aspect struct {
	Before func(args []interface{}) error
	After  func(args []interface{}, result []interface{}) error
	Around func(proceed func() ([]interface{}, error)) ([]interface{}, error)
}

// NewProxy creates a new proxy
func NewProxy(target interface{}) *Proxy {
	return &Proxy{
		target:  target,
		aspects: make([]Aspect, 0),
	}
}

// AddAspect adds an aspect to the proxy
func (p *Proxy) AddAspect(aspect Aspect) {
	p.aspects = append(p.aspects, aspect)
}

// Invoke invokes a method on the target with aspect weaving
func (p *Proxy) Invoke(methodName string, args []interface{}) ([]interface{}, error) {
	targetValue := reflect.ValueOf(p.target)
	method := targetValue.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	// Execute before aspects
	for _, aspect := range p.aspects {
		if aspect.Before != nil {
			if err := aspect.Before(args); err != nil {
				return nil, err
			}
		}
	}

	// Prepare method arguments
	methodType := method.Type()
	reflectArgs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if i < methodType.NumIn() {
			argType := methodType.In(i)
			reflectArgs[i] = reflect.ValueOf(arg).Convert(argType)
		}
	}

	// Execute around aspects or direct call
	var results []reflect.Value
	var err error

	if len(p.aspects) > 0 && p.aspects[0].Around != nil {
		// Use around aspect
		proceed := func() ([]interface{}, error) {
			results := method.Call(reflectArgs)
			resultInterfaces := make([]interface{}, len(results))
			for i, r := range results {
				resultInterfaces[i] = r.Interface()
			}
			return resultInterfaces, nil
		}
		resultInterfaces, err := p.aspects[0].Around(proceed)
		if err != nil {
			return nil, err
		}
		// Convert back to reflect.Value
		results = make([]reflect.Value, len(resultInterfaces))
		for i, r := range resultInterfaces {
			results[i] = reflect.ValueOf(r)
		}
	} else {
		// Direct call
		results = method.Call(reflectArgs)
	}

	// Convert results to interfaces
	resultInterfaces := make([]interface{}, len(results))
	for i, r := range results {
		resultInterfaces[i] = r.Interface()
	}

	// Execute after aspects
	for _, aspect := range p.aspects {
		if aspect.After != nil {
			if err := aspect.After(args, resultInterfaces); err != nil {
				return nil, err
			}
		}
	}

	return resultInterfaces, err
}
