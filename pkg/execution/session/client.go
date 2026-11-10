package session

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"sync"

	acp "github.com/coder/acp-go-sdk"

	host "github.com/gfx-labs/acphost/pkg/acphost"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// OpenTool is a running tool call's title and what it acts on.
type OpenTool struct{ Title, Target, Intent string }

// Traffic is what the agent's callbacks have said during the current turn:
// its open tool calls and the tail of its text. It has its own lock, so
// ACP callbacks never wait on the runner's. The zero value is ready.
type Traffic struct {
	mu        sync.Mutex
	tools     map[acp.ToolCallId]OpenTool
	toolOrder []acp.ToolCallId
	said      strings.Builder
}

// saidMax is the text size at which Said keeps only its newest saidKeep bytes.
const (
	saidMax  = 4096
	saidKeep = 1024
)

// Reset starts a turn: no open tools, no text.
func (t *Traffic) Reset() {
	t.mu.Lock()
	t.said.Reset()
	t.tools, t.toolOrder = nil, nil
	t.mu.Unlock()
}

// Said is the tail of the agent's text since Reset.
func (t *Traffic) Said() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.said.String()
}

func (t *Traffic) say(s string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.said.WriteString(s)
	if all := t.said.String(); len(all) > saidMax { // only the tail is reported
		t.said.Reset()
		t.said.WriteString(all[len(all)-saidKeep:])
	}
}

// Track records a tool call's label and drops it once it completes or
// fails, so Running names the latest open call (JOB-54). Empty fields
// keep what an earlier frame said.
func (t *Traffic) Track(id acp.ToolCallId, o OpenTool, st acp.ToolCallStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if st == acp.ToolCallStatusCompleted || st == acp.ToolCallStatusFailed {
		delete(t.tools, id)
		t.toolOrder = slices.DeleteFunc(t.toolOrder, func(x acp.ToolCallId) bool { return x == id })
		return
	}
	if t.tools == nil {
		t.tools = map[acp.ToolCallId]OpenTool{}
	}
	old, ok := t.tools[id]
	if !ok {
		t.toolOrder = append(t.toolOrder, id)
	}
	if o.Title == "" {
		o.Title = old.Title
	}
	if o.Target == "" {
		o.Target = old.Target
	}
	if o.Intent == "" {
		o.Intent = old.Intent
	}
	t.tools[id] = o
}

// Running is the latest open tool call, or zero.
func (t *Traffic) Running() OpenTool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if n := len(t.toolOrder); n > 0 {
		return t.tools[t.toolOrder[n-1]]
	}
	return OpenTool{}
}

// Client is the ACP client side the agent talks to. It auto-approves
// permission prompts: agents run unattended in their own slice and
// worktree. It holds no ticket, pool or store state.
type Client struct {
	// Traffic receives tool and text updates; nil drops them.
	Traffic *Traffic
	// TurnEnd is signalled without blocking when the host reports a turn
	// ended, including one that ended while no controller was attached.
	TurnEnd chan<- struct{}
	// Refuse reports whether a tool call of this kind is denied (a
	// read-only chat, #606); nil allows every call.
	Refuse func(kind *acp.ToolKind) bool
}

// RequestPermission answers a permission prompt with the first allow option, or a reject when Refuse denies the kind.
func (c *Client) RequestPermission(_ context.Context, p acp.RequestPermissionRequest) (acp.RequestPermissionResponse, error) {
	kinds := []acp.PermissionOptionKind{acp.PermissionOptionKindAllowAlways, acp.PermissionOptionKindAllowOnce}
	if c.Refuse != nil && c.Refuse(p.ToolCall.Kind) {
		// Once only: the same tool is allowed after the chat is writable.
		kinds = []acp.PermissionOptionKind{acp.PermissionOptionKindRejectOnce}
	}
	for _, kind := range kinds {
		for _, o := range p.Options {
			if o.Kind == kind {
				return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeSelected(o.OptionId)}, nil
			}
		}
	}
	return acp.RequestPermissionResponse{Outcome: acp.NewRequestPermissionOutcomeCancelled()}, nil
}

