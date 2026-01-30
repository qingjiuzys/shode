package database

import (
	"database/sql"
	"strings"
	"testing"
)

// MockModel for testing
type MockModel struct {
	ID    int64  `db:"id" auto_incr:"true"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func (m *MockModel) TableName() string {
	return "users"
}

func (m *MockModel) PrimaryKey() string {
	return "id"
}

func (m *MockModel) PrimaryKeyValue() interface{} {
	return m.ID
}

func TestGetModelInfo(t *testing.T) {
	model := &MockModel{ID: 1, Name: "John", Email: "john@example.com"}
	info := GetModelInfo(model)

	if info.TableName != "users" {
		t.Errorf("Expected table name 'users', got '%s'", info.TableName)
	}

	if info.PrimaryKey != "id" {
		t.Errorf("Expected primary key 'id', got '%s'", info.PrimaryKey)
	}

	if len(info.Fields) != 3 {
		t.Errorf("Expected 3 fields, got %d", len(info.Fields))
	}

	// Check ID field is marked as primary key and auto increment
	idField := info.Fields[0]
	if !idField.IsPK {
		t.Error("ID field should be marked as primary key")
	}
	if !idField.AutoIncr {
		t.Error("ID field should be marked as auto increment")
	}
}

func TestQueryBuilder_Build(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		Select("id", "name").
		Where("age > $1", 18).
		Where("status = $2", "active").
		OrderBy("name").
		Limit(10)

	query, args := builder.Build()

	expectedQuery := "SELECT id, name FROM users WHERE age > $1 AND status = $2 ORDER BY name LIMIT 10"
	if query != expectedQuery {
		t.Errorf("Expected query:\n%s\nGot:\n%s", expectedQuery, query)
	}

	if len(args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args))
	}
	if args[0] != 18 || args[1] != "active" {
		t.Errorf("Expected args [18 active], got %v", args)
	}
}

func TestQueryBuilder_Joins(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		Select("users.*, posts.title").
		Join("posts", "users.id = posts.user_id")

	query, args := builder.Build()

	if !stringContains(query, "INNER JOIN posts ON users.id = posts.user_id") {
		t.Errorf("Expected JOIN clause in query: %s", query)
	}

	if len(args) != 0 {
		t.Errorf("Expected 0 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		WhereIn("id", []interface{}{1, 2, 3})

	query, args := builder.Build()

	expectedQuery := "SELECT * FROM users WHERE id IN ($1, $2, $3)"
	if query != expectedQuery {
		t.Errorf("Expected query:\n%s\nGot:\n%s", expectedQuery, query)
	}

	if len(args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(args))
	}
}

func TestQueryBuilder_WhereLike(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		WhereLike("name", "%John%")

	query, args := builder.Build()

	if !stringContains(query, "name LIKE") {
		t.Errorf("Expected LIKE clause in query: %s", query)
	}

	if len(args) != 1 || args[0] != "%John%" {
		t.Errorf("Expected args [%%John%%], got %v", args)
	}
}

func TestQueryBuilder_WhereBetween(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		WhereBetween("age", 18, 65)

	query, args := builder.Build()

	if !stringContains(query, "age BETWEEN") {
		t.Errorf("Expected BETWEEN clause in query: %s", query)
	}

	if len(args) != 2 || args[0] != 18 || args[1] != 65 {
		t.Errorf("Expected args [18 65], got %v", args)
	}
}

func TestQueryBuilder_GroupByHaving(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		Select("role, COUNT(*) as count").
		GroupBy("role").
		Having("COUNT(*) > ?", 5)

	query, args := builder.Build()

	if !stringContains(query, "GROUP BY role") {
		t.Errorf("Expected GROUP BY clause in query: %s", query)
	}

	if !stringContains(query, "HAVING COUNT(*) >") {
		t.Errorf("Expected HAVING clause in query: %s", query)
	}

	if len(args) != 1 || args[0] != 5 {
		t.Errorf("Expected args [5], got %v", args)
	}
}

func TestQueryBuilder_Reset(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		Where("id = ?", 1).
		Limit(10)

	// Reset and build new query
	builder.Reset().
		Where("name = ?", "John")

	query, args := builder.Build()

	if stringContains(query, "LIMIT") {
		t.Error("Expected no LIMIT after reset")
	}

	if !stringContains(query, "name =") {
		t.Error("Expected name condition after reset")
	}

	if len(args) != 1 || args[0] != "John" {
		t.Errorf("Expected args [John], got %v", args)
	}
}

func TestPlaceholder_Postgres(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	if orm.placeholder(1) != "$1" {
		t.Errorf("Expected $1 for postgres, got %s", orm.placeholder(1))
	}
	if orm.placeholder(3) != "$3" {
		t.Errorf("Expected $3 for postgres, got %s", orm.placeholder(3))
	}
}

func TestPlaceholder_SQLite(t *testing.T) {
	orm := &ORM{driver: "sqlite3"}
	if orm.placeholder(1) != "?" {
		t.Errorf("Expected ? for sqlite, got %s", orm.placeholder(1))
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("postgres")

	if config.Driver != "postgres" {
		t.Errorf("Expected driver postgres, got %s", config.Driver)
	}

	if config.MaxOpenConns != 25 {
		t.Errorf("Expected MaxOpenConns 25, got %d", config.MaxOpenConns)
	}

	if config.MaxIdleConns != 5 {
		t.Errorf("Expected MaxIdleConns 5, got %d", config.MaxIdleConns)
	}
}

// Test with mock database
type MockDB struct {
	*sql.DB
}

// stringContains is a helper function to check if a string stringContains a substring
func stringContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestQueryBuilder_String(t *testing.T) {
	orm := &ORM{driver: "postgres"}
	builder := orm.QueryBuilder("users").
		Where("id = ?", 1).
		OrderBy("name")

	queryStr := builder.String()

	if !stringContains(queryStr, "SELECT") || !stringContains(queryStr, "FROM users") {
		t.Errorf("Expected SELECT query in string: %s", queryStr)
	}
}
