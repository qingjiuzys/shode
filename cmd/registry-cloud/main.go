package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"gitee.com/com_818cloud/shode/pkg/registry"
	"gitee.com/com_818cloud/shode/pkg/registry/cloud"
)

func main() {
	cfg := loadConfig()
	ctx := context.Background()

	service, err := cloud.NewService(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to start cloud registry: %v", err)
	}

	server := &apiServer{
		registry:       service,
		authToken:      os.Getenv("REGISTRY_TOKEN"),
		downloadExpiry: 15 * time.Minute,
	}

	addr := getenv("LISTEN_ADDR", ":8080")
	log.Printf("Shode cloud registry listening on %s", addr)
	if err := http.ListenAndServe(addr, server.routes()); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func loadConfig() *cloud.Config {
	useSSL := strings.ToLower(getenv("S3_USE_SSL", "true")) == "true"
	return &cloud.Config{
		DatabaseURL: getenv("DATABASE_URL", ""),
		S3Endpoint:  getenv("S3_ENDPOINT", "s3.amazonaws.com"),
		S3Bucket:    getenv("S3_BUCKET", "shode-packages"),
		S3AccessKey: getenv("S3_ACCESS_KEY", ""),
		S3SecretKey: getenv("S3_SECRET_KEY", ""),
		S3UseSSL:    useSSL,
		S3Region:    getenv("S3_REGION", "us-east-1"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type apiServer struct {
	registry       *cloud.Service
	authToken      string
	downloadExpiry time.Duration
}

func (s *apiServer) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/packages/", s.handlePackage)
	mux.HandleFunc("/api/packages", s.handlePublish)
	return mux
}

func (s *apiServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

func (s *apiServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query registry.SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		http.Error(w, "invalid query", http.StatusBadRequest)
		return
	}

	results, err := s.registry.SearchPackages(r.Context(), &query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, results)
}

func (s *apiServer) handlePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.authToken != "" {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token != s.authToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	var req registry.PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := s.registry.Publish(r.Context(), &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *apiServer) handlePackage(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/packages/")
	if path == "" {
		http.NotFound(w, r)
		return
	}

	segments := strings.Split(path, "/")
	name := segments[0]

	if len(segments) == 3 && segments[1] == "versions" && segments[2] == "download" {
		version := r.URL.Query().Get("version")
		if version == "" {
			http.Error(w, "version query parameter required", http.StatusBadRequest)
			return
		}
		s.handleDownload(w, r, name, version)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	meta, err := s.registry.GetPackage(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, meta)
}

func (s *apiServer) handleDownload(w http.ResponseWriter, r *http.Request, name, version string) {
	url, err := s.registry.GetDownloadURL(r.Context(), name, version, s.downloadExpiry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}
