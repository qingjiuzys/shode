// Package plugin 提供插件系统功能。
package plugin

import (
	"context"
	"fmt"
	"sync"
)

// Plugin 插件接口
type Plugin interface {
	// ID 返回插件唯一标识
	ID() string

	// Name 返回插件名称
	Name() string

	// Version 返回插件版本
	Version() string

	// Initialize 初始化插件
	Initialize(ctx context.Context, config map[string]interface{}) error

	// Start 启动插件
	Start(ctx context.Context) error

	// Stop 停止插件
	Stop(ctx context.Context) error

	// Status 返回插件状态
	Status() PluginStatus

	// Dependencies 返回依赖的插件 ID
	Dependencies() []string
}

// PluginStatus 插件状态
type PluginStatus int

const (
	StatusUnloaded PluginStatus = iota
	StatusLoaded
	StatusInitialized
	StatusStarted
	StatusStopped
	StatusError
)

// String 返回状态字符串
func (s PluginStatus) String() string {
	switch s {
	case StatusUnloaded:
		return "unloaded"
	case StatusLoaded:
		return "loaded"
	case StatusInitialized:
		return "initialized"
	case StatusStarted:
		return "started"
	case StatusStopped:
		return "stopped"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Author      string
	Homepage    string
	Status      PluginStatus
	Config      map[string]interface{}
}

// PluginManager 插件管理器
type PluginManager struct {
	plugins    map[string]Plugin
	configs    map[string]map[string]interface{}
	status     map[string]PluginStatus
	mu         sync.RWMutex
	hooks      map[string][]HookFunc
	ctx        context.Context
	cancel     context.CancelFunc
}

// HookFunc 钩子函数
type HookFunc func(ctx context.Context, data interface{}) error

// NewPluginManager 创建插件管理器
func NewPluginManager() *PluginManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &PluginManager{
		plugins: make(map[string]Plugin),
		configs: make(map[string]map[string]interface{}),
		status:  make(map[string]PluginStatus),
		hooks:   make(map[string][]HookFunc),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Register 注册插件
func (pm *PluginManager) Register(plugin Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pluginID := plugin.ID()

	if _, exists := pm.plugins[pluginID]; exists {
		return fmt.Errorf("plugin already registered: %s", pluginID)
	}

	pm.plugins[pluginID] = plugin
	pm.status[pluginID] = StatusLoaded

	fmt.Printf("Plugin registered: %s (%s)\n", plugin.Name(), pluginID)

	return nil
}

// Unregister 注销插件
func (pm *PluginManager) Unregister(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 如果插件已启动，先停止
	if pm.status[pluginID] == StatusStarted {
		if err := plugin.Stop(pm.ctx); err != nil {
			return fmt.Errorf("failed to stop plugin: %w", err)
		}
	}

	delete(pm.plugins, pluginID)
	delete(pm.configs, pluginID)
	delete(pm.status, pluginID)

	fmt.Printf("Plugin unregistered: %s\n", pluginID)

	return nil
}

// Initialize 初始化插件
func (pm *PluginManager) Initialize(pluginID string, config map[string]interface{}) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 保存配置
	if config != nil {
		pm.configs[pluginID] = config
	}

	// 初始化插件
	if err := plugin.Initialize(pm.ctx, config); err != nil {
		pm.status[pluginID] = StatusError
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	pm.status[pluginID] = StatusInitialized

	fmt.Printf("Plugin initialized: %s\n", pluginID)

	return nil
}

// Start 启动插件
func (pm *PluginManager) Start(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 检查依赖
	deps := plugin.Dependencies()
	for _, depID := range deps {
		if depStatus, exists := pm.status[depID]; !exists || depStatus != StatusStarted {
			return fmt.Errorf("dependency not satisfied: %s (requires %s)", pluginID, depID)
		}
	}

	// 启动插件
	if err := plugin.Start(pm.ctx); err != nil {
		pm.status[pluginID] = StatusError
		return fmt.Errorf("failed to start plugin: %w", err)
	}

	pm.status[pluginID] = StatusStarted

	fmt.Printf("Plugin started: %s\n", pluginID)

	// 触发钩子
	pm.triggerHook("plugin.started", map[string]interface{}{
		"plugin_id": pluginID,
		"plugin":    plugin,
	})

	return nil
}

// Stop 停止插件
func (pm *PluginManager) Stop(pluginID string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	// 检查是否有其他插件依赖此插件
	for id, p := range pm.plugins {
		if id == pluginID {
			continue
		}
		for _, dep := range p.Dependencies() {
			if dep == pluginID && pm.status[id] == StatusStarted {
				return fmt.Errorf("cannot stop plugin: %s is required by %s", pluginID, id)
			}
		}
	}

	// 停止插件
	if err := plugin.Stop(pm.ctx); err != nil {
		pm.status[pluginID] = StatusError
		return fmt.Errorf("failed to stop plugin: %w", err)
	}

	pm.status[pluginID] = StatusStopped

	fmt.Printf("Plugin stopped: %s\n", pluginID)

	// 触发钩子
	pm.triggerHook("plugin.stopped", map[string]interface{}{
		"plugin_id": pluginID,
		"plugin":    plugin,
	})

	return nil
}

// Get 获取插件
func (pm *PluginManager) Get(pluginID string) (Plugin, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	return plugin, exists
}

// List 列出所有插件
func (pm *PluginManager) List() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// GetStatus 获取插件状态
func (pm *PluginManager) GetStatus(pluginID string) PluginStatus {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	status, exists := pm.status[pluginID]
	if !exists {
		return StatusUnloaded
	}

	return status
}

// GetInfo 获取插件信息
func (pm *PluginManager) GetInfo(pluginID string) (*PluginInfo, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	config := pm.configs[pluginID]
	if config == nil {
		config = make(map[string]interface{})
	}

	return &PluginInfo{
		ID:      plugin.ID(),
		Name:    plugin.Name(),
		Version: plugin.Version(),
		Status:  pm.status[pluginID],
		Config:  config,
	}, nil
}

// StartAll 启动所有插件
func (pm *PluginManager) StartAll() error {
	pm.mu.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mu.RUnlock()

	// 拓扑排序启动（按依赖关系）
	for _, plugin := range plugins {
		if pm.GetStatus(plugin.ID()) != StatusStarted {
			if err := pm.Start(plugin.ID()); err != nil {
				return err
			}
		}
	}

	return nil
}

// StopAll 停止所有插件
func (pm *PluginManager) StopAll() error {
	pm.mu.RLock()
	plugins := make([]Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mu.RUnlock()

	// 逆序停止
	for i := len(plugins) - 1; i >= 0; i-- {
		plugin := plugins[i]
		if pm.GetStatus(plugin.ID()) == StatusStarted {
			if err := pm.Stop(plugin.ID()); err != nil {
				fmt.Printf("Failed to stop plugin %s: %v\n", plugin.ID(), err)
			}
		}
	}

	return nil
}

// RegisterHook 注册钩子
func (pm *PluginManager) RegisterHook(name string, fn HookFunc) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.hooks[name] = append(pm.hooks[name], fn)
}

// triggerHook 触发钩子
func (pm *PluginManager) triggerHook(name string, data interface{}) {
	pm.mu.RLock()
	hooks, exists := pm.hooks[name]
	pm.mu.RUnlock()

	if !exists {
		return
	}

	for _, hook := range hooks {
		go func(fn HookFunc) {
			if err := fn(pm.ctx, data); err != nil {
				fmt.Printf("Hook %s failed: %v\n", name, err)
			}
		}(hook)
	}
}

// Communicate 插件间通信
func (pm *PluginManager) Communicate(from, to string, data interface{}) (interface{}, error) {
	pm.mu.RLock()
	plugin, exists := pm.plugins[to]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", to)
	}

	// 如果插件实现了 CommunicationHandler 接口，调用它
	if handler, ok := plugin.(CommunicationHandler); ok {
		return handler.HandleMessage(pm.ctx, from, data)
	}

	return nil, fmt.Errorf("plugin does not support communication")
}

// CommunicationHandler 通信处理器接口
type CommunicationHandler interface {
	HandleMessage(ctx context.Context, from string, data interface{}) (interface{}, error)
}

// BasePlugin 基础插件
type BasePlugin struct {
	id      string
	name    string
	version string
	status  PluginStatus
}

// NewBasePlugin 创建基础插件
func NewBasePlugin(id, name, version string) *BasePlugin {
	return &BasePlugin{
		id:      id,
		name:    name,
		version: version,
		status:  StatusUnloaded,
	}
}

// ID 返回插件 ID
func (bp *BasePlugin) ID() string {
	return bp.id
}

// Name 返回插件名称
func (bp *BasePlugin) Name() string {
	return bp.name
}

// Version 返回插件版本
func (bp *BasePlugin) Version() string {
	return bp.version
}

// Initialize 初始化插件
func (bp *BasePlugin) Initialize(ctx context.Context, config map[string]interface{}) error {
	bp.status = StatusInitialized
	return nil
}

// Start 启动插件
func (bp *BasePlugin) Start(ctx context.Context) error {
	bp.status = StatusStarted
	return nil
}

// Stop 停止插件
func (bp *BasePlugin) Stop(ctx context.Context) error {
	bp.status = StatusStopped
	return nil
}

// Status 返回插件状态
func (bp *BasePlugin) Status() PluginStatus {
	return bp.status
}

// Dependencies 返回依赖
func (bp *BasePlugin) Dependencies() []string {
	return []string{}
}
