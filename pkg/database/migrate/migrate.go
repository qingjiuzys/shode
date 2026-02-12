// Package migrate 提供数据库迁移功能
package migrate

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Migration 迁移
type Migration struct {
	Version uint
	Name    string
	Up      string
	Down    string
}

// Migrator 迁移器
type Migrator struct {
	db          *sql.DB
	dialect     string
	tableName   string
	migrations  []*Migration
	currentVer  uint
	targetVer   uint
}

// Config 迁移配置
type Config struct {
	DB        *sql.DB
	Dialect   string
	TableName string
}

// NewMigrator 创建迁移器
func NewMigrator(config *Config) *Migrator {
	tableName := config.TableName
	if tableName == "" {
		tableName = "schema_migrations"
	}

	return &Migrator{
		db:         config.DB,
		dialect:    config.Dialect,
		tableName:  tableName,
		migrations: make([]*Migration, 0),
	}
}

// AddMigration 添加迁移
func (m *Migrator) AddMigration(version uint, name, up, down string) {
	m.migrations = append(m.migrations, &Migration{
		Version: version,
		Name:    name,
		Up:      up,
		Down:    down,
	})
}

// AddMigrationSQL 添加SQL迁移
func (m *Migrator) AddMigrationSQL(version uint, name, upSQL, downSQL string) {
	m.AddMigration(version, name, upSQL, downSQL)
}

// LoadMigrationsFromDir 从目录加载迁移
func (m *Migrator) LoadMigrationsFromDir(dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// 解析迁移文件: 001_initial.up.sql, 001_initial.down.sql
	pattern := regexp.MustCompile(`^(\d+)_([^\.]+)\.(up|down)\.sql$`)

	migrationMap := make(map[uint]*Migration)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		matches := pattern.FindStringSubmatch(file.Name())
		if matches == nil {
			continue
		}

		version, _ := strconv.ParseUint(matches[1], 10, 32)
		name := matches[2]
		direction := matches[3]

		filePath := filepath.Join(dir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
		}

		migration, ok := migrationMap[uint(version)]
		if !ok {
			migration = &Migration{
				Version: uint(version),
				Name:    name,
			}
			migrationMap[uint(version)] = migration
		}

		if direction == "up" {
			migration.Up = string(content)
		} else {
			migration.Down = string(content)
		}
	}

	// 转换为切片并排序
	for _, migration := range migrationMap {
		m.migrations = append(m.migrations, migration)
	}

	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	return nil
}

