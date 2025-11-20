package registry

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gitee.com/com_818cloud/shode/pkg/security"
)

// Client represents a registry client
type Client struct {
	config     *RegistryConfig
	httpClient *http.Client
	cache      *Cache
	trustStore *security.TrustStore
}

// NewClient creates a new registry client
func NewClient(config *RegistryConfig) (*Client, error) {
	if config == nil {
		config = &RegistryConfig{
			URL:           "https://registry.shode.io", // Default registry
			CacheDir:      filepath.Join(os.TempDir(), "shode-cache"),
			Timeout:       30,
			AllowUnsigned: true,
		}
	}

	// Create cache directory
	if err := os.MkdirAll(config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %v", err)
	}

	httpClient := &http.Client{
		Timeout: time.Duration(config.Timeout) * time.Second,
	}

	cache := NewCache(config.CacheDir)

	trustStore, err := security.LoadOrCreateTrustStore(config.TrustStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load trust store: %v", err)
	}

	return &Client{
		config:     config,
		httpClient: httpClient,
		cache:      cache,
		trustStore: trustStore,
	}, nil
}

// Search searches for packages in the registry
func (c *Client) Search(query *SearchQuery) ([]*SearchResult, error) {
	// Build query parameters
	url := fmt.Sprintf("%s/api/search", c.config.URL)

	// Prepare request body
	reqBody, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %v", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var results []*SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return results, nil
}

// GetPackage retrieves package metadata from the registry
func (c *Client) GetPackage(name string) (*PackageMetadata, error) {
	// Check cache first
	if cached, ok := c.cache.GetPackageMetadata(name); ok {
		return cached, nil
	}

	// Fetch from registry
	url := fmt.Sprintf("%s/api/packages/%s", c.config.URL, name)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if c.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.Token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("package not found: %s", name)
		}
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("get package failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var metadata PackageMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	// Cache the result
	c.cache.SetPackageMetadata(name, &metadata)

	return &metadata, nil
}

// GetPackageVersion retrieves a specific version of a package
func (c *Client) GetPackageVersion(name, version string) (*PackageVersion, error) {
	metadata, err := c.GetPackage(name)
	if err != nil {
		return nil, err
	}

	pkgVersion, exists := metadata.Versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found for package %s", version, name)
	}

	return pkgVersion, nil
}

// Download downloads a package tarball
func (c *Client) Download(name, version string) (string, error) {
	// Get package version metadata
	pkgVersion, err := c.GetPackageVersion(name, version)
	if err != nil {
		return "", err
	}

	// Check if already cached
	cacheKey := fmt.Sprintf("%s@%s", name, version)
	if cachedPath, ok := c.cache.GetTarball(cacheKey); ok {
		return cachedPath, nil
	}

	// Download tarball
	resp, err := c.httpClient.Get(pkgVersion.TarballURL)
	if err != nil {
		return "", fmt.Errorf("failed to download tarball: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Read tarball
	tarballData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read tarball: %v", err)
	}

	// Verify checksum
	checksum := calculateChecksum(tarballData)
	if checksum != pkgVersion.Shasum {
		return "", fmt.Errorf("checksum mismatch: expected %s, got %s", pkgVersion.Shasum, checksum)
	}

	if err := c.verifySignature(pkgVersion, tarballData); err != nil {
		return "", err
	}

	// Save to cache
	tarballPath := filepath.Join(c.config.CacheDir, fmt.Sprintf("%s-%s.tar.gz", name, version))
	if err := ioutil.WriteFile(tarballPath, tarballData, 0644); err != nil {
		return "", fmt.Errorf("failed to save tarball: %v", err)
	}

	// Update cache
	c.cache.SetTarball(cacheKey, tarballPath)

	return tarballPath, nil
}

// Publish publishes a package to the registry
func (c *Client) Publish(req *PublishRequest) error {
	if c.config.Token == "" {
		return fmt.Errorf("authentication token required for publishing")
	}

	// Verify checksum
	checksum := calculateChecksum(req.Tarball)
	if checksum != req.Checksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", req.Checksum, checksum)
	}

	// Prepare request
	url := fmt.Sprintf("%s/api/packages", c.config.URL)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.Token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("publish failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Install installs a package to the specified directory
func (c *Client) Install(name, version, targetDir string) error {
	// Download package
	tarballPath, err := c.Download(name, version)
	if err != nil {
		return fmt.Errorf("failed to download package: %v", err)
	}

	// Extract tarball
	packageDir := filepath.Join(targetDir, name)
	if err := extractTarball(tarballPath, packageDir); err != nil {
		return fmt.Errorf("failed to extract package: %v", err)
	}

	return nil
}

// GetConfig returns the registry configuration
func (c *Client) GetConfig() *RegistryConfig {
	return c.config
}

// SetToken sets the authentication token
func (c *Client) SetToken(token string) {
	c.config.Token = token
}

// SetAllowUnsigned configures whether unsigned packages are permitted
func (c *Client) SetAllowUnsigned(allow bool) {
	c.config.AllowUnsigned = allow
}

// ReloadTrustStore reloads trust store from disk (useful after modifications)
func (c *Client) ReloadTrustStore() error {
	store, err := security.LoadOrCreateTrustStore(c.config.TrustStorePath)
	if err != nil {
		return err
	}
	c.trustStore = store
	return nil
}

// calculateChecksum calculates SHA256 checksum of data
func calculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (c *Client) verifySignature(pkgVersion *PackageVersion, tarballData []byte) error {
	if pkgVersion == nil {
		return fmt.Errorf("package version metadata missing")
	}

	if pkgVersion.Signature == "" || pkgVersion.SignerID == "" {
		if c.config.AllowUnsigned {
			return nil
		}
		return fmt.Errorf("package %s is unsigned and allowUnsigned=false", pkgVersion.Version)
	}

	if c.trustStore == nil {
		if c.config.AllowUnsigned {
			return nil
		}
		return fmt.Errorf("trust store not initialized")
	}

	publicKey, ok := c.trustStore.GetPublicKey(pkgVersion.SignerID)
	if !ok {
		if c.config.AllowUnsigned {
			return nil
		}
		return fmt.Errorf("signer %s not trusted", pkgVersion.SignerID)
	}

	if err := security.VerifySignature(tarballData, pkgVersion.Signature, pkgVersion.SignatureAlgo, publicKey); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// extractTarball extracts a tar.gz archive to the target directory
func extractTarball(tarballPath, targetDir string) error {
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %v", err)
	}

	// Open tarball file
	file, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %v", err)
	}
	defer file.Close()

	// TODO: Implement actual tar.gz extraction
	// For now, this is a placeholder
	// In production, use archive/tar and compress/gzip packages

	// Placeholder: Just copy the file as-is
	destFile := filepath.Join(targetDir, "package.tar.gz")
	dest, err := os.Create(destFile)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		return fmt.Errorf("failed to copy tarball: %v", err)
	}

	return nil
}
