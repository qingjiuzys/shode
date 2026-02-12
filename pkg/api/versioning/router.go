// Package versioning 提供API版本控制功能
package versioning

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Version API版本
type Version struct {
	Major int
	Minor int
	Patch  int
	Pre    string
}

// VersionedRoute 版本化路由
type VersionedRoute struct {
	Version     *Version
	Path        string
	Handler     http.Handler
	Deprecated  bool
	DeprecatedIn *Version
}

// Router 版本路由器
type Router struct {
	versions        map[string]*Version
	routes          map[string][]*VersionedRoute
	defaultVersion  *Version
	latestVersion   *Version
	versionHeader   string
}

// NewRouter 创建版本路由器
func NewRouter() *Router {
	return &Router{
		versions:       make(map[string]*Version),
		routes:         make(map[string][]*VersionedRoute),
		versionHeader:   "API-Version",
		defaultVersion: ParseVersion("1.0.0"),
		latestVersion:  ParseVersion("1.0.0"),
	}
}

// ParseVersion 解析版本字符串
func ParseVersion(versionStr string) *Version {
	// 支持格式: 1.0.0, 1.0.0-alpha, 1.0.0-beta.1
	pattern := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$`)
	matches := pattern.FindStringSubmatch(versionStr)

	if len(matches) < 4 {
		return nil
	}

	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])
	pre := ""
	if len(matches) > 4 {
		pre = matches[4]
	}

	return &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
		Pre:   pre,
	}
}

// String 版本字符串
func (v *Version) String() string {
	if v.Pre != "" {
		return fmt.Sprintf("%d.%d.%d-%s", v.Major, v.Minor, v.Patch, v.Pre)
	}
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare 比较版本
// 返回值: -1 (v < other), 0 (v == other), 1 (v > other)
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	// 比较预发布版本
	if v.Pre != "" || other.Pre != "" {
		if v.Pre == "" {
			return 1 // 正式版本 > 预发布版本
		}
		if other.Pre == "" {
			return -1
		}

		// 简单比较预发布字符串
		if v.Pre < other.Pre {
			return -1
		}
		if v.Pre > other.Pre {
			return 1
		}
	}

	return 0
}

// IsGreaterThan 检查版本是否大于
func (v *Version) IsGreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// IsLessThan 检查版本是否小于
func (v *Version) IsLessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// Equals 检查版本是否相等
func (v *Version) Equals(other *Version) bool {
	return v.Compare(other) == 0
}

// AddVersion 添加API版本
func (r *Router) AddVersion(version string) error {
	v := ParseVersion(version)
	if v == nil {
		return fmt.Errorf("invalid version: %s", version)
	}

	r.versions[version] = v

	// 更新最新版本
	if r.latestVersion == nil || v.IsGreaterThan(r.latestVersion) {
		r.latestVersion = v
	}

	return nil
}

// SetDefaultVersion 设置默认版本
func (r *Router) SetDefaultVersion(version string) error {
	v := ParseVersion(version)
	if v == nil {
		return fmt.Errorf("invalid version: %s", version)
	}

	r.defaultVersion = v
	return nil
}

// SetLatestVersion 设置最新版本
func (r *Router) SetLatestVersion(version string) error {
	v := ParseVersion(version)
	if v == nil {
		return fmt.Errorf("invalid version: %s", version)
	}

	r.latestVersion = v
	return nil
}

// Register 注册版本化路由
func (r *Router) Register(version, path string, handler http.Handler) *VersionedRoute {
	route := &VersionedRoute{
		Version:    r.versions[version],
		Path:       path,
		Handler:    handler,
		Deprecated: false,
	}

	versionRoutes := r.routes[path]
	versionRoutes = append(versionRoutes, route)
	r.routes[path] = versionRoutes

	return route
}

// RegisterDeprecated 注册弃用路由
func (r *Router) RegisterDeprecated(version, path string, deprecatedIn string, handler http.Handler) (*VersionedRoute, error) {
	v := r.versions[version]
	if v == nil {
		return nil, fmt.Errorf("version not found: %s", version)
	}

	deprecatedVersion := r.versions[deprecatedIn]
	if deprecatedVersion == nil {
		return nil, fmt.Errorf("deprecated version not found: %s", deprecatedIn)
	}

	route := &VersionedRoute{
		Version:      v,
		Path:         path,
		Handler:      handler,
		Deprecated:   true,
		DeprecatedIn: deprecatedVersion,
	}

	versionRoutes := r.routes[path]
	versionRoutes = append(versionRoutes, route)
	r.routes[path] = versionRoutes

	return route, nil
}

// ServeHTTP 处理HTTP请求
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 检查版本头
	clientVersion := req.Header.Get(r.versionHeader)
	if clientVersion == "" {
		clientVersion = r.defaultVersion.String()
	}

	// 解析路径
	path := req.URL.Path

	// 查找匹配的路由
	if routes, ok := r.routes[path]; ok {
		// 优先使用客户端指定的版本
		route := r.findRouteForVersion(routes, clientVersion)
		if route != nil {
			r.serveRoute(w, req, route)
			return
		}

		// 尝试使用默认版本
		route = r.findRouteForVersion(routes, r.defaultVersion.String())
		if route != nil {
			r.serveRoute(w, req, route)
			return
		}
	}

	// 没有找到路由
	http.NotFound(w, req)
}

// findRouteForVersion 查找指定版本的路由
func (r *Router) findRouteForVersion(routes []*VersionedRoute, version string) *VersionedRoute {
	v := ParseVersion(version)
	if v == nil {
		return nil
	}

	// 首先尝试精确匹配
	for _, route := range routes {
		if route.Version.Equals(v) && !route.Deprecated {
			return route
		}
	}

	// 尝试找到兼容的版本
	for _, route := range routes {
		if !route.Deprecated && route.Version.Major == v.Major {
			// 检查版本兼容性：相同主版本，客户端版本 <= 服务端版本
			if v.IsLessThan(route.Version) || v.Equals(route.Version) {
				return route
			}
		}
	}

	return nil
}

// serveRoute 处理路由
func (r *Router) serveRoute(w http.ResponseWriter, req *http.Request, route *VersionedRoute) {
	// 添加版本信息到响应头
	w.Header().Set("API-Version", route.Version.String())

	if route.Deprecated {
		w.Header().Set("Deprecated", "true")
		if route.DeprecatedIn != nil {
			w.Header().Set("Sunset", route.DeprecatedIn.String())
		}
	}

	route.Handler.ServeHTTP(w, req)
}

// GetVersion 获取版本信息
func (r *Router) GetVersion() *VersionInfo {
	return &VersionInfo{
		Default: r.defaultVersion,
		Latest:  r.latestVersion,
		Versions: r.getAllVersions(),
	}
}

// VersionInfo 版本信息
type VersionInfo struct {
	Default  *Version            `json:"default"`
	Latest   *Version            `json:"latest"`
	Versions []string            `json:"versions"`
}

// getAllVersions 获取所有版本
func (r *Router) getAllVersions() []string {
	versions := make([]string, 0, len(r.versions))
	for v := range r.versions {
		versions = append(versions, v)
	}
	return versions
}

// GetSupportedVersions 获取支持的版本列表
func (r *Router) GetSupportedVersions(path string) []string {
	if routes, ok := r.routes[path]; ok {
		versions := make([]string, 0)
		for _, route := range routes {
			if !route.Deprecated {
				versions = append(versions, route.Version.String())
			}
		}
		return versions
	}
	return nil
}

// VersionMiddleware 版本中间件
type VersionMiddleware struct {
	router *Router
}

// NewVersionMiddleware 创建版本中间件
func NewVersionMiddleware(router *Router) *VersionMiddleware {
	return &VersionMiddleware{router: router}
}

// ServeHTTP 实现中间件接口
func (vm *VersionMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	vm.router.ServeHTTP(w, req)
}

// NegotiateVersion 协商版本
func (r *Router) NegotiateVersion(req *http.Request, versions []string) (*Version, error) {
	// 1. 检查请求头
	clientVersion := req.Header.Get(r.versionHeader)
	if clientVersion != "" {
		for _, v := range versions {
			if v == clientVersion {
				return ParseVersion(v), nil
			}
		}
	}

	// 2. 检查查询参数
	queryVersion := req.URL.Query().Get("version")
	if queryVersion != "" {
		for _, v := range versions {
			if v == queryVersion {
				return ParseVersion(v), nil
			}
		}
	}

	// 3. 使用默认版本
	return r.defaultVersion, nil
}

// CheckVersion 检查版本兼容性
func (r *Router) CheckVersion(clientVersion, minVersion, maxVersion string) error {
	cv := ParseVersion(clientVersion)
	minV := ParseVersion(minVersion)
	maxV := ParseVersion(maxVersion)

	if cv == nil || minV == nil || maxV == nil {
		return fmt.Errorf("invalid version format")
	}

	if cv.IsLessThan(minV) {
		return fmt.Errorf("client version %s is less than minimum required version %s", clientVersion, minVersion)
	}

	if cv.IsGreaterThan(maxV) {
		return fmt.Errorf("client version %s is greater than maximum supported version %s", clientVersion, maxVersion)
	}

	return nil
}

// VersionHandler 版本信息处理器
func (r *Router) VersionHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		info := r.GetVersion()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

// ParseVersionFromPath 从路径解析版本
func ParseVersionFromPath(path string) (*Version, string, error) {
	// 支持格式: /v1/users, /v1.0.0/users
	pattern := regexp.MustCompile(`^/v(\d+)(?:\.(\d+)(?:\.(\d+))?(?:/|$)`)
	matches := pattern.FindStringSubmatch(path)

	if len(matches) < 2 {
		return nil, "", fmt.Errorf("no version found in path")
	}

	major, _ := strconv.Atoi(matches[1])
	minor := 0
	patch := 0

	if len(matches) > 2 {
		minor, _ = strconv.Atoi(matches[2])
	}
	if len(matches) > 3 {
		patch, _ = strconv.Atoi(matches[3])
	}

	version := &Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	// 移除版本前缀
	remainingPath := path
	if len(matches[1]) > 0 {
		idx := strings.Index(path, matches[1])
		if idx > 0 {
			remainingPath = path[idx:]
		}
	}

	return version, remainingPath, nil
}

// BuildVersionedPath 构建版本化路径
func BuildVersionedPath(version, path string) string {
	v := ParseVersion(version)
	if v == nil {
		return "/" + path
	}

	return fmt.Sprintf("/v%d/%s", v.Major, path)
}

// GetDeprecationWarnings 获取弃用警告
func (r *Router) GetDeprecationWarnings() []DeprecationWarning {
	warnings := make([]DeprecationWarning, 0)

	for path, routes := range r.routes {
		for _, route := range routes {
			if route.Deprecated {
				warnings = append(warnings, DeprecationWarning{
					Path:        path,
					Version:     route.Version.String(),
					SunsetIn:    route.DeprecatedIn.String(),
					Message:     fmt.Sprintf("API %s at version %s is deprecated, please use %s", path, route.Version.String(), route.DeprecatedIn.String()),
				})
			}
		}
	}

	return warnings
}

// DeprecationWarning 弃用警告
type DeprecationWarning struct {
	Path     string `json:"path"`
	Version  string `json:"version"`
	SunsetIn string `json:"sunset_in"`
	Message  string `json:"message"`
}

// ValidateVersion 验证版本格式
func ValidateVersion(version string) bool {
	return ParseVersion(version) != nil
}

// GetLatestVersion 获取最新版本
func (r *Router) GetLatestVersion() *Version {
	return r.latestVersion
}

// GetDefaultVersion 获取默认版本
func (r *Router) GetDefaultVersion() *Version {
	return r.defaultVersion
}

// IsVersionSupported 检查版本是否支持
func (r *Router) IsVersionSupported(version string) bool {
	if _, ok := r.versions[version]; ok {
		return !strings.HasPrefix(version, "0.") // 排除0.x版本
	}
	return false
}

// GetVersionChangelog 获取版本变更日志
type Changelog struct {
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Changes     []string `json:"changes"`
	Date        string   `json:"date"`
	Deprecated  bool     `json:"deprecated"`
}

type ChangelogManager struct {
	changelogs map[string]*Changelog
}

// NewChangelogManager 创建变更日志管理器
func NewChangelogManager() *ChangelogManager {
	return &ChangelogManager{
		changelogs: make(map[string]*Changelog),
	}
}

// Add 添加变更日志
func (cm *ChangelogManager) Add(version, description string, changes []string) {
	cm.changelogs[version] = &Changelog{
		Version:     version,
		Description: description,
		Changes:     changes,
		Date:        "2024-01-01",
		Deprecated:  false,
	}
}

// Deprecate 弃用版本
func (cm *ChangelogManager) Deprecate(version, sunsetVersion string) error {
	log, ok := cm.changelogs[version]
	if !ok {
		return fmt.Errorf("version not found: %s", version)
	}

	log.Deprecated = true

	return nil
}

// GetChangelog 获取变更日志
func (cm *ChangelogManager) GetChangelog(version string) (*Changelog, bool) {
	log, ok := cm.changelogs[version]
	return log, ok
}

// GetAllChangelogs 获取所有变更日志
func (cm *ChangelogManager) GetAllChangelogs() []*Changelog {
	logs := make([]*Changelog, 0, len(cm.changelogs))
	for _, log := range cm.changelogs {
		logs = append(logs, log)
	}
	return logs
}

// GetChangelogsSince 获取指定版本之后的变更日志
func (cm *ChangelogManager) GetChangelogsSince(version string) []*Changelog {
	v := ParseVersion(version)
	if v == nil {
		return nil
	}

	logs := make([]*Changelog, 0)
	for _, log := range cm.changelogs {
		logVersion := ParseVersion(log.Version)
		if logVersion != nil && logVersion.IsGreaterThan(v) {
			logs = append(logs, log)
		}
	}

	return logs
}
