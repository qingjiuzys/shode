// Package fixtures 提供测试夹具功能
package fixtures

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// FixtureManager 夹具管理器
type FixtureManager struct {
	t          *testing.T
	basePath   string
	data       map[string]interface{}
	db         *sql.DB
	loaded     map[string]bool
	collections map[string]*Collection
}

// Collection 集合
type Collection struct {
	name  string
	items []interface{}
}

// New 创建夹具管理器
func New(t *testing.T) *FixtureManager {
	return &FixtureManager{
		t:          t,
		basePath:   "./testdata/fixtures",
		data:       make(map[string]interface{}),
		loaded:     make(map[string]bool),
		collections: make(map[string]*Collection),
	}
}

// SetBasePath 设置基础路径
func (f *FixtureManager) SetBasePath(path string) {
	f.basePath = path
}

// SetDB 设置数据库
func (f *FixtureManager) SetDB(db *sql.DB) {
	f.db = db
}

// Load 加载夹具
func (f *FixtureManager) Load(name string) error {
	if f.loaded[name] {
		return nil
	}

	filePath := filepath.Join(f.basePath, name+".json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read fixture file %s: %w", filePath, err)
	}

	var fixture interface{}
	if err := json.Unmarshal(data, &fixture); err != nil {
		return fmt.Errorf("failed to parse fixture file %s: %w", filePath, err)
	}

	f.data[name] = fixture
	f.loaded[name] = true

	return nil
}

// MustLoad 必须加载夹具
func (f *FixtureManager) MustLoad(name string) {
	if err := f.Load(name); err != nil {
		f.t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
}

// Get 获取夹具数据
func (f *FixtureManager) Get(name string) interface{} {
	if !f.loaded[name] {
		f.MustLoad(name)
	}

	return f.data[name]
}

// GetAs 获取夹具数据并转换类型
func (f *FixtureManager) GetAs(name string, v interface{}) error {
	data := f.Get(name)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return json.Unmarshal(jsonData, v)
}

// MustGetAs 必须获取并转换
func (f *FixtureManager) MustGetAs(name string, v interface{}) {
	if err := f.GetAs(name, v); err != nil {
		f.t.Fatalf("Failed to convert fixture %s: %v", name, err)
	}
}

// Collection 获取或创建集合
func (f *FixtureManager) Collection(name string) *Collection {
	if f.collections[name] == nil {
		f.collections[name] = &Collection{
			name:  name,
			items: make([]interface{}, 0),
		}
	}

	return f.collections[name]
}

// LoadToDB 加载到数据库
func (f *FixtureManager) LoadToDB(tableName string) error {
	if f.db == nil {
		return fmt.Errorf("database not set")
	}

	collection := f.Collection(tableName)
	for _, item := range collection.items {
		if err := f.insertToDB(tableName, item); err != nil {
			return err
		}
	}

	return nil
}

// insertToDB 插入到数据库
func (f *FixtureManager) insertToDB(table string, item interface{}) error {
	// 简化实现，实际应该根据item结构生成INSERT语句
	query := fmt.Sprintf("INSERT INTO %s VALUES (...)")

	_, err := f.db.Exec(query)
	return err
}

// Reset 重置夹具
func (f *FixtureManager) Reset() {
	f.data = make(map[string]interface{})
	f.loaded = make(map[string]bool)
	f.collections = make(map[string]*Collection)
}

// Add 添加到集合
func (c *Collection) Add(item interface{}) {
	c.items = append(c.items, item)
}

// Count 获取数量
func (c *Collection) Count() int {
	return len(c.items)
}

// Get 获取指定索引的项
func (c *Collection) Get(index int) interface{} {
	if index < 0 || index >= len(c.items) {
		return nil
	}
	return c.items[index]
}

// All 获取所有项
func (c *Collection) All() []interface{} {
	return c.items
}

// Clean 清理集合
func (c *Collection) Clean() {
	c.items = make([]interface{}, 0)
}

// DatabaseFixture 数据库夹具
type DatabaseFixture struct {
	db       *sql.DB
	t        *testing.T
	tearDown []func() error
}

// NewDatabaseFixture 创建数据库夹具
func NewDatabaseFixture(t *testing.T, db *sql.DB) *DatabaseFixture {
	return &DatabaseFixture{
		db:       db,
		t:        t,
		tearDown: make([]func() error, 0),
	}
}

// Setup 设置数据库
func (d *DatabaseFixture) Setup(queries ...string) {
	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			d.t.Fatalf("Failed to execute setup query: %v\nQuery: %s", err, query)
		}
	}
}

// Teardown 清理数据库
func (d *DatabaseFixture) Teardown() {
	for i := len(d.tearDown) - 1; i >= 0; i-- {
		if err := d.tearDown[i](); err != nil {
			d.t.Logf("Warning: teardown failed: %v", err)
		}
	}
}

// AddTeardown 添加清理函数
func (d *DatabaseFixture) AddTeardown(fn func() error) {
	d.tearDown = append(d.tearDown, fn)
}

// Table 表夹具
type Table struct {
	name string
	db   *sql.DB
	t    *testing.T
}

// NewTable 创建表夹具
func NewTable(t *testing.T, db *sql.DB, name string) *Table {
	return &Table{
		name: name,
		db:   db,
		t:    t,
	}
}

// Create 创建表
func (tbl *Table) Create(schema string) {
	if _, err := tbl.db.Exec(schema); err != nil {
		tbl.t.Fatalf("Failed to create table %s: %v", tbl.name, err)
	}
}

// Drop 删除表
func (tbl *Table) Drop() {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tbl.name)
	if _, err := tbl.db.Exec(query); err != nil {
		tbl.t.Logf("Warning: failed to drop table %s: %v", tbl.name, err)
	}
}

// Truncate 清空表
func (tbl *Table) Truncate() {
	query := fmt.Sprintf("DELETE FROM %s", tbl.name)
	if _, err := tbl.db.Exec(query); err != nil {
		tbl.t.Fatalf("Failed to truncate table %s: %v", tbl.name, err)
	}
}

// Insert 插入数据
func (tbl *Table) Insert(data map[string]interface{}) {
	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		columns = append(columns, col)
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tbl.name,
		joinStrings(columns, ", "),
		joinStrings(placeholders, ", "))

	if _, err := tbl.db.Exec(query, values...); err != nil {
		tbl.t.Fatalf("Failed to insert into %s: %v", tbl.name, err)
	}
}

// Count 统计行数
func (tbl *Table) Count() int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl.name)
	if err := tbl.db.QueryRow(query).Scan(&count); err != nil {
		tbl.t.Fatalf("Failed to count %s: %v", tbl.name, err)
	}
	return count
}

// Exists 检查数据是否存在
func (tbl *Table) Exists(where string, args ...interface{}) bool {
	query := fmt.Sprintf("SELECT 1 FROM %s WHERE %s LIMIT 1", tbl.name, where)
	var exists int
	err := tbl.db.QueryRow(query, args...).Scan(&exists)
	return err == nil && exists == 1
}

// joinStrings 连接字符串
func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
