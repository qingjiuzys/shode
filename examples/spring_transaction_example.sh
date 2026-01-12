#!/usr/bin/env shode

# Spring-like Transaction Management Example
# Demonstrates transactional operations

Println "=== Transaction Management Example ==="

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "transaction.db"
Println "Database connected"

# Create accounts table
Println "Creating accounts table..."
ExecDB "CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, balance REAL)"
Println "Table created"

# Initialize accounts
Println "Initializing accounts..."
ExecDB "INSERT OR REPLACE INTO accounts (id, balance) VALUES (?, ?)" "1" "1000.00"
ExecDB "INSERT OR REPLACE INTO accounts (id, balance) VALUES (?, ?)" "2" "500.00"
Println "Accounts initialized"

# Transactional transfer function
function transferMoney(fromId, toId, amount) {
    Println "Starting transaction: Transfer $" + amount + " from " + fromId + " to " + toId
    
    # Deduct from source account
    ExecDB "UPDATE accounts SET balance = balance - ? WHERE id = ?" amount fromId
    Println "Deducted from account " + fromId
    
    # Add to destination account
    ExecDB "UPDATE accounts SET balance = balance + ? WHERE id = ?" amount toId
    Println "Added to account " + toId
    
    Println "Transaction completed"
}

# Execute transfer
Println ""
Println "Executing transfer..."
transferMoney "1" "2" "100.00"

# Verify balances
Println ""
Println "Verifying balances..."
QueryDB "SELECT id, balance FROM accounts ORDER BY id"
result = GetQueryResult
Println "Account balances:"
Println result

Println ""
Println "=== Transaction Example Complete ==="
Println "Note: Full transaction management with rollback requires"
Println "      @Transactional annotation support (coming soon)"
