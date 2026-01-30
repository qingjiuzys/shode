#!/bin/sh
# Database abstraction implementation

DB_CONNECTION=""

# DbConnect establishes a database connection
DbConnect() {
    local db_type="$1"
    shift
    local connection_string="$*"
    
    case "$db_type" in
        mysql)
            if command -v mysql &> /dev/null; then
                DB_CONNECTION="mysql:$connection_string"
                echo "Connected to MySQL database"
            else
                echo "Error: mysql client not found" >&2
                return 1
            fi
            ;;
        postgresql|postgres)
            if command -v psql &> /dev/null; then
                DB_CONNECTION="postgres:$connection_string"
                echo "Connected to PostgreSQL database"
            else
                echo "Error: psql client not found" >&2
                return 1
            fi
            ;;
        sqlite)
            if [ -f "$connection_string" ]; then
                DB_CONNECTION="sqlite:$connection_string"
                echo "Connected to SQLite database: $connection_string"
            else
                echo "Error: SQLite file not found: $connection_string" >&2
                return 1
            fi
            ;;
        *)
            echo "Error: Unsupported database type: $db_type" >&2
            echo "Supported types: mysql, postgresql, sqlite" >&2
            return 1
            ;;
    esac
    
    return 0
}

# DbQuery executes a SELECT query
DbQuery() {
    local query="$1"
    
    _db_execute "$query" "query"
}

# DbExec executes an INSERT/UPDATE/DELETE query
DbExec() {
    local query="$1"
    
    _db_execute "$query" "exec"
}

# _db_execute executes a database query
_db_execute() {
    local query="$1"
    local query_type="$2"
    
    if [ -z "$DB_CONNECTION" ]; then
        echo "Error: No database connection. Call DbConnect first." >&2
        return 1
    fi
    
    local db_type="${DB_CONNECTION%%:*}"
    
    case "$db_type" in
        mysql)
            if command -v mysql &> /dev/null; then
                mysql "$query" 2>/dev/null
            fi
            ;;
        postgres)
            if command -v psql &> /dev/null; then
                echo "$query" | psql -t -A
            fi
            ;;
        sqlite)
            if command -v sqlite3 &> /dev/null; then
                local db_file="${DB_CONNECTION#*:}"
                sqlite3 "$db_file" "$query"
            fi
            ;;
    esac
}

# DbClose closes the database connection
DbClose() {
    DB_CONNECTION=""
    echo "Database connection closed"
}

# DbEscape escapes special characters for SQL
DbEscape() {
    local input="$1"
    
    # Basic SQL escaping
    echo "$input" | sed 's/\\/\\\\/g' | sed "s/'/\\\\'/g"
}
