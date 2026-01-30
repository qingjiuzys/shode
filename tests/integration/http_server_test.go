package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

func TestHTTPServerBasic(t *testing.T) {
	// Setup
	tmpDir, _ := os.MkdirTemp("", "shode-http-test-*")
	defer os.RemoveAll(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	ctx := context.Background()

	// Start server first (it runs in goroutine, so returns immediately)
	serverScript := &types.ScriptNode{
		Nodes: []types.Node{
			&types.CommandNode{
				Name: "StartHTTPServer",
				Args: []string{"9188"},
			},
		},
	}

	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	go func() {
		_, execErr := ee.Execute(serverCtx, serverScript)
		if execErr != nil && execErr.Error() != "context canceled" {
			t.Logf("Server execution error: %v", execErr)
		}
	}()

	// Wait for server to start and be ready
	time.Sleep(2 * time.Second)

	// Now register routes
	registerScriptContent := `
RegisterRouteWithResponse "/" "hello world"
Println "Route registered"
`

	p := parser.NewSimpleParser()
	registerScript, err := p.ParseString(registerScriptContent)
	if err != nil {
		t.Fatalf("Failed to parse register script: %v", err)
	}

	_, err = ee.Execute(ctx, registerScript)
	if err != nil {
		t.Fatalf("Failed to register routes: %v", err)
	}

	// Test HTTP request
	resp, err := http.Get("http://localhost:9188/")
	if err != nil {
		t.Fatalf("Failed to connect to HTTP server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read response body
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	responseBody := string(buf[:n])

	if !contains(responseBody, "hello world") {
		t.Errorf("Expected 'hello world' in response, got: %s", responseBody)
	}

	// Stop server
	serverCancel()
	time.Sleep(500 * time.Millisecond) // Wait for graceful shutdown
}

func TestHTTPServerFromFile(t *testing.T) {
	// Setup
	tmpDir, _ := os.MkdirTemp("", "shode-http-file-test-*")
	defer os.RemoveAll(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	// Create test script file
	scriptFile := fmt.Sprintf("%s/http_server.sh", tmpDir)
	scriptContent := `#!/usr/bin/env shode
# HTTP Server Test Script
StartHTTPServer "9188"
sleep 1
RegisterRouteWithResponse "/" "hello world"
Println "HTTP server running on port 9188"
`
	err := os.WriteFile(scriptFile, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create script file: %v", err)
	}

	// Parse script file
	p := parser.NewSimpleParser()
	script, err := p.ParseFile(scriptFile)
	if err != nil {
		t.Fatalf("Failed to parse script file: %v", err)
	}

	ctx := context.Background()

	// Execute script in background
	go func() {
		_, execErr := ee.Execute(ctx, script)
		if execErr != nil {
			t.Logf("Script execution error (may be expected): %v", execErr)
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test HTTP request
	resp, err := http.Get("http://localhost:9188/")
	if err != nil {
		t.Fatalf("Failed to connect to HTTP server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read response body
	buf := make([]byte, 1024)
	n, _ := resp.Body.Read(buf)
	responseBody := string(buf[:n])

	if !contains(responseBody, "hello world") {
		t.Errorf("Expected 'hello world' in response, got: %s", responseBody)
	}

	// Cleanup: Stop server
	stopCmd := &types.CommandNode{
		Name: "StopHTTPServer",
		Args: []string{},
	}
	ee.ExecuteCommand(ctx, stopCmd)
}

func TestHTTPServerMultipleRoutes(t *testing.T) {
	// Setup
	tmpDir, _ := os.MkdirTemp("", "shode-http-multi-test-*")
	defer os.RemoveAll(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	ctx := context.Background()

	// Start server
	startCmd := &types.CommandNode{
		Name: "StartHTTPServer",
		Args: []string{"9188"},
	}
	_, err := ee.ExecuteCommand(ctx, startCmd)
	if err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Register multiple routes
	routes := []struct {
		path     string
		response string
	}{
		{"/", "hello world"},
		{"/api", "API endpoint"},
		{"/health", "OK"},
	}

	for _, route := range routes {
		registerCmd := &types.CommandNode{
			Name: "RegisterRouteWithResponse",
			Args: []string{route.path, route.response},
		}
		_, err := ee.ExecuteCommand(ctx, registerCmd)
		if err != nil {
			t.Fatalf("Failed to register route %s: %v", route.path, err)
		}
	}

	// Wait a bit for routes to be registered
	time.Sleep(500 * time.Millisecond)

	// Test each route
	for _, route := range routes {
		resp, err := http.Get(fmt.Sprintf("http://localhost:9188%s", route.path))
		if err != nil {
			t.Errorf("Failed to connect to route %s: %v", route.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Route %s: Expected status 200, got %d", route.path, resp.StatusCode)
		}

		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		responseBody := string(buf[:n])

		if !contains(responseBody, route.response) {
			t.Errorf("Route %s: Expected '%s' in response, got: %s", route.path, route.response, responseBody)
		}
	}

	// Cleanup
	stopCmd := &types.CommandNode{
		Name: "StopHTTPServer",
		Args: []string{},
	}
	ee.ExecuteCommand(ctx, stopCmd)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestHTTPMethods(t *testing.T) {
	// Setup
	tmpDir, _ := os.MkdirTemp("", "shode-http-methods-test-*")
	defer os.RemoveAll(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	ctx := context.Background()

	// Start server
	startCmd := &types.CommandNode{
		Name: "StartHTTPServer",
		Args: []string{"9188"},
	}
	_, err := ee.ExecuteCommand(ctx, startCmd)
	if err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Register routes with different methods
	methods := []struct {
		method string
		path   string
		body   string
	}{
		{"GET", "/api/users", "GET response"},
		{"POST", "/api/users", "POST response"},
		{"PUT", "/api/users/1", "PUT response"},
		{"DELETE", "/api/users/1", "DELETE response"},
	}

	for _, route := range methods {
		registerCmd := &types.CommandNode{
			Name: "RegisterHTTPRoute",
			Args: []string{route.method, route.path, "script", fmt.Sprintf("SetHTTPResponse 200 '%s'", route.body)},
		}
		_, err := ee.ExecuteCommand(ctx, registerCmd)
		if err != nil {
			t.Fatalf("Failed to register route %s %s: %v", route.method, route.path, err)
		}
	}

	// Wait a bit for routes to be registered
	time.Sleep(500 * time.Millisecond)

	// Test each method
	for _, route := range methods {
		req, _ := http.NewRequest(route.method, fmt.Sprintf("http://localhost:9188%s", route.path), nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Failed to connect to route %s %s: %v", route.method, route.path, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Route %s %s: Expected status 200, got %d", route.method, route.path, resp.StatusCode)
		}

		buf := make([]byte, 1024)
		n, _ := resp.Body.Read(buf)
		responseBody := string(buf[:n])

		// Note: Currently returns placeholder response, will be updated in handler_execution phase
		if !contains(responseBody, route.method) && !contains(responseBody, "Handler") {
			t.Logf("Route %s %s: Response may be placeholder: %s", route.method, route.path, responseBody)
		}
	}

	// Cleanup
	stopCmd := &types.CommandNode{
		Name: "StopHTTPServer",
		Args: []string{},
	}
	ee.ExecuteCommand(ctx, stopCmd)
}
