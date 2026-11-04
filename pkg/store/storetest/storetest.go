// Package storetest is the conformance suite every store backend must pass.
package storetest

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// Clock is a settable test clock shared with the backend under test.
type Clock struct{ T time.Time }

// Now returns the current fake time.
func (c *Clock) Now() time.Time { return c.T }

// Advance moves the clock forward.
func (c *Clock) Advance(d time.Duration) { c.T = c.T.Add(d) }

// Factory returns a fresh, empty store that reads time from clock.
type Factory func(t *testing.T, clock *Clock) store.Store

const (
	proj  = "js"
	lease = 30 * time.Minute
)

// Run executes the suite against stores built by newStore.
func Run(t *testing.T, newStore Factory) {
	cases := []struct {
		name string
		fn   func(t *testing.T, s store.Store, c *Clock)
	}{
		{"CreateGet", testCreateGet},
		{"FieldsValidated", testFieldsValidated},
		{"UpdateCAS", testUpdateCAS},
		{"ReadyRespectsEdges", testReadyRespectsEdges},
		{"EdgeCycle", testEdgeCycle},
		{"ClaimNextTiers", testClaimNextTiers},
		{"FencedClose", testFencedClose},
		{"Heartbeat", testHeartbeat},
		{"ReleaseHolder", testReleaseHolder},
		{"SweepLeases", testSweepLeases},
		{"DeferUntil", testDeferUntil},
		{"LabelsCommentsHistory", testLabelsCommentsHistory},
		{"CreateGraphAtomic", testCreateGraphAtomic},
		{"SessionKeyUnique", testSessionKeyUnique},
		{"ListMetaNewest", testListMetaNewest},
		{"ListCreatedBy", testListCreatedBy},
		{"Orphaned", testOrphaned},
		{"Runs", testRuns},
		{"StepFailures", testStepFailures},
		{"ApplyGraph", testApplyGraph},
		{"Resources", testResources},
		{"PruneWriteKeys", testPruneWriteKeys},
		{"ConcurrentClaims", testConcurrentClaims},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Each case has its own store and clock, so cases run in parallel.
			t.Parallel()
			c := &Clock{T: time.Date(2026, 11, 1, 12, 0, 0, 0, time.UTC)}
			s := newStore(t, c)
			t.Cleanup(func() { _ = s.Close() })
			tc.fn(t, s, c)
		})
	}
}

func task(t *testing.T, s store.Store, title, pool string, opts ...func(*store.NewTicket)) store.Ticket {
	t.Helper()
	nt := store.NewTicket{Project: proj, Kind: store.KindTask, Title: title, Pool: pool}
	for _, o := range opts {
		o(&nt)
	}
	tk, err := s.Create(t.Context(), "test", nt)
	require.NoError(t, err)
	return tk
}

func prio(p int) func(*store.NewTicket) { return func(n *store.NewTicket) { n.Priority = &p } }

func ids(ts []store.Ticket) []string {
	out := make([]string, len(ts))
	for i, t := range ts {
		out[i] = t.ID
	}
	return out
}

func testListMetaNewest(t *testing.T, s store.Store, c *Clock) {
	withMeta := func(m string) func(*store.NewTicket) {
		return func(n *store.NewTicket) { n.Meta = json.RawMessage(m) }
	}
	a := task(t, s, "a", "p", withMeta(`{"trigger":{"name":"x"}}`))
	c.Advance(time.Second)
	task(t, s, "b", "p", withMeta(`{"trigger":{"name":"y"}}`))
	c.Advance(time.Second)
	b := task(t, s, "c", "p", withMeta(`{"trigger":{"name":"x"}}`), prio(4))
	task(t, s, "d", "p")

	l, err := s.List(t.Context(), store.Filter{MetaPack: "trigger", MetaKey: "name", MetaValue: "x"})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID, b.ID}, ids(l), "priority order by default")

	l, err = s.List(t.Context(), store.Filter{MetaPack: "trigger", MetaKey: "name", MetaValue: "x", Newest: true, Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID}, ids(l), "newest first")

	// Names are quoted into the JSON path, not interpreted.
	l, err = s.List(t.Context(), store.Filter{MetaPack: `trigger"."name`, MetaKey: "name", MetaValue: "x"})
	require.NoError(t, err)
	assert.Empty(t, l)
}

