package control

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/ingestqueue"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/store"
)

// admitComments writes n comments on the followed thread to the inbox
// and admits them, without dispatching.
func admitComments(t *testing.T, c *Controller, clk *scanClock, src *fakeThreads, n int) []string {
	t.Helper()
	th := thread{Repo: "acme/widget", Number: 7}
	var bodies []string
	for i := range n {
		body := fmt.Sprintf("note-%03d", i)
		cm := comment(clk, fmt.Sprintf("c%d", i), body)
		src.add("acme/widget#7", cm)
		require.NoError(t, c.inboxComment(t.Context(), th, cm))
		bodies = append(bodies, body)
	}
	for {
		ok, err := c.admitNext(t.Context())
		require.NoError(t, err)
		if !ok {
			return bodies
		}
	}
}

// deliveredCounts counts how often each body reached the session.
func deliveredCounts(t *testing.T, c *Controller, to string, bodies []string) map[string]int {
	t.Helper()
	got := map[string]int{}
	for _, m := range messagesTo(t, c, to) {
		for _, b := range bodies {
			got[b] += strings.Count(m.Body, b)
		}
	}
	return got
}

func setDispatch(t *testing.T, c *Controller, mode string) {
	t.Helper()
	cfg := *c.config()
	cfg.Ingest.Dispatch = mode
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
}

// TestDispatchRacingDispatchers: the controller and two workers dispatch
// the same admitted rows at once; every row is delivered exactly once.
func TestDispatchRacingDispatchers(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t)
	bodies := admitComments(t, c, clk, src, 30)

	var wg sync.WaitGroup
	errs := make([]error, 3)
	wg.Go(func() { _, errs[0] = c.dispatchAll(t.Context(), dispatchHolder) })
	wg.Go(func() { _, errs[1] = worker(c, "w1").DispatchOnce(t.Context()) })
	wg.Go(func() { _, errs[2] = worker(c, "w2").DispatchOnce(t.Context()) })
	wg.Wait()
	for _, err := range errs {
		require.NoError(t, err)
	}
	for _, b := range bodies {
		assert.Equal(t, 1, deliveredCounts(t, c, sess.ID, bodies)[b], b)
	}
	for _, st := range inboxStates(t, c) {
		assert.Equal(t, store.InboxDispatched, st)
	}
}

// TestDispatchWorkerOnlyWhenExternal: with ingest.dispatch external the
// controller's pass leaves admitted rows to `ingest work --dispatch`.
func TestDispatchWorkerOnlyWhenExternal(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t)
	setDispatch(t, c, config.IngestWorkersExternal)
	bodies := admitComments(t, c, clk, src, 2)
	c.ingestPass(t.Context())
	assert.Empty(t, messagesTo(t, c, sess.ID))
	assert.Equal(t, []string{store.InboxAdmitted, store.InboxAdmitted}, inboxStates(t, c))

	w := worker(c, "w1")
	w.Dispatch, w.NoAdmit = true, true
	_, d, err := w.Pass(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 2, d)
	for _, b := range bodies {
		assert.Equal(t, 1, deliveredCounts(t, c, sess.ID, bodies)[b], b)
	}
	require.Len(t, c.queue().Workers(), 1, "a dispatch call is a heartbeat")
}

// TestDispatchWorkerDiesMidLease: a worker that claims admitted rows and
// dies leaves them leased; after the lease another worker delivers them,
// once, and the dead worker cannot mark them.
func TestDispatchWorkerDiesMidLease(t *testing.T) {
	t.Parallel()
	c, clk, src, sess := dispatchController(t)
	setDispatch(t, c, config.IngestWorkersExternal)
	bodies := admitComments(t, c, clk, src, 1)
	dead, err := ingest.InboxClaim(t.Context(), c.store, []string{store.InboxAdmitted}, ingestqueue.WorkerPrefix+"dead", dispatchLease)
	require.NoError(t, err)

	n, err := worker(c, "w2").DispatchOnce(t.Context())
	require.NoError(t, err)
	assert.Zero(t, n, "leased by the dead worker")
	assert.Empty(t, messagesTo(t, c, sess.ID))

	clk.advance(dispatchLease + time.Second)
	n, err = worker(c, "w2").DispatchOnce(t.Context())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, deliveredCounts(t, c, sess.ID, bodies)[bodies[0]])
	assert.Equal(t, []string{store.InboxDispatched}, inboxStates(t, c))

	_, err = ingest.InboxSetState(t.Context(), c.store, dead.ID, ingestqueue.WorkerPrefix+"dead", store.InboxUpdate{State: store.InboxDispatched})
	require.ErrorIs(t, err, store.ErrConflict)
}
