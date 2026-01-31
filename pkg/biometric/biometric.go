// Package biometric 提供生物识别安全功能。
package biometric

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// BiometricEngine 生物识别引擎
type BiometricEngine struct {
	multimodal  *MultimodalBiometric
	behavioral  *BehavioralBiometric
	continuous  *ContinuousAuthentication
	liveness    *LivenessDetection
	fraud       *FraudDetection
	mu          sync.RWMutex
}

// NewBiometricEngine 创建生物识别引擎
func NewBiometricEngine() *BiometricEngine {
	return &BiometricEngine{
		multimodal:  NewMultimodalBiometric(),
		behavioral:  NewBehavioralBiometric(),
		continuous:  NewContinuousAuthentication(),
		liveness:    NewLivenessDetection(),
		fraud:       NewFraudDetection(),
	}
}

// Enroll 注册生物特征
func (be *BiometricEngine) Enroll(ctx context.Context, userID string, template *BiometricTemplate) (*EnrollmentResult, error) {
	return be.multimodal.Enroll(ctx, userID, template)
}

// Verify 验证生物特征
func (be *BiometricEngine) Verify(ctx context.Context, userID string, sample *BiometricSample) (*VerificationResult, error) {
	return be.multimodal.Verify(ctx, userID, sample)
}

// AnalyzeBehavior 分析行为
func (be *BiometricEngine) AnalyzeBehavior(ctx context.Context, userID string, behavior *BehaviorPattern) (*BehaviorScore, error) {
	return be.behavioral.Analyze(ctx, userID, behavior)
}

// StartContinuous 启动连续认证
func (be *BiometricEngine) StartContinuous(ctx context.Context, userID string) (*ContinuousSession, error) {
	return be.continuous.Start(ctx, userID)
}

// CheckLiveness 检查活体
func (be *BiometricEngine) CheckLiveness(ctx context.Context, sample *BiometricSample) (*LivenessResult, error) {
	return be.liveness.Check(ctx, sample)
}

// DetectFraud 检测欺诈
func (be *BiometricEngine) DetectFraud(ctx context.Context, transaction *Transaction) (*FraudAlert, error) {
	return be.fraud.Detect(ctx, transaction)
}

// MultimodalBiometric 多模态生物特征
type MultimodalBiometric struct {
	templates    map[string]*UserTemplate
	modalities   map[string]*ModalityConfig
	fusion       *FusionStrategy
	matchers     map[string]*BiometricMatcher
	mu           sync.RWMutex
}

// BiometricTemplate 生物特征模板
type BiometricTemplate struct {
	Fingerprint   *FingerprintTemplate   `json:"fingerprint,omitempty"`
	Face          *FaceTemplate          `json:"face,omitempty"`
	Iris          *IrisTemplate          `json:"iris,omitempty"`
	Voice         *VoiceTemplate         `json:"voice,omitempty"`
	Signature     *SignatureTemplate     `json:"signature,omitempty"`
}

// FingerprintTemplate 指纹模板
type FingerprintTemplate struct {
	Minutiae   []*Minutia `json:"minutiae"`
	ROI        *RegionOfInterest `json:"roi"`
	Quality    float64    `json:"quality"`
}

// Minutia 细节点
type Minutia struct {
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Angle     float64 `json:"angle"`
	Type      string  `json:"type"` // "ending", "bifurcation"
}

// RegionOfInterest 感兴趣区域
type RegionOfInterest struct {
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
}

// FaceTemplate 人脸模板
type FaceTemplate struct {
	Embedding   []float64              `json:"embedding"`
	Landmarks   []*FacialLandmark      `json:"landmarks"`
	Quality     float64                `json:"quality"`
	Attributes  map[string]string      `json:"attributes"`
}

// FacialLandmark 面部特征点
type FacialLandmark struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z,omitempty"`
}

// IrisTemplate 虹膜模板
type IrisTemplate struct {
	IrisCode    string   `json:"iris_code"`
	NoiseMask   []bool   `json:"noise_mask"`
	Quality     float64  `json:"quality"`
}

// VoiceTemplate 声纹模板
type VoiceTemplate struct {
	MFCC        []float64 `json:"mfcc"`
	Pitch       []float64 `json:"pitch"`
	Timbre      []float64 `json:"timbre"`
	Quality     float64   `json:"quality"`
}

