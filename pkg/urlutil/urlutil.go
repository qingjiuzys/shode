// Package urlutil 提供URL处理工具
package urlutil

import (
	"net/url"
	"path"
	"strings"
)

// Join 连接URL路径
func Join(base string, paths ...string) string {
	p := path.Join(paths...)
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(p, "/")
}

// Parse 解析URL
func Parse(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// MustParse 解析URL（panic on error）
func MustParse(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

// Build 构建URL
func Build(scheme, host, path string) string {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	return u.String()
}

// BuildWithQuery 构建带查询参数的URL
func BuildWithQuery(scheme, host, path string, query map[string]string) string {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// AddQueryParam 添加查询参数
func AddQueryParam(rawURL, key, value string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()

	return u.String()
}

// AddQueryParams 添加多个查询参数
func AddQueryParams(rawURL string, params map[string]string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// RemoveQueryParam 移除查询参数
func RemoveQueryParam(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()

	return u.String()
}

// GetQueryParam 获取查询参数
func GetQueryParam(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Query().Get(key)
}

// GetQueryParams 获取所有查询参数
func GetQueryParams(rawURL string) map[string]string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return make(map[string]string)
	}

	query := u.Query()
	result := make(map[string]string, len(query))

	for k, v := range query {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}

	return result
}

// ParseQuery 解析查询字符串
func ParseQuery(query string) map[string]string {
	values, err := url.ParseQuery(query)
	if err != nil {
		return make(map[string]string)
	}

	result := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) > 0 {
			result[k] = v[0]
		}
	}

	return result
}

// EncodeQuery 编码查询参数
func EncodeQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// GetScheme 获取协议
func GetScheme(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Scheme
}

// GetHost 获取主机
func GetHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Host
}

// GetPath 获取路径
func GetPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Path
}

// GetFragment 获取片段
func GetFragment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Fragment
}

// SetScheme 设置协议
func SetScheme(rawURL, scheme string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Scheme = scheme
	return u.String()
}

// SetHost 设置主机
func SetHost(rawURL, host string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Host = host
	return u.String()
}

// SetPath 设置路径
func SetPath(rawURL, path string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Path = path
	return u.String()
}

// SetFragment 设置片段
func SetFragment(rawURL, fragment string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.Fragment = fragment
	return u.String()
}

// Base 获取基础URL（不包含查询参数和片段）
func Base(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.RawQuery = ""
	u.Fragment = ""

	return u.String()
}

// Clean 清理URL路径
func Clean(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Path = path.Clean(u.Path)
	return u.String()
}

// Resolve 解析相对URL
func Resolve(base, rel string) string {
	u, err := url.Parse(base)
	if err != nil {
		return rel
	}

	baseURL, err := u.Parse(rel)
	if err != nil {
		return rel
	}

	return baseURL.String()
}

// IsAbsolute 检查是否为绝对URL
func IsAbsolute(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.IsAbs()
}

// IsRelative 检查是否为相对URL
func IsRelative(rawURL string) bool {
	return !IsAbsolute(rawURL)
}

// IsValid 检查URL是否有效
func IsValid(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}

// HasScheme 检查是否有协议
func HasScheme(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme != ""
}

// HasQuery 检查是否有查询参数
func HasQuery(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.RawQuery != ""
}

// HasFragment 检查是否有片段
func HasFragment(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Fragment != ""
}

// Encode 编码URL组件
func Encode(s string) string {
	return url.QueryEscape(s)
}

// Decode 解码URL组件
func Decode(s string) (string, error) {
	return url.QueryUnescape(s)
}

// EncodePath 编码路径
func EncodePath(s string) string {
	return url.PathEscape(s)
}

// DecodePath 解码路径
func DecodePath(s string) (string, error) {
	return url.PathUnescape(s)
}

// Split 分割URL
func Split(rawURL string) (base, query string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, ""
	}

	u.RawQuery = ""
	base = u.String()
	query = u.Query().Encode()

	return base, query
}

// GetFileName 从URL获取文件名
func GetFileName(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return path.Base(u.Path)
}

// GetExtension 从URL获取文件扩展名
func GetExtension(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return path.Ext(u.Path)
}

// TrimFragment 移除片段
func TrimFragment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Fragment = ""
	return u.String()
}

// TrimQuery 移除查询参数
func TrimQuery(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.RawQuery = ""
	return u.String()
}

// TrimQueryAndFragment 移除查询参数和片段
func TrimQueryAndFragment(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

// Normalize 规范化URL
func Normalize(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// 转换为小写协议
	u.Scheme = strings.ToLower(u.Scheme)

	// 转换为小写主机
	u.Host = strings.ToLower(u.Host)

	// 清理路径
	u.Path = path.Clean(u.Path)

	return u.String()
}

// WithUserInfo 添加用户信息
func WithUserInfo(rawURL, username, password string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.User = url.UserPassword(username, password)
	return u.String()
}

// GetUserInfo 获取用户信息
func GetUserInfo(rawURL string) (username, password string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", ""
	}

	if u.User == nil {
		return "", ""
	}

	username = u.User.Username()
	password, _ = u.User.Password()
	return username, password
}

