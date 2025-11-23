package security

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultTrustDirName  = ".shode/trust"
	trustStoreFileName   = "trusted_signers.json"
	trustStoreFilePerm   = 0644
	publicKeyBase64Label = "publicKey"
)

// TrustedSigner represents a trusted signer entry
type TrustedSigner struct {
	ID          string    `json:"id"`
	PublicKey   string    `json:"publicKey"`
	Description string    `json:"description,omitempty"`
	AddedAt     time.Time `json:"addedAt"`
}

// TrustStore manages trusted signers persisted on disk
type TrustStore struct {
	path    string
	signers map[string]*TrustedSigner
	mu      sync.RWMutex
}

// DefaultTrustStorePath returns the default path for the trust store file
func DefaultTrustStorePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, defaultTrustDirName, trustStoreFileName), nil
}

// LoadOrCreateTrustStore loads a trust store from disk or creates an empty one
func LoadOrCreateTrustStore(path string) (*TrustStore, error) {
	if path == "" {
		var err error
		path, err = DefaultTrustStorePath()
		if err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create trust store directory: %w", err)
	}

	store := &TrustStore{
		path:    path,
		signers: make(map[string]*TrustedSigner),
	}

	if _, err := os.Stat(path); err == nil {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, fmt.Errorf("failed to read trust store: %w", readErr)
		}

		if len(data) > 0 {
			if err := json.Unmarshal(data, &store.signers); err != nil {
				return nil, fmt.Errorf("failed to parse trust store: %w", err)
			}
		}
	}

	return store, nil
}

// Save persists the trust store to disk
func (ts *TrustStore) Save() error {
	ts.mu.RLock()
	snapshot := cloneSigners(ts.signers)
	ts.mu.RUnlock()
	return writeTrustStore(ts.path, snapshot)
}

// AddSigner adds or updates a trusted signer entry
func (ts *TrustStore) AddSigner(id string, publicKeyBase64 string, description string) error {
	if id == "" {
		return fmt.Errorf("signer ID is required")
	}

	if _, err := base64.StdEncoding.DecodeString(publicKeyBase64); err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}

	ts.mu.Lock()
	ts.signers[id] = &TrustedSigner{
		ID:          id,
		PublicKey:   publicKeyBase64,
		Description: description,
		AddedAt:     time.Now(),
	}
	snapshot := cloneSigners(ts.signers)
	ts.mu.Unlock()

	return writeTrustStore(ts.path, snapshot)
}

// RemoveSigner removes a signer from the trust store
func (ts *TrustStore) RemoveSigner(id string) error {
	ts.mu.Lock()
	delete(ts.signers, id)
	snapshot := cloneSigners(ts.signers)
	ts.mu.Unlock()
	return writeTrustStore(ts.path, snapshot)
}

// GetPublicKey returns the public key for a signer
func (ts *TrustStore) GetPublicKey(id string) (string, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	signer, ok := ts.signers[id]
	if !ok {
		return "", false
	}
	return signer.PublicKey, true
}

// ListSigners returns all trusted signers
func (ts *TrustStore) ListSigners() []*TrustedSigner {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	list := make([]*TrustedSigner, 0, len(ts.signers))
	for _, signer := range ts.signers {
		copySigner := *signer
		list = append(list, &copySigner)
	}
	return list
}

// Path returns the backing file path
func (ts *TrustStore) Path() string {
	return ts.path
}

func cloneSigners(src map[string]*TrustedSigner) map[string]*TrustedSigner {
	cloned := make(map[string]*TrustedSigner, len(src))
	for id, signer := range src {
		if signer == nil {
			continue
		}
		copySigner := *signer
		cloned[id] = &copySigner
	}
	return cloned
}

func writeTrustStore(path string, signers map[string]*TrustedSigner) error {
	data, err := json.MarshalIndent(signers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal trust store: %w", err)
	}
	return os.WriteFile(path, data, trustStoreFilePerm)
}
