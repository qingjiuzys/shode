package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// JITCompiler Just-In-Time 编译器
type JITCompiler struct {
	cacheDir      string
	cache         *CompilationCache
	optimizer     *CodeOptimizer
	stats         *JITStats
	mu            sync.RWMutex
	enableJIT     bool
	enableCache   bool
}

// CompilationCache 编译缓存
type CompilationCache struct {
	items map[string]*CachedCompilation
	mu    sync.RWMutex
}

// CachedCompilation 缓存的编译结果
type CachedCompilation struct {
	Bytecode    []byte      // 编译后的字节码
	AST         *types.ScriptNode // AST (用于快速重新编译)
	CompileTime time.Time   // 编译时间
	HitCount    int64      // 命中次数
	Size        int64      // 缓存大小
	Version     string     // Shode 版本
}

// CodeOptimizer 代码优化器
type CodeOptimizer struct {
	optimizations []Optimization
	stats         *OptimizationStats
}

// Optimization 优化接口
type Optimization interface {
	Name() string
	Apply(node *types.ScriptNode) (*types.ScriptNode, error)
	Enabled() bool
}

// OptimizationStats 优化统计
type OptimizationStats struct {
	OptimizationsApplied map[string]int64
	NodesOptimized      int64
	TimeSpent           time.Duration
}

// JITStats JIT 统计
type JITStats struct {
	CacheHits      int64
	CacheMisses    int64
	Compilations   int64
	Recompilations int64
	CacheSize      int64
	AverageCompile time.Duration
}

// Bytecode 字节码表示
type Bytecode struct {
	Magic    [4]byte  // 魔数: "SHBC"
	Version  uint16   // 版本号
	Length   uint32   // 字节码长度
	Sections []Section // 段
}

// Section 字节码段
type Section struct {
	Type  string   // 类型: code, data, const, etc.
	Name  string   // 段名
	Data  []byte   // 数据
	Align uint32   // 对齐
}

// CompiledScript 编译后的脚本
type CompiledScript struct {
	Bytecode   []byte
	Constants  []Constant
	Functions  []CompiledFunction
	Globals    map[string]interface{}
	SourceMap  *SourceMap
}

// CompiledFunction 编译后的函数
type CompiledFunction struct {
	Name       string
	Bytecode   []byte
	NumParams  int
	NumLocals  int
	StackSize  int
	SourcePos  int
}

// Constant 常量
type Constant struct {
	Type  string // string, number, boolean, etc.
	Value interface{}
}

// SourceMap 源码映射
type SourceMap struct {
	FilePath  string
	Mappings  []Mapping
}

// Mapping 源码映射项
type Mapping struct {
	GeneratedPos int
	SourcePos    int
	Name         string
}

// NewJITCompiler 创建 JIT 编译器
func NewJITCompiler(cacheDir string, enableJIT, enableCache bool) *JITCompiler {
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "shode-jit-cache")
	}

	jit := &JITCompiler{
		cacheDir:    cacheDir,
		cache:       NewCompilationCache(),
		optimizer:   NewCodeOptimizer(),
		stats:       &JITStats{},
		enableJIT:   enableJIT,
		enableCache: enableCache,
	}

	// 加载现有缓存
	if enableCache {
		jit.loadCache()
	}

	return jit
}

// Compile 编译脚本
func (jit *JITCompiler) Compile(ctx context.Context, script *types.ScriptNode, sourcePath string) (*CompiledScript, error) {
	cacheKey := jit.generateCacheKey(script, sourcePath)

	// 1. 检查缓存
	if jit.enableCache {
		if cached, found := jit.cache.Get(cacheKey); found {
			jit.stats.CacheHits++
			cached.HitCount++
			return jit.loadCompiledScript(cached)
		}
		jit.stats.CacheMisses++
	}

	// 2. 运行优化器
	optimized, err := jit.optimizer.Optimize(script)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	// 3. 编译为字节码
	compiled, err := jit.compileToBytecode(optimized, sourcePath)
	if err != nil {
		return nil, fmt.Errorf("bytecode compilation failed: %w", err)
	}

	// 4. 保存到缓存
	if jit.enableCache {
		cached := &CachedCompilation{
			Bytecode:    compiled.Bytecode,
			AST:         optimized,
			CompileTime: time.Now(),
			HitCount:    0,
			Size:        int64(len(compiled.Bytecode)),
			Version:     "0.7.0",
		}
		jit.cache.Set(cacheKey, cached)
		jit.stats.Compilations++
	}

	return compiled, nil
}

// compileToBytecode 编译为字节码
func (jit *JITCompiler) compileToBytecode(script *types.ScriptNode, sourcePath string) (*CompiledScript, error) {
	compiled := &CompiledScript{
		Constants: make([]Constant, 0),
		Functions: make([]CompiledFunction, 0),
		Globals:   make(map[string]interface{}),
		SourceMap: jit.buildSourceMap(sourcePath),
	}

	// 简化版字节码生成
	// TODO: 实现完整的字节码生成
	bytecode := &Bytecode{
		Magic:   [4]byte{'S', 'H', 'B', 'C'},
		Version: 1,
		Sections: []Section{
			{
				Type: "code",
				Name: "main",
				Data: jit.generateBytecode(script),
			},
		},
	}

	data, err := json.Marshal(bytecode)
	if err != nil {
		return nil, err
	}

	compiled.Bytecode = data
	return compiled, nil
}

