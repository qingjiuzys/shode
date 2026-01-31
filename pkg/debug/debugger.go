// Package debug 提供调试功能。
package debug

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Breakpoint 断点
type Breakpoint struct {
	File     string
	Line     int
	Condition string
	HitCount int
	Enabled  bool
}

// Debugger 调试器
type Debugger struct {
	breakpoints map[string]*Breakpoint // key: "file:line"
	callStack   []StackFrame
	variables   map[string]interface{}
	mu          sync.RWMutex
	running     bool
	paused      bool
	stepMode    StepMode
	onBreak     func(*Debugger)
}

// StackFrame 调用栈帧
type StackFrame struct {
	Function string
	File     string
	Line     int
	Locals   map[string]interface{}
}

// StepMode 单步执行模式
type StepMode int

const (
	StepNone StepMode = iota
	StepInto
	StepOver
	StepOut
)

// DebugEvent 调试事件
type DebugEvent struct {
	Type      string                 // "breakpoint", "step", "exception"
	File      string
	Line      int
	Message   string
	Variables map[string]interface{}
	Timestamp time.Time
}

// NewDebugger 创建调试器
func NewDebugger() *Debugger {
	return &Debugger{
		breakpoints: make(map[string]*Breakpoint),
		callStack:   make([]StackFrame, 0),
		variables:   make(map[string]interface{}),
		running:     false,
		paused:      false,
		stepMode:    StepNone,
	}
}

// Start 开始调试
func (d *Debugger) Start() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = true
	fmt.Println("Debugger started")
}

// Stop 停止调试
func (d *Debugger) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.running = false
	d.paused = false
	fmt.Println("Debugger stopped")
}

// Pause 暂停执行
func (d *Debugger) Pause() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = true
	fmt.Println("Execution paused")
}

// Continue 继续执行
func (d *Debugger) Continue() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = false
	d.stepMode = StepNone
	fmt.Println("Continuing execution...")
}

// StepInto 单步进入
func (d *Debugger) StepInto() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = false
	d.stepMode = StepInto
	fmt.Println("Stepping into...")
}

// StepOver 单步跳过
func (d *Debugger) StepOver() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = false
	d.stepMode = StepOver
	fmt.Println("Stepping over...")
}

// StepOut 单步跳出
func (d *Debugger) StepOut() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.paused = false
	d.stepMode = StepOut
	fmt.Println("Stepping out...")
}

// SetBreakpoint 设置断点
func (d *Debugger) SetBreakpoint(file string, line int, condition string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := fmt.Sprintf("%s:%d", file, line)
	d.breakpoints[key] = &Breakpoint{
		File:      file,
		Line:      line,
		Condition: condition,
		HitCount:  0,
		Enabled:   true,
	}

	fmt.Printf("Breakpoint set at %s:%d\n", file, line)
	return nil
}

// ClearBreakpoint 清除断点
func (d *Debugger) ClearBreakpoint(file string, line int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := fmt.Sprintf("%s:%d", file, line)
	delete(d.breakpoints, key)

	fmt.Printf("Breakpoint cleared at %s:%d\n", file, line)
	return nil
}

// ListBreakpoints 列出所有断点
func (d *Debugger) ListBreakpoints() []*Breakpoint {
	d.mu.RLock()
	defer d.mu.RUnlock()

	bps := make([]*Breakpoint, 0, len(d.breakpoints))
	for _, bp := range d.breakpoints {
		bps = append(bps, bp)
	}
	return bps
}

// CheckBreakpoint 检查是否命中断点
func (d *Debugger) CheckBreakpoint(file string, line int) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	key := fmt.Sprintf("%s:%d", file, line)
	bp, exists := d.breakpoints[key]
	if !exists || !bp.Enabled {
		return false
	}

	bp.HitCount++

	// 检查条件
	if bp.Condition != "" {
		// 简化的条件检查（实际应该用表达式求值）
		if !d.evaluateCondition(bp.Condition) {
			return false
		}
	}

	return true
}

// evaluateCondition 评估断点条件
func (d *Debugger) evaluateCondition(condition string) bool {
	// 简化实现，实际应该用完整的表达式求值器
	return true
}

// SetVariable 设置变量值
func (d *Debugger) SetVariable(name string, value interface{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.variables[name] = value
}

// GetVariable 获取变量值
func (d *Debugger) GetVariable(name string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	val, exists := d.variables[name]
	return val, exists
}

// ListVariables 列出所有变量
func (d *Debugger) ListVariables() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	vars := make(map[string]interface{})
	for k, v := range d.variables {
		vars[k] = v
	}
	return vars
}

// PushCallFrame 推入调用栈帧
func (d *Debugger) PushCallFrame(frame StackFrame) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.callStack = append(d.callStack, frame)
}

// PopCallFrame 弹出调用栈帧
func (d *Debugger) PopCallFrame() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.callStack) > 0 {
		d.callStack = d.callStack[:len(d.callStack)-1]
	}
}

// GetCallStack 获取调用栈
func (d *Debugger) GetCallStack() []StackFrame {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stack := make([]StackFrame, len(d.callStack))
	copy(stack, d.callStack)
	return stack
}

// OnBreak 设置断点回调
func (d *Debugger) OnBreak(handler func(*Debugger)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onBreak = handler
}