// Init 初始化迁移表
func (m *Migrator) Init() error {
	var createSQL string

	switch m.dialect {
	case "mysql":
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version BIGINT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`, m.tableName)
	case "postgres", "postgresql":
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version BIGINT PRIMARY KEY,
				name VARCHAR(255) NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`, m.tableName)
	case "sqlite3", "sqlite":
		createSQL = fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version INTEGER PRIMARY KEY,
				name TEXT NOT NULL,
				applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`, m.tableName)
	default:
		return fmt.Errorf("unsupported dialect: %s", m.dialect)
	}

	_, err := m.db.Exec(createSQL)
	return err
}

// CurrentVersion 获取当前版本
func (m *Migrator) CurrentVersion() (uint, error) {
	row := m.db.QueryRow(fmt.Sprintf("SELECT COALESCE(MAX(version), 0) FROM %s", m.tableName))

	var version uint
	err := row.Scan(&version)
	if err != nil {
		return 0, err
	}

	m.currentVer = version
	return version, nil
}

// Pending 获取待执行的迁移
func (m *Migrator) Pending() ([]*Migration, error) {
	current, err := m.CurrentVersion()
	if err != nil {
		return nil, err
	}

	var pending []*Migration
	for _, migration := range m.migrations {
		if migration.Version > current {
			pending = append(pending, migration)
		}
	}

	return pending, nil
}

// Up 执行所有待执行的迁移
func (m *Migrator) Up() error {
	pending, err := m.Pending()
	if err != nil {
		return err
	}

	for _, migration := range pending {
		if err := m.UpTo(migration.Version); err != nil {
			return err
		}
	}

	return nil
}

// UpTo 迁移到指定版本
func (m *Migrator) UpTo(version uint) error {
	current, err := m.CurrentVersion()
	if err != nil {
		return err
	}

	if version <= current {
		return fmt.Errorf("target version %d is not greater than current version %d", version, current)
	}

	for _, migration := range m.migrations {
		if migration.Version > current && migration.Version <= version {
			if err := m.applyMigration(migration, "up"); err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
			}
			fmt.Printf("✓ Applied migration %d: %s\n", migration.Version, migration.Name)
		}
	}

	return nil
}

// Down 回滚最近的迁移
func (m *Migrator) Down() error {
	current, err := m.CurrentVersion()
	if err != nil {
		return err
	}

	if current == 0 {
		return fmt.Errorf("no migration to rollback")
	}

	return m.DownTo(current - 1)
}

// DownTo 回滚到指定版本
func (m *Migrator) DownTo(version uint) error {
	current, err := m.CurrentVersion()
	if err != nil {
		return err
	}

	if version >= current {
		return fmt.Errorf("target version %d is not less than current version %d", version, current)
	}

	// 从后往前查找需要回滚的迁移
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if migration.Version <= current && migration.Version > version {
			if err := m.applyMigration(migration, "down"); err != nil {
				return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
			}
			fmt.Printf("✓ Rolled back migration %d: %s\n", migration.Version, migration.Name)
		}
	}

	return nil
}

// applyMigration 应用迁移
func (m *Migrator) applyMigration(migration *Migration, direction string) error {
	// 开始事务
	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 执行迁移SQL
	var sql string
	if direction == "up" {
		sql = migration.Up
	} else {
		sql = migration.Down
	}

	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// 更新迁移记录
	if direction == "up" {
		// 插入记录
		insertSQL := fmt.Sprintf("INSERT INTO %s (version, name, applied_at) VALUES (?, ?, ?)", m.tableName)
		if _, err := tx.Exec(insertSQL, migration.Version, migration.Name, time.Now()); err != nil {
			return err
		}
	} else {
		// 删除记录
		deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName)
		if _, err := tx.Exec(deleteSQL, migration.Version); err != nil {
			return err
		}
	}

	// 提交事务
	return tx.Commit()
}

// Status 获取迁移状态
func (m *Migrator) Status() (*Status, error) {
	current, err := m.CurrentVersion()
	if err != nil {
		return nil, err
	}

	pending, err := m.Pending()
	if err != nil {
		return nil, err
	}

	applied := make([]*Migration, 0)
	for _, migration := range m.migrations {
		if migration.Version <= current {
			applied = append(applied, migration)
		}
	}

	return &Status{
		Current: current,
		Applied: applied,
		Pending: pending,
	}, nil
}

// Status 迁移状态
type Status struct {
	Current uint
	Applied []*Migration
	Pending []*Migration
}

// PrintStatus 打印状态
func (m *Migrator) PrintStatus() error {
	status, err := m.Status()
	if err != nil {
		return err
	}

	fmt.Printf("\n=== Migration Status ===\n\n")
	fmt.Printf("Current Version: %d\n\n", status.Current)

	fmt.Printf("Applied Migrations (%d):\n", len(status.Applied))
	for _, migration := range status.Applied {
		fmt.Printf("  %d: %s\n", migration.Version, migration.Name)
	}

	fmt.Printf("\nPending Migrations (%d):\n", len(status.Pending))
	for _, migration := range status.Pending {
		fmt.Printf("  %d: %s\n", migration.Version, migration.Name)
	}

	fmt.Println("\n========================")
	return nil
}

// Create 创建新的迁移文件
func (m *Migrator) Create(dir, name string) error {
	// 获取下一个版本号
	nextVersion := uint(len(m.migrations) + 1)

	// 格式化版本号
	versionStr := fmt.Sprintf("%03d", nextVersion)

	// 创建文件名
	upFileName := fmt.Sprintf("%s_%s.up.sql", versionStr, strings.ToLower(strings.ReplaceAll(name, " ", "_")))
	downFileName := fmt.Sprintf("%s_%s.down.sql", versionStr, strings.ToLower(strings.ReplaceAll(name, " ", "_")))

	// 创建目录
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 创建 up 文件
	upPath := filepath.Join(dir, upFileName)
	upContent := fmt.Sprintf("-- Migration: %s\n-- Version: %d\n-- Up\n\n", name, nextVersion)
	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return err
	}

	// 创建 down 文件
	downPath := filepath.Join(dir, downFileName)
	downContent := fmt.Sprintf("-- Migration: %s\n-- Version: %d\n-- Down\n\n", name, nextVersion)
	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return err
	}

	fmt.Printf("✓ Created migration files:\n")
	fmt.Printf("  - %s\n", upPath)
	fmt.Printf("  - %s\n", downPath)

	return nil
}

// Validate 验证迁移
func (m *Migrator) Validate() error {
	// 检查版本号重复
	versions := make(map[uint]bool)
	for _, migration := range m.migrations {
		if versions[migration.Version] {
			return fmt.Errorf("duplicate version number: %d", migration.Version)
		}
		versions[migration.Version] = true
	}

	// 检查迁移完整性
	for _, migration := range m.migrations {
		if migration.Up == "" {
			return fmt.Errorf("migration %d (%s) missing up SQL", migration.Version, migration.Name)
		}
		if migration.Down == "" {
			return fmt.Errorf("migration %d (%s) missing down SQL", migration.Version, migration.Name)
		}
	}

	return nil
}

// Redo 重做最后一次迁移
func (m *Migrator) Redo() error {
	if err := m.Down(); err != nil {
		return err
	}
	return m.Up()
}

// Reset 重置所有迁移
func (m *Migrator) Reset() error {
	// 回滚所有迁移
	for i := len(m.migrations) - 1; i >= 0; i-- {
		migration := m.migrations[i]
		if err := m.applyMigration(migration, "down"); err != nil {
			return fmt.Errorf("failed to rollback migration %d: %w", migration.Version, err)
		}
	}

	// 重新应用所有迁移
	return m.Up()
}

// GetMigration 获取指定版本的迁移
func (m *Migrator) GetMigration(version uint) (*Migration, bool) {
	for _, migration := range m.migrations {
		if migration.Version == version {
			return migration, true
		}
	}
	return nil, false
}

// Version 获取目标版本
func (m *Migrator) Version() uint {
	return m.targetVer
}

// SetVersion 设置目标版本
func (m *Migrator) SetVersion(version uint) {
	m.targetVer = version
}