// generateBytecode 生成字节码
func (jit *JITCompiler) generateBytecode(script *types.ScriptNode) []byte {
	var bytecode []byte

	// 简单的字节码生成
	// 格式: [opcode][operand...]
	for _, node := range script.Nodes {
		bytecode = append(bytecode, jit.compileNode(node)...)
	}

	// 返回指令
	bytecode = append(bytecode, 0x00) // HALT

	return bytecode
}

// compileNode 编译节点
func (jit *JITCompiler) compileNode(node types.Node) []byte {
	var bytecode []byte

	switch n := node.(type) {
	case *types.CommandNode:
		bytecode = append(bytecode, 0x01) // CALL
		bytecode = append(bytecode, []byte(n.Name)...)
		bytecode = append(bytecode, 0x00) // NULL terminator

	case *types.AssignmentNode:
		bytecode = append(bytecode, 0x02) // STORE
		bytecode = append(bytecode, []byte(n.Name)...)
		bytecode = append(bytecode, 0x00)

	case *types.IfNode:
		bytecode = append(bytecode, 0x03) // IF
		bytecode = append(bytecode, jit.compileNode(n.Condition)...)
		bytecode = append(bytecode, jit.compileScript(n.Then)...)

	case *types.ForNode:
		bytecode = append(bytecode, 0x04) // FOR
		bytecode = append(bytecode, []byte(n.Variable)...)
		bytecode = append(bytecode, 0x00)
		bytecode = append(bytecode, jit.compileScript(n.Body)...)

	case *types.WhileNode:
		bytecode = append(bytecode, 0x05) // WHILE
		bytecode = append(bytecode, jit.compileNode(n.Condition)...)
		bytecode = append(bytecode, jit.compileScript(n.Body)...)

	case *types.FunctionNode:
		bytecode = append(bytecode, 0x06) // FUNC
		bytecode = append(bytecode, []byte(n.Name)...)
		bytecode = append(bytecode, 0x00)
		bytecode = append(bytecode, jit.compileScript(n.Body)...)

	case *types.PipeNode:
		bytecode = append(bytecode, 0x07) // PIPE
		bytecode = append(bytecode, jit.compileNode(n.Left)...)
		bytecode = append(bytecode, jit.compileNode(n.Right)...)

	default:
		// Unknown node type
	}

	return bytecode
}

// compileScript 编译脚本节点
func (jit *JITCompiler) compileScript(script *types.ScriptNode) []byte {
	var bytecode []byte

	for _, node := range script.Nodes {
		bytecode = append(bytecode, jit.compileNode(node)...)
	}

	return bytecode
}

// buildSourceMap 构建源码映射
func (jit *JITCompiler) buildSourceMap(sourcePath string) *SourceMap {
	return &SourceMap{
		FilePath: sourcePath,
		Mappings: []Mapping{},
	}
}

// loadCompiledScript 从缓存加载编译后的脚本
func (jit *JITCompiler) loadCompiledScript(cached *CachedCompilation) (*CompiledScript, error) {
	var compiled CompiledScript
	err := json.Unmarshal(cached.Bytecode, &compiled)
	if err != nil {
		return nil, err
	}
	return &compiled, nil
}

// generateCacheKey 生成缓存键
func (jit *JITCompiler) generateCacheKey(script *types.ScriptNode, sourcePath string) string {
	// 基于脚本内容生成哈希
	// 简化版：使用脚本字符串哈希
	// TODO: 使用更强大的哈希算法
	return fmt.Sprintf("%s:%x", sourcePath, script.String())
}

// loadCache 从磁盘加载缓存
func (jit *JITCompiler) loadCache() error {
	if err := os.MkdirAll(jit.cacheDir, 0755); err != nil {
		return err
	}

	files, err := os.ReadDir(jit.cacheDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".cache" {
			continue
		}

		cachePath := filepath.Join(jit.cacheDir, file.Name())
		data, err := os.ReadFile(cachePath)
		if err != nil {
			continue
		}

		var cached CachedCompilation
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}

		key := file.Name()[:len(file.Name())-6] // Remove .cache extension
		jit.cache.Set(key, &cached)
		jit.stats.CacheSize += cached.Size
	}

	return nil
}

// saveCache 保存缓存到磁盘
func (jit *JITCompiler) saveCache() error {
	if err := os.MkdirAll(jit.cacheDir, 0755); err != nil {
		return err
	}

	jit.cache.mu.RLock()
	defer jit.cache.mu.RUnlock()

	for key, cached := range jit.cache.items {
		data, err := json.Marshal(cached)
		if err != nil {
			continue
		}

		cachePath := filepath.Join(jit.cacheDir, key+".cache")
		if err := os.WriteFile(cachePath, data, 0644); err != nil {
			return err
		}
	}

	return nil
}

