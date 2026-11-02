package gateway

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

type fakeAdapter struct {
	mu     sync.Mutex
	posts  []string
	reacts []string
	opened int
}

func (f *fakeAdapter) Network() string { return "fake" }
func (f *fakeAdapter) Run(ctx context.Context, _ func(context.Context, Message)) error {
	<-ctx.Done()
	return ctx.Err()
}

func (f *fakeAdapter) Post(_ context.Context, t Thread, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.posts = append(f.posts, t.Room+"/"+t.Root+": "+text)
	return nil
}

func (f *fakeAdapter) React(_ context.Context, _ Thread, msgID, _ string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.reacts = append(f.reacts, msgID)
	return nil
}

func (f *fakeAdapter) Open(_ context.Context, room, text string) (Thread, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.opened++
	root := fmt.Sprintf("$open%d", f.opened)
	f.posts = append(f.posts, room+"/"+root+": "+text)
	return Thread{Room: room, Root: root}, nil
}

func (f *fakeAdapter) got() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.posts...)
}

// fakeController echoes each send as a turn on the subscription. A
// session that existed before (old) replays an old turn first.
type fakeController struct {
	mu     sync.Mutex
	sends  []rpc.SessionSendParams
	events chan rpc.SessionEvent
	old    bool
	ticket map[string]store.Ticket
}

func (f *fakeController) List(_ context.Context, p rpc.ListParams) ([]store.Ticket, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []store.Ticket
	for _, t := range f.ticket {
		if slices.Contains(t.Labels, p.Label) && (p.Status == "" || t.Status == p.Status) {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeController) Show(_ context.Context, id string) (rpc.ShowResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.ticket[id]
	if !ok {
		return rpc.ShowResult{}, store.ErrNotFound
	}
	return rpc.ShowResult{Ticket: t}, nil
}

func (f *fakeController) Update(_ context.Context, p rpc.UpdateParams) (store.Ticket, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t := f.ticket[p.ID]
	if p.IfRevision != 0 && p.IfRevision != t.Revision {
		return t, store.ErrConflict
	}
	t.Meta = p.Meta
	t.Revision++
	f.ticket[p.ID] = t
	return t, nil
}

func (f *fakeController) Send(_ context.Context, p rpc.SessionSendParams) (rpc.SessionSendResult, error) {
	f.mu.Lock()
	f.sends = append(f.sends, p)
	f.mu.Unlock()
	now := time.Now()
	go func() {
		f.events <- rpc.SessionEvent{Type: rpc.EventTurnStart, Time: &now, Text: p.Text}
		f.events <- rpc.SessionEvent{Type: rpc.EventText, Text: "echo: "}
		f.events <- rpc.SessionEvent{Type: rpc.EventText, Text: p.Text}
		f.events <- rpc.SessionEvent{Type: rpc.EventTurnEnd, StopReason: "end_turn"}
	}()
	return rpc.SessionSendResult{Session: "s-" + p.Key}, nil
}

func (f *fakeController) Subscribe(ctx context.Context, _ rpc.SessionSubscribeParams, fn func(rpc.SessionEvent)) (rpc.SessionSubscribeResult, error) {
	if f.old {
		then := time.Now().Add(-time.Hour)
		fn(rpc.SessionEvent{Type: rpc.EventTurnStart, Time: &then})
		fn(rpc.SessionEvent{Type: rpc.EventText, Text: "old reply"})
		fn(rpc.SessionEvent{Type: rpc.EventTurnEnd})
	}
	for {
		select {
		case <-ctx.Done():
			return rpc.SessionSubscribeResult{}, nil
		case ev := <-f.events:
			fn(ev)
		}
	}
}

func (f *fakeController) sent() []rpc.SessionSendParams {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]rpc.SessionSendParams(nil), f.sends...)
}

func setup(t *testing.T, old bool) (*Gateway, *fakeAdapter, *fakeController, context.Context) {
	ad := &fakeAdapter{}
	ctl := &fakeController{events: make(chan rpc.SessionEvent, 16), old: old}
	g := New(Config{Pool: "operator", Rooms: []string{"!room"}, Senders: []string{"@op"}}, ad, ctl, nil)
	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(func() { cancel(); g.wg.Wait() })
	return g, ad, ctl, ctx
}

func TestRouting(t *testing.T) {
	g, ad, ctl, ctx := setup(t, false)
	th := Thread{Room: "!room", Root: "$root"}

	// Others are ignored: wrong sender, wrong room.
	g.Handle(ctx, Message{Thread: th, ID: "$a", Sender: "@stranger", Text: "hi"})
	g.Handle(ctx, Message{Thread: Thread{Room: "!other", Root: "$x"}, ID: "$b", Sender: "@op", Text: "hi"})
	assert.Empty(t, ctl.sent())

	g.Handle(ctx, Message{Thread: th, ID: "$c", Sender: "@op", Text: "hello"})
	require.Eventually(t, func() bool { return len(ad.got()) == 1 }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "!room/$root: echo: hello", ad.got()[0])
	s := ctl.sent()
	require.Len(t, s, 1)
	assert.Equal(t, rpc.SessionSendParams{Pool: "operator", Key: "fake:!room:$root", Text: "hello"}, s[0])
	assert.Equal(t, []string{"$c"}, ad.reacts)

	// The marker sends with interrupt and is stripped.
	g.Handle(ctx, Message{Thread: th, ID: "$d", Sender: "@op", Text: "! stop that"})
	s = ctl.sent()
	require.Len(t, s, 2)
	assert.True(t, s[1].Interrupt)
	assert.Equal(t, "stop that", s[1].Text)
	// One watcher per session.
	g.mu.Lock()
	assert.Len(t, g.watched, 1)
	g.mu.Unlock()
}

func TestOldTurnsNotReposted(t *testing.T) {
	g, ad, _, ctx := setup(t, true)
	g.Handle(ctx, Message{Thread: Thread{Room: "!room", Root: "$r"}, ID: "$m", Sender: "@op", Text: "new"})
	require.Eventually(t, func() bool { return len(ad.got()) == 1 }, 5*time.Second, 10*time.Millisecond)
	assert.Equal(t, "!room/$r: echo: new", ad.got()[0])
}

func TestMatrixMessage(t *testing.T) {
	ev := func(c *event.MessageEventContent) *event.Event {
		return &event.Event{
			ID: "$e", RoomID: "!room", Sender: "@op", Type: event.EventMessage,
			Content: event.Content{Parsed: c},
		}
	}
	m, ok := matrixMessage(ev(&event.MessageEventContent{MsgType: event.MsgText, Body: "top"}))
	require.True(t, ok)
	assert.Equal(t, Thread{Room: "!room", Root: "$e"}, m.Thread, "a top-level message roots its own thread")

	in := &event.MessageEventContent{MsgType: event.MsgText, Body: "reply"}
	in.RelatesTo = (&event.RelatesTo{}).SetThread(id.EventID("$root"), "$prev")
	m, ok = matrixMessage(ev(in))
	require.True(t, ok)
	assert.Equal(t, "$root", m.Thread.Root)

	edit := &event.MessageEventContent{MsgType: event.MsgText, Body: "* fix"}
	edit.SetEdit("$e")
	_, ok = matrixMessage(ev(edit))
	assert.False(t, ok, "edits are not new messages")

	_, ok = matrixMessage(ev(&event.MessageEventContent{MsgType: event.MsgImage}))
	assert.False(t, ok)
}
