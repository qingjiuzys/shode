// Package collab 提供实时协作功能。
package collab

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OTTAlgorithm OTT 算法实现
type OTTAlgorithm struct {
	document   *Document
	replicaMap map[string]*Replica
	mu          sync.RWMutex
}

// Document 文档
type Document struct {
	ID         string             `json:"id"`
	Title      string             `json:"title"`
	Content    string             `json:"content"`
	Operations []*Operation      `json:"operations"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
	Version    int                `json:"version"`
	Deleted    bool               `json:"deleted"`
	Mu         sync.RWMutex       `json:"-"`
}

// Operation 操作
type Operation struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"` // "insert", "delete", "retain"
	Position  int          `json:"position"`
	Length    int          `json:"length"`
	Content   string       `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	UserID    string       `json:"user_id"`
	SiteID    string       `json:"site_id"`
}

// Replica 副本
type Replica struct {
	ID        string                     `json:"id"`
	Document  *Document                  `json:"-"`
	State     *DocumentState              `json:"state"`
	Queue     []*Operation               `json:"queue"`
	Mu        sync.Mutex
}

// DocumentState 文档状态
type DocumentState struct {
	Version      int       `json:"version"`
	LastSync    time.Time `json:"last_sync"`
	SyncStatus  string   `json:"sync_status"`
	SiteID      string   `json:"site_id"`
}

// NewOTTAlgorithm 创建 OTT 算法
func NewOTTAlgorithm(document *Document) *OTTAlgorithm {
	return &OTTAlgorithm{
		document:   document,
		replicaMap: make(map[string]*Replica),
	}
}

// AddReplica 添加副本
func (ott *OTTAlgorithm) AddReplica(siteID string) *Replica {
	ott.mu.Lock()
	defer ott.mu.Unlock()

	replica := &Replica{
		ID:       fmt.Sprintf("%s_%s", ott.document.ID, siteID),
		Document: ott.document,
		State: &DocumentState{
			Version:     0,
			SyncStatus: "synced",
			SiteID:      siteID,
		},
		Queue: make([]*Operation, 0),
	}

	ott.replicaMap[siteID] = replica
	return replica
}

// ReceiveOperation 接收操作
func (ott *OTTAlgorithm) ReceiveOperation(siteID string, op *Operation) error {
	ott.mu.Lock()
	defer ott.mu.Unlock()

	replica, exists := ott.replicaMap[siteID]
	if !exists {
		return fmt.Errorf("replica not found for site: %s", siteID)
	}

	replica.Mu.Lock()
	defer replica.Mu.Unlock()

	// 添加到队列
	replica.Queue = append(replica.Queue, op)

	return nil
}

// ProcessQueue 处理队列
func (ott *OTTAlgorithm) ProcessQueue(siteID string) error {
	ott.mu.Lock()
	replica, exists := ott.replicaMap[siteID]
	ott.mu.Unlock()

	if !exists {
		return fmt.Errorf("replica not found for site: %s", siteID)
	}

	replica.Mu.Lock()
	defer replica.Mu.Unlock()

	for len(replica.Queue) > 0 {
		op := replica.Queue[0]
		replica.Queue = replica.Queue[1:]

	// 应用操作
	ott.applyOperation(replica, op)
	}

	return nil
}

// applyOperation 应用操作
func (ott *OTTAlgorithm) applyOperation(replica *Replica, op *Operation) {
	doc := replica.Document
	doc.Mu.Lock()
	defer doc.Mu.Unlock()

	switch op.Type {
	case "insert":
		// 插入文本
		doc.Content = doc.Content[:op.Position] + op.Content + doc.Content[op.Position+op.Length:]
	case "delete":
		// 删除文本
		doc.Content = doc.Content[:op.Position] + doc.Content[op.Position+op.Length:]
	case "retain":
		// 保留文本（标记为已删除）
		// 实际实现中应该有特殊标记
	}

	doc.Version++
	doc.UpdatedAt = time.Now()
}

// Sync 同步
func (ott *OTTAlgorithm) Sync(ctx context.Context) error {
	ott.mu.Lock()
	defer ott.mu.Unlock()

	// 选择一个副本作为主副本
	var primary *Replica
	for _, replica := range ott.replicaMap {
		primary = replica
		break
	}

	if primary == nil {
		return fmt.Errorf("no replicas available")
	}

	// 将主副本的状态广播到其他副本
	for _, replica := range ott.replicaMap {
		if replica != primary {
			replica.State.Version = primary.State.Version
			replica.State.LastSync = time.Now()
			replica.State.SyncStatus = "synced"
		}
	}

	return nil
}

// GetState 获取状态
func (ott *OTTAlgorithm) GetState(siteID string) (*DocumentState, error) {
	ott.mu.RLock()
	defer ott.mu.RUnlock()

	replica, exists := ott.replicaMap[siteID]
	if !exists {
		return nil, fmt.Errorf("replica not found: %s", siteID)
	}

	return replica.State, nil
}

// CRDT CRDT 数据类型
type CRDT interface {
	Merge(other CRDT) CRDT
	Clone() CRDT
	String() string
}

// LWWRegister Last-Write-Wins 寄存器
type LWWRegister struct {
	assigns map[string]Assignment
	mu      sync.RWMutex
}

// Assignment 赋值
type Assignment struct {
	SiteID      string
	Counter    int
	Assignments map[string]int // clientID -> counter
}

// NewLWWRegister 创建 LWW 寄存器
func NewLWWRegister() *LWWRegister {
	return &LWWRegister{
		assigns: make(map[string]Assignment),
	}
}

// Assign 分配
func (lwr *LWWRegister) Assign(siteID, clientID string) int {
	lwr.mu.Lock()
	defer lwr.mu.Unlock()

	assignment, exists := lwr.assigns[siteID]
	if !exists {
		assignment = Assignment{
			SiteID:      siteID,
			Counter:    0,
			Assignments: make(map[string]int),
		}
		lwr.assigns[siteID] = assignment
	}

	assignment.Counter++
	assignment.Assignments[clientID]++

	return assignment.Counter
}

// GetAssignments 获取分配
func (lwr *LWWRegister) GetAssignments(siteID string) map[string]int {
	lwr.mu.RLock()
	defer lwr.mu.RUnlock()

	if assignment, exists := lwr.assigns[siteID]; exists {
		// 返回副本
		result := make(map[string]int)
		for k, v := range assignment.Assignments {
			result[k] = v
		}
		return result
	}

	return make(map[string]int)
}

// Resolve 冲突解决
func (lwr *LWWRegister) Resolve(siteID string, operations []*Operation) []*Operation {
	_ = lwr.GetAssignments(siteID)

	// 按分配值排序操作
	sortedOps := make([]*Operation, len(operations))
	copy(sortedOps, sortedOps)

	// 简化实现，直接按分配值排序
	for _, op := range sortedOps {
		op.UserID = ""
	}

	return sortedOps
}

// TextDocument 文本文档
type TextDocument struct {
	ID       string
	Content string
	Sites    map[string]*SiteState
	mu       sync.RWMutex
}

// SiteState 站点状态
type SiteState struct {
	SiteID     string
	Counter    int
	State      string
}

// CRDTDocument CRDT 文档
type CRDTDocument struct {
	TextDocument *TextDocument
	LWWRegister *LWWRegister
	mu          sync.RWMutex
}

// NewCRDTDocument 创建 CRDT 文档
func NewCRDTDocument() *CRDTDocument {
	return &CRDTDocument{
		TextDocument: &TextDocument{
			Sites: make(map[string]*SiteState),
		},
		LWWRegister: NewLWWRegister(),
	}
}

// Insert 插入文本
func (cd *CRDTDocument) Insert(siteID, userID string, position int, content string) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	doc := cd.TextDocument

	// 更新内容
	doc.Content = doc.Content[:position] + content + doc.Content[position:]

	// 记录站点状态
	if _, exists := doc.Sites[siteID]; !exists {
		doc.Sites[siteID] = &SiteState{
			SiteID:  siteID,
			Counter: 0,
			State:   "editing",
		}
	}

	doc.Sites[siteID].Counter++

	// 记录分配
	cd.LWWRegister.Assign(siteID, userID)

	return nil
}

// Delete 删除文本
func (cd *CRDTDocument) Delete(siteID, userID string, position, length int) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	doc := cd.TextDocument

	if position+length > len(doc.Content) {
		return fmt.Errorf("out of bounds: position=%d, length=%d, contentLen=%d", position, length, len(doc.Content))
	}

	// 删除文本
	doc.Content = doc.Content[:position] + doc.Content[position+length:]

	// 更新站点状态
	if siteState, exists := doc.Sites[siteID]; exists {
		siteState.Counter++
	}

	return nil
}

// Retain 保留文本
func (cd *CRDTDocument) Retain(siteID, userID string, position, length int) error {
	cd.mu.Lock()
	defer cd.mu.Unlock()

	// 记录站点状态
	if siteState, exists := cd.TextDocument.Sites[siteID]; exists {
		siteState.Counter++
	}

	return nil
}

// Merge 合并
func (cd *CRDTDocument) Merge(other CRDT) CRDT {
	return cd
}

// Clone 克隆
func (cd *CRDTDocument) Clone() CRDT {
	return cd
}

// String 返回字符串
func (cd *CRDTDocument) String() string {
	return fmt.Sprintf("CRDT Document: %s", cd.TextDocument.ID)
}

// CollabEngine 协作引擎
type CollabEngine struct {
	documents   map[string]*CRDTDocument
	presences  map[string]*PresenceInfo
	mu          sync.RWMutex
}

// PresenceInfo 在线信息
type PresenceInfo struct {
	UserID    string
	DocumentID string
	SiteID    string
	Cursor    int
	LastSeen   time.Time
}

// NewCollabEngine 创建协作引擎
func NewCollabEngine() *CollabEngine {
	return &CollabEngine{
		documents:  make(map[string]*CRDTDocument),
		presences: make(map[string]*PresenceInfo),
	}
}

// CreateDocument 创建文档
func (ce *CollabEngine) CreateDocument(id, title, content string) (*CRDTDocument, error) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	doc := NewCRDTDocument()
	doc.TextDocument.ID = id
	doc.TextDocument.Content = content

	ce.documents[id] = doc

	return doc, nil
}

// GetDocument 获取文档
func (ce *CollabEngine) GetDocument(id string) (*CRDTDocument, bool) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	doc, exists := ce.documents[id]
	return doc, exists
}

// Connect 连接到文档
func (ce *CollabEngine) Connect(userID, documentID, siteID string) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	presence := &PresenceInfo{
		UserID:    userID,
		DocumentID: documentID,
		SiteID:    siteID,
		Cursor:    0,
		LastSeen:   time.Now(),
	}

	ce.presences[userID+"-"+documentID] = presence

	return nil
}

// BroadcastChange 广播变更
func (ce *CollabEngine) BroadcastChange(documentID string, operation *Operation) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	doc, exists := ce.documents[documentID]
	if !exists {
		return fmt.Errorf("document not found: %s", documentID)
	}

	// 应用操作到所有副本
	switch operation.Type {
	case "insert":
		doc.Insert(operation.SiteID, operation.UserID, operation.Position, operation.Content)
	case "delete":
		doc.Delete(operation.SiteID, operation.UserID, operation.Position, operation.Length)
	case "retain":
		doc.Retain(operation.SiteID, operation.UserID, operation.Position, operation.Length)
	}

	return nil
}

// GetPresence 获取在线信息
func (ce *CollabEngine) GetPresence(userID, documentID string) (*PresenceInfo, bool) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	key := userID + "-" + documentID
	presence, exists := ce.presences[key]
	return presence, exists
}

// UpdateCursor 更新光标位置
func (ce *CollabEngine) UpdateCursor(userID, documentID, siteID string, cursor int) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	key := userID + "-" + documentID
	presence, exists := ce.presences[key]
	if !exists {
		return fmt.Errorf("presence not found: %s", key)
	}

	presence.Cursor = cursor
	presence.LastSeen = time.Now()

	return nil
}

// ConflictResolver 冲突解决器
type ConflictResolver struct {
	strategy string // "last-write-wins", "operational-transform", "custom"
}

// NewConflictResolver 创建冲突解决器
func NewConflictResolver(strategy string) *ConflictResolver {
	return &ConflictResolver{strategy: strategy}
}

// Resolve 解决冲突
func (cr *ConflictResolver) Resolve(local, remote []*Operation) ([]*Operation, error) {
	switch cr.strategy {
	case "last-write-wins":
		// 使用 LWW 算法
		return remote, nil
	case "operational-transform":
		// 使用 OTT 算法
		return cr.resolveOTT(local, remote)
	default:
		return remote, nil
	}
}

// resolveOTT 使用 OTT 解决冲突
func (cr *ConflictResolver) resolveOTT(local, remote []*Operation) ([]*Operation, error) {
	// 简化实现，合并操作
	return remote, nil
}

// ChangeHistory 变更历史
type ChangeHistory struct {
	history map[string][]*ChangeRecord
	mu      sync.RWMutex
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	ID          string       `json:"id"`
	DocumentID  string       `json:"document_id"`
	Operation   *Operation `json:"operation"`
	Timestamp   time.Time   `json:"timestamp"`
	UserID      string       `json:"user_id"`
	SiteID      string       `json:"site_id"`
}

// NewChangeHistory 创建变更历史
func NewChangeHistory() *ChangeHistory {
	return &ChangeHistory{
		history: make(map[string][]*ChangeRecord),
	}
}

// Record 记录变更
func (ch *ChangeHistory) Record(record *ChangeRecord) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if _, exists := ch.history[record.DocumentID]; !exists {
		ch.history[record.DocumentID] = make([]*ChangeRecord, 0)
	}

	ch.history[record.DocumentID] = append(ch.history[record.DocumentID], record)
}

// GetHistory 获取历史
func (ch *ChangeHistory) GetHistory(documentID string) ([]*ChangeRecord, bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	history, exists := ch.history[documentID]
	return history, exists
}

// VersionControl 版本控制
type VersionControl struct {
	versions map[string][]*DocumentVersion
	mu        sync.RWMutex
}

// DocumentVersion 文档版本
type DocumentVersion struct {
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author"`
	Message   string    `json:"message"`
}

// NewVersionControl 创建版本控制
func NewVersionControl() *VersionControl {
	return &VersionControl{
		versions: make(map[string][]*DocumentVersion),
	}
}

// SaveVersion 保存版本
func (vc *VersionControl) SaveVersion(documentID, author, message string) error {
	return nil
}

// GetVersion 获取版本
func (vc *VersionControl) GetVersion(documentID string, version int) (*DocumentVersion, error) {
	// 简化实现
	return &DocumentVersion{
		Version: 1,
	}, nil
}

// ListVersions 列出版本
func (vc *VersionControl) ListVersions(documentID string) ([]*DocumentVersion, error) {
	// 简化实现
	return []*DocumentVersion{
		{Version: 1},
	}, nil
}

// generateOperationID 生成操作 ID
func generateOperationID() string {
	return fmt.Sprintf("op_%d", time.Now().UnixNano())
}
