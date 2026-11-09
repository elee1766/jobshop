package control_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// events collects a subscription's events.
type events struct {
	mu  sync.Mutex
	all []rpc.SessionEvent
}

func (e *events) add(ev rpc.SessionEvent) { e.mu.Lock(); e.all = append(e.all, ev); e.mu.Unlock() }

func (e *events) get() []rpc.SessionEvent {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]rpc.SessionEvent(nil), e.all...)
}

func (e *events) has(typ, text string) bool {
	for _, ev := range e.get() {
		if ev.Type == typ && strings.Contains(ev.Text+ev.StopReason, text) {
			return true
		}
	}
	return false
}

// TestSessionSendSubscribe starts a chat session by pool and key, streams
// its turn through session.subscribe, resumes the stream by seq, and
// interrupts a busy turn.
func TestSessionSendSubscribe(t *testing.T) {
	root, start := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "  chat:\n    project: js\n    agent: fake\n    type: chat\n    prompt: CHAT-POOL-PROMPT\n    dir: %q\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())

	// Ticket pools do not start sessions by key.
	err = admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "dev", Key: "k", Text: "hi"}, nil)
	assert.ErrorIs(t, err, domain.ErrInvalid)

	var sent rpc.SessionSendResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:1", Text: "hello"}, &sent))
	assert.True(t, sent.Started)
	var again rpc.SessionSendResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:1", Text: "again"}, &again))
	assert.Equal(t, sent.Session, again.Session, "same key, same session")
	assert.False(t, again.Started)

	sub := dial(t, root.AdminSocket())
	var ev events
	subCtx, cancel := context.WithCancel(ctx)
	done := make(chan rpc.SessionSubscribeResult, 1)
	go func() {
		res, _ := sub.Subscribe(subCtx, rpc.SessionSubscribeParams{ID: sent.Session}, ev.add)
		done <- res
	}()
	passEventually(t, root.Dir, func() bool { return ev.has(rpc.EventText, "echo: again") && ev.has(rpc.EventTurnEnd, "end_turn") },
		"messages delivered")
	got := ev.get()
	// The pool prompt leads the first message turn only.
	prompted := 0
	for _, e := range got {
		if e.Type == rpc.EventTurnStart && strings.Contains(e.Context, "CHAT-POOL-PROMPT") {
			prompted++
			assert.NotContains(t, e.Text, "<messages>")
			assert.Contains(t, e.RawPrompt, "<messages>")
			assert.NotEmpty(t, e.Messages)
		}
	}
	assert.Equal(t, 1, prompted, "pool prompt on the first chat turn")
	assert.Equal(t, rpc.EventState, got[0].Type)
	assert.Equal(t, domain.StatusOpen, got[0].Status)
	assert.EqualValues(t, 1, got[0].First, "full replay starts at the journal beginning")
	var tool *rpc.ToolSummary
	for _, e := range got {
		if e.Type == rpc.EventTool && e.Tool.Status == "completed" {
			tool = e.Tool
		}
	}
	require.NotNil(t, tool)
	assert.Equal(t, "Reading", tool.Title, "title carried from the start frame")
	cancel()
	res := <-done
	require.Positive(t, res.Seq)

	// Tail starts at the last N frames (#614): nothing older is sent.
	var evTail events
	tailCtx, tailCancel := context.WithCancel(ctx)
	go func() {
		r, _ := sub.Subscribe(tailCtx, rpc.SessionSubscribeParams{ID: sent.Session, Tail: 3}, evTail.add)
		done <- r
	}()
	passEventually(t, root.Dir, func() bool {
		for _, e := range evTail.get() {
			if e.Seq == res.Seq {
				return true
			}
		}
		return false
	}, "tail reached the last frame")
	tailCancel()
	<-done
	require.Greater(t, evTail.get()[0].First, int64(1), "tail advertises the raw history cursor")
	for _, e := range evTail.get() {
		if e.Seq > 0 {
			assert.Greater(t, e.Seq, res.Seq-3, "only the last 3 frames")
		}
	}
	// Reconnect from the last seq: only new events arrive.
	var ev2 events
	subCtx, cancel = context.WithCancel(ctx)
	defer cancel()
	go func() {
		res, _ := sub.Subscribe(subCtx, rpc.SessionSubscribeParams{ID: sent.Session, SinceSeq: res.Seq}, ev2.add)
		done <- res
	}()
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: sent.Session, Text: "hold on"}, nil))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "held"))
		return err == nil
	}, "turn held")
	// Interrupt the held turn; the message is delivered right after.
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: sent.Session, Text: "stop", Interrupt: true}, nil))
	passEventually(t, root.Dir, func() bool { return ev2.has(rpc.EventTurnEnd, "cancelled") && ev2.has(rpc.EventText, "echo: stop") },
		"interrupt delivered")
	for _, e := range ev2.get() {
		if e.Seq > 0 {
			assert.Greater(t, e.Seq, res.Seq)
		}
		assert.NotContains(t, e.Text, "again", "old events not replayed")
	}

	// Stop (#535): session.cancel ends the held turn and sends nothing.
	passEventually(t, root.Dir, func() bool {
		var sessions []rpc.SessionInfo
		if err := admin.Call(ctx, rpc.MethodSessionList, rpc.SessionListParams{}, &sessions); err != nil {
			return false
		}
		for _, session := range sessions {
			if session.ID == sent.Session {
				return !session.Busy
			}
		}
		return false
	}, "interrupt reply finished before idle cancel")
	require.NoError(t, os.Remove(filepath.Join(root.Dir, "held")))
	var idle rpc.SessionCancelResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionCancel, rpc.IDParams{ID: sent.Session}, &idle))
	assert.False(t, idle.Cancelled)
	cancels := func() int {
		n := 0
		for _, e := range ev2.get() {
			if e.Type == rpc.EventTurnEnd && e.StopReason == "cancelled" {
				n++
			}
		}
		return n
	}
	before := cancels()
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: sent.Session, Text: "hold on"}, nil))
	passEventually(t, root.Dir, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "held"))
		return err == nil
	}, "turn held")
	var stopped rpc.SessionCancelResult
	passEventually(t, root.Dir, func() bool {
		return admin.Call(ctx, rpc.MethodSessionCancel, rpc.IDParams{ID: sent.Session}, &stopped) == nil && stopped.Cancelled
	}, "session cancelled")
	passEventually(t, root.Dir, func() bool { return cancels() == before+1 }, "cancel turn ended")
	require.ErrorIs(t, admin.Call(ctx, rpc.MethodSessionCancel, rpc.IDParams{ID: "js-nope"}, nil), domain.ErrInvalid)

	// Closing the session ends the stream with a state event.
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sent.Session}, nil))
	// Closing wakes the stream through the controller's event subscription.
	var r rpc.SessionSubscribeResult
	passEventually(t, root.Dir, func() bool {
		select {
		case r = <-done:
			return true
		default:
			return false
		}
	}, "subscribe did not end on close")
	assert.True(t, r.Closed)
	assert.Equal(t, "done", r.Resolution)
	last := ev2.get()[len(ev2.get())-1]
	assert.Equal(t, rpc.EventState, last.Type)
	assert.Equal(t, domain.StatusClosed, last.Status)
}

