// Package securityplus 提供增强的安全功能。
package securityplus

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"sync"
	"time"
)

// SecurityPlusEngine 安全增强引擎
type SecurityPlusEngine struct {
	zeroTrust   *ZeroTrustEngine
	kms         *KMSManager
	audit       *AuditLogger
	rbac        *RBACManager
	abac        *ABACManager
	scanner     *SecurityScanner
	masking     *DataMasking
	encryption  *EncryptionManager
	mu          sync.RWMutex
}

// NewSecurityPlusEngine 创建安全增强引擎
func NewSecurityPlusEngine() *SecurityPlusEngine {
	return &SecurityPlusEngine{
		zeroTrust:  NewZeroTrustEngine(),
		kms:        NewKMSManager(),
		audit:      NewAuditLogger(),
		rbac:       NewRBACManager(),
		abac:       NewABACManager(),
		scanner:    NewSecurityScanner(),
		masking:    NewDataMasking(),
		encryption: NewEncryptionManager(),
	}
}

// VerifyIdentity 验证身份
func (spe *SecurityPlusEngine) VerifyIdentity(ctx context.Context, token string) (*Identity, error) {
	return spe.zeroTrust.Verify(ctx, token)
}

// EncryptData 加密数据
func (spe *SecurityPlusEngine) EncryptData(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	return spe.encryption.Encrypt(ctx, keyID, plaintext)
}

// DecryptData 解密数据
func (spe *SecurityPlusEngine) DecryptData(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	return spe.encryption.Decrypt(ctx, keyID, ciphertext)
}

// LogAudit 审计日志
func (spe *SecurityPlusEngine) LogAudit(ctx context.Context, event *AuditEvent) error {
	return spe.audit.Log(ctx, event)
}

// CheckPermission 检查权限
func (spe *SecurityPlusEngine) CheckPermission(ctx context.Context, subject, object, action string) (bool, error) {
	// 先检查 RBAC
	allowed, err := spe.rbac.Check(ctx, subject, object, action)
	if err != nil {
		return false, err
	}

	if !allowed {
		// 再检查 ABAC
		return spe.abac.Check(ctx, subject, object, action)
	}

	return true, nil
}

// Scan 扫描
func (spe *SecurityPlusEngine) Scan(ctx context.Context, target string) (*SecurityReport, error) {
	return spe.scanner.Scan(ctx, target)
}

// Mask 脱敏
func (spe *SecurityPlusEngine) Mask(data string, maskType string) string {
	return spe.masking.Mask(data, maskType)
}

// ZeroTrustEngine 零信任引擎
type ZeroTrustEngine struct {
	policies   map[string]*TrustPolicy
	identities map[string]*Identity
	sessions   map[string]*Session
	mu         sync.RWMutex
}

// TrustPolicy 信任策略
type TrustPolicy struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Rules       []*TrustRule `json:"rules"`
	Enabled     bool         `json:"enabled"`
}

// TrustRule 信任规则
type TrustRule struct {
	Type     string                 `json:"type"` // "device", "location", "time", "behavior"
	Operator string                 `json:"operator"` // "equals", "contains", "between"
	Value    interface{}            `json:"value"`
	Weight   int                    `json:"weight"`
}

// Identity 身份
type Identity struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // "user", "service", "device"
	Attributes map[string]interface{} `json:"attributes"`
	Groups    []string               `json:"groups"`
	Status    string                 `json:"status"`
}

// Session 会话
type Session struct {
	ID          string                 `json:"id"`
	IdentityID  string                 `json:"identity_id"`
	Token       string                 `json:"token"`
	Context     map[string]interface{} `json:"context"`
	CreatedAt   time.Time              `json:"created_at"`
	ExpiresAt   time.Time              `json:"expires_at"`
	LastUsedAt  time.Time              `json:"last_used_at"`
}

// NewZeroTrustEngine 创建零信任引擎
func NewZeroTrustEngine() *ZeroTrustEngine {
	return &ZeroTrustEngine{
		policies:   make(map[string]*TrustPolicy),
		identities: make(map[string]*Identity),
		sessions:   make(map[string]*Session),
	}
}

// Verify 验证
func (zte *ZeroTrustEngine) Verify(ctx context.Context, token string) (*Identity, error) {
	zte.mu.RLock()
	defer zte.mu.RUnlock()

	// 查找会话
	for _, session := range zte.sessions {
		if session.Token == token {
			if time.Now().After(session.ExpiresAt) {
				return nil, fmt.Errorf("session expired")
			}

			identity, exists := zte.identities[session.IdentityID]
			if !exists {
				return nil, fmt.Errorf("identity not found")
			}

			return identity, nil
		}
	}

	return nil, fmt.Errorf("invalid token")
}

