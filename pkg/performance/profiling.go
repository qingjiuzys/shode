package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"
)

// ProfilerEnhanced 增强的性能分析器
type ProfilerEnhanced struct {
	cpuProfiler     *CPUProfiler
	memoryProfiler  *MemoryProfiler
	blockProfiler   *BlockProfiler
	mutexProfiler   *MutexProfiler
	flameGraph      *FlameGraphGenerator
	callGraph       *CallGraphBuilder
	active          bool
	outputDir       string
	mu              sync.RWMutex
}

// CPUProfiler CPU性能分析器
type CPUProfiler struct {
	enabled       bool
	sampleRate    int
	duration      time.Duration
	samples       []*CPUSample
	mu            sync.RWMutex
}

// CPUSample CPU样本
type CPUSample struct {
	Timestamp    time.Time
	Stack        []StackFrame
	Count        int
}

// StackFrame 栈帧
type StackFrame struct {
	Function string
	File     string
	Line     int
	Package  string
}

// MemoryProfiler 内存性能分析器
type MemoryProfiler struct {
	enabled      bool
	sampleRate   int
	allocations  []*MemoryAllocation
	frees        []*MemoryFree
	mu           sync.RWMutex
	trackStacks  bool
}

// MemoryAllocation 内存分配
type MemoryAllocation struct {
	Timestamp   time.Time
	Size        uint64
	Stack       []StackFrame
	Type        string
	Pointer     uintptr
}

// MemoryFree 内存释放
type MemoryFree struct {
	Timestamp   time.Time
	Pointer     uintptr
	Size        uint64
	Lifetime    time.Duration
}

// BlockProfiler 阻塞性能分析器
type BlockProfiler struct {
	enabled     bool
	events      []*BlockEvent
	mu          sync.RWMutex
}

// BlockEvent 阻塞事件
type BlockEvent struct {
	Timestamp   time.Time
	Duration    time.Duration
	Stack       []StackFrame
	Type        string // "channel", "mutex", "cond", "select"
}

// MutexProfiler 互斥锁性能分析器
type MutexProfiler struct {
	enabled     bool
	contentions []*MutexContention
	mu          sync.RWMutex
}

// MutexContention 互斥锁竞争
type MutexContention struct {
	Timestamp    time.Time
	WaitDuration time.Duration
	LockAddress  uintptr
	Stack        []StackFrame
	HolderStack  []StackFrame
}

// FlameGraphGenerator 火焰图生成器
type FlameGraphGenerator struct {
	roots     []*FlameNode
	mu        sync.RWMutex
}

// FlameNode 火焰图节点
type FlameNode struct {
	Name       string
	Value      int64
	Children   []*FlameNode
	Depth      int
	Percentage float64
	Color      string
}

// CallGraphBuilder 调用图构建器
type CallGraphBuilder struct {
	nodes     map[string]*CallGraphNode
	edges     []*CallGraphEdge
	mu        sync.RWMutex
}

// CallGraphNode 调用图节点
type CallGraphNode struct {
	ID         string
	Function   string
	File       string
	Line       int
	CallCount  int
	TotalTime  time.Duration
	SelfTime   time.Duration
}

// CallGraphEdge 调用图边
type CallGraphEdge struct {
	From     string
	To       string
	CallCount int
	TotalTime time.Duration
}

// NewProfilerEnhanced 创建增强的性能分析器
func NewProfilerEnhanced(outputDir string) *ProfilerEnhanced {
	return &ProfilerEnhanced{
		cpuProfiler:    &CPUProfiler{sampleRate: 100},
		memoryProfiler: &MemoryProfiler{sampleRate: 1, trackStacks: true},
		blockProfiler:  &BlockProfiler{},
		mutexProfiler:  &MutexProfiler{},
		flameGraph:     &FlameGraphGenerator{roots: make([]*FlameNode, 0)},
		callGraph:      &CallGraphBuilder{
			nodes: make(map[string]*CallGraphNode),
			edges: make([]*CallGraphEdge, 0),
		},
		outputDir:      outputDir,
	}
}

