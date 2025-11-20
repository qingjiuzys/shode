package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// SignatureAlgoEd25519 represents the default signing algorithm
	SignatureAlgoEd25519 = "ed25519"
	defaultKeyDirName    = ".shode/keys"
	privateKeyExt        = ".ed25519"
	publicKeyExt         = ".pub"
)

// KeyInfo contains metadata about a signing key pair stored on disk
type KeyInfo struct {
	SignerID        string
	PrivateKeyPath  string
	PublicKeyPath   string
	PublicKeyBase64 string
}

// DefaultKeyDir returns the default directory for storing keys
func DefaultKeyDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine home directory: %w", err)
	}
	return filepath.Join(home, defaultKeyDirName), nil
}

// EnsureKeyDir makes sure the key directory exists
func EnsureKeyDir(dir string) error {
	return os.MkdirAll(dir, 0700)
}

// GenerateKeyPair generates and stores an ed25519 key pair for the given signer ID
func GenerateKeyPair(signerID, dir string) (*KeyInfo, error) {
	if signerID == "" {
		return nil, fmt.Errorf("signer ID is required")
	}

	if dir == "" {
		var err error
		dir, err = DefaultKeyDir()
		if err != nil {
			return nil, err
		}
	}

	if err := EnsureKeyDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create key directory: %w", err)
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	privatePath := filepath.Join(dir, signerID+privateKeyExt)
	publicPath := filepath.Join(dir, signerID+publicKeyExt)

	if err := os.WriteFile(privatePath, priv, 0600); err != nil {
		return nil, fmt.Errorf("failed to write private key: %w", err)
	}

	if err := os.WriteFile(publicPath, pub, 0644); err != nil {
		return nil, fmt.Errorf("failed to write public key: %w", err)
	}

	return &KeyInfo{
		SignerID:        signerID,
		PrivateKeyPath:  privatePath,
		PublicKeyPath:   publicPath,
		PublicKeyBase64: base64.StdEncoding.EncodeToString(pub),
	}, nil
}

// LoadPrivateKey loads an ed25519 private key from disk
func LoadPrivateKey(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	if len(data) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: got %d bytes", len(data))
	}

	return ed25519.PrivateKey(data), nil
}

// LoadPublicKey loads an ed25519 public key from disk
func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	if len(data) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d bytes", len(data))
	}

	return ed25519.PublicKey(data), nil
}

// SignData signs bytes using the provided private key file path
func SignData(data []byte, privateKeyPath string) (string, error) {
	priv, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		return "", err
	}

	signature := ed25519.Sign(priv, data)
	return base64.StdEncoding.EncodeToString(signature), nil
}

// VerifySignature verifies signature with Base64 encoded public key
func VerifySignature(data []byte, signature string, algo string, publicKeyBase64 string) error {
	if algo == "" {
		algo = SignatureAlgoEd25519
	}

	if algo != SignatureAlgoEd25519 {
		return fmt.Errorf("unsupported signature algorithm: %s", algo)
	}

	pubBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return fmt.Errorf("invalid public key encoding: %w", err)
	}

	if len(pubBytes) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: %d", len(pubBytes))
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("invalid signature encoding: %w", err)
	}

	if len(sigBytes) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature size: %d", len(sigBytes))
	}

	if !ed25519.Verify(ed25519.PublicKey(pubBytes), data, sigBytes) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// ListKeyPairs lists available signing key pairs in the directory
func ListKeyPairs(dir string) ([]*KeyInfo, error) {
	if dir == "" {
		var err error
		dir, err = DefaultKeyDir()
		if err != nil {
			return nil, err
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*KeyInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read key directory: %w", err)
	}

	var keys []*KeyInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != privateKeyExt {
			continue
		}

		signerID := entry.Name()[:len(entry.Name())-len(privateKeyExt)]
		privatePath := filepath.Join(dir, entry.Name())
		publicPath := filepath.Join(dir, signerID+publicKeyExt)
		publicBytes, err := os.ReadFile(publicPath)
		if err != nil {
			continue
		}

		keys = append(keys, &KeyInfo{
			SignerID:        signerID,
			PrivateKeyPath:  privatePath,
			PublicKeyPath:   publicPath,
			PublicKeyBase64: base64.StdEncoding.EncodeToString(publicBytes),
		})
	}

	return keys, nil
}
