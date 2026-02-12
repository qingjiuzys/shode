// Package http 提供HTTP测试辅助功能
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// HTTPHelper HTTP测试辅助器
type HTTPHelper struct {
	T      *testing.T
	Router http.Handler
}

// NewHelper 创建HTTP辅助器
func NewHelper(t *testing.T, handler http.Handler) *HTTPHelper {
	return &HTTPHelper{
		T:      t,
		Router: handler,
	}
}

// Request 发送HTTP请求
func (h *HTTPHelper) Request(method, url string, body interface{}, headers map[string]string) *ResponseRecorder {
	var reqBody io.Reader

	if body != nil {
		switch v := body.(type) {
		case string:
			reqBody = strings.NewReader(v)
		case []byte:
			reqBody = bytes.NewReader(v)
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				h.T.Fatalf("Failed to marshal request body: %v", err)
			}
			reqBody = bytes.NewReader(jsonData)
		}
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		h.T.Fatalf("Failed to create request: %v", err)
	}

	// 设置headers
	if headers != nil {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	// 默认设置Content-Type
	if req.Header.Get("Content-Type") == "" && body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// 记录响应
	w := httptest.NewRecorder()
	h.Router.ServeHTTP(w, req)

	return &ResponseRecorder{
		ResponseRecorder: w,
		T:                h.T,
	}
}

// GET 发送GET请求
func (h *HTTPHelper) GET(url string, headers ...map[string]string) *ResponseRecorder {
	headerMap := make(map[string]string)
	if len(headers) > 0 {
		headerMap = headers[0]
	}
	return h.Request("GET", url, nil, headerMap)
}

// POST 发送POST请求
func (h *HTTPHelper) POST(url string, body interface{}, headers ...map[string]string) *ResponseRecorder {
	headerMap := make(map[string]string)
	if len(headers) > 0 {
		headerMap = headers[0]
	}
	return h.Request("POST", url, body, headerMap)
}

// PUT 发送PUT请求
func (h *HTTPHelper) PUT(url string, body interface{}, headers ...map[string]string) *ResponseRecorder {
	headerMap := make(map[string]string)
	if len(headers) > 0 {
		headerMap = headers[0]
	}
	return h.Request("PUT", url, body, headerMap)
}

// DELETE 发送DELETE请求
func (h *HTTPHelper) DELETE(url string, headers ...map[string]string) *ResponseRecorder {
	headerMap := make(map[string]string)
	if len(headers) > 0 {
		headerMap = headers[0]
	}
	return h.Request("DELETE", url, nil, headerMap)
}

// PATCH 发送PATCH请求
func (h *HTTPHelper) PATCH(url string, body interface{}, headers ...map[string]string) *ResponseRecorder {
	headerMap := make(map[string]string)
	if len(headers) > 0 {
		headerMap = headers[0]
	}
	return h.Request("PATCH", url, body, headerMap)
}

// ResponseRecorder 响应记录器
type ResponseRecorder struct {
	*httptest.ResponseRecorder
	T *testing.T
}

// StatusCode 获取状态码
func (r *ResponseRecorder) StatusCode() int {
	return r.ResponseRecorder.Code
}

// Body 获取响应体
func (r *ResponseRecorder) Body() string {
	return r.ResponseRecorder.Body.String()
}

// BodyBytes 获取响应体字节
func (r *ResponseRecorder) BodyBytes() []byte {
	return r.ResponseRecorder.Body.Bytes()
}

// JSON 解析响应体为JSON
func (r *ResponseRecorder) JSON(v interface{}) error {
	return json.Unmarshal(r.BodyBytes(), v)
}

// AssertStatus 断言状态码
func (r *ResponseRecorder) AssertStatus(expectedCode int) *ResponseRecorder {
	if r.ResponseRecorder.Code != expectedCode {
		r.T.Errorf("Expected status code %d, got %d", expectedCode, r.ResponseRecorder.Code)
	}
	return r
}

// AssertOK 断言200
func (r *ResponseRecorder) AssertOK() *ResponseRecorder {
	return r.AssertStatus(http.StatusOK)
}

// AssertCreated 断言201
func (r *ResponseRecorder) AssertCreated() *ResponseRecorder {
	return r.AssertStatus(http.StatusCreated)
}

// AssertNoContent 断言204
func (r *ResponseRecorder) AssertNoContent() *ResponseRecorder {
	return r.AssertStatus(http.StatusNoContent)
}

