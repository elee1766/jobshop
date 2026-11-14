package session

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	acp "github.com/coder/acp-go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	host "github.com/gfx-labs/acphost/pkg/acphost"
)

// TestRunningTool: the latest open tool call names the running tool; a
// completed or failed call drops out, and an update without a title or
// input keeps what it had (JOB-54).
func TestRunningTool(t *testing.T) {
	var tr Traffic
	assert.Zero(t, tr.Running())
	tr.Track("a", OpenTool{Title: "Read file"}, acp.ToolCallStatusPending)
	tr.Track("b", OpenTool{Title: "Running shell command"}, acp.ToolCallStatusPending)
	tr.Track("b", OpenTool{Target: "go test ./...", Intent: "Run tests"}, acp.ToolCallStatusInProgress)
	assert.Equal(t, OpenTool{"Running shell command", "go test ./...", "Run tests"}, tr.Running())
	tr.Track("b", OpenTool{}, acp.ToolCallStatusCompleted)
	assert.Equal(t, "Read file", tr.Running().Title)
	tr.Track("a", OpenTool{}, acp.ToolCallStatusFailed)
	assert.Zero(t, tr.Running())
}

// TestTrafficSaid keeps only the newest text once it passes the limit, and
// Reset clears text and tools for the next turn.
func TestTrafficSaid(t *testing.T) {
	var tr Traffic
	c := &Client{Traffic: &tr}
	chunk := func(s string) acp.SessionNotification {
		return acp.SessionNotification{Update: acp.SessionUpdate{AgentMessageChunk: &acp.SessionUpdateAgentMessageChunk{
			Content: acp.TextBlock(s),
		}}}
	}
	require.NoError(t, c.SessionUpdate(context.Background(), chunk("hello ")))
	require.NoError(t, c.SessionUpdate(context.Background(), chunk("world")))
	assert.Equal(t, "hello world", tr.Said())

	require.NoError(t, c.SessionUpdate(context.Background(), chunk(strings.Repeat("x", saidMax))))
	require.NoError(t, c.SessionUpdate(context.Background(), chunk("tail")))
	got := tr.Said()
	assert.LessOrEqual(t, len(got), saidMax)
	assert.True(t, strings.HasSuffix(got, "tail"))

	tr.Track("a", OpenTool{Title: "x"}, acp.ToolCallStatusPending)
	tr.Reset()
	assert.Empty(t, tr.Said())
	assert.Zero(t, tr.Running())
}

func TestClientPermission(t *testing.T) {
	req := acp.RequestPermissionRequest{Options: []acp.PermissionOption{
		{OptionId: "once", Kind: acp.PermissionOptionKindAllowOnce},
		{OptionId: "always", Kind: acp.PermissionOptionKindAllowAlways},
		{OptionId: "no", Kind: acp.PermissionOptionKindRejectOnce},
	}}
	pick := func(c *Client) string {
		resp, err := c.RequestPermission(context.Background(), req)
		require.NoError(t, err)
		require.NotNil(t, resp.Outcome.Selected)
		return string(resp.Outcome.Selected.OptionId)
	}
	assert.Equal(t, "always", pick(&Client{}))
	assert.Equal(t, "no", pick(&Client{Refuse: func(*acp.ToolKind) bool { return true }}))
	assert.Equal(t, "always", pick(&Client{Refuse: func(*acp.ToolKind) bool { return false }}))
}

// TestClientTurnEnd: the signal never blocks, a second one while the first
// is unread is dropped, and other extension methods are refused.
func TestClientTurnEnd(t *testing.T) {
	ch := make(chan struct{}, 1)
	c := &Client{TurnEnd: ch}
	for range 2 {
		_, err := c.HandleExtensionMethod(context.Background(), host.MethodTurnEnd, nil)
		require.NoError(t, err)
	}
	assert.Len(t, ch, 1)
	_, err := c.HandleExtensionMethod(context.Background(), host.MethodExportStats, nil)
	require.NoError(t, err)
	_, err = c.HandleExtensionMethod(context.Background(), "nope", nil)
	require.Error(t, err)
	_, err = c.ReadTextFile(context.Background(), acp.ReadTextFileRequest{})
	require.Error(t, err)
}

