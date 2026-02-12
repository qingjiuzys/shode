// Package repl REPL 交互式执行环境
package repl

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// REPL 交互式执行环境
type REPL struct {
	prompt       string
	continuation string
	history      []string
	historyIndex int
	variables    map[string]interface{}
	multiline    bool
	buffer       strings.Builder
	running      bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// REPLConfig REPL 配置
type REPLConfig struct {
	Prompt       string
	Continuation string
	HistorySize  int
	EnableColor  bool
	EnableAutoComplete bool
}

// NewREPL 创建 REPL
func NewREPL(config *REPLConfig) *REPL {
	if config == nil {
		config = &REPLConfig{
			Prompt:       ">>> ",
			Continuation: "... ",
			HistorySize:  1000,
			EnableColor:  true,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &REPL{
		prompt:       config.Prompt,
		continuation: config.Continuation,
		history:      make([]string, 0, config.HistorySize),
		variables:    make(map[string]interface{}),
		running:      false,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start 启动 REPL
func (r *REPL) Start() error {
	r.running = true

	// 显示欢迎信息
	r.printWelcome()

	// 主循环
	reader := bufio.NewReader(os.Stdin)

	for r.running {
		// 显示提示符
		prompt := r.prompt
		if r.multiline {
			prompt = r.continuation
		}

		fmt.Print(prompt)

		// 读取输入
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		line = strings.TrimSpace(line)

		// 处理命令
		if err := r.handleInput(line); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	return nil
}

// Stop 停止 REPL
func (r *REPL) Stop() {
	r.running = false
	r.cancel()
}

// handleInput 处理输入
func (r *REPL) handleInput(input string) error {
	// 空行
	if input == "" {
		return nil
	}

	// 特殊命令
	if strings.HasPrefix(input, ".") {
		return r.handleCommand(input)
	}

	// 多行输入
	if r.multiline {
		r.buffer.WriteString("\n")
		r.buffer.WriteString(input)

		// 检查是否结束多行输入
		if r.isComplete(r.buffer.String()) {
			r.multiline = false
			input = r.buffer.String()
			r.buffer.Reset()
		} else {
			return nil
		}
	} else if !r.isComplete(input) {
		// 开始多行输入
		r.multiline = true
		r.buffer.WriteString(input)
		return nil
	}

	// 添加到历史
	r.addToHistory(input)

	// 解析和执行
	result, err := r.execute(input)
	if err != nil {
		return fmt.Errorf("execution error: %w", err)
	}

	// 显示结果
	if result != nil {
		fmt.Println(formatResult(result))
	}

	return nil
}

// handleCommand 处理特殊命令
func (r *REPL) handleCommand(cmd string) error {
	parts := strings.Fields(cmd)
	command := parts[0]

	switch command {
	case ".help":
		r.printHelp()
	case ".exit", ".quit":
		r.Stop()
	case ".clear":
		r.variables = make(map[string]interface{})
		fmt.Println("Cleared all variables")
	case ".history":
		r.printHistory()
	case ".vars":
		r.printVariables()
	case ".load":
		if len(parts) < 2 {
			return fmt.Errorf("usage: .load <filename>")
		}
		return r.loadFile(parts[1])
	case ".save":
		if len(parts) < 2 {
			return fmt.Errorf("usage: .save <filename>")
		}
		return r.saveFile(parts[1])
	case ".time":
		if len(parts) < 2 {
			return fmt.Errorf("usage: .time <expression>")
		}
		return r.timeExecution(strings.Join(parts[1:], " "))
	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}

// execute 执行代码
func (r *REPL) execute(input string) (interface{}, error) {
	// 简化实现：解析并执行
	// TODO: 实际实现需要集成解析器和执行器

	// 示例：简单的表达式求值
	if strings.Contains(input, "+") {
		parts := strings.Split(input, "+")
		if len(parts) == 2 {
			// 简化：只处理两个数相加
			return fmt.Sprintf("result: %s + %s", strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])), nil
		}
	}

	return fmt.Sprintf("executed: %s", input), nil
}

// createContext 创建执行上下文（简化实现）
func (r *REPL) createContext() context.Context {
	return context.WithValue(r.ctx, "variables", r.variables)
}

// timeExecution 测量执行时间
func (r *REPL) timeExecution(code string) error {
	start := time.Now()

	result, err := r.execute(code)
	duration := time.Since(start)

	if err != nil {
		return err
	}

	fmt.Printf("Result: %v\n", result)
	fmt.Printf("Time: %v\n", duration)

	return nil
}

// loadFile 加载文件
func (r *REPL) loadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	_, err = r.execute(string(content))
	return err
}

// saveFile 保存会话到文件
func (r *REPL) saveFile(filename string) error {
	content := strings.Join(r.history, "\n")

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Saved %d commands to %s\n", len(r.history), filename)
	return nil
}

// isComplete 检查代码是否完整
func (r *REPL) isComplete(code string) bool {
	// 简化实现：检查括号是否匹配
	openBraces := strings.Count(code, "{")
	closeBraces := strings.Count(code, "}")

	return openBraces == closeBraces
}

// addToHistory 添加到历史
func (r *REPL) addToHistory(input string) {
	r.history = append(r.history, input)
	r.historyIndex = len(r.history)
}

// printWelcome 打印欢迎信息
func (r *REPL) printWelcome() {
	fmt.Println("Shode REPL v1.0.0")
	fmt.Println("Type .help for available commands")
	fmt.Println("Type .exit to quit")
	fmt.Println()
}

// printHelp 打印帮助
func (r *REPL) printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  .help      Show this help message")
	fmt.Println("  .exit      Exit REPL")
	fmt.Println("  .clear     Clear all variables")
	fmt.Println("  .history   Show command history")
	fmt.Println("  .vars      Show all variables")
	fmt.Println("  .load      Load and execute a file")
	fmt.Println("  .save      Save session to file")
	fmt.Println("  .time      Measure execution time")
}

// printHistory 打印历史
func (r *REPL) printHistory() {
	fmt.Println("Command history:")
	for i, cmd := range r.history {
		fmt.Printf("  %d: %s\n", i+1, cmd)
	}
}

// printVariables 打印变量
func (r *REPL) printVariables() {
	fmt.Println("Variables:")
	for name, value := range r.variables {
		fmt.Printf("  %s = %v\n", name, value)
	}
}

// formatResult 格式化结果
func formatResult(result interface{}) string {
	if result == nil {
		return "undefined"
	}

	switch v := result.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// SetPrompt 设置提示符
func (r *REPL) SetPrompt(prompt string) {
	r.prompt = prompt
}

// GetHistory 获取历史
func (r *REPL) GetHistory() []string {
	return r.history
}

// GetVariables 获取变量
func (r *REPL) GetVariables() map[string]interface{} {
	return r.variables
}
