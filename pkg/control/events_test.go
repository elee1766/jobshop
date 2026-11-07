package control_test

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type eventLog struct {
	mu  sync.Mutex
	evs []rpc.Event
}

func (l *eventLog) add(ev rpc.Event) { l.mu.Lock(); l.evs = append(l.evs, ev); l.mu.Unlock() }

func (l *eventLog) get() []rpc.Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]rpc.Event(nil), l.evs...)
}

// has reports whether an event with kind, op and id arrived.
func (l *eventLog) has(kind, op, id string) bool {
	for _, ev := range l.get() {
		if ev.Kind == kind && ev.Op == op && ev.ID == id {
			return true
		}
	}
	return false
}

func subscribe(ctx context.Context, t *testing.T, sock string, p rpc.EventsSubscribeParams) (*eventLog, <-chan rpc.EventsSubscribeResult) {
	t.Helper()
	c := dial(t, sock)
	l := &eventLog{}
	ready := make(chan struct{})
	var once sync.Once
	done := make(chan rpc.EventsSubscribeResult, 1)
	go func() {
		res, _ := c.Events(ctx, p, func(ev rpc.Event) {
			if ev.Kind == rpc.EventKindReady {
				once.Do(func() { close(ready) })
				return
			}
			l.add(ev)
		})
		done <- res
	}()
	select {
	case <-ready:
	case <-time.After(5 * time.Second):
		t.Fatal("no ready event")
	}
	return l, done
}

func TestEventsSubscribe(t *testing.T) {
	t.Parallel()
	dir, err := os.MkdirTemp("", "js")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), []byte("projects:\n  js: {}\ndefault_project: js\n"), 0o600))
	root := config.Root{Dir: dir}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- control.Run(ctx, root, control.Options{Tick: time.Hour}) }()
	t.Cleanup(func() { cancel(); require.NoError(t, <-done) })
	admin := dial(t, root.AdminSocket())

	all, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{})
	labeled, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{Labels: []string{"needs-human"}})
	sessions, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{Kinds: []string{rpc.EventKindSession}})

	var a, b store.Ticket
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "a", Pool: "dev"}, &a))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "b"}, &b))
	require.NoError(t, admin.Call(ctx, rpc.MethodLabelAdd, rpc.LabelParams{ID: a.ID, Label: "needs-human"}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: a.ID, Body: "hi"}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodClaim, rpc.ClaimParams{ID: a.ID, Holder: "h"}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodRelease, rpc.ReleaseParams{ID: a.ID}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodLabelRemove, rpc.LabelParams{ID: a.ID, Label: "needs-human"}, nil))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: a.ID}, nil))
	var sess rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &sess))
	require.NoError(t, admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: sess.Ticket.ID}, nil))

	for _, op := range []string{"create", "label_add", "comment", "claim", "release", "label_remove", "close"} {
		require.Eventually(t, func() bool { return all.has(rpc.EventKindTicket, op, a.ID) }, dialWait, 10*time.Millisecond, op)
	}
	require.Eventually(t, func() bool { return all.has(rpc.EventKindSession, rpc.OpSessionStop, sess.Ticket.ID) }, dialWait, 10*time.Millisecond)
	assert.True(t, all.has(rpc.EventKindSession, rpc.OpSessionStart, sess.Ticket.ID))
	evs := all.get()
	var lastSeq int64
	for _, ev := range evs {
		if ev.Seq == 0 {
			continue // Transient state notifications have no replay cursor.
		}
		assert.Greater(t, ev.Seq, lastSeq, "in seq order")
		lastSeq = ev.Seq
	}

	// Label filter: label_add and label_remove name the label; other
	// events of a only match while it still had the label.
	require.Eventually(t, func() bool { return labeled.has(rpc.EventKindTicket, "label_remove", a.ID) }, dialWait, 10*time.Millisecond)
	for _, ev := range labeled.get() {
		assert.NotEqual(t, b.ID, ev.ID, "b never had the label")
	}
	assert.True(t, labeled.has(rpc.EventKindTicket, "label_add", a.ID))
	// Kind filter.
	require.Eventually(t, func() bool { return sessions.has(rpc.EventKindSession, rpc.OpSessionStop, sess.Ticket.ID) }, dialWait, 10*time.Millisecond)
	for _, ev := range sessions.get() {
		assert.Equal(t, rpc.EventKindSession, ev.Kind)
	}

	// Replay: a subscriber from a's create seq sees everything after it,
	// without any new write.
	var first int64
	for _, ev := range evs {
		if ev.ID == a.ID && ev.Op == "create" {
			first = ev.Seq
		}
	}
	require.NotZero(t, first)
	replay, _ := subscribe(ctx, t, root.AdminSocket(), rpc.EventsSubscribeParams{SinceSeq: first})
	got := replay.get()
	require.NotEmpty(t, got)
	assert.Equal(t, b.ID, got[0].ID, "replay starts after since_seq")
	assert.True(t, replay.has(rpc.EventKindTicket, "close", a.ID))
	assert.True(t, replay.has(rpc.EventKindSession, rpc.OpSessionStop, sess.Ticket.ID))

	// Session sockets cannot subscribe.
	var s2 rpc.SessionNewResult
	require.NoError(t, admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode", Pool: "dev"}, &s2))
	_, err = dial(t, s2.Socket).Events(ctx, rpc.EventsSubscribeParams{}, func(rpc.Event) {})
	assert.ErrorIs(t, err, rpc.ErrForbidden)
}
