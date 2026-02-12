// Package mock 提供Mock功能
package mock

import (
	"fmt"
	"reflect"
	"sync"
)

// Mock 对象
type Mock struct {
	mu       sync.Mutex
	calls    []Call
	expected []ExpectedCall
	//	arguments map[string][]Argument
}

// Call 方法调用
type Call struct {
	Method string
	Args   []interface{}
}

// ExpectedCall 期望的调用
type ExpectedCall struct {
	Method   string
	Args     []interface{}
	Returns  []interface{}
	Calls    int
	MinCalls int
	MaxCalls int
}

// New 创建Mock对象
func New() *Mock {
	return &Mock{
		calls:    make([]Call, 0),
		expected: make([]ExpectedCall, 0),
	}
}

// On 期望方法被调用
func (m *Mock) On(method string, args ...interface{}) *ExpectedCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := ExpectedCall{
		Method:   method,
		Args:     args,
		MinCalls: 1,
		MaxCalls: 1,
	}

	m.expected = append(m.expected, call)
	return &m.expected[len(m.expected)-1]
}

// Recorded 记录方法调用
func (m *Mock) Recorded(method string, args ...interface{}) *Call {
	m.mu.Lock()
	defer m.mu.Unlock()

	call := Call{
		Method: method,
		Args:   args,
	}

	m.calls = append(m.calls, call)
	return &m.calls[len(m.calls)-1]
}

// Called 检查方法是否被调用
func (m *Mock) Called(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CalledTimes 获取方法调用次数
func (m *Mock) CalledTimes(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, call := range m.calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// CalledWith 检查方法是否用指定参数被调用
func (m *Mock) CalledWith(method string, args ...interface{}) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, call := range m.calls {
		if call.Method == method && argsMatch(call.Args, args) {
			return true
		}
	}
	return false
}

// AssertExpectations 断言所有期望都满足
func (m *Mock) AssertExpectations() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, expected := range m.expected {
		if expected.Calls < expected.MinCalls {
			return fmt.Errorf("expected method %s to be called at least %d times, got %d",
				expected.Method, expected.MinCalls, expected.Calls)
		}

		if expected.MaxCalls > 0 && expected.Calls > expected.MaxCalls {
			return fmt.Errorf("expected method %s to be called at most %d times, got %d",
				expected.Method, expected.MaxCalls, expected.Calls)
		}
	}

	return nil
}

// Reset 重置Mock
func (m *Mock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = make([]Call, 0)
	m.expected = make([]ExpectedCall, 0)
}

// FindExpected 查找期望的调用
func (m *Mock) FindExpected(method string, args []interface{}) *ExpectedCall {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.expected {
		expected := &m.expected[i]
		if expected.Method == method && argsMatch(expected.Args, args) {
			expected.Calls++
			return expected
		}
	}

	return nil
}

// argsMatch 检查参数是否匹配
func argsMatch(expected, actual []interface{}) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if !argMatch(expected[i], actual[i]) {
			return false
		}
	}

	return true
}

// argMatch 检查单个参数是否匹配
func argMatch(expected, actual interface{}) bool {
	// Anything 匹配任何值
	if _, ok := expected.(AnythingOfType); ok {
		return true
	}

	if _, ok := expected.(Anything); ok {
		return true
	}

	// 使用反射比较
	return reflect.DeepEqual(expected, actual)
}

// Return 设置返回值
func (e *ExpectedCall) Return(returns ...interface{}) *ExpectedCall {
	e.Returns = returns
	return e
}

// Once 期望调用一次
func (e *ExpectedCall) Once() *ExpectedCall {
	e.MinCalls = 1
	e.MaxCalls = 1
	return e
}

// Twice 期望调用两次
func (e *ExpectedCall) Twice() *ExpectedCall {
	e.MinCalls = 2
	e.MaxCalls = 2
	return e
}

// Times 期望调用指定次数
func (e *ExpectedCall) Times(n int) *ExpectedCall {
	e.MinCalls = n
	e.MaxCalls = n
	return e
}

// AtLeast 期望至少调用n次
func (e *ExpectedCall) AtLeast(n int) *ExpectedCall {
	e.MinCalls = n
	e.MaxCalls = 0
	return e
}

// AtMost 期望最多调用n次
func (e *ExpectedCall) AtMost(n int) *ExpectedCall {
	e.MinCalls = 0
	e.MaxCalls = n
	return e
}

// Maybe 期望调用0次或1次
func (e *ExpectedCall) Maybe() *ExpectedCall {
	e.MinCalls = 0
	e.MaxCalls = 1
	return e
}

// Anything 匹配任何值
type Anything struct{}

// AnythingOfType 匹配任何指定类型的值
type AnythingOfType string

// Arg 匹配参数
func Arg() Anything {
	return Anything{}
}

// Any 任何值的别名
func Any() Anything {
	return Anything{}
}

// AnyOfType 任何指定类型的值
func AnyOfType(typ string) AnythingOfType {
	return AnythingOfType(typ)
}

// GetReturns 获取返回值
func (m *Mock) GetReturns(method string, args ...interface{}) []interface{} {
	expected := m.FindExpected(method, args)
	if expected == nil || expected.Returns == nil {
		return nil
	}

	return expected.Returns
}

// MethodCaller 方法调用器
type MethodCaller struct {
	mock   *Mock
	method string
}

// Method 创建方法调用器
func (m *Mock) Method(method string) *MethodCaller {
	return &MethodCaller{
		mock:   m,
		method: method,
	}
}

// With 指定参数
func (c *MethodCaller) With(args ...interface{}) *MethodCaller {
	return c
}

// Return 返回值
func (c *MethodCaller) Return(returns ...interface{}) *MethodCaller {
	c.mock.On(c.method).Return(returns...)
	return c
}

// Get 获取返回值
func (c *MethodCaller) Get(args ...interface{}) []interface{} {
	return c.mock.GetReturns(c.method, args...)
}
