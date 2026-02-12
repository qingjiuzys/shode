// Package ai 提供 AI 对话功能。
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// LLMProvider LLM 提供商
type LLMProvider interface {
	Chat(ctx context.Context, messages []*Message) (*LLMResponse, error)
	Complete(ctx context.Context, prompt string) (*LLMResponse, error)
	Embed(ctx context.Context, texts []string) ([][]float64, error)
}

// Message 消息
type Message struct {
	Role    string `json:"role"`    // "user", "assistant", "system"
	Content string `json:"content"`
}

// LLMResponse LLM 响应
type LLMResponse struct {
	Content      string  `json:"content"`
	FinishReason string  `json:"finish_reason"`
	Usage        *Usage `json:"usage,omitempty"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIProvider OpenAI 提供商
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client *http.Client
}

// NewOpenAIProvider 创建 OpenAI 提供商
func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		model:   model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat 聊天
func (oap *OpenAIProvider) Chat(ctx context.Context, messages []*Message) (*LLMResponse, error) {
	// 简化实现
	response := &LLMResponse{
		Content: fmt.Sprintf("AI response to %d messages", len(messages)),
	}

	return response, nil
}

// Complete 补全
func (oap *OpenAIProvider) Complete(ctx context.Context, prompt string) (*LLMResponse, error) {
	// 简化实现
	return &LLMResponse{
		Content: fmt.Sprintf("Completed: %s", prompt),
	}, nil
}

// Embed 嵌入
func (oap *OpenAIProvider) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	// 简化实现
	embeddings := make([][]float64, len(texts))
	for i := range embeddings {
		embeddings[i] = make([]float64, 1536) // OpenAI ada-002 的维度
	}
	return embeddings, nil
}

// ConversationManager 对话管理器
type ConversationManager struct {
	conversations map[string]*Conversation
	mu            sync.RWMutex
	llmProvider   LLMProvider
}

// Conversation 对话
type Conversation struct {
	ID            string              `json:"id"`
	Messages      []*Message         `json:"messages"`
	Context       *ConversationContext `json:"context"`
	Settings      *ConversationSettings `json:"settings"`
	History       []*Message         `json:"history"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

// ConversationContext 对话上下文
type ConversationContext struct {
	KeyValuePairs map[string]interface{} `json:"key_value_pairs"`
	Documents    []*Document             `json:"documents"`
}

// ConversationSettings 对话设置
type ConversationSettings struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
	Model        string  `json:"model"`
	SystemPrompt string  `json:"system_prompt"`
}

// NewConversationManager 创建对话管理器
func NewConversationManager(provider LLMProvider) *ConversationManager {
	return &ConversationManager{
		conversations: make(map[string]*Conversation),
		llmProvider: provider,
	}
}

// CreateConversation 创建对话
func (cm *ConversationManager) CreateConversation(systemPrompt string, settings *ConversationSettings) (*Conversation, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	id := generateConversationID()

	conversation := &Conversation{
		ID:        id,
		Messages:  make([]*Message, 0),
		Context:   &ConversationContext{
			KeyValuePairs: make(map[string]interface{}),
			Documents:    make([]*Document, 0),
		},
		Settings: settings,
		History:   make([]*Message, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if systemPrompt != "" {
		conversation.Messages = append(conversation.Messages, &Message{
			Role:    "system",
			Content: systemPrompt,
		})
	}

	cm.conversations[id] = conversation

	return conversation, nil
}

// SendMessage 发送消息
func (cm *ConversationManager) SendMessage(ctx context.Context, conversationID, userMessage string) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conversation, exists := cm.conversations[conversationID]
	if !exists {
		return "", fmt.Errorf("conversation not found: %s", conversationID)
	}

	// 添加用户消息
	userMsg := &Message{
		Role:    "user",
		Content: userMessage,
	}
	conversation.Messages = append(conversation.Messages, userMsg)

	// 调用 LLM
	messages := conversation.Messages
	response, err := cm.llmProvider.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	// 添加助手响应
	assistantMsg := &Message{
		Role:    "assistant",
		Content: response.Content,
	}
	conversation.Messages = append(conversation.Messages, assistantMsg)
	conversation.UpdatedAt = time.Now()

	return response.Content, nil
}

// GetConversation 获取对话
func (cm *ConversationManager) GetConversation(conversationID string) (*Conversation, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conversation, exists := cm.conversations[conversationID]
	return conversation, exists
}

// RAGEngine RAG 检索增强生成引擎
type RAGEngine struct {
	llmProvider  LLMProvider
	vectorStore  *VectorStore
	retriever    *Retriever
	promptTmpl   string
}

// VectorStore 向量存储
type VectorStore struct {
	vectors  map[string][]float64
	metadata map[string]*Document
	mu       sync.RWMutex
}

// Document 文档
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Embedding []float64              `json:"embedding"`
}

