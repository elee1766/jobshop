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
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
	_, err = fmt.Fprintf(f, "  chat:\n    project: js\n    agent: fake\n    type: chat\n    dir: %q\n", root.Dir)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	stop := start()
	defer stop()
	ctx := t.Context()
	admin := dial(t, root.AdminSocket())

	// Ticket pools do not start sessions by key.
	err = admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{Pool: "dev", Key: "k", Text: "hi"}, nil)
	assert.ErrorIs(t, err, store.ErrInvalid)

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
	require.Eventually(t, func() bool { return ev.has(rpc.EventText, "echo: again") && ev.has(rpc.EventTurnEnd, "end_turn") },
		20*time.Second, 50*time.Millisecond)
	got := ev.get()
	assert.Equal(t, rpc.EventState, got[0].Type)
	assert.Equal(t, store.StatusOpen, got[0].Status)
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

	// Reconnect from the last seq: only new events arrive.
	var ev2 events
	subCtx, cancel = context.WithCancel(ctx)
	defer cancel()
	go func() {
		res, _ := sub.Subscribe(subCtx, rpc.SessionSubscribeParams{ID: sent.Session, SinceSeq: res.Seq}, ev2.add)
		done <- res
	}()
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: sent.Session, Text: "hold on"}, nil))
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root.Dir, "held"))
		return err == nil
	}, 20*time.Second, 20*time.Millisecond)
	// Interrupt the held turn; the message is delivered right after.
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionSend, rpc.SessionSendParams{ID: sent.Session, Text: "stop", Interrupt: true}, nil))
	require.Eventually(t, func() bool { return ev2.has(rpc.EventTurnEnd, "cancelled") && ev2.has(rpc.EventText, "echo: stop") },
		20*time.Second, 50*time.Millisecond)
	for _, e := range ev2.get() {
		if e.Seq > 0 {
			assert.Greater(t, e.Seq, res.Seq)
		}
		assert.NotContains(t, e.Text, "again", "old events not replayed")
	}

	// Closing the session ends the stream with a state event.
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sent.Session}, nil))
	select {
	case r := <-done:
		assert.True(t, r.Closed)
		assert.Equal(t, "done", r.Resolution)
	case <-time.After(20 * time.Second):
		t.Fatal("subscribe did not end on close")
	}
	last := ev2.get()[len(ev2.get())-1]
	assert.Equal(t, rpc.EventState, last.Type)
	assert.Equal(t, store.StatusClosed, last.Status)
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
	var manual store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Project: "js", Agent: "fake"}, &manual))
	_, err = send(rpc.SessionSendParams{ID: manual.ID, Attach: true})
	assert.ErrorIs(t, err, store.ErrInvalid)

	// The same key after a controller restart reaches the same session.
	stop()
	stop = start()
	defer func() { stop() }()
	admin = dial(t, root.AdminSocket())
	again, err := send(rpc.SessionSendParams{Pool: "chat", Key: "matrix:r:a", Text: "back"})
	require.NoError(t, err)
	assert.Equal(t, a.Session, again.Session, "same key, same session after restart")
	assert.False(t, again.Started)

	// With no more messages the idle TTL closes it as idle.
	var s rpc.ShowResult
	require.Eventually(t, func() bool {
		return admin.Call(ctx, rpc.MethodTicketShow, rpc.IDParams{ID: b.Session}, &s) == nil && s.Ticket.Status == store.StatusClosed
	}, 30*time.Second, 100*time.Millisecond)
	assert.Equal(t, "idle", s.Ticket.Resolution)
}