// Start 开始性能分析
func (pe *ProfilerEnhanced) Start(ctx context.Context) error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if pe.active {
		return fmt.Errorf("profiler already active")
	}

	// 启动CPU分析
	if err := pe.cpuProfiler.Start(); err != nil {
		return err
	}

	// 启动内存分析
	if err := pe.memoryProfiler.Start(); err != nil {
		return err
	}

	// 启动阻塞分析
	if err := pe.blockProfiler.Start(); err != nil {
		return err
	}

	// 启动互斥锁分析
	if err := pe.mutexProfiler.Start(); err != nil {
		return err
	}

	pe.active = true

	// 设置自动停止
	go func() {
		<-ctx.Done()
		pe.Stop()
	}()

	return nil
}

// Stop 停止性能分析
func (pe *ProfilerEnhanced) Stop() error {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if !pe.active {
		return nil
	}

	// 停止各个分析器
	pe.cpuProfiler.Stop()
	pe.memoryProfiler.Stop()
	pe.blockProfiler.Stop()
	pe.mutexProfiler.Stop()

	// 生成火焰图
	pe.generateFlameGraph()

	// 生成调用图
	pe.generateCallGraph()

	// 生成报告
	pe.generateReport()

	pe.active = false

	return nil
}

// Start 启动CPU分析
func (cp *CPUProfiler) Start() error {
	if cp.enabled {
		return fmt.Errorf("CPU profiler already enabled")
	}

	cp.enabled = true
	cp.samples = make([]*CPUSample, 0)

	// 设置采样率
	runtime.SetCPUProfileRate(cp.sampleRate)

	// 启动采样goroutine
	go cp.sample()

	return nil
}

// Stop 停止CPU分析
func (cp *CPUProfiler) Stop() {
	cp.enabled = false
	runtime.SetCPUProfileRate(0)
}

// sample 采样CPU
func (cp *CPUProfiler) sample() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for cp.enabled {
		select {
		case <-ticker.C:
			// 获取当前goroutine的栈
			buf := make([]byte, 1024*1024)
			n := runtime.Stack(buf, false)
			stack := parseStack(buf[:n])

			cp.mu.Lock()
			cp.samples = append(cp.samples, &CPUSample{
				Timestamp: time.Now(),
				Stack:     stack,
				Count:     1,
			})
			cp.mu.Unlock()
		}
	}
}

// Start 启动内存分析
func (mp *MemoryProfiler) Start() error {
	if mp.enabled {
		return fmt.Errorf("memory profiler already enabled")
	}

	mp.enabled = true
	mp.allocations = make([]*MemoryAllocation, 0)
	mp.frees = make([]*MemoryFree, 0)

	// 设置内存采样率
	runtime.MemProfileRate = mp.sampleRate

	return nil
}

// Stop 停止内存分析
func (mp *MemoryProfiler) Stop() {
	mp.enabled = false
	runtime.MemProfileRate = 0
}

// Start 启动阻塞分析
func (bp *BlockProfiler) Start() error {
	if bp.enabled {
		return fmt.Errorf("block profiler already enabled")
	}

	bp.enabled = true
	bp.events = make([]*BlockEvent, 0)

	// 设置阻塞分析率
	runtime.SetBlockProfileRate(1)

	return nil
}

// Stop 停止阻塞分析
func (bp *BlockProfiler) Stop() {
	bp.enabled = false
	runtime.SetBlockProfileRate(0)
}

// Start 启动互斥锁分析
func (mup *MutexProfiler) Start() error {
	if mup.enabled {
		return fmt.Errorf("mutex profiler already enabled")
	}

	mup.enabled = true
	mup.contentions = make([]*MutexContention, 0)

	// 设置互斥锁分析率
	runtime.SetMutexProfileFraction(1)

	return nil
}

