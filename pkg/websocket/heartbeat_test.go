package websocket

import (
	"testing"
	"time"
)

// TestHeartbeatManagerCreation tests creating a heartbeat manager
func TestHeartbeatManagerCreation(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	if hm == nil {
		t.Fatal("NewHeartbeatManager returned nil")
	}

	if hm.manager != manager {
		t.Error("HeartbeatManager manager reference not set correctly")
	}
}

// TestHeartbeatManagerDefaultConfig tests using default configuration
func TestHeartbeatManagerDefaultConfig(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	if hm.config.Interval != 30*time.Second {
		t.Errorf("Expected default interval 30s, got %v", hm.config.Interval)
	}

	if hm.config.Timeout != 60*time.Second {
		t.Errorf("Expected default timeout 60s, got %v", hm.config.Timeout)
	}

	if !hm.config.UseWebSocketPing {
		t.Error("Expected default UseWebSocketPing to be true")
	}
}

// TestHeartbeatManagerCustomConfig tests using custom configuration
func TestHeartbeatManagerCustomConfig(t *testing.T) {
	customConfig := &HeartbeatConfig{
		Interval:        10 * time.Second,
		Timeout:         20 * time.Second,
		PingMessage:     []byte("custom_ping"),
		PongMessage:     []byte("custom_pong"),
		UseWebSocketPing: false,
	}

	manager := NewManager()
	hm := NewHeartbeatManager(customConfig, manager)

	if hm.config.Interval != 10*time.Second {
		t.Errorf("Expected custom interval 10s, got %v", hm.config.Interval)
	}

	if hm.config.Timeout != 20*time.Second {
		t.Errorf("Expected custom timeout 20s, got %v", hm.config.Timeout)
	}

	if hm.config.UseWebSocketPing {
		t.Error("Expected custom UseWebSocketPing to be false")
	}
}

// TestHeartbeatManagerRegisterUnregister tests connection registration
func TestHeartbeatManagerRegisterUnregister(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	// Register a connection
	connID := "test_conn_1"
	hm.RegisterConnection(connID)

	if !hm.IsActive(connID) {
		t.Error("Connection should be active after registration")
	}

	// Unregister the connection
	hm.UnregisterConnection(connID)

	if hm.IsActive(connID) {
		t.Error("Connection should not be active after unregistration")
	}
}

// TestHeartbeatManagerUpdatePong tests pong time updates
func TestHeartbeatManagerUpdatePong(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	connID := "test_conn_2"
	hm.RegisterConnection(connID)

	// Get initial pong time
	pong1 := hm.GetLastPong(connID)
	if pong1.IsZero() {
		t.Error("Pong time should be set after registration")
	}

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	hm.UpdatePong(connID)

	pong2 := hm.GetLastPong(connID)
	if !pong2.After(pong1) {
		t.Error("Pong time should be updated after UpdatePong")
	}
}

// TestHeartbeatManagerHandleMessage tests message handling
func TestHeartbeatManagerHandleMessage(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	connID := "test_conn_3"
	hm.RegisterConnection(connID)

	// Test handling WebSocket protocol pong
	handled := hm.HandleMessage(connID, PongMessage, []byte{})
	if !handled {
		t.Error("PongMessage should be handled")
	}

	// Test handling application-level pong
	hm.config.PongMessage = []byte("pong")
	handled = hm.HandleMessage(connID, TextMessage, []byte("pong"))
	if !handled {
		t.Error("Application-level pong should be handled")
	}

	// Test handling non-pong message
	handled = hm.HandleMessage(connID, TextMessage, []byte("hello"))
	if handled {
		t.Error("Regular message should not be handled as pong")
	}
}

// TestHeartbeatManagerGetStats tests statistics
func TestHeartbeatManagerGetStats(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	// Register some connections
	hm.RegisterConnection("conn1")
	hm.RegisterConnection("conn2")
	hm.RegisterConnection("conn3")

	stats := hm.GetStats()

	if stats.TotalConnections != 3 {
		t.Errorf("Expected 3 total connections, got %d", stats.TotalConnections)
	}

	if stats.ActiveConnections != 3 {
		t.Errorf("Expected 3 active connections, got %d", stats.ActiveConnections)
	}

	if stats.Config != hm.config {
		t.Error("Stats config should match manager config")
	}
}

