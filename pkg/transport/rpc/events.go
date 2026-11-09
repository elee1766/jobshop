package rpc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	_ "image/gif" // Only raster formats are accepted in transcripts.
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"
)

// ContextOpen and ContextClose delimit session.send context in a prompt.
// ACP text blocks carry no audience on the wire (the SDK drops it), so
// the event reader splits the block out by these lines.
const (
	ContextOpen  = "[jobshop context]"
	ContextClose = "[end of jobshop context]"
)

// SplitContext separates a prompt into its visible text and the context
// block, when it has one.
func SplitContext(s string) (text, ctx string) {
	before, rest, ok := strings.Cut(s, ContextOpen+"\n")
	if !ok {
		return s, ""
	}
	ctx, after, ok := strings.Cut(rest, "\n"+ContextClose)
	if !ok {
		return s, ""
	}
	return strings.TrimSpace(before + strings.TrimPrefix(after, "\n\n")), ctx
}

// MaxToolIO bounds a tool call's input and output in a SessionEvent, so
// a large file read does not flood a chat client.
const MaxToolIO = 8 << 10

// EventDecoder turns journal frames into session events. It keeps tool
// titles across frames (updates often omit them). Session
// subscribe and Desk's older-history paging (#535) share it, so both
// show a transcript the same way. Redact, when set, is applied to tool
// input and output.
type EventDecoder struct {
	tools  map[string]toolLabel
	Redact func(string) string
	// loading is set from a controller session/load until the agent's
	// reply: the agent replays the whole conversation as session/update
	// frames, which are history already in the journal, not new output.
	loading bool
}

type journalMsg struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Decode returns the event for one frame; ok is false for frames that
// show nothing.
func (e *EventDecoder) Decode(seq, t int64, from string, msg json.RawMessage) (SessionEvent, bool) {
	if e.tools == nil {
		e.tools = map[string]toolLabel{}
	}
	out := SessionEvent{Seq: seq}
	if t > 0 {
		at := time.UnixMilli(t).UTC()
		out.Time = &at
	}
	var m journalMsg
	if json.Unmarshal(msg, &m) != nil {
		return out, false
	}
	if from == "controller" && m.Method == "session/load" {
		e.loading = true
		return out, false
	}
	if e.loading {
		if from == "agent" && m.Method == "" && (len(m.Result) > 0 || m.Error != nil) {
			e.loading = false
		}
		return out, false
	}
	switch {
	case m.Method == "session/prompt":
		var p struct {
			Meta   map[string]json.RawMessage `json:"_meta"`
			Prompt []struct {
				Text     string `json:"text"`
				Type     string `json:"type"`
				Data     string `json:"data"`
				MimeType string `json:"mimeType"`
			} `json:"prompt"`
		}
		_ = json.Unmarshal(m.Params, &p)
		var parts []string
		for _, b := range p.Prompt {
			if b.Text != "" {
				parts = append(parts, b.Text)
			}
			if b.Type == "image" {
				out.Images = appendImage(out.Images, b.MimeType, b.Data)
			}
		}
		out.Type = EventTurnStart
		raw := strings.Join(parts, "\n")
		out.Text, out.Context = SplitContext(raw)
		var presentation PromptPresentation
		if from == "controller" && json.Unmarshal(p.Meta[MetaPresentation], &presentation) == nil && presentation.Version == 1 {
			out.Messages, out.Context, out.RawPrompt = presentation.Messages, presentation.Context, raw
			var visible []string
			for _, message := range presentation.Messages {
				visible = append(visible, message.Text)
			}
			out.Text = strings.Join(visible, "\n\n")
		}
		return out, true
	case m.Method == "session/update":
		return e.update(out, m.Params)
	case m.Method == "" && len(m.Result) > 0:
		// Only session/prompt results carry a stopReason.
		var res struct {
			StopReason string `json:"stopReason"`
		}
		if json.Unmarshal(m.Result, &res) == nil && res.StopReason != "" {
			out.Type, out.StopReason = EventTurnEnd, res.StopReason
			return out, true
		}
	case m.Method == "" && m.Error != nil && from == "agent":
		out.Type, out.Error = EventTurnEnd, m.Error.Message
		return out, true
	}
	return out, false
}

