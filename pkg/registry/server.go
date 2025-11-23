package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/security"
)

// Server represents a local registry server
type Server struct {
	packages   map[string]*PackageMetadata
	mu         sync.RWMutex
	storageDir string
	port       int
	authToken  string
	trustStore *security.TrustStore
}

const serverTrustStoreFile = "trusted_signers.json"

// NewServer creates a new registry server
func NewServer(storageDir string, port int) (*Server, error) {
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %v", err)
	}

	trustStorePath := filepath.Join(storageDir, serverTrustStoreFile)
	trustStore, err := security.LoadOrCreateTrustStore(trustStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load server trust store: %w", err)
	}

	server := &Server{
		packages:   make(map[string]*PackageMetadata),
		storageDir: storageDir,
		port:       port,
		authToken:  generateAuthToken(),
		trustStore: trustStore,
	}

	// Load existing packages
	if err := server.loadPackages(); err != nil {
		return nil, err
	}

	return server, nil
}

// Start starts the registry server
func (s *Server) Start() error {
	http.HandleFunc("/api/search", s.handleSearch)
	http.HandleFunc("/api/packages/", s.handlePackages)
	http.HandleFunc("/api/packages", s.handlePublish)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Registry server starting on %s\n", addr)
	fmt.Printf("Auth token: %s\n", s.authToken)

	return http.ListenAndServe(addr, nil)
}

// handleSearch handles package search requests
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query
	var query SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "Invalid query", http.StatusBadRequest)
		return
	}

	// Search packages
	results := s.searchPackages(&query)

	// Return results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handlePackages handles package metadata requests
func (s *Server) handlePackages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract package name from URL
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid package name", http.StatusBadRequest)
		return
	}
	packageName := parts[3]

	// Get package metadata
	s.mu.RLock()
	metadata, exists := s.packages[packageName]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, "Package not found", http.StatusNotFound)
		return
	}

	// Increment download counter
	s.mu.Lock()
	metadata.Downloads++
	s.mu.Unlock()

	// Return metadata
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}

// handlePublish handles package publish requests
func (s *Server) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify authentication
	authHeader := r.Header.Get("Authorization")
	if !s.verifyAuth(authHeader) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse publish request
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify checksum
	checksum := calculateChecksum(req.Tarball)
	if checksum != req.Checksum {
		http.Error(w, "Checksum mismatch", http.StatusBadRequest)
		return
	}

	// Verify signature
	if s.trustStore == nil {
		http.Error(w, "Server trust store not configured", http.StatusInternalServerError)
		return
	}

	if req.SignerID == "" || req.Signature == "" {
		http.Error(w, "Signature and signerId are required", http.StatusBadRequest)
		return
	}

	publicKey, ok := s.trustStore.GetPublicKey(req.SignerID)
	if !ok {
		http.Error(w, "Signer not trusted", http.StatusForbidden)
		return
	}

	if err := security.VerifySignature(req.Tarball, req.Signature, req.SignatureAlgo, publicKey); err != nil {
		http.Error(w, fmt.Sprintf("Signature verification failed: %v", err), http.StatusForbidden)
		return
	}

	// Store package
	if err := s.storePackage(&req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store package: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Package published successfully",
	})
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"packages": len(s.packages),
		"time":     time.Now().Unix(),
	})
}

