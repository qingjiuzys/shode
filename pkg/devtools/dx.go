// Package devtools 提供开发者体验增强功能。
package devtools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// DeveloperExperience 开发者体验引擎
type DeveloperExperience struct {
	cli         *CLI
	hotReload   *HotReload
	debugger    *Debugger
	profiler    *Profiler
	docGen      *DocGenerator
	linter      *Linter
	formatter   *Formatter
	tester      *Tester
	mu          sync.RWMutex
}

// NewDeveloperExperience 创建开发者体验引擎
func NewDeveloperExperience() *DeveloperExperience {
	return &DeveloperExperience{
		cli:       NewCLI(),
		hotReload: NewHotReload(),
		debugger:  NewDebugger(),
		profiler:  NewProfiler(),
		docGen:    NewDocGenerator(),
		linter:    NewLinter(),
		formatter: NewFormatter(),
		tester:    NewTester(),
	}
}

// RunCLI 运行 CLI
func (dx *DeveloperExperience) RunCLI(ctx context.Context, args []string) error {
	return dx.cli.Run(ctx, args)
}

// EnableHotReload 启用热重载
func (dx *DeveloperExperience) EnableHotReload(ctx context.Context, dirs []string) error {
	return dx.hotReload.Watch(ctx, dirs)
}

// StartDebugging 开始调试
func (dx *DeveloperExperience) StartDebugging(ctx context.Context, config *DebugConfig) (*DebugSession, error) {
	return dx.debugger.Start(ctx, config)
}

// Profile 性能分析
func (dx *DeveloperExperience) Profile(ctx context.Context, target string, duration time.Duration) (*ProfileResult, error) {
	return dx.profiler.Profile(ctx, target, duration)
}

// GenerateDocs 生成文档
func (dx *DeveloperExperience) GenerateDocs(ctx context.Context, source string, output string) error {
	return dx.docGen.Generate(ctx, source, output)
}

// Lint 代码检查
func (dx *DeveloperExperience) Lint(ctx context.Context, files []string) ([]*LintResult, error) {
	return dx.linter.Lint(ctx, files)
}

// Format 格式化代码
func (dx *DeveloperExperience) Format(ctx context.Context, files []string) ([]*FormatResult, error) {
	return dx.formatr.Format(ctx, files)
}

// RunTests 运行测试
func (dx *DeveloperExperience) RunTests(ctx context.Context, pattern string) (*TestResult, error) {
	return dx.tester.Run(ctx, pattern)
}

// CLI 命令行工具
type CLI struct {
	commands map[string]*Command
	config   *CLIConfig
	plugins  []*CLIPlugin
	mu       sync.RWMutex
}

// Command 命令
type Command struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Usage       string                 `json:"usage"`
	Handler     func(ctx context.Context, args []string) error `json:"-"`
	Flags       []*Flag                `json:"flags"`
	Aliases     []string               `json:"aliases"`
}

// Flag 标志
type Flag struct {
	Name        string `json:"name"`
	Short       string `json:"short"`
	Description string `json:"description"`
	Default     string `json:"default"`
	Required    bool   `json:"required"`
}

// CLIConfig CLI 配置
type CLIConfig struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Prompt      string `json:"prompt"`
}

// CLIPlugin CLI 插件
type CLIPlugin struct {
	Name     string                 `json:"name"`
	Version  string                 `json:"version"`
	Commands []*Command             `json:"commands"`
	Hooks    map[string]interface{} `json:"hooks"`
}

// NewCLI 创建 CLI
func NewCLI() *CLI {
	return &CLI{
		commands: make(map[string]*Command),
		config: &CLIConfig{
			Name:    "shode",
			Version: "0.15.0",
			Prompt:  "> ",
		},
		plugins: make([]*CLIPlugin, 0),
	}
}

// Register 注册命令
func (cli *CLI) Register(command *Command) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	cli.commands[command.Name] = command

	// 注册别名
	for _, alias := range command.Aliases {
		cli.commands[alias] = command
	}
}

// Run 运行
func (cli *CLI) Run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return cli.help()
	}

	cmdName := args[0]
	cmdArgs := args[1:]

	cli.mu.RLock()
	command, exists := cli.commands[cmdName]
	cli.mu.RUnlock()

	if !exists {
		return fmt.Errorf("command not found: %s", cmdName)
	}

	return command.Handler(ctx, cmdArgs)
}

// help 帮助
func (cli *CLI) help() error {
	fmt.Printf("%s v%s - %s\n\n", cli.config.Name, cli.config.Version, cli.config.Description)
	fmt.Println("Commands:")

	cli.mu.RLock()
	defer cli.mu.RUnlock()

	for _, cmd := range cli.commands {
		if cmd.Name == cmd.Name { // 避免重复打印别名
			fmt.Printf("  %-20s %s\n", cmd.Name, cmd.Description)
		}
	}

	return nil
}

