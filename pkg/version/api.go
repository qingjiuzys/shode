// Package version 提供 API 版本控制功能。
package version

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// VersioningStrategy 版本策略
type VersioningStrategy int

const (
	URLStrategy VersioningStrategy = iota
	HeaderStrategy
	QueryParamStrategy
	ContentTypeStrategy
)

// Version 版本
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion 解析版本
func ParseVersion(versionStr string) (*Version, error) {
	re := regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(versionStr)
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// String 返回版本字符串
func (v *Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare 比较版本
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

// IsGreaterThan 是否大于
func (v *Version) IsGreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// IsLessThan 是否小于
func (v *Version) IsLessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// APIVersion API 版本
type APIVersion struct {
	Version  *Version
	Strategy VersioningStrategy
	Deprecated bool
	SunsetAt   string // 版本弃用日期
}

// VersionManager 版本管理器
type VersionManager struct {
	versions    map[string]*APIVersion
	defaultVersion *Version
	mu          sync.RWMutex
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: make(map[string]*APIVersion),
	}
}

// RegisterVersion 注册版本
func (vm *VersionManager) RegisterVersion(version string, strategy VersioningStrategy) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	v, err := ParseVersion(version)
	if err != nil {
		return err
	}

	vm.versions[version] = &APIVersion{
		Version:  v,
		Strategy: strategy,
	}

	return nil
}

// SetDefault 设置默认版本
func (vm *VersionManager) SetDefault(version string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	v, err := ParseVersion(version)
	if err != nil {
		return err
	}

	vm.defaultVersion = v
	return nil
}

// GetDefault 获取默认版本
func (vm *VersionManager) GetDefault() *Version {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	return vm.defaultVersion
}

// GetVersion 获取版本
func (vm *VersionManager) GetVersion(version string) (*APIVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	apiVer, exists := vm.versions[version]
	if !exists {
		return nil, fmt.Errorf("version not found: %s", version)
	}

	return apiVer, nil
}

// DeprecateVersion 弃用版本
func (vm *VersionManager) DeprecateVersion(version, sunsetAt string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if apiVer, exists := vm.versions[version]; exists {
		apiVer.Deprecated = true
		apiVer.SunsetAt = sunsetAt
		return nil
	}

	return fmt.Errorf("version not found: %s", version)
}

// URLVersioning URL 版本控制
type URLVersioning struct {
	manager *VersionManager
	prefix  string
}

// NewURLVersioning 创建 URL 版本控制
func NewURLVersioning(manager *VersionManager, prefix string) *URLVersioning {
	return &URLVersioning{
		manager: manager,
		prefix:  prefix,
	}
}