func testListCreatedBy(t *testing.T, s store.Store, _ *Clock) {
	a, err := s.Create(t.Context(), "js-sess1", store.NewTicket{Project: "js", Kind: store.KindTask, Title: "a"})
	require.NoError(t, err)
	_, err = s.Create(t.Context(), "admin", store.NewTicket{Project: "js", Kind: store.KindTask, Title: "b"})
	require.NoError(t, err)
	l, err := s.List(t.Context(), store.Filter{CreatedBy: "js-sess1"})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(l))
}

func testCreateGet(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev", func(n *store.NewTicket) {
		n.Labels = []string{"x", "y"}
		n.Meta = json.RawMessage(`{"gx":{"linear_id":"ENG-1"}}`)
	})
	assert.Regexp(t, `^js-[0-9a-z]{13}$`, a.ID)
	assert.Equal(t, store.StatusOpen, a.Status)
	assert.Equal(t, store.DefaultPriority, a.Priority)
	assert.EqualValues(t, 1, a.Revision)

	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"x", "y"}, got.Labels)
	assert.JSONEq(t, `{"gx":{"linear_id":"ENG-1"}}`, string(got.Meta))
	assert.Equal(t, "test", got.CreatedBy)

	_, err = s.Get(ctx, "js-missing")
	assert.ErrorIs(t, err, store.ErrNotFound)

	_, err = s.Create(ctx, "test", store.NewTicket{ID: a.ID, Project: proj, Kind: store.KindTask})
	assert.ErrorIs(t, err, store.ErrAlreadyExist)

	_, err = s.Create(ctx, "test", store.NewTicket{Project: proj, Kind: store.KindTask, ParentID: "js-missing"})
	assert.ErrorIs(t, err, store.ErrNotFound)

	l, err := s.List(ctx, store.Filter{Project: proj, Label: "x"})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(l))
}

func testFieldsValidated(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	cases := []store.NewTicket{
		{Project: proj, Kind: "nope"},
		{Project: proj, Kind: store.KindMessage, Fields: json.RawMessage(`{}`)},
		{Project: proj, Kind: store.KindMessage, Fields: json.RawMessage(`{"to":"pool:"}`)},
		{Project: proj, Kind: store.KindSession, Fields: json.RawMessage(`{"agent":"a","extra":1}`)},
		{Project: proj, Kind: store.KindTask, Meta: json.RawMessage(`{"flat":"string"}`)},
		{Kind: store.KindTask},
	}
	for _, nt := range cases {
		_, err := s.Create(ctx, "test", nt)
		assert.ErrorIs(t, err, store.ErrInvalid, "%+v", nt)
	}
	m, err := s.Create(ctx, "test", store.NewTicket{
		Project: proj, Kind: store.KindMessage,
		Fields: json.RawMessage(`{"to":"pool:dev","from":"me"}`),
	})
	require.NoError(t, err)
	f, err := store.DecodeFields[store.MessageFields](m)
	require.NoError(t, err)
	assert.Equal(t, store.MessageFields{To: "pool:dev", From: "me"}, f)
}

func testUpdateCAS(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	title := "b"
	u, err := s.Update(ctx, "test", a.ID, store.Update{Title: &title, IfRevision: a.Revision})
	require.NoError(t, err)
	assert.Equal(t, "b", u.Title)
	assert.Equal(t, a.Revision+1, u.Revision)

	_, err = s.Update(ctx, "test", a.ID, store.Update{Title: &title, IfRevision: a.Revision})
	assert.ErrorIs(t, err, store.ErrConflict)

	_, err = s.Update(ctx, "test", a.ID, store.Update{Fields: json.RawMessage(`{"bad":1}`)})
	assert.ErrorIs(t, err, store.ErrInvalid)
}

