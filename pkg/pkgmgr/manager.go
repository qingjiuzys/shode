package pkg

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/errors"
	"gitee.com/com_818cloud/shode/pkg/registry"
)

// PackageManager manages Shode package dependencies
type PackageManager struct {
	envManager     *environment.EnvironmentManager
	config         *PackageConfig
	configPath     string
	registryClient *registry.Client
}

// PackageConfig represents the shode.json configuration
type PackageConfig struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description,omitempty"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Scripts      map[string]string `json:"scripts,omitempty"`
}

// PackageInfo represents information about an installed package
type PackageInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
	Main        string `json:"main,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	Repository  string `json:"repository,omitempty"`
}

// NewPackageManager creates a new package manager instance.
//
// The package manager handles package initialization, dependency management,
// script execution, and registry operations. It automatically initializes
// a registry client with default configuration.
//
// Returns a new PackageManager instance ready to use.
//
// Example:
//
//	pm := pkg.NewPackageManager()
//	err := pm.Init("my-package", "1.0.0")
func NewPackageManager() *PackageManager {
	// Initialize registry client with default config
	registryClient, _ := registry.NewClient(nil)
	
	return &PackageManager{
		envManager:     environment.NewEnvironmentManager(),
		config:         &PackageConfig{},
		registryClient: registryClient,
	}
}

// Init initializes a new package configuration
func (pm *PackageManager) Init(name, version string) error {
	pm.config = &PackageConfig{
		Name:        name,
		Version:     version,
		Dependencies: make(map[string]string),
		DevDependencies: make(map[string]string),
		Scripts:     make(map[string]string),
	}

	// Set default config path
	wd := pm.envManager.GetWorkingDir()
	pm.configPath = filepath.Join(wd, "shode.json")

	return pm.SaveConfig()
}

// LoadConfig loads the package configuration from shode.json
func (pm *PackageManager) LoadConfig() error {
	wd := pm.envManager.GetWorkingDir()
	configPath := filepath.Join(wd, "shode.json")
	pm.configPath = configPath

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.NewFileNotFoundError(configPath).
			WithContext("message", "Run 'shode pkg init' first")
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return errors.WrapError(errors.ErrFileNotFound,
			"failed to read shode.json", err).
			WithContext("path", configPath)
	}

	if err := json.Unmarshal(data, &pm.config); err != nil {
		return errors.WrapError(errors.ErrParseError,
			"failed to parse shode.json", err).
			WithContext("path", configPath)
	}

	// Initialize maps if they are nil
	if pm.config.Dependencies == nil {
		pm.config.Dependencies = make(map[string]string)
	}
	if pm.config.DevDependencies == nil {
		pm.config.DevDependencies = make(map[string]string)
	}
	if pm.config.Scripts == nil {
		pm.config.Scripts = make(map[string]string)
	}

	return nil
}

// SaveConfig saves the package configuration to shode.json
func (pm *PackageManager) SaveConfig() error {
	if pm.configPath == "" {
		return errors.NewExecutionError(errors.ErrInvalidInput,
			"config path not set")
	}

	data, err := json.MarshalIndent(pm.config, "", "  ")
	if err != nil {
		return errors.WrapError(errors.ErrExecutionFailed,
			"failed to marshal config", err).
			WithContext("path", pm.configPath)
	}

	return ioutil.WriteFile(pm.configPath, data, 0644)
}

// AddDependency adds a package dependency
func (pm *PackageManager) AddDependency(name, version string, dev bool) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	if dev {
		pm.config.DevDependencies[name] = version
	} else {
		pm.config.Dependencies[name] = version
	}

	return pm.SaveConfig()
}

// RemoveDependency removes a package dependency
func (pm *PackageManager) RemoveDependency(name string, dev bool) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	if dev {
		delete(pm.config.DevDependencies, name)
	} else {
		delete(pm.config.Dependencies, name)
	}

	return pm.SaveConfig()
}

// AddScript adds a script to the configuration
func (pm *PackageManager) AddScript(name, command string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	pm.config.Scripts[name] = command
	return pm.SaveConfig()
}

// RemoveScript removes a script from the configuration
func (pm *PackageManager) RemoveScript(name string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	delete(pm.config.Scripts, name)
	return pm.SaveConfig()
}

// Install installs all dependencies
func (pm *PackageManager) Install() error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	fmt.Println("Installing dependencies...")

	// Create sh_models directory if it doesn't exist
	wd := pm.envManager.GetWorkingDir()
	shModelsPath := filepath.Join(wd, "sh_models")
	if err := os.MkdirAll(shModelsPath, 0755); err != nil {
		return errors.WrapError(errors.ErrFileNotFound,
			"failed to create sh_models directory", err).
			WithContext("path", shModelsPath)
	}

	// Install dependencies
	allDeps := make(map[string]string)
	for name, version := range pm.config.Dependencies {
		allDeps[name] = version
	}
	for name, version := range pm.config.DevDependencies {
		allDeps[name] = version
	}

	for name, version := range allDeps {
		fmt.Printf("Installing %s@%s\n", name, version)
		if err := pm.installPackageFromRegistry(name, version, shModelsPath); err != nil {
			// Fallback to local installation if registry fails
			fmt.Printf("  Registry installation failed, using local fallback...\n")
			if err := pm.installPackage(name, version); err != nil {
				return errors.WrapError(errors.ErrExecutionFailed,
					fmt.Sprintf("failed to install %s", name), err).
					WithContext("package", name).
					WithContext("version", version)
			}
		}
	}

	fmt.Println("All dependencies installed successfully!")
	return nil
}

// installPackageFromRegistry installs a package from the remote registry
func (pm *PackageManager) installPackageFromRegistry(name, version, targetDir string) error {
	// Try to install from registry
	if pm.registryClient == nil {
		return fmt.Errorf("registry client not initialized")
	}

	// Install package using registry client
	if err := pm.registryClient.Install(name, version, targetDir); err != nil {
		return err
	}

	fmt.Printf("  Installed %s@%s from registry\n", name, version)
	return nil
}

// installPackage installs a single package
func (pm *PackageManager) installPackage(name, version string) error {
	wd := pm.envManager.GetWorkingDir()

	// For now, we'll simulate package installation
	// In a real implementation, this would download from a registry
	packagePath := filepath.Join(wd, "sh_models", name)
	if err := os.MkdirAll(packagePath, 0755); err != nil {
		return err
	}

	// Create a simple package.json for the installed package
	packageInfo := PackageInfo{
		Name:    name,
		Version: version,
		Main:    "index.sh",
	}

	infoData, err := json.MarshalIndent(packageInfo, "", "  ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(packagePath, "package.json"), infoData, 0644); err != nil {
		return err
	}

	// Create a simple index.sh file
	indexContent := fmt.Sprintf(`#!/bin/sh