// InstallPlugin 安装插件
func (cli *CLI) InstallPlugin(plugin *CLIPlugin) {
	cli.mu.Lock()
	defer cli.mu.Unlock()

	cli.plugins = append(cli.plugins, plugin)

	// 注册插件的命令
	for _, cmd := range plugin.Commands {
		cli.commands[cmd.Name] = cmd
	}
}

// HotReload 热重载
type HotReload struct {
	watchers  map[string]*FileWatcher
	debounce  time.Duration
	ignores   []string
	hooks     []ReloadHook
	mu        sync.RWMutex
}

// FileWatcher 文件监视器
type FileWatcher struct {
	Path      string
	Recursive bool
	Events    chan *FileEvent
	Stop      chan bool
}

// FileEvent 文件事件
type FileEvent struct {
	Path     string
	Type     string // "create", "write", "remove", "rename"
	Time     time.Time
}

// ReloadHook 重载钩子
type ReloadHook func(event *FileEvent) error

// NewHotReload 创建热重载
func NewHotReload() *HotReload {
	return &HotReload{
		watchers: make(map[string]*FileWatcher),
		debounce: 100 * time.Millisecond,
		ignores:  []string{".git", "node_modules", ".DS_Store"},
		hooks:    make([]ReloadHook, 0),
	}
}

// Watch 监视
func (hr *HotReload) Watch(ctx context.Context, dirs []string) error {
	for _, dir := range dirs {
		watcher := &FileWatcher{
			Path:     dir,
			Events:   make(chan *FileEvent, 100),
			Stop:     make(chan bool),
		}

		hr.watchers[dir] = watcher

		// 启动监视
		go hr.watch(watcher)
	}

	return nil
}

// watch 监视
func (hr *HotReload) watch(watcher *FileWatcher) {
	ticker := time.NewTicker(hr.debounce)
	defer ticker.Stop()

	for {
		select {
		case event := <-watcher.Events:
			// 执行钩子
			for _, hook := range hr.hooks {
				_ = hook(event)
			}
		case <-watcher.Stop:
			return
		case <-ticker.C:
			// 检查文件变化
			hr.checkChanges(watcher)
		}
	}
}

// checkChanges 检查变化
func (hr *HotReload) checkChanges(watcher *FileWatcher) {
	// 简化实现，不实际检查
}

// AddHook 添加钩子
func (hr *HotReload) AddHook(hook ReloadHook) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.hooks = append(hr.hooks, hook)
}

// Stop 停止
func (hr *HotReload) Stop(path string) {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	if watcher, exists := hr.watchers[path]; exists {
		watcher.Stop <- true
		delete(hr.watchers, path)
	}
}

// Debugger 调试器
type Debugger struct {
	sessions map[string]*DebugSession
	breakpoints map[string][]*Breakpoint
	mu        sync.RWMutex
}

// DebugSession 调试会话
type DebugSession struct {
	ID          string                 `json:"id"`
	Target      string                 `json:"target"`
	Config      *DebugConfig           `json:"config"`
	Status      string                 `json:"status"`
	Variables   map[string]interface{} `json:"variables"`
	CallStack   []*StackFrame          `json:"call_stack"`
	StartTime   time.Time              `json:"start_time"`
}

// DebugConfig 调试配置
type DebugConfig struct {
	Program     string                 `json:"program"`
	Args        []string               `json:"args"`
	Env         map[string]string      `json:"env"`
	WorkingDir  string                 `json:"working_dir"`
	Port        int                    `json:"port"`
}

// Breakpoint 断点
type Breakpoint struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Condition string `json:"condition"`
	Enabled  bool   `json:"enabled"`
	HitCount int    `json:"hit_count"`
}

// StackFrame 栈帧
type StackFrame struct {
	Func   string `json:"func"`
	File   string `json:"file"`
	Line   int    `json:"line"`
	Vars   map[string]interface{} `json:"vars"`
}

// NewDebugger 创建调试器
func NewDebugger() *Debugger {
	return &Debugger{
		sessions:    make(map[string]*DebugSession),
		breakpoints: make(map[string][]*Breakpoint),
	}
}

// Start 开始调试
func (d *Debugger) Start(ctx context.Context, config *DebugConfig) (*DebugSession, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	session := &DebugSession{
		ID:        generateSessionID(),
		Target:    config.Program,
		Config:    config,
		Status:    "running",
		Variables: make(map[string]interface{}),
		CallStack: make([]*StackFrame, 0),
		StartTime: time.Now(),
	}

	d.sessions[session.ID] = session

	return session, nil
}