// ClearCache 清空缓存
func (jit *JITCompiler) ClearCache() error {
	jit.cache.mu.Lock()
	jit.cache.items = make(map[string]*CachedCompilation)
	jit.cache.mu.Unlock()

	// 清空磁盘缓存
	if err := os.RemoveAll(jit.cacheDir); err != nil {
		return err
	}

	return os.MkdirAll(jit.cacheDir, 0755)
}

// GetStats 获取 JIT 统计
func (jit *JITCompiler) GetStats() *JITStats {
	return jit.stats
}

// WarmupCache 预热缓存
func (jit *JITCompiler) WarmupCache(directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sh" {
			path := filepath.Join(directory, file.Name())

			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			parser := parser.NewSimpleParser()
			script, err := parser.ParseString(string(content))
			if err != nil {
				continue
			}

			// 编译并缓存
			_, err = jit.Compile(context.Background(), script, path)
			if err != nil {
				continue
			}
		}
	}

	return jit.saveCache()
}

// NewCompilationCache 创建编译缓存
func NewCompilationCache() *CompilationCache {
	return &CompilationCache{
		items: make(map[string]*CachedCompilation),
	}
}

// Get 获取缓存项
func (c *CompilationCache) Get(key string) (*CachedCompilation, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	return item, found
}

// Set 设置缓存项
func (c *CompilationCache) Set(key string, cached *CachedCompilation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = cached
}

// Delete 删除缓存项
func (c *CompilationCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear 清空缓存
func (c *CompilationCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*CachedCompilation)
}

// Size 获取缓存大小
func (c *CompilationCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// NewCodeOptimizer 创建代码优化器
func NewCodeOptimizer() *CodeOptimizer {
	return &CodeOptimizer{
		optimizations: []Optimization{
			&DeadCodeElimination{},
			&ConstantFolding{},
			&LoopUnrolling{},
			&InlineExpansion{},
		},
		stats: &OptimizationStats{
			OptimizationsApplied: make(map[string]int64),
		},
	}
}

// Optimize 优化脚本
func (co *CodeOptimizer) Optimize(script *types.ScriptNode) (*types.ScriptNode, error) {
	start := time.Now()

	optimized := script

	// 应用所有优化
	for _, opt := range co.optimizations {
		if !opt.Enabled() {
			continue
		}

		var err error
		optimized, err = opt.Apply(optimized)
		if err != nil {
			// 优化失败，使用未优化的版本
			continue
		}

		co.stats.OptimizationsApplied[opt.Name()]++
		co.stats.NodesOptimized++
	}

	co.stats.TimeSpent = time.Since(start)
	return optimized, nil
}

// DeadCodeElimination 死代码消除
type DeadCodeElimination struct{}

func (dce *DeadCodeElimination) Name() string {
	return "dead_code_elimination"
}

func (dce *DeadCodeElimination) Enabled() bool {
	return true
}

func (dce *DeadCodeElimination) Apply(node *types.ScriptNode) (*types.ScriptNode, error) {
	var optimizedNodes []types.Node

	for _, child := range node.Nodes {
		if !dce.isDeadCode(child) {
			optimizedNodes = append(optimizedNodes, child)
		}
	}

	return &types.ScriptNode{
		Pos:   node.Pos,
		Nodes: optimizedNodes,
	}, nil
}

func (dce *DeadCodeElimination) isDeadCode(node types.Node) bool {
	// 简化版：检测注释和空行
	switch n := node.(type) {
	case *types.CommandNode:
		return n.Name == "" || n.Name == "#"
	default:
		return false
	}
}

// ConstantFolding 常量折叠
type ConstantFolding struct{}

func (cf *ConstantFolding) Name() string {
	return "constant_folding"
}

func (cf *ConstantFolding) Enabled() bool {
	return true
}

func (cf *ConstantFolding) Apply(node *types.ScriptNode) (*types.ScriptNode, error) {
	// TODO: 实现常量折叠优化
	// 例如: x = 1 + 2 → x = 3
	return node, nil
}

// LoopUnrolling 循环展开
type LoopUnrolling struct{}

func (lu *LoopUnrolling) Name() string {
	return "loop_unrolling"
}

func (lu *LoopUnrolling) Enabled() bool {
	return false // 默认禁用，可通过配置启用
}

func (lu *LoopUnrolling) Apply(node *types.ScriptNode) (*types.ScriptNode, error) {
	// TODO: 实现循环展开优化
	return node, nil
}

// InlineExpansion 内联展开
type InlineExpansion struct{}

func (ie *InlineExpansion) Name() string {
	return "inline_expansion"
}

func (ie *InlineExpansion) Enabled() bool {
	return true
}

func (ie *InlineExpansion) Apply(node *types.ScriptNode) (*types.ScriptNode, error) {
	// TODO: 实现函数内联优化
	return node, nil
}

// SaveMetrics 保存 JIT 指标
func (jit *JITCompiler) SaveMetrics(w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(jit.stats)
}
