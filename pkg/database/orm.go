package database

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// ORM provides a simple ORM-like interface for database operations
type ORM struct {
	db     *sql.DB
	driver string
}

// NewORM creates a new ORM instance
func NewORM(db *sql.DB, driver string) *ORM {
	return &ORM{
		db:     db,
		driver: driver,
	}
}

// Model defines the interface for database models
type Model interface {
	// TableName returns the database table name
	TableName() string
	// PrimaryKey returns the primary key field name
	PrimaryKey() string
	// PrimaryKeyValue returns the primary key value
	PrimaryKeyValue() interface{}
}

// ModelInfo stores metadata about a model
type ModelInfo struct {
	TableName  string
	PrimaryKey string
	Fields     []FieldInfo
}

// FieldInfo stores metadata about a model field
type FieldInfo struct {
	Name     string
	Column   string
	Type     reflect.Type
	IsPK     bool
	AutoIncr bool
}

// GetModelInfo extracts model metadata using reflection
func GetModelInfo(model Model) *ModelInfo {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	info := &ModelInfo{
		TableName:  model.TableName(),
		PrimaryKey: model.PrimaryKey(),
		Fields:     make([]FieldInfo, 0),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		column := field.Tag.Get("db")
		if column == "" {
			column = strings.ToLower(field.Name)
		}
		isPK := column == info.PrimaryKey
		autoIncr := field.Tag.Get("auto_incr") == "true"

		info.Fields = append(info.Fields, FieldInfo{
			Name:     field.Name,
			Column:   column,
			Type:     field.Type,
			IsPK:     isPK,
			AutoIncr: autoIncr,
		})
	}

	return info
}

// Create inserts a new record into the database
func (orm *ORM) Create(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	// Build column and value lists
	var columns []string
	var placeholders []string
	var values []interface{}

	for _, field := range info.Fields {
		if field.AutoIncr {
			continue // Skip auto-increment fields
		}
		columns = append(columns, field.Column)
		placeholders = append(placeholders, orm.placeholder(len(columns)))
		values = append(values, getFieldValue(model, field.Name))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		info.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := orm.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	// If auto-increment primary key, set it back to the model
	if info.PrimaryKey != "" {
		for _, field := range info.Fields {
			if field.AutoIncr && field.IsPK {
				id, err := result.LastInsertId()
				if err != nil {
					return fmt.Errorf("failed to get last insert id: %w", err)
				}
				setFieldValue(model, field.Name, id)
				break
			}
		}
	}

	return nil
}

// FindByID retrieves a record by its primary key
func (orm *ORM) FindByID(ctx context.Context, model Model, id interface{}) error {
	info := GetModelInfo(model)

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = $1",
		orm.columnList(info.Fields),
		info.TableName,
		info.PrimaryKey,
	)

	row := orm.db.QueryRowContext(ctx, query, id)
	return orm.scanRow(row, model, info.Fields)
}

// Find retrieves multiple records with optional conditions
func (orm *ORM) Find(ctx context.Context, model Model, where string, args ...interface{}) ([]Model, error) {
	info := GetModelInfo(model)

	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		orm.columnList(info.Fields),
		info.TableName,
	)

	if where != "" {
		query += " WHERE " + where
	}

	rows, err := orm.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	var results []Model
	for rows.Next() {
		// Create new instance of the same type
		newModel := reflect.New(reflect.ValueOf(model).Elem().Type()).Interface().(Model)
		if err := orm.scanRow(rows, newModel, info.Fields); err != nil {
			return nil, err
		}
		results = append(results, newModel)
	}

	return results, rows.Err()
}

// Update updates a record in the database
func (orm *ORM) Update(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	var setParts []string
	var values []interface{}

	for _, field := range info.Fields {
		if field.IsPK {
			continue // Skip primary key in SET clause
		}
		setParts = append(setParts, fmt.Sprintf("%s = %s", field.Column, orm.placeholder(len(setParts)+1)))
		values = append(values, getFieldValue(model, field.Name))
	}

	// Add primary key to WHERE clause
	values = append(values, model.PrimaryKeyValue())

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		info.TableName,
		strings.Join(setParts, ", "),
		info.PrimaryKey,
		len(setParts)+1,
	)

	_, err := orm.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete removes a record from the database