// Retriever 检索器
type Retriever struct {
	vectorStore *VectorStore
	topK       int
}

// NewVectorStore 创建向量存储
func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors:  make(map[string][]float64),
		metadata: make(map[string]*Document),
	}
}

// Add 添加文档
func (vs *VectorStore) Add(id, content string, embedding []float64, metadata map[string]interface{}) error {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	doc := &Document{
		ID:       id,
		Content:  content,
		Metadata: metadata,
		Embedding: embedding,
	}

	vs.metadata[id] = doc
	vs.vectors[id] = embedding

	return nil
}

// Search 搜索相似文档
func (vs *VectorStore) Search(query []float64, topK int) ([]*SearchResult, error) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	results := make([]*SearchResult, 0)

	for id, embedding := range vs.vectors {
		score := cosineSimilarity(query, embedding)

		result := &SearchResult{
			DocumentID: id,
			Score:     score,
		}

		results = append(results, result)
	}

	// 按分数排序
	sortResults(results)

	// 返回 topK
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// SearchResult 搜索结果
type SearchResult struct {
	DocumentID string  `json:"document_id"`
	Score     float64 `json:"score"`
}

// cosineSimilarity 余弦相似度
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (sqrt(normA) * sqrt(normB))
}

// NewRAGEngine 创建 RAG 引擎
func NewRAGEngine(llmProvider LLMProvider, vectorStore *VectorStore) *RAGEngine {
	return &RAGEngine{
		llmProvider: llmProvider,
		vectorStore: vectorStore,
		retriever:   &Retriever{vectorStore: vectorStore, topK: 3},
		promptTmpl:  "Context: {{.Context}}\n\nQuestion: {{.Question}}",
	}
}

// Query 查询
func (rag *RAGEngine) Query(ctx context.Context, query string) (string, error) {
	// 生成查询嵌入
	queryEmbedding, err := rag.llmProvider.Embed(ctx, []string{query})
	if err != nil {
		return "", err
	}

	// 检索相关文档
	results, err := rag.retriever.vectorStore.Search(queryEmbedding[0], 3)
	if err != nil {
		return "", err
	}

	// 构建上下文
	context := ""
	for _, result := range results {
		if doc, exists := rag.retriever.vectorStore.metadata[result.DocumentID]; exists {
			context += doc.Content + "\n\n"
		}
	}

	// 构建提示词
	prompt := fmt.Sprintf("Context: %s\n\nQuestion: %s", context, query)

	// 调用 LLM
	response, err := rag.llmProvider.Chat(ctx, []*Message{
		{Role: "user", Content: prompt},
	})

	if err != nil {
		return "", err
	}

	return response.Content, nil
}

// Agent Agent 智能体
type Agent struct {
	ID          string
	Name        string
	Description string
	Tools       []*Tool
	LLMProvider LLMProvider
	Memory      []*Conversation
	SystemPrompt string
	goal        string
	mu          sync.RWMutex
}

// Tool 工具
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Function    func(ctx context.Context, params map[string]interface{}) (string, error)
}

