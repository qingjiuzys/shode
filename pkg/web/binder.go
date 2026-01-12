package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Binder binds request data to structs or values
type Binder struct{}

// NewBinder creates a new binder
func NewBinder() *Binder {
	return &Binder{}
}

// BindQuery binds query parameters
func (b *Binder) BindQuery(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// BindQueryInt binds query parameter as integer
func (b *Binder) BindQueryInt(r *http.Request, key string, defaultValue int) int {
	value := b.BindQuery(r, key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// BindPath binds path parameter (from URL pattern like /users/:id)
func (b *Binder) BindPath(r *http.Request, key string) string {
	// Extract from path - simplified implementation
	// In full implementation, this would use a router that extracts path params
	path := r.URL.Path
	parts := strings.Split(path, "/")
	
	// Simple pattern matching - look for key in path
	keyWithColon := ":" + key
	for i, part := range parts {
		if part == keyWithColon && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// BindJSON binds JSON request body
func (b *Binder) BindJSON(r *http.Request, target interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %v", err)
	}
	defer r.Body.Close()

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return nil
}

// BindHeader binds request header
func (b *Binder) BindHeader(r *http.Request, key string) string {
	return r.Header.Get(key)
}

// BindForm binds form data
func (b *Binder) BindForm(r *http.Request, key string) string {
	return r.FormValue(key)
}