// SignatureTemplate 签名模板
type SignatureTemplate struct {
	Points      []*SignaturePoint `json:"points"`
	Pressure    []float64         `json:"pressure"`
	Velocity    []float64         `json:"velocity"`
	Timing      []time.Duration   `json:"timing"`
}

// SignaturePoint 签名点
type SignaturePoint struct {
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Timestamp time.Time `json:"timestamp"`
}

// UserTemplate 用户模板
type UserTemplate struct {
	UserID      string                 `json:"user_id"`
	Templates   []*BiometricTemplate   `json:"templates"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Version     int                    `json:"version"`
}

// ModalityConfig 模态配置
type ModalityConfig struct {
	Type        string  `json:"type"` // "fingerprint", "face", "iris", "voice", "signature"`
	Enabled     bool    `json:"enabled"`
	Required    bool    `json:"required"`
	Weight      float64 `json:"weight"`
	Threshold   float64 `json:"threshold"`
}

// FusionStrategy 融合策略
type FusionStrategy struct {
	Type        string                 `json:"type"` // "score", "decision", "feature", "rank"`
	Algorithm   string                 `json:"algorithm"` // "weighted_sum", "svm", "neural"
	Weights     map[string]float64     `json:"weights"`
}

// BiometricMatcher 生物特征匹配器
type BiometricMatcher struct {
	Algorithm   string                 `json:"algorithm"`
	Speed       time.Duration          `json:"speed"`
	Accuracy    float64                `json:"accuracy"`
}

