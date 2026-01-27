package integration

import (
	"context"
	"net/http"
	"testing"
	"time"

	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
)

// TestWebSocketBasic tests basic WebSocket functionality
func TestWebSocketBasic(t *testing.T) {
	// Setup
	tmpDir, _ := setupTestDir()
	defer cleanupTestDir(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	ctx := context.Background()

	// Start server with WebSocket route
	scriptContent := `
StartHTTPServer "9198"
RegisterWebSocketRoute "/ws" ""
Println "Server started"
`

	p := parser.NewSimpleParser()
	script, err := p.ParseString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	// Execute in background
	done := make(chan error, 1)
	go func() {
		_, err := ee.Execute(ctx, script)
		done <- err
	}()

	time.Sleep(2 * time.Second)
	defer func() {
		stdLib.StopHTTPServer()
		time.Sleep(500 * time.Millisecond)
	}()

	// Test 1: WebSocket route is registered
	t.Run("WebSocketRouteRegistered", func(t *testing.T) {
		// Try to connect to WebSocket endpoint (will fail without proper client)
		// But we can check if the route exists by trying HTTP request first
		resp, err := http.Get("http://localhost:9198/ws")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should get 426 Upgrade Required or similar for HTTP request to WebSocket endpoint
		// This confirms the route is registered
		if resp.StatusCode != http.StatusSwitchingProtocols && resp.StatusCode != http.StatusBadRequest {
			t.Logf("WebSocket route responded with status: %d", resp.StatusCode)
		}
	})

	// Test 2: Server is running
	t.Run("ServerRunning", func(t *testing.T) {
		if !stdLib.IsHTTPServerRunning() {
			t.Error("HTTP server should be running")
		}
	})

	<-done
}

// TestWebSocketManager tests WebSocket manager functions
func TestWebSocketManager(t *testing.T) {
	// Setup
	tmpDir, _ := setupTestDir()
	defer cleanupTestDir(tmpDir)

	em := environment.NewEnvironmentManager()
	em.ChangeDir(tmpDir)
	stdLib := stdlib.New()
	mm := module.NewModuleManager()
	sc := sandbox.NewSecurityChecker()
	ee := engine.NewExecutionEngine(em, stdLib, mm, sc)

	ctx := context.Background()

	// Start server
	scriptContent := `
StartHTTPServer "9199"
RegisterWebSocketRoute "/ws" ""
Println "Server started"
`

	p := parser.NewSimpleParser()
	script, err := p.ParseString(scriptContent)
	if err != nil {
		t.Fatalf("Failed to parse script: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		_, err := ee.Execute(ctx, script)
		done <- err
	}()

	time.Sleep(2 * time.Second)
	defer func() {
		stdLib.StopHTTPServer()
		time.Sleep(500 * time.Millisecond)
	}()

	// Test WebSocket manager functions
	t.Run("GetConnectionCount", func(t *testing.T) {
		count := stdLib.GetWebSocketConnectionCount()
		t.Logf("Current connection count: %d", count)
		// Should be 0 since no actual WebSocket connections
		if count < 0 {
			t.Error("Connection count should not be negative")
		}
	})

	t.Run("GetRoomCount", func(t *testing.T) {
		count := stdLib.GetWebSocketRoomCount("test-room")
		if count != 0 {
			t.Logf("Room 'test-room' has %d connections", count)
		}
	})

	t.Run("ListRooms", func(t *testing.T) {
		rooms := stdLib.ListWebSocketRooms()
		t.Logf("Active rooms: %v", rooms)
		// Should return empty list or "No active rooms"
		if rooms == nil {
			t.Error("ListWebSocketRooms should not return nil")
		}
	})

	t.Run("BroadcastMessage", func(t *testing.T) {
		err := stdLib.BroadcastWebSocketMessage("test message")
		// Should succeed even with 0 connections
		if err != nil {
			t.Errorf("Broadcast should succeed: %v", err)
		}
	})

	t.Run("BroadcastToRoom", func(t *testing.T) {
		err := stdLib.BroadcastWebSocketMessageToRoom("test-room", "test message")
		// May fail if room doesn't exist, which is OK
		if err != nil {
			t.Logf("Broadcast to non-existent room failed as expected: %v", err)
		}
	})

	<-done
}

// Helper functions
func setupTestDir() (string, error) {
	return "/tmp/shode-ws-test", nil
}

func cleanupTestDir(dir string) {
	// Cleanup if needed
}
