package transaction

import (
	"database/sql"
	"fmt"
	"sync"
)

// TransactionManager manages database transactions
type TransactionManager struct {
	db     *sql.DB
	active map[string]*sql.Tx
	mu     sync.RWMutex
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(db *sql.DB) *TransactionManager {
	return &TransactionManager{
		db:     db,
		active: make(map[string]*sql.Tx),
	}
}

// Begin starts a new transaction
func (tm *TransactionManager) Begin(txID string) (*sql.Tx, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.active[txID]; exists {
		return nil, fmt.Errorf("transaction %s already exists", txID)
	}

	tx, err := tm.db.Begin()
	if err != nil {
		return nil, err
	}

	tm.active[txID] = tx
	return tx, nil
}

// Get retrieves an active transaction
func (tm *TransactionManager) Get(txID string) (*sql.Tx, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	tx, exists := tm.active[txID]
	return tx, exists
}

// Commit commits a transaction
func (tm *TransactionManager) Commit(txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.active[txID]
	if !exists {
		return fmt.Errorf("transaction %s not found", txID)
	}

	err := tx.Commit()
	delete(tm.active, txID)
	return err
}

// Rollback rolls back a transaction
func (tm *TransactionManager) Rollback(txID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tx, exists := tm.active[txID]
	if !exists {
		return fmt.Errorf("transaction %s not found", txID)
	}

	err := tx.Rollback()
	delete(tm.active, txID)
	return err
}

// PropagationBehavior defines transaction propagation behavior
type PropagationBehavior int

const (
	PropagationRequired PropagationBehavior = iota
	PropagationRequiresNew
	PropagationNested
	PropagationSupports
	PropagationNotSupported
	PropagationNever
	PropagationMandatory
)

// TransactionContext holds transaction context
type TransactionContext struct {
	TxID      string
	Propagation PropagationBehavior
	Isolation  sql.IsolationLevel
	ReadOnly   bool
}

// NewTransactionContext creates a new transaction context
func NewTransactionContext(txID string) *TransactionContext {
	return &TransactionContext{
		TxID:        txID,
		Propagation: PropagationRequired,
		Isolation:   sql.LevelDefault,
		ReadOnly:    false,
	}
}