// toolInput is rawInput as JSON for rpc.ToolTarget; nil when absent.
func toolInput(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}

func toolPaths(l []acp.ToolCallLocation) []string {
	var out []string
	for _, x := range l {
		out = append(out, x.Path)
	}
	return out
}

// SessionUpdate records tool calls and agent text in Traffic.
func (c *Client) SessionUpdate(_ context.Context, n acp.SessionNotification) error {
	if c.Traffic == nil {
		return nil
	}
	if tc := n.Update.ToolCall; tc != nil {
		in := toolInput(tc.RawInput)
		c.Traffic.Track(tc.ToolCallId, OpenTool{tc.Title, rpc.ToolTarget(in, toolPaths(tc.Locations)), rpc.ToolIntent(in)}, tc.Status)
	}
	if tu := n.Update.ToolCallUpdate; tu != nil {
		title := ""
		if tu.Title != nil {
			title = *tu.Title
		}
		var st acp.ToolCallStatus
		if tu.Status != nil {
			st = *tu.Status
		}
		in := toolInput(tu.RawInput)
		c.Traffic.Track(tu.ToolCallId, OpenTool{title, rpc.ToolTarget(in, toolPaths(tu.Locations)), rpc.ToolIntent(in)}, st)
	}
	if m := n.Update.AgentMessageChunk; m != nil && m.Content.Text != nil {
		c.Traffic.say(m.Content.Text.Text)
	}
	return nil
}

// HandleExtensionMethod signals TurnEnd on the host's turn-end frame and ignores the retired stats export.
func (c *Client) HandleExtensionMethod(_ context.Context, method string, _ json.RawMessage) (any, error) {
	if method == host.MethodExportStats {
		// Hosts no longer export (#530); an older host may still report.
		return nil, nil
	}
	if method == host.MethodTurnEnd {
		select {
		case c.TurnEnd <- struct{}{}:
		default:
		}
		return nil, nil
	}
	return nil, acp.NewMethodNotFound(method)
}

var errUnsupported = acp.NewMethodNotFound("not supported by jobshop")

// ReadTextFile is refused: the agent reads files itself.
func (c *Client) ReadTextFile(context.Context, acp.ReadTextFileRequest) (acp.ReadTextFileResponse, error) {
	return acp.ReadTextFileResponse{}, errUnsupported
}

// WriteTextFile is refused: the agent writes files itself.
func (c *Client) WriteTextFile(context.Context, acp.WriteTextFileRequest) (acp.WriteTextFileResponse, error) {
	return acp.WriteTextFileResponse{}, errUnsupported
}

// CreateTerminal is refused: the agent runs commands itself.
func (c *Client) CreateTerminal(context.Context, acp.CreateTerminalRequest) (acp.CreateTerminalResponse, error) {
	return acp.CreateTerminalResponse{}, errUnsupported
}

// KillTerminal is refused: terminals are not offered.
func (c *Client) KillTerminal(context.Context, acp.KillTerminalRequest) (acp.KillTerminalResponse, error) {
	return acp.KillTerminalResponse{}, errUnsupported
}

// TerminalOutput is refused: terminals are not offered.
func (c *Client) TerminalOutput(context.Context, acp.TerminalOutputRequest) (acp.TerminalOutputResponse, error) {
	return acp.TerminalOutputResponse{}, errUnsupported
}

// ReleaseTerminal is refused: terminals are not offered.
func (c *Client) ReleaseTerminal(context.Context, acp.ReleaseTerminalRequest) (acp.ReleaseTerminalResponse, error) {
	return acp.ReleaseTerminalResponse{}, errUnsupported
}

// WaitForTerminalExit is refused: terminals are not offered.
func (c *Client) WaitForTerminalExit(context.Context, acp.WaitForTerminalExitRequest) (acp.WaitForTerminalExitResponse, error) {
	return acp.WaitForTerminalExitResponse{}, errUnsupported
}
