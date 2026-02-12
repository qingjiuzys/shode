// Package httpclient 提供HTTP客户端工具
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client HTTP客户端
type Client struct {
	client      *http.Client
	baseURL     string
	headers     map[string]string
	maxRetries  int
	timeout     time.Duration
}

// Config 客户端配置
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	Headers    map[string]string
}

// DefaultConfig 默认配置
var DefaultConfig = Config{
	Timeout:    30 * time.Second,
	MaxRetries: 3,
}

// NewClient 创建HTTP客户端
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = DefaultConfig.Timeout
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = DefaultConfig.MaxRetries
	}

	return &Client{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		baseURL:    config.BaseURL,
		headers:    config.Headers,
		maxRetries: config.MaxRetries,
		timeout:    config.Timeout,
	}
}

// SetBaseURL 设置基础URL
func (c *Client) SetBaseURL(baseURL string) *Client {
	c.baseURL = baseURL
	return c
}

// SetTimeout 设置超时
func (c *Client) SetTimeout(timeout time.Duration) *Client {
	c.timeout = timeout
	c.client.Timeout = timeout
	return c
}

// SetHeader 设置请求头
func (c *Client) SetHeader(key, value string) *Client {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
	return c
}

// SetHeaders 设置多个请求头
func (c *Client) SetHeaders(headers map[string]string) *Client {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// Get GET请求
func (c *Client) Get(path string) (*Response, error) {
	return c.Do(http.MethodGet, path, nil)
}

// Post POST请求
func (c *Client) Post(path string, body any) (*Response, error) {
	return c.Do(http.MethodPost, path, body)
}

// Put PUT请求
func (c *Client) Put(path string, body any) (*Response, error) {
	return c.Do(http.MethodPut, path, body)
}

// Patch PATCH请求
func (c *Client) Patch(path string, body any) (*Response, error) {
	return c.Do(http.MethodPatch, path, body)
}

// Delete DELETE请求
func (c *Client) Delete(path string) (*Response, error) {
	return c.Do(http.MethodDelete, path, nil)
}

// Do 执行HTTP请求
func (c *Client) Do(method, path string, body any) (*Response, error) {
	// 构建完整URL
	fullURL := c.baseURL + path

	// 序列化请求体
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}

	// 创建请求
	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 自动设置Content-Type
	if body != nil && req.Header.Get("Content-Type") == "" {
		switch body.(type) {
		case string, []byte, io.Reader:
			// 不自动设置
		default:
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// 执行请求（带重试）
	var resp *http.Response
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// 等待一段时间后重试
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		resp, err = c.client.Do(req)
		if err == nil {
			break
		}
		lastErr = err
	}

	if err != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", c.maxRetries+1, lastErr)
	}

	return NewResponse(resp), nil
}

// GetWithContext 带context的GET请求
func (c *Client) GetWithContext(ctx context.Context, path string) (*Response, error) {
	return c.DoWithContext(ctx, http.MethodGet, path, nil)
}