// Stop 停止调试
func (d *Debugger) Stop(sessionID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	session, exists := d.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Status = "stopped"

	return nil
}

// AddBreakpoint 添加断点
func (d *Debugger) AddBreakpoint(file string, line int, condition string) (*Breakpoint, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	bp := &Breakpoint{
		ID:        generateBreakpointID(),
		File:      file,
		Line:      line,
		Condition: condition,
		Enabled:   true,
		HitCount:  0,
	}

	d.breakpoints[file] = append(d.breakpoints[file], bp)

	return bp, nil
}

// GetVariables 获取变量
func (d *Debugger) GetVariables(sessionID string) (map[string]interface{}, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	session, exists := d.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session.Variables, nil
}

// Profiler 性能分析器
type Profiler struct {
	profiles map[string]*ProfileResult
	mu       sync.RWMutex
}

// ProfileResult 分析结果
type ProfileResult struct {
	ID          string             `json:"id"`
	Target      string             `json:"target"`
	Duration    time.Duration      `json:"duration"`
	CPUProfile  *CPUProfile        `json:"cpu_profile"`
	MemoryProfile *MemoryProfile   `json:"memory_profile"`
	HeapProfile *HeapProfile       `json:"heap_profile"`
	Timestamp   time.Time          `json:"timestamp"`
}

// CPUProfile CPU 分析
type CPUProfile struct {
	Samples    []*Sample   `json:"samples"`
	Function   string      `json:"function"`
	Duration   time.Duration `json:"duration"`
}

// Sample 样本
type Sample struct {
	Function string `json:"function"`
	Count    int    `json:"count"`
}

// MemoryProfile 内存分析
type MemoryProfile struct {
	HeapSize   int64 `json:"heap_size"`
	StackSize  int64 `json:"stack_size"`
	Goroutines int   `json:"goroutines"`
}

// HeapProfile 堆分析
type HeapProfile struct {
	Objects    []*HeapObject `json:"objects"`
	TotalSize  int64         `json:"total_size"`
}

// HeapObject 堆对象
type HeapObject struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
	Size  int64  `json:"size"`
}

// NewProfiler 创建性能分析器
func NewProfiler() *Profiler {
	return &Profiler{
		profiles: make(map[string]*ProfileResult),
	}
}

// Profile 分析
func (p *Profiler) Profile(ctx context.Context, target string, duration time.Duration) (*ProfileResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := &ProfileResult{
		ID:       generateProfileID(),
		Target:   target,
		Duration: duration,
		CPUProfile: &CPUProfile{
			Samples: make([]*Sample, 0),
		},
		MemoryProfile: &MemoryProfile{},
		HeapProfile: &HeapProfile{
			Objects: make([]*HeapObject, 0),
		},
		Timestamp: time.Now(),
	}

	p.profiles[result.ID] = result

	return result, nil
}

// GetResult 获取结果
func (p *Profiler) GetResult(profileID string) (*ProfileResult, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result, exists := p.profiles[profileID]
	return result, exists
}

// DocGenerator 文档生成器
type DocGenerator struct {
	templates map[string]*DocTemplate
	formats   []string // "markdown", "html", "json"
	mu        sync.RWMutex
}

// DocTemplate 文档模板
type DocTemplate struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Format   string `json:"format"`
}

// NewDocGenerator 创建文档生成器
func NewDocGenerator() *DocGenerator {
	return &DocGenerator{
		templates: make(map[string]*DocTemplate),
		formats:   []string{"markdown", "html", "json"},
	}
}

// Generate 生成文档
func (dg *DocGenerator) Generate(ctx context.Context, source string, output string) error {
	// 扫描源文件
	files, err := dg.scanSource(source)
	if err != nil {
		return err
	}

	// 解析注释
	docs := dg.parseComments(files)

	// 生成文档
	for format := range dg.templates {
		outputPath := filepath.Join(output, format)
		_ = os.MkdirAll(outputPath, 0755)

		for _, doc := range docs {
			_ = dg.writeDoc(doc, outputPath, format)
		}
	}

	return nil
}

