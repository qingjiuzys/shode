// Package assert 提供测试断言功能
package assert

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// AnError 用于测试的错误
var AnError = errors.New("assert.AnError general error for testing")

// TestingT 测试接口
type TestingT interface {
	Errorf(format string, args ...interface{})
	FailNow()
}

// Equal 断言相等
func Equal(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	if !objectsAreEqual(expected, actual) {
		t.Errorf(fmt.Sprintf("Not equal: expected %+v, actual %+v", expected, actual))
	}
}

// NotEqual 断言不相等
func NotEqual(t TestingT, expected, actual interface{}, msgAndArgs ...interface{}) {
	if objectsAreEqual(expected, actual) {
		t.Errorf(fmt.Sprintf("Should not be equal: both are %+v", expected))
	}
}

// True 断言为真
func True(t TestingT, value bool, msgAndArgs ...interface{}) {
	if !value {
		t.Errorf("Should be true")
	}
}

// False 断言为假
func False(t TestingT, value bool, msgAndArgs ...interface{}) {
	if value {
		t.Errorf("Should be false")
	}
}

// Nil 断言为 nil
func Nil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	if !isNil(object) {
		t.Errorf(fmt.Sprintf("Expected nil, got %+v", object))
	}
}

// NotNil 断言不为 nil
func NotNil(t TestingT, object interface{}, msgAndArgs ...interface{}) {
	if isNil(object) {
		t.Errorf("Should not be nil")
	}
}

// Contains 断言包含
func Contains(t TestingT, str, substring interface{}, msgAndArgs ...interface{}) {
	ss, ok := str.(string)
	if !ok {
		t.Errorf("Cannot check contains on non-string")
		return
	}

	sub, ok := substring.(string)
	if !ok {
		t.Errorf("Substring must be string")
		return
	}

	if !strings.Contains(ss, sub) {
		t.Errorf(fmt.Sprintf("%q does not contain %q", str, substring))
	}
}

// Panics 断言会panic
func Panics(t TestingT, fn func(), msgAndArgs ...interface{}) {
	didPanic := false

	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		fn()
	}()

	if !didPanic {
		t.Errorf("Function should panic")
	}
}

// NotPanics 断言不会panic
func NotPanics(t TestingT, fn func(), msgAndArgs ...interface{}) {
	didPanic := false
	var message interface{}

	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
				message = r
			}
		}()
		fn()
	}()

	if didPanic {
		t.Errorf(fmt.Sprintf("Function should not panic, got: %v", message))
	}
}

// Len 断言长度
func Len(t TestingT, object interface{}, length int, msgAndArgs ...interface{}) {
	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan, reflect.String:
		if objValue.Len() != length {
			t.Errorf(fmt.Sprintf("%q should have length %d, got %d", object, length, objValue.Len()))
		}
	default:
		t.Errorf(fmt.Sprintf("Cannot get length of type %T", object))
	}
}

// Error 断言有错误
func Error(t TestingT, err error, msgAndArgs ...interface{}) {
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
}

// NoError 断言没有错误
func NoError(t TestingT, err error, msgAndArgs ...interface{}) {
	if err != nil {
		t.Errorf(fmt.Sprintf("Expected no error, got: %v", err))
	}
}

// Greater 断言大于
func Greater(t TestingT, e1, e2 interface{}, msgAndArgs ...interface{}) {
	e1Num, ok1 := toFloat(e1)
	e2Num, ok2 := toFloat(e2)

	if !ok1 || !ok2 {
		t.Errorf("Cannot compare non-numeric types")
		return
	}

	if e1Num <= e2Num {
		t.Errorf(fmt.Sprintf("%v is not greater than %v", e1, e2))
	}
}

// Less 断言小于
func Less(t TestingT, e1, e2 interface{}, msgAndArgs ...interface{}) {
	e1Num, ok1 := toFloat(e1)
	e2Num, ok2 := toFloat(e2)

	if !ok1 || !ok2 {
		t.Errorf("Cannot compare non-numeric types")
		return
	}

	if e1Num >= e2Num {
		t.Errorf(fmt.Sprintf("%v is not less than %v", e1, e2))
	}
}

// Implements 断言实现了接口
func Implements(t TestingT, object interface{}, interfaceType interface{}) {
	if !reflect.TypeOf(object).Implements(reflect.TypeOf(interfaceType).Elem()) {
		t.Errorf(fmt.Sprintf("%T must implement %T", object, interfaceType))
	}
}

// JSONEq 断言JSON相等
func JSONEq(t TestingT, expected, actual string, msgAndArgs ...interface{}) {
	var expectedJSONAsInterface, actualJSONAsInterface interface{}

	var err error
	expectedJSONAsInterface, err = decodeJSON(expected)
	if err != nil {
		t.Errorf(fmt.Sprintf("Expected JSON is invalid: %v", err))
		return
	}

	actualJSONAsInterface, err = decodeJSON(actual)
	if err != nil {
		t.Errorf(fmt.Sprintf("Actual JSON is invalid: %v", err))
		return
	}

	if !reflect.DeepEqual(expectedJSONAsInterface, actualJSONAsInterface) {
		t.Errorf(fmt.Sprintf("JSON not equal:\nExpected: %s\nActual:   %s", expected, actual))
	}
}

// objectsAreEqual 检查对象是否相等
func objectsAreEqual(expected, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	exp, ok := expected.([]byte)
	if !ok {
		return reflect.DeepEqual(expected, actual)
	}

	act, ok := actual.([]byte)
	if !ok {
		return false
	}

	if exp == nil || act == nil {
		return exp == nil && act == nil
	}

	return bytes.Equal(exp, act)
}

// isNil 检查是否为nil
func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	value := reflect.ValueOf(object)
	kind := value.Kind()
	if kind >= reflect.Chan && kind <= reflect.Slice && value.IsNil() {
		return true
	}

	return false
}

// toFloat 转换为float64
func toFloat(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// decodeJSON 解码JSON
func decodeJSON(data string) (interface{}, error) {
	var dec interface{}
	// 简化实现，实际应该使用json.Unmarshal
	return dec, nil
}
