package debugger

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"

	"gitee.com/com_818cloud/shode/pkg/engine"
)

type dapRequest struct {
	Seq     int             `json:"seq"`
	Type    string          `json:"type"`
	Command string          `json:"command"`
	Args    json.RawMessage `json:"arguments"`
}

type dapResponse struct {
	Seq        int         `json:"seq"`
	Type       string      `json:"type"`
	RequestSeq int         `json:"request_seq"`
	Command    string      `json:"command"`
	Success    bool        `json:"success"`
	Message    string      `json:"message,omitempty"`
	Body       interface{} `json:"body,omitempty"`
}

type dapEvent struct {
	Seq   int         `json:"seq"`
	Type  string      `json:"type"`
	Event string      `json:"event"`
	Body  interface{} `json:"body,omitempty"`
}

// DAPServer implements a minimal Debug Adapter Protocol server over stdio.
type DAPServer struct {
	reader *bufio.Reader
	writer io.Writer

	session *Session
	lineMu  sync.Mutex

	nextSeq  int
	state    string
	lastLine int
}

// NewDAPServer creates a server bound to the provided IO streams.
func NewDAPServer(r io.Reader, w io.Writer) *DAPServer {
	return &DAPServer{
		reader:  bufio.NewReader(r),
		writer:  w,
		session: NewSession(),
		nextSeq: 1,
		state:   "initialized",
	}
}

// Run consumes requests until EOF or error.
func (s *DAPServer) Run(ctx context.Context) error {
	for {
		raw, err := s.readMessage()
		if err != nil {
			return err
		}
		if len(raw) == 0 {
			continue
		}

		var req dapRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			continue
		}

		if err := s.handleRequest(ctx, &req); err != nil {
			return err
		}
	}
}

func (s *DAPServer) handleRequest(ctx context.Context, req *dapRequest) error {
	switch req.Command {
	case "initialize":
		body := map[string]interface{}{
			"supportsConfigurationDoneRequest": true,
			"supportsEvaluateForHovers":        false,
			"supportsSetVariable":              false,
		}
		s.state = "started"
		s.sendResponse(req, true, "", body)
	case "launch":
		var args struct {
			Program     string `json:"program"`
			StopOnEntry bool   `json:"stopOnEntry"`
		}
		json.Unmarshal(req.Args, &args)
		if args.Program == "" {
			s.sendResponse(req, false, "program missing", nil)
			return nil
		}
		if err := s.session.LoadProgram(args.Program, args.StopOnEntry); err != nil {
			s.sendResponse(req, false, err.Error(), nil)
			return nil
		}
		s.state = "launched"
		s.sendResponse(req, true, "", nil)
		s.sendEvent("initialized", nil)
	case "setBreakpoints":
		var args struct {
			Source struct {
				Path string `json:"path"`
			} `json:"source"`
			Breakpoints []struct {
				Line int `json:"line"`
			} `json:"breakpoints"`
		}
		json.Unmarshal(req.Args, &args)
		lines := make([]int, 0, len(args.Breakpoints))
		for _, bp := range args.Breakpoints {
			lines = append(lines, bp.Line)
		}
		valid := s.session.SetBreakpoints(lines)
		respBps := make([]map[string]interface{}, len(valid))
		for i, line := range valid {
			respBps[i] = map[string]interface{}{
				"verified": true,
				"line":     line,
			}
		}
		s.sendResponse(req, true, "", map[string]interface{}{
			"breakpoints": respBps,
		})
	case "configurationDone":
		s.sendResponse(req, true, "", nil)
		go s.startRun(ctx, RunModeContinue)
	case "threads":
		body := map[string]interface{}{
			"threads": []map[string]interface{}{
				{
					"id":   1,
					"name": "main",
				},
			},
		}
		s.sendResponse(req, true, "", body)
	case "stackTrace":
		line := s.currentLine()
		if line == 0 {
			line = 1
		}
		body := map[string]interface{}{
			"stackFrames": []map[string]interface{}{
				{
					"id":     1,
					"name":   "shode",
					"line":   line,
					"column": 1,
					"source": map[string]interface{}{
						"path": s.session.Program(),
					},
				},
			},
			"totalFrames": 1,
		}
		s.sendResponse(req, true, "", body)
	case "scopes":
		body := map[string]interface{}{
			"scopes": []map[string]interface{}{
				{
					"name":               "Globals",
					"variablesReference": 0,
					"presentationHint":   "locals",
				},
			},
		}
		s.sendResponse(req, true, "", body)
	case "variables":
		body := map[string]interface{}{
			"variables": []interface{}{},
		}
		s.sendResponse(req, true, "", body)
	case "continue":
		s.sendResponse(req, true, "", map[string]bool{"allThreadsContinued": true})
		go s.startRun(ctx, RunModeContinue)
	case "next":
		s.sendResponse(req, true, "", nil)
		go s.startRun(ctx, RunModeStep)
	case "disconnect":
		s.sendResponse(req, true, "", nil)
		s.sendEvent("terminated", nil)
		return io.EOF
	default:
		s.sendResponse(req, false, fmt.Sprintf("unsupported command %s", req.Command), nil)
	}
	return nil
}