func (orm *ORM) Delete(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		info.TableName,
		info.PrimaryKey,
	)

	_, err := orm.db.ExecContext(ctx, query, model.PrimaryKeyValue())
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// Count returns the number of records matching the condition
func (orm *ORM) Count(ctx context.Context, model Model, where string, args ...interface{}) (int64, error) {
	info := GetModelInfo(model)

	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", info.TableName)
	if where != "" {
		query += " WHERE " + where
	}

	var count int64
	err := orm.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Exists checks if any record matches the condition
func (orm *ORM) Exists(ctx context.Context, model Model, where string, args ...interface{}) (bool, error) {
	count, err := orm.Count(ctx, model, where, args...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// First retrieves the first record matching the condition
func (orm *ORM) First(ctx context.Context, model Model, where string, args ...interface{}) (Model, error) {
	info := GetModelInfo(model)

	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		orm.columnList(info.Fields),
		info.TableName,
	)

	if where != "" {
		query += " WHERE " + where
	}
	query += " LIMIT 1"

	row := orm.db.QueryRowContext(ctx, query, args...)
	// Create new instance of the same type
	newModel := reflect.New(reflect.ValueOf(model).Elem().Type()).Interface().(Model)
	if err := orm.scanRow(row, newModel, info.Fields); err != nil {
		return nil, err
	}
	return newModel, nil
}

// Helper functions

func (orm *ORM) placeholder(n int) string {
	switch orm.driver {
	case "postgres":
		return fmt.Sprintf("$%d", n)
	case "mysql":
		return "?"
	default:
		return "?"
	}
}

func (orm *ORM) columnList(fields []FieldInfo) string {
	columns := make([]string, len(fields))
	for i, field := range fields {
		columns[i] = field.Column
	}
	return strings.Join(columns, ", ")
}

func (orm *ORM) scanRow(row Scanner, model Model, fields []FieldInfo) error {
	v := reflect.ValueOf(model).Elem()
	dests := make([]interface{}, len(fields))

	for i, field := range fields {
		fieldValue := v.FieldByName(field.Name)
		if !fieldValue.IsValid() {
			return fmt.Errorf("field %s not found in model", field.Name)
		}
		dests[i] = fieldValue.Addr().Interface()
	}

	return row.Scan(dests...)
}

// Scanner interface matches sql.Row and sql.Rows
type Scanner interface {
	Scan(dest ...interface{}) error
}

func getFieldValue(model Model, fieldName string) interface{} {
	v := reflect.ValueOf(model).Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}

func setFieldValue(model Model, fieldName string, value interface{}) error {
	v := reflect.ValueOf(model).Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found", fieldName)
	}

	// Convert value to field type
	val := reflect.ValueOf(value)
	if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot convert %v to %v", val.Type(), field.Type())
}

// Transaction wraps the function in a database transaction
func (orm *ORM) Transaction(ctx context.Context, fn func(*Tx) error) error {
	tx, err := orm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback happens if panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic
		}
	}()

	t := &Tx{tx: tx, orm: orm}

	if err := fn(t); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %w", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// Tx represents a database transaction
type Tx struct {
	tx  *sql.Tx
	orm *ORM
}

// Create creates a record within a transaction
func (t *Tx) Create(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	var columns []string
	var placeholders []string
	var values []interface{}

	for _, field := range info.Fields {
		if field.AutoIncr {
			continue
		}
		columns = append(columns, field.Column)
		placeholders = append(placeholders, t.orm.placeholder(len(columns)))
		values = append(values, getFieldValue(model, field.Name))
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		info.TableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	result, err := t.tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	if info.PrimaryKey != "" {
		for _, field := range info.Fields {
			if field.AutoIncr && field.IsPK {
				id, err := result.LastInsertId()
				if err != nil {
					return fmt.Errorf("failed to get last insert id: %w", err)
				}
				setFieldValue(model, field.Name, id)
				break
			}
		}
	}

	return nil
}

// Update updates a record within a transaction
func (t *Tx) Update(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	var setParts []string
	var values []interface{}

	for _, field := range info.Fields {
		if field.IsPK {
			continue
		}
		setParts = append(setParts, fmt.Sprintf("%s = %s", field.Column, t.orm.placeholder(len(setParts)+1)))
		values = append(values, getFieldValue(model, field.Name))
	}

	values = append(values, model.PrimaryKeyValue())

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = $%d",
		info.TableName,
		strings.Join(setParts, ", "),
		info.PrimaryKey,
		len(setParts)+1,
	)

	_, err := t.tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	return nil
}

// Delete removes a record within a transaction
func (t *Tx) Delete(ctx context.Context, model Model) error {
	info := GetModelInfo(model)

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s = $1",
		info.TableName,
		info.PrimaryKey,
	)

	_, err := t.tx.ExecContext(ctx, query, model.PrimaryKeyValue())
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

// Exec executes a custom query within a transaction
func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

// Query executes a query within a transaction
func (t *Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (t *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

// DB returns the underlying database connection
func (orm *ORM) DB() *sql.DB {
	return orm.db
}

// Close closes the database connection
func (orm *ORM) Close() error {
	return orm.db.Close()
}

// Stats returns database statistics
func (orm *ORM) Stats() sql.DBStats {
	return orm.db.Stats()
}

// Ping verifies a connection to the database is still alive
func (orm *ORM) Ping(ctx context.Context) error {
	return orm.db.PingContext(ctx)
}
