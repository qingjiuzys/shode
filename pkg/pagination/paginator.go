// Package pagination 提供分页功能
package pagination

import (
	"math"
)

// PageRequest 分页请求
type PageRequest struct {
	// 页码（从1开始）
	Page int `json:"page" form:"page"`
	// 每页大小
	Size int `json:"size" form:"size"`
	// 排序字段
	Sort string `json:"sort" form:"sort"`
	// 排序方向（asc, desc）
	Order string `json:"order" form:"order"`
}

// PageResponse 分页响应
type PageResponse struct {
	// 当前页码
	Page int `json:"page"`
	// 每页大小
	Size int `json:"size"`
	// 总记录数
	Total int64 `json:"total"`
	// 总页数
	Pages int `json:"pages"`
	// 是否有上一页
	HasPrev bool `json:"has_prev"`
	// 是否有下一页
	HasNext bool `json:"has_next"`
	// 上一页页码
	PrevPage int `json:"prev_page,omitempty"`
	// 下一页页码
	NextPage int `json:"next_page,omitempty"`
	// 数据列表
	Items any `json:"items"`
}

// Paginator 分页器
type Paginator struct {
	page int
	size int
	sort string
	order string
	total int64
}

// NewPaginator 创建分页器
func NewPaginator(page, size int) *Paginator {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100 // 最大每页100条
	}

	return &Paginator{
		page: page,
		size: size,
		order: "desc",
	}
}

// Page 获取页码
func (p *Paginator) Page() int {
	return p.page
}

// Size 获取每页大小
func (p *Paginator) Size() int {
	return p.size
}

// Sort 获取排序字段
func (p *Paginator) Sort() string {
	return p.sort
}

// Order 获取排序方向
func (p *Paginator) Order() string {
	return p.order
}

// SetSort 设置排序
func (p *Paginator) SetSort(sort string) *Paginator {
	p.sort = sort
	return p
}

// SetOrder 设置排序方向
func (p *Paginator) SetOrder(order string) *Paginator {
	if order == "asc" || order == "desc" {
		p.order = order
	}
	return p
}

// Offset 计算偏移量
func (p *Paginator) Offset() int {
	return (p.page - 1) * p.size
}

// Limit 获取限制数量
func (p *Paginator) Limit() int {
	return p.size
}

// TotalPages 计算总页数
func (p *Paginator) TotalPages() int {
	if p.total == 0 {
		return 0
	}
	return int(math.Ceil(float64(p.total) / float64(p.size)))
}

// HasPrev 是否有上一页
func (p *Paginator) HasPrev() bool {
	return p.page > 1
}

// HasNext 是否有下一页
func (p *Paginator) HasNext() bool {
	return p.page < p.TotalPages()
}

// PrevPage 上一页页码
func (p *Paginator) PrevPage() int {
	if p.HasPrev() {
		return p.page - 1
	}
	return 0
}

// NextPage 下一页页码
func (p *Paginator) NextPage() int {
	if p.HasNext() {
		return p.page + 1
	}
	return 0
}

// SetTotal 设置总数
func (p *Paginator) SetTotal(total int64) *Paginator {
	p.total = total
	return p
}

// Response 生成分页响应
func (p *Paginator) Response(items any) *PageResponse {
	return &PageResponse{
		Page:     p.page,
		Size:     p.size,
		Total:    p.total,
		Pages:    p.TotalPages(),
		HasPrev:  p.HasPrev(),
		HasNext:  p.HasNext(),
		PrevPage: p.PrevPage(),
		NextPage: p.NextPage(),
		Items:    items,
	}
}

// Paginate 分页切片
func Paginate[T any](items []T, page, size int) ([]T, *PageResponse) {
	p := NewPaginator(page, size)
	total := int64(len(items))
	p.SetTotal(total)

	offset := p.Offset()
	limit := p.Limit()

	if offset >= len(items) {
		return []T{}, p.Response([]T{})
	}

	end := offset + limit
	if end > len(items) {
		end = len(items)
	}

	pagedItems := items[offset:end]
	return pagedItems, p.Response(pagedItems)
}

// PageSlice 分页切片（简化版本）
func PageSlice[T any](items []T, page, size int) []T {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	offset := (page - 1) * size
	if offset >= len(items) {
		return []T{}
	}

	end := offset + size
	if end > len(items) {
		end = len(items)
	}

	return items[offset:end]
}

// CursorBasedPaginator 基于游标的分页器（适合大数据量）
type CursorBasedPaginator struct {
	cursor string
	limit  int
	items  any
	hasMore bool
}

// NewCursorBasedPaginator 创建游标分页器
func NewCursorBasedPaginator(cursor string, limit int) *CursorBasedPaginator {
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	return &CursorBasedPaginator{
		cursor: cursor,
		limit:  limit,
	}
}

// Cursor 获取游标
func (c *CursorBasedPaginator) Cursor() string {
	return c.cursor
}

// Limit 获取限制数量
func (c *CursorBasedPaginator) Limit() int {
	return c.limit
}

// HasMore 是否有更多数据
func (c *CursorBasedPaginator) HasMore() bool {
	return c.hasMore
}

// SetHasMore 设置是否有更多数据
func (c *CursorBasedPaginator) SetHasMore(hasMore bool) *CursorBasedPaginator {
	c.hasMore = hasMore
	return c
}

