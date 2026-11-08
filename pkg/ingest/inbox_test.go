package ingest_test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// inboxStore opens a store whose clock the test moves.
func inboxStore(t *testing.T) (*sqlite.Store, *atomic.Int64) {
	t.Helper()
	var clock atomic.Int64
	clock.Store(time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC).UnixMilli())
	s, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "j.db"), sqlite.Options{
		Now: func() time.Time { return time.UnixMilli(clock.Load()).UTC() },
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s, &clock
}

func TestInboxPutDedupe(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, _ := inboxStore(t)
	it := store.InboxItem{Source: "github", DeliveryID: "d1", Endpoint: "gh-app", Kind: "issue_comment", RepoOrTeam: "a/b", Number: 7, Payload: json.RawMessage(`{"x":1}`)}
	a, dup, err := ingest.InboxPut(ctx, s, it)
	require.NoError(t, err)
	assert.Equal(t, "gh-app", a.Endpoint, "endpoint stored (#339)")
	assert.False(t, dup)
	assert.Equal(t, store.InboxPending, a.State)

	it.Payload = json.RawMessage(`{"x":2}`)
	b, dup, err := ingest.InboxPut(ctx, s, it)
	require.NoError(t, err)
	assert.True(t, dup, "same delivery id")
	assert.Equal(t, a.ID, b.ID)
	assert.JSONEq(t, `{"x":1}`, string(b.Payload), "first payload kept")

	_, _, err = ingest.InboxPut(ctx, s, store.InboxItem{DeliveryID: "d2"})
	require.ErrorIs(t, err, store.ErrInvalid)
	_, dup, err = ingest.InboxPut(ctx, s, store.InboxItem{Source: "generic", DeliveryID: "generic:d3"})
	require.NoError(t, err, "any endpoint source is stored (#399)")
	assert.False(t, dup)
}

func TestInboxClaimConcurrent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, _ := inboxStore(t)
	const n = 20
	for i := range n {
		_, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "linear", DeliveryID: string(rune('a' + i))})
		require.NoError(t, err)
	}
	var mu sync.Mutex
	got := map[int64]string{}
	var wg sync.WaitGroup
	for w := range 8 {
		holder := string(rune('A' + w))
		wg.Go(func() {
			for {
				it, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, holder, time.Minute)
				if err != nil {
					assert.ErrorIs(t, err, store.ErrNotFound)
					return
				}
				mu.Lock()
				prev, seen := got[it.ID]
				got[it.ID] = holder
				mu.Unlock()
				assert.False(t, seen, "item %d claimed by %s and %s", it.ID, prev, holder)
			}
		})
	}
	wg.Wait()
	assert.Len(t, got, n)
}

func TestInboxLeaseExpiry(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, clock := inboxStore(t)
	put, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: "d1"})
	require.NoError(t, err)

	a, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w1", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, put.ID, a.ID)
	assert.EqualValues(t, 1, a.Attempts)
	_, err = ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w2", time.Minute)
	require.ErrorIs(t, err, store.ErrNotFound, "leased")

	_, err = ingest.InboxRenew(ctx, s, a.ID, "w1", time.Minute)
	require.NoError(t, err)
	_, err = ingest.InboxRenew(ctx, s, a.ID, "w2", time.Minute)
	require.ErrorIs(t, err, store.ErrConflict)

	clock.Add((2 * time.Minute).Milliseconds())
	b, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w2", time.Minute)
	require.NoError(t, err, "expired lease is reclaimable")
	assert.Equal(t, "w2", b.Holder)
	assert.EqualValues(t, 2, b.Attempts)

	// The stale holder cannot move the row.
	_, err = ingest.InboxSetState(ctx, s, a.ID, "w1", store.InboxUpdate{State: store.InboxAdmitted})
	require.ErrorIs(t, err, store.ErrConflict)

	c, err := ingest.InboxSetState(ctx, s, b.ID, "w2", store.InboxUpdate{
		State: store.InboxAdmitted, Admitted: json.RawMessage(`{"ok":true}`), HookResults: json.RawMessage(`[{"hook":"h","action":"allow"}]`),
	})

	require.NoError(t, err)
	assert.Equal(t, store.InboxAdmitted, c.State)
	assert.Empty(t, c.Holder, "lease dropped for the next stage")
	assert.JSONEq(t, `{"ok":true}`, string(c.Admitted))

	d, err := ingest.InboxClaim(ctx, s, []string{store.InboxAdmitted}, "dispatch", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, c.ID, d.ID)

	_, err = ingest.InboxSetState(ctx, s, 999, "w2", store.InboxUpdate{State: store.InboxFailed})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestInboxPrune(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, clock := inboxStore(t)
	old := time.UnixMilli(clock.Load()).UTC()
	for _, id := range []string{"done", "denied", "pending", "held"} {
		_, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: id, ReceivedAt: old})
		require.NoError(t, err)
	}
	_, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: "new-done", ReceivedAt: old})
	require.NoError(t, err)
	move := func(state string) {
		it, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w", time.Minute)
		require.NoError(t, err)
		_, err = ingest.InboxSetState(ctx, s, it.ID, "w", store.InboxUpdate{State: state})
		require.NoError(t, err)
	}
	// Claims go oldest first, by received_at then id: done, denied, pending, held, new-done.
	move(store.InboxDispatched)
	move(store.InboxDenied)
	// Leave it pending, still leased so the next claims skip it.
	_, err = ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w", time.Hour)
	require.NoError(t, err)
	move(store.InboxHeld)
	// new-done finishes two days later: prune goes by updated_at.
	clock.Add((48 * time.Hour).Milliseconds())
	move(store.InboxDispatched)

	n, err := ingest.InboxPrune(ctx, s, old.Add(24*time.Hour))
	require.NoError(t, err)
	assert.EqualValues(t, 2, n, "only old terminal rows")
	left, err := ingest.InboxList(ctx, s, store.InboxFilter{})
	require.NoError(t, err)
	var ids []string
	for _, it := range left {
		ids = append(ids, it.DeliveryID)
	}
	assert.ElementsMatch(t, []string{"pending", "held", "new-done"}, ids)
}

