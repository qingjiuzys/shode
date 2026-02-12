// Package context 提供上下文管理功能
package contextx

import (
	"context"
	"sync"
	"time"
)

// ContextKey 上下文键类型
type ContextKey string

// 常用上下文键
const (
	RequestIDKey    ContextKey = "request_id"
	UserIDKey       ContextKey = "user_id"
	TraceIDKey      ContextKey = "trace_id"
	SessionIDKey    ContextKey = "session_id"
	CorrelationIDKey ContextKey = "correlation_id"
	LanguageKey     ContextKey = "language"
	TimeZoneKey     ContextKey = "timezone"
	IPKey           ContextKey = "ip"
	UserAgentKey    ContextKey = "user_agent"
	HostKey         ContextKey = "host"
)

// WithRequestID 设置请求ID
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID 获取请求ID
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithUserID 设置用户ID
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserID 获取用户ID
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return ""
}

// WithTraceID 设置追踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 获取追踪ID
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithSessionID 设置会话ID
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, SessionIDKey, sessionID)
}

// GetSessionID 获取会话ID
func GetSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value(SessionIDKey).(string); ok {
		return sessionID
	}
	return ""
}

// WithCorrelationID 设置关联ID
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// GetCorrelationID 获取关联ID
func GetCorrelationID(ctx context.Context) string {
	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}
	return ""
}

// WithLanguage 设置语言
func WithLanguage(ctx context.Context, language string) context.Context {
	return context.WithValue(ctx, LanguageKey, language)
}

// GetLanguage 获取语言
func GetLanguage(ctx context.Context) string {
	if language, ok := ctx.Value(LanguageKey).(string); ok {
		return language
	}
	return "en"
}

// WithTimeZone 设置时区
func WithTimeZone(ctx context.Context, timezone string) context.Context {
	return context.WithValue(ctx, TimeZoneKey, timezone)
}

// GetTimeZone 获取时区
func GetTimeZone(ctx context.Context) string {
	if timezone, ok := ctx.Value(TimeZoneKey).(string); ok {
		return timezone
	}
	return "UTC"
}

// WithIP 设置IP地址
func WithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IPKey, ip)
}

// GetIP 获取IP地址
func GetIP(ctx context.Context) string {
	if ip, ok := ctx.Value(IPKey).(string); ok {
		return ip
	}
	return ""
}

// WithUserAgent 设置User-Agent
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

// GetUserAgent 获取User-Agent
func GetUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(UserAgentKey).(string); ok {
		return userAgent
	}
	return ""
}

// WithHost 设置Host
func WithHost(ctx context.Context, host string) context.Context {
	return context.WithValue(ctx, HostKey, host)
}

// GetHost 获取Host
func GetHost(ctx context.Context) string {
	if host, ok := ctx.Value(HostKey).(string); ok {
		return host
	}
	return ""
}

// SetValue 设置任意值
func SetValue(ctx context.Context, key ContextKey, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

// GetValue 获取任意值
func GetValue(ctx context.Context, key ContextKey) any {
	return ctx.Value(key)
}

// ContextMetadata 上下文元数据
type ContextMetadata struct {
	mu     sync.RWMutex
	values map[ContextKey]any
}

// NewContextMetadata 创建上下文元数据
func NewContextMetadata() *ContextMetadata {
	return &ContextMetadata{
		values: make(map[ContextKey]any),
	}
}

// Set 设置值
func (m *ContextMetadata) Set(key ContextKey, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values[key] = value
}

// Get 获取值
func (m *ContextMetadata) Get(key ContextKey) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.values[key]
}

// Delete 删除值
func (m *ContextMetadata) Delete(key ContextKey) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.values, key)
}

// Clear 清空所有值
func (m *ContextMetadata) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = make(map[ContextKey]any)
}

// Keys 获取所有键
func (m *ContextMetadata) Keys() []ContextKey {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]ContextKey, 0, len(m.values))
	for key := range m.values {
		keys = append(keys, key)
	}
	return keys
}

// Size 获取大小
func (m *ContextMetadata) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.values)
}

// Clone 克隆元数据
func (m *ContextMetadata) Clone() *ContextMetadata {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := NewContextMetadata()
	for key, value := range m.values {
		clone.values[key] = value
	}
	return clone
}

// ToMap 转换为map
func (m *ContextMetadata) ToMap() map[ContextKey]any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[ContextKey]any, len(m.values))
	for key, value := range m.values {
		result[key] = value
	}
	return result
}

// WithMetadata 在上下文中添加元数据
func WithMetadata(ctx context.Context, metadata *ContextMetadata) context.Context {
	return SetValue(ctx, "__metadata__", metadata)
}

