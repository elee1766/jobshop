package gateway

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/clock"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const a2aToken = "s3cret-a2a-token"

// a2aPeer is an httptest A2A peer. mode picks how it answers SendMessage:
// message, task (working until GetTask polls done), input, rejected.
type a2aPeer struct {
	t    *testing.T
	mode string
	srv  *httptest.Server

	mu    sync.Mutex
	sends []a2aMessage
	gets  int
	// workPolls is how many GetTask calls still answer working.
	workPolls int
	// calls is closed and replaced on every request, so tests wait on it.
	calls chan struct{}
}

func newA2APeer(t *testing.T, mode string) *a2aPeer {
	p := &a2aPeer{t: t, mode: mode, workPolls: 1, calls: make(chan struct{})}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/agent-card.json", func(w http.ResponseWriter, r *http.Request) {
		if !p.authed(w, r) {
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name": "peer",
			"supportedInterfaces": []map[string]string{
				{"url": p.srv.URL + "/grpc", "protocolBinding": "GRPC", "protocolVersion": "1.0"},
				{"url": p.srv.URL + "/rpc", "protocolBinding": "JSONRPC", "protocolVersion": "1.0"},
			},
		})
	})
	mux.HandleFunc("POST /rpc", func(w http.ResponseWriter, r *http.Request) {
		if !p.authed(w, r) {
			return
		}
		b, _ := io.ReadAll(r.Body)
		var req struct {
			ID     int64           `json:"id"`
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}
		require.NoError(t, json.Unmarshal(b, &req))
		res := p.answer(req.Method, req.Params)
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": res})
		p.mu.Lock()
		close(p.calls)
		p.calls = make(chan struct{})
		p.mu.Unlock()
	})
	p.srv = httptest.NewServer(mux)
	t.Cleanup(p.srv.Close)
	return p
}

func (p *a2aPeer) authed(w http.ResponseWriter, r *http.Request) bool {
	if p.mode == "auth" || r.Header.Get("Authorization") != "Bearer "+a2aToken {
		http.Error(w, "no", http.StatusUnauthorized)
		p.mu.Lock()
		close(p.calls)
		p.calls = make(chan struct{})
		p.mu.Unlock()
		return false
	}
	return true
}

func task(id, ctxID, state, text string) map[string]any {
	st := map[string]any{"state": state}
	if text != "" {
		st["message"] = map[string]any{"messageId": "m", "role": "ROLE_AGENT", "parts": []map[string]string{{"text": text}}}
	}
	return map[string]any{"task": map[string]any{"id": id, "contextId": ctxID, "status": st}}
}

func (p *a2aPeer) answer(method string, raw json.RawMessage) any {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch method {
	case "SendMessage":
		var params struct {
			Message a2aMessage `json:"message"`
		}
		_ = json.Unmarshal(raw, &params)
		p.sends = append(p.sends, params.Message)
		m := params.Message
		switch p.mode {
		case "message":
			return map[string]any{"message": map[string]any{"messageId": "r", "role": "ROLE_AGENT", "contextId": m.ContextID, "parts": []map[string]string{{"text": "close it, drill acknowledged"}}}}
		case "input":
			if m.TaskID == "" {
				return task("t-1", m.ContextID, "TASK_STATE_INPUT_REQUIRED", "which run?")
			}
			return task("t-1", m.ContextID, "TASK_STATE_COMPLETED", "thanks, resolved")
		case "rejected":
			return task("t-1", m.ContextID, "TASK_STATE_REJECTED", "not my job")
		}
		return task("t-1", m.ContextID, "TASK_STATE_WORKING", "")
	case "GetTask":
		p.gets++
		if p.workPolls > 0 {
			p.workPolls--
			return task("t-1", "", "TASK_STATE_WORKING", "")
		}
		return task("t-1", "", "TASK_STATE_COMPLETED", "close it, drill acknowledged")
	}
	return nil
}

func (p *a2aPeer) sent() []a2aMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]a2aMessage(nil), p.sends...)
}

// a2aCtl is fakeController with comments and a journal seq on events.
type a2aCtl struct {
	fakeController
	cmu      sync.Mutex
	comments []string
	seq      int64
}

func (f *a2aCtl) Comment(_ context.Context, p rpc.CommentParams) error {
	f.cmu.Lock()
	defer f.cmu.Unlock()
	f.comments = append(f.comments, p.Body)
	return nil
}