// AssertBadRequest 断言400
func (r *ResponseRecorder) AssertBadRequest() *ResponseRecorder {
	return r.AssertStatus(http.StatusBadRequest)
}

// AssertUnauthorized 断言401
func (r *ResponseRecorder) AssertUnauthorized() *ResponseRecorder {
	return r.AssertStatus(http.StatusUnauthorized)
}

// AssertForbidden 断言403
func (r *ResponseRecorder) AssertForbidden() *ResponseRecorder {
	return r.AssertStatus(http.StatusForbidden)
}

// AssertNotFound 断言404
func (r *ResponseRecorder) AssertNotFound() *ResponseRecorder {
	return r.AssertStatus(http.StatusNotFound)
}

// AssertInternalServerError 断言500
func (r *ResponseRecorder) AssertInternalServerError() *ResponseRecorder {
	return r.AssertStatus(http.StatusInternalServerError)
}

// AssertContentType 断言Content-Type
func (r *ResponseRecorder) AssertContentType(contentType string) *ResponseRecorder {
	actualCT := r.ResponseRecorder.Header().Get("Content-Type")
	if !strings.Contains(actualCT, contentType) {
		r.T.Errorf("Expected Content-Type %s, got %s", contentType, actualCT)
	}
	return r
}

// AssertJSON 断言JSON响应
func (r *ResponseRecorder) AssertJSON() *ResponseRecorder {
	return r.AssertContentType("application/json")
}

// AssertBody 断言响应体
func (r *ResponseRecorder) AssertBody(expected string) *ResponseRecorder {
	actual := r.Body()
	if actual != expected {
		r.T.Errorf("Expected body %q, got %q", expected, actual)
	}
	return r
}

// AssertContains 断言响应体包含
func (r *ResponseRecorder) AssertContains(substring string) *ResponseRecorder {
	if !strings.Contains(r.Body(), substring) {
		r.T.Errorf("Expected body to contain %q, got %q", substring, r.Body())
	}
	return r
}

// AssertJSONEq 断言JSON相等
func (r *ResponseRecorder) AssertJSONEq(expectedJSON string) *ResponseRecorder {
	var expected, actual interface{}

	if err := json.Unmarshal([]byte(expectedJSON), &expected); err != nil {
		r.T.Fatalf("Failed to parse expected JSON: %v", err)
	}

	if err := json.Unmarshal(r.BodyBytes(), &actual); err != nil {
		r.T.Fatalf("Failed to parse actual JSON: %v", err)
	}

	if !jsonEqual(expected, actual) {
		r.T.Errorf("JSON not equal.\nExpected: %s\nActual:   %s", expectedJSON, r.Body())
	}

	return r
}

// GetHeader 获取header
func (r *ResponseRecorder) GetHeader(key string) string {
	return r.ResponseRecorder.Header().Get(key)
}

// AssertHeader 断言header
func (r *ResponseRecorder) AssertHeader(key, expectedValue string) *ResponseRecorder {
	actualValue := r.GetHeader(key)
	if actualValue != expectedValue {
		r.T.Errorf("Expected header %s=%q, got %q", key, expectedValue, actualValue)
	}
	return r
}

// GetCookie 获取cookie
func (r *ResponseRecorder) GetCookie(name string) *http.Cookie {
	for _, c := range r.ResponseRecorder.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// AssertCookie 断言cookie
func (r *ResponseRecorder) AssertCookie(name string) *http.Cookie {
	cookie := r.GetCookie(name)
	if cookie == nil {
		r.T.Errorf("Cookie %q not found", name)
	}
	return cookie
}

// FormEncoder 表单编码器
type FormEncoder struct {
	values url.Values
}

// NewFormEncoder 创建表单编码器
func NewFormEncoder() *FormEncoder {
	return &FormEncoder{
		values: make(url.Values),
	}
}

// Add 添加字段
func (e *FormEncoder) Add(key, value string) {
	e.values.Add(key, value)
}

// Encode 编码
func (e *FormEncoder) Encode() string {
	return e.values.Encode()
}

// Bytes 返回字节
func (e *FormEncoder) Bytes() []byte {
	return []byte(e.Encode())
}

// String 返回字符串
func (e *FormEncoder) String() string {
	return e.Encode()
}

// jsonEqual 比较JSON
func jsonEqual(a, b interface{}) bool {
	// 简化实现，实际应该使用reflect.DeepEqual
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
