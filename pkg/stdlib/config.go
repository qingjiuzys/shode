package stdlib

import (
	"fmt"
)

// LoadConfig loads configuration from a file
// Usage: LoadConfig "application.json"
func (sl *StdLib) LoadConfig(path string) error {
	return sl.configManager.LoadConfigFile(path)
}

// LoadConfigWithEnv loads configuration with environment substitution
// Usage: LoadConfigWithEnv "application.json" "prod"
func (sl *StdLib) LoadConfigWithEnv(path, env string) error {
	return sl.configManager.LoadConfigFileWithEnv(path, env)
}

// GetConfig retrieves a configuration value
// Usage: GetConfig "server.port"
func (sl *StdLib) GetConfig(key string) (string, error) {
	value, err := sl.configManager.Get(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", value), nil
}

// GetConfigString retrieves a string configuration value with default
// Usage: GetConfigString "server.port" "9188"
func (sl *StdLib) GetConfigString(key, defaultValue string) string {
	return sl.configManager.GetString(key, defaultValue)
}

// GetConfigInt retrieves an integer configuration value with default
// Usage: GetConfigInt "server.port" 9188
func (sl *StdLib) GetConfigInt(key string, defaultValue int) int {
	return sl.configManager.GetInt(key, defaultValue)
}

// GetConfigBool retrieves a boolean configuration value with default
// Usage: GetConfigBool "server.enabled" true
func (sl *StdLib) GetConfigBool(key string, defaultValue bool) bool {
	return sl.configManager.GetBool(key, defaultValue)
}

// SetConfig sets a configuration value
// Usage: SetConfig "server.port" "9188"
func (sl *StdLib) SetConfig(key string, value interface{}) {
	sl.configManager.Set(key, value)
}
