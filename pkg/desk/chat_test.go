package desk

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func postJSON(h http.Handler, path, body string, hdr map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Host = "127.0.0.1:7480"
	req.Header.Set(CSRFHeader, "1")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestChatSend(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s1"}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()

	// By key: the gateway's key in the chat pool, so Matrix and Desk share it.
	w := postJSON(h, "/api/chat/send", `{"key":"matrix:!r:$root","text":"hi"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodSessionSend+` {"pool":"operator","key":"matrix:!r:$root","text":"hi"}`, ctl.calls[len(ctl.calls)-1])

	// By session: an rw attach, which the controller checks against the pool.
	w = postJSON(h, "/api/chat/send", `{"session":"s9","text":"stop","interrupt":true}`, nil)
	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, rpc.MethodSessionSend+` {"id":"s9","text":"stop","interrupt":true,"attach":true}`, ctl.calls[len(ctl.calls)-1])

	for name, body := range map[string]string{
		"empty":   `{"key":"k","text":"  "}`,
		"both":    `{"key":"k","session":"s","text":"x"}`,
		"neither": `{"text":"x"}`,
		"bad":     `{`,
		"long":    `{"key":"k","text":"` + strings.Repeat("a", maxChatText+1) + `"}`,
	} {
		assert.Equal(t, http.StatusBadRequest, postJSON(h, "/api/chat/send", body, nil).Code, name)
	}

	// A refused rw attach is a 400 with the controller's reason.
	ctl.results[rpc.MethodSessionSend] = fmt.Errorf("%w: pool does not allow rw attach", store.ErrInvalid)
	w = postJSON(h, "/api/chat/send", `{"session":"s9","text":"x"}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "rw attach")
}

func TestChatWritesGuarded(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s1"}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/chat/send", strings.NewReader(`{"key":"k","text":"x"}`))
	req.Host = "127.0.0.1:7480"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code, "no CSRF header")
	w = postJSON(h, "/api/chat/send", `{"key":"k","text":"x"}`, map[string]string{"Origin": "http://evil.example"})
	assert.Equal(t, http.StatusForbidden, w.Code, "cross origin")
	w = postJSON(h, "/api/chat/send", `{"key":"k","text":"x"}`, map[string]string{"Origin": "http://127.0.0.1:7480"})
	assert.Equal(t, http.StatusOK, w.Code, "same origin")
	assert.Len(t, ctl.calls, 1, "refused writes reach no controller")

	// No chat pool: key sends and new sessions are refused.
	h = New(Config{}, ctl, nil).Handler()
	assert.Equal(t, http.StatusBadRequest, postJSON(h, "/api/chat/send", `{"key":"k","text":"x"}`, nil).Code)
	assert.Equal(t, http.StatusBadRequest, postJSON(h, "/api/chat/new", `{"text":"x"}`, nil).Code)
}

func TestChatNew(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s2", Started: true}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	w := postJSON(h, "/api/chat/new", `{"text":"why is dev quarantined?"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var n ChatNew
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))
	assert.Equal(t, "s2", n.Session)
	assert.Regexp(t, `^desk:[a-z0-9]{12}$`, n.Key)
	var p rpc.SessionSendParams
	_, raw, _ := strings.Cut(ctl.calls[0], " ")
	require.NoError(t, json.Unmarshal([]byte(raw), &p))
	assert.Equal(t, rpc.SessionSendParams{Pool: "operator", Key: n.Key, Text: "why is dev quarantined?"}, p)

	w2 := postJSON(h, "/api/chat/new", `{"text":"again"}`, nil)
	var n2 ChatNew
	require.NoError(t, json.Unmarshal(w2.Body.Bytes(), &n2))
	assert.NotEqual(t, n.Key, n2.Key, "each click is a fresh session")
}

func TestChatAccess(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s1"}}}
	h := New(Config{}, ctl, nil).Handler()
	var a ChatAccess
	get(t, h, "/api/session/s1/access", &a)
	assert.True(t, a.Write)
	// The probe is an attach send with no text, which delivers nothing.
	assert.Equal(t, rpc.MethodSessionSend+` {"id":"s1","text":"","attach":true}`, ctl.calls[0])

	ctl.results[rpc.MethodSessionSend] = fmt.Errorf("%w: set attach: rw", store.ErrInvalid)
	get(t, h, "/api/session/s1/access", &a)
	assert.False(t, a.Write)
	assert.Contains(t, a.Reason, "attach: rw")
}

func TestChatList(t *testing.T) {
	sess := func(id, pool, key string) rpc.SessionInfo {
		f, _ := json.Marshal(store.SessionFields{Key: key, Agent: "a"})
		return rpc.SessionInfo{ID: id, Alive: true, Ticket: &store.Ticket{ID: id, Pool: pool, Fields: f}}
	}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionList: []rpc.SessionInfo{
		sess("s1", "operator", "matrix:!r:$e"), sess("s2", "dev", ""), sess("s3", "operator", "desk:abc"), sess("s4", "operator", "escalation:js-9"),
	}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	var out struct {
		Pool     string        `json:"pool"`
		Sessions []ChatSession `json:"sessions"`
	}
	get(t, h, "/api/chat", &out)
	assert.Equal(t, "operator", out.Pool)
	require.Len(t, out.Sessions, 3)
	assert.Equal(t, "matrix:!r:$e", out.Sessions[0].Key)
	assert.Equal(t, "desk:abc", out.Sessions[1].Key)
	assert.Equal(t, "Escalation js-9", out.Sessions[2].Title, "escalation keys are labeled by ticket (#379)")
}

// TestEscalationChatSend (#379): the ticket view carries meta.escalation,
// and /api/chat/send with its key reaches the controller's session in the
// chat pool; the attention item names the key.
func TestEscalationChatSend(t *testing.T) {
	esc := store.Ticket{
		ID: "js-9", Kind: store.KindTask, Labels: []string{NeedsHuman},
		Meta: json.RawMessage(`{"escalation":{"key":"escalation:js-9","session":"s7","pool":"operator"}}`),
	}
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodTicketShow:       rpc.ShowResult{Ticket: esc},
		rpc.MethodTicketList:       []store.Ticket{esc},
		rpc.MethodSessionSend:      rpc.SessionSendResult{Session: "s7"},
		rpc.MethodRunList:          []store.Ticket{},
		rpc.MethodPoolList:         rpc.PoolStatusResult{},
		rpc.MethodTicketList + ":": []store.Ticket{esc},
	}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()

	var tv TicketView
	get(t, h, "/api/ticket/js-9", &tv)
	var m struct {
		Escalation struct{ Key, Session string } `json:"escalation"`
	}
	require.NoError(t, json.Unmarshal(tv.Ticket.Meta, &m))
	assert.Equal(t, "escalation:js-9", m.Escalation.Key)
	assert.Equal(t, "s7", m.Escalation.Session)

	w := postJSON(h, "/api/chat/send", `{"key":"`+m.Escalation.Key+`","text":"resume it"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodSessionSend+` {"pool":"operator","key":"escalation:js-9","text":"resume it"}`, ctl.calls[len(ctl.calls)-1])

	var items []AttentionItem
	get(t, h, "/api/attention", &items)
	require.NotEmpty(t, items)
	assert.Equal(t, "escalation:js-9", items[0].Chat)
}