// TestHeartbeatManagerTimeout tests timeout detection
func TestHeartbeatManagerTimeout(t *testing.T) {
	// Create a config with very short timeout
	config := &HeartbeatConfig{
		Interval:    50 * time.Millisecond,
		Timeout:     100 * time.Millisecond,
		PingMessage: []byte("ping"),
		PongMessage: []byte("pong"),
	}

	manager := NewManager()
	hm := NewHeartbeatManager(config, manager)

	// Track timeout events
	timeoutDetected := make(chan string, 1)
	hm.Start(func(connID string) {
		timeoutDetected <- connID
	})
	defer hm.Stop()

	// Register a connection but don't send pong
	connID := "timeout_conn"
	hm.RegisterConnection(connID)

	// Wait for timeout
	select {
	case detectedID := <-timeoutDetected:
		if detectedID != connID {
			t.Errorf("Expected timeout for %s, got %s", connID, detectedID)
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("Timeout was not detected within expected time")
	}

	// Connection should no longer be active
	if hm.IsActive(connID) {
		t.Error("Connection should not be active after timeout")
	}
}

// TestHeartbeatManagerIntegration tests heartbeat with simulated connection
func TestHeartbeatManagerIntegration(t *testing.T) {
	// Use text message ping/pong for simpler testing
	manager := NewManager()
	config := &HeartbeatConfig{
		Interval:        50 * time.Millisecond,
		Timeout:         150 * time.Millisecond,
		PingMessage:     []byte("ping"),
		PongMessage:     []byte("pong"),
		UseWebSocketPing: false, // Use text messages for simpler test
	}
	hm := NewHeartbeatManager(config, manager)

	// Simulate an active connection that responds to pings
	connID := "simulated_conn"
	manager.AddConnection(&Connection{
		ID:        connID,
		WriteChan: make(chan []byte, 10),
		CloseChan: make(chan bool),
	})
	hm.RegisterConnection(connID)

	// Simulate pong responses
	stopSimulation := make(chan bool)
	go func() {
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				hm.HandlePong(connID)
			case <-stopSimulation:
				return
			}
		}
	}()

	// Start heartbeat
	timeoutDetected := make(chan string, 1)
	hm.Start(func(cID string) {
		timeoutDetected <- cID
	})
	defer hm.Stop()

	// Wait for a few heartbeat cycles
	time.Sleep(200 * time.Millisecond)

	// Stop simulation
	stopSimulation <- true
	time.Sleep(50 * time.Millisecond)

	// Connection should still be active
	if !hm.IsActive(connID) {
		t.Error("Connection should still be active with proper ping/pong")
	}

	// Check stats
	stats := hm.GetStats()
	t.Logf("Heartbeat stats: Total=%d, Active=%d", stats.TotalConnections, stats.ActiveConnections)

	if stats.TotalConnections != 1 {
		t.Errorf("Expected 1 total connection, got %d", stats.TotalConnections)
	}

	if stats.ActiveConnections != 1 {
		t.Errorf("Expected 1 active connection, got %d", stats.ActiveConnections)
	}

	// No timeout should have been detected
	select {
	case <-timeoutDetected:
		t.Error("Timeout should not be detected with active connection")
	default:
		// Good, no timeout
	}
}

// TestHeartbeatManagerStop tests stopping the heartbeat manager
func TestHeartbeatManagerStop(t *testing.T) {
	manager := NewManager()
	config := &HeartbeatConfig{
		Interval: 50 * time.Millisecond,
		Timeout:  100 * time.Millisecond,
	}
	hm := NewHeartbeatManager(config, manager)

	// Start heartbeat
	hm.Start(func(connID string) {
		t.Logf("Timeout detected: %s", connID)
	})

	// Stop immediately - should not panic
	hm.Stop()

	// Register connection after stop
	hm.RegisterConnection("test_conn")

	// Should not cause any issues
	time.Sleep(100 * time.Millisecond)
}

// TestHeartbeatManagerConcurrentAccess tests concurrent access to heartbeat manager
func TestHeartbeatManagerConcurrentAccess(t *testing.T) {
	manager := NewManager()
	hm := NewHeartbeatManager(nil, manager)

	done := make(chan bool)

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		go func(n int) {
			connID := "conn_" + string(rune('0'+n))
			hm.RegisterConnection(connID)
			hm.UpdatePong(connID)
			hm.IsActive(connID)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not have any race conditions
	stats := hm.GetStats()
	t.Logf("Concurrent test completed. Total connections: %d", stats.TotalConnections)
}
