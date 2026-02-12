// Package configcenter 提供配置中心功能。
package configcenter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ConfigItem 配置项
type ConfigItem struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Version     int64                  `json:"version"`
	ContentType string                 `json:"content_type"`
	Labels      map[string]string      `json:"labels"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Encrypted   bool                   `json:"encrypted"`
}

// ConfigChange 配置变更
type ConfigChange struct {
	Type     string // "create", "update", "delete"
	Key      string
	OldValue interface{}
	NewValue interface{}
	Version  int64
	Timestamp time.Time
}

// ConfigStore 配置存储接口
type ConfigStore interface {
	Get(ctx context.Context, key string) (*ConfigItem, error)
	Set(ctx context.Context, item *ConfigItem) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]*ConfigItem, error)
	Watch(ctx context.Context, key string) (<-chan *ConfigChange, error)
}

// MemoryConfigStore 内存配置存储
type MemoryConfigStore struct {
	items    map[string]*ConfigItem
	watchers map[string][]chan *ConfigChange
	mu       sync.RWMutex
}

// NewMemoryConfigStore 创建内存配置存储
func NewMemoryConfigStore() *MemoryConfigStore {
	return &MemoryConfigStore{
		items:    make(map[string]*ConfigItem),
		watchers: make(map[string][]chan *ConfigChange),
	}
}

// Get 获取配置
func (mcs *MemoryConfigStore) Get(ctx context.Context, key string) (*ConfigItem, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	item, exists := mcs.items[key]
	if !exists {
		return nil, fmt.Errorf("config not found: %s", key)
	}

	return item, nil
}

// Set 设置配置
func (mcs *MemoryConfigStore) Set(ctx context.Context, item *ConfigItem) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	oldValue := mcs.items[item.Key]
	oldVersion := int64(0)
	if oldValue != nil {
		oldVersion = oldValue.Version
	}

	item.Version = oldVersion + 1
	item.UpdatedAt = time.Now()

	mcs.items[item.Key] = item

	// 通知观察者
	change := &ConfigChange{
		Type:     "update",
		Key:      item.Key,
		OldValue: nil,
		NewValue: item.Value,
		Version:  item.Version,
		Timestamp: time.Now(),
	}

	if oldValue == nil {
		change.Type = "create"
	}

	mcs.notifyWatchers(item.Key, change)

	return nil
}

// Delete 删除配置
func (mcs *MemoryConfigStore) Delete(ctx context.Context, key string) error {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	if _, exists := mcs.items[key]; !exists {
		return fmt.Errorf("config not found: %s", key)
	}

	delete(mcs.items, key)

	change := &ConfigChange{
		Type:     "delete",
		Key:      key,
		Version:  0,
		Timestamp: time.Now(),
	}

	mcs.notifyWatchers(key, change)

	return nil
}

// List 列出配置
func (mcs *MemoryConfigStore) List(ctx context.Context, prefix string) ([]*ConfigItem, error) {
	mcs.mu.RLock()
	defer mcs.mu.RUnlock()

	items := make([]*ConfigItem, 0)
	for key, item := range mcs.items {
		if len(prefix) == 0 || key == prefix || len(key) > len(prefix) && key[:len(prefix)] == prefix {
			items = append(items, item)
		}
	}

	return items, nil
}

// Watch 监听配置变更
func (mcs *MemoryConfigStore) Watch(ctx context.Context, key string) (<-chan *ConfigChange, error) {
	mcs.mu.Lock()
	defer mcs.mu.Unlock()

	ch := make(chan *ConfigChange, 100)

	if _, exists := mcs.watchers[key]; !exists {
		mcs.watchers[key] = make([]chan *ConfigChange, 0)
	}
	mcs.watchers[key] = append(mcs.watchers[key], ch)

	return ch, nil
}

// notifyWatchers 通知观察者
func (mcs *MemoryConfigStore) notifyWatchers(key string, change *ConfigChange) {
	if watchers, exists := mcs.watchers[key]; exists {
		for _, ch := range watchers {
			select {
			case ch <- change:
			default:
				// 通道满，跳过
			}
		}
	}
}

// ConfigCenter 配置中心
type ConfigCenter struct {
	store      ConfigStore
	cache      *ConfigCache
	encryptor  *Encryptor
	version    *VersionManager
	mu         sync.RWMutex
}

// NewConfigCenter 创建配置中心
func NewConfigCenter(store ConfigStore) *ConfigCenter {
	return &ConfigCenter{
		store:     store,
		cache:     NewConfigCache(),
		encryptor:  NewEncryptor(),
		version:   NewVersionManager(),
	}
}

// Get 获取配置
func (cc *ConfigCenter) Get(ctx context.Context, key string) (interface{}, error) {
	// 先查缓存
	if value, exists := cc.cache.Get(key); exists {
		return value, nil
	}

	// 从存储加载
	item, err := cc.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// 解密
	value := item.Value
	if item.Encrypted {
		value, err = cc.encryptor.Decrypt(value.(string))
		if err != nil {
			return nil, err
		}
	}

	// 缓存
	cc.cache.Set(key, value, 5*time.Minute)

	return value, nil
}

// Set 设置配置
func (cc *ConfigCenter) Set(ctx context.Context, key string, value interface{}, encrypt bool) error {
	storedValue := value

	// 加密
	if encrypt {
		encrypted, err := cc.encryptor.Encrypt(fmt.Sprintf("%v", value))
		if err != nil {
			return err
		}
		storedValue = encrypted
	}

	item := &ConfigItem{
		Key:         key,
		Value:       storedValue,
		ContentType: "application/json",
		Labels:      make(map[string]string),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Encrypted:   encrypt,
	}

	return cc.store.Set(ctx, item)
}

// Delete 删除配置
func (cc *ConfigCenter) Delete(ctx context.Context, key string) error {
	// 清除缓存
	cc.cache.Delete(key)

	return cc.store.Delete(ctx, key)
}

// List 列出配置
func (cc *ConfigCenter) List(ctx context.Context, prefix string) (map[string]interface{}, error) {
	items, err := cc.store.List(ctx, prefix)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, item := range items {
		value := item.Value
		if item.Encrypted {
			decrypted, err := cc.encryptor.Decrypt(value.(string))
			if err != nil {
				return nil, err
			}
			value = decrypted
		}
		result[item.Key] = value
	}

	return result, nil
}

// Watch 监听配置变更
func (cc *ConfigCenter) Watch(ctx context.Context, key string) (<-chan *ConfigChange, error) {
	return cc.store.Watch(ctx, key)
}

// Publish 发布配置
func (cc *ConfigCenter) Publish(ctx context.Context, key string, value interface{}) error {
	return cc.Set(ctx, key, value, false)
}

// Subscribe 订阅配置
func (cc *ConfigCenter) Subscribe(ctx context.Context, pattern string, handler func(*ConfigChange)) error {
	ch, err := cc.store.Watch(ctx, pattern)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case change, ok := <-ch:
				if !ok {
					return
				}
				handler(change)
			}
		}
	}()

	return nil
}

// ConfigCache 配置缓存
type ConfigCache struct {
	items map[string]*cacheEntry
	mu    sync.RWMutex
}

// cacheEntry 缓存条目
type cacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// NewConfigCache 创建配置缓存
func NewConfigCache() *ConfigCache {
	return &ConfigCache{
		items: make(map[string]*cacheEntry),
	}
}

// Get 获取
func (cc *ConfigCache) Get(key string) (interface{}, bool) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	entry, exists := cc.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		delete(cc.items, key)
		return nil, false
	}

	return entry.Value, true
}

// Set 设置
func (cc *ConfigCache) Set(key string, value interface{}, ttl time.Duration) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.items[key] = &cacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete 删除
func (cc *ConfigCache) Delete(key string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.items, key)
}

// Clear 清空
func (cc *ConfigCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.items = make(map[string]*cacheEntry)
}

// Encryptor 加密器
type Encryptor struct {
	key []byte
}

// NewEncryptor 创建加密器
func NewEncryptor() *Encryptor {
	return &Encryptor{
		key: []byte("default-key-32-bytes-long-!"),
	}
}

// Encrypt 加密
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	// 简化实现，实际应该用 AES
	return fmt.Sprintf("encrypted:%s", plaintext), nil
}

// Decrypt 解密
func (e *Encryptor) Decrypt(ciphertext string) (interface{}, error) {
	// 简化实现
	if strings.HasPrefix(ciphertext, "encrypted:") {
		return strings.TrimPrefix(ciphertext, "encrypted:"), nil
	}
	return nil, fmt.Errorf("invalid ciphertext")
}

// VersionManager 版本管理器
type VersionManager struct {
	versions map[string][]*ConfigVersion
	mu       sync.RWMutex
}

// ConfigVersion 配置版本
type ConfigVersion struct {
	Version   string
	Value     interface{}
	CreatedAt time.Time
	Active    bool
}

// NewVersionManager 创建版本管理器
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions: make(map[string][]*ConfigVersion),
	}
}

// SaveVersion 保存版本
func (vm *VersionManager) SaveVersion(key string, version string, value interface{}) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if _, exists := vm.versions[key]; !exists {
		vm.versions[key] = make([]*ConfigVersion, 0)
	}

	configVersion := &ConfigVersion{
		Version:   version,
		Value:     value,
		CreatedAt: time.Now(),
		Active:    true,
	}

	vm.versions[key] = append(vm.versions[key], configVersion)
	return nil
}

// GetVersion 获取版本
func (vm *VersionManager) GetVersion(key, version string) (interface{}, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	for _, v := range versions {
		if v.Version == version {
			return v.Value, nil
		}
	}

	return nil, fmt.Errorf("version not found: %s", version)
}

// ListVersions 列出版本
func (vm *VersionManager) ListVersions(key string) ([]*ConfigVersion, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	versions, exists := vm.versions[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return versions, nil
}

// Rollback 回滚版本
func (vm *VersionManager) Rollback(ctx context.Context, key string, version string) error {
	// 简化实现，获取版本并设置
	value, err := vm.GetVersion(key, version)
	if err != nil {
		return err
	}

	// 保存为当前版本
	currentVersion := fmt.Sprintf("v%d", time.Now().Unix())
	return vm.SaveVersion(key, currentVersion, value)
}

// GrayRelease 灰度发布
type GrayRelease struct {
	configs map[string]*GrayConfig
	mu      sync.RWMutex
}

// GrayConfig 灰度配置
type GrayConfig struct {
	Key         string
	Version     string
	Percentage  int // 流量百分比
	Conditions  map[string]string
	CreatedAt   time.Time
}

// NewGrayRelease 创建灰度发布
func NewGrayRelease() *GrayRelease {
	return &GrayRelease{
		configs: make(map[string]*GrayConfig),
	}
}

// Publish 发布灰度配置
func (gr *GrayRelease) Publish(key, version string, percentage int, conditions map[string]string) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	config := &GrayConfig{
		Key:        key,
		Version:    version,
		Percentage: percentage,
		Conditions: conditions,
		CreatedAt:  time.Now(),
	}

	gr.configs[key+"."+version] = config
	return nil
}

// GetConfig 获取配置（灰度）
func (gr *GrayRelease) GetConfig(key string, labels map[string]string) (string, error) {
	gr.mu.RLock()
	defer gr.mu.RUnlock()

	// 简化实现，返回第一个匹配的配置
	for _, config := range gr.configs {
		if strings.HasPrefix(config.Key, key) {
			return config.Version, nil
		}
	}

	return "", fmt.Errorf("config not found: %s", key)
}

// UpdatePercentage 更新百分比
func (gr *GrayRelease) UpdatePercentage(key, version string, percentage int) error {
	gr.mu.Lock()
	defer gr.mu.Unlock()

	configKey := key + "." + version
	if config, exists := gr.configs[configKey]; exists {
		config.Percentage = percentage
		return nil
	}

	return fmt.Errorf("gray config not found: %s", configKey)
}

// PushConfig 推送配置
func (cc *ConfigCenter) PushConfig(ctx context.Context, appID string, config map[string]interface{}) error {
	// 批量设置配置
	for key, value := range config {
		fullKey := fmt.Sprintf("%s.%s", appID, key)
		if err := cc.Set(ctx, fullKey, value, false); err != nil {
			return err
		}
	}
	return nil
}

// PullConfig 拉取配置
func (cc *ConfigCenter) PullConfig(ctx context.Context, appID string) (map[string]interface{}, error) {
	return cc.List(ctx, appID)
}

// Validate 验证配置
func (cc *ConfigCenter) Validate(key string, value interface{}) error {
	// 简化实现，检查 JSON 格式
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return fmt.Errorf("invalid JSON structure: %w", err)
	}

	return nil
}

// Template 模板
type Template struct {
	Name       string
	Content    string
	Variables  map[string]interface{}
}

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*Template
	mu        sync.RWMutex
}

// NewTemplateManager 创建模板管理器
func NewTemplateManager() *TemplateManager {
	return &TemplateManager{
		templates: make(map[string]*Template),
	}
}

// AddTemplate 添加模板
func (tm *TemplateManager) AddTemplate(template *Template) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.templates[template.Name] = template
}

// Render 渲染模板
func (tm *TemplateManager) Render(name string, variables map[string]interface{}) (string, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	template, exists := tm.templates[name]
	if !exists {
		return "", fmt.Errorf("template not found: %s", name)
	}

	content := template.Content
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}

	return content, nil
}