func (f *a2aCtl) got() []string {
	f.cmu.Lock()
	defer f.cmu.Unlock()
	return append([]string(nil), f.comments...)
}

// Send records without echoing: the peer's reply is the test's subject.
func (f *a2aCtl) Send(_ context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sends = append(f.sends, p)
	return rpc.SessionSendResult{Session: "s-" + p.Key}, nil
}

// turn emits one escalation session turn with journal seqs.
func (f *a2aCtl) turn(text string) {
	f.seq += 3
	now := time.Now()
	f.events <- rpc.SessionEvent{Type: rpc.EventTurnStart, Seq: f.seq - 2, Time: &now}
	f.events <- rpc.SessionEvent{Type: rpc.EventText, Seq: f.seq - 1, Text: text}
	f.events <- rpc.SessionEvent{Type: rpc.EventTurnEnd, Seq: f.seq}
}

func a2aEsc(id, cause string) store.Ticket {
	return store.Ticket{
		ID: id, Title: "drill needs a human", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
		Meta: []byte(`{"escalation":{"key":"escalation:` + id + `","pool":"chat","session":"s-esc","cause":"` + cause + `"}}`),
	}
}

func a2aSetup(t *testing.T, mode string, tickets ...store.Ticket) (*A2A, *a2aPeer, *a2aCtl, *clock.Fake, context.Context) {
	peer := newA2APeer(t, mode)
	ctl := &a2aCtl{fakeController: fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{}}}
	for _, tk := range tickets {
		ctl.ticket[tk.ID] = tk
	}
	clk := clock.NewFake(time.Date(2026, 11, 5, 8, 0, 0, 0, time.UTC))
	a := NewA2A(A2AConfig{Name: "hermes", URL: peer.srv.URL, Token: a2aToken, Pool: "operator", Escalations: []string{"run"}, Poll: 30 * time.Second, Clock: clk}, ctl, nil)
	// Each gateway gets its own context, so a test can stop one.
	actx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); a.wg.Wait() })
	return a, peer, ctl, clk, actx
}

func (f *a2aCtl) state(t *testing.T, id string) A2AState {
	t.Helper()
	f.mu.Lock()
	tk := f.ticket[id]
	f.mu.Unlock()
	e, _ := metaEscalationAny(tk)
	if s := e.A2A["hermes"]; s != nil {
		return *s
	}
	return A2AState{}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	require.Eventually(t, cond, 5*time.Second, time.Millisecond)
}

func TestA2AMessageReply(t *testing.T) {
	a, peer, ctl, _, ctx := a2aSetup(t, "message", a2aEsc("js-esc", "run"), a2aEsc("js-pool", "pool"))
	a.Resync(ctx)
	ctl.turn("Run js-run failed in implement; retry or close?")
	waitFor(t, func() bool { return len(ctl.sent()) == 1 })

	s := peer.sent()
	require.Len(t, s, 1, "the pool escalation is not routed")
	assert.Equal(t, "js-esc", s[0].ContextID, "context id is the ticket id")
	assert.Contains(t, s[0].Parts[0].Text, "Jobshop escalation js-esc (run)")
	assert.Contains(t, s[0].Parts[0].Text, "retry or close?")
	assert.Equal(t, rpc.SessionSendParams{Pool: "chat", Key: "escalation:js-esc", Text: "close it, drill acknowledged"}, ctl.sent()[0])
	waitFor(t, func() bool { return ctl.state(t, "js-esc").Turns == 1 })
	assert.Equal(t, A2AState{Context: "js-esc", Seq: 3, Turns: 1}, ctl.state(t, "js-esc"))
	assert.Empty(t, ctl.state(t, "js-pool").Context, "a cause not in escalations is not sent")

	// Resolved: one last message to the peer, then the conversation ends.
	a.event(ctx, rpc.Event{Kind: rpc.EventKindTicket, Op: "close", ID: "js-esc", Resolution: "done", Actor: "s-esc"})
	waitFor(t, func() bool { return ctl.state(t, "js-esc").Done })
	s = peer.sent()
	require.Len(t, s, 2)
	assert.Contains(t, s[1].Parts[0].Text, "was resolved (done) by s-esc")
}

