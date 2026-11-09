package storetest

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Race cases for JOB-38. Each one starts goroutines on a barrier so the
// writes overlap, then checks an invariant: one holder per ticket, fences
// only increase, no lost update, no double claim. The clock is never
// advanced while goroutines run: Clock is not safe for concurrent use.

const racers = 16

// race runs f(i) for i in [0,n) after all goroutines are started.
func race(n int, f func(i int)) {
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			<-start
			f(i)
		})
	}
	close(start)
	wg.Wait()
}

// retryConflict retries f while it fails with ErrConflict: a serialization
// failure surfaces as a conflict, and the caller is expected to retry.
func retryConflict(f func() error) error {
	var err error
	for range 50 {
		if err = f(); !errors.Is(err, domain.ErrConflict) {
			return err
		}
	}
	return err
}

// testRaceClaimNextExtend: holders claim and re-claim their own work
// concurrently. ClaimNext returns a holder's in-progress ticket first, so
// every claim of a ticket must name one holder, and its fence never goes
// down.
func testRaceClaimNextExtend(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	const n = 6
	for i := range n {
		task(t, s, fmt.Sprintf("t%d", i), "dev")
	}
	var mu sync.Mutex
	holders := map[string]map[string]bool{}
	fences := map[string][]int64{} // ticket -> fences handed out
	race(racers, func(i int) {
		holder := fmt.Sprintf("h%d", i%n)
		for range 4 {
			tk, c, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: holder, Lease: lease})
			if errors.Is(err, domain.ErrNoWork) || errors.Is(err, domain.ErrConflict) {
				continue
			}
			if !assert.NoError(t, err, holder) {
				return
			}
			assert.Equal(t, holder, c.Holder)
			mu.Lock()
			if holders[tk.ID] == nil {
				holders[tk.ID] = map[string]bool{}
			}
			holders[tk.ID][holder] = true
			fences[tk.ID] = append(fences[tk.ID], c.Fence)
			mu.Unlock()
		}
	})
	for id, hs := range holders {
		assert.Len(t, hs, 1, "one holder for %s", id)
		got, err := s.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusInProgress, got.Status)
		for h := range hs {
			assert.Equal(t, h, got.Holder)
		}
		// Goroutines record claims out of commit order, so only check
		// that the stored fence is not below any fence handed out.
		assert.GreaterOrEqual(t, got.Fence, slices.Max(fences[id]), "fence went down on %s", id)
	}
}

// testRaceReleaseClaimFences: one ticket is released and re-claimed by
// many holders in turn. Every successful claim gets a fence greater than
// every earlier one, and a claim that was released can no longer close.
func testRaceReleaseClaimFences(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	var mu sync.Mutex
	var claims []store.Claim
	race(racers, func(i int) {
		holder := fmt.Sprintf("h%d", i)
		for range 3 {
			_, c, err := s.Claim(ctx, a.ID, holder, lease)
			if err != nil {
				continue
			}
			mu.Lock()
			claims = append(claims, c)
			mu.Unlock()
			_, err = s.Release(ctx, "", a.ID, &c, store.Outcome{Reason: "race"})
			assert.NoError(t, err, "the holder releases its own claim")
		}
	})
	require.NotEmpty(t, claims)
	seen := map[int64]bool{}
	for _, c := range claims {
		assert.False(t, seen[c.Fence], "fence %d handed out twice", c.Fence)
		seen[c.Fence] = true
		_, err := store.Close(ctx, s, c.Holder, a.ID, &c, "stale")
		assert.Error(t, err, "a released claim cannot close")
	}
	_, last, err := s.Claim(ctx, a.ID, "final", lease)
	require.NoError(t, err)
	for _, c := range claims {
		assert.Greater(t, last.Fence, c.Fence)
	}
}

// testRaceReleaseExpire: holders release their own tickets while the
// controller expires the same holders. Each ticket ends open with no
// holder, and every call either succeeds or reports the claim is gone.
func testRaceReleaseExpire(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	const n = 8
	type held struct {
		id string
		c  store.Claim
	}
	var hs []held
	for i := range n {
		tk := task(t, s, fmt.Sprintf("t%d", i), "dev")
		_, c, err := s.Claim(ctx, tk.ID, fmt.Sprintf("h%d", i), lease)
		require.NoError(t, err)
		hs = append(hs, held{tk.ID, c})
	}
	race(2*n, func(i int) {
		h := hs[i%n]
		if i < n {
			_, err := s.Release(ctx, "", h.id, &h.c, store.Outcome{Reason: "done"})
			if err != nil {
				assert.True(t, errors.Is(err, domain.ErrNotHolder) || errors.Is(err, domain.ErrConflict), "release %s: %v", h.id, err)
			}
			return
		}
		err := retryConflict(func() error {
			_, err := s.Expire(ctx, "controller", h.c.Holder, "died")
			return err
		})
		assert.NoError(t, err)
	})
	for _, h := range hs {
		got, err := s.Get(ctx, h.id)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusOpen, got.Status, h.id)
		assert.Empty(t, got.Holder, h.id)
		_, err = s.Heartbeat(ctx, h.c, lease)
		assert.ErrorIs(t, err, domain.ErrNotHolder)
	}
}