func (s *DAPServer) startRun(ctx context.Context, mode RunMode) {
	reason, line, err := s.session.Continue(ctx, mode, s.handleCommandResult)
	if err != nil {
		s.sendOutput("stderr", err.Error())
		s.sendEvent("terminated", nil)
		return
	}

	s.lineMu.Lock()
	if line > 0 {
		s.lastLine = line
	}
	s.lineMu.Unlock()

	switch reason {
	case StopReasonCompleted:
		s.sendEvent("terminated", map[string]interface{}{"restart": false})
		s.sendEvent("exited", map[string]interface{}{"exitCode": 0})
	case StopReasonBreakpoint:
		s.sendStopped("breakpoint", line)
	case StopReasonEntry:
		s.sendStopped("entry", line)
	case StopReasonStep:
		s.sendStopped("step", line)
	}
}

func (s *DAPServer) handleCommandResult(result *engine.CommandResult) {
	if result == nil {
		return
	}
	if strings.TrimSpace(result.Output) != "" {
		s.sendOutput("stdout", result.Output)
	}
	if strings.TrimSpace(result.Error) != "" {
		s.sendOutput("stderr", result.Error)
	}
}

func (s *DAPServer) sendStopped(reason string, line int) {
	s.sendEvent("stopped", map[string]interface{}{
		"reason":   reason,
		"threadId": 1,
		"line":     line,
	})
}

func (s *DAPServer) currentLine() int {
	s.lineMu.Lock()
	defer s.lineMu.Unlock()
	if s.lastLine != 0 {
		return s.lastLine
	}
	return s.session.CurrentLine()
}

func (s *DAPServer) sendResponse(req *dapRequest, success bool, message string, body interface{}) {
	resp := dapResponse{
		Seq:        s.nextSeqValue(),
		Type:       "response",
		RequestSeq: req.Seq,
		Command:    req.Command,
		Success:    success,
		Message:    message,
		Body:       body,
	}
	s.writeMessage(resp)
}

func (s *DAPServer) sendEvent(event string, body interface{}) {
	evt := dapEvent{
		Seq:   s.nextSeqValue(),
		Type:  "event",
		Event: event,
		Body:  body,
	}
	s.writeMessage(evt)
}

func (s *DAPServer) sendOutput(category, output string) {
	if strings.HasSuffix(output, "\n") {
		// noop
	} else {
		output += "\n"
	}
	s.sendEvent("output", map[string]interface{}{
		"category": category,
		"output":   output,
	})
}

func (s *DAPServer) nextSeqValue() int {
	s.nextSeq++
	return s.nextSeq
}

func (s *DAPServer) writeMessage(v interface{}) {
	data, _ := json.Marshal(v)
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	s.writer.Write([]byte(header))
	s.writer.Write(data)
}

func (s *DAPServer) readMessage() ([]byte, error) {
	length := 0
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			value := strings.TrimSpace(line[len("content-length:"):])
			length, err = strconv.Atoi(value)
			if err != nil {
				return nil, err
			}
		}
	}

	if length == 0 {
		return nil, nil
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(s.reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
