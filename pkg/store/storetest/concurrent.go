package storetest

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
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
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: blocked.ID, To: blocker.ID, Kind: domain.EdgeBlocks}))

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
				if errors.Is(err, domain.ErrNoWork) {
					return
				}
				if !assert.NoError(t, err, holder) {
					return
				}
				mu.Lock()
				got[tk.ID] = append(got[tk.ID], holder)
				mu.Unlock()
				if _, err := store.Close(ctx, s, holder, tk.ID, &c, "done"); !assert.NoError(t, err) {
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

// testConcurrentResourceClaims: several holders race for steps that share
// a capacity-1 resource (#400, #406). The slot check is part of the
// guarded claim write, so exactly one step is in progress at a time.
func testConcurrentResourceClaims(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	require.NoError(t, setCaps(ctx, s, map[string]int{"emu": 1}))
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	nts := []store.NewTicket{{
		ID: runID, Project: proj, Kind: domain.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","state":"running"}`),
	}}
	for i := range 6 {
		id := fmt.Sprintf("js-r%d", i)
		nts = append(nts, store.NewTicket{
			ID: id, Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":%q,"retries":5,"resources":["emu"]}`, runID, id)),
		})
	}
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: nts, Edges: nil})
	require.NoError(t, err)

	var mu sync.Mutex
	var held []string
	var wg sync.WaitGroup
	for i := range 8 {
		holder := fmt.Sprintf("h%d", i)
		wg.Go(func() {
			tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: holder, Lease: lease})
			if errors.Is(err, domain.ErrNoWork) || errors.Is(err, domain.ErrConflict) {
				return
			}
			if assert.NoError(t, err, holder) {
				mu.Lock()
				held = append(held, tk.ID)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	assert.Len(t, held, 1, "one slot, one claim")
}

// testClaimAfterSkippedBlocker: a step whose blocker closed skipped is
// ready, and the guarded claim write accepts it too (#389, #406).
func testClaimAfterSkippedBlocker(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	mk := func(id string) store.NewTicket {
		return store.NewTicket{
			ID: id, Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":%q}`, runID, id)),
		}
	}
	run := store.NewTicket{
		ID: runID, Project: proj, Kind: domain.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","state":"running"}`),
	}
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, mk("js-sa"), mk("js-sb")}, Edges: []domain.Edge{{From: "js-sb", To: "js-sa", Kind: domain.EdgeBlocks}}})

	require.NoError(t, err)
	_, err = store.Close(ctx, s, "controller", "js-sa", nil, domain.StepSkipped)
	require.NoError(t, err)

	tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "h", Lease: lease})
	require.NoError(t, err)
	assert.Equal(t, "js-sb", tk.ID)
}
