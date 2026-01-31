// Package database 提供数据库迁移功能。
package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Migration 迁移
type Migration struct {
	Version     int
	Name        string
	Description string
	Up          string
	Down        string
	AppliedAt   *time.Time
}

// Migrator 迁移器
type Migrator struct {
	db          *sql.DB
	tableName   string
	migrations  []*Migration
	mu          sync.Mutex
}

// NewMigrator 创建迁移器
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{
		db:        db,
		tableName: "schema_migrations",
		migrations: make([]*Migration, 0),
	}
}

// SetTableName 设置迁移表名
func (m *Migrator) SetTableName(name string) {
	m.tableName = name
}

// AddMigration 添加迁移
func (m *Migrator) AddMigration(version int, name, description, up, down string) {
	m.migrations = append(m.migrations, &Migration{
		Version:     version,
		Name:        name,
		Description: description,
		Up:          up,
		Down:        down,
	})
}

// Init 初始化迁移表
func (m *Migrator) Init(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			version INTEGER PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			applied_at TIMESTAMP NOT NULL
		);
	`, m.tableName)

	_, err := m.db.ExecContext(ctx, createTableSQL)
	return err
}

// Up 执行所有未应用的迁移
func (m *Migrator) Up(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 确保迁移表存在
	if err := m.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	// 获取已应用的迁移
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 排序迁移
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// 执行未应用的迁移
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			continue
		}

		if err := m.applyMigration(ctx, migration); err != nil {
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		fmt.Printf("Applied migration %d: %s\n", migration.Version, migration.Name)
	}

	return nil
}

// Down 回滚最后一个迁移
func (m *Migrator) Down(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取已应用的迁移
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		return fmt.Errorf("no migrations to rollback")
	}

	// 找到最后一个应用的迁移
	var lastMigration *Migration
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			lastMigration = migration
		}
	}

	if lastMigration == nil {
		return fmt.Errorf("no migrations to rollback")
	}

	// 执行回滚
	if err := m.rollbackMigration(ctx, lastMigration); err != nil {
		return fmt.Errorf("failed to rollback migration %d: %w", lastMigration.Version, err)
	}

	fmt.Printf("Rolled back migration %d: %s\n", lastMigration.Version, lastMigration.Name)

	return nil
}

// applyMigration 应用迁移
func (m *Migrator) applyMigration(ctx context.Context, migration *Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// 执行迁移 SQL
	if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// 记录迁移
	now := time.Now()
	insertSQL := fmt.Sprintf(`
		INSERT INTO %s (version, name, description, applied_at)
		VALUES (?, ?, ?, ?)
	`, m.tableName)

	if _, err := tx.ExecContext(ctx, insertSQL, migration.Version, migration.Name, migration.Description, now); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return tx.Commit()
}

// rollbackMigration 回滚迁移
func (m *Migrator) rollbackMigration(ctx context.Context, migration *Migration) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// 执行回滚 SQL
	if _, err := tx.ExecContext(ctx, migration.Down); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	// 删除迁移记录
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE version = ?", m.tableName)
	if _, err := tx.ExecContext(ctx, deleteSQL, migration.Version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete migration record: %w", err)
	}

	return tx.Commit()
}

// getAppliedVersions 获取已应用的迁移版本
func (m *Migrator) getAppliedVersions(ctx context.Context) (map[int]bool, error) {
	query := fmt.Sprintf("SELECT version FROM %s ORDER BY version", m.tableName)

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	versions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		versions[version] = true
	}

	return versions, rows.Err()
}

// Status 获取迁移状态
func (m *Migrator) Status(ctx context.Context) ([]*Migration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取已应用的迁移
	applied, err := m.getAppliedVersions(ctx)
	if err != nil {
		return nil, err
	}

	// 标记已应用的迁移
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			migration.AppliedAt = &[]time.Time{time.Now()}[0]
		}
	}

	return m.migrations, nil
}

// LoadMigrationsFromDir 从目录加载迁移文件
func (m *Migrator) LoadMigrationsFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// 解析迁移文件
		// 文件名格式: 001_create_users.up.sql 或 001_create_users.down.sql
		filename := filepath.Base(file)
		parts := strings.Split(filename, "_")

		if len(parts) < 2 {
			continue
		}

		// 解析版本号
		var version int
		fmt.Sscanf(parts[0], "%d", &version)

		// 解析名称
		name := strings.Join(parts[1:len(parts)-1], "_")
		extension := parts[len(parts)-1]

		// 解析类型
		migrationType := strings.Split(extension, ".")[0] // up 或 down

		// 查找或创建迁移
		var migration *Migration
		for _, m := range m.migrations {
			if m.Version == version {
				migration = m
				break
			}
		}

		if migration == nil {
			migration = &Migration{
				Version: version,
				Name:    name,
			}
			m.migrations = append(m.migrations, migration)
		}

		// 设置 SQL
		if migrationType == "up" {
			migration.Up = string(content)
		} else if migrationType == "down" {
			migration.Down = string(content)
		}
	}

	return nil
}

// CreateMigration 创建新的迁移文件
func (m *Migrator) CreateMigration(dir, name string) error {
	// 获取下一个版本号
	version := len(m.migrations) + 1

	// 创建 up 文件
	upFilename := fmt.Sprintf("%03d_%s.up.sql", version, name)
	upPath := filepath.Join(dir, upFilename)
	upContent := fmt.Sprintf("-- Migration: %s\n-- Version: %d\n\n", name, version)

	if err := os.WriteFile(upPath, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration: %w", err)
	}

	// 创建 down 文件
	downFilename := fmt.Sprintf("%03d_%s.down.sql", version, name)
	downPath := filepath.Join(dir, downFilename)
	downContent := fmt.Sprintf("-- Rollback: %s\n-- Version: %d\n\n", name, version)

	if err := os.WriteFile(downPath, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration: %w", err)
	}

	fmt.Printf("Created migration files:\n  %s\n  %s\n", upFilename, downFilename)

	return nil
}