// Stop 停止互斥锁分析
func (mup *MutexProfiler) Stop() {
	mup.enabled = false
	runtime.SetMutexProfileFraction(0)
}

// generateFlameGraph 生成火焰图
func (pe *ProfilerEnhanced) generateFlameGraph() {
	// 从CPU样本构建火焰图
	pe.flameGraph.mu.Lock()
	defer pe.flameGraph.mu.Unlock()

	// 清空现有节点
	pe.flameGraph.roots = make([]*FlameNode, 0)

	// 构建调用树
	callTree := make(map[string]*FlameNode)

	pe.cpuProfiler.mu.RLock()
	defer pe.cpuProfiler.mu.RUnlock()

	for _, sample := range pe.cpuProfiler.samples {
		currentLevel := pe.flameGraph.roots

		for i, frame := range sample.Stack {
			nodeKey := fmt.Sprintf("%s:%d", frame.Function, frame.Line)

			node, exists := callTree[nodeKey]
			if !exists {
				node = &FlameNode{
					Name:     frame.Function,
					Value:    0,
					Children: make([]*FlameNode, 0),
					Depth:    i,
				}
				callTree[nodeKey] = node
				currentLevel = append(currentLevel, node)
			} else {
				// 找到现有节点
				found := false
				for _, n := range currentLevel {
					if n.Name == frame.Function {
						node = n
						found = true
						break
					}
				}
				if !found {
					currentLevel = append(currentLevel, node)
				}
			}

			node.Value += int64(sample.Count)
			currentLevel = node.Children
		}
	}

	// 计算百分比
	total := int64(0)
	for _, root := range pe.flameGraph.roots {
		total += root.Value
	}

	for _, root := range pe.flameGraph.roots {
		calculatePercentage(root, total)
		assignColors(root, 0)
	}
}

// calculatePercentage 计算百分比
func calculatePercentage(node *FlameNode, total int64) int64 {
	sum := node.Value
	for _, child := range node.Children {
		childSum := calculatePercentage(child, total)
		sum += childSum
	}

	node.Percentage = float64(node.Value) / float64(total) * 100
	return sum
}

// assignColors 分配颜色
func assignColors(node *FlameNode, depth int) {
	// 基于深度和名称生成颜色
	hue := (depth * 37) % 360
	saturation := 60 + (depth*7)%30
	lightness := 50 + (depth*5)%20

	node.Color = fmt.Sprintf("hsl(%d, %d%%, %d%%)", hue, saturation, lightness)

	for _, child := range node.Children {
		assignColors(child, depth+1)
	}
}

// generateCallGraph 生成调用图
func (pe *ProfilerEnhanced) generateCallGraph() {
	// 从CPU样本构建调用图
	pe.callGraph.mu.Lock()
	defer pe.callGraph.mu.Unlock()

	pe.callGraph.nodes = make(map[string]*CallGraphNode)
	pe.callGraph.edges = make([]*CallGraphEdge, 0)

	pe.cpuProfiler.mu.RLock()
	defer pe.cpuProfiler.mu.RUnlock()

	for _, sample := range pe.cpuProfiler.samples {
		for i := 0; i < len(sample.Stack)-1; i++ {
			caller := sample.Stack[i]
			callee := sample.Stack[i+1]

			// 添加/更新节点
			callerID := fmt.Sprintf("%s:%d", caller.Function, caller.Line)
			calleeID := fmt.Sprintf("%s:%d", callee.Function, callee.Line)

			if _, exists := pe.callGraph.nodes[callerID]; !exists {
				pe.callGraph.nodes[callerID] = &CallGraphNode{
					ID:       callerID,
					Function: caller.Function,
					File:     caller.File,
					Line:     caller.Line,
				}
			}

			if _, exists := pe.callGraph.nodes[calleeID]; !exists {
				pe.callGraph.nodes[calleeID] = &CallGraphNode{
					ID:       calleeID,
					Function: callee.Function,
					File:     callee.File,
					Line:     callee.Line,
				}
			}

			// 添加/更新边
			edge := findEdge(pe.callGraph.edges, callerID, calleeID)
			if edge == nil {
				edge = &CallGraphEdge{
					From:      callerID,
					To:        calleeID,
					CallCount: 1,
				}
				pe.callGraph.edges = append(pe.callGraph.edges, edge)
			} else {
				edge.CallCount++
			}

			// 更新统计
			pe.callGraph.nodes[callerID].CallCount++
		}
	}
}