func testReadyRespectsEdges(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev", prio(1))
	b := task(t, s, "b", "dev", prio(2))
	c := task(t, s, "c", "dev", prio(3))
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: a.ID, To: b.ID, Kind: store.EdgeBlocks}))
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: c.ID, To: b.ID, Kind: store.EdgeTracks}))

	r, err := s.Ready(ctx, proj, "", 0)
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID, c.ID}, ids(r), "a is blocked; tracks does not block")

	_, err = s.CloseTicket(ctx, "test", b.ID, nil, "done")
	require.NoError(t, err)
	r, err = s.Ready(ctx, proj, "dev", 0)
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID, c.ID}, ids(r))

	e, err := s.Edges(ctx, b.ID)
	require.NoError(t, err)
	assert.Len(t, e, 2)
	require.NoError(t, s.RemoveEdge(ctx, "test", store.Edge{From: c.ID, To: b.ID, Kind: store.EdgeTracks}))
	assert.ErrorIs(t, s.RemoveEdge(ctx, "test", store.Edge{From: c.ID, To: b.ID, Kind: store.EdgeTracks}), store.ErrNotFound)
}

func testEdgeCycle(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a, b, c := task(t, s, "a", ""), task(t, s, "b", ""), task(t, s, "c", "")
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: a.ID, To: b.ID, Kind: store.EdgeBlocks}))
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: b.ID, To: c.ID, Kind: store.EdgeWaitsFor}))
	assert.ErrorIs(t, s.AddEdge(ctx, "test", store.Edge{From: c.ID, To: a.ID, Kind: store.EdgeBlocks}), store.ErrCycle)
	assert.ErrorIs(t, s.AddEdge(ctx, "test", store.Edge{From: a.ID, To: a.ID, Kind: store.EdgeBlocks}), store.ErrInvalid)
	// tracks never gates readiness, so it may close a loop.
	require.NoError(t, s.AddEdge(ctx, "test", store.Edge{From: c.ID, To: a.ID, Kind: store.EdgeTracks}))
	assert.ErrorIs(t, s.AddEdge(ctx, "test", store.Edge{From: a.ID, To: "js-missing", Kind: store.EdgeBlocks}), store.ErrNotFound)
}

func testOrphaned(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	none := task(t, s, "none", "")
	gone := task(t, s, "gone", "gone")
	_ = task(t, s, "queued", "dev")
	_ = task(t, s, "held", "gone", func(n *store.NewTicket) { n.Holder = "s1" })
	closed := task(t, s, "closed", "")
	_, err := s.CloseTicket(ctx, "test", closed.ID, nil, "done")
	require.NoError(t, err)

	got, err := s.Orphaned(ctx, "", []string{"dev"}, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{none.ID, gone.ID}, ids(got))

	got, err = s.Orphaned(ctx, "other", []string{"dev"}, 0)
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = s.Orphaned(ctx, proj, nil, 0)
	require.NoError(t, err)
	assert.Len(t, got, 3, "no known pools: every pool is unknown")
}