// testRaceHeartbeatSweep: leases have expired and holders heartbeat while
// the sweep runs. A ticket the sweep returned is open and its heartbeat
// failed; a ticket it did not return is still held.
func testRaceHeartbeatSweep(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	const n = 8
	var claims []store.Claim
	for i := range n {
		tk := task(t, s, fmt.Sprintf("t%d", i), "dev")
		_, cl, err := s.Claim(ctx, tk.ID, fmt.Sprintf("h%d", i), lease)
		require.NoError(t, err)
		claims = append(claims, cl)
	}
	c.Advance(lease + time.Minute)
	var mu sync.Mutex
	swept := map[string]bool{}
	beat := map[string]bool{}
	race(2*n, func(i int) {
		if i < n {
			cl := claims[i]
			err := retryConflict(func() error {
				_, err := s.Heartbeat(ctx, cl, lease)
				return err
			})
			if err == nil {
				mu.Lock()
				beat[cl.TicketID] = true
				mu.Unlock()
			} else {
				assert.ErrorIs(t, err, domain.ErrNotHolder)
			}
			return
		}
		var ids []string
		err := retryConflict(func() error {
			var err error
			ids, err = s.Expire(ctx, "controller", "", "")
			return err
		})
		if !assert.NoError(t, err) {
			return
		}
		mu.Lock()
		for _, id := range ids {
			assert.False(t, swept[id], "%s swept twice", id)
			swept[id] = true
		}
		mu.Unlock()
	})
	for _, cl := range claims {
		got, err := s.Get(ctx, cl.TicketID)
		require.NoError(t, err)
		if swept[cl.TicketID] {
			assert.Equal(t, domain.StatusOpen, got.Status, cl.TicketID)
			_, err = s.Heartbeat(ctx, cl, lease)
			assert.ErrorIs(t, err, domain.ErrNotHolder, "swept %s cannot heartbeat", cl.TicketID)
		} else {
			assert.True(t, beat[cl.TicketID], "%s neither swept nor renewed", cl.TicketID)
			assert.Equal(t, domain.StatusInProgress, got.Status, cl.TicketID)
			assert.Equal(t, cl.Holder, got.Holder)
		}
	}
}

// testRaceUpdateClaim: label updates race with claims of the same ticket.
// No update is lost (every label lands) and exactly one claim wins.
func testRaceUpdateClaim(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	var mu sync.Mutex
	var winners []string
	race(racers, func(i int) {
		if i%2 == 0 {
			err := retryConflict(func() error {
				return store.AddLabel(ctx, s, "test", a.ID, fmt.Sprintf("l%d", i))
			})
			assert.NoError(t, err)
			return
		}
		holder := fmt.Sprintf("h%d", i)
		if _, _, err := s.Claim(ctx, a.ID, holder, lease); err == nil {
			mu.Lock()
			winners = append(winners, holder)
			mu.Unlock()
		}
	})
	assert.Len(t, winners, 1, "one claim wins")
	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	for i := 0; i < racers; i += 2 {
		assert.Contains(t, got.Labels, fmt.Sprintf("l%d", i), "label update lost")
	}
	if len(winners) == 1 {
		assert.Equal(t, winners[0], got.Holder)
	}
}

// testRaceCASUpdates: compare-and-swap updates from one revision race;
// exactly one wins and the revision moves by one.
func testRaceCASUpdates(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	var mu sync.Mutex
	var won []string
	race(racers, func(i int) {
		title := fmt.Sprintf("t%d", i)
		_, err := s.Update(ctx, "test", a.ID, store.Update{Title: &title, IfRevision: a.Revision})
		if err == nil {
			mu.Lock()
			won = append(won, title)
			mu.Unlock()
			return
		}
		assert.ErrorIs(t, err, domain.ErrConflict)
	})
	require.Len(t, won, 1)
	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, won[0], got.Title)
	assert.Equal(t, a.Revision+1, got.Revision)
}