# %s v%s - Shode package
echo "Package %s version %s is installed"
`, name, version, name, version)

	if err := ioutil.WriteFile(filepath.Join(packagePath, "index.sh"), []byte(indexContent), 0755); err != nil {
		return err
	}

	return nil
}

// RunScript runs a script from the configuration
func (pm *PackageManager) RunScript(name string) error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	script, exists := pm.config.Scripts[name]
	if !exists {
		return fmt.Errorf("script '%s' not found in shode.json", name)
	}

	fmt.Printf("Running script: %s\n", script)
	fmt.Println("(Script execution will be implemented in the execution engine)")

	return nil
}

// ListDependencies lists all dependencies
func (pm *PackageManager) ListDependencies() error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	fmt.Println("Dependencies:")
	for name, version := range pm.config.Dependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	fmt.Println("\nDev Dependencies:")
	for name, version := range pm.config.DevDependencies {
		fmt.Printf("  %s: %s\n", name, version)
	}

	return nil
}

// GetConfig returns the current package configuration
func (pm *PackageManager) GetConfig() *PackageConfig {
	return pm.config
}

// GetConfigPath returns the path to the config file
func (pm *PackageManager) GetConfigPath() string {
	return pm.configPath
}

// Search searches for packages in the registry
func (pm *PackageManager) Search(query string) ([]*registry.SearchResult, error) {
	if pm.registryClient == nil {
		return nil, fmt.Errorf("registry client not initialized")
	}

	searchQuery := &registry.SearchQuery{
		Query: query,
		Limit: 20,
	}

	return pm.registryClient.Search(searchQuery)
}

// Publish publishes the current package to the registry
func (pm *PackageManager) Publish() error {
	if err := pm.LoadConfig(); err != nil {
		return err
	}

	// Create package data
	pkg := &registry.Package{
		Name:        pm.config.Name,
		Version:     pm.config.Version,
		Description: pm.config.Description,
		Scripts:     pm.config.Scripts,
		Dependencies: pm.config.Dependencies,
		DevDependencies: pm.config.DevDependencies,
		Main:        "index.sh",
	}

	// Create tarball from package files
	wd := pm.envManager.GetWorkingDir()
	tarballData, err := createTarball(wd)
	if err != nil {
		return fmt.Errorf("failed to create tarball: %v", err)
	}
	checksum := calculateChecksum(tarballData)

	req := &registry.PublishRequest{
		Package:  pkg,
		Tarball:  tarballData,
		Checksum: checksum,
	}

	return pm.registryClient.Publish(req)
}

// GetRegistryClient returns the registry client
func (pm *PackageManager) GetRegistryClient() *registry.Client {
	return pm.registryClient
}

// calculateChecksum calculates SHA256 checksum of data
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// createTarball creates a tar.gz archive from the package directory
func createTarball(sourceDir string) ([]byte, error) {
	var buf bytes.Buffer
	
	// Create gzip writer
	gzw := gzip.NewWriter(&buf)
	defer gzw.Close()
	
	// Create tar writer
	tw := tar.NewWriter(gzw)
	defer tw.Close()
	
	// Files and directories to exclude
	excludePatterns := []string{
		".git",
		"node_modules",
		"sh_models",
		".shode",
		"*.log",
		".DS_Store",
		"Thumbs.db",
	}
	
	// Walk the directory
	err := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if path == sourceDir {
			return nil
		}
		
		// Check if path should be excluded
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		
		// Check exclusion patterns
		for _, pattern := range excludePatterns {
			if matched, _ := filepath.Match(pattern, relPath); matched {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			// Check if path starts with pattern (for directories)
			if strings.HasPrefix(relPath, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		
		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		
		// Set the name to be relative to source directory
		header.Name = relPath
		
		// Write header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		
		// Write file content if it's a regular file
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Close writers to flush data
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}