// GetMetadata 从上下文中获取元数据
func GetMetadata(ctx context.Context) *ContextMetadata {
	if metadata, ok := ctx.Value(ContextKey("__metadata__")).(*ContextMetadata); ok {
		return metadata
	}
	return NewContextMetadata()
}

// ContextBag 上下文包（用于存储临时数据）
type ContextBag struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewContextBag 创建上下文包
func NewContextBag() *ContextBag {
	return &ContextBag{
		data: make(map[string]any),
	}
}

// Set 设置数据
func (b *ContextBag) Set(key string, value any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data[key] = value
}

// Get 获取数据
func (b *ContextBag) Get(key string) any {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.data[key]
}

// GetString 获取字符串
func (b *ContextBag) GetString(key string) string {
	if val := b.Get(key); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// GetInt 获取整数
func (b *ContextBag) GetInt(key string) int {
	if val := b.Get(key); val != nil {
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

// GetBool 获取布尔值
func (b *ContextBag) GetBool(key string) bool {
	if val := b.Get(key); val != nil {
		if bl, ok := val.(bool); ok {
			return bl
		}
	}
	return false
}

// Delete 删除数据
func (b *ContextBag) Delete(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.data, key)
}

// Has 检查是否存在
func (b *ContextBag) Has(key string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	_, ok := b.data[key]
	return ok
}

// Clear 清空数据
func (b *ContextBag) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = make(map[string]any)
}

// Keys 获取所有键
func (b *ContextBag) Keys() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	keys := make([]string, 0, len(b.data))
	for key := range b.data {
		keys = append(keys, key)
	}
	return keys
}

// Size 获取大小
func (b *ContextBag) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.data)
}

// WithBag 在上下文中添加包
func WithBag(ctx context.Context, bag *ContextBag) context.Context {
	return SetValue(ctx, "__bag__", bag)
}

// GetBag 从上下文中获取包
func GetBag(ctx context.Context) *ContextBag {
	if bag, ok := ctx.Value(ContextKey("__bag__")).(*ContextBag); ok {
		return bag
	}
	return NewContextBag()
}

// ContextManager 上下文管理器
type ContextManager struct {
	mu       sync.RWMutex
	contexts map[string]*ManagedContext
}

// ManagedContext 托管上下文
type ManagedContext struct {
	ctx        context.Context
	cancel     context.CancelFunc
	created    time.Time
	metadata   *ContextMetadata
}

// NewContextManager 创建上下文管理器
func NewContextManager() *ContextManager {
	return &ContextManager{
		contexts: make(map[string]*ManagedContext),
	}
}

// Create 创建新的托管上下文
func (cm *ContextManager) Create(id string, parent context.Context) *ManagedContext {
	ctx, cancel := context.WithCancel(parent)

	mc := &ManagedContext{
		ctx:      ctx,
		cancel:   cancel,
		created:  time.Now(),
		metadata: NewContextMetadata(),
	}

	cm.mu.Lock()
	cm.contexts[id] = mc
	cm.mu.Unlock()

	return mc
}

// CreateWithTimeout 创建带超时的托管上下文
func (cm *ContextManager) CreateWithTimeout(id string, parent context.Context, timeout time.Duration) *ManagedContext {
	ctx, cancel := context.WithTimeout(parent, timeout)

	mc := &ManagedContext{
		ctx:      ctx,
		cancel:   cancel,
		created:  time.Now(),
		metadata: NewContextMetadata(),
	}

	cm.mu.Lock()
	cm.contexts[id] = mc
	cm.mu.Unlock()

	return mc
}

// Get 获取托管上下文
func (cm *ContextManager) Get(id string) (*ManagedContext, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	mc, ok := cm.contexts[id]
	return mc, ok
}

// Cancel 取消托管上下文
func (cm *ContextManager) Cancel(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if mc, ok := cm.contexts[id]; ok {
		mc.cancel()
		delete(cm.contexts, id)
	}
}

// Delete 删除托管上下文（不取消）
func (cm *ContextManager) Delete(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.contexts, id)
}

// Clear 清空所有托管上下文
func (cm *ContextManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 取消所有上下文
	for _, mc := range cm.contexts {
		mc.cancel()
	}

	cm.contexts = make(map[string]*ManagedContext)
}

// Count 获取托管上下文数量
func (cm *ContextManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.contexts)
}

// IDs 获取所有ID
func (cm *ContextManager) IDs() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ids := make([]string, 0, len(cm.contexts))
	for id := range cm.contexts {
		ids = append(ids, id)
	}
	return ids
}

