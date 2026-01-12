#!/usr/bin/env shode

# Account Transfer Example
# Demonstrates database transaction simulation (account transfer)

Println "=== Account Transfer Demo ==="

# Connect to database
Println "Connecting to database..."
ConnectDB "sqlite" "accounts.db"
Println "Database connected"

# Create accounts table
Println "Creating accounts table..."
ExecDB "CREATE TABLE IF NOT EXISTS accounts (id INTEGER PRIMARY KEY, balance REAL)"

# Create two accounts with initial balances
Println "Creating accounts..."
ExecDB "INSERT OR REPLACE INTO accounts (id, balance) VALUES (?, ?)" "1" "1000.00"
ExecDB "INSERT OR REPLACE INTO accounts (id, balance) VALUES (?, ?)" "2" "500.00"
Println "Accounts created:"
Println "  Account 1: $1000.00"
Println "  Account 2: $500.00"

# Query initial balances
Println "Querying initial balances..."
QueryDB "SELECT id, balance FROM accounts ORDER BY id"
initialBalances = GetQueryResult
Println "Initial balances: " + initialBalances

# Simulate transfer: $100 from account 1 to account 2
transferAmount = "100.00"
Println ""
Println "Transferring $" + transferAmount + " from Account 1 to Account 2..."

# Step 1: Deduct from account 1
Println "Deducting from Account 1..."
ExecDB "UPDATE accounts SET balance = balance - ? WHERE id = ?" transferAmount "1"
Println "Deduction complete"

# Step 2: Add to account 2
Println "Adding to Account 2..."
ExecDB "UPDATE accounts SET balance = balance + ? WHERE id = ?" transferAmount "2"
Println "Addition complete"

# Query final balances
Println "Querying final balances..."
QueryDB "SELECT id, balance FROM accounts ORDER BY id"
finalBalances = GetQueryResult
Println "Final balances: " + finalBalances

# Verify transfer
Println ""
Println "Transfer verification:"
Println "  Account 1 should have: $900.00"
Println "  Account 2 should have: $600.00"

# Close database
Println "Closing database..."
CloseDB
Println "Database closed"

Println ""
Println "=== Account Transfer Demo Complete ==="