func testRuns(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(runID, proj+"-"))
	run := store.NewTicket{
		ID: runID, Project: proj, Kind: store.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","key":"k1","state":"running"}`),
	}
	step := func(i int, id string) store.NewTicket {
		return store.NewTicket{
			ID: id, Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"s%d","index":%d,"retries":2}`, runID, i, i)),
		}
	}
	// Inserted out of order: ListChildren orders by fields.index (10 > 2).
	_, err = s.CreateGraph(ctx, "test", []store.NewTicket{run, step(10, "js-b"), step(2, "js-a")}, nil)
	require.NoError(t, err)
	kids, err := s.ListChildren(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, []string{"js-a", "js-b"}, ids(kids))

	got, err := s.FindOpenRunByKey(ctx, "w", "k1")
	require.NoError(t, err)
	assert.Equal(t, runID, got.ID)
	_, err = s.FindOpenRunByKey(ctx, "other", "k1")
	assert.ErrorIs(t, err, store.ErrNotFound)

	r, err := s.ListRuns(ctx, proj, store.RunRunning, 0)
	require.NoError(t, err)
	assert.Equal(t, []string{runID}, ids(r))
	r, err = s.ListRuns(ctx, "", store.RunNeedsHuman, 0)
	require.NoError(t, err)
	assert.Empty(t, r)

	// Steps are claimable; runs are not.
	ready, err := s.Ready(ctx, proj, "", 0)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"js-a", "js-b"}, ids(ready))

	_, err = s.CloseTicket(ctx, "test", runID, nil, "done")
	require.NoError(t, err)
	_, err = s.FindOpenRunByKey(ctx, "w", "k1")
	assert.ErrorIs(t, err, store.ErrNotFound, "closed runs do not dedup")
}

// testStepFailures: each failure source bumps fields.failures in its own
// write; below retries the step reopens, at the limit it closes failed and
// keeps the next step blocked.
func testStepFailures(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	mk := func(id string, i int) store.NewTicket {
		return store.NewTicket{
			ID: id, Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: "js-run",
			Fields: json.RawMessage(fmt.Sprintf(`{"run":"js-run","step":"s%d","index":%d,"retries":4}`, i, i)),
		}
	}
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: store.KindRun, Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)}
	_, err := s.CreateGraph(ctx, "test", []store.NewTicket{run, mk("js-a", 0), mk("js-b", 1)},
		[]store.Edge{{From: "js-b", To: "js-a", Kind: store.EdgeBlocks}})
	require.NoError(t, err)
	fails := func() int {
		t.Helper()
		tk, err := s.Get(ctx, "js-a")
		require.NoError(t, err)
		f, err := store.DecodeFields[store.StepFields](tk)
		require.NoError(t, err)
		return f.Failures
	}

	// Close failed reopens below the limit.
	tk, err := s.CloseTicket(ctx, "test", "js-a", nil, "failed")
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, tk.Status)
	assert.Equal(t, 1, fails())

	// A rebase close reopens without counting, however often (#424).
	for range 5 {
		tk, err = s.CloseTicket(ctx, "test", "js-a", nil, store.ResolutionRebase)
		require.NoError(t, err)
		assert.Equal(t, store.StatusOpen, tk.Status)
		assert.Empty(t, tk.Resolution)
	}
	assert.Equal(t, 1, fails())

	// Release.
	_, cl, err := s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, cl, "gave up", time.Time{})
	require.NoError(t, err)
	assert.Equal(t, 2, fails())

	// A deferred release is not a failure and hides the step until then (#139).
	_, cl, err = s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)
	tk, err = s.Release(ctx, cl, "not yet", c.T.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, tk.Status)
	assert.Equal(t, 2, fails())
	r, err := s.Ready(ctx, proj, "dev", 0)
	require.NoError(t, err)
	assert.NotContains(t, ids(r), "js-a")
	c.Advance(time.Hour)
	r, err = s.Ready(ctx, proj, "dev", 0)
	require.NoError(t, err)
	assert.Contains(t, ids(r), "js-a")

	// Dead holder.
	_, _, err = s.Claim(ctx, "js-a", "s2", lease)
	require.NoError(t, err)
	_, err = s.ReleaseHolder(ctx, "controller", "s2", "session died")
	require.NoError(t, err)
	assert.Equal(t, 3, fails())

	// Lease expiry reaches the limit: closed failed.
	_, _, err = s.Claim(ctx, "js-a", "s3", lease)
	require.NoError(t, err)
	c.Advance(2 * lease)
	_, err = s.SweepLeases(ctx, "controller")
	require.NoError(t, err)
	assert.Equal(t, 4, fails())
	tk, err = s.Get(ctx, "js-a")
	require.NoError(t, err)
	assert.Equal(t, store.StatusClosed, tk.Status)
	assert.Equal(t, "failed", tk.Resolution)
	cs, err := s.Comments(ctx, "js-a")
	require.NoError(t, err)
	require.Len(t, cs, 10, "one comment per counted failure, per rebase close, plus the deferral note")
	assert.Contains(t, cs[1].Body, "Closed for a rebase")
	assert.Contains(t, cs[7].Body, "Released until")
	assert.Contains(t, cs[9].Body, "Retry limit 4 reached")

	// A step closed failed keeps the next step blocked.
	ready, err := s.Ready(ctx, proj, "", 0)
	require.NoError(t, err)
	assert.Empty(t, ids(ready))
	_, _, err = s.Claim(ctx, "js-b", "s4", lease)
	assert.ErrorIs(t, err, store.ErrNoWork)

	// Tasks are not counted.
	task := task(t, s, "plain", "dev")
	tk, err = s.CloseTicket(ctx, "test", task.ID, nil, "failed")
	require.NoError(t, err)
	assert.Equal(t, store.StatusClosed, tk.Status)
}

func testClaimNextTiers(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	queue := task(t, s, "queue", "dev", prio(0))
	assigned := task(t, s, "assigned", "other", prio(4), func(n *store.NewTicket) { n.Holder = "s1" })
	_ = task(t, s, "elsewhere", "other", prio(0))
	req := store.ClaimNext{Project: proj, Pool: "dev", Holder: "s1", Lease: lease}

	got, c1, err := s.ClaimNext(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, assigned.ID, got.ID, "pre-assigned beats queue regardless of priority")
	assert.Equal(t, store.StatusInProgress, got.Status)
	assert.EqualValues(t, 1, c1.Fence)

	again, c2, err := s.ClaimNext(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, assigned.ID, again.ID, "in-progress work resumes")
	assert.Equal(t, c1.Fence, c2.Fence)

	_, err = s.CloseTicket(ctx, "s1", assigned.ID, &c2, "done")
	require.NoError(t, err)
	got, _, err = s.ClaimNext(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, queue.ID, got.ID)

	_, _, err = s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s2", Lease: lease})
	assert.ErrorIs(t, err, store.ErrNoWork)

	_, _, err = s.Claim(ctx, queue.ID, "s2", lease)
	assert.ErrorIs(t, err, store.ErrNotHolder)
}

func testFencedClose(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	_, old, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, old, "handoff", time.Time{})
	require.NoError(t, err)
	_, cur, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	assert.Greater(t, cur.Fence, old.Fence)

	_, err = s.CloseTicket(ctx, "s1", a.ID, &old, "stale")
	assert.ErrorIs(t, err, store.ErrNotHolder, "stale fence is rejected")
	c.Advance(time.Second)
	closed, err := s.CloseTicket(ctx, "s1", a.ID, &cur, "done")
	require.NoError(t, err)
	assert.Equal(t, store.StatusClosed, closed.Status)
	assert.Equal(t, "done", closed.Resolution)
	assert.Equal(t, c.T, closed.ClosedAt)
	assert.True(t, closed.LeaseUntil.IsZero())

	_, err = s.CloseTicket(ctx, "admin", a.ID, nil, "again")
	assert.ErrorIs(t, err, store.ErrClosed)
	re, err := s.Reopen(ctx, "admin", a.ID)
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, re.Status)
	assert.Empty(t, re.Resolution)
}

func testHeartbeat(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	tk, cl, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	c.Advance(10 * time.Minute)
	cl2, err := s.Heartbeat(ctx, cl, lease)
	require.NoError(t, err)
	assert.Equal(t, c.T.Add(lease), cl2.LeaseUntil)

	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, tk.Revision, got.Revision, "heartbeat does not bump revision")
	assert.Equal(t, cl2.LeaseUntil, got.LeaseUntil)

	bad := cl
	bad.Holder = "s2"
	_, err = s.Heartbeat(ctx, bad, lease)
	assert.ErrorIs(t, err, store.ErrNotHolder)
}

func testReleaseHolder(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	b := task(t, s, "b", "dev", func(n *store.NewTicket) { n.Holder = "s1" })
	other := task(t, s, "other", "dev")
	_, _, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	_, _, err = s.Claim(ctx, other.ID, "s2", lease)
	require.NoError(t, err)

	released, err := s.ReleaseHolder(ctx, "controller", "s1", "session died")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{a.ID, b.ID}, released)
	for _, id := range released {
		got, err := s.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, store.StatusOpen, got.Status)
		assert.Empty(t, got.Holder)
	}
	got, err := s.Get(ctx, other.ID)
	require.NoError(t, err)
	assert.Equal(t, "s2", got.Holder)
}

func testSweepLeases(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	b := task(t, s, "b", "dev")
	_, ca, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	_, cb, err := s.Claim(ctx, b.ID, "s2", lease)
	require.NoError(t, err)
	c.Advance(20 * time.Minute)
	_, err = s.Heartbeat(ctx, cb, lease)
	require.NoError(t, err)
	c.Advance(11 * time.Minute)

	swept, err := s.SweepLeases(ctx, "controller")
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, swept)
	_, err = s.Heartbeat(ctx, ca, lease)
	assert.ErrorIs(t, err, store.ErrNotHolder, "swept holder lost the claim")

	h, err := s.History(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, "lease_expired", h[len(h)-1].Op)
}

func testDeferUntil(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	later := c.T.Add(time.Hour)
	a := task(t, s, "a", "dev", func(n *store.NewTicket) { n.DeferUntil = later })
	r, err := s.Ready(ctx, proj, "dev", 0)
	require.NoError(t, err)
	assert.Empty(t, r)
	c.Advance(time.Hour)
	r, err = s.Ready(ctx, proj, "dev", 0)
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(r))
}

func testLabelsCommentsHistory(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "")
	require.NoError(t, s.AddLabel(ctx, "u", a.ID, "bug"))
	require.NoError(t, s.AddLabel(ctx, "u", a.ID, "bug"))
	require.NoError(t, s.RemoveLabel(ctx, "u", a.ID, "missing"))
	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"bug"}, got.Labels)
	assert.ErrorIs(t, s.AddLabel(ctx, "u", "js-missing", "x"), store.ErrNotFound)

	_, err = s.AddComment(ctx, "u", a.ID, "first")
	require.NoError(t, err)
	_, err = s.AddComment(ctx, "u", a.ID, "second")
	require.NoError(t, err)
	_, err = s.AddComment(ctx, "u", "js-missing", "x")
	assert.ErrorIs(t, err, store.ErrNotFound)
	cs, err := s.Comments(ctx, a.ID)
	require.NoError(t, err)
	require.Len(t, cs, 2)
	assert.Equal(t, "first", cs[0].Body)

	h, err := s.History(ctx, a.ID)
	require.NoError(t, err)
	ops := make([]string, len(h))
	for i, e := range h {
		ops[i] = e.Op
	}
	assert.Equal(t, []string{"create", "label_add", "comment", "comment"}, ops)

	// Changes is the global history after a seq, in order.
	ch, err := s.Changes(ctx, h[0].Seq, 0)
	require.NoError(t, err)
	require.Len(t, ch, 3)
	assert.Equal(t, h[1:], ch)
	last, err := s.LastSeq(ctx)
	require.NoError(t, err)
	assert.Equal(t, h[3].Seq, last)
	ch, err = s.Changes(ctx, 0, 1)
	require.NoError(t, err)
	assert.Len(t, ch, 1)
}

func testCreateGraphAtomic(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: store.KindTask}
	s1 := store.NewTicket{ID: "js-s1", Project: proj, Kind: store.KindTask, ParentID: "js-run"}
	s2 := store.NewTicket{ID: "js-s2", Project: proj, Kind: store.KindTask, ParentID: "js-run"}
	out, err := s.CreateGraph(ctx, "test", []store.NewTicket{run, s1, s2},
		[]store.Edge{{From: "js-s2", To: "js-s1", Kind: store.EdgeBlocks}})
	require.NoError(t, err)
	assert.Len(t, out, 3)
	kids, err := s.List(ctx, store.Filter{ParentID: "js-run"})
	require.NoError(t, err)
	assert.Equal(t, []string{"js-s1", "js-s2"}, ids(kids))

	// A cycle aborts the whole graph.
	x := store.NewTicket{ID: "js-x", Project: proj, Kind: store.KindTask}
	y := store.NewTicket{ID: "js-y", Project: proj, Kind: store.KindTask}
	_, err = s.CreateGraph(ctx, "test", []store.NewTicket{x, y}, []store.Edge{
		{From: "js-x", To: "js-y", Kind: store.EdgeBlocks},
		{From: "js-y", To: "js-x", Kind: store.EdgeBlocks},
	})
	assert.ErrorIs(t, err, store.ErrCycle)
	_, err = s.Get(ctx, "js-x")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func testSessionKeyUnique(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	sess := func() (store.Ticket, error) {
		return s.Create(ctx, "test", store.NewTicket{
			Project: proj, Kind: store.KindSession,
			Fields: json.RawMessage(`{"agent":"jcode","key":"matrix:!r"}`),
		})
	}
	a, err := sess()
	require.NoError(t, err)
	_, err = sess()
	assert.ErrorIs(t, err, store.ErrAlreadyExist)
	_, err = s.CloseTicket(ctx, "test", a.ID, nil, "stopped")
	require.NoError(t, err)
	_, err = sess()
	assert.NoError(t, err, "key is reusable once the session closes")
	_, err = s.Create(ctx, "test", store.NewTicket{
		Project: proj, Kind: store.KindSession,
		Fields: json.RawMessage(`{"agent":"jcode"}`),
	})
	assert.NoError(t, err)
	_, err = s.Create(ctx, "test", store.NewTicket{
		Project: proj, Kind: store.KindSession,
		Fields: json.RawMessage(`{"agent":"jcode"}`),
	})
	assert.NoError(t, err, "sessions without a key never collide")
}

// testApplyGraph: re-parent, insert and edges land together; a held
// Unclaimed ticket aborts the whole apply.
func testApplyGraph(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: store.KindRun, Fields: json.RawMessage(`{"workflow":"w","key":"issue-1","state":"running"}`)}
	step := store.NewTicket{
		ID: "js-s", Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: "js-run",
		Fields: json.RawMessage(`{"run":"js-run","step":"s","index":0,"retries":1}`),
	}
	_, err := s.CreateGraph(ctx, "test", []store.NewTicket{run, step}, nil)
	require.NoError(t, err)
	other := task(t, s, "other", "dev")
	batch := func(id string) store.NewTicket {
		return store.NewTicket{ID: id, Project: proj, Kind: store.KindBatch, Fields: json.RawMessage(`{"key":"k","issues":["1"],"state":"open"}`)}
	}

	// Held: nothing is written.
	_, _, err = s.Claim(ctx, "js-s", "sess", time.Minute)
	require.NoError(t, err)
	_, err = s.ApplyGraph(ctx, "test", store.Graph{
		Tickets: []store.NewTicket{batch("js-b1")}, Reparent: map[string]string{"js-run": "js-b1"},
		Edges: []store.Edge{{From: "js-s", To: other.ID, Kind: store.EdgeBlocks}}, Unclaimed: []string{"js-s"},
	})
	require.ErrorIs(t, err, store.ErrNotHolder)
	_, err = s.Get(ctx, "js-b1")
	require.ErrorIs(t, err, store.ErrNotFound)
	_, err = s.FindOpenBatchByKey(ctx, proj, "k")
	require.ErrorIs(t, err, store.ErrNotFound)

	out, err := s.ApplyGraph(ctx, "test", store.Graph{
		Tickets: []store.NewTicket{batch("js-b2")}, Reparent: map[string]string{"js-run": "js-b2"},
	})
	require.NoError(t, err)
	require.Len(t, out, 1)
	got, err := s.Get(ctx, "js-run")
	require.NoError(t, err)
	assert.Equal(t, "js-b2", got.ParentID)
	kids, err := s.ListChildren(ctx, "js-b2")
	require.NoError(t, err)
	assert.Equal(t, []string{"js-run"}, ids(kids))
	b, err := s.FindOpenBatchByKey(ctx, proj, "k")
	require.NoError(t, err)
	assert.Equal(t, "js-b2", b.ID)
	ready, err := s.Ready(ctx, proj, "", 0)
	require.NoError(t, err)
	assert.NotContains(t, ids(ready), "js-b2", "batches are never claimable")
}

// testPruneWriteKeys: keys older than the cutoff go, done or pending;
// newer keys stay (#415).
func testPruneWriteKeys(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	_, err := s.ClaimWriteKey(ctx, "js/old-done", "op", "a")
	require.NoError(t, err)
	_, err = s.FinishWriteKey(ctx, "js/old-done", "r")
	require.NoError(t, err)
	_, err = s.ClaimWriteKey(ctx, "js/old-pending", "op", "a")
	require.NoError(t, err)
	c.Advance(47 * time.Hour)
	_, err = s.ClaimWriteKey(ctx, "js/new", "op", "a")
	require.NoError(t, err)
	c.Advance(2 * time.Hour)

	n, err := s.PruneWriteKeys(ctx, c.Now().Add(-48*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
	k, err := s.ClaimWriteKey(ctx, "js/old-done", "op", "a")
	require.NoError(t, err)
	assert.True(t, k.New, "pruned key is claimed fresh")
	k, err = s.ClaimWriteKey(ctx, "js/new", "op", "a")
	require.NoError(t, err)
	assert.False(t, k.New, "newer key kept")
}
