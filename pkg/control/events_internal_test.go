package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

// A full subscriber buffer marks it lagged and drops it; publish (the
// pump, so every write's wake-up) never waits on it, and other
// subscribers still get every event.
func TestHubLaggedDoesNotBlock(t *testing.T) {
	t.Parallel()
	h := newHub(2)
	slow, _ := h.add()
	fast, _ := h.add()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := int64(1); i <= 5; i++ {
			h.publish([]rpc.Event{{Seq: i}})
			<-fast.ch // fast keeps up
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("publish blocked on a slow subscriber")
	}
	select {
	case <-slow.lagged:
	default:
		t.Fatal("slow subscriber not marked lagged")
	}
	assert.Len(t, slow.ch, 2, "events before the overflow stay queued")
	assert.Equal(t, int64(5), h.last)
	h.mu.Lock()
	_, still := h.subs[slow]
	h.mu.Unlock()
	assert.False(t, still)
	for range 100 {
		h.kick() // never blocks
	}
}

func TestToEvent(t *testing.T) {
	t.Parallel()
	at := time.Unix(100, 0)
	sess := &domain.Ticket{ID: "js-s", Kind: domain.KindSession, Status: domain.StatusClosed, Resolution: "crashed"}
	ev := toEvent(store.HistoryEntry{Seq: 3, TicketID: "js-s", At: at, Op: "close", Revision: 4, Diff: json.RawMessage(`{"resolution":"crashed"}`)}, sess)
	want := rpc.Event{
		Seq: 3, Time: at, Kind: rpc.EventKindSession, Op: rpc.OpSessionCrash, ID: "js-s", Revision: 4,
		TicketKind: domain.KindSession, Status: domain.StatusClosed, Resolution: "crashed",
	}
	assert.Equal(t, want, ev)

	run := &domain.Ticket{ID: "js-r", Kind: domain.KindRun, Status: domain.StatusOpen, Fields: json.RawMessage(`{"workflow":"w","state":"needs_human"}`)}
	ev = toEvent(store.HistoryEntry{Seq: 4, TicketID: "js-r", Op: "update", Diff: json.RawMessage(`{"fields":{"workflow":"w","state":"needs_human"}}`)}, run)
	assert.Equal(t, rpc.EventKindRun, ev.Kind)
	assert.Equal(t, rpc.OpRunState, ev.Op)
	assert.Equal(t, domain.RunNeedsHuman, ev.State)

	ev = toEvent(store.HistoryEntry{Seq: 5, TicketID: "js-t", Op: "label_add", Diff: json.RawMessage(`{"label":"needs-human"}`)}, nil)
	assert.Equal(t, rpc.EventKindTicket, ev.Kind)
	assert.Equal(t, "needs-human", ev.Label)
	require.True(t, eventMatches(rpc.EventsSubscribeParams{Labels: []string{"needs-human"}}, ev))
	assert.False(t, eventMatches(rpc.EventsSubscribeParams{Kinds: []string{rpc.EventKindRun}}, ev))
	assert.False(t, eventMatches(rpc.EventsSubscribeParams{Labels: []string{"other"}}, ev))
}
