// Package transaction 提供分布式事务功能。
package transaction

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Transaction 事务接口
type Transaction interface {
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	Status() TransactionStatus
}

// TransactionStatus 事务状态
type TransactionStatus int

const (
	StatusActive TransactionStatus = iota
	StatusCommitted
	StatusRolledBack
	StatusUnknown
	StatusTimedOut
)

// String 返回状态字符串
func (s TransactionStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusCommitted:
		return "committed"
	case StatusRolledBack:
		return "rolled_back"
	case StatusUnknown:
		return "unknown"
	case StatusTimedOut:
		return "timed_out"
	default:
		return "unknown"
	}
}

// TwoPhaseCommit 两阶段提交
type TwoPhaseCommit struct {
	participants map[string]*Participant
	coordinator  *Coordinator
	timeout      time.Duration
	mu           sync.RWMutex
}

// Participant 参与者
type Participant struct {
	ID       string
	Endpoint string
	Status   ParticipantStatus
	Prepare  func(ctx context.Context) error
	Commit   func(ctx context.Context) error
	Rollback func(ctx context.Context) error
}

// ParticipantStatus 参与者状态
type ParticipantStatus int

const (
	ParticipantReady ParticipantStatus = iota
	ParticipantPrepared
	ParticipantCommitted
	ParticipantAborted
)

// Coordinator 协调器
type Coordinator struct {
	transactionID string
	log           *TransactionLog
}

// TransactionLog 事务日志
type TransactionLog struct {
	entries []LogEntry
	mu      sync.Mutex
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time
	Phase     string
	ParticipantID string
	Status    string
}

// NewTwoPhaseCommit 创建两阶段提交
func NewTwoPhaseCommit(timeout time.Duration) *TwoPhaseCommit {
	return &TwoPhaseCommit{
		participants: make(map[string]*Participant),
		timeout:      timeout,
		coordinator: &Coordinator{
			log: &TransactionLog{
				entries: make([]LogEntry, 0),
			},
		},
	}
}

// AddParticipant 添加参与者
func (tpc *TwoPhaseCommit) AddParticipant(participant *Participant) {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()

	tpc.participants[participant.ID] = participant
}

// Begin 开始事务
func (tpc *TwoPhaseCommit) Begin(ctx context.Context) error {
	tpc.coordinator.transactionID = generateTransactionID()
	tpc.coordinator.log.Append(LogEntry{
		Timestamp: time.Now(),
		Phase:     "begin",
		Status:    "started",
	})

	return nil
}

// Prepare 准备阶段
func (tpc *TwoPhaseCommit) Prepare(ctx context.Context) error {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()

	// 向所有参与者发送准备请求
	for _, participant := range tpc.participants {
		if err := participant.Prepare(ctx); err != nil {
			// 准备失败，回滚所有已准备的参与者
			tpc.rollbackPrepared(ctx)
			return fmt.Errorf("participant %s failed to prepare: %w", participant.ID, err)
		}
		participant.Status = ParticipantPrepared
	}

	return nil
}

// Commit 提交阶段
func (tpc *TwoPhaseCommit) Commit(ctx context.Context) error {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()

	// 记录提交日志
	tpc.coordinator.log.Append(LogEntry{
		Timestamp: time.Now(),
		Phase:     "commit",
		Status:    "started",
	})

	// 向所有参与者发送提交请求
	for _, participant := range tpc.participants {
		if err := participant.Commit(ctx); err != nil {
			// 提交失败，记录但不回滚（已提交的不能回滚）
			tpc.coordinator.log.Append(LogEntry{
				Timestamp: time.Now(),
				Phase:     "commit",
				ParticipantID: participant.ID,
				Status:    "failed",
			})
			return fmt.Errorf("participant %s failed to commit: %w", participant.ID, err)
		}
		participant.Status = ParticipantCommitted
	}

	return nil
}

// Rollback 回滚
func (tpc *TwoPhaseCommit) Rollback(ctx context.Context) error {
	tpc.mu.Lock()
	defer tpc.mu.Unlock()

	for _, participant := range tpc.participants {
		if participant.Status == ParticipantPrepared {
			if err := participant.Rollback(ctx); err != nil {
				// 回滚失败，记录错误
				tpc.coordinator.log.Append(LogEntry{
					Timestamp: time.Now(),
					Phase:     "rollback",
					ParticipantID: participant.ID,
					Status:    "failed",
				})
			}
			participant.Status = ParticipantAborted
		}
	}

	return nil
}