// TestInboxPruneBatches: a backlog larger than one delete batch goes in
// one prune call. The batch is 2 here, not 200, so 5 items cross it
// twice: 450 items took 4-7s under -race (JOB-117). Not parallel: it
// sets the package's batch size.
func TestInboxPruneBatches(t *testing.T) {
	t.Cleanup(ingest.SetInboxPruneBatch(2))
	ctx := t.Context()
	s, clock := inboxStore(t)
	const n = 5
	for i := range n {
		_, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: fmt.Sprint(i)})
		require.NoError(t, err)
		it, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w", time.Minute)
		require.NoError(t, err)
		_, err = ingest.InboxSetState(ctx, s, it.ID, "w", store.InboxUpdate{State: store.InboxDispatched})
		require.NoError(t, err)
	}
	clock.Add(time.Hour.Milliseconds())
	got, err := ingest.InboxPrune(ctx, s, time.UnixMilli(clock.Load()))
	require.NoError(t, err)
	assert.EqualValues(t, n, got)
	counts, err := ingest.InboxCounts(ctx, s)
	require.NoError(t, err)
	assert.Zero(t, counts[store.InboxDispatched])
}

// TestInboxRequeue (#353): held -> pending only from the named state and
// only while no one leases the item.
func TestInboxRequeue(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, _ := inboxStore(t)
	put, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: "d1"})
	require.NoError(t, err)

	_, err = ingest.InboxRequeue(ctx, s, put.ID, store.InboxHeld, store.InboxPending)
	require.ErrorIs(t, err, store.ErrConflict, "pending is not held")

	it, err := ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w", time.Minute)
	require.NoError(t, err)
	_, err = ingest.InboxSetState(ctx, s, it.ID, "w", store.InboxUpdate{State: store.InboxHeld})
	require.NoError(t, err)
	got, err := ingest.InboxRequeue(ctx, s, it.ID, store.InboxHeld, store.InboxPending)
	require.NoError(t, err)
	assert.Equal(t, store.InboxPending, got.State)
	assert.Empty(t, got.Holder)

	// A leased item is not moved under its holder.
	_, err = ingest.InboxClaim(ctx, s, []string{store.InboxPending}, "w", time.Minute)
	require.NoError(t, err)
	_, err = ingest.InboxRequeue(ctx, s, it.ID, store.InboxPending, store.InboxHeld)
	require.ErrorIs(t, err, store.ErrConflict)

	_, err = ingest.InboxRequeue(ctx, s, 9999, store.InboxHeld, store.InboxPending)
	require.ErrorIs(t, err, store.ErrNotFound)
	_, err = ingest.InboxRequeue(ctx, s, it.ID, "bogus", store.InboxPending)
	require.ErrorIs(t, err, store.ErrInvalid)
}