// IsPaused 检查是否暂停
func (d *Debugger) IsPaused() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.paused
}

// GetStepMode 获取单步模式
func (d *Debugger) GetStepMode() StepMode {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.stepMode
}

// SetBreakHandler 设置断点处理器
func (d *Debugger) SetBreakHandler(handler func(*Debugger)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.onBreak = handler
}

// REPLEDebugger REPL 调试器
type REPLDebugger struct {
	debugger *Debugger
}

// NewREPLDebugger 创建 REPL 调试器
func NewREPLDebugger() *REPLDebugger {
	return &REPLDebugger{
		debugger: NewDebugger(),
	}
}

// Start 启动 REPL 调试
func (rd *REPLDebugger) Start() {
	rd.debugger.Start()
	fmt.Println("Shode REPL Debugger")
	fmt.Println("====================")
	fmt.Println("Commands:")
	fmt.Println("  b <file>:<line>  - Set breakpoint")
	fmt.Println("  c <file>:<line>  - Clear breakpoint")
	fmt.Println("  lb               - List breakpoints")
	fmt.Println("  s                - Step")
	fmt.Println("  n                - Next")
	fmt.Println("  fin              - Finish")
	fmt.Println("  c                - Continue")
	fmt.Println("  p <var>          - Print variable")
	fmt.Println("  l                - List variables")
	fmt.Println("  bt               - Backtrace")
	fmt.Println("  q                - Quit")

	scanner := bufio.NewScanner(os.Stdin)

	for rd.debugger.running {
		fmt.Printf("(dbg) ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		rd.handleCommand(input)
	}
}

// handleCommand 处理调试命令
func (rd *REPLDebugger) handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	cmd := parts[0]

	switch cmd {
	case "b":
		if len(parts) < 2 {
			fmt.Println("Usage: b <file>:<line>")
			return
		}
		rd.handleSetBreakpoint(parts[1])

	case "c":
		if len(parts) < 2 {
			rd.debugger.Continue()
		} else {
			rd.handleClearBreakpoint(parts[1])
		}

	case "lb":
		rd.handleListBreakpoints()

	case "s":
		rd.debugger.StepInto()

	case "n":
		rd.debugger.StepOver()

	case "fin":
		rd.debugger.StepOut()

	case "p":
		if len(parts) < 2 {
			fmt.Println("Usage: p <variable>")
			return
		}
		rd.handlePrintVariable(parts[1])

	case "l":
		rd.handleListVariables()

	case "bt":
		rd.handleBacktrace()

	case "q":
		rd.debugger.Stop()

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

// handleSetBreakpoint 处理设置断点
func (rd *REPLDebugger) handleSetBreakpoint(spec string) {
	parts := strings.Split(spec, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid breakpoint format. Use <file>:<line>")
		return
	}

	file := parts[0]
	line := 0
	fmt.Sscanf(parts[1], "%d", &line)

	if err := rd.debugger.SetBreakpoint(file, line, ""); err != nil {
		fmt.Printf("Error setting breakpoint: %v\n", err)
	}
}

// handleClearBreakpoint 处理清除断点
func (rd *REPLDebugger) handleClearBreakpoint(spec string) {
	parts := strings.Split(spec, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid breakpoint format. Use <file>:<line>")
		return
	}

	file := parts[0]
	line := 0
	fmt.Sscanf(parts[1], "%d", &line)

	if err := rd.debugger.ClearBreakpoint(file, line); err != nil {
		fmt.Printf("Error clearing breakpoint: %v\n", err)
	}
}

// handleListBreakpoints 处理列出断点
func (rd *REPLDebugger) handleListBreakpoints() {
	bps := rd.debugger.ListBreakpoints()
	if len(bps) == 0 {
		fmt.Println("No breakpoints set")
		return
	}

	for _, bp := range bps {
		enabled := "y"
		if !bp.Enabled {
			enabled = "n"
		}
		fmt.Printf("%s:%d\t(hit count: %d\tenabled: %s)\n", bp.File, bp.Line, bp.HitCount, enabled)
	}
}

// handlePrintVariable 处理打印变量
func (rd *REPLDebugger) handlePrintVariable(name string) {
	val, exists := rd.debugger.GetVariable(name)
	if !exists {
		fmt.Printf("Variable '%s' not found\n", name)
		return
	}
	fmt.Printf("%s = %v\n", name, val)
}

// handleListVariables 处理列出变量
func (rd *REPLDebugger) handleListVariables() {
	vars := rd.debugger.ListVariables()
	if len(vars) == 0 {
		fmt.Println("No variables")
		return
	}

	for name, val := range vars {
		fmt.Printf("%s = %v\n", name, val)
	}
}

// handleBacktrace 处理回溯
func (rd *REPLDebugger) handleBacktrace() {
	stack := rd.debugger.GetCallStack()
	if len(stack) == 0 {
		fmt.Println("No stack trace available")
		return
	}

	for i, frame := range stack {
		fmt.Printf("#%d %s at %s:%d\n", i, frame.Function, frame.File, frame.Line)
	}
}

// GetStackTrace 获取堆栈跟踪
func GetStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// GetCaller 获取调用者信息
func GetCaller(skip int) (pc uintptr, file string, line int, ok bool) {
	return runtime.Caller(skip + 1)
}

// GetFunctionName 获取函数名
func GetFunctionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	return fn.Name()
}
