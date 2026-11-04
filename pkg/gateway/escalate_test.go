package gateway

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// escMeta is a run escalation whose chat the controller opened (#379):
// session s-escalation:<id> in pool chat, keyed escalation:<id>.
func escMeta(id string) []byte {
	return []byte(`{"run":{"id":"js-run","workflow":"drill","escalation":"js-run/1"},` +
		`"escalation":{"key":"escalation:` + id + `","pool":"chat","session":"s-escalation:` + id + `"}}`)
}

// summary emits one session turn, as the controller's first prompt does.
func summary(ctl *fakeController, text string) {
	now := time.Now()
	ctl.events <- rpc.SessionEvent{Type: rpc.EventTurnStart, Time: &now}
	ctl.events <- rpc.SessionEvent{Type: rpc.EventText, Text: text}
	ctl.events <- rpc.SessionEvent{Type: rpc.EventTurnEnd, StopReason: "end_turn"}
}

// TestEscalationThread: the gateway mirrors the controller's escalation
// chat. It opens one thread, binds it to the chat's key and pool, posts
// the session's turns, sends no first prompt, and posts one final
// message when the ticket closes.
func TestEscalationThread(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-esc": {
			ID: "js-esc", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Body: "Run js-run needs a human.", CreatedAt: time.Now().Add(-time.Minute), Meta: escMeta("js-esc"),
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
	assert.Empty(t, ctl.sent(), "the controller sent the first prompt, not the gateway")
	ctl.mu.Lock()
	tk := ctl.ticket["js-esc"]
	ctl.mu.Unlock()
	e, ok := metaEscalation(tk)
	require.True(t, ok)
	assert.Equal(t, Escalation{Gateway: "local", Room: "!esc", Root: "$open1", Key: "escalation:js-esc", Session: "s-escalation:js-esc"}, e)
	assert.Contains(t, string(tk.Meta), `"run":`, "other meta packs are kept")

	// The session's summary goes to the thread.
	summary(ctl, "Run js-run failed in implement.")
	require.Eventually(t, func() bool { return len(ad.got()) == 2 }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "!esc/$open1: Run js-run failed in implement.", ad.got()[1])

	// The operator replies in the thread: the chat's key and pool.
	g.Handle(ctx, Message{Thread: Thread{Room: "!esc", Root: "$open1"}, ID: "$r", Sender: "@op", Text: "resume it"})
	s := ctl.sent()
	require.Len(t, s, 1)
	assert.Equal(t, "escalation:js-esc", s[0].Key)
	assert.Equal(t, "chat", s[0].Pool)
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

// TestEscalationNoChat: without a controller chat (no escalation chat
// config) the gateway opens no thread and starts no session.
func TestEscalationNoChat(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-esc": {
			ID: "js-esc", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Meta: []byte(`{"run":{"id":"js-run","workflow":"drill","escalation":"js-run/1"}}`),
		},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	g.Escalations(t.Context())
	assert.Zero(t, ad.opened)
	assert.Empty(t, ctl.sent())
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

// TestPoolEscalationThread: a pool quarantine escalation (#146) with a
// controller chat opens a thread from the event alone, no poll.
func TestPoolEscalationThread(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-pool": {
			ID: "js-pool", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Body: "Pool dev starts no sessions.",
			Meta: []byte(`{"pool":{"name":"dev","quarantine":"spawn: repo /x: no such file"},"escalation":{"key":"escalation:js-pool","pool":"chat","session":"s-escalation:js-pool"}}`),
		},
	}}
	ctl.changes = make(chan rpc.Event)
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx) }()
	t.Cleanup(func() { cancel(); assert.NoError(t, <-done) })

	ctl.changes <- rpc.Event{Kind: rpc.EventKindTicket, Op: "create", ID: "js-pool", Status: store.StatusOpen}
	require.Eventually(t, func() bool { ad.mu.Lock(); defer ad.mu.Unlock(); return ad.opened == 1 }, 5*time.Second, 10*time.Millisecond)
	assert.Empty(t, ctl.sent())
	ctl.mu.Lock()
	tk := ctl.ticket["js-pool"]
	ctl.mu.Unlock()
	assert.Contains(t, openingText(tk), "pool dev is quarantined (spawn: repo /x: no such file)")
}

