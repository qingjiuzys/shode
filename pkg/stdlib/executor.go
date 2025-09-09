package stdlib

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ExecuteFunction executes a standard library function by name with arguments
func (sl *StdLib) ExecuteFunction(name string, args ...interface{}) (interface{}, error) {
	// Get the function implementation
	fn, exists := FunctionMap[name]
	if !exists {
		return nil, fmt.Errorf("function not found: %s", name)
	}

	// Convert function to reflect value
	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function: %s", name)
	}

	// Prepare arguments
	fnType := fnValue.Type()
	if fnType.NumIn() != len(args)+1 { // +1 for the receiver
		return nil, fmt.Errorf("invalid number of arguments for %s: expected %d, got %d", 
			name, fnType.NumIn()-1, len(args))
	}

	// Convert arguments to expected types
	callArgs := make([]reflect.Value, len(args)+1)
	callArgs[0] = reflect.ValueOf(sl) // receiver

	for i, arg := range args {
		expectedType := fnType.In(i + 1)
		argValue, err := convertArgument(arg, expectedType)
		if err != nil {
			return nil, fmt.Errorf("argument %d for %s: %v", i+1, name, err)
		}
		callArgs[i+1] = argValue
	}

	// Call the function
	results := fnValue.Call(callArgs)

	// Handle results
	if len(results) == 0 {
		return nil, nil
	}

	// Check for error
	if len(results) > 1 {
		lastResult := results[len(results)-1]
		if lastResult.Type().Implements(errorInterface) {
			if !lastResult.IsNil() {
				return nil, lastResult.Interface().(error)
			}
			// Return other results if no error
			if len(results) == 2 {
				return results[0].Interface(), nil
			}
		}
	}

	return results[0].Interface(), nil
}

// convertArgument converts an argument to the expected type
func convertArgument(arg interface{}, expectedType reflect.Type) (reflect.Value, error) {
	argValue := reflect.ValueOf(arg)
	argType := argValue.Type()

	// If types match exactly, return as is
	if argType == expectedType {
		return argValue, nil
	}

	// Handle string conversion
	if expectedType.Kind() == reflect.String {
		if str, ok := arg.(string); ok {
			return reflect.ValueOf(str), nil
		}
		// Convert any type to string
		return reflect.ValueOf(fmt.Sprintf("%v", arg)), nil
	}

	// Handle int conversion
	if expectedType.Kind() == reflect.Int {
		switch v := arg.(type) {
		case int:
			return reflect.ValueOf(v), nil
		case int64:
			return reflect.ValueOf(int(v)), nil
		case float64:
			return reflect.ValueOf(int(v)), nil
		case string:
			var i int
			_, err := fmt.Sscanf(v, "%d", &i)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert %q to int", v)
			}
			return reflect.ValueOf(i), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to int", arg)
		}
	}

	// Handle int64 conversion
	if expectedType.Kind() == reflect.Int64 {
		switch v := arg.(type) {
		case int64:
			return reflect.ValueOf(v), nil
		case int:
			return reflect.ValueOf(int64(v)), nil
		case float64:
			return reflect.ValueOf(int64(v)), nil
		case string:
			var i int64
			_, err := fmt.Sscanf(v, "%d", &i)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert %q to int64", v)
			}
			return reflect.ValueOf(i), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to int64", arg)
		}
	}

	// Handle bool conversion
	if expectedType.Kind() == reflect.Bool {
		switch v := arg.(type) {
		case bool:
			return reflect.ValueOf(v), nil
		case string:
			lower := strings.ToLower(v)
			if lower == "true" || lower == "1" || lower == "yes" {
				return reflect.ValueOf(true), nil
			}
			if lower == "false" || lower == "0" || lower == "no" {
				return reflect.ValueOf(false), nil
			}
			return reflect.Value{}, fmt.Errorf("cannot convert %q to bool", v)
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to bool", arg)
		}
	}

	// Handle time.Duration conversion
	if expectedType == durationType {
		switch v := arg.(type) {
		case time.Duration:
			return reflect.ValueOf(v), nil
		case string:
			duration, err := time.ParseDuration(v)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("cannot convert %q to duration: %v", v, err)
			}
			return reflect.ValueOf(duration), nil
		case int:
			return reflect.ValueOf(time.Duration(v) * time.Second), nil
		case int64:
			return reflect.ValueOf(time.Duration(v) * time.Second), nil
		default:
			return reflect.Value{}, fmt.Errorf("cannot convert %T to duration", arg)
		}
	}

	// Handle slice conversion
	if expectedType.Kind() == reflect.Slice {
		if argType.Kind() == reflect.Slice {
			// Check if element types are compatible
			if argType.Elem().ConvertibleTo(expectedType.Elem()) {
				return argValue.Convert(expectedType), nil
			}
		}
		// Handle string to []string conversion
		if expectedType.Elem().Kind() == reflect.String && argType.Kind() == reflect.String {
			// Split string by spaces for simple cases
			parts := strings.Fields(arg.(string))
			return reflect.ValueOf(parts), nil
		}
	}

	return reflect.Value{}, fmt.Errorf("cannot convert %T to %v", arg, expectedType)
}

// Helper variables for type comparison
var (
	errorInterface = reflect.TypeOf((*error)(nil)).Elem()
	durationType   = reflect.TypeOf(time.Duration(0))
)

// ExecuteFunctionSafe safely executes a function with error handling
func (sl *StdLib) ExecuteFunctionSafe(name string, args ...interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in function %s: %v", name, r)
		}
	}()

	return sl.ExecuteFunction(name, args...)
}

// FunctionSignature returns the signature of a function
func (sl *StdLib) FunctionSignature(name string) (string, error) {
	fn, exists := FunctionMap[name]
	if !exists {
		return "", fmt.Errorf("function not found: %s", name)
	}

	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return "", fmt.Errorf("not a function: %s", name)
	}

	var signature strings.Builder
	signature.WriteString(name)
	signature.WriteString("(")

	// Skip the first parameter (receiver)
	for i := 1; i < fnType.NumIn(); i++ {
		if i > 1 {
			signature.WriteString(", ")
		}
		paramType := fnType.In(i)
		signature.WriteString(paramType.String())
	}

	signature.WriteString(")")

	// Add return types
	if fnType.NumOut() > 0 {
		signature.WriteString(" ")
		if fnType.NumOut() > 1 {
			signature.WriteString("(")
		}
		for i := 0; i < fnType.NumOut(); i++ {
			if i > 0 {
				signature.WriteString(", ")
			}
			returnType := fnType.Out(i)
			signature.WriteString(returnType.String())
		}
		if fnType.NumOut() > 1 {
			signature.WriteString(")")
		}
	}

	return signature.String(), nil
}