// BiometricSample 生物特征样本
type BiometricSample struct {
	Type        string                 `json:"type"`
	Data        []byte                 `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
	Quality     float64                `json:"quality"`
}

// EnrollmentResult 注册结果
type EnrollmentResult struct {
	UserID      string                 `json:"user_id"`
	Success     bool                   `json:"success"`
	Quality     float64                `json:"quality"`
	Attempts    int                    `json:"attempts"`
	Errors      []string               `json:"errors,omitempty"`
	EnrolledAt  time.Time              `json:"enrolled_at"`
}

// VerificationResult 验证结果
type VerificationResult struct {
	UserID      string                 `json:"user_id"`
	Match       bool                   `json:"match"`
	Score       float64                `json:"score"`
	Confidence  float64                `json:"confidence"`
	Modality    string                 `json:"modality"`
	Details     map[string]interface{} `json:"details"`
	VerifiedAt  time.Time              `json:"verified_at"`
}

// NewMultimodalBiometric 创建多模态生物特征
func NewMultimodalBiometric() *MultimodalBiometric {
	return &MultimodalBiometric{
		templates:  make(map[string]*UserTemplate),
		modalities: make(map[string]*ModalityConfig),
		fusion:     &FusionStrategy{Type: "score", Algorithm: "weighted_sum"},
		matchers:   make(map[string]*BiometricMatcher),
	}
}

// Enroll 注册
func (mb *MultimodalBiometric) Enroll(ctx context.Context, userID string, template *BiometricTemplate) (*EnrollmentResult, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	userTemplate := &UserTemplate{
		UserID:    userID,
		Templates: []*BiometricTemplate{template},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Version:   1,
	}

	mb.templates[userID] = userTemplate

	result := &EnrollmentResult{
		UserID:     userID,
		Success:    true,
		Quality:    0.95,
		Attempts:   1,
		EnrolledAt: time.Now(),
	}

	return result, nil
}

// Verify 验证
func (mb *MultimodalBiometric) Verify(ctx context.Context, userID string, sample *BiometricSample) (*VerificationResult, error) {
	mb.mu.RLock()
	defer mb.mu.RUnlock()

	template, exists := mb.templates[userID]
	if !exists {
		return nil, fmt.Errorf("user not enrolled")
	}

	// 简化实现 - 模拟匹配
	score := rand.Float64()
	match := score > 0.7

	result := &VerificationResult{
		UserID:     userID,
		Match:      match,
		Score:      score,
		Confidence: 0.92,
		Modality:   sample.Type,
		Details: map[string]interface{}{
			"template_version": template.Version,
			"match_algorithm":  "hamming_distance",
		},
		VerifiedAt: time.Now(),
	}

	return result, nil
}

// BehavioralBiometric 行为生物特征
type BehavioralBiometric struct {
	profiles   map[string]*BehaviorProfile
	patterns   map[string]*BehaviorPattern
	scorers    map[string]*BehaviorScorer
	mu         sync.RWMutex
}

// BehaviorPattern 行为模式
type BehaviorPattern struct {
	Type       string                 `json:"type"` // "keystroke", "mouse", "gait", "swipe"`
	Data       []float64              `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Context    map[string]interface{} `json:"context"`
}

// BehaviorProfile 行为画像
type BehaviorProfile struct {
	UserID       string                 `json:"user_id"`
	Keystroke    *KeystrokeProfile      `json:"keystroke,omitempty"`
	Mouse        *MouseProfile          `json:"mouse,omitempty"`
	Gait         *GaitProfile           `json:"gait,omitempty"`
	Swipe        *SwipeProfile          `json:"swipe,omitempty"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// KeystrokeProfile 按键画像
type KeystrokeProfile struct {
	HoldTimes    []float64 `json:"hold_times"`     // key hold duration
	FlightTimes  []float64 `json:"flight_times"`   // time between keys
	TypingSpeed  float64   `json:"typing_speed"`   // words per minute
	ErrorRate    float64   `json:"error_rate"`     // percentage
}

// MouseProfile 鼠标画像
type MouseProfile struct {
	Movement     *MovementPattern       `json:"movement"`
	ClickPattern *ClickPattern          `json:"click_pattern"`
	ScrollSpeed  float64                `json:"scroll_speed"`
}

// MovementPattern 移动模式
type MovementPattern struct {
	Velocity     []float64 `json:"velocity"`
	Acceleration []float64 `json:"acceleration"`
	Deviation    float64   `json:"deviation"` // from straight line
}

// ClickPattern 点击模式
type ClickPattern struct {
	DwellTime    []float64 `json:"dwell_time"`    // time before click
	Interval     []float64 `json:"interval"`      // time between clicks
	Pressure     []float64 `json:"pressure"`
}

// GaitProfile 步态画像
type GaitProfile struct {
	StrideLength []float64 `json:"stride_length"`
	Cadence      []float64 `json:"cadence"`       // steps per minute`
	SwingTime    []float64 `json:"swing_time"`
	StanceTime   []float64 `json:"stance_time"`
}

// SwipeProfile 滑动画像
type SwipeProfile struct {
	Velocity     []float64 `json:"velocity"`
	Pressure     []float64 `json:"pressure"`
	Direction    []string `json:"direction"`
	Acceleration []float64 `json:"acceleration"`
}

// BehaviorScorer 行为打分器
type BehaviorScorer struct {
	Type      string  `json:"type"`
	Threshold float64 `json:"threshold"`
	Sensitivity float64 `json:"sensitivity"`
}

// BehaviorScore 行为分数
type BehaviorScore struct {
	UserID     string                 `json:"user_id"`
	Score      float64                `json:"score"`
	Risk       string                 `json:"risk"` // "low", "medium", "high"
	Confidence float64                `json:"confidence"`
	Anomalies  []*Anomaly             `json:"anomalies"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Anomaly 异常
type Anomaly struct {
	Type      string  `json:"type"`
	Severity  float64 `json:"severity"`
	Context   string  `json:"context"`
}

// NewBehavioralBiometric 创建行为生物特征
func NewBehavioralBiometric() *BehavioralBiometric {
	return &BehavioralBiometric{
		profiles: make(map[string]*BehaviorProfile),
		patterns: make(map[string]*BehaviorPattern),
		scorers:  make(map[string]*BehaviorScorer),
	}
}

// Analyze 分析
func (bb *BehavioralBiometric) Analyze(ctx context.Context, userID string, pattern *BehaviorPattern) (*BehaviorScore, error) {
	bb.mu.Lock()
	defer bb.mu.Unlock()

	profile, exists := bb.profiles[userID]
	if !exists {
		profile = &BehaviorProfile{
			UserID:    userID,
			UpdatedAt: time.Now(),
		}
		bb.profiles[userID] = profile
	}

	// 简化实现
	score := rand.Float64()
	risk := "low"
	if score < 0.5 {
		risk = "medium"
	}
	if score < 0.3 {
		risk = "high"
	}

	result := &BehaviorScore{
		UserID:     userID,
		Score:      score,
		Risk:       risk,
		Confidence: 0.88,
		Anomalies:  make([]*Anomaly, 0),
		Timestamp:  time.Now(),
	}

	return result, nil
}

// ContinuousAuthentication 连续认证
type ContinuousAuthentication struct {
	sessions   map[string]*ContinuousSession
	monitors   map[string]*ActivityMonitor
	policies   map[string]*AuthPolicy
	mu         sync.RWMutex
}

// ContinuousSession 连续认证会话
type ContinuousSession struct {
	SessionID  string                 `json:"session_id"`
	UserID     string                 `json:"user_id"`
	StartTime  time.Time              `json:"start_time"`
	LastActive time.Time              `json:"last_active"`
	Score      float64                `json:"score"`
	Status     string                 `json:"status"` // "active", "warning", "terminated"`
	Events     []*AuthEvent           `json:"events"`
}

// ActivityMonitor 活动监控器
type ActivityMonitor struct {
	Type        string                 `json:"type"`
	SampleRate  time.Duration          `json:"sample_rate"`
	BufferSize  int                    `json:"buffer_size"`
}

// AuthPolicy 认证策略
type AuthPolicy struct {
	Name            string        `json:"name"`
	RiskThreshold   float64       `json:"risk_threshold"`
	Action          string        `json:"action"` // "monitor", "challenge", "terminate"`
	CooldownPeriod  time.Duration `json:"cooldown_period"`
}

// AuthEvent 认证事件
type AuthEvent struct {
	Type      string    `json:"type"`
	Score     float64   `json:"score"`
	Timestamp time.Time `json:"timestamp"`
}

// NewContinuousAuthentication 创建连续认证
func NewContinuousAuthentication() *ContinuousAuthentication {
	return &ContinuousAuthentication{
		sessions: make(map[string]*ContinuousSession),
		monitors: make(map[string]*ActivityMonitor),
		policies: make(map[string]*AuthPolicy),
	}
}

// Start 启动
func (ca *ContinuousAuthentication) Start(ctx context.Context, userID string) (*ContinuousSession, error) {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	session := &ContinuousSession{
		SessionID:  generateSessionID(),
		UserID:     userID,
		StartTime:  time.Now(),
		LastActive: time.Now(),
		Score:      1.0,
		Status:     "active",
		Events:     make([]*AuthEvent, 0),
	}

	ca.sessions[session.SessionID] = session

	return session, nil
}

// Update 更新
func (ca *ContinuousAuthentication) Update(ctx context.Context, sessionID string, score float64) error {
	ca.mu.Lock()
	defer ca.mu.Unlock()

	session, exists := ca.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Score = score
	session.LastActive = time.Now()

	if score < 0.3 {
		session.Status = "terminated"
	} else if score < 0.6 {
		session.Status = "warning"
	}

	event := &AuthEvent{
		Type:      "score_update",
		Score:     score,
		Timestamp: time.Now(),
	}

	session.Events = append(session.Events, event)

	return nil
}

// LivenessDetection 活体检测
type LivenessDetection struct {
	detectors  map[string]*LivenessDetector
	challenges map[string]*Challenge
	mu         sync.RWMutex
}

// LivenessDetector 活体检测器
type LivenessDetector struct {
	Modality   string  `json:"modality"` // "face", "voice", "fingerprint"
	Algorithm  string  `json:"algorithm"`
	Accuracy   float64 `json:"accuracy"`
}

// Challenge 挑战
type Challenge struct {
	Type       string                 `json:"type"` // "passive", "active"
	Prompt     string                 `json:"prompt"`
	Expected   map[string]interface{} `json:"expected"`
	Timeout    time.Duration          `json:"timeout"`
}

// LivenessResult 活体检测结果
type LivenessResult struct {
	Live       bool                   `json:"live"`
	Confidence float64                `json:"confidence"`
	Method     string                 `json:"method"`
	Score      float64                `json:"score"`
	Details    map[string]interface{} `json:"details"`
	Timestamp  time.Time              `json:"timestamp"`
}

// NewLivenessDetection 创建活体检测
func NewLivenessDetection() *LivenessDetection {
	return &LivenessDetection{
		detectors:  make(map[string]*LivenessDetector),
		challenges: make(map[string]*Challenge),
	}
}

// Check 检查
func (ld *LivenessDetection) Check(ctx context.Context, sample *BiometricSample) (*LivenessResult, error) {
	ld.mu.RLock()
	defer ld.mu.RUnlock()

	// 简化实现
	result := &LivenessResult{
		Live:       true,
		Confidence: 0.96,
		Method:     "passive",
		Score:      0.94,
		Details: map[string]interface{}{
			"texture_analysis": 0.95,
			"motion_analysis":  0.93,
			"depth_analysis":   0.94,
		},
		Timestamp: time.Now(),
	}

	return result, nil
}

// FraudDetection 欺诈检测
type FraudDetection struct {
	profiles   map[string]*FraudProfile
	incidents  map[string]*FraudIncident
	models     map[string]*FraudModel
	mu         sync.RWMutex
}

// Transaction 交易
type Transaction struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Type         string                 `json:"type"`
	Amount       float64                `json:"amount"`
	Location     *Location              `json:"location"`
	Device       string                 `json:"device"`
	Timestamp    time.Time              `json:"timestamp"`
	Attributes   map[string]interface{} `json:"attributes"`
}

