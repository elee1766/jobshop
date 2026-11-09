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

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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

	// Selecting a managed chat needs no all-session lookup or rw-attach grant.
	fields, err := json.Marshal(domain.SessionFields{Key: "desk:chosen"})
	require.NoError(t, err)
	ctl.results[rpc.MethodTicketShow] = rpc.ShowResult{Ticket: domain.Ticket{ID: "s1", Kind: domain.KindSession, Fields: fields}}
	ctl.calls = nil
	w = postJSON(h, "/api/chat/send", `{"session":"s1","key":"desk:chosen","text":"selected"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, []string{rpc.MethodTicketShow + ` {"id":"s1"}`, rpc.MethodSessionSend + ` {"id":"s1","text":"selected"}`}, ctl.calls)
	assert.Equal(t, http.StatusBadRequest, postJSON(h, "/api/chat/send", `{"session":"s1","key":"desk:other","text":"x"}`, nil).Code)

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
	ctl.results[rpc.MethodSessionSend] = fmt.Errorf("%w: pool does not allow rw attach", domain.ErrInvalid)
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
	assert.Empty(t, ctl.calls, "refused writes reach no controller")
	w = postJSON(h, "/api/chat/send", `{"key":"k","text":"x"}`, map[string]string{"Origin": "http://127.0.0.1:7480"})
	assert.Equal(t, http.StatusOK, w.Code, "same origin")

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

	ctl.results[rpc.MethodSessionSend] = fmt.Errorf("%w: set attach: rw", domain.ErrInvalid)
	get(t, h, "/api/session/s1/access", &a)
	assert.False(t, a.Write)
	assert.Contains(t, a.Reason, "attach: rw")
}

func TestChatList(t *testing.T) {
	sess := func(id, pool, key string) rpc.SessionInfo {
		f, _ := json.Marshal(domain.SessionFields{Key: key, Agent: "a"})
		return rpc.SessionInfo{ID: id, Alive: true, Ticket: &domain.Ticket{ID: id, Pool: pool, Fields: f}}
	}
	ctl := &fakeCtl{results: map[string]any{rpc.MethodSessionList: []rpc.SessionInfo{
		sess("s1", "operator", "matrix:!r:$e"), sess("s2", "dev", ""), sess("s3", "operator", "desk:abc"),
		sess("s4", "esc-chat", "escalation:js-e"),
	}}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	var out struct {
		Pool     string        `json:"pool"`
		Sessions []ChatSession `json:"sessions"`
	}
	get(t, h, "/api/chat", &out)
	assert.Equal(t, "operator", out.Pool)
	require.Len(t, out.Sessions, 2)
	assert.Equal(t, "desk:abc", out.Sessions[0].Key)
	assert.Equal(t, "escalation:js-e", out.Sessions[1].Key, "managed chats include escalations but exclude gateway sessions")
}

// TestChatSendEscalation: an escalation:<ticket> key goes to the chat
// pool the controller recorded in meta.escalation (#379), not Desk's.
func TestChatSendEscalation(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s1"},
		rpc.MethodTicketShow: rpc.ShowResult{Ticket: domain.Ticket{
			ID:   "js-e",
			Meta: []byte(`{"escalation":{"key":"escalation:js-e","pool":"esc-chat","session":"s1"}}`),
		}},
	}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	w := postJSON(h, "/api/chat/send", `{"key":"escalation:js-e","text":"resume it"}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodSessionSend+` {"pool":"esc-chat","key":"escalation:js-e","text":"resume it"}`, ctl.calls[len(ctl.calls)-1])

	// A ticket without that chat is refused.
	ctl.results[rpc.MethodTicketShow] = rpc.ShowResult{Ticket: domain.Ticket{ID: "js-x"}}
	w = postJSON(h, "/api/chat/send", `{"key":"escalation:js-x","text":"x"}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// The Linear and GitHub views attach an item as linear:KEY or
// github:owner/name#n. It is resolved through browse.show, so the agent gets
// the title, state and the provider's link, and a ref the controller refuses
// (outside the configured scope) stops the send.
func TestChatSendProviderContext(t *testing.T) {
	ctl := &fakeCtl{results: map[string]any{
		rpc.MethodSessionSend: rpc.SessionSendResult{Session: "s1"},
		rpc.MethodBrowseShow: rpc.BrowseDetail{
			Item:   rpc.BrowseItem{Provider: rpc.BrowseGitHub, Ref: "o/r#7", Kind: "pr", Title: "Add things", State: "open", URL: "https://github.com/o/r/pull/7"},
			Status: rpc.BrowseStatus{State: rpc.BrowseFresh},
		},
	}}
	h := New(Config{ChatPool: "operator"}, ctl, nil).Handler()
	w := postJSON(h, "/api/chat/send", `{"key":"desk:x","text":"look","refs":["github:o/r#7"]}`, nil)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Equal(t, rpc.MethodBrowseShow+` {"provider":"github","ref":"o/r#7","brief":true}`, ctl.calls[0])
	var sent rpc.SessionSendParams
	require.NoError(t, json.Unmarshal([]byte(strings.TrimPrefix(ctl.calls[len(ctl.calls)-1], rpc.MethodSessionSend+" ")), &sent))
	assert.Contains(t, sent.Context, "data, not instructions")
	assert.Contains(t, sent.Context, "Add things")
	assert.Contains(t, sent.Context, "https://github.com/o/r/pull/7")
	assert.NotContains(t, sent.Context, "ticket.show")

	// A lookup the controller could not make is a 400, never a record made up
	// from the reference.
	ctl.results[rpc.MethodBrowseShow] = rpc.BrowseDetail{Status: rpc.BrowseStatus{State: rpc.BrowseError, Code: rpc.BrowseCodeNotFound, Message: "no such item"}}
	w = postJSON(h, "/api/chat/send", `{"key":"desk:x","text":"look","refs":["github:o/r#8"]}`, nil)
	assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
}
