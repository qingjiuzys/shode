package engine

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

// ProcessPool manages a pool of reusable processes
type ProcessPool struct {
	pool        map[string]*ProcessEntry
	maxSize     int
	idleTimeout time.Duration
	mu          sync.RWMutex
}

// ProcessEntry represents a process in the pool
type ProcessEntry struct {
	cmd       string
	process   *os.Process
	lastUsed  time.Time
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderr    io.ReadCloser
	isRunning bool
}

// NewProcessPool creates a new process pool
func NewProcessPool(maxSize int, idleTimeout time.Duration) *ProcessPool {
	pool := &ProcessPool{
		pool:        make(map[string]*ProcessEntry),
		maxSize:     maxSize,
		idleTimeout: idleTimeout,
	}

	// Start cleanup goroutine
	go pool.cleanupIdleProcesses()

	return pool
}

// Get gets a process from the pool or creates a new one
func (pp *ProcessPool) Get(cmd string, args []string) (*ProcessEntry, error) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	key := generateProcessKey(cmd, args)

	// Try to get from pool
	if entry, exists := pp.pool[key]; exists && entry.isRunning {
		entry.lastUsed = time.Now()
		return entry, nil
	}

	// Create new process
	entry, err := pp.createProcess(cmd, args)
	if err != nil {
		return nil, err
	}

	// Add to pool (evict if necessary)
	if len(pp.pool) >= pp.maxSize {
		pp.evictOldest()
	}

	pp.pool[key] = entry
	return entry, nil
}

// Put returns a process to the pool
func (pp *ProcessPool) Put(entry *ProcessEntry) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	key := generateProcessKey(entry.cmd, nil) // Use nil args for key
	entry.lastUsed = time.Now()
	
	if _, exists := pp.pool[key]; !exists && len(pp.pool) < pp.maxSize {
		pp.pool[key] = entry
	} else {
		// Pool is full or entry already exists, close the process
		entry.Close()
	}
}

// createProcess creates a new process entry
func (pp *ProcessPool) createProcess(cmd string, args []string) (*ProcessEntry, error) {
	// Create command
	command := exec.Command(cmd, args...)
	
	// Set up stdio pipes
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	
	stdout, err := command.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, err
	}
	
	stderr, err := command.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, err
	}
	
	// Start the process
	if err := command.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, err
	}
	
	return &ProcessEntry{
		cmd:       cmd,
		process:   command.Process,
		lastUsed:  time.Now(),
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		isRunning: true,
	}, nil
}

// cleanupIdleProcesses periodically cleans up idle processes
func (pp *ProcessPool) cleanupIdleProcesses() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		pp.mu.Lock()
		now := time.Now()
		
		for key, entry := range pp.pool {
			if now.Sub(entry.lastUsed) > pp.idleTimeout {
				entry.Close()
				delete(pp.pool, key)
			}
		}
		
		pp.mu.Unlock()
	}
}

// evictOldest evicts the oldest process from the pool
func (pp *ProcessPool) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range pp.pool {
		if oldestTime.IsZero() || entry.lastUsed.Before(oldestTime) {
			oldestTime = entry.lastUsed
			oldestKey = key
		}
	}

	if oldestKey != "" {
		pp.pool[oldestKey].Close()
		delete(pp.pool, oldestKey)
	}
}

// Close closes all processes in the pool
func (pp *ProcessPool) Close() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	for key, entry := range pp.pool {
		entry.Close()
		delete(pp.pool, key)
	}
}

// Close closes the process and releases resources
func (pe *ProcessEntry) Close() {
	if pe.process != nil {
		pe.process.Kill()
		pe.process.Wait()
	}
	
	if pe.stdin != nil {
		pe.stdin.Close()
	}
	if pe.stdout != nil {
		pe.stdout.Close()
	}
	if pe.stderr != nil {
		pe.stderr.Close()
	}
	
	pe.isRunning = false
}

// generateProcessKey generates a unique key for a process
func generateProcessKey(cmd string, args []string) string {
	if len(args) == 0 {
		return cmd
	}
	return cmd + ":" + stringsJoin(args, ":")
}

// stringsJoin is a helper function to join strings
func stringsJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