// ExtractVersion 从 URL 提取版本
func (uv *URLVersioning) ExtractVersion(path string) (*Version, error) {
	// 格式: /api/v1/users, /v2/ping
	re := regexp.MustCompile(`^/api/(\d+\.\d+\.\d+)/|^/v(\d+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 2 {
		return ParseVersion(matches[2])
	} else if len(matches) > 1 {
		return ParseVersion(matches[1])
	}

	return uv.manager.GetDefault(), nil
}

// BuildURL 构建 URL
func (uv *URLVersioning) BuildURL(basePath, version, path string) string {
	return fmt.Sprintf("%s/v%s/%s", uv.prefix, version, path)
}

// HeaderVersioning Header 版本控制
type HeaderVersioning struct {
	manager       *VersionManager
	headerName    string
	defaultHeader string
}

// NewHeaderVersioning 创建 Header 版本控制
func NewHeaderVersioning(manager *VersionManager, headerName, defaultHeader string) *HeaderVersioning {
	return &HeaderVersioning{
		manager:       manager,
		headerName:    headerName,
		defaultHeader: defaultHeader,
	}
}

// ExtractVersion 从 Header 提取版本
func (hv *HeaderVersioning) ExtractVersion(r *http.Request) (*Version, error) {
	version := r.Header.Get(hv.headerName)
	if version == "" {
		version = hv.defaultHeader
	}

	return ParseVersion(version)
}

// AddVersion 添加版本到 Header
func (hv *HeaderVersioning) AddVersion(header http.Header, version string) {
	header.Set(hv.headerName, version)
}

// QueryParamVersioning Query 参数版本控制
type QueryParamVersioning struct {
	manager        *VersionManager
	paramName      string
	defaultVersion string
}

// NewQueryParamVersioning 创建 Query 参数版本控制
func NewQueryParamVersioning(manager *VersionManager, paramName, defaultVersion string) *QueryParamVersioning {
	return &QueryParamVersioning{
		manager:        manager,
		paramName:      paramName,
		defaultVersion: defaultVersion,
	}
}

// ExtractVersion 从 Query 参数提取版本
func (qv *QueryParamVersioning) ExtractVersion(r *http.Request) (*Version, error) {
	version := r.URL.Query().Get(qv.paramName)
	if version == "" {
		version = qv.defaultVersion
	}

	return ParseVersion(version)
}

// BuildURL 构建 URL
func (qv *QueryParamVersioning) BuildURL(baseURL, version string) string {
	return fmt.Sprintf("%s?%s=%s", baseURL, qv.paramName, version)
}

// ContentTypeVersioning Content-Type 版本控制
type ContentTypeVersioning struct {
	manager      *VersionManager
	versionParam string
}

// NewContentTypeVersioning 创建 Content-Type 版本控制
func NewContentTypeVersioning(manager *VersionManager, versionParam string) *ContentTypeVersioning {
	return &ContentTypeVersioning{
		manager:      manager,
		versionParam: versionParam,
	}
}

// ExtractVersion 从 Content-Type 提取版本
func (ctv *ContentTypeVersioning) ExtractVersion(r *http.Request) (*Version, error) {
	contentType := r.Header.Get("Content-Type")
	// 格式: application/json; version=v1
	params := strings.Split(contentType, ";")
	for _, param := range params {
		param = strings.TrimSpace(param)
		if strings.HasPrefix(param, ctv.versionParam+"=") {
			version := strings.TrimPrefix(param, ctv.versionParam+"=")
			return ParseVersion(version)
		}
	}

	return ctv.manager.GetDefault(), nil
}

// VersionNegotiation 版本协商
type VersionNegotiation struct {
	manager  *VersionManager
	supported []*Version
}

// NewVersionNegotiation 创建版本协商
func NewVersionNegotiation(manager *VersionManager) *VersionNegotiation {
	return &VersionNegotiation{
		manager: manager,
		supported: make([]*Version, 0),
	}
}

// AddSupported 添加支持的版本
func (vn *VersionNegotiation) AddSupported(version string) error {
	v, err := ParseVersion(version)
	if err != nil {
		return err
	}

	vn.supported = append(vn.supported, v)
	return nil
}

// Negotiate 协商版本
func (vn *VersionNegotiation) Negotiate(requestedVersion string) (*Version, error) {
	// 解析请求的版本
	requested, err := ParseVersion(requestedVersion)
	if err != nil {
		// 如果解析失败，返回默认版本
		return vn.manager.GetDefault(), nil
	}

	// 查找最匹配的版本
	for _, supported := range vn.supported {
		if supported.Compare(requested) == 0 {
			return supported, nil
		}
	}

	// 返回默认版本
	return vn.manager.GetDefault(), nil
}

// BackwardCompatibility 向后兼容
type BackwardCompatibility struct {
	manager        *VersionManager
	compatibility map[string]*CompatibilityRule
}

// CompatibilityRule 兼容规则
type CompatibilityRule struct {
	FromVersion string
	ToVersion   string
	Type        string // "added", "removed", "changed"
	Field       string
	Description string
}

// NewBackwardCompatibility 创建向后兼容
func NewBackwardCompatibility(manager *VersionManager) *BackwardCompatibility {
	return &BackwardCompatibility{
		manager:        manager,
		compatibility: make(map[string]*CompatibilityRule),
	}
}

// AddRule 添加兼容规则
func (bc *BackwardCompatibility) AddRule(rule *CompatibilityRule) {
	bc.compatibility[rule.FromVersion+"->"+rule.ToVersion] = rule
}

// CheckCompatibility 检查兼容性
func (bc *BackwardCompatibility) CheckCompatibility(fromVersion, toVersion string) (*CompatibilityReport, error) {
	report := &CompatibilityReport{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Compatible:  true,
		Warnings:    make([]string, 0),
		Errors:      make([]string, 0),
	}

	// 检查主版本变更
	from, _ := ParseVersion(fromVersion)
	to, _ := ParseVersion(toVersion)

	if to.Major > from.Major {
		report.Compatible = false
		report.Errors = append(report.Errors, "Major version increment indicates breaking changes")
	}

	return report, nil
}

// CompatibilityReport 兼容报告
type CompatibilityReport struct {
	FromVersion string
	ToVersion   string
	Compatible  bool
	Warnings    []string
	Errors      []string
}

// Contract 契约
type Contract struct {
	Version       string
	Request       map[string]*FieldSpec
	Response      map[string]*FieldSpec
	Required      []string
	Optional      []string
	Enums         map[string]*EnumSpec
	Constants     map[string]interface{}
}

// FieldSpec 字段规范
type FieldSpec struct {
	Type        string
	Required    bool
	Nullable    bool
	Description string
	Example     interface{}
	Validation  *ValidationSpec
}

// EnumSpec 枚举规范
type EnumSpec struct {
	Values []string
}

// ValidationSpec 验证规范
type ValidationSpec struct {
	MinLength *int
	MaxLength *int
	Min       *float64
	Max       *float64
	Pattern   *string
}

// ContractTester 契约测试
type ContractTester struct {
	contracts map[string]*Contract
	mu        sync.RWMutex
}

// NewContractTester 创建契约测试器
func NewContractTester() *ContractTester {
	return &ContractTester{
		contracts: make(map[string]*Contract),
	}
}

// AddContract 添加契约
func (ct *ContractTester) AddContract(endpoint string, contract *Contract) {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	ct.contracts[endpoint] = contract
}

// TestRequest 测试请求
func (ct *ContractTester) TestRequest(endpoint string, data interface{}) (*TestResult, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	contract, exists := ct.contracts[endpoint]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", endpoint)
	}

	result := &TestResult{
		Passed: true,
		Errors: make([]string, 0),
	}

	// 验证必需字段
	for _, field := range contract.Required {
		if !hasField(data, field) {
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("missing required field: %s", field))
		}
	}

	// 验证类型
	for field, spec := range contract.Request {
		if hasField(data, field) {
			if err := validateType(getField(data, field), spec.Type); err != nil {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("field %s validation failed: %v", field, err))
			}
		}
	}

	return result, nil
}

// TestResponse 测试响应
func (ct *ContractTester) TestResponse(endpoint string, data interface{}) (*TestResult, error) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	contract, exists := ct.contracts[endpoint]
	if !exists {
		return nil, fmt.Errorf("contract not found: %s", endpoint)
	}

	result := &TestResult{
		Passed: true,
		Errors: make([]string, 0),
	}

	// 验证响应类型
	for field, spec := range contract.Response {
		if hasField(data, field) {
			if err := validateType(getField(data, field), spec.Type); err != nil {
				result.Passed = false
				result.Errors = append(result.Errors, fmt.Sprintf("response field %s validation failed: %v", field, err))
			}
		}
	}

	return result, nil
}

// TestResult 测试结果
type TestResult struct {
	Passed bool
	Errors []string
}

// hasField 检查字段是否存在
func hasField(data interface{}, field string) bool {
	// 简化实现
	return false
}

// getField 获取字段值
func getField(data interface{}, field string) interface{} {
	// 简化实现
	return nil
}

// validateType 验证类型
func validateType(value interface{}, typ string) error {
	// 简化实现
	return nil
}

// VersionMigration 版本迁移
type VersionMigration struct {
	fromVersion string
	toVersion   string
	migrations  []MigrationFunc
}

// MigrationFunc 迁移函数
type MigrationFunc func(ctx context.Context, data interface{}) (interface{}, error)

// NewVersionMigration 创建版本迁移
func NewVersionMigration(fromVersion, toVersion string) *VersionMigration {
	return &VersionMigration{
		fromVersion: fromVersion,
		toVersion:   toVersion,
		migrations:  make([]MigrationFunc, 0),
	}
}

// AddMigration 添加迁移
func (vm *VersionMigration) AddMigration(fn MigrationFunc) {
	vm.migrations = append(vm.migrations, fn)
}

// Migrate 执行迁移
func (vm *VersionMigration) Migrate(ctx context.Context, data interface{}) (interface{}, error) {
	var err error
	for _, migration := range vm.migrations {
		data, err = migration(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("migration failed: %w", err)
		}
	}
	return data, nil
}

// VersionRouter 版本路由器
type VersionRouter struct {
	manager  *VersionManager
	routes   map[string]map[string]*Route // version -> path -> route
	mu       sync.RWMutex
}

// Route 路由
type Route struct {
	Path    string
	Handler http.Handler
	Version string
}

// NewVersionRouter 创建版本路由器
func NewVersionRouter(manager *VersionManager) *VersionRouter {
	return &VersionRouter{
		manager: manager,
		routes:  make(map[string]map[string]*Route),
	}
}

// Register 注册路由
func (vr *VersionRouter) Register(version, path string, handler http.Handler) error {
	vr.mu.Lock()
	defer vr.mu.Unlock()

	if _, exists := vr.routes[version]; !exists {
		vr.routes[version] = make(map[string]*Route)
	}

	vr.routes[version][path] = &Route{
		Path:    path,
		Handler: handler,
		Version: version,
	}

	return nil
}

// Route 路由请求
func (vr *VersionRouter) Route(r *http.Request) (http.Handler, *Version, error) {
	// 提取版本
	version, err := vr.extractVersion(r)
	if err != nil {
		return nil, nil, err
	}

	// 查找路由
	vr.mu.RLock()
	defer vr.mu.RUnlock()

	if routes, exists := vr.routes[version.String()]; exists {
		if route, exists := routes[r.URL.Path]; exists {
			return route.Handler, version, nil
		}
	}

	return nil, nil, fmt.Errorf("route not found: %s %s", version.String(), r.URL.Path)
}

// extractVersion 提取版本
func (vr *VersionRouter) extractVersion(r *http.Request) (*Version, error) {
	// 尝试从 URL 提取
	if version, err := ParseVersionFromURL(r.URL.Path); err == nil {
		return version, nil
	}

	// 尝试从 Header 提取
	if version, err := ParseVersionFromHeader(r); err == nil {
		return version, nil
	}

	// 返回默认版本
	return vr.manager.GetDefault(), nil
}

// ParseVersionFromURL 从 URL 解析版本
func ParseVersionFromURL(path string) (*Version, error) {
	re := regexp.MustCompile(`/v(\d+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) < 2 {
		return nil, fmt.Errorf("no version in URL")
	}

	return &Version{Major: 1, Minor: 0, Patch: 0}, nil
}

// ParseVersionFromHeader 从 Header 解析版本
func ParseVersionFromHeader(r *http.Request) (*Version, error) {
	version := r.Header.Get("API-Version")
	if version == "" {
		return nil, fmt.Errorf("no version in header")
	}

	return ParseVersion(version)
}
