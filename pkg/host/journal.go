package host

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Frame is one journal line.
type Frame struct {
	Seq  int64           `json:"seq"`
	T    int64           `json:"t"` // unix ms
	From string          `json:"from"`
	Msg  json.RawMessage `json:"msg"`
}

// ReadJournal returns the frames in path with seq > since. With since < 0
// it returns the last -since frames.
func ReadJournal(path string, since int64) ([]Frame, int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = f.Close() }()
	var out []Frame
	off, err := scanFrames(f, 0, func(fr Frame) {
		if since < 0 || fr.Seq > since {
			out = append(out, fr)
		}
	})
	if since < 0 && len(out) > int(-since) {
		out = out[len(out)-int(-since):]
	}
	return out, off, err
}

// FollowJournal calls fn for each frame appended to path after byte offset
// off until ctx ends. The journal is append-only, so polling is enough.
func FollowJournal(ctx context.Context, path string, off int64, fn func(Frame)) error {
	for {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		off, err = scanFrames(f, off, fn)
		_ = f.Close()
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(200 * time.Millisecond):
		}
	}
}

// scanFrames reads complete lines from off and returns the offset after the
// last complete one, so a half-written frame is read again next time.
func scanFrames(f *os.File, off int64, fn func(Frame)) (int64, error) {
	if _, err := f.Seek(off, io.SeekStart); err != nil {
		return off, err
	}
	br := bufio.NewReaderSize(f, 1<<20)
	for {
		line, err := br.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return off, nil
			}
			return off, err
		}
		off += int64(len(line))
		var fr Frame
		if json.Unmarshal(line, &fr) == nil {
			fn(fr)
		}
	}
}

// Renderer turns journal frames into readable text: agent message chunks
// joined, one line per tool call state change, plans, and turn boundaries.
// Frames it does not understand are skipped.
type Renderer struct {
	W      io.Writer
	inText bool              // last output was an unterminated message chunk
	tools  map[string]string // toolCallId -> title
}

type rpcMsg struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

type sessionUpdate struct {
	Update struct {
		Kind    string `json:"sessionUpdate"`
		Content struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		ToolCallID string `json:"toolCallId"`
		Title      string `json:"title"`
		Status     string `json:"status"`
		Entries    []struct {
			Content string `json:"content"`
			Status  string `json:"status"`
		} `json:"entries"`
	} `json:"update"`
}

// Render writes one frame.
func (r *Renderer) Render(fr Frame) {
	if r.tools == nil {
		r.tools = map[string]string{}
	}
	var m rpcMsg
	if json.Unmarshal(fr.Msg, &m) != nil {
		return
	}
	switch {
	case m.Method == "session/update":
		r.update(m.Params)
	case m.Method == "session/prompt":
		var p struct {
			Prompt []struct {
				Text string `json:"text"`
			} `json:"prompt"`
		}
		_ = json.Unmarshal(m.Params, &p)
		var text string
		if len(p.Prompt) > 0 {
			text = p.Prompt[0].Text
		}
		r.line(fmt.Sprintf("=== turn %s: %s", stamp(fr.T), firstLine(text, 100)))
	case m.Method == "session/cancel":
		r.line("=== cancel requested")
	case m.Method == "" && len(m.Result) > 0:
		// Only session/prompt results carry a stopReason.
		var res struct {
			StopReason string `json:"stopReason"`
		}
		if json.Unmarshal(m.Result, &res) == nil && res.StopReason != "" {
			r.line(fmt.Sprintf("=== turn end %s: %s", stamp(fr.T), res.StopReason))
		}
	case m.Method == "" && m.Error != nil && fr.From == "agent":
		r.line("=== error: " + firstLine(m.Error.Message, 200))
	}
}

func (r *Renderer) update(params json.RawMessage) {
	var u sessionUpdate
	if json.Unmarshal(params, &u) != nil {
		return
	}
	up := u.Update
	switch up.Kind {
	case "agent_message_chunk":
		if up.Content.Type == "text" || up.Content.Type == "" {
			_, _ = io.WriteString(r.W, up.Content.Text)
			r.inText = up.Content.Text != "" && !strings.HasSuffix(up.Content.Text, "\n")
		}
	case "tool_call", "tool_call_update":
		title := up.Title
		if title == "" {
			title = r.tools[up.ToolCallID]
		}
		if title != "" {
			r.tools[up.ToolCallID] = title
		}
		// Updates without a status are output deltas; skip them.
		if up.Status == "" && up.Kind == "tool_call_update" {
			return
		}
		status := up.Status
		if status == "" {
			status = "pending"
		}
		r.line(fmt.Sprintf("[tool] %s (%s)", title, status))
	case "plan":
		r.line("[plan]")
		for _, e := range up.Entries {
			r.line(fmt.Sprintf("  %s %s", planMark(e.Status), e.Content))
		}
	}
}

// line writes s on its own line, ending a pending message chunk first.
func (r *Renderer) line(s string) {
	if r.inText {
		_, _ = io.WriteString(r.W, "\n")
		r.inText = false
	}
	_, _ = io.WriteString(r.W, s+"\n")
}

func planMark(status string) string {
	switch status {
	case "completed":
		return "[x]"
	case "in_progress":
		return "[>]"
	}
	return "[ ]"
}

func firstLine(s string, n int) string {
	s, _, _ = strings.Cut(strings.TrimSpace(s), "\n")
	if len(s) > n {
		s = s[:n] + "..."
	}
	return s
}

func stamp(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format("15:04:05")
}

// RawFrame re-encodes a frame as one JSON line, for --json.
func RawFrame(fr Frame) []byte {
	b, _ := json.Marshal(fr)
	return append(bytes.TrimSpace(b), '\n')
}
