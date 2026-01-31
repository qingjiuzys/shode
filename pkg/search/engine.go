// Package search 提供搜索引擎功能。
package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// Document 文档
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Title    string                 `json:"title,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	Metadata map[string]string      `json:"metadata,omitempty"`
}

// SearchEngine 搜索引擎接口
type SearchEngine interface {
	Index(ctx context.Context, doc *Document) error
	BulkIndex(ctx context.Context, docs []*Document) error
	Search(ctx context.Context, query *Query) (*SearchResult, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, doc *Document) error
}

// Query 查询
type Query struct {
	Term        string
	Fields      []string
	Filters     map[string]interface{}
	Sort        []SortOption
	From        int
	Size        int
	Highlight   bool
	Aggregations []*Aggregation
}

// SortOption 排序选项
type SortOption struct {
	Field     string
	Ascending bool
}

// SearchResult 搜索结果
type SearchResult struct {
	Total      int               `json:"total"`
	Hits       []*Hit            `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Took       int64             `json:"took"`
}

// Hit 命中
type Hit struct {
	ID        string                 `json:"id"`
	Score     float64                `json:"score"`
	Source    interface{}            `json:"source"`
	Highlight map[string][]string    `json:"highlight,omitempty"`
}

// Aggregation 聚合
type Aggregation struct {
	Name   string
	Field  string
	Type   string // terms, range, stats, etc.
	Size   int
}

// InvertedIndex 倒排索引
type InvertedIndex struct {
	index map[string]map[string]float64 // term -> docID -> score
	mu    sync.RWMutex
}

// NewInvertedIndex 创建倒排索引
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		index: make(map[string]map[string]float64),
	}
}

// Add 添加文档
func (ii *InvertedIndex) Add(docID string, text string) {
	ii.mu.Lock()
	defer ii.mu.Unlock()

	// 分词
	terms := tokenize(text)

	// 计算词频
	termFreq := make(map[string]int)
	for _, term := range terms {
		termFreq[term]++
	}

	// 添加到索引
	for term, freq := range termFreq {
		if _, exists := ii.index[term]; !exists {
			ii.index[term] = make(map[string]float64)
		}
		// 简化的 TF-IDF 计算
		ii.index[term][docID] = float64(freq)
	}
}

// Search 搜索
func (ii *InvertedIndex) Search(query string) map[string]float64 {
	ii.mu.RLock()
	defer ii.mu.RUnlock()

	terms := tokenize(query)
	scores := make(map[string]float64)

	for _, term := range terms {
	 postings, exists := ii.index[term]
		if !exists {
			continue
		}

		for docID, score := range postings {
			scores[docID] += score
		}
	}

	return scores
}

// tokenize 分词
func tokenize(text string) []string {
	// 简化实现，按空格分词
	words := strings.Fields(text)
	terms := make([]string, 0)

	for _, word := range words {
		// 转小写
		term := strings.ToLower(word)
		// 去除标点
		term = strings.Trim(term, ".,!?;:\"'")
		if term != "" {
			terms = append(terms, term)
		}
	}

	return terms
}

// MemorySearchEngine 内存搜索引擎
type MemorySearchEngine struct {
	docs        map[string]*Document
	index       *InvertedIndex
	mu          sync.RWMutex
}

// NewMemorySearchEngine 创建内存搜索引擎
func NewMemorySearchEngine() *MemorySearchEngine {
	return &MemorySearchEngine{
		docs:  make(map[string]*Document),
		index: NewInvertedIndex(),
	}
}

// Index 索引文档
func (mse *MemorySearchEngine) Index(ctx context.Context, doc *Document) error {
	mse.mu.Lock()
	defer mse.mu.Unlock()

	mse.docs[doc.ID] = doc

	// 索引标题和内容
	text := doc.Title + " " + doc.Content
	mse.index.Add(doc.ID, text)

	return nil
}