func TestPromptErrors(t *testing.T) {
	req := func(code int, msg string) error { return &acp.RequestError{Code: code, Message: msg} }
	assert.True(t, UnknownSession(req(-32602, "Unknown ACP session id")))
	assert.False(t, UnknownSession(req(-32000, "Unknown ACP session id")))
	assert.False(t, UnknownSession(assert.AnError))
	// towa's shape (JOB-214): -32002 counts only with data.sessionId.
	notFound := func(data any) error {
		return &acp.RequestError{Code: -32002, Message: "Resource not found", Data: data}
	}
	assert.True(t, UnknownSession(notFound(map[string]any{"sessionId": "x"})))
	assert.False(t, UnknownSession(notFound(nil)))
	assert.False(t, UnknownSession(notFound(map[string]any{"path": "x"})))
	assert.False(t, UnknownSession(notFound("sessionId")))
	assert.True(t, RateLimited(req(-32000, "HTTP 429 Too Many Requests")))
	assert.True(t, RateLimited(req(-32000, "model is unavailable")))
	assert.False(t, RateLimited(assert.AnError))
	assert.True(t, ProviderUnreachable(req(-32000, "Failed to send request: Connect")))
	assert.False(t, ProviderUnreachable(req(-32000, "boom")))
	// towa's typed stop (JOB-215): the kind decides, not the text or code.
	stop := func(code int, kind, msg string) error {
		return &acp.RequestError{Code: code, Message: "Internal error", Data: map[string]any{
			"error": "turn stopped: " + kind + ": " + msg,
			"towa":  map[string]any{"stop": map[string]any{"kind": kind, "account": "a-ant", "message": msg}},
		}}
	}
	network := stop(-32603, "network", "send request")
	assert.True(t, ProviderUnreachable(network))
	assert.False(t, RateLimited(network))
	// The payload seen 2026-11-13 (dead proxy), decoded as the client does.
	var seen acp.RequestError
	require.NoError(t, json.Unmarshal([]byte(`{"code":-32603,"message":"Internal error","data":{"error":"turn stopped: network: provider: network (retry): send request: at claude:a-ant/claude-sonnet-5-5: Post \"https://api.anthropic.com/v1/messages?beta=true\": proxyconnect tcp: dial tcp 127.0.0.1:1: connect: connection refused","towa":{"stop":{"kind":"network","account":"a-ant","model":"claude-sonnet-5-5","message":"send request"}}}}`), &seen))
	assert.True(t, ProviderUnreachable(&seen))
	assert.False(t, RateLimited(&seen))
	for _, kind := range []string{"rate_limit", "overload"} {
		// For rate_limit the kind is also in the printed data, so only
		// overload proves the kind alone decides.
		err := stop(-32603, kind, "chain exhausted")
		assert.True(t, RateLimited(err), kind)
		assert.False(t, ProviderUnreachable(err), kind)
	}
	assert.True(t, RateLimited(stop(-32000, "overload", "x")), "code is not part of the check")
	for _, kind := range []string{"request", "auth", "context_full", "limit", "denied"} {
		err := stop(-32603, kind, "x")
		assert.False(t, RateLimited(err), kind)
		assert.False(t, ProviderUnreachable(err), kind)
	}
	// Malformed data never matches: wrong types at each level.
	for _, data := range []any{
		"network",
		map[string]any{"towa": "network"},
		map[string]any{"towa": map[string]any{"stop": "network"}},
		map[string]any{"towa": map[string]any{"stop": map[string]any{"kind": 1}}},
	} {
		err := &acp.RequestError{Code: -32603, Message: "Internal error", Data: data}
		assert.False(t, ProviderUnreachable(err), "%v", data)
	}
	assert.True(t, AgentBusy(req(-32000, "Already processing a prompt")))
	assert.False(t, AgentBusy(assert.AnError))
}
