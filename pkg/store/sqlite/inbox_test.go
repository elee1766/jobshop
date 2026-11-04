package sqlite_test

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// inboxStore opens a store whose clock the test moves.
func inboxStore(t *testing.T) (*sqlite.Store, *atomic.Int64) {
	t.Helper()
	var clock atomic.Int64
	clock.Store(time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC).UnixMilli())
	s := openTest(t, sqlite.Options{
		Now: func() time.Time { return time.UnixMilli(clock.Load()).UTC() },
	})
	t.Cleanup(func() { _ = s.Close() })
	return s, &clock
}

func TestInboxPutDedupe(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, _ := inboxStore(t)
	it := store.InboxItem{Source: "github", DeliveryID: "d1", Endpoint: "gh-app", Kind: "issue_comment", RepoOrTeam: "a/b", Number: 7, Payload: json.RawMessage(`{"x":1}`)}
	a, dup, err := s.InboxPut(ctx, it)
	require.NoError(t, err)
	assert.Equal(t, "gh-app", a.Endpoint, "endpoint stored (#339)")
	assert.False(t, dup)
	assert.Equal(t, store.InboxPending, a.State)

	it.Payload = json.RawMessage(`{"x":2}`)
	b, dup, err := s.InboxPut(ctx, it)
	require.NoError(t, err)
	assert.True(t, dup, "same delivery id")
	assert.Equal(t, a.ID, b.ID)
	assert.JSONEq(t, `{"x":1}`, string(b.Payload), "first payload kept")

	_, _, err = s.InboxPut(ctx, store.InboxItem{Source: "jira", DeliveryID: "d2"})
	require.ErrorIs(t, err, store.ErrInvalid)
}

func TestInboxClaimConcurrent(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, _ := inboxStore(t)
	const n = 20
	for i := range n {
		_, _, err := s.InboxPut(ctx, store.InboxItem{Source: "linear", DeliveryID: string(rune('a' + i))})
		require.NoError(t, err)
	}
	var mu sync.Mutex
	got := map[int64]string{}
	var wg sync.WaitGroup
	for w := range 8 {
		holder := string(rune('A' + w))
		wg.Go(func() {
			for {
				it, err := s.InboxClaim(ctx, []string{store.InboxPending}, holder, time.Minute)
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
	put, _, err := s.InboxPut(ctx, store.InboxItem{Source: "github", DeliveryID: "d1"})
	require.NoError(t, err)

	a, err := s.InboxClaim(ctx, []string{store.InboxPending}, "w1", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, put.ID, a.ID)
	assert.EqualValues(t, 1, a.Attempts)
	_, err = s.InboxClaim(ctx, []string{store.InboxPending}, "w2", time.Minute)
	require.ErrorIs(t, err, store.ErrNotFound, "leased")

	_, err = s.InboxRenew(ctx, a.ID, "w1", time.Minute)
	require.NoError(t, err)
	_, err = s.InboxRenew(ctx, a.ID, "w2", time.Minute)
	require.ErrorIs(t, err, store.ErrConflict)

	clock.Add((2 * time.Minute).Milliseconds())
	b, err := s.InboxClaim(ctx, []string{store.InboxPending}, "w2", time.Minute)
	require.NoError(t, err, "expired lease is reclaimable")
	assert.Equal(t, "w2", b.Holder)
	assert.EqualValues(t, 2, b.Attempts)

	// The stale holder cannot move the row.
	_, err = s.InboxSetState(ctx, a.ID, "w1", store.InboxUpdate{State: store.InboxAdmitted})
	require.ErrorIs(t, err, store.ErrConflict)

	c, err := s.InboxSetState(ctx, b.ID, "w2", store.InboxUpdate{
		State: store.InboxAdmitted, Admitted: json.RawMessage(`{"ok":true}`), HookResults: json.RawMessage(`[{"hook":"h","action":"allow"}]`),
	})
	require.NoError(t, err)
	assert.Equal(t, store.InboxAdmitted, c.State)
	assert.Empty(t, c.Holder, "lease dropped for the next stage")
	assert.JSONEq(t, `{"ok":true}`, string(c.Admitted))

	d, err := s.InboxClaim(ctx, []string{store.InboxAdmitted}, "dispatch", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, c.ID, d.ID)

	_, err = s.InboxSetState(ctx, 999, "w2", store.InboxUpdate{State: store.InboxFailed})
	require.ErrorIs(t, err, store.ErrNotFound)
}

func TestInboxPrune(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, clock := inboxStore(t)
	old := time.UnixMilli(clock.Load()).UTC()
	for _, id := range []string{"done", "denied", "pending", "held"} {
		_, _, err := s.InboxPut(ctx, store.InboxItem{Source: "github", DeliveryID: id, ReceivedAt: old})
		require.NoError(t, err)
	}
	_, _, err := s.InboxPut(ctx, store.InboxItem{Source: "github", DeliveryID: "new-done", ReceivedAt: old.Add(48 * time.Hour)})
	require.NoError(t, err)
	move := func(state string) {
		it, err := s.InboxClaim(ctx, []string{store.InboxPending}, "w", time.Minute)
		require.NoError(t, err)
		_, err = s.InboxSetState(ctx, it.ID, "w", store.InboxUpdate{State: state})
		require.NoError(t, err)
	}
	// Claims go oldest first, by received_at then id: done, denied, pending, held, new-done.
	move(store.InboxDispatched)
	move(store.InboxDenied)
	// Leave it pending, still leased so the next claims skip it.
	_, err = s.InboxClaim(ctx, []string{store.InboxPending}, "w", time.Hour)
	require.NoError(t, err)
	move(store.InboxHeld)
	move(store.InboxDispatched)

	n, err := s.InboxPrune(ctx, old.Add(24*time.Hour))
	require.NoError(t, err)
	assert.EqualValues(t, 2, n, "only old terminal rows")
	left, err := s.InboxList(ctx, store.InboxFilter{})
	require.NoError(t, err)
	var ids []string
	for _, it := range left {
		ids = append(ids, it.DeliveryID)
	}
	assert.ElementsMatch(t, []string{"pending", "held", "new-done"}, ids)
}