// BulkIndex 批量索引
func (mse *MemorySearchEngine) BulkIndex(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		if err := mse.Index(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// Search 搜索
func (mse *MemorySearchEngine) Search(ctx context.Context, query *Query) (*SearchResult, error) {
	mse.mu.RLock()
	defer mse.mu.RUnlock()

	// 执行搜索
	scores := mse.index.Search(query.Term)

	// 排序
	hits := make([]*Hit, 0, len(scores))
	for docID, score := range scores {
		if doc, exists := mse.docs[docID]; exists {
			hit := &Hit{
				ID:    docID,
				Score: score,
				Source: doc,
			}

			// 高亮
			if query.Highlight {
				hit.Highlight = mse.highlight(doc, query.Term)
			}

			hits = append(hits, hit)
		}
	}

	// 排序（按分数降序）
	sortHits(hits)

	// 分页
	from := query.From
	if from < 0 {
		from = 0
	}
	size := query.Size
	if size <= 0 {
		size = 10
	}

	to := from + size
	if to > len(hits) {
		to = len(hits)
	}

	if from >= len(hits) {
		return &SearchResult{
			Total: len(scores),
			Hits:  []*Hit{},
		}, nil
	}

	return &SearchResult{
		Total: len(scores),
		Hits:  hits[from:to],
	}, nil
}

// highlight 高亮
func (mse *MemorySearchEngine) highlight(doc *Document, term string) map[string][]string {
	highlights := make(map[string][]string)

	// 在标题中高亮
	if strings.Contains(strings.ToLower(doc.Title), strings.ToLower(term)) {
		highlighted := strings.ReplaceAll(
			strings.ToLower(doc.Title),
			strings.ToLower(term),
			fmt.Sprintf("<em>%s</em>", term),
		)
		highlights["title"] = []string{highlighted}
	}

	// 在内容中高亮
	if strings.Contains(strings.ToLower(doc.Content), strings.ToLower(term)) {
		// 截取上下文
		index := strings.Index(strings.ToLower(doc.Content), strings.ToLower(term))
		start := index - 50
		if start < 0 {
			start = 0
		}
		end := index + len(term) + 50
		if end > len(doc.Content) {
			end = len(doc.Content)
		}

		context := doc.Content[start:end]
		highlighted := strings.ReplaceAll(
			strings.ToLower(context),
			strings.ToLower(term),
			fmt.Sprintf("<em>%s</em>", term),
		)
		highlights["content"] = []string{"..." + highlighted + "..."}
	}

	return highlights
}

// Delete 删除文档
func (mse *MemorySearchEngine) Delete(ctx context.Context, id string) error {
	mse.mu.Lock()
	defer mse.mu.Unlock()

	delete(mse.docs, id)
	return nil
}

// Update 更新文档
func (mse *MemorySearchEngine) Update(ctx context.Context, doc *Document) error {
	mse.mu.Lock()
	defer mse.mu.Unlock()

	mse.docs[doc.ID] = doc
	text := doc.Title + " " + doc.Content
	mse.index.Add(doc.ID, text)

	return nil
}

// sortHits 排序命中
func sortHits(hits []*Hit) {
	// 简化的冒泡排序（按分数降序）
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[i].Score < hits[j].Score {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
}

// IndexManager 索引管理器
type IndexManager struct {
	engine SearchEngine
	stats  *IndexStats
}

// IndexStats 索引统计
type IndexStats struct {
	DocumentCount int
	IndexSize     int64
	LastUpdated   string
}

// NewIndexManager 创建索引管理器
func NewIndexManager(engine SearchEngine) *IndexManager {
	return &IndexManager{
		engine: engine,
		stats:  &IndexStats{},
	}
}

// CreateIndex 创建索引
func (im *IndexManager) CreateIndex(name string) error {
	fmt.Printf("Creating index: %s\n", name)
	return nil
}

// DeleteIndex 删除索引
func (im *IndexManager) DeleteIndex(name string) error {
	fmt.Printf("Deleting index: %s\n", name)
	return nil
}

// GetStats 获取统计
func (im *IndexManager) GetStats() *IndexStats {
	return im.stats
}

// UpdateStats 更新统计
func (im *IndexManager) UpdateStats() {
	// 简化实现
}

// Aggregator 聚合器
type Aggregator struct {
	engine SearchEngine
}

// NewAggregator 创建聚合器
func NewAggregator(engine SearchEngine) *Aggregator {
	return &Aggregator{engine: engine}
}

// TermsAggregation 词项聚合
func (a *Aggregator) TermsAggregation(ctx context.Context, field string, size int) (map[string]int, error) {
	// 简化实现
	return make(map[string]int), nil
}

// RangeAggregation 范围聚合
func (a *Aggregator) RangeAggregation(ctx context.Context, field string, ranges []Range) (map[string]int, error) {
	// 简化实现
	return make(map[string]int), nil
}

// Range 范围
type Range struct {
	From float64
	To   float64
	Name string
}

// StatsAggregation 统计聚合
func (a *Aggregator) StatsAggregation(ctx context.Context, field string) (*Stats, error) {
	return &Stats{}, nil
}

// Stats 统计
type Stats struct {
	Count int
	Sum   float64
	Avg   float64
	Min   float64
	Max   float64
}

// Suggester 建议器
type Suggester struct {
	engine SearchEngine
}

// NewSuggester 创建建议器
func NewSuggester(engine SearchEngine) *Suggester {
	return &Suggester{engine: engine}
}

// TermSuggestion 词项建议
func (s *Suggester) TermSuggestion(ctx context.Context, term string, size int) ([]string, error) {
	// 简化实现，返回基于编辑距离的建议
	return []string{}, nil
}

// PhraseSuggestion 短语建议
func (s *Suggester) PhraseSuggestion(ctx context.Context, phrase string, size int) ([]string, error) {
	// 简化实现
	return []string{}, nil
}

// CompletionSuggestion 补全建议
func (s *Suggester) CompletionSuggestion(ctx context.Context, prefix string, size int) ([]string, error) {
	// 简化实现
	return []string{}, nil
}

// QueryBuilder 查询构建器
type QueryBuilder struct {
	query *Query
}

// NewQueryBuilder 创建查询构建器
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		query: &Query{
			Filters: make(map[string]interface{}),
		},
	}
}

// WithTerm 设置搜索词
func (qb *QueryBuilder) WithTerm(term string) *QueryBuilder {
	qb.query.Term = term
	return qb
}

// WithFields 设置搜索字段
func (qb *QueryBuilder) WithFields(fields ...string) *QueryBuilder {
	qb.query.Fields = fields
	return qb
}

// WithFilter 添加过滤器
func (qb *QueryBuilder) WithFilter(field string, value interface{}) *QueryBuilder {
	qb.query.Filters[field] = value
	return qb
}

// WithSort 添加排序
func (qb *QueryBuilder) WithSort(field string, ascending bool) *QueryBuilder {
	qb.query.Sort = append(qb.query.Sort, SortOption{
		Field:     field,
		Ascending: ascending,
	})
	return qb
}

// WithPagination 设置分页
func (qb *QueryBuilder) WithPagination(from, size int) *QueryBuilder {
	qb.query.From = from
	qb.query.Size = size
	return qb
}

// WithHighlight 启用高亮
func (qb *QueryBuilder) WithHighlight() *QueryBuilder {
	qb.query.Highlight = true
	return qb
}

// Build 构建查询
func (qb *QueryBuilder) Build() *Query {
	return qb.query
}

// ToJSON 转换为 JSON
func (q *Query) ToJSON() (string, error) {
	data, err := json.MarshalIndent(q, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseQuery 解析查询字符串
func ParseQuery(queryStr string) (*Query, error) {
	// 简化实现，解析 "term field:value field:value" 格式
	query := &Query{
		Filters: make(map[string]interface{}),
	}

	parts := strings.Fields(queryStr)
	for i, part := range parts {
		if i == 0 && !strings.Contains(part, ":") {
			query.Term = part
		} else if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				query.Filters[kv[0]] = kv[1]
			}
		}
	}

	return query, nil
}