// findEdge 查找边
func findEdge(edges []*CallGraphEdge, from, to string) *CallGraphEdge {
	for _, edge := range edges {
		if edge.From == from && edge.To == to {
			return edge
		}
	}
	return nil
}

// generateReport 生成报告
func (pe *ProfilerEnhanced) generateReport() error {
	report := &ProfilingReport{
		GeneratedAt:  time.Now(),
		CPUStats:     pe.getCPUStats(),
		MemoryStats:  pe.getMemoryStats(),
		BlockStats:   pe.getBlockStats(),
		MutexStats:   pe.getMutexStats(),
		HotFunctions: pe.getHotFunctions(),
		TopAllocators: pe.getTopAllocators(),
	}

	// 保存JSON报告
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	reportPath := fmt.Sprintf("%s/profile_report_%d.json", pe.outputDir, time.Now().Unix())
	return os.WriteFile(reportPath, data, 0644)
}

// ProfilingReport 性能分析报告
type ProfilingReport struct {
	GeneratedAt    time.Time            `json:"generated_at"`
	Duration       time.Duration        `json:"duration"`
	CPUStats       *CPUStats            `json:"cpu_stats"`
	MemoryStats    *MemoryProfileStats  `json:"memory_stats"`
	BlockStats     *BlockProfileStats   `json:"block_stats"`
	MutexStats     *MutexProfileStats   `json:"mutex_stats"`
	HotFunctions   []*HotFunction       `json:"hot_functions"`
	TopAllocators  []*AllocatorInfo     `json:"top_allocators"`
}

// CPUStats CPU统计
type CPUStats struct {
	SampleCount     int               `json:"sample_count"`
	Duration        time.Duration     `json:"duration"`
	TopFunctions    []*FunctionStat   `json:"top_functions"`
}

// MemoryProfileStats 内存分析统计
type MemoryProfileStats struct {
	TotalAllocations uint64              `json:"total_allocations"`
	TotalFrees       uint64              `json:"total_frees"`
	CurrentUsage     uint64              `json:"current_usage"`
	TopAllocators    []*AllocatorInfo    `json:"top_allocators"`
}

// BlockProfileStats 阻塞分析统计
type BlockProfileStats struct {
	BlockCount     int              `json:"block_count"`
	TotalDuration  time.Duration    `json:"total_duration"`
	TopBlockers    []*BlockStat     `json:"top_blockers"`
}

// MutexProfileStats 互斥锁分析统计
type MutexProfileStats struct {
	ContentionCount int                `json:"contention_count"`
	TotalWait       time.Duration      `json:"total_wait"`
	TopContentions  []*MutexContentionStat `json:"top_contentions"`
}

// HotFunction 热点函数
type HotFunction struct {
	Function    string        `json:"function"`
	File        string        `json:"file"`
	Line        int           `json:"line"`
	SampleCount int           `json:"sample_count"`
	Percentage  float64       `json:"percentage"`
	SelfTime    time.Duration `json:"self_time"`
}

