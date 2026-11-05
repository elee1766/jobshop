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

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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

// TestAttentionChatStale: a stale ticket has no chat, so the first click
// starts one keyed attention:stale:<id> with a context prompt and records
// the key on the ticket; the second click finds the open session.
func TestAttentionChatStale(t *testing.T) {
	tk := store.Ticket{ID: "js-9", Title: "orphan task", Revision: 4, Meta: json.RawMessage(`{"github":{"repo":"o/r","issue":12}}`)}
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketShow:  rpc.ShowResult{Ticket: tk},
		rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s7", Started: true},
	}}
	h := New(Config{ChatPool: "operator", Root: t.TempDir()}, ctl, nil).Handler()
	w := postJSON(h, "/api/attention/chat", `{"kind":"stale","id":"js-9","title":"orphan task"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var c AttentionChat
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	assert.Equal(t, AttentionChat{Kind: "stale", ID: "js-9", Key: "attention:stale:js-9", Session: "s7", Pool: "operator", Started: true}, c)

	sends := calls(ctl, rpc.MethodSessionSend)
	require.Len(t, sends, 1)
	var p rpc.SessionSendParams
	require.NoError(t, json.Unmarshal([]byte(sends[0]), &p))
	assert.Equal(t, "operator", p.Pool)
	assert.Equal(t, "attention:stale:js-9", p.Key)
	assert.Contains(t, p.Text, "Desk attention item: stale js-9")
	assert.Contains(t, p.Text, "https://github.com/o/r/issues/12")

	ups := calls(ctl, rpc.MethodTicketUpdate)
	require.Len(t, ups, 1)
	var u rpc.UpdateParams
	require.NoError(t, json.Unmarshal([]byte(ups[0]), &u))
	assert.Equal(t, int64(4), u.IfRevision, "meta is replaced, so the write is conditional")
	assert.JSONEq(t, `{"github":{"repo":"o/r","issue":12},"desk":{"attention_chat":"attention:stale:js-9"}}`, string(u.Meta))

	// Second click: the session for the key is open, nothing is sent.
	tk.Meta = u.Meta
	ctl.results[rpc.MethodTicketShow] = rpc.ShowResult{Ticket: tk}
	sess := store.Ticket{ID: "s7", Kind: store.KindSession, Status: store.StatusOpen, Fields: json.RawMessage(`{"key":"attention:stale:js-9"}`)}
	ctl.results[rpc.MethodSessionList] = []rpc.SessionInfo{{ID: "s7", Ticket: &sess}}
	w = postJSON(h, "/api/attention/chat", `{"kind":"stale","id":"js-9"}`, nil)
	c = AttentionChat{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	assert.Equal(t, "s7", c.Session)
	assert.False(t, c.Started)
	assert.Len(t, calls(ctl, rpc.MethodSessionSend), 1)
}

// TestAttentionChatEscalation: a ticket with the controller's escalation
// chat opens that chat and starts nothing.
func TestAttentionChatEscalation(t *testing.T) {
	tk := store.Ticket{ID: "js-e1", Meta: json.RawMessage(`{"escalation":{"key":"escalation:js-e1","session":"js-s1","pool":"ops"}}`)}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodTicketShow: rpc.ShowResult{Ticket: tk}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	w := postJSON(h, "/api/attention/chat", `{"kind":"ticket","id":"js-e1"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var c AttentionChat
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &c))
	assert.Equal(t, AttentionChat{Kind: "ticket", ID: "js-e1", Key: "escalation:js-e1", Session: "js-s1", Pool: "ops"}, c)
	assert.Empty(t, calls(ctl, rpc.MethodSessionSend))
}

// TestAttentionChatNotTicket: a config item keeps its key in
// desk/attention.json, and ack hides it until it leaves the list.
func TestAttentionChatNotTicket(t *testing.T) {
	root := t.TempDir()
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s3"},
		rpc.MethodAttention: rpc.AttentionResult{Items: []rpc.AttentionItem{
			{Kind: rpc.AttentionConfig, ID: "config", Title: "pack failed", Since: time.Now()},
			{Kind: rpc.AttentionQuestion, ID: "js-q", Title: "which branch?", Ref: "#4", Since: time.Now()},
		}},
	}}
	h := New(Config{ChatPool: "operator", Root: root}, ctl, nil).Handler()
	w := postJSON(h, "/api/attention/chat", `{"kind":"config","id":"config"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	b, err := os.ReadFile(filepath.Join(root, "desk", "attention.json"))
	require.NoError(t, err)
	assert.Contains(t, string(b), `"config:config": "attention:config:config"`)
	fi, err := os.Stat(filepath.Join(root, "desk", "attention.json"))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fi.Mode().Perm())

	var items []AttentionItem
	get(t, h, "/api/attention", &items)
	require.Len(t, items, 2)
	assert.Equal(t, "question", items[1].Kind)
	assert.Equal(t, "#4", items[1].Ref)

	w = postJSON(h, "/api/attention/ack", `{"kind":"config","id":"config"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	get(t, h, "/api/attention", &items)
	require.Len(t, items, 1)
	assert.Equal(t, "js-q", items[0].ID)
}

// TestSessionEvents pages the journal before a seq and decodes it the
// way session.subscribe does.
func TestSessionEvents(t *testing.T) {
	frames := []rpc.JournalFrame{
		{Seq: 5, T: 1000, From: "controller", Msg: json.RawMessage(`{"method":"session/prompt","params":{"prompt":[{"type":"text","text":"hi"}]}}`)},
		{Seq: 6, T: 1001, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"agent_thought_chunk","content":{"type":"text","text":"hmm"}}}}`)},
		{Seq: 7, T: 1002, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call","toolCallId":"t1","title":"Read","status":"in_progress","rawInput":{"path":"a.go"}}}}`)},
		{Seq: 8, T: 1003, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"tool_call_update","toolCallId":"t1","status":"completed","rawOutput":"package a"}}}`)},
		{Seq: 9, T: 1004, From: "agent", Msg: json.RawMessage(`{"method":"session/update","params":{"update":{"sessionUpdate":"usage_update","used":1200,"size":200000}}}`)},
	}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionJournal: rpc.SessionJournalResult{Frames: frames}}}
	h := New(Config{}, ctl, nil).Handler()
	var got SessionEvents
	get(t, h, "/api/session/s1/events?before=10&limit=6", &got)
	assert.Equal(t, int64(4), got.First)
	assert.Equal(t, rpc.MethodSessionJournal+` {"id":"s1","since_seq":3,"limit":6}`, ctl.calls[0])
	var types []string
	for _, e := range got.Events {
		types = append(types, e.Type)
	}
	assert.Equal(t, []string{"turn_start", "reasoning", "tool", "tool", "usage"}, types)
	assert.Equal(t, `{"path":"a.go"}`, got.Events[2].Tool.Input)
	assert.Equal(t, "Read", got.Events[3].Tool.Title, "title carried from the start frame")
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
