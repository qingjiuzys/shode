package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// DatabaseManager manages database connections and operations
type DatabaseManager struct {
	db         *sql.DB
	dbType     string
	dsn        string
	mu         sync.RWMutex
	lastResult *QueryResult
}

// QueryResult holds the result of a database query
type QueryResult struct {
	Rows   []map[string]interface{} `json:"rows"`
	Row    map[string]interface{}   `json:"row,omitempty"`
	Count  int                      `json:"count"`
	LastID int64                    `json:"lastId,omitempty"`
	mu     sync.RWMutex
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager() *DatabaseManager {
	return &DatabaseManager{}
}

// Connect connects to a database
// dbType: "mysql", "postgres", or "sqlite"
// dsn: connection string
func (dm *DatabaseManager) Connect(dbType, dsn string) error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Close existing connection if any
	if dm.db != nil {
		dm.db.Close()
	}

	// Set connection parameters based on database type
	var driverName string
	switch dbType {
	case "mysql":
		driverName = "mysql"
		// Add timeout parameters if not present
		if !contains(dsn, "timeout") {
			if contains(dsn, "?") {
				dsn += "&timeout=10s&readTimeout=10s&writeTimeout=10s"
			} else {
				dsn += "?timeout=10s&readTimeout=10s&writeTimeout=10s"
			}
		}
	case "postgres":
		driverName = "postgres"
	case "sqlite":
		driverName = "sqlite3"
	default:
		return fmt.Errorf("unsupported database type: %s (supported: mysql, postgres, sqlite)", dbType)
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %v", err)
	}

	dm.db = db
	dm.dbType = dbType
	dm.dsn = dsn

	return nil
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if dm.db == nil {
		return nil
	}

	err := dm.db.Close()
	dm.db = nil
	dm.dbType = ""
	dm.dsn = ""

	return err
}

// IsConnected checks if the database is connected
func (dm *DatabaseManager) IsConnected() bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()

	if dm.db == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return dm.db.PingContext(ctx) == nil
}

// Query executes a SELECT query and returns results
func (dm *DatabaseManager) Query(sql string, args ...interface{}) (*QueryResult, error) {
	dm.mu.RLock()
	db := dm.db
	dm.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	var resultRows []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Create a map for this row
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if val != nil {
				// Convert []byte to string
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			} else {
				row[col] = nil
			}
		}

		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	result := &QueryResult{
		Rows:  resultRows,
		Count: len(resultRows),
	}

	dm.mu.Lock()
	dm.lastResult = result
	dm.mu.Unlock()

	return result, nil
}

// QueryRow executes a SELECT query and returns a single row
func (dm *DatabaseManager) QueryRow(sql string, args ...interface{}) (*QueryResult, error) {
	dm.mu.RLock()
	db := dm.db
	dm.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	// For simplicity, we'll use Query and take the first row
	result, err := dm.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	if result.Count > 0 {
		result.Row = result.Rows[0]
		result.Rows = nil // Clear rows for single row result
	}

	dm.mu.Lock()
	dm.lastResult = result
	dm.mu.Unlock()

	return result, nil
}

// Exec executes a non-query SQL statement (INSERT, UPDATE, DELETE)
func (dm *DatabaseManager) Exec(sql string, args ...interface{}) (*QueryResult, error) {
	dm.mu.RLock()
	db := dm.db
	dm.mu.RUnlock()

	if db == nil {
		return nil, fmt.Errorf("database not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %v", err)
	}

	queryResult := &QueryResult{
		Count: 0,
	}

	// Get affected rows
	rowsAffected, err := result.RowsAffected()
	if err == nil {
		queryResult.Count = int(rowsAffected)
	}

	// Get last insert ID (for INSERT statements)
	lastID, err := result.LastInsertId()
	if err == nil && lastID > 0 {
		queryResult.LastID = lastID
	}

	dm.mu.Lock()
	dm.lastResult = queryResult
	dm.mu.Unlock()

	return queryResult, nil
}

// GetLastResult returns the last query result
func (dm *DatabaseManager) GetLastResult() *QueryResult {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.lastResult
}

// GetLastResultJSON returns the last query result as JSON
func (dm *DatabaseManager) GetLastResultJSON() (string, error) {
	result := dm.GetLastResult()
	if result == nil {
		return "{}", nil
	}

	result.mu.RLock()
	defer result.mu.RUnlock()

	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(data), nil
}

// ToJSON converts a QueryResult to JSON string
func (qr *QueryResult) ToJSON() (string, error) {
	qr.mu.RLock()
	defer qr.mu.RUnlock()

	data, err := json.Marshal(qr)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %v", err)
	}

	return string(data), nil
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