func TestA2ATaskPolledAndResumed(t *testing.T) {
	a, peer, ctl, clk, ctx := a2aSetup(t, "task", a2aEsc("js-esc", "run"))
	oldCtx, stopOld := context.WithCancel(ctx)
	a.Resync(oldCtx)
	ctl.turn("summary")
	waitFor(t, func() bool { return ctl.state(t, "js-esc").State == A2AWorking })
	assert.Equal(t, "t-1", ctl.state(t, "js-esc").Task)
	assert.Empty(t, ctl.sent(), "a working task has no reply yet")

	// Restart the gateway: the new one resumes polling from meta.gateway.
	stopOld()
	a.wg.Wait()
	// Its own clock: the old gateway's abandoned poll timer stays on clk.
	clk2 := clock.NewFake(clk.Now())
	cfg := a.cfg
	cfg.Clock = clk2
	a2 := NewA2A(cfg, ctl, nil)
	t.Cleanup(a2.wg.Wait)
	a2.Resync(ctx)

	// Poll 1: still working. Poll 2: completed, the reply reaches the session.
	for peer.getsN() < 2 {
		require.True(t, clk2.WaitFor(1, t.Context().Done()))
		ch := peer.waitCall()
		clk2.Advance(30 * time.Second)
		<-ch
	}
	waitFor(t, func() bool { return len(ctl.sent()) == 1 })
	assert.Equal(t, "close it, drill acknowledged", ctl.sent()[0].Text)
	assert.Len(t, peer.sent(), 1, "the restarted gateway did not resend the turn")
	waitFor(t, func() bool { return ctl.state(t, "js-esc").Task == "" })
}

func (p *a2aPeer) getsN() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.gets
}

func (p *a2aPeer) waitCall() <-chan struct{} {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func TestA2AInputRequired(t *testing.T) {
	a, peer, ctl, _, ctx := a2aSetup(t, "input", a2aEsc("js-esc", "run"))
	a.Resync(ctx)
	ctl.turn("summary")
	waitFor(t, func() bool { return len(ctl.sent()) == 1 })
	assert.Equal(t, "which run?", ctl.sent()[0].Text)
	assert.Equal(t, A2AInputRequired, ctl.state(t, "js-esc").State)

	// The session answers: it continues the same task.
	ctl.turn("js-run")
	waitFor(t, func() bool { return len(ctl.sent()) == 2 })
	s := peer.sent()
	require.Len(t, s, 2)
	assert.Equal(t, "t-1", s[1].TaskID)
	assert.Equal(t, "thanks, resolved", ctl.sent()[1].Text)
}

func TestA2ARejectedStops(t *testing.T) {
	a, _, ctl, _, ctx := a2aSetup(t, "rejected", a2aEsc("js-esc", "run"))
	a.Resync(ctx)
	ctl.turn("summary")
	waitFor(t, func() bool { return ctl.state(t, "js-esc").Done })
	require.Len(t, ctl.got(), 1)
	assert.Contains(t, ctl.got()[0], "ended rejected: not my job")
	assert.Empty(t, ctl.sent())
}

func TestA2AAuthFailureCommentsOnce(t *testing.T) {
	logs := &lockedBuf{}
	a, peer, ctl, clk, ctx := a2aSetup(t, "auth", a2aEsc("js-esc", "run"))
	a.log = slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	a.Resync(ctx)
	ctl.turn("summary")
	waitFor(t, func() bool { return len(ctl.got()) == 1 })
	// Two retries: still one comment.
	for range 2 {
		require.True(t, clk.WaitFor(1, ctx.Done()))
		ch := peer.waitCall()
		clk.Advance(10 * time.Minute)
		<-ch
	}
	assert.Len(t, ctl.got(), 1)
	assert.Contains(t, ctl.got()[0], "peer refused the token")
	for _, c := range ctl.got() {
		assert.NotContains(t, c, a2aToken)
	}
	ctl.mu.Lock()
	meta := string(ctl.ticket["js-esc"].Meta)
	ctl.mu.Unlock()
	assert.NotContains(t, meta, a2aToken)
	assert.NotContains(t, logs.String(), a2aToken)
	assert.Contains(t, logs.String(), "peer refused the token")
}

// lockedBuf is a log sink safe for the gateway's goroutines.
type lockedBuf struct {
	mu sync.Mutex
	b  strings.Builder
}

func (l *lockedBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}