// TestChatPoolKeys: two keys get two sessions, a key keeps its session
// across a controller restart, rw attach sends need attach: rw, and an
// idle chat session is reaped after its pool's idle TTL.
func TestChatPoolKeys(t *testing.T) {
	root, start := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "  chat:\n    project: js\n    agent: fake\n    type: chat\n    attach: rw\n    idle_timeout: 3s\n    dir: %q\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())

	send := func(p rpc.SessionSendParams) (rpc.SessionSendResult, error) {
		var res rpc.SessionSendResult
		err := admin.Call(ctx, rpc.MethodSessionSend, p, &res)
		return res, err
	}
	a, err := send(rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:a", Text: "one"})
	require.NoError(t, err)
	b, err := send(rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:b", Text: "two"})
	require.NoError(t, err)
	assert.True(t, a.Started)
	assert.True(t, b.Started)
	assert.NotEqual(t, a.Session, b.Session, "two keys, two sessions")

	// rw attach: allowed on the chat pool, a probe sends nothing.
	probe, err := send(rpc.SessionSendParams{ID: a.Session, Attach: true})
	require.NoError(t, err)
	assert.Empty(t, probe.Message.ID)
	_, err = send(rpc.SessionSendParams{ID: a.Session, Text: "via attach", Attach: true})
	require.NoError(t, err)

	// A manual session in no pool refuses rw attach.
	var manual domain.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Project: "js", Agent: "fake"}, &manual))
	_, err = send(rpc.SessionSendParams{ID: manual.ID, Attach: true})
	assert.ErrorIs(t, err, domain.ErrInvalid)

	// The same key after a controller restart reaches the same session.
	stop()
	stop = start()
	defer func() { stop() }()
	admin = dial(t, root.AdminSocket())
	again, err := send(rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:a", Text: "back"})
	require.NoError(t, err)
	assert.Equal(t, a.Session, again.Session, "same key, same session after restart")
	assert.False(t, again.Started)

	// With no more messages the idle TTL closes it as idle: move the
	// controller's fake clock past it (JOB-73). The restarted session may
	// not have armed its idle poll yet, so each attempt moves it again.
	var s rpc.ShowResult
	passEventually(t, root.Dir, func() bool {
		advance(t, root.Dir, 4*time.Second)
		return admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: b.Session}, &s) == nil && s.Ticket.Status == domain.StatusClosed
	}, "idle chat session closed")
	assert.Equal(t, "idle", s.Ticket.Resolution)
}

// TestHumanSlot: a chat pool with max_sessions 1 starts one session over
// the cap for a human send, then refuses the next new key (#190).
func TestHumanSlot(t *testing.T) {
	root, start := poolRoot(t, "1m")
	f, err := os.OpenFile(filepath.Join(root.Dir, config.FileName), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fmt.Fprintf(f, "  chat:\n    project: js\n    agent: fake\n    type: chat\n    max_sessions: 1\n    dir: %q\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())
	send := func(key string) error {
		var res rpc.SessionSendResult
		return admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "chat", Key: key, Text: "hi"}, &res)
	}
	require.NoError(t, send("a"))
	require.NoError(t, send("b"), "the reserved human slot")
	assert.ErrorIs(t, send("c"), domain.ErrConflict)
	require.NoError(t, send("a"), "an open key still reaches its session")
	var res rpc.SessionNewResult
	err = admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Project: "js", Pool: "chat"}, &res)
	assert.ErrorIs(t, err, domain.ErrConflict)
}
