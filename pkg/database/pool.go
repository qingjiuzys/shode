// Package database 提供数据库连接池功能。
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxOpenConns     int           // 最大打开连接数
	MaxIdleConns     int           // 最大空闲连接数
	ConnMaxLifetime  time.Duration // 连接最大生命周期
	ConnMaxIdleTime  time.Duration // 连接最大空闲时间
	ConnWaitTimeout  time.Duration // 等待连接超时时间
	HealthCheck      bool          // 是否启用健康检查
	HealthCheckInterval time.Duration // 健康检查间隔
}

// DefaultConnectionPoolConfig 默认连接池配置
var DefaultConnectionPoolConfig = ConnectionPoolConfig{
	MaxOpenConns:     25,
	MaxIdleConns:     10,
	ConnMaxLifetime:  1 * time.Hour,
	ConnMaxIdleTime:  10 * time.Minute,
	ConnWaitTimeout:  30 * time.Second,
	HealthCheck:      true,
	HealthCheckInterval: 30 * time.Second,
}

// ConnectionPool 连接池
type ConnectionPool struct {
	db     *sql.DB
	config ConnectionPoolConfig
	stats  PoolStatistics
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// PoolStatistics 池统计
type PoolStatistics struct {
	OpenConnections int64
	InUse           int64
	Idle            int64
	WaitCount       int64
	WaitDuration    int64
	MaxIdleClosed   int64
	MaxLifetimeClosed int64
}

// NewConnectionPool 创建连接池
func NewConnectionPool(driver, dsn string, config ConnectionPoolConfig) (*ConnectionPool, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// 配置连接池
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	ctx, cancel := context.WithCancel(context.Background())

	pool := &ConnectionPool{
		db:     db,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}

	// 启动健康检查
	if config.HealthCheck {
		pool.wg.Add(1)
		go pool.healthCheckLoop()
	}

	return pool, nil
}

// healthCheckLoop 健康检查循环
func (p *ConnectionPool) healthCheckLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(p.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			p.healthCheck()
		}
	}
}

// healthCheck 健康检查
func (p *ConnectionPool) healthCheck() {
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()

	if err := p.db.PingContext(ctx); err != nil {
		// 健康检查失败，记录日志
		fmt.Printf("Database health check failed: %v\n", err)
	}
}

// GetDB 获取数据库连接
func (p *ConnectionPool) GetDB() *sql.DB {
	return p.db
}

// Query 查询
func (p *ConnectionPool) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, query, args...)
}

// QueryRow 查询单行
func (p *ConnectionPool) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

// Exec 执行
func (p *ConnectionPool) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

// Begin 开始事务
func (p *ConnectionPool) Begin(ctx context.Context) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, nil)
}

// BeginTx 开始事务（带选项）
func (p *ConnectionPool) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return p.db.BeginTx(ctx, opts)
}

// GetStatistics 获取统计信息
func (p *ConnectionPool) GetStatistics() PoolStatistics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := p.db.Stats()

	return PoolStatistics{
		OpenConnections:     int64(stats.OpenConnections),
		InUse:              int64(stats.InUse),
		Idle:               int64(stats.Idle),
		WaitCount:          int64(stats.WaitCount),
		WaitDuration:       int64(stats.WaitDuration),
		MaxIdleClosed:      int64(stats.MaxIdleClosed),
		MaxLifetimeClosed:  int64(stats.MaxLifetimeClosed),
	}
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.cancel()
	p.wg.Wait()
	return p.db.Close()
}

// TransactionManager 事务管理器
type TransactionManager struct {
	pool *ConnectionPool
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(pool *ConnectionPool) *TransactionManager {
	return &TransactionManager{pool: pool}
}

// RunInTransaction 在事务中运行
func (tm *TransactionManager) RunInTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // 重新抛出 panic
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// QueryBuilder 增强的查询构建器
type QueryBuilder struct {
	table      string
	selects    []string
	wheres     []WhereClause
	joins      []JoinClause
	orders     []string
	groups     []string
	having     []string
	limit      int
	offset     int
	args       []interface{}
}

// WhereClause where 子句
type WhereClause struct {
	Condition string
	Args      []interface{}
	Or        bool
}

// JoinClause join 子句
type JoinClause struct {
	Table    string
	On       string
	Args     []interface{}
	JoinType string // INNER, LEFT, RIGHT, FULL
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder(table string) *QueryBuilder {
	return &QueryBuilder{
		table:   table,
		selects: []string{"*"},
		wheres:  make([]WhereClause, 0),
		joins:   make([]JoinClause, 0),
		orders:  make([]string, 0),
		groups:  make([]string, 0),
		having:  make([]string, 0),
		args:    make([]interface{}, 0),
	}
}

// Select 选择字段
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selects = columns
	return qb
}

// Where 添加 where 条件
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	qb.wheres = append(qb.wheres, WhereClause{
		Condition: condition,
		Args:      args,
		Or:        false,
	})
	qb.args = append(qb.args, args...)
	return qb
}

