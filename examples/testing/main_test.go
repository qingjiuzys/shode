// Package main 测试示例
package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/testing/assert"
	httptest "gitee.com/com_818cloud/shode/pkg/testing/http"
	"gitee.com/com_818cloud/shode/pkg/testing/mock"
)

// TestAssert 断言测试示例
func TestAssert(t *testing.T) {
	// Equal
	assert.Equal(t, 1, 1)
	assert.Equal(t, "hello", "hello")

	// NotEqual
	assert.NotEqual(t, 1, 2)

	// True/False
	assert.True(t, true)
	assert.False(t, false)

	// Nil/NotNil
	assert.Nil(t, nil)
	var ptr *int
	assert.Nil(t, ptr)

	value := 42
	assert.NotNil(t, &value)

	// Contains
	assert.Contains(t, "hello world", "hello")

	// Len
	slice := []int{1, 2, 3}
	assert.Len(t, slice, 3)

	// Greater/Less
	assert.Greater(t, 5, 3)
	assert.Less(t, 3, 5)

	// Error/NoError
	err := nil
	assert.NoError(t, err)

	err = assert.AnError
	assert.Error(t, err)
}

// TestHTTP HTTP测试示例
func TestHTTP(t *testing.T) {
	// 创建测试handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message":"ok"}`))
		case "/created":
			w.WriteHeader(http.StatusCreated)
		case "/notfound":
			w.WriteHeader(http.StatusNotFound)
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"name": "test"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	h := httptest.NewHelper(t, handler)

	// 测试GET请求
	h.GET("/ok").AssertOK().
		AssertContains("ok").
		AssertJSON()

	// 测试POST请求
	h.POST("/created", map[string]string{"name": "test"}).
		AssertCreated()

	// 测试404
	h.GET("/notfound").AssertNotFound()

	// 测试JSON响应
	h.GET("/json").AssertJSON().
		AssertJSONEq(`{"name":"test"}`)
}

// TestMock Mock测试示例
func TestMock(t *testing.T) {
	m := mock.New()

	// 设置期望
	m.On("DoSomething", 1, 2).Return(3)
	m.On("DoSomething", 4, 5).Return(9)

	// 记录调用
	m.Recorded("DoSomething", 1, 2)
	m.Recorded("DoSomething", 4, 5)

	// 断言调用
	assert.True(t, m.Called("DoSomething"))
	assert.True(t, m.CalledWith("DoSomething", 1, 2))
	assert.Equal(t, 2, m.CalledTimes("DoSomething"))

	// 断言期望
	err := m.AssertExpectations()
	assert.NoError(t, err)
}

// TestMockWithAnyArgs Mock使用任意参数测试
func TestMockWithAnyArgs(t *testing.T) {
	m := mock.New()

	// 使用Anything匹配任何参数
	m.On("Process", mock.Any(), mock.Any()).Return("processed")

	// 使用AnythingOfType匹配任何指定类型
	m.On("Handle", mock.AnyOfType("string"), mock.AnyOfType("int")).Return(true)

	// 这些调用都会匹配
	m.Recorded("Process", 1, 2)
	m.Recorded("Process", "hello", "world")
	m.Recorded("Handle", "test", 42)

	assert.Equal(t, 3, m.CalledTimes("Process"))
	assert.Equal(t, 1, m.CalledTimes("Handle"))

	err := m.AssertExpectations()
	assert.NoError(t, err)
}

// TestMockMultipleCalls Mock多次调用测试
func TestMockMultipleCalls(t *testing.T) {
	m := mock.New()

	// 期望调用3次
	m.On("Method").Times(3)

	// 记录调用
	for i := 0; i < 3; i++ {
		m.Recorded("Method")
	}

	assert.Equal(t, 3, m.CalledTimes("Method"))

	err := m.AssertExpectations()
	assert.NoError(t, err)
}

// TestMockAtLeast Mock至少调用测试
func TestMockAtLeast(t *testing.T) {
	m := mock.New()

	// 期望至少调用2次
	m.On("Method").AtLeast(2)

	// 记录调用3次
	for i := 0; i < 3; i++ {
		m.Recorded("Method")
	}

	assert.Equal(t, 3, m.CalledTimes("Method"))

	err := m.AssertExpectations()
	assert.NoError(t, err)
}

// ExampleService 示例服务
type ExampleService struct {
	mock *mock.Mock
}

// NewExampleService 创建服务
func NewExampleService() *ExampleService {
	return &ExampleService{
		mock: mock.New(),
	}
}

// GetData 获取数据（示例方法）
func (s *ExampleService) GetData(id int) (string, error) {
	// 记录方法调用
	s.mock.Recorded("GetData", id)

	// 返回mock数据
	returns := s.mock.GetReturns("GetData", id)
	if returns != nil && len(returns) > 0 {
		if data, ok := returns[0].(string); ok {
			return data, nil
		}
	}

	return "default", nil
}

// TestExampleService 服务测试示例
func TestExampleService(t *testing.T) {
	service := NewExampleService()

	// 设置期望
	service.mock.On("GetData", 1).Return("data1")
	service.mock.On("GetData", 2).Return("data2")

	// 测试
	data, err := service.GetData(1)
	assert.NoError(t, err)
	assert.Equal(t, "data1", data)

	data, err = service.GetData(2)
	assert.NoError(t, err)
	assert.Equal(t, "data2", data)

	// 断言期望
	err = service.mock.AssertExpectations()
	assert.NoError(t, err)
}

// BenchmarkExample 基准测试示例
func BenchmarkExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// 被测试的代码
		result := 0
		for j := 0; j < 100; j++ {
			result += j
		}
		_ = result
	}
}