// DoWithContext 带context的HTTP请求
func (c *Client) DoWithContext(ctx context.Context, method, path string, body any) (*Response, error) {
	// 构建完整URL
	fullURL := c.baseURL + path

	// 序列化请求体
	var bodyReader io.Reader
	if body != nil {
		switch v := body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			jsonData, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	// 自动设置Content-Type
	if body != nil && req.Header.Get("Content-Type") == "" {
		switch body.(type) {
		case string, []byte, io.Reader:
			// 不自动设置
		default:
			req.Header.Set("Content-Type", "application/json")
		}
	}

	// 执行请求
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return NewResponse(resp), nil
}

// Response HTTP响应
type Response struct {
	*http.Response
}

// NewResponse 创建响应
func NewResponse(resp *http.Response) *Response {
	return &Response{Response: resp}
}

// Bytes 获取响应体字节数组
func (r *Response) Bytes() ([]byte, error) {
	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// String 获取响应体字符串
func (r *Response) String() (string, error) {
	data, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// JSON 解析JSON响应
func (r *Response) JSON(v any) error {
	data, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// IsSuccess 检查是否成功
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError 检查是否错误
func (r *Response) IsError() bool {
	return !r.IsSuccess()
}

// RequestBuilder 请求构建器
type RequestBuilder struct {
	method  string
	url     string
	headers map[string]string
	query   map[string]string
	body    any
	client  *Client
}

// NewRequestBuilder 创建请求构建器
func NewRequestBuilder(client *Client) *RequestBuilder {
	return &RequestBuilder{
		headers: make(map[string]string),
		query:   make(map[string]string),
		client:  client,
	}
}

// SetMethod 设置方法
func (rb *RequestBuilder) SetMethod(method string) *RequestBuilder {
	rb.method = method
	return rb
}

// SetURL 设置URL
func (rb *RequestBuilder) SetURL(url string) *RequestBuilder {
	rb.url = url
	return rb
}

// SetHeader 设置请求头
func (rb *RequestBuilder) SetHeader(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// SetHeaders 设置多个请求头
func (rb *RequestBuilder) SetHeaders(headers map[string]string) *RequestBuilder {
	for k, v := range headers {
		rb.headers[k] = v
	}
	return rb
}

// SetQueryParam 设置查询参数
func (rb *RequestBuilder) SetQueryParam(key, value string) *RequestBuilder {
	rb.query[key] = value
	return rb
}

// SetQueryParams 设置多个查询参数
func (rb *RequestBuilder) SetQueryParams(params map[string]string) *RequestBuilder {
	for k, v := range params {
		rb.query[k] = v
	}
	return rb
}

// SetBody 设置请求体
func (rb *RequestBuilder) SetBody(body any) *RequestBuilder {
	rb.body = body
	return rb
}

// SetJSON 设置JSON请求体
func (rb *RequestBuilder) SetJSON(body any) *RequestBuilder {
	rb.body = body
	rb.headers["Content-Type"] = "application/json"
	return rb
}

// Build 构建请求
func (rb *RequestBuilder) Build() (*http.Request, error) {
	// 构建URL
	u := rb.url
	if len(rb.query) > 0 {
		values := url.Values{}
		for k, v := range rb.query {
			values.Set(k, v)
		}
		u += "?" + values.Encode()
	}

	// 准备请求体
	var bodyReader io.Reader
	if rb.body != nil {
		switch v := rb.body.(type) {
		case string:
			bodyReader = strings.NewReader(v)
		case []byte:
			bodyReader = bytes.NewReader(v)
		case io.Reader:
			bodyReader = v
		default:
			jsonData, err := json.Marshal(rb.body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(jsonData)
		}
	}

	// 创建请求
	req, err := http.NewRequest(rb.method, u, bodyReader)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	for k, v := range rb.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// Execute 执行请求
func (rb *RequestBuilder) Execute() (*Response, error) {
	if rb.client == nil {
		rb.client = NewClient(DefaultConfig)
	}
	return rb.client.Do(rb.method, rb.url, rb.body)
}

// FormEncoder 表单编码器
type FormEncoder struct {
	fields map[string]string
}

// NewFormEncoder 创建表单编码器
func NewFormEncoder() *FormEncoder {
	return &FormEncoder{
		fields: make(map[string]string),
	}
}

// Add 添加字段
func (fe *FormEncoder) Add(key, value string) *FormEncoder {
	fe.fields[key] = value
	return fe
}

// Encode 编码表单
func (fe *FormEncoder) Encode() string {
	values := url.Values{}
	for k, v := range fe.fields {
		values.Set(k, v)
	}
	return values.Encode()
}

// EncodeReader 编码为io.Reader
func (fe *FormEncoder) EncodeReader() io.Reader {
	return strings.NewReader(fe.Encode())
}

// MultipartEncoder 多部分编码器
type MultipartEncoder struct {
	fields map[string]io.Reader
}

// NewMultipartEncoder 创建多部分编码器
func NewMultipartEncoder() *MultipartEncoder {
	return &MultipartEncoder{
		fields: make(map[string]io.Reader),
	}
}

// Add 添加字段
func (me *MultipartEncoder) Add(key string, reader io.Reader) *MultipartEncoder {
	me.fields[key] = reader
	return me
}

// AddString 添加字符串字段
func (me *MultipartEncoder) AddString(key, value string) *MultipartEncoder {
	me.fields[key] = strings.NewReader(value)
	return me
}

// Build 构建multipart请求
func (me *MultipartEncoder) Build() (string, io.Reader, string) {
	// 简化实现，实际应该使用multipart.Writer
	body := &bytes.Buffer{}
	boundary := fmt.Sprintf("boundary-%d", time.Now().UnixNano())

	return boundary, body, boundary
}

// RetryFunc 重试函数类型
type RetryFunc func(resp *Response, err error) bool

// DefaultRetryFunc 默认重试函数
var DefaultRetryFunc RetryFunc = func(resp *Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode >= 500 || resp.StatusCode == 429
}

// WithRetry 带重试的请求
func (c *Client) WithRetry(method, path string, body any, retryFunc RetryFunc) (*Response, error) {
	var resp *Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// 指数退避
			time.Sleep(time.Duration(1<<uint(attempt)) * time.Second)
		}

		resp, err = c.Do(method, path, body)

		// 检查是否需要重试
		if retryFunc == nil || !retryFunc(resp, err) {
			break
		}
	}

	return resp, err
}

// GetWithRetry 带重试的GET请求
func (c *Client) GetWithRetry(path string, retryFunc RetryFunc) (*Response, error) {
	return c.WithRetry(http.MethodGet, path, nil, retryFunc)
}

// PostWithRetry 带重试的POST请求
func (c *Client) PostWithRetry(path string, body any, retryFunc RetryFunc) (*Response, error) {
	return c.WithRetry(http.MethodPost, path, body, retryFunc)
}
