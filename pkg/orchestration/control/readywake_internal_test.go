package control

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// listCounter counts List calls, the query the per-pass reads used to
// repeat once per session.
type listCounter struct {
	store.Store
	n atomic.Int64
}

func (l *listCounter) List(ctx context.Context, f store.Filter) ([]domain.Ticket, error) {
	l.n.Add(1)
	return l.Store.List(ctx, f)
}

// passListCalls makes n open sessions of pool dev, every other one holding
// a ticket, and returns how many List calls one supervise makes.
func passListCalls(t *testing.T, n int) int64 {
	t.Helper()
	c, _, _ := scanController(t)
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	ctx := t.Context()
	counter := &listCounter{Store: c.store}
	c.store = counter
	for i := range n {
		s, err := store.Create(ctx, c.store, "controller", store.NewTicket{
			Project: "js", Kind: domain.KindSession, Title: "s", Pool: "dev",
			Fields: []byte(`{"agent":"fake"}`),
		})
		require.NoError(t, err)
		if i%2 == 0 {
			_, err = store.Create(ctx, c.store, "controller", store.NewTicket{Project: "js", Kind: domain.KindTask, Title: "t", Pool: "dev"})
			require.NoError(t, err)
			_, _, err = c.store.ClaimNext(ctx, store.ClaimNext{Project: "js", Pool: "dev", Holder: s.ID, Lease: c.leaseDur()})
			require.NoError(t, err)
		}
	}
	counter.n.Store(0)
	c.supervise(ctx, c.config(), c.leaseDur())
	return counter.n.Load()
}

// TestSuperviseQueriesDoNotGrowWithSessions (JOB-197): the held state for
// heartbeats, gauges and the scale plan is one read per pass, so a pass
// over 20 sessions lists as often as one over 2.
func TestSuperviseQueriesDoNotGrowWithSessions(t *testing.T) {
	t.Parallel()
	few := passListCalls(t, 2)
	many := passListCalls(t, 20)
	assert.Equal(t, few, many, "List calls per supervise pass, 2 sessions vs 20")
	assert.Positive(t, few)
}

// TestReadyWake (JOB-197): a wake closes the channel every waiter of the
// pool holds, one taken after it is fresh, and another pool is left alone.
func TestReadyWake(t *testing.T) {
	t.Parallel()
	c := &Controller{}
	a1, a2, b := c.readyChan("a"), c.readyChan("a"), c.readyChan("b")
	c.wakePool("a")
	for _, ch := range []<-chan struct{}{a1, a2} {
		select {
		case <-ch:
		default:
			t.Fatal("waiter of the woken pool did not wake")
		}
	}
	select {
	case <-b:
		t.Fatal("another pool woke")
	default:
	}
	select {
	case <-c.readyChan("a"):
		t.Fatal("a channel taken after the wake is already closed")
	default:
	}
	c.wakeAllPools()
	select {
	case <-b:
	default:
		t.Fatal("wakeAllPools left pool b asleep")
	}
}

// TestNoteChangeWakes (JOB-197): history rows that make work claimable
// wake the pool; rows that cannot do not.
func TestNoteChangeWakes(t *testing.T) {
	t.Parallel()
	open := domain.Ticket{Kind: domain.KindTask, Status: domain.StatusOpen, Pool: "dev"}
	cases := []struct {
		name string
		op   string
		t    domain.Ticket
		wake bool
	}{
		{"create", "create", open, true},
		{"release", "release", open, true},
		{"lease expired", "lease_expired", open, true},
		{"reopen", "reopen", open, true},
		{"update of an open ticket (wait cleared)", "update", open, true},
		{"label", "label_add", open, false},
		{"claim", "claim", domain.Ticket{Kind: domain.KindTask, Status: domain.StatusInProgress, Pool: "dev"}, false},
		{"open ticket with no pool", "create", domain.Ticket{Kind: domain.KindTask, Status: domain.StatusOpen}, false},
		{"session create", "create", domain.Ticket{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: "dev"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Controller{runners: map[string]*runner{}}
			ch := c.readyChan("dev")
			c.noteChange(store.HistoryEntry{Op: tc.op}, &tc.t)
			select {
			case <-ch:
				assert.True(t, tc.wake, "woke")
			default:
				assert.False(t, tc.wake, "did not wake")
			}
		})
	}

	// A close can ready a dependent in any pool.
	c := &Controller{runners: map[string]*runner{}}
	other := c.readyChan("other")
	c.noteChange(store.HistoryEntry{Op: "close"}, &domain.Ticket{Kind: domain.KindTask, Status: domain.StatusClosed, Pool: "dev"})
	select {
	case <-other:
	default:
		t.Fatal("a close did not wake the other pool")
	}

	// A message to a pool wakes it.
	c = &Controller{runners: map[string]*runner{}}
	ch := c.readyChan("dev")
	c.noteChange(store.HistoryEntry{Op: "create"}, &domain.Ticket{Kind: domain.KindMessage, Fields: []byte(`{"to":"pool:dev"}`)})
	select {
	case <-ch:
	default:
		t.Fatal("a pool message did not wake the pool")
	}
}
