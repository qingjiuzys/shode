package repository

import (
	"database/sql"
	"fmt"
	"strings"
)

// Repository is the base repository interface
type Repository interface {
	FindByID(id interface{}) (interface{}, error)
	FindAll() ([]interface{}, error)
	Create(entity interface{}) error
	Update(entity interface{}) error
	Delete(id interface{}) error
}

// BaseRepository provides base CRUD operations
type BaseRepository struct {
	db     *sql.DB
	table  string
	idColumn string
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *sql.DB, table, idColumn string) *BaseRepository {
	return &BaseRepository{
		db:       db,
		table:    table,
		idColumn: idColumn,
	}
}

// FindByID finds an entity by ID
func (r *BaseRepository) FindByID(id interface{}) (*sql.Row, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", r.table, r.idColumn)
	return r.db.QueryRow(query, id), nil
}

// FindAll finds all entities
func (r *BaseRepository) FindAll() (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT * FROM %s", r.table)
	return r.db.Query(query)
}

// Create creates a new entity
func (r *BaseRepository) Create(entity map[string]interface{}) (int64, error) {
	columns := make([]string, 0, len(entity))
	placeholders := make([]string, 0, len(entity))
	values := make([]interface{}, 0, len(entity))

	for col, val := range entity {
		if col != r.idColumn { // Skip ID column for insert
			columns = append(columns, col)
			placeholders = append(placeholders, "?")
			values = append(values, val)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		r.table,
		fmt.Sprintf("%v", columns),
		fmt.Sprintf("%v", placeholders))

	result, err := r.db.Exec(query, values...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Update updates an entity
func (r *BaseRepository) Update(id interface{}, entity map[string]interface{}) error {
	sets := make([]string, 0, len(entity))
	values := make([]interface{}, 0, len(entity))

	for col, val := range entity {
		if col != r.idColumn {
			sets = append(sets, fmt.Sprintf("%s = ?", col))
			values = append(values, val)
		}
	}

	values = append(values, id)
	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
		r.table,
		strings.Join(sets, ", "),
		r.idColumn)

	_, err := r.db.Exec(query, values...)
	return err
}

// Delete deletes an entity by ID
func (r *BaseRepository) Delete(id interface{}) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", r.table, r.idColumn)
	_, err := r.db.Exec(query, id)
	return err
}