// testRaceSlotRelease: slot claims race with ReleaseSlots and closes. A
// slot is never held by two tickets, and once every claim has ended and
// ReleaseSlots ran, no slot is still held.
func testRaceSlotRelease(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	require.NoError(t, setProviders(ctx, s, []string{"emu"}))
	var steps []string
	for i := range 8 {
		steps = append(steps, fmt.Sprintf("js-q%d", i))
	}
	slotSteps(t, s, "emu", steps...)
	for _, id := range []string{"e1", "e2"} {
		_, err := addSlot(ctx, s, "emu", id, nil)
		require.NoError(t, err)
	}
	race(racers, func(i int) {
		if i%4 == 0 {
			_, err := s.ReleaseSlots(ctx)
			if err != nil {
				assert.ErrorIs(t, err, domain.ErrConflict)
			}
			return
		}
		holder := fmt.Sprintf("h%d", i)
		tk, c, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: holder, Lease: lease})
		if err != nil {
			return
		}
		all, err := allSlots(ctx, s)
		if assert.NoError(t, err) {
			n := 0
			for _, sl := range all {
				if sl.State == domain.SlotHeld && sl.Ticket == tk.ID {
					n++
				}
			}
			assert.Equal(t, 1, n, "%s holds one slot", tk.ID)
		}
		_, err = store.Close(ctx, s, holder, tk.ID, &c, "done")
		assert.NoError(t, err)
	})
	_, err := s.ReleaseSlots(ctx)
	require.NoError(t, err)
	all, err := allSlots(ctx, s)
	require.NoError(t, err)
	for _, sl := range all {
		assert.NotEqual(t, domain.SlotHeld, sl.State, "slot %s held after every claim ended", sl.ID)
	}
}

// testReadAfterWrite: after a write returns, Get and List see it at once,
// from any goroutine (SQLite reads go through a reader pool). Writes
// retry on ErrConflict: SQLite reports SQLITE_BUSY as a conflict.
func testReadAfterWrite(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	race(racers, func(i int) {
		tk, err := store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindTask, Title: fmt.Sprintf("raw%d", i), Pool: "raw"})
		if !assert.NoError(t, err) {
			return
		}
		got, err := s.Get(ctx, tk.ID)
		if assert.NoError(t, err, "get after create") {
			assert.Equal(t, tk.Revision, got.Revision)
		}
		assert.True(t, listed(t, s, store.Filter{Project: proj, Pool: "raw"}, tk.ID), "list after create")

		title := fmt.Sprintf("raw%d-b", i)
		var u domain.Ticket
		err = retryConflict(func() error {
			u, err = s.Update(ctx, "test", tk.ID, store.Update{Title: &title, AddLabels: []string{"x"}})
			return err
		})
		if !assert.NoError(t, err) {
			return
		}
		got, err = s.Get(ctx, tk.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, title, got.Title, "get after update")
			assert.Equal(t, u.Revision, got.Revision)
		}
		assert.True(t, listed(t, s, store.Filter{Project: proj, Label: "x"}, tk.ID), "list after label")

		holder := fmt.Sprintf("h%d", i)
		var c store.Claim
		err = retryConflict(func() error {
			_, c, err = s.Claim(ctx, tk.ID, holder, lease)
			return err
		})
		if !assert.NoError(t, err) {
			return
		}
		got, err = s.Get(ctx, tk.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, domain.StatusInProgress, got.Status, "get after claim")
			assert.Equal(t, holder, got.Holder)
		}
		assert.True(t, listed(t, s, store.Filter{Project: proj, Holder: holder}, tk.ID), "list after claim")

		err = retryConflict(func() error {
			_, err := store.Close(ctx, s, holder, tk.ID, &c, "done")
			return err
		})
		if !assert.NoError(t, err) {
			return
		}
		got, err = s.Get(ctx, tk.ID)
		if assert.NoError(t, err) {
			assert.Equal(t, domain.StatusClosed, got.Status, "get after close")
		}
		assert.True(t, listed(t, s, store.Filter{Project: proj, Status: domain.StatusClosed, Pool: "raw"}, tk.ID), "list after close")
	})
}

func listed(t *testing.T, s store.Store, f store.Filter, id string) bool {
	t.Helper()
	ts, err := s.List(t.Context(), f)
	if !assert.NoError(t, err) {
		return false
	}
	return slices.Contains(ids(ts), id)
}