// GetPort 获取端口
func GetPort(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Port()
}

// SetPort 设置端口
func SetPort(rawURL string, port string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Host = u.Hostname() + ":" + port
	return u.String()
}

// GetHostname 获取主机名（不含端口）
func GetHostname(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return u.Hostname()
}

// Clone 克隆URL
func Clone(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	return u.String()
}

// Merge 合并查询参数
func Merge(base string, params map[string]string) string {
	return AddQueryParams(base, params)
}

// Equals 比较两个URL是否相等（忽略片段）
func Equals(url1, url2 string) bool {
	u1, err1 := url.Parse(url1)
	u2, err2 := url.Parse(url2)

	if err1 != nil || err2 != nil {
		return false
	}

	u1.Fragment = ""
	u2.Fragment = ""

	return u1.String() == u2.String()
}

// IsHTTP 检查是否为HTTP协议
func IsHTTP(rawURL string) bool {
	return GetScheme(rawURL) == "http"
}

// IsHTTPS 检查是否为HTTPS协议
func IsHTTPS(rawURL string) bool {
	return GetScheme(rawURL) == "https"
}

// IsWS 检查是否为WebSocket协议
func IsWS(rawURL string) bool {
	return GetScheme(rawURL) == "ws"
}

// IsWSS 检查是否为WebSocket Secure协议
func IsWSS(rawURL string) bool {
	return GetScheme(rawURL) == "wss"
}

// ToHTTPS 转换为HTTPS
func ToHTTPS(rawURL string) string {
	return SetScheme(rawURL, "https")
}

// ToHTTP 转换为HTTP
func ToHTTP(rawURL string) string {
	return SetScheme(rawURL, "http")
}

// StripPrefix 移除路径前缀
func StripPrefix(rawURL, prefix string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Path = strings.TrimPrefix(u.Path, prefix)
	return u.String()
}

// AddPrefix 添加路径前缀
func AddPrefix(rawURL, prefix string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Path = prefix + u.Path
	return u.String()
}

// TrimSlash 移除路径两端的斜杠
func TrimSlash(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	u.Path = strings.Trim(u.Path, "/")
	return u.String()
}

// EnsureLeadingSlash 确保路径以斜杠开头
func EnsureLeadingSlash(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if !strings.HasPrefix(u.Path, "/") {
		u.Path = "/" + u.Path
	}

	return u.String()
}

// EnsureTrailingSlash 确保路径以斜杠结尾
func EnsureTrailingSlash(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if !strings.HasSuffix(u.Path, "/") {
		if u.Path == "" {
			u.Path = "/"
		} else {
			u.Path += "/"
		}
	}

	return u.String()
}

// Builder URL构建器
type Builder struct {
	u *url.URL
}

// NewBuilder 创建URL构建器
func NewBuilder() *Builder {
	return &Builder{
		u: &url.URL{},
	}
}

// ParseBuilder 从URL解析创建构建器
func ParseBuilder(rawURL string) (*Builder, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	return &Builder{u: u}, nil
}

// SetScheme 设置协议
func (b *Builder) SetScheme(scheme string) *Builder {
	b.u.Scheme = scheme
	return b
}

// SetHost 设置主机
func (b *Builder) SetHost(host string) *Builder {
	b.u.Host = host
	return b
}

// SetPath 设置路径
func (b *Builder) SetPath(p string) *Builder {
	b.u.Path = p
	return b
}

// AddPath 添加路径
func (b *Builder) AddPath(p string) *Builder {
	b.u.Path = path.Join(b.u.Path, p)
	return b
}

// SetFragment 设置片段
func (b *Builder) SetFragment(fragment string) *Builder {
	b.u.Fragment = fragment
	return b
}

// AddQueryParam 添加查询参数
func (b *Builder) AddQueryParam(key, value string) *Builder {
	q := b.u.Query()
	q.Set(key, value)
	b.u.RawQuery = q.Encode()
	return b
}

// AddQueryParams 添加多个查询参数
func (b *Builder) AddQueryParams(params map[string]string) *Builder {
	q := b.u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	b.u.RawQuery = q.Encode()
	return b
}

// SetQueryParam 设置查询参数
func (b *Builder) SetQueryParam(key, value string) *Builder {
	q := b.u.Query()
	q.Set(key, value)
	b.u.RawQuery = q.Encode()
	return b
}

// RemoveQueryParam 移除查询参数
func (b *Builder) RemoveQueryParam(key string) *Builder {
	q := b.u.Query()
	q.Del(key)
	b.u.RawQuery = q.Encode()
	return b
}

// Build 构建URL
func (b *Builder) Build() string {
	return b.u.String()
}

// String 字符串表示
func (b *Builder) String() string {
	return b.u.String()
}

// Get 获取url.URL
func (b *Builder) Get() *url.URL {
	return b.u
}
