package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Builder provides a fluent interface for building SQL queries
type Builder struct {
	orm       *ORM
	tableName string

	// SELECT clause
	selectColumns []string

	// WHERE clause
	whereClause []string
	whereArgs   []interface{}
	andOr       string // AND or OR for chaining conditions

	// JOIN clause
	joins []joinClause

	// ORDER BY clause
	orderBys []string

	// LIMIT and OFFSET
	limit  int
	offset int

	// GROUP BY clause
	groupBys []string

	// HAVING clause
	havingClause string
	havingArgs   []interface{}
}

type joinClause struct {
	joinType string
	table    string
	on       string
	args     []interface{}
}

// QueryBuilder creates a new query builder for a table
func (orm *ORM) QueryBuilder(tableName string) *Builder {
	return &Builder{
		orm:       orm,
		tableName: tableName,
		andOr:     "AND",
	}
}

// Select specifies columns to select
func (b *Builder) Select(columns ...string) *Builder {
	b.selectColumns = columns
	return b
}

// Where adds a WHERE condition (uses AND by default)
func (b *Builder) Where(condition string, args ...interface{}) *Builder {
	b.whereClause = append(b.whereClause, condition)
	b.whereArgs = append(b.whereArgs, args...)
	b.andOr = "AND"
	return b
}

// OrWhere adds an OR WHERE condition
func (b *Builder) OrWhere(condition string, args ...interface{}) *Builder {
	b.whereClause = append(b.whereClause, condition)
	b.whereArgs = append(b.whereArgs, args...)
	b.andOr = "OR"
	return b
}

// WhereIn adds a WHERE IN condition
func (b *Builder) WhereIn(column string, values []interface{}) *Builder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = b.orm.placeholder(i + 1)
		b.whereArgs = append(b.whereArgs, values[i])
	}
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	b.whereClause = append(b.whereClause, condition)
	b.andOr = "AND"
	return b
}

// WhereNotIn adds a WHERE NOT IN condition
func (b *Builder) WhereNotIn(column string, values []interface{}) *Builder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = b.orm.placeholder(i + 1)
		b.whereArgs = append(b.whereArgs, values[i])
	}
	condition := fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", "))
	b.whereClause = append(b.whereClause, condition)
	b.andOr = "AND"
	return b
}

// WhereLike adds a WHERE LIKE condition
func (b *Builder) WhereLike(column, pattern string) *Builder {
	condition := fmt.Sprintf("%s LIKE %s", column, b.orm.placeholder(1))
	b.whereClause = append(b.whereClause, condition)
	b.whereArgs = append(b.whereArgs, pattern)
	b.andOr = "AND"
	return b
}

// WhereBetween adds a WHERE BETWEEN condition
func (b *Builder) WhereBetween(column string, min, max interface{}) *Builder {
	condition := fmt.Sprintf("%s BETWEEN %s AND %s", column, b.orm.placeholder(1), b.orm.placeholder(2))
	b.whereClause = append(b.whereClause, condition)
	b.whereArgs = append(b.whereArgs, min, max)
	b.andOr = "AND"
	return b
}

// WhereNull adds a WHERE IS NULL condition
func (b *Builder) WhereNull(column string) *Builder {
	condition := fmt.Sprintf("%s IS NULL", column)
	b.whereClause = append(b.whereClause, condition)
	b.andOr = "AND"
	return b
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (b *Builder) WhereNotNull(column string) *Builder {
	condition := fmt.Sprintf("%s IS NOT NULL", column)
	b.whereClause = append(b.whereClause, condition)
	b.andOr = "AND"
	return b
}

// Join adds an INNER JOIN clause
func (b *Builder) Join(table, on string, args ...interface{}) *Builder {
	b.joins = append(b.joins, joinClause{joinType: "INNER JOIN", table: table, on: on, args: args})
	return b
}

// LeftJoin adds a LEFT JOIN clause
func (b *Builder) LeftJoin(table, on string, args ...interface{}) *Builder {
	b.joins = append(b.joins, joinClause{joinType: "LEFT JOIN", table: table, on: on, args: args})
	return b
}

// RightJoin adds a RIGHT JOIN clause
func (b *Builder) RightJoin(table, on string, args ...interface{}) *Builder {
	b.joins = append(b.joins, joinClause{joinType: "RIGHT JOIN", table: table, on: on, args: args})
	return b
}

// OrderBy adds an ORDER BY clause
func (b *Builder) OrderBy(column string) *Builder {
	b.orderBys = append(b.orderBys, column)
	return b
}

// OrderByDesc adds an ORDER BY DESC clause
func (b *Builder) OrderByDesc(column string) *Builder {
	b.orderBys = append(b.orderBys, column+" DESC")
	return b
}

// GroupBy adds a GROUP BY clause
func (b *Builder) GroupBy(column string) *Builder {
	b.groupBys = append(b.groupBys, column)
	return b
}

// Having adds a HAVING clause
func (b *Builder) Having(condition string, args ...interface{}) *Builder {
	b.havingClause = condition
	b.havingArgs = args
	return b
}

// Limit adds a LIMIT clause
func (b *Builder) Limit(n int) *Builder {
	b.limit = n
	return b
}

// Offset adds an OFFSET clause
func (b *Builder) Offset(n int) *Builder {
	b.offset = n
	return b
}

// Build builds and returns the query string and arguments
func (b *Builder) Build() (string, []interface{}) {
	var query string
	var args []interface{}

	// SELECT
	if len(b.selectColumns) > 0 {
		query = fmt.Sprintf("SELECT %s FROM %s", strings.Join(b.selectColumns, ", "), b.tableName)
	} else {
		query = fmt.Sprintf("SELECT * FROM %s", b.tableName)
	}

	// JOINs
	for _, join := range b.joins {
		query += fmt.Sprintf(" %s %s ON %s", join.joinType, join.table, join.on)
		args = append(args, join.args...)
	}

	// WHERE
	if len(b.whereClause) > 0 {
		query += " WHERE " + strings.Join(b.whereClause, " "+b.andOr+" ")
		args = append(args, b.whereArgs...)
	}

	// GROUP BY
	if len(b.groupBys) > 0 {
		query += " GROUP BY " + strings.Join(b.groupBys, ", ")
	}

	// HAVING
	if b.havingClause != "" {
		query += " HAVING " + b.havingClause
		args = append(args, b.havingArgs...)
	}

	// ORDER BY
	if len(b.orderBys) > 0 {
		query += " ORDER BY " + strings.Join(b.orderBys, ", ")
	}

	// LIMIT
	if b.limit > 0 {
		switch b.orm.driver {
		case "postgres":
			query += fmt.Sprintf(" LIMIT %d", b.limit)
		default:
			query += fmt.Sprintf(" LIMIT %d", b.limit)
		}
	}

	// OFFSET
	if b.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", b.offset)
	}

	return query, args
}