// Response 生成响应
func (c *CursorBasedPaginator) Response(items any, nextCursor string) map[string]any {
	return map[string]any{
		"items":      items,
		"limit":      c.limit,
		"cursor":     c.cursor,
		"next_cursor": nextCursor,
		"has_more":   c.hasMore,
	}
}

// InfiniteScrollPaginator 无限滚动分页器
type InfiniteScrollPaginator struct {
	page  int
	size  int
	total int64
}

// NewInfiniteScrollPaginator 创建无限滚动分页器
func NewInfiniteScrollPaginator(page, size int) *InfiniteScrollPaginator {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if size > 50 {
		size = 50
	}

	return &InfiniteScrollPaginator{
		page: page,
		size: size,
	}
}

// Offset 计算偏移量
func (i *InfiniteScrollPaginator) Offset() int {
	return (i.page - 1) * i.size
}

// Limit 获取限制数量
func (i *InfiniteScrollPaginator) Limit() int {
	return i.size
}

// SetTotal 设置总数
func (i *InfiniteScrollPaginator) SetTotal(total int64) *InfiniteScrollPaginator {
	i.total = total
	return i
}

// HasMore 是否有更多数据
func (i *InfiniteScrollPaginator) HasMore() bool {
	return int64(i.Offset()+i.size) < i.total
}

// Response 生成响应
func (i *InfiniteScrollPaginator) Response(items any) map[string]any {
	return map[string]any{
		"page":     i.page,
		"size":     i.size,
		"total":    i.total,
		"has_more": i.HasMore(),
		"items":    items,
	}
}

// PageIterator 分页迭代器
type PageIterator struct {
	currentPage int
	pageSize    int
	total       int64
}

// NewPageIterator 创建分页迭代器
func NewPageIterator(pageSize, total int64) *PageIterator {
	return &PageIterator{
		currentPage: 0,
		pageSize:    int(pageSize),
		total:       total,
	}
}

// Next 下一页
func (pi *PageIterator) Next() bool {
	pi.currentPage++
	return !pi.IsLast()
}

// IsLast 是否最后一页
func (pi *PageIterator) IsLast() bool {
	return int64((pi.currentPage)*pi.pageSize) >= pi.total
}

// Current 当前页码
func (pi *PageIterator) Current() int {
	return pi.currentPage
}

// Offset 当前偏移量
func (pi *PageIterator) Offset() int {
	return (pi.currentPage - 1) * pi.pageSize
}

// Limit 限制数量
func (pi *PageIterator) Limit() int {
	return pi.pageSize
}

// Reset 重置迭代器
func (pi *PageIterator) Reset() {
	pi.currentPage = 0
}

// PageInfo 分页信息
type PageInfo struct {
	CurrentPage  int   `json:"current_page"`
	PageSize     int   `json:"page_size"`
	TotalPages   int   `json:"total_pages"`
	TotalItems   int64 `json:"total_items"`
	HasPrev      bool  `json:"has_prev"`
	HasNext      bool  `json:"has_next"`
	FirstPage    int   `json:"first_page"`
	LastPage     int   `json:"last_page"`
	PrevPage     int   `json:"prev_page,omitempty"`
	NextPage     int   `json:"next_page,omitempty"`
	StartIndex   int   `json:"start_index"`
	EndIndex     int   `json:"end_index"`
}

// BuildPageInfo 构建分页信息
func BuildPageInfo(page, size int, total int64) *PageInfo {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	totalPages := int(math.Ceil(float64(total) / float64(size)))
	if totalPages == 0 {
		totalPages = 1
	}

	startIndex := (page - 1) * size
	endIndex := startIndex + size - 1
	if endIndex >= int(total) {
		endIndex = int(total) - 1
	}

	info := &PageInfo{
		CurrentPage: page,
		PageSize:    size,
		TotalPages:  totalPages,
		TotalItems:  total,
		FirstPage:   1,
		LastPage:    totalPages,
		StartIndex:  startIndex,
		EndIndex:    endIndex,
		HasPrev:     page > 1,
		HasNext:     page < totalPages,
	}

	if info.HasPrev {
		info.PrevPage = page - 1
	}
	if info.HasNext {
		info.NextPage = page + 1
	}

	return info
}

// ParsePageRequest 解析分页请求
func ParsePageRequest(page, size int, defaultSize int) *PageRequest {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = defaultSize
	}
	if size > 100 {
		size = 100
	}

	return &PageRequest{
		Page:  page,
		Size:  size,
		Order: "desc",
	}
}

// CalculatePageRange 计算页码范围（用于显示页码按钮）
func CalculatePageRange(currentPage, totalPages, displayCount int) []int {
	if displayCount < 1 {
		displayCount = 7
	}

	if totalPages <= displayCount {
		pages := make([]int, totalPages)
		for i := 0; i < totalPages; i++ {
			pages[i] = i + 1
		}
		return pages
	}

	half := displayCount / 2
	start := currentPage - half
	end := currentPage + half

	if start < 1 {
		start = 1
		end = displayCount
	}

	if end > totalPages {
		end = totalPages
		start = totalPages - displayCount + 1
		if start < 1 {
			start = 1
		}
	}

	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}

	return pages
}

// SliceToBatch 将切片分批处理
func SliceToBatch[T any](items []T, batchSize int) [][]T {
	if batchSize <= 0 {
		batchSize = 100
	}

	var batches [][]T

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}
		batches = append(batches, items[i:end])
	}

	return batches
}
