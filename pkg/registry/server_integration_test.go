package registry

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gitee.com/com_818cloud/shode/pkg/security"
)

func TestPublishRequiresValidSignature(t *testing.T) {
	tempDir := t.TempDir()
	server, err := NewServer(tempDir, 8089)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	keyInfo, err := security.GenerateKeyPair("integration", filepath.Join(tempDir, "keys"))
	if err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	publicBytes, err := os.ReadFile(keyInfo.PublicKeyPath)
	if err != nil {
		t.Fatalf("failed to read public key: %v", err)
	}
	publicBase64 := base64.StdEncoding.EncodeToString(publicBytes)

	if err := server.trustStore.AddSigner("integration", publicBase64, "integration test"); err != nil {
		t.Fatalf("failed to trust signer: %v", err)
	}

	tarball := []byte("integration tarball content")
	checksum := calculateChecksum(tarball)
	signature, err := security.SignData(tarball, keyInfo.PrivateKeyPath)
	if err != nil {
		t.Fatalf("failed to sign tarball: %v", err)
	}

	reqBody := PublishRequest{
		Package: &Package{
			Name:    "example",
			Version: "1.0.0",
			Main:    "index.sh",
		},
		Tarball:       tarball,
		Checksum:      checksum,
		Signature:     signature,
		SignatureAlgo: security.SignatureAlgoEd25519,
		SignerID:      "integration",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/packages", bytes.NewReader(bodyBytes))
	request.Header.Set("Authorization", "Bearer "+server.authToken)
	recorder := httptest.NewRecorder()

	server.handlePublish(recorder, request)

	if status := recorder.Result().StatusCode; status != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (%s)", http.StatusCreated, status, recorder.Body.String())
	}

	metadata, exists := server.packages["example"]
	if !exists {
		t.Fatalf("package metadata not stored")
	}
	if !metadata.Verified {
		t.Fatalf("expected package to be marked verified")
	}
	version := metadata.Versions["1.0.0"]
	if version == nil || version.Signature == "" {
		t.Fatalf("expected signature metadata to be stored")
	}
}

func TestPublishRejectsUnknownSigner(t *testing.T) {
	tempDir := t.TempDir()
	server, err := NewServer(tempDir, 8090)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	tarball := []byte("unsigned content")
	reqBody := PublishRequest{
		Package: &Package{
			Name:    "bad-example",
			Version: "0.0.1",
			Main:    "index.sh",
		},
		Tarball:       tarball,
		Checksum:      calculateChecksum(tarball),
		Signature:     "invalid",
		SignatureAlgo: security.SignatureAlgoEd25519,
		SignerID:      "unknown",
	}
	payload, _ := json.Marshal(reqBody)
	request := httptest.NewRequest(http.MethodPost, "/api/packages", bytes.NewReader(payload))
	request.Header.Set("Authorization", "Bearer "+server.authToken)
	recorder := httptest.NewRecorder()

	server.handlePublish(recorder, request)

	if status := recorder.Result().StatusCode; status != http.StatusForbidden {
		t.Fatalf("expected forbidden for unknown signer, got %d", status)
	}
}