// FunctionStat 函数统计
type FunctionStat struct {
	Function   string `json:"function"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	CallCount  int    `json:"call_count"`
	Percentage float64 `json:"percentage"`
}

// AllocatorInfo 分配器信息
type AllocatorInfo struct {
	Function     string `json:"function"`
	AllocCount   int    `json:"alloc_count"`
	TotalBytes   uint64 `json:"total_bytes"`
	AverageBytes uint64 `json:"average_bytes"`
	Percentage   float64 `json:"percentage"`
}

// BlockStat 阻塞统计
type BlockStat struct {
	Function    string        `json:"function"`
	BlockCount  int           `json:"block_count"`
	TotalTime   time.Duration `json:"total_time"`
	AverageTime time.Duration `json:"average_time"`
}

// MutexContentionStat 互斥锁竞争统计
type MutexContentionStat struct {
	LockAddress string        `json:"lock_address"`
	WaitCount   int           `json:"wait_count"`
	TotalWait   time.Duration `json:"total_wait"`
	AverageWait time.Duration `json:"average_wait"`
}

// getCPUStats 获取CPU统计
func (pe *ProfilerEnhanced) getCPUStats() *CPUStats {
	pe.cpuProfiler.mu.RLock()
	defer pe.cpuProfiler.mu.RUnlock()

	// 统计函数调用次数
	funcCounts := make(map[string]int)
	totalSamples := len(pe.cpuProfiler.samples)

	for _, sample := range pe.cpuProfiler.samples {
		if len(sample.Stack) > 0 {
			frame := sample.Stack[0]
			key := fmt.Sprintf("%s:%d:%s", frame.Function, frame.Line, frame.File)
			funcCounts[key]++
		}
	}

	// 转换为FunctionStat
	topFunctions := make([]*FunctionStat, 0)
	for key, count := range funcCounts {
		percentage := float64(count) / float64(totalSamples) * 100
		// 解析key（简化）
		function := key
		topFunctions = append(topFunctions, &FunctionStat{
			Function:   function,
			CallCount:  count,
			Percentage: percentage,
		})
	}

	// 排序
	sort.Slice(topFunctions, func(i, j int) bool {
		return topFunctions[i].CallCount > topFunctions[j].CallCount
	})

	// 只保留前20个
	if len(topFunctions) > 20 {
		topFunctions = topFunctions[:20]
	}

	return &CPUStats{
		SampleCount:  totalSamples,
		TopFunctions: topFunctions,
	}
}

// getMemoryStats 获取内存统计
func (pe *ProfilerEnhanced) getMemoryStats() *MemoryProfileStats {
	pe.memoryProfiler.mu.RLock()
	defer pe.memoryProfiler.mu.RUnlock()

	// 统计分配
	allocCounts := make(map[string]*AllocatorInfo)
	totalAlloc := uint64(0)

	for _, alloc := range pe.memoryProfiler.allocations {
		key := alloc.Type
		if key == "" {
			key = "unknown"
		}

		info, exists := allocCounts[key]
		if !exists {
			info = &AllocatorInfo{
				Function: key,
			}
			allocCounts[key] = info
		}

		info.AllocCount++
		info.TotalBytes += alloc.Size
		totalAlloc += alloc.Size
	}

	// 计算统计
	topAllocators := make([]*AllocatorInfo, 0)
	for _, info := range allocCounts {
		if totalAlloc > 0 {
			info.Percentage = float64(info.TotalBytes) / float64(totalAlloc) * 100
			info.AverageBytes = info.TotalBytes / uint64(info.AllocCount)
		}
		topAllocators = append(topAllocators, info)
	}

	// 排序
	sort.Slice(topAllocators, func(i, j int) bool {
		return topAllocators[i].TotalBytes > topAllocators[j].TotalBytes
	})

	// 只保留前20个
	if len(topAllocators) > 20 {
		topAllocators = topAllocators[:20]
	}

	return &MemoryProfileStats{
		TotalAllocations: uint64(len(pe.memoryProfiler.allocations)),
		TotalFrees:       uint64(len(pe.memoryProfiler.frees)),
		TopAllocators:    topAllocators,
	}
}

// getBlockStats 获取阻塞统计
func (pe *ProfilerEnhanced) getBlockStats() *BlockProfileStats {
	pe.blockProfiler.mu.RLock()
	defer pe.blockProfiler.mu.RUnlock()

	// 统计阻塞
	blockCounts := make(map[string]*BlockStat)
	totalDuration := time.Duration(0)

	for _, event := range pe.blockProfiler.events {
		key := event.Type
		if len(event.Stack) > 0 {
			frame := event.Stack[0]
			key = fmt.Sprintf("%s:%s", event.Type, frame.Function)
		}

		stat, exists := blockCounts[key]
		if !exists {
			function := key
			stat = &BlockStat{
				Function: function,
			}
			blockCounts[key] = stat
		}

		stat.BlockCount++
		stat.TotalTime += event.Duration
		totalDuration += event.Duration
	}

	// 计算平均时间
	for _, stat := range blockCounts {
		if stat.BlockCount > 0 {
			stat.AverageTime = stat.TotalTime / time.Duration(stat.BlockCount)
		}
	}

	// 转换为数组并排序
	topBlockers := make([]*BlockStat, 0)
	for _, stat := range blockCounts {
		topBlockers = append(topBlockers, stat)
	}

	sort.Slice(topBlockers, func(i, j int) bool {
		return topBlockers[i].TotalTime > topBlockers[j].TotalTime
	})

	// 只保留前20个
	if len(topBlockers) > 20 {
		topBlockers = topBlockers[:20]
	}

	return &BlockProfileStats{
		BlockCount:    len(pe.blockProfiler.events),
		TotalDuration: totalDuration,
		TopBlockers:   topBlockers,
	}
}

// getMutexStats 获取互斥锁统计
func (pe *ProfilerEnhanced) getMutexStats() *MutexProfileStats {
	pe.mutexProfiler.mu.RLock()
	defer pe.mutexProfiler.mu.RUnlock()

	// 统计竞争
	contentionCounts := make(map[string]*MutexContentionStat)
	totalWait := time.Duration(0)

	for _, contention := range pe.mutexProfiler.contentions {
		key := fmt.Sprintf("lock_%x", contention.LockAddress)

		stat, exists := contentionCounts[key]
		if !exists {
			stat = &MutexContentionStat{
				LockAddress: key,
			}
			contentionCounts[key] = stat
		}

		stat.WaitCount++
		stat.TotalWait += contention.WaitDuration
		totalWait += contention.WaitDuration
	}

	// 计算平均等待时间
	for _, stat := range contentionCounts {
		if stat.WaitCount > 0 {
			stat.AverageWait = stat.TotalWait / time.Duration(stat.WaitCount)
		}
	}

	// 转换为数组并排序
	topContentions := make([]*MutexContentionStat, 0)
	for _, stat := range contentionCounts {
		topContentions = append(topContentions, stat)
	}

	sort.Slice(topContentions, func(i, j int) bool {
		return topContentions[i].TotalWait > topContentions[j].TotalWait
	})

	// 只保留前20个
	if len(topContentions) > 20 {
		topContentions = topContentions[:20]
	}

	return &MutexProfileStats{
		ContentionCount: len(pe.mutexProfiler.contentions),
		TotalWait:       totalWait,
		TopContentions:  topContentions,
	}
}

// getHotFunctions 获取热点函数
func (pe *ProfilerEnhanced) getHotFunctions() []*HotFunction {
	pe.cpuProfiler.mu.RLock()
	defer pe.cpuProfiler.mu.RUnlock()

	funcCounts := make(map[string]int)
	totalSamples := len(pe.cpuProfiler.samples)

	for _, sample := range pe.cpuProfiler.samples {
		if len(sample.Stack) > 0 {
			frame := sample.Stack[0]
			key := fmt.Sprintf("%s:%d:%s", frame.Function, frame.Line, frame.File)
			funcCounts[key]++
		}
	}

	// 转换为HotFunction
	hotFunctions := make([]*HotFunction, 0)
	for key, count := range funcCounts {
		percentage := float64(count) / float64(totalSamples) * 100
		// 解析key（简化）
		function := key
		hotFunctions = append(hotFunctions, &HotFunction{
			Function:    function,
			SampleCount: count,
			Percentage:  percentage,
		})
	}

	// 排序
	sort.Slice(hotFunctions, func(i, j int) bool {
		return hotFunctions[i].SampleCount > hotFunctions[j].SampleCount
	})

	// 只保留前10个
	if len(hotFunctions) > 10 {
		hotFunctions = hotFunctions[:10]
	}

	return hotFunctions
}

// getTopAllocators 获取顶级分配器
func (pe *ProfilerEnhanced) getTopAllocators() []*AllocatorInfo {
	stats := pe.getMemoryStats()
	return stats.TopAllocators
}

// ExportFlameGraph 导出火焰图
func (pe *ProfilerEnhanced) ExportFlameGraph(writer io.Writer) error {
	pe.flameGraph.mu.RLock()
	defer pe.flameGraph.mu.RUnlock()

	// 生成JSON格式的火焰图
	data, err := json.MarshalIndent(pe.flameGraph.roots, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// ExportCallGraph 导出调用图
func (pe *ProfilerEnhanced) ExportCallGraph(writer io.Writer) error {
	pe.callGraph.mu.RLock()
	defer pe.callGraph.mu.RUnlock()

	graph := map[string]interface{}{
		"nodes": pe.callGraph.nodes,
		"edges": pe.callGraph.edges,
	}

	data, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// GetFlameGraph 获取火焰图
func (pe *ProfilerEnhanced) GetFlameGraph() []*FlameNode {
	pe.flameGraph.mu.RLock()
	defer pe.flameGraph.mu.RUnlock()

	return pe.flameGraph.roots
}

// GetCallGraph 获取调用图
func (pe *ProfilerEnhanced) GetCallGraph() (map[string]*CallGraphNode, []*CallGraphEdge) {
	pe.callGraph.mu.RLock()
	defer pe.callGraph.mu.RUnlock()

	return pe.callGraph.nodes, pe.callGraph.edges
}

// parseStack 解析栈信息
func parseStack(data []byte) []StackFrame {
	// 简化实现，实际应该解析Go的栈格式
	return []StackFrame{
		{
			Function: "unknown",
			File:     "unknown",
			Line:     0,
			Package:  "main",
		},
	}
}

// RecordAllocation 记录内存分配
func (mp *MemoryProfiler) RecordAllocation(size uint64, ptr uintptr, objType string) {
	if !mp.enabled || !mp.trackStacks {
		return
	}

	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	stack := parseStack(buf[:n])

	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.allocations = append(mp.allocations, &MemoryAllocation{
		Timestamp: time.Now(),
		Size:      size,
		Stack:     stack,
		Type:      objType,
		Pointer:   ptr,
	})
}

// RecordFree 记录内存释放
func (mp *MemoryProfiler) RecordFree(ptr uintptr, size uint64) {
	if !mp.enabled {
		return
	}

	mp.mu.Lock()
	defer mp.mu.Unlock()

	// 查找对应的分配
	var alloc *MemoryAllocation
	var allocIndex int
	for i, a := range mp.allocations {
		if a.Pointer == ptr {
			alloc = a
			allocIndex = i
			break
		}
	}

	if alloc != nil {
		lifetime := time.Since(alloc.Timestamp)
		mp.frees = append(mp.frees, &MemoryFree{
			Timestamp: time.Now(),
			Pointer:   ptr,
			Size:      size,
			Lifetime:  lifetime,
		})

		// 从分配列表中移除
		mp.allocations = append(mp.allocations[:allocIndex], mp.allocations[allocIndex+1:]...)
	}
}

// GetReport 获取分析报告
func (pe *ProfilerEnhanced) GetReport() *ProfilingReport {
	return &ProfilingReport{
		GeneratedAt:  time.Now(),
		CPUStats:     pe.getCPUStats(),
		MemoryStats:  pe.getMemoryStats(),
		BlockStats:   pe.getBlockStats(),
		MutexStats:   pe.getMutexStats(),
		HotFunctions: pe.getHotFunctions(),
		TopAllocators: pe.getTopAllocators(),
	}
}
