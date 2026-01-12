package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// ConfigManager manages application configuration
type ConfigManager struct {
	configs    map[string]interface{}
	sources    []ConfigSource
	mu         sync.RWMutex
	watchFiles map[string]bool
}

// ConfigSource represents a configuration source
type ConfigSource interface {
	Load() (map[string]interface{}, error)
	Priority() int // Higher priority sources override lower priority ones
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		configs:    make(map[string]interface{}),
		sources:    make([]ConfigSource, 0),
		watchFiles: make(map[string]bool),
	}
}

// AddSource adds a configuration source
func (cm *ConfigManager) AddSource(source ConfigSource) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.sources = append(cm.sources, source)
}

// Load loads configuration from all sources
func (cm *ConfigManager) Load() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Sort sources by priority (higher first)
	sortedSources := make([]ConfigSource, len(cm.sources))
	copy(sortedSources, cm.sources)
	
	// Simple sort by priority
	for i := 0; i < len(sortedSources)-1; i++ {
		for j := i + 1; j < len(sortedSources); j++ {
			if sortedSources[i].Priority() < sortedSources[j].Priority() {
				sortedSources[i], sortedSources[j] = sortedSources[j], sortedSources[i]
			}
		}
	}

	// Load from all sources, higher priority overrides lower
	for _, source := range sortedSources {
		config, err := source.Load()
		if err != nil {
			return fmt.Errorf("failed to load config from source: %v", err)
		}

		// Merge configs (higher priority overrides)
		cm.mergeConfigs(cm.configs, config)
	}

	return nil
}

// mergeConfigs merges source into target
func (cm *ConfigManager) mergeConfigs(target, source map[string]interface{}) {
	for key, value := range source {
		if existing, exists := target[key]; exists {
			// If both are maps, merge recursively
			if targetMap, ok := existing.(map[string]interface{}); ok {
				if sourceMap, ok := value.(map[string]interface{}); ok {
					cm.mergeConfigs(targetMap, sourceMap)
					continue
				}
			}
		}
		// Override or set new value
		target[key] = value
	}
}

// Get retrieves a configuration value by key path (e.g., "server.port")
func (cm *ConfigManager) Get(key string) (interface{}, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	keys := strings.Split(key, ".")
	return cm.getNested(cm.configs, keys)
}

// GetString retrieves a string configuration value
func (cm *ConfigManager) GetString(key string, defaultValue string) string {
	value, err := cm.Get(key)
	if err != nil {
		return defaultValue
	}

	// Handle different types and convert to string
	switch v := value.(type) {
	case string:
		return v
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float64:
		// JSON numbers are parsed as float64
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt retrieves an integer configuration value
func (cm *ConfigManager) GetInt(key string, defaultValue int) int {
	value, err := cm.Get(key)
	if err != nil {
		return defaultValue
	}

	// Handle different numeric types
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	}

	return defaultValue
}

// GetBool retrieves a boolean configuration value
func (cm *ConfigManager) GetBool(key string, defaultValue bool) bool {
	value, err := cm.Get(key)
	if err != nil {
		return defaultValue
	}

	if b, ok := value.(bool); ok {
		return b
	}

	return defaultValue
}

// getNested retrieves a nested value from a map
func (cm *ConfigManager) getNested(m map[string]interface{}, keys []string) (interface{}, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("empty key path")
	}

	current := m
	for i, key := range keys[:len(keys)-1] {
		value, exists := current[key]
		if !exists {
			return nil, fmt.Errorf("config key '%s' not found", strings.Join(keys[:i+1], "."))
		}

		nested, ok := value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("config key '%s' is not a nested object", strings.Join(keys[:i+1], "."))
		}

		current = nested
	}

	lastKey := keys[len(keys)-1]
	value, exists := current[lastKey]
	if !exists {
		return nil, fmt.Errorf("config key '%s' not found", strings.Join(keys, "."))
	}

	return value, nil
}

// Set sets a configuration value
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	keys := strings.Split(key, ".")
	cm.setNested(cm.configs, keys, value)
}

// setNested sets a nested value in a map
func (cm *ConfigManager) setNested(m map[string]interface{}, keys []string, value interface{}) {
	current := m
	for _, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		current = current[key].(map[string]interface{})
	}

	lastKey := keys[len(keys)-1]
	current[lastKey] = value
}

// GetAll returns all configuration
func (cm *ConfigManager) GetAll() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Return a copy
	result := make(map[string]interface{})
	cm.deepCopy(cm.configs, result)
	return result
}

// deepCopy performs a deep copy of a map
func (cm *ConfigManager) deepCopy(src, dst map[string]interface{}) {
	for k, v := range src {
		if nested, ok := v.(map[string]interface{}); ok {
			dst[k] = make(map[string]interface{})
			cm.deepCopy(nested, dst[k].(map[string]interface{}))
		} else {
			dst[k] = v
		}
	}
}

// FileSource loads configuration from a JSON file
type FileSource struct {
	path     string
	priority int
}

// NewFileSource creates a new file source
func NewFileSource(path string, priority int) *FileSource {
	return &FileSource{
		path:     path,
		priority: priority,
	}
}

// Load loads configuration from file
func (fs *FileSource) Load() (map[string]interface{}, error) {
	data, err := os.ReadFile(fs.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return config, nil
}

// Priority returns the priority of this source
func (fs *FileSource) Priority() int {
	return fs.priority
}

// EnvSource loads configuration from environment variables
type EnvSource struct {
	prefix   string
	priority int
}

// NewEnvSource creates a new environment variable source
func NewEnvSource(prefix string, priority int) *EnvSource {
	return &EnvSource{
		prefix:   prefix,
		priority: priority,
	}
}

// Load loads configuration from environment variables
func (es *EnvSource) Load() (map[string]interface{}, error) {
	config := make(map[string]interface{})

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		value := parts[1]

		// Filter by prefix if specified
		if es.prefix != "" && !strings.HasPrefix(key, es.prefix) {
			continue
		}

		// Remove prefix and convert to nested keys
		if es.prefix != "" {
			key = strings.TrimPrefix(key, es.prefix)
		}

		// Convert KEY_NAME to nested map (e.g., SERVER_PORT -> server.port)
		keys := strings.Split(strings.ToLower(key), "_")
		es.setNested(config, keys, value)
	}

	return config, nil
}

// setNested sets a nested value in config map
func (es *EnvSource) setNested(m map[string]interface{}, keys []string, value interface{}) {
	current := m
	for _, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		current = current[key].(map[string]interface{})
	}

	lastKey := keys[len(keys)-1]
	current[lastKey] = value
}

// Priority returns the priority of this source
func (es *EnvSource) Priority() int {
	return es.priority
}

// LoadConfigFile loads a configuration file
func (cm *ConfigManager) LoadConfigFile(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", absPath)
	}

	// Clear existing sources to avoid duplicates
	cm.mu.Lock()
	cm.sources = make([]ConfigSource, 0)
	cm.mu.Unlock()

	source := NewFileSource(absPath, 10)
	cm.AddSource(source)
	return cm.Load()
}

// LoadConfigFileWithEnv loads a configuration file with environment variable substitution
func (cm *ConfigManager) LoadConfigFileWithEnv(path string, env string) error {
	// Try to load environment-specific config first
	envPath := strings.Replace(path, ".json", fmt.Sprintf("-%s.json", env), 1)
	if _, err := os.Stat(envPath); err == nil {
		if err := cm.LoadConfigFile(envPath); err != nil {
			return err
		}
	}

	// Load base config
	return cm.LoadConfigFile(path)
}
