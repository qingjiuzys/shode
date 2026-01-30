package pkg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"gitee.com/com_818cloud/shode/pkg/semver"
)

// LockFileManager manages shode-lock.json
type LockFileManager struct {
	configPath string
}

// LockFile represents the shode-lock.json structure
type LockFile struct {
	LockfileVersion int                    `json:"lockfileVersion"`
	GeneratedAt     time.Time              `json:"generatedAt"`
	Resolved        map[string]*LockEntry  `json:"resolved"`
	Dependencies    map[string]*DepEntry   `json:"dependencies"`
	DevDependencies map[string]*DepEntry   `json:"devDependencies"`
}

// LockEntry represents a locked package entry
type LockEntry struct {
	Version      string            `json:"version"`
	Integrity    string            `json:"integrity"`
	Resolved     string            `json:"resolved"`
	Dependencies map[string]string `json:"dependencies"`
}

// DepEntry represents a dependency entry in lock file
type DepEntry struct {
	Version  string   `json:"version"`
	Requires []string `json:"requires"`
}

// NewLockFileManager creates a new lock file manager
func NewLockFileManager(configPath string) *LockFileManager {
	return &LockFileManager{
		configPath: configPath,
	}
}

// Load loads the lock file from disk
func (lfm *LockFileManager) Load() (*LockFile, error) {
	// Get lock file path
	lockfilePath := lfm.getLockFilePath()

	// Check if lock file exists
	if _, err := os.Stat(lockfilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("lock file not found: %s", lockfilePath)
	}

	data, err := ioutil.ReadFile(lockfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read lock file: %v", err)
	}

	var lockfile LockFile
	if err := json.Unmarshal(data, &lockfile); err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %v", err)
	}

	// Validate lock file version
	if lockfile.LockfileVersion != 1 {
		return nil, fmt.Errorf("unsupported lock file version: %d", lockfile.LockfileVersion)
	}

	return &lockfile, nil
}

// Save saves the lock file to disk
func (lfm *LockFileManager) Save(lockfile *LockFile) error {
	lockfilePath := lfm.getLockFilePath()

	// Ensure directory exists
	lockDir := filepath.Dir(lockfilePath)
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return fmt.Errorf("failed to create lock file directory: %v", err)
	}

	data, err := json.MarshalIndent(lockfile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal lock file: %v", err)
	}

	return ioutil.WriteFile(lockfilePath, data, 0644)
}

// Generate generates a lock file from resolved dependencies
func (lfm *LockFileManager) Generate(resolved []*ResolvedDependency) (*LockFile, error) {
	lockfile := &LockFile{
		LockfileVersion: 1,
		GeneratedAt:     time.Now(),
		Resolved:        make(map[string]*LockEntry),
		Dependencies:    make(map[string]*DepEntry),
		DevDependencies: make(map[string]*DepEntry),
	}

	// Flatten dependency tree
	for _, dep := range resolved {
		lfm.flattenDependency(dep, lockfile, false)
	}

	return lockfile, nil
}

// Validate validates the lock file against package config
func (lfm *LockFileManager) Validate(config *PackageConfig) error {
	lockfile, err := lfm.Load()
	if err != nil {
		return err
	}

	// Check if all dependencies are satisfied
	for name, constraint := range config.Dependencies {
		entry, exists := lockfile.Resolved[name]
		if !exists {
			return fmt.Errorf("dependency %s not found in lock file", name)
		}

		// Validate version satisfies constraint
		version, err := semver.ParseVersion(entry.Version)
		if err != nil {
			return fmt.Errorf("invalid version %s for %s: %v", entry.Version, name, err)
		}

		r, err := semver.ParseRange(constraint)
		if err != nil {
			return fmt.Errorf("invalid constraint %s for %s: %v", constraint, name, err)
		}

		if !r.Match(version) {
			return fmt.Errorf("locked version %s for %s does not satisfy constraint %s", entry.Version, name, constraint)
		}
	}

	return nil
}

// Update updates a specific package in the lock file
func (lfm *LockFileManager) Update(packageName, newVersion string) error {
	lockfile, err := lfm.Load()
	if err != nil {
		return err
	}

	entry, exists := lockfile.Resolved[packageName]
	if !exists {
		return fmt.Errorf("package %s not found in lock file", packageName)
	}

	// Update version
	entry.Version = newVersion

	// Update timestamp
	lockfile.GeneratedAt = time.Now()

	return lfm.Save(lockfile)
}

// Verify checks integrity of locked packages
func (lfm *LockFileManager) Verify() error {
	lockfile, err := lfm.Load()
	if err != nil {
		return err
	}

	// Verify each locked package
	for name, entry := range lockfile.Resolved {
		// In a real implementation, this would check file checksums
		// For now, just validate version format
		if _, err := semver.ParseVersion(entry.Version); err != nil {
			return fmt.Errorf("invalid version %s for package %s: %v", entry.Version, name, err)
		}

		// Validate integrity checksum format
		if entry.Integrity != "" && len(entry.Integrity) < 10 {
			return fmt.Errorf("invalid integrity checksum for package %s", name)
		}
	}

	return nil
}

// Exists checks if lock file exists
func (lfm *LockFileManager) Exists() bool {
	lockfilePath := lfm.getLockFilePath()
	_, err := os.Stat(lockfilePath)
	return err == nil
}

// flattenDependency flattens a dependency tree into lock file entries
func (lfm *LockFileManager) flattenDependency(dep *ResolvedDependency, lockfile *LockFile, isDev bool) {
	// Check if already processed
	if _, exists := lockfile.Resolved[dep.Name]; exists {
		return
	}

	// Create lock entry
	lockEntry := &LockEntry{
		Version:      dep.Version.String(),
		Integrity:    calculateIntegrity(dep.Name, dep.Version),
		Resolved:     fmt.Sprintf("https://registry.shode.io/%s/-/%s-%s.tgz", dep.Name, dep.Name, dep.Version),
		Dependencies: make(map[string]string),
	}

	// Add to resolved
	lockfile.Resolved[dep.Name] = lockEntry

	// Create dependency entry
	requires := []string{}
	for _, child := range dep.Dependencies {
		requires = append(requires, fmt.Sprintf("%s@%s", child.Name, child.Version))
		lockEntry.Dependencies[child.Name] = child.Version.String()

		// Recursively process child dependencies
		lfm.flattenDependency(child, lockfile, isDev)
	}

	depEntry := &DepEntry{
		Version:  dep.Version.String(),
		Requires: requires,
	}

	if isDev {
		lockfile.DevDependencies[dep.Name] = depEntry
	} else {
		lockfile.Dependencies[dep.Name] = depEntry
	}
}

// getLockFilePath returns the path to the lock file
func (lfm *LockFileManager) getLockFilePath() string {
	if lfm.configPath == "" {
		// Default to current directory
		return "shode-lock.json"
	}

	// Use same directory as config file
	configDir := filepath.Dir(lfm.configPath)
	return filepath.Join(configDir, "shode-lock.json")
}

// calculateIntegrity calculates a mock integrity checksum
// In a real implementation, this would calculate SHA256 of the tarball
func calculateIntegrity(name string, version *semver.Version) string {
	return fmt.Sprintf("sha512-%s-%s", name, version.String())
}

// ResolvedDependency represents a resolved dependency with its transitive dependencies
type ResolvedDependency struct {
	Name         string
	Version      *semver.Version
	Dependencies []*ResolvedDependency
}
