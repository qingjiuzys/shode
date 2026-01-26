package module

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/errors"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// ModuleManager manages Shode module loading and resolution
type ModuleManager struct {
	envManager *environment.EnvironmentManager
	parser     *parser.SimpleParser
	modules    map[string]*Module
}

// Module represents a loaded Shode module
type Module struct {
	Name     string
	Path     string
	Exports  map[string]*types.CommandNode
	Imports  map[string]*Module
	IsLoaded bool
}

// ModuleInfo contains information about a module
type ModuleInfo struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Main        string            `json:"main,omitempty"`
	Exports     map[string]string `json:"exports,omitempty"`
}

// NewModuleManager creates a new module manager instance.
//
// The module manager handles module loading, resolution, and export/import
// functionality. It automatically initializes an environment manager and
// parser for module operations.
//
// Returns a new ModuleManager instance ready to use.
//
// Example:
//
//	mm := module.NewModuleManager()
//	mod, err := mm.LoadModule("./my-module")
func NewModuleManager() *ModuleManager {
	return &ModuleManager{
		envManager: environment.NewEnvironmentManager(),
		parser:     parser.NewSimpleParser(),
		modules:    make(map[string]*Module),
	}
}

// LoadModule loads a module from the given path
func (mm *ModuleManager) LoadModule(path string) (*Module, error) {
	// Check if module is already loaded
	if module, exists := mm.modules[path]; exists && module.IsLoaded {
		return module, nil
	}

	// Resolve absolute path
	absPath, err := mm.resolveModulePath(path)
	if err != nil {
		return nil, err
	}

	// Check if module exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("module not found: %s", path)
	}

	// Try to get module name from package.json first
	moduleName := filepath.Base(absPath)
	packageJsonPath := filepath.Join(absPath, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		if pkgInfo, err := mm.loadPackageJson(packageJsonPath); err == nil && pkgInfo.Name != "" {
			moduleName = pkgInfo.Name
		}
	}

	// Create new module
	module := &Module{
		Name:     moduleName,
		Path:     absPath,
		Exports:  make(map[string]*types.CommandNode),
		Imports:  make(map[string]*Module),
		IsLoaded: false,
	}

	// Load module exports
	if err := mm.loadModuleExports(module); err != nil {
		return nil, err
	}

	// Mark as loaded and store
	module.IsLoaded = true
	mm.modules[path] = module

	return module, nil
}

// resolveModulePath resolves a module path to an absolute path
func (mm *ModuleManager) resolveModulePath(path string) (string, error) {
	// Handle relative paths
	if !filepath.IsAbs(path) {
		wd := mm.envManager.GetWorkingDir()

		// Check if it's a local file
		localPath := filepath.Join(wd, path)
		if _, err := os.Stat(localPath); err == nil {
			return localPath, nil
		}

		// Check sh_models
		shModelsPath := filepath.Join(wd, "sh_models", path)
		if _, err := os.Stat(shModelsPath); err == nil {
			return shModelsPath, nil
		}

		return "", fmt.Errorf("module not found: %s", path)
	}

	return path, nil
}

// loadModuleExports loads exports from a module
func (mm *ModuleManager) loadModuleExports(module *Module) error {
	// Check for package.json first
	packageJsonPath := filepath.Join(module.Path, "package.json")
	if _, err := os.Stat(packageJsonPath); err == nil {
		// Load package.json
		pkgInfo, err := mm.loadPackageJson(packageJsonPath)
		if err != nil {
			return fmt.Errorf("failed to load package.json: %v", err)
		}

		// Use main entry point from package.json if specified
		if pkgInfo.Main != "" {
			mainPath := filepath.Join(module.Path, pkgInfo.Main)
			if _, err := os.Stat(mainPath); err == nil {
				return mm.loadScriptExports(module, mainPath)
			}
			// If main path doesn't exist, fall through to default behavior
		}
	}

	// Look for index.sh (default entry point)
	indexPath := filepath.Join(module.Path, "index.sh")
	if _, err := os.Stat(indexPath); err == nil {
		return mm.loadScriptExports(module, indexPath)
	}

	// Look for <module-name>.sh
	moduleScriptPath := filepath.Join(module.Path, module.Name+".sh")
	if _, err := os.Stat(moduleScriptPath); err == nil {
		return mm.loadScriptExports(module, moduleScriptPath)
	}

	return errors.NewExecutionError(errors.ErrFileNotFound,
		fmt.Sprintf("no module entry point found in %s", module.Path)).
		WithContext("module_path", module.Path)
}