// TestEscalationEvents drives threads from events.subscribe through Run,
// with no poll: ready resyncs, the update recording the chat opens a
// thread, close posts the final message, and lagged reconnects and
// resyncs.
func TestEscalationEvents(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), changes: make(chan rpc.Event), ticket: map[string]store.Ticket{}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	g.retry = 10 * time.Millisecond
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx) }()
	t.Cleanup(func() { cancel(); assert.NoError(t, <-done) })

	esc := func(id string) store.Ticket {
		return store.Ticket{ID: id, Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1, Meta: escMeta(id)}
	}
	ctl.mu.Lock()
	ctl.ticket["js-a"] = esc("js-a")
	ctl.mu.Unlock()
	ctl.changes <- rpc.Event{Kind: rpc.EventKindTicket, Op: "update", ID: "js-a", Status: store.StatusOpen}
	opened := func() int { ad.mu.Lock(); defer ad.mu.Unlock(); return ad.opened }
	require.Eventually(t, func() bool { return opened() == 1 }, 5*time.Second, 10*time.Millisecond)

	// Close event: final message.
	ctl.mu.Lock()
	tk := ctl.ticket["js-a"]
	tk.Status, tk.Resolution = store.StatusClosed, "done"
	tk.Revision++
	ctl.ticket["js-a"] = tk
	ctl.mu.Unlock()
	ctl.changes <- rpc.Event{Kind: rpc.EventKindTicket, Op: "close", ID: "js-a", Status: store.StatusClosed}
	require.Eventually(t, func() bool {
		for _, p := range ad.got() {
			if strings.Contains(p, "escalation js-a closed (done)") {
				return true
			}
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)

	// A ticket filed while the stream lagged is found by the resync.
	ctl.mu.Lock()
	ctl.ticket["js-b"] = esc("js-b")
	ctl.mu.Unlock()
	ctl.changes <- rpc.Event{Kind: rpc.EventKindLagged}
	require.Eventually(t, func() bool { return opened() == 2 }, 5*time.Second, 10*time.Millisecond)
}

// TestEscalationRethread: posts to a thread whose root the network
// rejects (#197) open a new thread, recorded on the ticket, and replies
// there reach the same session.
func TestEscalationRethread(t *testing.T) {
	ad := &fakeAdapter{gone: map[string]bool{"$open1": true}}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-esc": {
			ID: "js-esc", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			CreatedAt: time.Now().Add(-time.Minute), Meta: escMeta("js-esc"),
		},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); g.wg.Wait() })

	g.Escalations(ctx)
	// The summary fails on $open1 and lands in $open2.
	summary(ctl, "summary")
	require.Eventually(t, func() bool {
		return slices.Contains(ad.got(), "!esc/$open2: summary")
	}, 5*time.Second, 10*time.Millisecond)
	ctl.mu.Lock()
	e, _ := metaEscalation(ctl.ticket["js-esc"])
	ctl.mu.Unlock()
	assert.Equal(t, "$open2", e.Root)
	assert.Equal(t, "escalation:js-esc", e.Key, "the session key stays")

	g.Handle(ctx, Message{Thread: Thread{Room: "!esc", Root: "$open2"}, ID: "$r", Sender: "@op", Text: "resume it"})
	s := ctl.sent()
	require.Len(t, s, 1)
	assert.Equal(t, "escalation:js-esc", s[0].Key)
	assert.Equal(t, "chat", s[0].Pool)
	require.Eventually(t, func() bool {
		return slices.Contains(ad.got(), "!esc/$open2: echo: resume it")
	}, 5*time.Second, 10*time.Millisecond)

	// A restarted gateway maps the recorded thread back to the key.
	g2 := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	g2.Escalations(ctx)
	assert.Equal(t, "escalation:js-esc", g2.threadKey(Thread{Room: "!esc", Root: "$open2"}))
}

func TestQuestionThread(t *testing.T) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), ticket: map[string]store.Ticket{
		"js-q": {
			ID: "js-q", Status: store.StatusOpen, Labels: []string{"needs-human"}, Revision: 1,
			Body: "Which base branch?",
			Meta: []byte(`{"question":{"ticket":"js-step","session":"js-s1"}}`),
		},
	}}
	g := New(Config{Name: "local", Pool: "operator", Rooms: []string{"!esc"}, Senders: []string{"@op"}, EscalationRoom: "!esc"}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); g.wg.Wait() })

	g.Escalations(ctx)
	assert.Equal(t, 1, ad.opened)
	assert.Empty(t, ctl.sent(), "a question starts no chat session")
	require.Len(t, ad.got(), 1)
	assert.Contains(t, ad.got()[0], "Question js-q from js-s1 on js-step")
	assert.Contains(t, ad.got()[0], "Which base branch?")

	// The reply is the answer, not a chat message.
	g.Handle(ctx, Message{Thread: Thread{Room: "!esc", Root: "$open1"}, ID: "$r", Sender: "@op", Text: "use main"})
	assert.Empty(t, ctl.sent())
	ctl.mu.Lock()
	ans := ctl.answers
	ctl.mu.Unlock()
	assert.Equal(t, []rpc.AnswerParams{{ID: "js-q", Text: "use main", From: "@op"}}, ans)

	g.Escalations(ctx) // the fake closed the question
	got := ad.got()
	assert.Equal(t, "!esc/$open1: jobshop: question js-q closed (answered); the answer went to js-s1 and is a comment on js-step.", got[len(got)-1])
}