// Context 返回context.Context
func (mc *ManagedContext) Context() context.Context {
	return mc.ctx
}

// Metadata 返回元数据
func (mc *ManagedContext) Metadata() *ContextMetadata {
	return mc.metadata
}

// Created 返回创建时间
func (mc *ManagedContext) Created() time.Time {
	return mc.created
}

// Done 返回done通道
func (mc *ManagedContext) Done() <-chan struct{} {
	return mc.ctx.Done()
}

// Err 返回错误
func (mc *ManagedContext) Err() error {
	return mc.ctx.Err()
}

// Deadline 返回截止时间
func (mc *ManagedContext) Deadline() (deadline time.Time, ok bool) {
	return mc.ctx.Deadline()
}

// Value 返回值
func (mc *ManagedContext) Value(key any) any {
	return mc.ctx.Value(key)
}

// CancelFunc 取消函数
func (mc *ManagedContext) CancelFunc() context.CancelFunc {
	return mc.cancel
}

// ContextPool 上下文池
type ContextPool struct {
	mu      sync.RWMutex
	context map[string]context.Context
	factory func() context.Context
}

// NewContextPool 创建上下文池
func NewContextPool(factory func() context.Context) *ContextPool {
	return &ContextPool{
		context: make(map[string]context.Context),
		factory: factory,
	}
}

// Get 获取上下文
func (cp *ContextPool) Get(key string) context.Context {
	cp.mu.RLock()
	if ctx, ok := cp.context[key]; ok {
		cp.mu.RUnlock()
		return ctx
	}
	cp.mu.RUnlock()

	cp.mu.Lock()
	defer cp.mu.Unlock()

	// 双重检查
	if ctx, ok := cp.context[key]; ok {
		return ctx
	}

	ctx := cp.factory()
	cp.context[key] = ctx
	return ctx
}

// Remove 移除上下文
func (cp *ContextPool) Remove(key string) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	delete(cp.context, key)
}

// Clear 清空上下文池
func (cp *ContextPool) Clear() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.context = make(map[string]context.Context)
}

// Size 获取大小
func (cp *ContextPool) Size() int {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	return len(cp.context)
}

// Keys 获取所有键
func (cp *ContextPool) Keys() []string {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	keys := make([]string, 0, len(cp.context))
	for key := range cp.context {
		keys = append(keys, key)
	}
	return keys
}

// ChainContextBuilder 链式上下文构建器
type ChainContextBuilder struct {
	ctx context.Context
}

// NewChainContextBuilder 创建链式上下文构建器
func NewChainContextBuilder(parent context.Context) *ChainContextBuilder {
	if parent == nil {
		parent = context.Background()
	}

	return &ChainContextBuilder{
		ctx: parent,
	}
}

// WithRequestID 添加请求ID
func (b *ChainContextBuilder) WithRequestID(requestID string) *ChainContextBuilder {
	b.ctx = WithRequestID(b.ctx, requestID)
	return b
}

// WithUserID 添加用户ID
func (b *ChainContextBuilder) WithUserID(userID string) *ChainContextBuilder {
	b.ctx = WithUserID(b.ctx, userID)
	return b
}

// WithTraceID 添加追踪ID
func (b *ChainContextBuilder) WithTraceID(traceID string) *ChainContextBuilder {
	b.ctx = WithTraceID(b.ctx, traceID)
	return b
}

// WithSessionID 添加会话ID
func (b *ChainContextBuilder) WithSessionID(sessionID string) *ChainContextBuilder {
	b.ctx = WithSessionID(b.ctx, sessionID)
	return b
}

// WithLanguage 添加语言
func (b *ChainContextBuilder) WithLanguage(language string) *ChainContextBuilder {
	b.ctx = WithLanguage(b.ctx, language)
	return b
}

// WithTimeZone 添加时区
func (b *ChainContextBuilder) WithTimeZone(timezone string) *ChainContextBuilder {
	b.ctx = WithTimeZone(b.ctx, timezone)
	return b
}

// WithIP 添加IP
func (b *ChainContextBuilder) WithIP(ip string) *ChainContextBuilder {
	b.ctx = WithIP(b.ctx, ip)
	return b
}

// WithValue 添加任意值
func (b *ChainContextBuilder) WithValue(key ContextKey, value any) *ChainContextBuilder {
	b.ctx = SetValue(b.ctx, key, value)
	return b
}

// WithTimeout 添加超时
func (b *ChainContextBuilder) WithTimeout(timeout time.Duration) (*ChainContextBuilder, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	b.ctx = ctx
	return b, cancel
}

// Build 构建上下文
func (b *ChainContextBuilder) Build() context.Context {
	return b.ctx
}
