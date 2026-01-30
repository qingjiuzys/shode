package database

import (
	"database/sql"
	"fmt"
	"time"

	// Import database drivers
	_ "github.com/lib/pq"        // PostgreSQL
	_ "github.com/mattn/go-sqlite3" // SQLite
)

// Config holds database connection configuration
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultConfig returns default configuration for the given driver
func DefaultConfig(driver string) *Config {
	return &Config{
		Driver:          driver,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

// Open opens a database connection with the given configuration
func Open(config *Config) (*sql.DB, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	db, err := sql.Open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	if config.MaxOpenConns > 0 {
		db.SetMaxOpenConns(config.MaxOpenConns)
	}
	if config.MaxIdleConns > 0 {
		db.SetMaxIdleConns(config.MaxIdleConns)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}

	// Verify connection is alive
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// OpenSQLite opens a SQLite database connection
func OpenSQLite(filepath string) (*ORM, error) {
	config := DefaultConfig("sqlite3")
	config.DSN = filepath
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return NewORM(db, "sqlite3"), nil
}

// OpenPostgreSQL opens a PostgreSQL database connection
func OpenPostgreSQL(host, port, user, password, dbname string) (*ORM, error) {
	config := DefaultConfig("postgres")
	config.DSN = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return NewORM(db, "postgres"), nil
}

// OpenPostgreSQLWithDSN opens a PostgreSQL database with a custom DSN
func OpenPostgreSQLWithDSN(dsn string) (*ORM, error) {
	config := DefaultConfig("postgres")
	config.DSN = dsn
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return NewORM(db, "postgres"), nil
}

// OpenMySQL opens a MySQL database connection
func OpenMySQL(host, port, user, password, dbname string) (*ORM, error) {
	config := DefaultConfig("mysql")
	config.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		user, password, host, port, dbname)
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return NewORM(db, "mysql"), nil
}

// OpenMySQLWithDSN opens a MySQL database with a custom DSN
func OpenMySQLWithDSN(dsn string) (*ORM, error) {
	config := DefaultConfig("mysql")
	config.DSN = dsn
	db, err := Open(config)
	if err != nil {
		return nil, err
	}
	return NewORM(db, "mysql"), nil
}