// FraudProfile 欺诈画像
type FraudProfile struct {
	UserID          string                 `json:"user_id"`
	RiskScore       float64                `json:"risk_score"`
	BehaviorBaseline map[string]float64    `json:"behavior_baseline"`
	TransactionHistory []*Transaction      `json:"transaction_history"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// FraudIncident 欺诈事件
type FraudIncident struct {
	ID          string                 `json:"id"`
	Transaction string                 `json:"transaction"`
	Type        string                 `json:"type"` // "account_takeover", "synthetic_identity", "payment_fraud"`
	Severity    string                 `json:"severity"` // "low", "medium", "high", "critical"`
	Status      string                 `json:"status"` // "detected", "investigating", "confirmed", "false_positive"`
	DetectedAt  time.Time              `json:"detected_at"`
	ResolvedAt  time.Time              `json:"resolved_at,omitempty"`
}

// FraudModel 欺诈模型
type FraudModel struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "supervised", "unsupervised", "hybrid"`
	Algorithm   string                 `json:"algorithm"`
	Accuracy    float64                `json:"accuracy"`
	FalsePositiveRate float64          `json:"false_positive_rate"`
}

// FraudAlert 欺诈告警
type FraudAlert struct {
	IncidentID  string                 `json:"incident_id"`
	Transaction string                 `json:"transaction"`
	RiskScore   float64                `json:"risk_score"`
	Reasons     []string               `json:"reasons"`
	Action      string                 `json:"action"` // "block", "flag", "monitor"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewFraudDetection 创建欺诈检测
func NewFraudDetection() *FraudDetection {
	return &FraudDetection{
		profiles:  make(map[string]*FraudProfile),
		incidents: make(map[string]*FraudIncident),
		models:    make(map[string]*FraudModel),
	}
}

// Detect 检测
func (fd *FraudDetection) Detect(ctx context.Context, transaction *Transaction) (*FraudAlert, error) {
	fd.mu.Lock()
	defer fd.mu.Unlock()

	profile, exists := fd.profiles[transaction.UserID]
	if !exists {
		profile = &FraudProfile{
			UserID:          transaction.UserID,
			RiskScore:       0.1,
			BehaviorBaseline: make(map[string]float64),
			TransactionHistory: make([]*Transaction, 0),
			UpdatedAt:       time.Now(),
		}
		fd.profiles[transaction.UserID] = profile
	}

	// 简化实现 - 基于规则的欺诈检测
	riskScore := 0.1
	reasons := make([]string, 0)

	if transaction.Amount > 10000 {
		riskScore += 0.3
		reasons = append(reasons, "high_amount")
	}

	hour := transaction.Timestamp.Hour()
	if hour < 6 || hour > 23 {
		riskScore += 0.2
		reasons = append(reasons, "unusual_time")
	}

	if riskScore > 0.5 {
		alert := &FraudAlert{
			IncidentID: generateIncidentID(),
			Transaction: transaction.ID,
			RiskScore:   riskScore,
			Reasons:     reasons,
			Action:      "flag",
			Timestamp:   time.Now(),
		}

		incident := &FraudIncident{
			ID:          alert.IncidentID,
			Transaction: transaction.ID,
			Type:        "payment_fraud",
			Severity:    "medium",
			Status:      "detected",
			DetectedAt:  time.Now(),
		}

		fd.incidents[incident.ID] = incident

		return alert, nil
	}

	return &FraudAlert{
		IncidentID:  "",
		Transaction: transaction.ID,
		RiskScore:   riskScore,
		Reasons:     reasons,
		Action:      "monitor",
		Timestamp:   time.Now(),
	}, nil
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	data := fmt.Sprintf("%d", time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])[:16]
}

// generateIncidentID 生成事件 ID
func generateIncidentID() string {
	return fmt.Sprintf("incident_%d", time.Now().UnixNano())
}
