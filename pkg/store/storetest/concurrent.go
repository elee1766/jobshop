package storetest

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// testConcurrentClaims runs many claimers against one store (#406): each
// ticket ends with exactly one holder, nobody claims a blocked ticket,
// and a claimer that loses a race moves on to the next ticket instead of
// failing. It must hold without a single-writer database.
func testConcurrentClaims(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	const tickets, claimers = 12, 8
	free := map[string]bool{}
	for i := range tickets {
		free[task(t, s, fmt.Sprintf("t%d", i), "dev").ID] = true
	}
	blocker := task(t, s, "blocker", "other")
	blocked := task(t, s, "blocked", "dev")
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: blocked.ID, To: blocker.ID, Kind: store.EdgeBlocks}))

	var mu sync.Mutex
	got := map[string][]string{} // ticket -> holders that claimed it
	var wg sync.WaitGroup
	for i := range claimers {
		holder := fmt.Sprintf("s%d", i)
		wg.Go(func() {
			// Each claimer takes work until none is left. A holder's own
			// in-progress ticket comes back first, so close it each time.
			for {
				tk, c, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: holder, Lease: lease})
				if errors.Is(err, store.ErrNoWork) {
					return
				}
				if !assert.NoError(t, err, holder) {
					return
				}
				mu.Lock()
				got[tk.ID] = append(got[tk.ID], holder)
				mu.Unlock()
				if _, err := s.CloseTicket(ctx, holder, tk.ID, &c, "done"); !assert.NoError(t, err) {
					return
				}
			}
		})
	}
	wg.Wait()
	assert.Empty(t, got[blocked.ID], "a blocked ticket is never claimed")
	for id := range free {
		assert.Len(t, got[id], 1, "exactly one holder for %s", id)
	}
	assert.Len(t, got, tickets)

	// Direct claims of one ticket race too: exactly one wins.
	one := task(t, s, "one", "dev")
	wins := make(chan string, claimers)
	for i := range claimers {
		holder := fmt.Sprintf("d%d", i)
		wg.Go(func() {
			if _, _, err := s.Claim(ctx, one.ID, holder, lease); err == nil {
				wins <- holder
			}
		})
	}
	wg.Wait()
	close(wins)
	var winners []string
	for w := range wins {
		winners = append(winners, w)
	}
	assert.Len(t, winners, 1)
}
