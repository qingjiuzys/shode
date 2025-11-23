package debugger

import (
	"context"
	"fmt"

	"gitee.com/com_818cloud/shode/pkg/engine"
	"gitee.com/com_818cloud/shode/pkg/environment"
	"gitee.com/com_818cloud/shode/pkg/module"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/sandbox"
	"gitee.com/com_818cloud/shode/pkg/stdlib"
	"gitee.com/com_818cloud/shode/pkg/types"
)

// RunMode determines how the debugger resumes execution.
type RunMode int

const (
	RunModeContinue RunMode = iota
	RunModeStep
)

// StopReason indicates why execution stopped.
type StopReason int

const (
	StopReasonNone StopReason = iota
	StopReasonEntry
	StopReasonBreakpoint
	StopReasonStep
	StopReasonCompleted
)

// CommandCallback is invoked after every executed command.
type CommandCallback func(result *engine.CommandResult)

// Session manages debugging state for a single script.
type Session struct {
	engine    *engine.ExecutionEngine
	parser    *parser.SimpleParser
	script    *types.ScriptNode
	program   string
	breaks    map[int]struct{}
	index     int
	stopEntry bool

	skipNextBreakpoint bool
}

// NewSession creates a new debugging session with a fresh execution engine.
func NewSession() *Session {
	env := environment.NewEnvironmentManager()
	std := stdlib.New()
	mod := module.NewModuleManager()
	sec := sandbox.NewSecurityChecker()

	return &Session{
		engine: engine.NewExecutionEngine(env, std, mod, sec),
		parser: parser.NewSimpleParser(),
		breaks: make(map[int]struct{}),
		index:  0,
		script: &types.ScriptNode{},
	}
}

// LoadProgram parses the given script file.
func (s *Session) LoadProgram(path string, stopOnEntry bool) error {
	script, err := s.parser.ParseFile(path)
	if err != nil {
		return err
	}

	s.script = script
	s.program = path
	s.index = 0
	s.stopEntry = stopOnEntry
	s.skipNextBreakpoint = false
	return nil
}

// Program returns the current script path.
func (s *Session) Program() string {
	return s.program
}

// SetBreakpoints configures active breakpoint lines.
func (s *Session) SetBreakpoints(lines []int) []int {
	s.breaks = make(map[int]struct{}, len(lines))
	validated := make([]int, 0, len(lines))
	for _, line := range lines {
		if line <= 0 {
			continue
		}
		s.breaks[line] = struct{}{}
		validated = append(validated, line)
	}
	return validated
}

// CurrentLine returns the line number for the current command.
func (s *Session) CurrentLine() int {
	if s.index >= len(s.script.Nodes) {
		if len(s.script.Nodes) == 0 {
			return 1
		}
		last := s.script.Nodes[len(s.script.Nodes)-1]
		return last.Position().Line
	}
	return s.script.Nodes[s.index].Position().Line
}

// Continue executes until the next stop condition.
func (s *Session) Continue(
	ctx context.Context,
	mode RunMode,
	cb CommandCallback,
) (StopReason, int, error) {
	if s.script == nil {
		return StopReasonCompleted, 0, fmt.Errorf("script not loaded")
	}

	targetIndex := -1
	if mode == RunModeStep {
		targetIndex = s.index + 1
	}

	for s.index < len(s.script.Nodes) {
		line := s.CurrentLine()

		if s.stopEntry {
			s.stopEntry = false
			s.skipNextBreakpoint = true
			return StopReasonEntry, line, nil
		}

		if err := ctx.Err(); err != nil {
			return StopReasonCompleted, line, err
		}

		if s.skipNextBreakpoint {
			s.skipNextBreakpoint = false
		} else if _, ok := s.breaks[line]; ok {
			s.skipNextBreakpoint = true
			return StopReasonBreakpoint, line, nil
		}

		result, err := s.executeCurrent(ctx)
		if err != nil {
			return StopReasonCompleted, line, err
		}
		if cb != nil && result != nil {
			cb(result)
		}

		s.index++

		if targetIndex >= 0 && s.index >= targetIndex {
			return StopReasonStep, s.CurrentLine(), nil
		}
	}

	return StopReasonCompleted, 0, nil
}

func (s *Session) executeCurrent(ctx context.Context) (*engine.CommandResult, error) {
	if s.index >= len(s.script.Nodes) {
		return nil, fmt.Errorf("no command to execute")
	}

	node := s.script.Nodes[s.index]
	switch n := node.(type) {
	case *types.CommandNode:
		return s.engine.ExecuteCommand(ctx, n)
	default:
		return nil, fmt.Errorf("unsupported node type: %T", n)
	}
}