// scanSource 扫描源文件
func (dg *DocGenerator) scanSource(source string) ([]string, error) {
	var files []string

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (filepath.Ext(path) == ".go" || filepath.Ext(path) == ".js") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// parseComments 解析注释
func (dg *DocGenerator) parseComments(files []string) []*DocSection {
	sections := make([]*DocSection, 0)

	for _, file := range files {
		section := &DocSection{
			File:    file,
			Name:    filepath.Base(file),
			Package: "main",
			Content: make([]*DocContent, 0),
		}

		sections = append(sections, section)
	}

	return sections
}

// writeDoc 写入文档
func (dg *DocGenerator) writeDoc(section *DocSection, output string, format string) error {
	var content string

	switch format {
	case "markdown":
		content = fmt.Sprintf("# %s\n\n%s\n", section.Name, section.Package)
	case "html":
		content = fmt.Sprintf("<h1>%s</h1><p>%s</p>", section.Name, section.Package)
	}

	outputFile := filepath.Join(output, section.Name+"."+format)

	return os.WriteFile(outputFile, []byte(content), 0644)
}

// DocSection 文档节
type DocSection struct {
	File    string         `json:"file"`
	Name    string         `json:"name"`
	Package string         `json:"package"`
	Content []*DocContent  `json:"content"`
}

// DocContent 文档内容
type DocContent struct {
	Type        string `json:"type"` // "function", "struct", "const", "var"
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  []string `json:"parameters"`
	Returns     []string `json:"returns"`
}

// Linter 代码检查器
type Linter struct {
	rules   map[string]*LintRule
	config  *LintConfig
	mu      sync.RWMutex
}

// LintRule 检查规则
type LintRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"` // "error", "warning", "info"
	Category    string `json:"category"`
	Enabled     bool   `json:"enabled"`
}

// LintConfig 检查配置
type LintConfig struct {
	Rules       []string `json:"rules"`
	Excludes    []string `json:"excludes"`
	MaxComplexity int    `json:"max_complexity"`
}

// LintResult 检查结果
type LintResult struct {
	File      string       `json:"file"`
	Line      int          `json:"line"`
	Column    int          `json:"column"`
	Rule      string       `json:"rule"`
	Severity  string       `json:"severity"`
	Message   string       `json:"message"`
	Source    string       `json:"source"`
}

// NewLinter 创建检查器
func NewLinter() *Linter {
	return &Linter{
		rules:  make(map[string]*LintRule),
		config: &LintConfig{},
	}
}

// AddRule 添加规则
func (l *Linter) AddRule(rule *LintRule) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.rules[rule.Name] = rule
}

// Lint 检查
func (l *Linter) Lint(ctx context.Context, files []string) ([]*LintResult, error) {
	results := make([]*LintResult, 0)

	for _, file := range files {
		// 简化实现，返回空结果
		_ = file
	}

	return results, nil
}

// Formatter 格式化器
type Formatter struct {
	config  *FormatConfig
	mu      sync.RWMutex
}

// FormatConfig 格式化配置
type FormatConfig struct {
	IndentSize  int    `json:"indent_size"`
	IndentStyle string `json:"indent_style"` // "tab", "space"
	LineWidth   int    `json:"line_width"`
}

// FormatResult 格式化结果
type FormatResult struct {
	File     string `json:"file"`
	Modified bool   `json:"modified"`
	Changes  int    `json:"changes"`
}

// NewFormatter 创建格式化器
func NewFormatter() *Formatter {
	return &Formatter{
		config: &FormatConfig{
			IndentSize:  4,
			IndentStyle: "tab",
			LineWidth:   120,
		},
	}
}

// Format 格式化
func (f *Formatter) Format(ctx context.Context, files []string) ([]*FormatResult, error) {
	results := make([]*FormatResult, 0)

	for _, file := range files {
		result := &FormatResult{
			File:     file,
			Modified: false,
			Changes:  0,
		}

		// 简化实现
		results = append(results, result)
	}

	return results, nil
}

// Tester 测试器
type Tester struct {
	config   *TestConfig
	coverage *Coverage
	mu       sync.RWMutex
}

// TestConfig 测试配置
type TestConfig struct {
	Pattern    string   `json:"pattern"`
	Timeout    time.Duration `json:"timeout"`
	Parallel   bool     `json:"parallel"`
	Verbose    bool     `json:"verbose"`
	Coverage   bool     `json:"coverage"`
	Exclude    []string `json:"exclude"`
}

// TestResult 测试结果
type TestResult struct {
	Total     int               `json:"total"`
	Passed    int               `json:"passed"`
	Failed    int               `json:"failed"`
	Skipped   int               `json:"skipped"`
	Duration  time.Duration     `json:"duration"`
	Output    string            `json:"output"`
	Tests     []*TestCase       `json:"tests"`
	Coverage  *CoverageReport   `json:"coverage"`
}

// TestCase 测试用例
type TestCase struct {
	Name     string        `json:"name"`
	Package  string        `json:"package"`
	Status   string        `json:"status"`
	Duration time.Duration `json:"duration"`
	Output   string        `json:"output"`
}