// PackageJson represents a package.json structure
type PackageJson struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description,omitempty"`
	Main        string            `json:"main,omitempty"`
	Exports     map[string]string `json:"exports,omitempty"`
	Scripts     map[string]string `json:"scripts,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
}

// loadPackageJson loads and parses a package.json file
func (mm *ModuleManager) loadPackageJson(path string) (*PackageJson, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var pkg PackageJson
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse package.json: %v", err)
	}

	// Set default main if not specified
	if pkg.Main == "" {
		pkg.Main = "index.sh"
	}

	return &pkg, nil
}

// loadScriptExports loads exports from a script file
func (mm *ModuleManager) loadScriptExports(module *Module, scriptPath string) error {
	// Read script content
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read module script: %v", err)
	}

	// Parse script
	script, err := mm.parser.ParseString(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse module script: %v", err)
	}

	// Extract exports (functions starting with export_)
	for _, node := range script.Nodes {
		if cmdNode, ok := node.(*types.CommandNode); ok {
			if strings.HasPrefix(cmdNode.Name, "export_") {
				exportName := strings.TrimPrefix(cmdNode.Name, "export_")
				module.Exports[exportName] = cmdNode
			}
		}
	}

	return nil
}

// Import imports a module and returns its exports
func (mm *ModuleManager) Import(path string) (map[string]*types.CommandNode, error) {
	module, err := mm.LoadModule(path)
	if err != nil {
		return nil, err
	}

	return module.Exports, nil
}

// GetModule returns a loaded module by path
func (mm *ModuleManager) GetModule(path string) (*Module, error) {
	module, exists := mm.modules[path]
	if !exists || !module.IsLoaded {
		return nil, fmt.Errorf("module not loaded: %s", path)
	}
	return module, nil
}

// ListModules returns all loaded modules
func (mm *ModuleManager) ListModules() []*Module {
	var modules []*Module
	for _, module := range mm.modules {
		if module.IsLoaded {
			modules = append(modules, module)
		}
	}
	return modules
}

// UnloadModule unloads a module
func (mm *ModuleManager) UnloadModule(path string) error {
	if _, exists := mm.modules[path]; !exists {
		return fmt.Errorf("module not found: %s", path)
	}
	delete(mm.modules, path)
	return nil
}

// ClearModules unloads all modules
func (mm *ModuleManager) ClearModules() {
	mm.modules = make(map[string]*Module)
}

// ResolveImport resolves an import statement
func (mm *ModuleManager) ResolveImport(importPath string) (string, error) {
	return mm.resolveModulePath(importPath)
}

// GetExport gets a specific export from a module
func (mm *ModuleManager) GetExport(modulePath, exportName string) (*types.CommandNode, error) {
	module, err := mm.GetModule(modulePath)
	if err != nil {
		return nil, err
	}

	// Try exact match first
	export, exists := module.Exports[exportName]
	if exists {
		return export, nil
	}

	// Try with parentheses for function-style exports
	export, exists = module.Exports[exportName+"()"]
	if exists {
		return export, nil
	}

	return nil, fmt.Errorf("export %s not found in module %s", exportName, modulePath)
}

// HasExport checks if a module has a specific export
func (mm *ModuleManager) HasExport(modulePath, exportName string) (bool, error) {
	module, err := mm.GetModule(modulePath)
	if err != nil {
		return false, err
	}

	// Try exact match first
	_, exists := module.Exports[exportName]
	if exists {
		return true, nil
	}

	// Try with parentheses for function-style exports
	_, exists = module.Exports[exportName+"()"]
	return exists, nil
}

// GetModuleInfo gets information about a module
func (mm *ModuleManager) GetModuleInfo(path string) (*ModuleInfo, error) {
	module, err := mm.GetModule(path)
	if err != nil {
		return nil, err
	}

	info := &ModuleInfo{
		Name:    module.Name,
		Exports: make(map[string]string),
	}

	// Collect export names
	for exportName := range module.Exports {
		info.Exports[exportName] = "function"
	}

	return info, nil
}

// IsExportedFunction checks if a function name is exported by any loaded module
// This is used by the execution engine to determine if a command should be
// executed in interpreted mode (for module exports) or process mode
func (mm *ModuleManager) IsExportedFunction(funcName string) bool {
	// Check all loaded modules for this export
	for _, module := range mm.modules {
		if !module.IsLoaded {
			continue
		}

		// Try exact match
		if _, exists := module.Exports[funcName]; exists {
			return true
		}

		// Try with parentheses for function-style exports
		if _, exists := module.Exports[funcName+"()"]; exists {
			return true
		}
	}

	return false
}
