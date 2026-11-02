package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func TestEscalationThread(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-esc": {
			ID: "js-esc", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Body: "Run js-run needs a human.",
			Meta: []byte(`{"run":{"id":"js-run","workflow":"drill","escalation":"js-run/1"}}`),
		},
		// Not an escalation: no meta.run.
		"js-other": {ID: "js-other", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); g.wg.Wait() })

	g.Escalations(ctx)
	g.Escalations(ctx) // dedup: a second pass opens nothing
	assert.Equal(t, 1, ad.opened)
	s := ctl.sent()
	require.Len(t, s, 1)
	assert.Equal(t, "operator", s[0].Pool)
	assert.Equal(t, "fake:!esc:$open1", s[0].Key)
	assert.Contains(t, s[0].Text, "jobshop run resume js-run")
	assert.Contains(t, s[0].Text, "Run js-run needs a human.")
	ctl.mu.Lock()
	tk := ctl.ticket["js-esc"]
	ctl.mu.Unlock()
	e, ok := metaEscalation(tk)
	require.True(t, ok)
	assert.Equal(t, Escalation{Gateway: "local", Room: "!esc", Root: "$open1", Key: "fake:!esc:$open1", Session: "s-fake:!esc:$open1"}, e)
	assert.Contains(t, string(tk.Meta), `"run":`, "other meta packs are kept")

	// The session's summary goes to the thread.
	require.Eventually(t, func() bool { return len(ad.got()) == 2 }, 5*time.Second, 10*time.Millisecond)
	assert.True(t, strings.HasPrefix(ad.got()[1], "!esc/$open1: echo: You are handling escalation ticket js-esc"))

	// The operator replies in the thread: same key, same session.
	g.Handle(ctx, Message{Thread: Thread{Room: "!esc", Root: "$open1"}, ID: "$r", Sender: "@op", Text: "resume it"})
	s = ctl.sent()
	require.Len(t, s, 2)
	assert.Equal(t, "fake:!esc:$open1", s[1].Key)
	require.Eventually(t, func() bool { return len(ad.got()) == 3 }, 5*time.Second, 10*time.Millisecond)

	// The session closes the ticket: one final message, then no more.
	ctl.mu.Lock()
	tk = ctl.ticket["js-esc"]
	tk.Status, tk.Resolution = store.StatusClosed, "done"
	tk.Revision++
	ctl.ticket["js-esc"] = tk
	ctl.mu.Unlock()
	g.Escalations(ctx)
	g.Escalations(ctx)
	got := ad.got()
	require.Len(t, got, 4)
	assert.Equal(t, "!esc/$open1: jobshop: escalation js-esc closed (done). Replies here still reach the same session.", got[3])
	ctl.mu.Lock()
	e, _ = metaEscalation(ctl.ticket["js-esc"])
	ctl.mu.Unlock()
	assert.True(t, e.Final)
}

func TestEscalationOwnedElsewhere(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-esc": {
			ID: "js-esc", Status: store.StatusInProgress, Labels: []string{"needs-human"}, Revision: 3,
			Meta: []byte(`{"run":{"id":"js-run","workflow":"drill","escalation":"js-run/1"},"gateway":{"gateway":"other","room":"!x","root":"$y","key":"k"}}`),
		},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	g.Escalations(t.Context())
	assert.Zero(t, ad.opened)
	assert.Empty(t, ctl.sent())
}

// TestPoolEscalationThread: a pool quarantine escalation (#146) opens a
// thread like a run escalation, with a prompt about the pool.
func TestPoolEscalationThread(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-pool": {
			ID: "js-pool", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Body: "Pool dev starts no sessions.",
			Meta: []byte(`{"pool":{"name":"dev","quarantine":"spawn: repo /x: no such file"}}`),
		},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); g.wg.Wait() })

	g.Escalations(ctx)
	assert.Equal(t, 1, ad.opened)
	s := ctl.sent()
	require.Len(t, s, 1)
	assert.Contains(t, s[0].Text, "jobshop pool unquarantine dev")
	assert.Contains(t, s[0].Text, "Pool dev starts no sessions.")
	ctl.mu.Lock()
	tk := ctl.ticket["js-pool"]
	ctl.mu.Unlock()
	assert.Contains(t, openingText(tk), "pool dev is quarantined (spawn: repo /x: no such file)")
}