func (e *EventDecoder) update(out SessionEvent, params json.RawMessage) (SessionEvent, bool) {
	var u struct {
		Update struct {
			Kind       string          `json:"sessionUpdate"`
			Content    json.RawMessage `json:"content"`
			ToolCallID string          `json:"toolCallId"`
			Title      string          `json:"title"`
			Status     string          `json:"status"`
			RawInput   json.RawMessage `json:"rawInput"`
			RawOutput  json.RawMessage `json:"rawOutput"`
			Locations  []struct {
				Path string `json:"path"`
			} `json:"locations"`
			Used int `json:"used"`
			Size int `json:"size"`
		} `json:"update"`
	}
	if json.Unmarshal(params, &u) != nil {
		return out, false
	}
	up := u.Update
	switch up.Kind {
	case "agent_message_chunk", "agent_thought_chunk":
		var c struct {
			Type     string `json:"type"`
			Text     string `json:"text"`
			Data     string `json:"data"`
			MimeType string `json:"mimeType"`
		}
		_ = json.Unmarshal(up.Content, &c)
		if c.Type == "image" && up.Kind == "agent_message_chunk" {
			out.Images = appendImage(nil, c.MimeType, c.Data)
			out.Type = EventImage
			return out, len(out.Images) > 0
		}
		if (c.Type == "text" || c.Type == "") && c.Text != "" {
			out.Type, out.Text = EventText, c.Text
			if up.Kind == "agent_thought_chunk" {
				out.Type = EventReasoning
			}
			return out, true
		}
	case "usage_update":
		out.Type, out.Usage = EventUsage, &Usage{Used: up.Used, Size: up.Size}
		return out, true
	case "tool_call", "tool_call_update":
		l := e.tools[up.ToolCallID]
		if up.Title != "" {
			l.title = up.Title
		}
		var locs []string
		for _, x := range up.Locations {
			locs = append(locs, x.Path)
		}
		if t := ToolTarget(up.RawInput, locs); t != "" {
			l.target = e.redact(t)
		}
		if i := ToolIntent(up.RawInput); i != "" {
			l.intent = e.redact(i)
		}
		e.tools[up.ToolCallID] = l
		status := up.Status
		if status == "" {
			status = "pending"
		}
		out.Type = EventTool
		var blocks []struct {
			Content struct {
				Type     string `json:"type"`
				MimeType string `json:"mimeType"`
				Data     string `json:"data"`
			} `json:"content"`
		}
		_ = json.Unmarshal(up.Content, &blocks)
		for _, b := range blocks {
			if b.Content.Type == "image" {
				out.Images = appendImage(out.Images, b.Content.MimeType, b.Content.Data)
			}
		}
		// Text deltas stay bounded in tool summaries, but image deltas are content.
		if up.Status == "" && up.Kind == "tool_call_update" && len(out.Images) == 0 {
			return out, false
		}
		out.Tool = &ToolSummary{ID: up.ToolCallID, Title: l.title, Status: status, Input: e.io(up.RawInput), Output: e.io(up.RawOutput), Target: l.target, Intent: l.intent}
		if out.Tool.Output == "" {
			out.Tool.Output = e.io(toolText(up.Content))
		}
		return out, true
	}
	return out, false
}

func appendImage(images []SessionImage, mime, data string) []SessionImage {
	// Bound each event as well as each image before allocating decoded bytes.
	total := len(data)
	for _, img := range images {
		total += len(img.Data)
	}
	if total > base64.StdEncoding.EncodedLen(MaxAttachmentBytes) {
		return images
	}
	b, err := base64.StdEncoding.DecodeString(data)
	if err == nil && ValidImage(b, mime) {
		return append(images, SessionImage{MimeType: mime, Data: data})
	}
	return images
}

// ValidImage bounds raster image size and verifies the declared MIME type.
func ValidImage(data []byte, mime string) bool {
	if len(data) == 0 || len(data) > MaxAttachmentBytes || (mime != "image/png" && mime != "image/jpeg" && mime != "image/gif") {
		return false
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	return err == nil && http.DetectContentType(data) == mime && cfg.Width > 0 && cfg.Height > 0 && int64(cfg.Width)*int64(cfg.Height) <= 64_000_000
}

// toolText is the text of a tool call's content blocks, as JSON string.
func toolText(raw json.RawMessage) json.RawMessage {
	var blocks []struct {
		Content struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if json.Unmarshal(raw, &blocks) != nil {
		return nil
	}
	var b strings.Builder
	for _, x := range blocks {
		b.WriteString(x.Content.Text)
	}
	if b.Len() == 0 {
		return nil
	}
	s, _ := json.Marshal(b.String())
	return s
}

// toolLabel is what the decoder keeps of a tool call between frames:
// updates often omit the title and input.
type toolLabel struct{ title, target, intent string }

func (e *EventDecoder) redact(s string) string {
	if e.Redact != nil {
		return e.Redact(s)
	}
	return s
}

// io renders a raw tool value: a JSON string as its text, anything else
// as compact JSON, cut to MaxToolIO.
func (e *EventDecoder) io(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		s = string(raw)
	}
	if e.Redact != nil {
		s = e.Redact(s)
	}
	if len(s) > MaxToolIO {
		s = s[:MaxToolIO] + "\n[cut]"
	}
	return s
}
