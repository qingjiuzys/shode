// Package alert 提供监控告警功能。
package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Metric 指标
type Metric struct {
	Name      string
	Value     float64
	Labels    map[string]string
	Timestamp time.Time
}

// AlertRule 告警规则
type AlertRule struct {
	ID          string
	Name        string
	Condition   string // ">", "<", "==", "!=", "in"
	Threshold   float64
	Duration    time.Duration
	Labels      map[string]string
	Annotations map[string]string
	Enabled     bool
	Severity    string
	Actions     []Action
}

// Action 动作
type Action struct {
	Type     string
	Channel  string
	Template string
	Params   map[string]interface{}
}

// RuleEngine 规则引擎
type RuleEngine struct {
	rules   map[string]*AlertRule
	metrics chan *Metric
	mu      sync.RWMutex
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules:   make(map[string]*AlertRule),
		metrics: make(chan *Metric, 1000),
	}
}

// AddRule 添加规则
func (re *RuleEngine) AddRule(rule *AlertRule) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules[rule.ID] = rule
}

// RemoveRule 移除规则
func (re *RuleEngine) RemoveRule(ruleID string) {
	re.mu.Lock()
	defer re.mu.Unlock()

	delete(re.rules, ruleID)
}

// GetRule 获取规则
func (re *RuleEngine) GetRule(ruleID string) (*AlertRule, bool) {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rule, exists := re.rules[ruleID]
	return rule, exists
}

// ListRules 列出规则
func (re *RuleEngine) ListRules() []*AlertRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(re.rules))
	for _, rule := range re.rules {
		rules = append(rules, rule)
	}
	return rules
}