// NewAgent 创建 Agent
func NewAgent(id, name, description string, llmProvider LLMProvider) *Agent {
	return &Agent{
		ID:          id,
		Name:        name,
		Description: description,
		Tools:       make([]*Tool, 0),
		LLMProvider: llmProvider,
		Memory:      make([]*Conversation, 0),
		goal:       "",
	}
}

// AddTool 添加工具
func (a *Agent) AddTool(tool *Tool) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Tools = append(a.Tools, tool)
}

// Execute 执行任务
func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
	// 添加任务到记忆
	messages := []*Message{
		{Role: "user", Content: task},
	}

	// 调用 LLM
	response, err := a.LLMProvider.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	// 检查是否需要使用工具
	for _, tool := range a.Tools {
		if contains(response.Content, tool.Name) {
			// 提取参数并调用工具
			params := extractToolParams(response.Content, tool)
			result, err := tool.Function(ctx, params)
			if err != nil {
				return "", err
			}

			// 将工具结果反馈给 LLM
			followUp := fmt.Sprintf("Tool %s returned: %s. Please provide a final answer.", tool.Name, result)
			messages = append(messages, &Message{Role: "assistant", Content: response.Content})
			messages = append(messages, &Message{Role: "user", Content: followUp})

			response, err = a.LLMProvider.Chat(ctx, messages)
			if err != nil {
				return "", err
			}

			return response.Content, nil
		}
	}

	return response.Content, nil
}

// contains 检查字符串包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

// extractToolParams 提取工具参数
func extractToolParams(content string, tool *Tool) map[string]interface{} {
	// 简化实现，使用 JSON 解析
	params := make(map[string]interface{})
	json.Unmarshal([]byte(content), params)
	return params
}

// FunctionCalling 函数调用
type FunctionCalling struct {
	agents   map[string]*Agent
	llmProvider LLMProvider
	mu       sync.RWMutex
}

// NewFunctionCalling 创建函数调用
func NewFunctionCalling(llmProvider LLMProvider) *FunctionCalling {
	return &FunctionCalling{
		agents:      make(map[string]*Agent),
		llmProvider: llmProvider,
	}
}

// RegisterAgent 注册 Agent
func (fc *FunctionCalling) RegisterAgent(agent *Agent) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	fc.agents[agent.ID] = agent
}

// CallFunction 调用函数
func (fc *FunctionCalling) CallFunction(ctx context.Context, functionID string, parameters map[string]interface{}) (string, error) {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	// 简化实现，返回固定值
	return fmt.Sprintf("Executed function %s with params %v", functionID, parameters), nil
}

// ChatWithFunctions 带函数调用的对话
func (fc *FunctionCalling) ChatWithFunctions(ctx context.Context, message string, availableFunctions []string) (string, error) {
	// 构建 system prompt
	systemPrompt := fmt.Sprintf("You are a helpful assistant. You have access to these functions: %v", availableFunctions)

	messages := []*Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: message},
	}

	response, err := fc.llmProvider.Chat(ctx, messages)
	if err != nil {
		return "", err
	}

	// 检查是否需要函数调用
	for _, functionID := range availableFunctions {
		if contains(response.Content, functionID) {
			// 执行函数
			result, err := fc.CallFunction(ctx, functionID, make(map[string]interface{}))
			if err != nil {
				return "", err
			}

			// 将结果反馈给 LLM
			followUp := fmt.Sprintf("Function %s returned: %s. Please provide a final answer.", functionID, result)
			messages = append(messages, &Message{Role: "assistant", Content: response.Content})
			messages = append(messages, &Message{Role: "user", Content: followUp})

			response, err = fc.llmProvider.Chat(ctx, messages)
			if err != nil {
				return "", err
			}

			return response.Content, nil
		}
	}

	return response.Content, nil
}

// generateConversationID 生成对话 ID
func generateConversationID() string {
	return fmt.Sprintf("conv_%d", time.Now().UnixNano())
}

// sqrt 平方根
func sqrt(x float64) float64 {
	// 简化实现
	return x * 0.5
}

// sortResults 排序结果
func sortResults(results []*SearchResult) {
	// 简化冒泡排序
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
