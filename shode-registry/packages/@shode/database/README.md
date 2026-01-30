# @shode/database

Database abstraction layer for Shode applications.

## Features

- Support for MySQL, PostgreSQL, SQLite
- Query execution
- Connection management
- SQL escaping

## Installation

```bash
shode pkg add @shode/database ^1.0.0
```

## Usage

```bash
. sh_modules/@shode/database/index.sh

# Connect to MySQL
DbConnect mysql "host=localhost user=root password=secret dbname=mydb"

# Connect to PostgreSQL
DbConnect postgres "host=localhost user=postgres dbname=mydb"

# Connect to SQLite
DbConnect sqlite "/path/to/database.db"

# Execute SELECT query
DbQuery "SELECT * FROM users"

# Execute INSERT/UPDATE/DELETE
DbExec "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"

# Close connection
DbClose
```

## API

### Functions

- `DbConnect(type, connection_string)` - Connect to database
- `DbQuery(query)` - Execute SELECT query
- `DbExec(query)` - Execute INSERT/UPDATE/DELETE query
- `DbClose()` - Close connection
- `DbEscape(input)` - Escape SQL special characters

## Supported Databases

- MySQL
- PostgreSQL
- SQLite

## Requirements

- MySQL: mysql client
- PostgreSQL: psql client
- SQLite: sqlite3

## License

MIT