// CreateSession 创建会话
func (zte *ZeroTrustEngine) CreateSession(ctx context.Context, identityID string, ttl time.Duration) (*Session, error) {
	zte.mu.Lock()
	defer zte.mu.Unlock()

	session := &Session{
		ID:         generateSessionID(),
		IdentityID: identityID,
		Token:      generateToken(),
		Context:    make(map[string]interface{}),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(ttl),
		LastUsedAt: time.Now(),
	}

	zte.sessions[session.ID] = session

	return session, nil
}

// AddIdentity 添加身份
func (zte *ZeroTrustEngine) AddIdentity(identity *Identity) {
	zte.mu.Lock()
	defer zte.mu.Unlock()

	zte.identities[identity.ID] = identity
}

// AddPolicy 添加策略
func (zte *ZeroTrustEngine) AddPolicy(policy *TrustPolicy) {
	zte.mu.Lock()
	defer zte.mu.Unlock()

	zte.policies[policy.ID] = policy
}

// Evaluate 评估
func (zte *ZeroTrustEngine) Evaluate(ctx context.Context, identityID, action string) (bool, error) {
	zte.mu.RLock()
	defer zte.mu.RUnlock()

	// 简化实现，总是返回 true
	return true, nil
}

// KMSManager 密钥管理器
type KMSManager struct {
	keys       map[string]*Key
	masterKeys map[string]*MasterKey
	rotation   *KeyRotation
	mu         sync.RWMutex
}

// Key 密钥
type Key struct {
	ID         string       `json:"id"`
	Type       string       `json:"type"` // "symmetric", "asymmetric"
	Value      []byte       `json:"-"`
	Metadata   map[string]string `json:"metadata"`
	CreatedAt  time.Time    `json:"created_at"`
	ExpiresAt  time.Time    `json:"expires_at"`
	RotatedAt  time.Time    `json:"rotated_at"`
}

// MasterKey 主密钥
type MasterKey struct {
	ID        string    `json:"id"`
	Value     []byte    `json:"-"`
	Algorithm string    `json:"algorithm"`
	CreatedAt time.Time `json:"created_at"`
}

// KeyRotation 密钥轮换
type KeyRotation struct {
	Policies map[string]*RotationPolicy
	Schedule map[string]time.Time
	mu       sync.RWMutex
}

// RotationPolicy 轮换策略
type RotationPolicy struct {
	KeyID     string        `json:"key_id"`
	Interval  time.Duration `json:"interval"`
	AutoRotate bool         `json:"auto_rotate"`
}

// NewKMSManager 创建密钥管理器
func NewKMSManager() *KMSManager {
	return &KMSManager{
		keys:       make(map[string]*Key),
		masterKeys: make(map[string]*MasterKey),
		rotation:   &KeyRotation{
			Policies: make(map[string]*RotationPolicy),
			Schedule: make(map[string]time.Time),
		},
	}
}

// CreateKey 创建密钥
func (kms *KMSManager) CreateKey(keyType string, metadata map[string]string) (*Key, error) {
	kms.mu.Lock()
	defer kms.mu.Unlock()

	key := &Key{
		ID:        generateKeyID(),
		Type:      keyType,
		Metadata:  metadata,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
	}

	// 生成密钥值
	if keyType == "symmetric" {
		value := make([]byte, 32)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		key.Value = value
	} else if keyType == "asymmetric" {
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, err
		}
		key.Value = x509.MarshalPKCS1PrivateKey(privateKey)
	}

	kms.keys[key.ID] = key

	return key, nil
}

// GetKey 获取密钥
func (kms *KMSManager) GetKey(keyID string) (*Key, error) {
	kms.mu.RLock()
	defer kms.mu.RUnlock()

	key, exists := kms.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	return key, nil
}

// RotateKey 轮换密钥
func (kms *KMSManager) RotateKey(keyID string) (*Key, error) {
	kms.mu.Lock()
	defer kms.mu.Unlock()

	oldKey, exists := kms.keys[keyID]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	// 创建新密钥
	newKey := &Key{
		ID:        generateKeyID(),
		Type:      oldKey.Type,
		Metadata:  oldKey.Metadata,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour),
		RotatedAt: time.Now(),
	}

	// 生成新值
	if oldKey.Type == "symmetric" {
		value := make([]byte, 32)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		newKey.Value = value
	}

	kms.keys[newKey.ID] = newKey

	return newKey, nil
}

// DeleteKey 删除密钥
func (kms *KMSManager) DeleteKey(keyID string) error {
	kms.mu.Lock()
	defer kms.mu.Unlock()

	delete(kms.keys, keyID)

	return nil
}