// rollbackPrepared 回滚已准备的参与者
func (tpc *TwoPhaseCommit) rollbackPrepared(ctx context.Context) {
	for _, participant := range tpc.participants {
		if participant.Status == ParticipantPrepared {
			participant.Rollback(ctx)
			participant.Status = ParticipantAborted
		}
	}
}

// Status 获取状态
func (tpc *TwoPhaseCommit) Status() TransactionStatus {
	// 简化实现
	return StatusActive
}

// Append 追加日志
func (tl *TransactionLog) Append(entry LogEntry) {
	tl.mu.Lock()
	defer tl.mu.Unlock()
	tl.entries = append(tl.entries, entry)
}

// ThreePhaseCommit 三阶段提交
type ThreePhaseCommit struct {
	participants map[string]*Participant
	canCommit    bool
	phase        int
	mu           sync.RWMutex
}

// NewThreePhaseCommit 创建三阶段提交
func NewThreePhaseCommit() *ThreePhaseCommit {
	return &ThreePhaseCommit{
		participants: make(map[string]*Participant),
	}
}

// AddParticipant 添加参与者
func (stc *ThreePhaseCommit) AddParticipant(participant *Participant) {
	stc.mu.Lock()
	defer stc.mu.Unlock()
	stc.participants[participant.ID] = participant
}

// Begin 开始事务
func (stc *ThreePhaseCommit) Begin(ctx context.Context) error {
	stc.phase = 1
	return nil
}

// PrepareCanCommit 准备可以提交
func (stc *ThreePhaseCommit) PrepareCanCommit(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	// 阶段 1: CanCommit?
	stc.canCommit = true
	for _, participant := range stc.participants {
		if err := participant.Prepare(ctx); err != nil {
			stc.canCommit = false
			break
		}
		participant.Status = ParticipantPrepared
	}

	stc.phase = 2
	return nil
}

// PreCommit 预提交
func (stc *ThreePhaseCommit) PreCommit(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	if !stc.canCommit {
		return fmt.Errorf("cannot commit")
	}

	// 阶段 2: PreCommit
	// 简化实现，直接进入下一阶段
	stc.phase = 3
	return nil
}

// DoCommit 执行提交
func (stc *ThreePhaseCommit) DoCommit(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	// 阶段 3: DoCommit
	for _, participant := range stc.participants {
		if err := participant.Commit(ctx); err != nil {
			return err
		}
		participant.Status = ParticipantCommitted
	}

	return nil
}

// Rollback 回滚
func (stc *ThreePhaseCommit) Rollback(ctx context.Context) error {
	stc.mu.Lock()
	defer stc.mu.Unlock()

	for _, participant := range stc.participants {
		participant.Rollback(ctx)
		participant.Status = ParticipantAborted
	}

	return nil
}

// Status 获取状态
func (stc *ThreePhaseCommit) Status() TransactionStatus {
	return StatusActive
}

// SagaTransaction Saga 事务
type SagaTransaction struct {
	steps      []*SagaStep
	current    int
	compensating bool
	mu         sync.Mutex
}

// SagaStep Saga 步骤
type SagaStep struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
	Status     string
}

// NewSagaTransaction 创建 Saga 事务
func NewSagaTransaction(steps ...*SagaStep) *SagaTransaction {
	return &SagaTransaction{
		steps: steps,
		current: 0,
	}
}