// First executes the query and returns the first result
func (b *Builder) First(ctx context.Context, dest interface{}) error {
	query, args := b.Limit(1).Build()
	row := b.orm.db.QueryRowContext(ctx, query, args...)
	return row.Scan(dest)
}

// All executes the query and returns all results
func (b *Builder) All(ctx context.Context) (*sql.Rows, error) {
	query, args := b.Build()
	return b.orm.db.QueryContext(ctx, query, args...)
}

// Count executes a COUNT query
func (b *Builder) Count(ctx context.Context) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", b.tableName)
	args := []interface{}{}

	if len(b.whereClause) > 0 {
		query += " WHERE " + strings.Join(b.whereClause, " "+b.andOr+" ")
		args = append(args, b.whereArgs...)
	}

	var count int64
	err := b.orm.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count: %w", err)
	}

	return count, nil
}

// Exists checks if any records match the query
func (b *Builder) Exists(ctx context.Context) (bool, error) {
	count, err := b.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Pluck executes the query and returns a single column's values
func (b *Builder) Pluck(ctx context.Context, column string) ([]interface{}, error) {
	selectColumns := b.selectColumns
	b.selectColumns = []string{column}
	defer func() { b.selectColumns = selectColumns }()

	query, args := b.Build()
	rows, err := b.orm.db.QueryContext(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []interface{}
	for rows.Next() {
		var value interface{}
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		results = append(results, value)
	}

	return results, rows.Err()
}

// Update executes an UPDATE query with the given data
func (b *Builder) Update(ctx context.Context, data map[string]interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data to update")
	}

	var setParts []string
	var args []interface{}
	argIdx := 1

	for column, value := range data {
		setParts = append(setParts, fmt.Sprintf("%s = %s", column, b.orm.placeholder(argIdx)))
		args = append(args, value)
		argIdx++
	}

	query := fmt.Sprintf("UPDATE %s SET %s", b.tableName, strings.Join(setParts, ", "))

	if len(b.whereClause) > 0 {
		query += " WHERE " + strings.Join(b.whereClause, " "+b.andOr+" ")
		args = append(args, b.whereArgs...)
	}

	return b.orm.db.ExecContext(ctx, query, args...)
}

// Delete executes a DELETE query
func (b *Builder) Delete(ctx context.Context) (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %s", b.tableName)

	var args []interface{}
	if len(b.whereClause) > 0 {
		query += " WHERE " + strings.Join(b.whereClause, " "+b.andOr+" ")
		args = append(args, b.whereArgs...)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("deleting without WHERE clause is not allowed")
	}

	return b.orm.db.ExecContext(ctx, query, args...)
}

// String returns the query string (for debugging)
func (b *Builder) String() string {
	query, _ := b.Build()
	return query
}

// Reset resets the builder for reuse
func (b *Builder) Reset() *Builder {
	b.selectColumns = nil
	b.whereClause = nil
	b.whereArgs = nil
	b.andOr = "AND"
	b.joins = nil
	b.orderBys = nil
	b.limit = 0
	b.offset = 0
	b.groupBys = nil
	b.havingClause = ""
	b.havingArgs = nil
	return b
}
