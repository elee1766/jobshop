package desk

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

func calls(ctl *fakeCtl, method string) []string {
	var out []string
	for _, c := range ctl.calls {
		if m, body, _ := strings.Cut(c, " "); m == method {
			out = append(out, body)
		}
	}
	return out
}

// TestAttentionChatAttaches (#606): a click asks the controller for the
// item's chat and only attaches: no session.send, no ticket write, so no
// new turn starts.
func TestAttentionChatAttaches(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodAttentionChat: rpc.AttentionChatResult{Kind: "stale", ID: "js-9", Key: "attention:stale:js-9", Session: "s7", Pool: "operator"},
	}}
	h := New(Config{ChatPool: "operator", Root: t.TempDir()}, ctl, nil).Handler()
	w := postJSON(h, "/api/attention/chat", `{"kind":"stale","id":"js-9","title":"orphan task"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var c AttentionChat
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	assert.Equal(t, AttentionChat{Kind: "stale", ID: "js-9", Key: "attention:stale:js-9", Session: "s7", Pool: "operator"}, c)
	chats := calls(ctl, rpc.MethodAttentionChat)
	require.Len(t, chats, 1)
	var p rpc.AttentionChatParams
	require.NoError(t, json.Unmarshal([]byte(chats[0]), &p))
	assert.Equal(t, rpc.AttentionChatParams{Kind: "stale", ID: "js-9", Title: "orphan task"}, p)
	assert.Empty(t, calls(ctl, rpc.MethodSessionSend))
	assert.Empty(t, calls(ctl, rpc.MethodTicketUpdate))

	w = postJSON(h, "/api/attention/chat", `{"kind":"stale","id":"a b"}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestAttentionAck: ack hides an item until it leaves the list, in a
// 0600 desk/attention.json, and closes the item's chat.
func TestAttentionAck(t *testing.T) {
	root := t.TempDir()
	sess := domain.Ticket{ID: "s3", Kind: domain.KindSession, Status: domain.StatusOpen, Fields: json.RawMessage(`{"key":"attention:config:config"}`)}
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodSessionList: []rpc.SessionInfo{{ID: "s3", Ticket: &sess}},
		rpc.MethodTicketClose: domain.Ticket{ID: "s3"},
		rpc.MethodAttention: rpc.AttentionResult{Items: []rpc.AttentionItem{
			{Kind: rpc.AttentionConfig, ID: "config", Title: "pack failed", Since: time.Now()},
			{Kind: rpc.AttentionQuestion, ID: "js-q", Title: "which branch?", Ref: "#4", Since: time.Now()},
		}},
	}}
	h := New(Config{ChatPool: "operator", Root: root}, ctl, nil).Handler()

	var items []AttentionItem
	get(t, h, "/api/attention", &items)
	require.Len(t, items, 2)
	assert.Equal(t, "question", items[1].Kind)
	assert.Equal(t, "#4", items[1].Ref)

	w := postJSON(h, "/api/attention/ack", `{"kind":"config","id":"config"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	fi, err := os.Stat(filepath.Join(root, "desk", "attention.json"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
	get(t, h, "/api/attention", &items)
	require.Len(t, items, 1)
	assert.Equal(t, "js-q", items[0].ID)
	closes := calls(ctl, rpc.MethodTicketClose)
	require.Len(t, closes, 1)
	assert.Contains(t, closes[0], `"id":"s3"`)
}

// TestSessionEvents pages the journal before a seq and decodes it the
// way session.subscribe does.
func TestSessionEvents(t *testing.T) {
	frames := []rpc.JournalFrame{
		{Seq: 5, T: 1000, From: "controller", Msg: json.RawMessage(`{"method":"session/prompt","params":{"prompt":[{"type":"text","text":"hi"}]}}`)},
		{Seq: 6, T: 1001, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"agent_thought_chunk","content":{"type":"text","text":"hmm"}}}}`)},
		{Seq: 7, T: 1002, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"t1","title":"Read","status":"in_progress","rawInput":{"intent":"Look at a","path":"a.go"}}}}`)},
		{Seq: 8, T: 1003, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call_update","toolCallId":"t1","status":"completed","rawOutput":"package a"}}}`)},
		{Seq: 9, T: 1004, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"usage_update","used":1200,"size":200000}}}`)},
		// A reconnect resumes with session/load: the agent replays the
		// conversation, which is history and shows nothing.
		{Seq: 10, T: 1005, From: "controller", Msg: json.RawMessage(`{"id":2,"method":"session/load","params":{}}`)},
		{Seq: 11, T: 1005, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"old reply"}}}}`)},
		{Seq: 12, T: 1005, From: "agent", Msg: json.RawMessage(`{"id":2,"result":{}}`)},
	}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionJournal: rpc.SessionJournalResult{Frames: frames}}}
	h := New(Config{}, ctl, nil).Handler()
	var got SessionEvents
	get(t, h, "/api/session/s1/events?before=13&limit=9", &got)
	assert.Equal(t, int64(4), got.First)
	assert.Equal(t, rpc.MethodSessionJournal+` {"id":"s1","since_seq":3,"limit":9}`, ctl.calls[0])
	var types []string
	for _, e := range got.Events {
		types = append(types, e.Type)
	}
	assert.Equal(t, []string{"turn_start", "reasoning", "tool", "tool", "usage"}, types)
	assert.Equal(t, `{"intent":"Look at a","path":"a.go"}`, got.Events[2].Tool.Input)
	assert.Equal(t, "Read", got.Events[3].Tool.Title, "title carried from the start frame")
	assert.Equal(t, "a.go", got.Events[3].Tool.Target, "target carried from the start frame")
	assert.Equal(t, "Look at a", got.Events[3].Tool.Intent)
	assert.Equal(t, "package a", got.Events[3].Tool.Output)
	assert.Equal(t, &rpc.Usage{Used: 1200, Size: 200000}, got.Events[4].Usage)

	w := get(t, h, "/api/session/s1/events?before=1", nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestChatCancel(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionCancel: rpc.SessionCancelResult{ID: "s1", Cancelled: true}}}
	h := New(Config{}, ctl, nil).Handler()
	w := postJSON(h, "/api/chat/cancel", `{"session":"s1"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.JSONEq(t, `{"id":"s1","cancelled":true}`, w.Body.String())
	assert.Equal(t, rpc.MethodSessionCancel+` {"id":"s1"}`, ctl.calls[0])
}