// Coverage 覆盖率
type Coverage struct {
	packages map[string]*PackageCoverage
	mu       sync.RWMutex
}

// PackageCoverage 包覆盖率
type PackageCoverage struct {
	Name        string             `json:"name"`
	Files       map[string]*FileCoverage `json:"files"`
	Coverage    float64            `json:"coverage"`
}

// FileCoverage 文件覆盖率
type FileCoverage struct {
	Name     string `json:"name"`
	Coverage float64 `json:"coverage"`
	Lines    []int  `json:"lines"`
}

// CoverageReport 覆盖率报告
type CoverageReport struct {
	Total    float64 `json:"total"`
	Covered  int     `json:"covered"`
	TotalLines int   `json:"total_lines"`
}

// NewTester 创建测试器
func NewTester() *Tester {
	return &Tester{
		config: &TestConfig{
			Pattern:  "*_test.go",
			Timeout:  30 * time.Second,
			Parallel: true,
			Coverage: true,
		},
		coverage: &Coverage{
			packages: make(map[string]*PackageCoverage),
		},
	}
}

// Run 运行测试
func (t *Tester) Run(ctx context.Context, pattern string) (*TestResult, error) {
	result := &TestResult{
		Total:    0,
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Duration: 0,
		Tests:    make([]*TestCase, 0),
	}

	// 执行测试
	cmd := exec.CommandContext(ctx, "go", "test", "-v", "-cover", pattern)
	output, err := cmd.CombinedOutput()

	result.Output = string(output)

	if err != nil {
		result.Failed++
	} else {
		result.Passed++
	}

	return result, nil
}

// GetCoverage 获取覆盖率
func (t *Tester) GetCoverage(packageName string) (*PackageCoverage, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	coverage, exists := t.coverage.packages[packageName]
	return coverage, exists
}

// ProjectManager 项目管理器
type ProjectManager struct {
	projects map[string]*Project
	mu       sync.RWMutex
}

// Project 项目
type Project struct {
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Type        string                 `json:"type"`
	Config      map[string]interface{} `json:"config"`
	Dependencies []string              `json:"dependencies"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NewProjectManager 创建项目管理器
func NewProjectManager() *ProjectManager {
	return &ProjectManager{
		projects: make(map[string]*Project),
	}
}

// CreateProject 创建项目
func (pm *ProjectManager) CreateProject(name, path, projectType string) (*Project, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	project := &Project{
		Name:     name,
		Path:     path,
		Type:     projectType,
		Config:   make(map[string]interface{}),
		Dependencies: make([]string, 0),
		CreatedAt: time.Now(),
	}

	pm.projects[name] = project

	// 创建目录
	_ = os.MkdirAll(path, 0755)

	return project, nil
}

// GetProject 获取项目
func (pm *ProjectManager) GetProject(name string) (*Project, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	project, exists := pm.projects[name]
	return project, exists
}

// CodeSnippets 代码片段管理器
type CodeSnippets struct {
	snippets map[string]*Snippet
	categories map[string][]*Snippet
	mu       sync.RWMutex
}

// Snippet 代码片段
type Snippet struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Code        string   `json:"code"`
	Language    string   `json:"language"`
	Tags        []string `json:"tags"`
}

// NewCodeSnippets 创建代码片段管理器
func NewCodeSnippets() *CodeSnippets {
	return &CodeSnippets{
		snippets:   make(map[string]*Snippet),
		categories: make(map[string][]*Snippet),
	}
}

// AddSnippet 添加片段
func (cs *CodeSnippets) AddSnippet(snippet *Snippet) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.snippets[snippet.ID] = snippet

	for _, tag := range snippet.Tags {
		cs.categories[tag] = append(cs.categories[tag], snippet)
	}
}

// GetSnippet 获取片段
func (cs *CodeSnippets) GetSnippet(id string) (*Snippet, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	snippet, exists := cs.snippets[id]
	return snippet, exists
}

// Search 搜索片段
func (cs *CodeSnippets) Search(query string) []*Snippet {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	results := make([]*Snippet, 0)

	for _, snippet := range cs.snippets {
		if contains(snippet.Name, query) || contains(snippet.Description, query) {
			results = append(results, snippet)
		}
	}

	return results
}

// contains 包含
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// generateBreakpointID 生成断点 ID
func generateBreakpointID() string {
	return fmt.Sprintf("bp_%d", time.Now().UnixNano())
}

// generateProfileID 生成分析 ID
func generateProfileID() string {
	return fmt.Sprintf("profile_%d", time.Now().UnixNano())
}