// Evaluate 评估指标
func (re *RuleEngine) Evaluate(metric *Metric) []*Alert {
	re.mu.RLock()
	defer re.mu.RUnlock()

	alerts := make([]*Alert, 0)

	for _, rule := range re.rules {
		if !rule.Enabled {
			continue
		}

		if re.matchRule(rule, metric) {
			alert := &Alert{
				RuleID:     rule.ID,
				RuleName:   rule.Name,
				Severity:   rule.Severity,
				Labels:     rule.Labels,
				Annotations: rule.Annotations,
				Metric:     metric,
				Timestamp:  time.Now(),
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// matchRule 匹配规则
func (re *RuleEngine) matchRule(rule *AlertRule, metric *Metric) bool {
	// 检查标签匹配
	for k, v := range rule.Labels {
		if metricVal, exists := metric.Labels[k]; !exists || metricVal != v {
			return false
		}
	}

	// 检查条件
	switch rule.Condition {
	case ">":
		return metric.Value > rule.Threshold
	case "<":
		return metric.Value < rule.Threshold
	case "==":
		return metric.Value == rule.Threshold
	case "!=":
		return metric.Value != rule.Threshold
	}

	return false
}

// Alert 告警
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Metric      *Metric                `json:"metric"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	Silenced    bool                   `json:"silenced"`
}

// AlertManager 告警管理器
type AlertManager struct {
	ruleEngine   *RuleEngine
	notifier     *Notifier
	silences    map[string]*Silence
	alerts      map[string]*Alert
	mu          sync.RWMutex
}

// Silence 静默
type Silence struct {
	ID        string
	Matchers  map[string]string
	StartAt   time.Time
	EndAt     time.Time
	CreatedBy string
	Comment   string
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	return &AlertManager{
		ruleEngine: NewRuleEngine(),
		notifier:   NewNotifier(),
		silences:  make(map[string]*Silence),
		alerts:     make(map[string]*Alert),
	}
}

// ProcessMetric 处理指标
func (am *AlertManager) ProcessMetric(metric *Metric) error {
	// 评估规则
	alerts := am.ruleEngine.Evaluate(metric)

	for _, alert := range alerts {
		// 检查沉默
		if am.isSilenced(alert) {
			continue
		}

		// 记录告警
		am.mu.Lock()
		alert.ID = generateAlertID()
		am.alerts[alert.ID] = alert
		am.mu.Unlock()

		// 发送通知
		am.notifier.Send(alert)
	}

	return nil
}

// isSilenced 检查是否沉默
func (am *AlertManager) isSilenced(alert *Alert) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, silence := range am.silences {
		if time.Now().Before(silence.StartAt) || time.Now().After(silence.EndAt) {
			continue
		}

		// 检查匹配
		matched := true
		for k, v := range silence.Matchers {
			if alertVal, exists := alert.Labels[k]; !exists || alertVal != v {
				matched = false
				break
			}
		}

		if matched {
			return true
		}
	}

	return false
}

// GetAlert 获取告警
func (am *AlertManager) GetAlert(alertID string) (*Alert, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alert, exists := am.alerts[alertID]
	return alert, exists
}

// ListAlerts 列出告警
func (am *AlertManager) ListAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.alerts))
	for _, alert := range am.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// ResolveAlert 解决告警
func (am *AlertManager) ResolveAlert(alertID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if alert, exists := am.alerts[alertID]; exists {
		alert.Resolved = true
		alert.Status = "resolved"
		return nil
	}

	return fmt.Errorf("alert not found: %s", alertID)
}

// SilenceAlert 沉默告警
func (am *AlertManager) SilenceAlert(silence *Silence) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.silences[silence.ID] = silence
}

// Notifier 通知器
type Notifier struct {
	channels map[string]NotificationChannel
	mu       sync.RWMutex
}

// NotificationChannel 通知渠道
type NotificationChannel interface {
	Send(alert *Alert) error
	Type() string
}

// NewNotifier 创建通知器
func NewNotifier() *Notifier {
	return &Notifier{
		channels: make(map[string]NotificationChannel),
	}
}

// RegisterChannel 注册渠道
func (n *Notifier) RegisterChannel(channel NotificationChannel) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.channels[channel.Type()] = channel
}

// Send 发送通知
func (n *Notifier) Send(alert *Alert) error {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, channel := range n.channels {
		if err := channel.Send(alert); err != nil {
			return err
		}
	}

	return nil
}

// EmailChannel 邮件渠道
type EmailChannel struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
}

// NewEmailChannel 创建邮件渠道
func NewEmailChannel(smtpHost string, smtpPort int, username, password, from string, to []string) *EmailChannel {
	return &EmailChannel{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
		to:       to,
	}
}

// Send 发送邮件
func (ec *EmailChannel) Send(alert *Alert) error {
	// 简化实现
	fmt.Printf("Sending email for alert %s to %v\n", alert.ID, ec.to)
	return nil
}

// Type 返回类型
func (ec *EmailChannel) Type() string {
	return "email"
}

// SMSChannel 短信渠道
type SMSChannel struct {
	provider string
	apiKey   string
	phones   []string
}

// NewSMSChannel 创建短信渠道
func NewSMSChannel(provider, apiKey string, phones []string) *SMSChannel {
	return &SMSChannel{
		provider: provider,
		apiKey:   apiKey,
		phones:   phones,
	}
}

// Send 发送短信
func (sc *SMSChannel) Send(alert *Alert) error {
	fmt.Printf("Sending SMS for alert %s to %v\n", alert.ID, sc.phones)
	return nil
}

// Type 返回类型
func (sc *SMSChannel) Type() string {
	return "sms"
}

// WebhookChannel Webhook 渠道
type WebhookChannel struct {
	url     string
	headers map[string]string
}

// NewWebhookChannel 创建 Webhook 渠道
func NewWebhookChannel(url string, headers map[string]string) *WebhookChannel {
	return &WebhookChannel{
		url:     url,
		headers: headers,
	}
}

// Send 发送 Webhook
func (wc *WebhookChannel) Send(alert *Alert) error {
	fmt.Printf("Sending webhook to %s for alert %s\n", wc.url, alert.ID)
	return nil
}

// Type 返回类型
func (wc *WebhookChannel) Type() string {
	return "webhook"
}

// SlackChannel Slack 渠道
type SlackChannel struct {
	webhookURL string
	channel    string
}

// NewSlackChannel 创建 Slack 渠道
func NewSlackChannel(webhookURL, channel string) *SlackChannel {
	return &SlackChannel{
		webhookURL: webhookURL,
		channel:    channel,
	}
}

// Send 发送到 Slack
func (sc *SlackChannel) Send(alert *Alert) error {
	fmt.Printf("Sending to Slack channel %s for alert %s\n", sc.channel, alert.ID)
	return nil
}

// Type 返回类型
func (sc *SlackChannel) Type() string {
	return "slack"
}

// DingTalkChannel 钉钉渠道
type DingTalkChannel struct {
	webhookURL string
	atMobiles  []string
	atUsers    []string
}

// NewDingTalkChannel 创建钉钉渠道
func NewDingTalkChannel(webhookURL string, atMobiles, atUsers []string) *DingTalkChannel {
	return &DingTalkChannel{
		webhookURL: webhookURL,
		atMobiles:  atMobiles,
		atUsers:    atUsers,
	}
}

// Send 发送到钉钉
func (dtc *DingTalkChannel) Send(alert *Alert) error {
	fmt.Printf("Sending to DingTalk for alert %s\n", alert.ID)
	return nil
}

// Type 返回类型
func (dtc *DingTalkChannel) Type() string {
	return "dingtalk"
}

// AlertDeduplicator 告警去重器
type AlertDeduplicator struct {
	alerts map[string]*DedupKey
	mu     sync.RWMutex
}

// DedupKey 去重键
type DedupKey struct {
	RuleID    string
	Labels    map[string]string
	LastAlert time.Time
	Count     int
}

// NewAlertDeduplicator 创建告警去重器
func NewAlertDeduplicator() *AlertDeduplicator {
	return &AlertDeduplicator{
		alerts: make(map[string]*DedupKey),
	}
}

// Deduplicate 去重
func (ad *AlertDeduplicator) Deduplicate(alert *Alert) bool {
	ad.mu.Lock()
	defer ad.mu.Unlock()

	key := ad.generateKey(alert)

	if existing, exists := ad.alerts[key]; exists {
		// 检查时间窗口
		if time.Since(existing.LastAlert) < 5*time.Minute {
			existing.Count++
			return false // 重复告警
		}
	}

	ad.alerts[key] = &DedupKey{
		RuleID:    alert.RuleID,
		Labels:    alert.Labels,
		LastAlert: time.Now(),
		Count:     1,
	}

	return true // 新告警
}

// generateKey 生成去重键
func (ad *AlertDeduplicator) generateKey(alert *Alert) string {
	return fmt.Sprintf("%s:%s", alert.RuleID, hashLabels(alert.Labels))
}

// hashLabels 哈希标签
func hashLabels(labels map[string]string) string {
	// 简化实现
	return ""
}

// DutyScheduler 值班管理
type DutyScheduler struct {
	schedules map[string]*DutySchedule
	mu        sync.RWMutex
}

// DutySchedule 值班表
type DutySchedule struct {
	Name    string
	Rotations []*Rotation
	mu      sync.RWMutex
}

// Rotation 轮换
type Rotation struct {
	Type     string // "daily", "weekly", "custom"
	Users    []string
	StartTime time.Time
	EndTime   time.Time
}

// NewDutyScheduler 创建值班管理
func NewDutyScheduler() *DutyScheduler {
	return &DutyScheduler{
		schedules: make(map[string]*DutySchedule),
	}
}

// GetOnDuty 获取值班人
func (ds *DutyScheduler) GetOnDuty(scheduleName string, t time.Time) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	schedule, exists := ds.schedules[scheduleName]
	if !exists {
		return "", fmt.Errorf("schedule not found: %s", scheduleName)
	}

	// 简化实现，返回第一个用户
	if len(schedule.Rotations) > 0 && len(schedule.Rotations[0].Users) > 0 {
		return schedule.Rotations[0].Users[0], nil
	}

	return "", fmt.Errorf("no on-duty user found")
}

// AddSchedule 添加值班表
func (ds *DutyScheduler) AddSchedule(schedule *DutySchedule) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.schedules[schedule.Name] = schedule
}

// IncidentManager 事件管理
type IncidentManager struct {
	incidents map[string]*Incident
	mu        sync.RWMutex
}

// Incident 事件
type Incident struct {
	ID          string
	Title       string
	Description string
	Severity    string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	AssignedTo  string
	Alerts      []*Alert
}

// NewIncidentManager 创建事件管理器
func NewIncidentManager() *IncidentManager {
	return &IncidentManager{
		incidents: make(map[string]*Incident),
	}
}

// CreateIncident 创建事件
func (im *IncidentManager) CreateIncident(title, description, severity string, alert *Alert) *Incident {
	im.mu.Lock()
	defer im.mu.Unlock()

	incident := &Incident{
		ID:          generateIncidentID(),
		Title:       title,
		Description: description,
		Severity:    severity,
		Status:      "open",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Alerts:      []*Alert{alert},
	}

	im.incidents[incident.ID] = incident
	return incident
}

// ResolveIncident 解决事件
func (im *IncidentManager) ResolveIncident(incidentID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	incident, exists := im.incidents[incidentID]
	if !exists {
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	incident.Status = "resolved"
	incident.UpdatedAt = time.Now()
	return nil
}

// GetIncident 获取事件
func (im *IncidentManager) GetIncident(incidentID string) (*Incident, bool) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	incident, exists := im.incidents[incidentID]
	return incident, exists
}

// ListIncidents 列出事件
func (im *IncidentManager) ListIncidents() []*Incident {
	im.mu.RLock()
	defer im.mu.RUnlock()

	incidents := make([]*Incident, 0, len(im.incidents))
	for _, incident := range im.incidents {
		incidents = append(incidents, incident)
	}
	return incidents
}

// generateAlertID 生成告警 ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateIncidentID 生成事件 ID
func generateIncidentID() string {
	return fmt.Sprintf("incident_%d", time.Now().UnixNano())
}