// OrWhere 添加 or where 条件
func (qb *QueryBuilder) OrWhere(condition string, args ...interface{}) *QueryBuilder {
	qb.wheres = append(qb.wheres, WhereClause{
		Condition: condition,
		Args:      args,
		Or:        true,
	})
	qb.args = append(qb.args, args...)
	return qb
}

// Join 添加 join
func (qb *QueryBuilder) Join(table, on string, args ...interface{}) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Table:    table,
		On:       on,
		Args:     args,
		JoinType: "INNER",
	})
	qb.args = append(qb.args, args...)
	return qb
}

// LeftJoin 添加左连接
func (qb *QueryBuilder) LeftJoin(table, on string, args ...interface{}) *QueryBuilder {
	qb.joins = append(qb.joins, JoinClause{
		Table:    table,
		On:       on,
		Args:     args,
		JoinType: "LEFT",
	})
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy 排序
func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	qb.orders = append(qb.orders, column)
	return qb
}

// GroupBy 分组
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groups = append(qb.groups, columns...)
	return qb
}

// Limit 限制
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset 偏移
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Build 构建 SQL
func (qb *QueryBuilder) Build() (string, []interface{}) {
	var sql string

	// SELECT
	sql += fmt.Sprintf("SELECT %s FROM %s", joinColumns(qb.selects), qb.table)

	// JOIN
	for _, join := range qb.joins {
		sql += fmt.Sprintf(" %s JOIN %s ON %s", join.JoinType, join.Table, join.On)
	}

	// WHERE
	if len(qb.wheres) > 0 {
		sql += " WHERE"
		for i, where := range qb.wheres {
			if i > 0 {
				if where.Or {
					sql += " OR"
				} else {
					sql += " AND"
				}
			}
			sql += " " + where.Condition
		}
	}

	// GROUP BY
	if len(qb.groups) > 0 {
		sql += " GROUP BY " + joinColumns(qb.groups)
	}

	// ORDER BY
	if len(qb.orders) > 0 {
		sql += " ORDER BY " + joinColumns(qb.orders)
	}

	// LIMIT
	if qb.limit > 0 {
		sql += fmt.Sprintf(" LIMIT %d", qb.limit)
	}

	// OFFSET
	if qb.offset > 0 {
		sql += fmt.Sprintf(" OFFSET %d", qb.offset)
	}

	return sql, qb.args
}

// joinColumns 连接列
func joinColumns(columns []string) string {
	if len(columns) == 0 {
		return "*"
	}

	result := ""
	for i, col := range columns {
		if i > 0 {
			result += ", "
		}
		result += col
	}
	return result
}

// ReadOnlyReplica 只读副本
type ReadOnlyReplica struct {
	pool *ConnectionPool
	weight int
}

// ReadWriteSplit 读写分离
type ReadWriteSplit struct {
	master   *ConnectionPool
	replicas []*ReadOnlyReplica
	mu       sync.RWMutex
	current  uint32
}

// NewReadWriteSplit 创建读写分离
func NewReadWriteSplit(master *ConnectionPool, replicas []*ReadOnlyReplica) *ReadWriteSplit {
	return &ReadWriteSplit{
		master:   master,
		replicas: replicas,
	}
}

// GetMaster 获取主库
func (rws *ReadWriteSplit) GetMaster() *ConnectionPool {
	return rws.master
}

// GetReplica 获取从库（轮询）
func (rws *ReadWriteSplit) GetReplica() *ConnectionPool {
	rws.mu.Lock()
	defer rws.mu.Unlock()

	if len(rws.replicas) == 0 {
		return rws.master
	}

	// 简单的轮询
	index := atomic.AddUint32(&rws.current, 1) - 1
	return rws.replicas[index%uint32(len(rws.replicas))].pool
}

// Query 查询（使用从库）
func (rws *ReadWriteSplit) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return rws.GetReplica().Query(ctx, query, args...)
}

// Exec 执行（使用主库）
func (rws *ReadWriteSplit) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return rws.master.Exec(ctx, query, args...)
}