// searchPackages performs package search
func (s *Server) searchPackages(query *SearchQuery) []*SearchResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*SearchResult, 0)
	queryLower := strings.ToLower(query.Query)

	for _, pkg := range s.packages {
		score := 0.0

		// Match name
		if strings.Contains(strings.ToLower(pkg.Name), queryLower) {
			score += 10.0
		}

		// Match description
		if strings.Contains(strings.ToLower(pkg.Description), queryLower) {
			score += 5.0
		}

		// Match keywords
		for _, keyword := range pkg.Keywords {
			if strings.Contains(strings.ToLower(keyword), queryLower) {
				score += 3.0
			}
		}

		// Match author
		if query.Author != "" && strings.Contains(strings.ToLower(pkg.Author), strings.ToLower(query.Author)) {
			score += 2.0
		}

		if score > 0 {
			result := &SearchResult{
				Name:        pkg.Name,
				Version:     pkg.LatestVersion,
				Description: pkg.Description,
				Author:      pkg.Author,
				Keywords:    pkg.Keywords,
				Downloads:   pkg.Downloads,
				Verified:    pkg.Verified,
				Score:       score,
			}
			results = append(results, result)
		}
	}

	// Sort by score (simple bubble sort for small datasets)
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	// Apply limit
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results
}

// storePackage stores a published package
func (s *Server) storePackage(req *PublishRequest) error {
	pkg := req.Package

	// Create package directory
	pkgDir := filepath.Join(s.storageDir, "packages", pkg.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}

	// Save tarball
	tarballPath := filepath.Join(pkgDir, fmt.Sprintf("%s-%s.tar.gz", pkg.Name, pkg.Version))
	if err := ioutil.WriteFile(tarballPath, req.Tarball, 0644); err != nil {
		return err
	}

	// Update or create metadata
	s.mu.Lock()
	defer s.mu.Unlock()

	metadata, exists := s.packages[pkg.Name]
	if !exists {
		metadata = &PackageMetadata{
			Name:        pkg.Name,
			Description: pkg.Description,
			Author:      pkg.Author,
			License:     pkg.License,
			Homepage:    pkg.Homepage,
			Repository:  pkg.Repository,
			Keywords:    pkg.Keywords,
			Versions:    make(map[string]*PackageVersion),
			CreatedAt:   time.Now(),
		}
	}

	// Add version
	pkgVersion := &PackageVersion{
		Version:         pkg.Version,
		Description:     pkg.Description,
		Author:          pkg.Author,
		Main:            pkg.Main,
		Dependencies:    pkg.Dependencies,
		DevDependencies: pkg.DevDependencies,
		TarballURL:      fmt.Sprintf("http://localhost:%d/tarballs/%s/%s-%s.tar.gz", s.port, pkg.Name, pkg.Name, pkg.Version),
		Shasum:          req.Checksum,
		Signature:       req.Signature,
		SignatureAlgo:   req.SignatureAlgo,
		SignerID:        req.SignerID,
		PublishedAt:     time.Now(),
	}
	metadata.Versions[pkg.Version] = pkgVersion
	metadata.LatestVersion = pkg.Version
	metadata.UpdatedAt = time.Now()
	metadata.Verified = true

	s.packages[pkg.Name] = metadata

	// Save metadata to disk
	return s.saveMetadata(pkg.Name, metadata)
}

// saveMetadata saves package metadata to disk
func (s *Server) saveMetadata(name string, metadata *PackageMetadata) error {
	metadataPath := filepath.Join(s.storageDir, "metadata", name+".json")

	// Create metadata directory
	if err := os.MkdirAll(filepath.Dir(metadataPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(metadataPath, data, 0644)
}

// loadPackages loads existing packages from disk
func (s *Server) loadPackages() error {
	metadataDir := filepath.Join(s.storageDir, "metadata")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(metadataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		data, err := ioutil.ReadFile(filepath.Join(metadataDir, entry.Name()))
		if err != nil {
			continue
		}

		var metadata PackageMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			continue
		}

		s.packages[metadata.Name] = &metadata
	}

	return nil
}

// verifyAuth verifies authentication token
func (s *Server) verifyAuth(authHeader string) bool {
	if authHeader == "" {
		return false
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return false
	}

	return parts[1] == s.authToken
}

// GetAuthToken returns the server's authentication token
func (s *Server) GetAuthToken() string {
	return s.authToken
}

// generateAuthToken generates a simple auth token
func generateAuthToken() string {
	return fmt.Sprintf("shode_%d", time.Now().Unix())
}