// AuditLogger 审计日志器
type AuditLogger struct {
	logs   map[string]*AuditEvent
	index  map[string][]string // resource -> log IDs
	export *AuditExport
	mu     sync.RWMutex
}

// AuditEvent 审计事件
type AuditEvent struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Actor     string                 `json:"actor"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"` // "success", "failure"
	Details   map[string]interface{} `json:"details"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
}

// AuditExport 审计导出
type AuditExport struct {
	Formats []string // "json", "csv", "syslog"
}

// NewAuditLogger 创建审计日志器
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs:   make(map[string]*AuditEvent),
		index:  make(map[string][]string),
		export: &AuditExport{
			Formats: []string{"json"},
		},
	}
}

// Log 记录日志
func (al *AuditLogger) Log(ctx context.Context, event *AuditEvent) error {
	al.mu.Lock()
	defer al.mu.Unlock()

	event.ID = generateAuditID()
	event.Timestamp = time.Now()

	al.logs[event.ID] = event
	al.index[event.Resource] = append(al.index[event.Resource], event.ID)

	return nil
}

// Query 查询日志
func (al *AuditLogger) Query(ctx context.Context, filter *AuditFilter) ([]*AuditEvent, error) {
	al.mu.RLock()
	defer al.mu.RUnlock()

	results := make([]*AuditEvent, 0)

	for _, log := range al.logs {
		if al.match(log, filter) {
			results = append(results, log)
		}
	}

	return results, nil
}

// match 匹配
func (al *AuditLogger) match(log *AuditEvent, filter *AuditFilter) bool {
	if filter == nil {
		return true
	}

	if filter.Actor != "" && log.Actor != filter.Actor {
		return false
	}

	if filter.Action != "" && log.Action != filter.Action {
		return false
	}

	if filter.Resource != "" && log.Resource != filter.Resource {
		return false
	}

	if !filter.StartTime.IsZero() && log.Timestamp.Before(filter.StartTime) {
		return false
	}

	if !filter.EndTime.IsZero() && log.Timestamp.After(filter.EndTime) {
		return false
	}

	return true
}

// AuditFilter 审计过滤器
type AuditFilter struct {
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// RBACManager RBAC 管理器
type RBACManager struct {
	roles     map[string]*Role
	perms     map[string]*Permission
	userRoles map[string][]string // user -> roles
	rolePerms map[string][]string // role -> permissions
	mu        sync.RWMutex
}

// Role 角色
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

// Permission 权限
type Permission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
}

// NewRBACManager 创建 RBAC 管理器
func NewRBACManager() *RBACManager {
	return &RBACManager{
		roles:     make(map[string]*Role),
		perms:     make(map[string]*Permission),
		userRoles: make(map[string][]string),
		rolePerms: make(map[string][]string),
	}
}

// AddRole 添加角色
func (rm *RBACManager) AddRole(role *Role) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.roles[role.ID] = role
}

// AssignRole 分配角色
func (rm *RBACManager) AssignRole(userID, roleID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.userRoles[userID] = append(rm.userRoles[userID], roleID)
}

// Check 检查
func (rm *RBACManager) Check(ctx context.Context, subject, object, action string) (bool, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 获取用户角色
	roles, exists := rm.userRoles[subject]
	if !exists {
		return false, nil
	}

	// 检查每个角色的权限
	for _, roleID := range roles {
		role := rm.roles[roleID]
		for _, permID := range role.Permissions {
			perm := rm.perms[permID]
			if perm.Resource == object && perm.Action == action {
				return true, nil
			}
		}
	}

	return false, nil
}

// ABACManager ABAC 管理器
type ABACManager struct {
	policies map[string]*ABACPolicy
	mu       sync.RWMutex
}

// ABACPolicy ABAC 策略
type ABACPolicy struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Target      string            `json:"target"`
	Conditions  map[string]string `json:"conditions"`
	Effect      string            `json:"effect"` // "allow", "deny"
}

// NewABACManager 创建 ABAC 管理器
func NewABACManager() *ABACManager {
	return &ABACManager{
		policies: make(map[string]*ABACPolicy),
	}
}

// AddPolicy 添加策略
func (am *ABACManager) AddPolicy(policy *ABACPolicy) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.policies[policy.ID] = policy
}

// Check 检查
func (am *ABACManager) Check(ctx context.Context, subject, object, action string) (bool, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	// 简化实现，返回 true
	return true, nil
}

// SecurityScanner 安全扫描器
type SecurityScanner struct {
	rules   map[string]*ScanRule
	scans   map[string]*SecurityReport
	mu      sync.RWMutex
}

// ScanRule 扫描规则
type ScanRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "sast", "dependency", "config"
	Severity    string   `json:"severity"`
	Patterns    []string `json:"patterns"`
}

// SecurityReport 安全报告
type SecurityReport struct {
	ID          string           `json:"id"`
	Target      string           `json:"target"`
	ScannedAt   time.Time        `json:"scanned_at"`
	Vulnerabilities []*Vulnerability `json:"vulnerabilities"`
	Summary     *ScanSummary     `json:"summary"`
}

// Vulnerability 漏洞
type Vulnerability struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Location    string `json:"location"`
	Fix         string `json:"fix"`
}

// ScanSummary 扫描摘要
type ScanSummary struct {
	Total        int            `json:"total"`
	Critical     int            `json:"critical"`
	High         int            `json:"high"`
	Medium       int            `json:"medium"`
	Low          int            `json:"low"`
	Score        int            `json:"score"`
}

// NewSecurityScanner 创建安全扫描器
func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		rules: make(map[string]*ScanRule),
		scans: make(map[string]*SecurityReport),
	}
}

// Scan 扫描
func (ss *SecurityScanner) Scan(ctx context.Context, target string) (*SecurityReport, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	report := &SecurityReport{
		ID:        generateScanID(),
		Target:    target,
		ScannedAt: time.Now(),
		Vulnerabilities: make([]*Vulnerability, 0),
		Summary:   &ScanSummary{},
	}

	ss.scans[report.ID] = report

	return report, nil
}

// DataMasking 数据脱敏
type DataMasking struct {
	rules map[string]*MaskRule
	mu    sync.RWMutex
}

// MaskRule 脱敏规则
type MaskRule struct {
	Type     string `json:"type"`
	Pattern  string `json:"pattern"`
	Replace  string `json:"replace"`
	Keep     int    `json:"keep"`
}

// NewDataMasking 创建数据脱敏
func NewDataMasking() *DataMasking {
	return &DataMasking{
		rules: make(map[string]*MaskRule),
	}
}

// Mask 脱敏
func (dm *DataMasking) Mask(data, maskType string) string {
	switch maskType {
	case "email":
		return dm.maskEmail(data)
	case "phone":
		return dm.maskPhone(data)
	case "creditcard":
		return dm.maskCreditCard(data)
	default:
		return dm.maskDefault(data)
	}
}

// maskEmail 脱敏邮箱
func (dm *DataMasking) maskEmail(email string) string {
	if len(email) < 3 {
		return "***"
	}
	return email[:2] + "***@" + email[len(email)-3:]
}

// maskPhone 脱敏手机
func (dm *DataMasking) maskPhone(phone string) string {
	if len(phone) < 7 {
		return "***"
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

// maskCreditCard 脱敏信用卡
func (dm *DataMasking) maskCreditCard(card string) string {
	if len(card) < 8 {
		return "***"
	}
	return "**** **** **** " + card[len(card)-4:]
}

// maskDefault 默认脱敏
func (dm *DataMasking) maskDefault(data string) string {
	if len(data) < 3 {
		return "***"
	}
	return data[:1] + "***" + data[len(data)-1:]
}

// EncryptionManager 加密管理器
type EncryptionManager struct {
	keys    map[string]*Key
	algorithms map[string]bool
	mu      sync.RWMutex
}

// NewEncryptionManager 创建加密管理器
func NewEncryptionManager() *EncryptionManager {
	return &EncryptionManager{
		keys:    make(map[string]*Key),
		algorithms: map[string]bool{
			"aes-256-gcm": true,
			"aes-256-cbc": true,
			"rsa-2048":    true,
		},
	}
}

// Encrypt 加密
func (em *EncryptionManager) Encrypt(ctx context.Context, keyID string, plaintext []byte) ([]byte, error) {
	em.mu.RLock()
	key, exists := em.keys[keyID]
	em.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	block, err := aes.NewCipher(key.Value)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

// Decrypt 解密
func (em *EncryptionManager) Decrypt(ctx context.Context, keyID string, ciphertext []byte) ([]byte, error) {
	em.mu.RLock()
	key, exists := em.keys[keyID]
	em.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key not found: %s", keyID)
	}

	block, err := aes.NewCipher(key.Value)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GenerateToken 生成令牌
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b, b)
}

// Hash 哈希
func Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return base64.URLEncoding.EncodeToString(hash[:], hash[:])
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// generateToken 生成令牌
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b, b)
}

// generateKeyID 生成密钥 ID
func generateKeyID() string {
	return fmt.Sprintf("key_%d", time.Now().UnixNano())
}

// generateAuditID 生成审计 ID
func generateAuditID() string {
	return fmt.Sprintf("audit_%d", time.Now().UnixNano())
}

// generateScanID 生成扫描 ID
func generateScanID() string {
	return fmt.Sprintf("scan_%d", time.Now().UnixNano())
}