// Execute 执行事务
func (st *SagaTransaction) Execute(ctx context.Context) error {
	st.mu.Lock()
	defer st.mu.Unlock()

	for i := st.current; i < len(st.steps); i++ {
		step := st.steps[i]
		if err := step.Execute(ctx); err != nil {
			step.Status = "failed"
			// 执行补偿
			st.compensate(ctx, i)
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
		step.Status = "completed"
		st.current = i + 1
	}

	return nil
}

// compensate 补偿
func (st *SagaTransaction) compensate(ctx context.Context, failedStep int) {
	st.compensating = true
	for i := failedStep - 1; i >= 0; i-- {
		step := st.steps[i]
		if step.Compensate != nil {
			if err := step.Compensate(ctx); err != nil {
				// 补偿失败，记录错误
				step.Status = "compensation_failed"
			} else {
				step.Status = "compensated"
			}
		}
	}
}

// Status 获取状态
func (st *SagaTransaction) Status() TransactionStatus {
	return StatusActive
}

// TCCTransaction TCC 事务
type TCCTransaction struct {
	actions map[string]*TCCAction
	mu      sync.RWMutex
}

// TCCAction TCC 动作
type TCCAction struct {
	Name      string
	Try       func(ctx context.Context) error
	Confirm   func(ctx context.Context) error
	Cancel    func(ctx context.Context) error
	Status    string
}

// NewTCCTransaction 创建 TCC 事务
func NewTCCTransaction() *TCCTransaction {
	return &TCCTransaction{
		actions: make(map[string]*TCCAction),
	}
}

// Register 注册动作
func (tcc *TCCTransaction) Register(action *TCCAction) {
	tcc.mu.Lock()
	defer tcc.mu.Unlock()
	tcc.actions[action.Name] = action
}

// Try 尝试阶段
func (tcc *TCCTransaction) Try(ctx context.Context) error {
	tcc.mu.Lock()
	defer tcc.mu.Unlock()

	for _, action := range tcc.actions {
		if err := action.Try(ctx); err != nil {
			action.Status = "failed"
			// 取消所有已成功的 Try
			tcc.cancel(ctx)
			return err
		}
		action.Status = "tried"
	}

	return nil
}

// Confirm 确认阶段
func (tcc *TCCTransaction) Confirm(ctx context.Context) error {
	tcc.mu.Lock()
	defer tcc.mu.Unlock()

	for _, action := range tcc.actions {
		if action.Status != "tried" {
			continue
		}

		if err := action.Confirm(ctx); err != nil {
			action.Status = "confirm_failed"
			return err
		}
		action.Status = "confirmed"
	}

	return nil
}

// Cancel 取消阶段
func (tcc *TCCTransaction) Cancel(ctx context.Context) error {
	tcc.mu.Lock()
	defer tcc.mu.Unlock()

	for _, action := range tcc.actions {
		if action.Status != "tried" {
			continue
		}

		if err := action.Cancel(ctx); err != nil {
			action.Status = "cancel_failed"
			return err
		}
		action.Status = "cancelled"
	}

	return nil
}

// cancel 内部取消
func (tcc *TCCTransaction) cancel(ctx context.Context) {
	for _, action := range tcc.actions {
		if action.Status == "tried" {
			action.Cancel(ctx)
			action.Status = "cancelled"
		}
	}
}

// Status 获取状态
func (tcc *TCCTransaction) Status() TransactionStatus {
	return StatusActive
}

// DistributedLock 分布式锁
type DistributedLock struct {
	locks     map[string]*Lock
	timeout   time.Duration
	mu        sync.RWMutex
}

// Lock 锁
type Lock struct {
	Key       string
	Owner     string
	ExpiresAt time.Time
}

// NewDistributedLock 创建分布式锁
func NewDistributedLock(timeout time.Duration) *DistributedLock {
	return &DistributedLock{
		locks:   make(map[string]*Lock),
		timeout: timeout,
	}
}

// Acquire 获取锁
func (dl *DistributedLock) Acquire(ctx context.Context, key, owner string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// 检查是否已存在
	if lock, exists := dl.locks[key]; exists {
		if time.Now().Before(lock.ExpiresAt) {
			return fmt.Errorf("lock already held by %s", lock.Owner)
		}
		// 锁已过期，删除
		delete(dl.locks, key)
	}

	// 创建新锁
	dl.locks[key] = &Lock{
		Key:       key,
		Owner:     owner,
		ExpiresAt: time.Now().Add(dl.timeout),
	}

	return nil
}

// Release 释放锁
func (dl *DistributedLock) Release(key, owner string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if lock, exists := dl.locks[key]; exists {
		if lock.Owner != owner {
			return fmt.Errorf("lock not held by %s", owner)
		}
		delete(dl.locks, key)
		return nil
	}

	return fmt.Errorf("lock not found: %s", key)
}

// Renew 续期锁
func (dl *DistributedLock) Renew(key, owner string) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	if lock, exists := dl.locks[key]; exists {
		if lock.Owner != owner {
			return fmt.Errorf("lock not held by %s", owner)
		}
		lock.ExpiresAt = time.Now().Add(dl.timeout)
		return nil
	}

	return fmt.Errorf("lock not found: %s", key)
}

// generateTransactionID 生成事务 ID
func generateTransactionID() string {
	return fmt.Sprintf("txn_%d", time.Now().UnixNano())
}

