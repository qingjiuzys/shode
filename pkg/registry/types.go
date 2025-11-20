package registry

import "time"

// Package represents a package in the registry
type Package struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	License         string            `json:"license"`
	Homepage        string            `json:"homepage,omitempty"`
	Repository      string            `json:"repository,omitempty"`
	Keywords        []string          `json:"keywords,omitempty"`
	Main            string            `json:"main"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Downloads       int               `json:"downloads"`
	Verified        bool              `json:"verified"`
}

// PackageVersion represents a specific version of a package
type PackageVersion struct {
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	Main            string            `json:"main"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	TarballURL      string            `json:"tarballUrl"`
	Shasum          string            `json:"shasum"`
	Signature       string            `json:"signature,omitempty"`
	SignatureAlgo   string            `json:"signatureAlgo,omitempty"`
	SignerID        string            `json:"signerId,omitempty"`
	PublishedAt     time.Time         `json:"publishedAt"`
	Deprecated      bool              `json:"deprecated,omitempty"`
}

// PackageMetadata represents complete package metadata with all versions
type PackageMetadata struct {
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	Author        string                     `json:"author"`
	License       string                     `json:"license"`
	Homepage      string                     `json:"homepage,omitempty"`
	Repository    string                     `json:"repository,omitempty"`
	Keywords      []string                   `json:"keywords,omitempty"`
	Versions      map[string]*PackageVersion `json:"versions"`
	LatestVersion string                     `json:"latestVersion"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	Downloads     int                        `json:"downloads"`
	Verified      bool                       `json:"verified"`
}

// SearchResult represents a package search result
type SearchResult struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Keywords    []string `json:"keywords,omitempty"`
	Downloads   int      `json:"downloads"`
	Verified    bool     `json:"verified"`
	Score       float64  `json:"score"` // Search relevance score
}

// SearchQuery represents a package search query
type SearchQuery struct {
	Query    string   `json:"query"`
	Keywords []string `json:"keywords,omitempty"`
	Author   string   `json:"author,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// PublishRequest represents a package publish request
type PublishRequest struct {
	Package       *Package `json:"package"`
	Tarball       []byte   `json:"tarball"` // Gzipped tar archive
	Checksum      string   `json:"checksum"`
	Signature     string   `json:"signature"`
	SignatureAlgo string   `json:"signatureAlgo"`
	SignerID      string   `json:"signerId"`
}

// RegistryConfig represents registry configuration
type RegistryConfig struct {
	URL            string `json:"url"`
	Token          string `json:"token,omitempty"`
	CacheDir       string `json:"cacheDir"`
	Timeout        int    `json:"timeout"` // seconds
	TrustStorePath string `json:"trustStorePath,omitempty"`
	AllowUnsigned  bool   `json:"allowUnsigned,omitempty"`
}

// DownloadInfo represents package download information
type DownloadInfo struct {
	PackageName string `json:"packageName"`
	Version     string `json:"version"`
	TarballURL  string `json:"tarballUrl"`
	Checksum    string `json:"checksum"`
	Size        int64  `json:"size"`
}
