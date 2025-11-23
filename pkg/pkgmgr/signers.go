package pkg

import (
	"encoding/base64"
	"fmt"
	"os"

	"gitee.com/com_818cloud/shode/pkg/security"
)

// SignerManager manages signing keys and trusted signers
type SignerManager struct {
	keyDir         string
	trustStorePath string
}

// NewSignerManager creates a new signer manager with default paths
func NewSignerManager() (*SignerManager, error) {
	keyDir, err := security.DefaultKeyDir()
	if err != nil {
		return nil, err
	}

	trustPath, err := security.DefaultTrustStorePath()
	if err != nil {
		return nil, err
	}

	return &SignerManager{
		keyDir:         keyDir,
		trustStorePath: trustPath,
	}, nil
}

// GenerateKey generates a signing key pair
func (sm *SignerManager) GenerateKey(signerID string) (*security.KeyInfo, error) {
	return security.GenerateKeyPair(signerID, sm.keyDir)
}

// ListKeys lists available key pairs
func (sm *SignerManager) ListKeys() ([]*security.KeyInfo, error) {
	return security.ListKeyPairs(sm.keyDir)
}

// TrustPublicKeyFile adds a trusted signer from a public key file
func (sm *SignerManager) TrustPublicKeyFile(id string, publicKeyPath string, description string) error {
	data, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public key: %w", err)
	}

	publicKeyBase64 := base64.StdEncoding.EncodeToString(data)

	store, err := security.LoadOrCreateTrustStore(sm.trustStorePath)
	if err != nil {
		return err
	}

	return store.AddSigner(id, publicKeyBase64, description)
}

// TrustPublicKey adds a trusted signer using base64 encoded key
func (sm *SignerManager) TrustPublicKey(id string, publicKeyBase64 string, description string) error {
	store, err := security.LoadOrCreateTrustStore(sm.trustStorePath)
	if err != nil {
		return err
	}
	return store.AddSigner(id, publicKeyBase64, description)
}

// ListTrustedSigners lists trusted signers
func (sm *SignerManager) ListTrustedSigners() ([]*security.TrustedSigner, error) {
	store, err := security.LoadOrCreateTrustStore(sm.trustStorePath)
	if err != nil {
		return nil, err
	}
	return store.ListSigners(), nil
}

// RemoveTrustedSigner removes a signer from trust store
func (sm *SignerManager) RemoveTrustedSigner(id string) error {
	store, err := security.LoadOrCreateTrustStore(sm.trustStorePath)
	if err != nil {
		return err
	}
	return store.RemoveSigner(id)
}

// KeyDir returns the directory containing key pairs
func (sm *SignerManager) KeyDir() string {
	return sm.keyDir
}

// TrustStorePath returns the trust store path
func (sm *SignerManager) TrustStorePath() string {
	return sm.trustStorePath
}
