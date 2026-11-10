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

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
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
		{"ApplySwapsEdges", testApplySwapsEdges},
		{"ClaimNextTiers", testClaimNextTiers},
		{"FencedClose", testFencedClose},
		{"Handoff", testHandoff},
		{"ReleaseOutcomeRules", testReleaseOutcomeRules},
		{"Heartbeat", testHeartbeat},
		{"ReleaseHolder", testReleaseHolder},
		{"SweepLeases", testSweepLeases},
		{"DeferUntil", testDeferUntil},
		{"LabelsCommentsHistory", testLabelsCommentsHistory},
		{"CommentsFor", testCommentsFor},
		{"CreateGraphAtomic", testCreateGraphAtomic},
		{"SessionKeyUnique", testSessionKeyUnique},
		{"ListMetaNewest", testListMetaNewest},
		{"ListCreatedBy", testListCreatedBy},
		{"ListFilters", testListFilters},
		{"ListShapes", testListShapes},
		{"ListQuery", testListQuery},
		{"Orphaned", testOrphaned},
		{"Runs", testRuns},
		{"StepFailures", testStepFailures},
		{"RequestChanges", testRequestChanges},
		{"ApplyGraph", testApplyGraph},
		{"ApplyRevisionedGraph", testApplyRevisionedGraph},
		{"StructuredStepResultFences", testStructuredStepResultFences},
		{"Resources", testResources},
		{"ProviderSlots", testProviderSlots},
		{"ConcurrentSlotClaims", testConcurrentSlotClaims},
		{"ConcurrentClaims", testConcurrentClaims},
		{"ConcurrentResourceClaims", testConcurrentResourceClaims},
		{"ClaimAfterSkippedBlocker", testClaimAfterSkippedBlocker},
		{"RaceClaimNextExtend", testRaceClaimNextExtend},
		{"RaceReleaseClaimFences", testRaceReleaseClaimFences},
		{"RaceReleaseExpire", testRaceReleaseExpire},
		{"RaceHeartbeatSweep", testRaceHeartbeatSweep},
		{"RaceUpdateClaim", testRaceUpdateClaim},
		{"RaceCASUpdates", testRaceCASUpdates},
		{"RaceSlotRelease", testRaceSlotRelease},
		{"ReadAfterWrite", testReadAfterWrite},
		{"WaitGates", testWaitGates},
		{"ReopenReady", testReopenReady},
		{"TracksDoesNotGate", testTracksDoesNotGate},
		{"ListMoreFilters", testListMoreFilters},
		{"RecallFilters", testRecallFilters},
		{"CollectionSpec", testCollectionSpec},
		{"CollectionPutGet", testCollectionPutGet},
		{"CollectionUnique", testCollectionUnique},
		{"CollectionUpdate", testCollectionUpdate},
		{"CollectionClaim", testCollectionClaim},
		{"CollectionListCount", testCollectionListCount},
		{"CollectionDelete", testCollectionDelete},
		{"CollectionSearch", testCollectionSearch},
		{"CollectionConcurrentClaims", testCollectionConcurrentClaims},
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

func task(t *testing.T, s store.Store, title, pool string, opts ...func(*store.NewTicket)) domain.Ticket {
	t.Helper()
	nt := store.NewTicket{Project: proj, Kind: domain.KindTask, Title: title, Pool: pool}
	for _, o := range opts {
		o(&nt)
	}
	tk, err := store.Create(t.Context(), s, "test", nt)
	require.NoError(t, err)
	return tk
}

func prio(p int) func(*store.NewTicket) { return func(n *store.NewTicket) { n.Priority = &p } }

func ids(ts []domain.Ticket) []string {
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

	l, err = s.List(t.Context(), store.Filter{MetaPack: "trigger", MetaKey: "name", MetaValue: "x", Order: store.OrderNewest, Limit: 1})
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID}, ids(l), "newest first")

	// Names are quoted into the JSON path, not interpreted.
	l, err = s.List(t.Context(), store.Filter{MetaPack: `trigger"."name`, MetaKey: "name", MetaValue: "x"})
	require.NoError(t, err)
	assert.Empty(t, l)
}

func testListFilters(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	withMeta := func(meta string) func(*store.NewTicket) {
		return func(n *store.NewTicket) { n.Meta = json.RawMessage(meta) }
	}
	a := task(t, s, "a", "p", withMeta(`{"myflow":{"stage":"review","owner":"x"}}`))
	c.Advance(time.Second)
	b := task(t, s, "b", "p", withMeta(`{"myflow":{"stage":"review","owner":"y"}}`))
	c.Advance(time.Second)
	cc := task(t, s, "c", "p", withMeta(`{"myflow":{"stage":"done","owner":"x"}}`))

	list := func(f store.Filter) []string {
		t.Helper()
		l, err := s.List(ctx, f)
		require.NoError(t, err)
		return ids(l)
	}
	assert.Equal(t, []string{a.ID, b.ID}, list(store.Filter{Meta: []store.MetaMatch{{Pack: "myflow", Key: "stage", Value: "review"}}}))
	assert.Equal(t, []string{a.ID}, list(store.Filter{Meta: []store.MetaMatch{{Pack: "myflow", Key: "stage", Value: "review"}, {Pack: "myflow", Key: "owner", Value: "x"}}}))
	assert.Equal(t, []string{a.ID, cc.ID}, list(store.Filter{MetaPack: "myflow", MetaKey: "owner", MetaValue: "x"}))

	// Cursor paging in both orders.
	assert.Equal(t, []string{a.ID}, list(store.Filter{Limit: 1}))
	assert.Equal(t, []string{b.ID}, list(store.Filter{Limit: 1, After: a.ID}))
	assert.Equal(t, []string{cc.ID}, list(store.Filter{After: b.ID}))
	assert.Equal(t, []string{b.ID, a.ID}, list(store.Filter{Order: store.OrderNewest, After: cc.ID}))

	// Since: updated at or after.
	c.Advance(time.Minute)
	mark := c.Now()
	b2 := "b2"
	_, err := s.Update(ctx, "test", b.ID, store.Update{Title: &b2})
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID}, list(store.Filter{Since: mark}))

	_, err = s.List(ctx, store.Filter{Meta: make([]store.MetaMatch, store.MaxMetaFilters+1)})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

func testListCreatedBy(t *testing.T, s store.Store, _ *Clock) {
	a, err := store.Create(t.Context(), s, "js-sess1", store.NewTicket{Project: "js", Kind: domain.KindTask, Title: "a"})
	require.NoError(t, err)
	_, err = store.Create(t.Context(), s, "admin", store.NewTicket{Project: "js", Kind: domain.KindTask, Title: "b"})
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
	assert.Equal(t, domain.StatusOpen, a.Status)
	assert.Equal(t, store.DefaultPriority, a.Priority)
	assert.EqualValues(t, 1, a.Revision)

	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"x", "y"}, got.Labels)
	assert.JSONEq(t, `{"gx":{"linear_id":"ENG-1"}}`, string(got.Meta))
	assert.Equal(t, "test", got.CreatedBy)

	_, err = s.Get(ctx, "js-missing")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	_, err = store.Create(ctx, s, "test", store.NewTicket{ID: a.ID, Project: proj, Kind: domain.KindTask})
	assert.ErrorIs(t, err, domain.ErrAlreadyExist)

	_, err = store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindTask, ParentID: "js-missing"})
	assert.ErrorIs(t, err, domain.ErrNotFound)

	l, err := s.List(ctx, store.Filter{Project: proj, Label: "x"})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(l))
}

func testFieldsValidated(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	cases := []store.NewTicket{
		{Project: proj, Kind: "nope"},
		{Project: proj, Kind: domain.KindMessage, Fields: json.RawMessage(`{}`)},
		{Project: proj, Kind: domain.KindMessage, Fields: json.RawMessage(`{"to":"pool:"}`)},
		{Project: proj, Kind: domain.KindSession, Fields: json.RawMessage(`{"agent":"a","extra":1}`)},
		{Project: proj, Kind: domain.KindTask, Meta: json.RawMessage(`{"flat":"string"}`)},
		{Kind: domain.KindTask},
	}
	for _, nt := range cases {
		_, err := store.Create(ctx, s, "test", nt)
		assert.ErrorIs(t, err, domain.ErrInvalid, "%+v", nt)
	}
	m, err := store.Create(ctx, s, "test", store.NewTicket{
		Project: proj, Kind: domain.KindMessage,
		Fields: json.RawMessage(`{"to":"pool:dev","from":"me"}`),
	})

	require.NoError(t, err)
	f, err := domain.DecodeFields[domain.MessageFields](m)
	require.NoError(t, err)
	assert.Equal(t, domain.MessageFields{To: "pool:dev", From: "me"}, f)
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
	assert.ErrorIs(t, err, domain.ErrConflict)

	_, err = s.Update(ctx, "test", a.ID, store.Update{Fields: json.RawMessage(`{"bad":1}`)})
	assert.ErrorIs(t, err, domain.ErrInvalid)
}

func testReadyRespectsEdges(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev", prio(1))
	b := task(t, s, "b", "dev", prio(2))
	c := task(t, s, "c", "dev", prio(3))
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: a.ID, To: b.ID, Kind: domain.EdgeBlocks}))
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: c.ID, To: b.ID, Kind: domain.EdgeTracks}))

	r, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
	require.NoError(t, err)
	assert.Equal(t, []string{b.ID, c.ID}, ids(r), "a is blocked; tracks does not block")

	_, err = store.Close(ctx, s, "test", b.ID, nil, "done")
	require.NoError(t, err)
	r, err = s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID, c.ID}, ids(r))

	e, err := store.Edges(ctx, s, b.ID)
	require.NoError(t, err)
	assert.Len(t, e, 2)
	require.NoError(t, store.RemoveEdge(ctx, s, "test", domain.Edge{From: c.ID, To: b.ID, Kind: domain.EdgeTracks}))
	assert.ErrorIs(t, store.RemoveEdge(ctx, s, "test", domain.Edge{From: c.ID, To: b.ID, Kind: domain.EdgeTracks}), domain.ErrNotFound)
}

func testEdgeCycle(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a, b, c := task(t, s, "a", ""), task(t, s, "b", ""), task(t, s, "c", "")
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: a.ID, To: b.ID, Kind: domain.EdgeBlocks}))
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: b.ID, To: c.ID, Kind: domain.EdgeWaitsFor}))
	assert.ErrorIs(t, store.AddEdge(ctx, s, "test", domain.Edge{From: c.ID, To: a.ID, Kind: domain.EdgeBlocks}), domain.ErrCycle)
	assert.ErrorIs(t, store.AddEdge(ctx, s, "test", domain.Edge{From: a.ID, To: a.ID, Kind: domain.EdgeBlocks}), domain.ErrInvalid)
	// tracks never gates readiness, so it may close a loop.
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: c.ID, To: a.ID, Kind: domain.EdgeTracks}))
	assert.ErrorIs(t, store.AddEdge(ctx, s, "test", domain.Edge{From: a.ID, To: "js-missing", Kind: domain.EdgeBlocks}), domain.ErrNotFound)
}

// testApplySwapsEdges moves a blocking edge in one Apply: removals run
// before additions, and a failed removal writes nothing.
func testApplySwapsEdges(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a, b, c := task(t, s, "a", ""), task(t, s, "b", ""), task(t, s, "c", "")
	ab := domain.Edge{From: a.ID, To: b.ID, Kind: domain.EdgeBlocks}
	ba := domain.Edge{From: b.ID, To: a.ID, Kind: domain.EdgeBlocks}
	ac := domain.Edge{From: a.ID, To: c.ID, Kind: domain.EdgeBlocks}
	require.NoError(t, store.AddEdge(ctx, s, "test", ab))

	// b -> a would be a cycle unless a -> b is removed first.
	_, err := s.Apply(ctx, "test", store.Graph{RemoveEdges: []domain.Edge{ab}, Edges: []domain.Edge{ba}})
	require.NoError(t, err)

	_, err = s.Apply(ctx, "test", store.Graph{RemoveEdges: []domain.Edge{ab}, Edges: []domain.Edge{ac}})
	require.ErrorIs(t, err, domain.ErrNotFound)
	e, err := store.Edges(ctx, s, a.ID)
	require.NoError(t, err)
	require.Len(t, e, 1, "the failed apply added nothing")
	assert.Equal(t, b.ID, e[0].From)
}

func testOrphaned(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	none := task(t, s, "none", "")
	gone := task(t, s, "gone", "gone")
	_ = task(t, s, "queued", "dev")
	_ = task(t, s, "held", "gone", func(n *store.NewTicket) { n.Holder = "s1" })
	closed := task(t, s, "closed", "")
	_, err := store.Close(ctx, s, "test", closed.ID, nil, "done")
	require.NoError(t, err)

	orphaned := func(project string, pools []string) ([]domain.Ticket, error) {
		return s.List(ctx, store.Filter{Project: project, Kinds: domain.ClaimableKinds, Status: domain.StatusOpen, Unheld: true, PoolNotIn: pools})
	}
	got, err := orphaned("", []string{"dev"})
	require.NoError(t, err)
	assert.Equal(t, []string{none.ID, gone.ID}, ids(got))

	got, err = orphaned("other", []string{"dev"})
	require.NoError(t, err)
	assert.Empty(t, got)

	got, err = orphaned(proj, nil)
	require.NoError(t, err)
	assert.Len(t, got, 3, "no known pools: every pool is unknown")
}

func testRuns(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	runID, err := s.NewID(proj)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(runID, proj+"-"))
	run := store.NewTicket{
		ID: runID, Project: proj, Kind: domain.KindRun, Title: "r",
		Fields: json.RawMessage(`{"workflow":"w","key":"k1","state":"running"}`),
	}
	step := func(i int, id string) store.NewTicket {
		return store.NewTicket{
			ID: id, Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: runID,
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"s%d","index":%d,"retries":2}`, runID, i, i)),
		}
	}
	// Inserted out of order: ListChildren orders by fields.index (10 > 2).
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, step(10, "js-b"), step(2, "js-a")}, Edges: nil})
	require.NoError(t, err)
	kids, err := store.Children(ctx, s, runID)
	require.NoError(t, err)
	assert.Equal(t, []string{"js-a", "js-b"}, ids(kids))

	got, err := store.OpenRun(ctx, s, "w", "k1")
	require.NoError(t, err)
	assert.Equal(t, runID, got.ID)
	_, err = store.OpenRun(ctx, s, "other", "k1")
	assert.ErrorIs(t, err, domain.ErrNotFound)

	r, err := s.List(ctx, store.Filter{Kind: domain.KindRun, Project: proj, State: domain.RunRunning, Order: store.OrderNewest, Limit: 0})
	require.NoError(t, err)
	assert.Equal(t, []string{runID}, ids(r))
	r, err = s.List(ctx, store.Filter{Kind: domain.KindRun, State: domain.RunNeedsHuman, Order: store.OrderNewest})
	require.NoError(t, err)
	assert.Empty(t, r)

	// Steps are claimable; runs are not.
	ready, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"js-a", "js-b"}, ids(ready))

	_, err = store.Close(ctx, s, "test", runID, nil, "done")
	require.NoError(t, err)
	_, err = store.OpenRun(ctx, s, "w", "k1")
	assert.ErrorIs(t, err, domain.ErrNotFound, "closed runs do not dedup")
}

// testStepFailures: each failure source bumps fields.failures in its own
// write; below retries the step reopens, at the limit it closes failed and
// keeps the next step blocked.
func testStepFailures(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	mk := func(id string, i int) store.NewTicket {
		return store.NewTicket{
			ID: id, Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: "js-run",
			Fields: json.RawMessage(fmt.Sprintf(`{"run":"js-run","step":"s%d","index":%d,"retries":4}`, i, i)),
		}
	}
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)}
	_, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, mk("js-a", 0), mk("js-b", 1)}, Edges: []domain.Edge{{From: "js-b", To: "js-a", Kind: domain.EdgeBlocks}}})

	require.NoError(t, err)
	fails := func() int {
		t.Helper()
		tk, err := s.Get(ctx, "js-a")
		require.NoError(t, err)
		f, err := domain.DecodeFields[domain.StepFields](tk)
		require.NoError(t, err)
		return f.Failures
	}

	// Close failed reopens below the limit.
	tk, err := store.Close(ctx, s, "test", "js-a", nil, "failed")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, tk.Status)
	assert.Equal(t, 1, fails())

	// A rebase close reopens without counting, however often (#424).
	for range 5 {
		tk, err = store.Close(ctx, s, "test", "js-a", nil, domain.ResolutionRebase)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusOpen, tk.Status)
		assert.Empty(t, tk.Resolution)
	}
	assert.Equal(t, 1, fails())

	// Release.
	_, cl, err := s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: "gave up"})
	require.NoError(t, err)
	assert.Equal(t, 2, fails())

	// A deferred release is not a failure and hides the step until then (#139).
	_, cl, err = s.Claim(ctx, "js-a", "s1", lease)
	require.NoError(t, err)
	tk, err = s.Release(ctx, "", cl.TicketID, &cl, store.Outcome{Reason: "not yet", Until: c.T.Add(time.Hour)})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, tk.Status)
	assert.Equal(t, 2, fails())
	r, err := s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.NotContains(t, ids(r), "js-a")
	c.Advance(time.Hour)
	r, err = s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Contains(t, ids(r), "js-a")

	// Dead holder.
	_, _, err = s.Claim(ctx, "js-a", "s2", lease)
	require.NoError(t, err)
	_, err = s.Expire(ctx, "controller", "s2", "session died")
	require.NoError(t, err)
	assert.Equal(t, 3, fails())

	// Lease expiry reaches the limit: closed failed.
	_, _, err = s.Claim(ctx, "js-a", "s3", lease)
	require.NoError(t, err)
	c.Advance(2 * lease)
	_, err = s.Expire(ctx, "controller", "", "")
	require.NoError(t, err)
	assert.Equal(t, 4, fails())
	tk, err = s.Get(ctx, "js-a")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, tk.Status)
	assert.Equal(t, "failed", tk.Resolution)
	cs, err := store.Comments(ctx, s, "js-a")
	require.NoError(t, err)
	require.Len(t, cs, 10, "one comment per counted failure, per rebase close, plus the deferral note")
	assert.Contains(t, cs[1].Body, "Closed for a rebase")
	assert.Contains(t, cs[7].Body, "Released until")
	assert.Contains(t, cs[9].Body, "Retry limit 4 reached")

	// A step closed failed keeps the next step blocked.
	ready, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
	require.NoError(t, err)
	assert.Empty(t, ids(ready))
	_, _, err = s.Claim(ctx, "js-b", "s4", lease)
	assert.ErrorIs(t, err, domain.ErrNoWork)

	// Tasks are not counted.
	task := task(t, s, "plain", "dev")
	tk, err = store.Close(ctx, s, "test", task.ID, nil, "failed")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, tk.Status)
}

// testRequestChanges: a request_changes close of a step reopens it and
// the step it reviews in one write, with the findings on the reviewed
// step and no failure counted; past the round cap the reviewer stays
// closed and nothing reopens (JOB-98).
func testRequestChanges(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)}
	deliver := store.NewTicket{
		ID: "js-a", Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: "js-run",
		Fields: json.RawMessage(`{"run":"js-run","step":"deliver","index":0,"retries":2}`),
	}
	review := store.NewTicket{
		ID: "js-b", Project: proj, Kind: domain.KindStep, Pool: "review", ParentID: "js-run",
		Fields: json.RawMessage(`{"run":"js-run","step":"review","index":1,"retries":2,"on_request_changes":"deliver","max_review_rounds":2}`),
	}
	_, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, deliver, review}, Edges: []domain.Edge{{From: "js-b", To: "js-a", Kind: domain.EdgeBlocks}}})
	require.NoError(t, err)
	fields := func(id string) (domain.Ticket, domain.StepFields) {
		t.Helper()
		tk, err := s.Get(ctx, id)
		require.NoError(t, err)
		f, err := domain.DecodeFields[domain.StepFields](tk)
		require.NoError(t, err)
		return tk, f
	}
	requestChanges := func(findings string) domain.Ticket {
		t.Helper()
		_, err := store.Close(ctx, s, "test", "js-a", nil, "done")
		require.NoError(t, err)
		_, cl, err := s.Claim(ctx, "js-b", "s1", lease)
		require.NoError(t, err)
		tk, err := s.Release(ctx, "", "js-b", &cl, store.Outcome{Close: domain.ResolutionRequestChanges, Reopen: "js-a", Note: findings})
		require.NoError(t, err)
		return tk
	}

	for round := 1; round <= 2; round++ {
		tk := requestChanges(fmt.Sprintf("fix %d", round))
		assert.Equal(t, domain.StatusOpen, tk.Status, "round %d reopens the reviewer", round)
		assert.Empty(t, tk.Holder)
		_, rf := fields("js-b")
		assert.Equal(t, round, rf.ReviewRounds)
		assert.Zero(t, rf.Failures, "not a failed attempt")
		a, af := fields("js-a")
		assert.Equal(t, domain.StatusOpen, a.Status)
		assert.Empty(t, a.Resolution)
		require.NotNil(t, af.Changes)
		assert.Equal(t, domain.StepChanges{Step: "review", Ticket: "js-b", Round: round, Max: 2, Findings: fmt.Sprintf("fix %d", round)}, *af.Changes)
		assert.Zero(t, af.Failures)
		// The reviewer waits behind the reopened step.
		ready, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
		require.NoError(t, err)
		assert.Equal(t, []string{"js-a"}, ids(ready))
	}
	cs, err := store.Comments(ctx, s, "js-a")
	require.NoError(t, err)
	require.Len(t, cs, 2)
	assert.Contains(t, cs[1].Body, "Changes requested by step review (js-b), review round 2 of 2")

	// Round 3 is over the cap: the reviewer stays closed, deliver is left
	// alone (closed done, which cleared the request).
	tk := requestChanges("fix 3")
	assert.Equal(t, domain.StatusClosed, tk.Status)
	assert.Equal(t, domain.ResolutionRequestChanges, tk.Resolution)
	_, rf := fields("js-b")
	assert.Equal(t, 3, rf.ReviewRounds)
	a, af := fields("js-a")
	assert.Equal(t, domain.StatusClosed, a.Status)
	assert.Nil(t, af.Changes, "a done close clears the request")

	// A target that is not closed rolls the whole close back.
	_, err = s.Update(ctx, "test", "js-a", store.Update{Reopen: true})
	require.NoError(t, err)
	_, err = s.Update(ctx, "test", "js-b", store.Update{Reopen: true, Fields: json.RawMessage(`{"run":"js-run","step":"review","index":1,"retries":2}`)})
	require.NoError(t, err)
	_, err = s.Release(ctx, "test", "js-b", nil, store.Outcome{Close: domain.ResolutionRequestChanges, Reopen: "js-a", Note: "x"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	b, bf := fields("js-b")
	assert.Equal(t, domain.StatusOpen, b.Status, "rolled back")
	assert.Zero(t, bf.ReviewRounds, "rolled back")
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
	assert.Equal(t, domain.StatusInProgress, got.Status)
	assert.EqualValues(t, 1, c1.Fence)

	again, c2, err := s.ClaimNext(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, assigned.ID, again.ID, "in-progress work resumes")
	assert.Equal(t, c1.Fence, c2.Fence)

	_, err = store.Close(ctx, s, "s1", assigned.ID, &c2, "done")
	require.NoError(t, err)
	got, _, err = s.ClaimNext(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, queue.ID, got.ID)

	_, _, err = s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s2", Lease: lease})
	assert.ErrorIs(t, err, domain.ErrNoWork)

	_, _, err = s.Claim(ctx, queue.ID, "s2", lease)
	assert.ErrorIs(t, err, domain.ErrNotHolder)
}

// testHandoff: pool, release, fence and note in one revision (#492).
func testHandoff(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	held, cl, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	stale := cl
	stale.Fence--
	_, err = s.Release(ctx, "s1", a.ID, &stale, store.Outcome{Pool: "refinery", Note: "x"})
	require.ErrorIs(t, err, domain.ErrNotHolder)

	got, err := s.Release(ctx, "s1", a.ID, &cl, store.Outcome{Pool: "refinery", Note: "over to you"})
	require.NoError(t, err)
	assert.Equal(t, "refinery", got.Pool)
	assert.Equal(t, domain.StatusOpen, got.Status)
	assert.Empty(t, got.Holder)
	assert.Greater(t, got.Fence, cl.Fence)
	assert.Equal(t, held.Revision+1, got.Revision)
	cs, err := store.Comments(ctx, s, a.ID)
	require.NoError(t, err)
	require.Len(t, cs, 1)
	assert.Equal(t, "over to you", cs[0].Body)
	assert.Equal(t, "s1", cs[0].Author)

	_, err = s.Release(ctx, "admin", a.ID, nil, store.Outcome{})
	require.ErrorIs(t, err, domain.ErrInvalid)
}

// testReleaseOutcomeRules (#579): close and pool are exclusive, a claim
// for another ticket is rejected, and an empty id or actor comes from
// the claim.
func testReleaseOutcomeRules(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	b := task(t, s, "b", "dev")
	_, cl, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)

	_, err = s.Release(ctx, "s1", a.ID, &cl, store.Outcome{Close: "done", Pool: "refinery"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = s.Release(ctx, "s1", b.ID, &cl, store.Outcome{Close: "done"})
	require.ErrorIs(t, err, domain.ErrInvalid)

	got, err := s.Release(ctx, "", "", &cl, store.Outcome{Close: "done"})
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, got.Status)
	assert.Equal(t, "done", got.Resolution)
	h, err := store.History(ctx, s, a.ID)
	require.NoError(t, err)
	assert.Equal(t, "s1", h[len(h)-1].Actor)
}

func testFencedClose(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev")
	_, old, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, "", old.TicketID, &old, store.Outcome{Reason: "handoff"})
	require.NoError(t, err)
	_, cur, err := s.Claim(ctx, a.ID, "s1", lease)
	require.NoError(t, err)
	assert.Greater(t, cur.Fence, old.Fence)

	_, err = store.Close(ctx, s, "s1", a.ID, &old, "stale")
	assert.ErrorIs(t, err, domain.ErrNotHolder, "stale fence is rejected")
	c.Advance(time.Second)
	closed, err := store.Close(ctx, s, "s1", a.ID, &cur, "done")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, closed.Status)
	assert.Equal(t, "done", closed.Resolution)
	assert.Equal(t, c.T, closed.ClosedAt)
	assert.True(t, closed.LeaseUntil.IsZero())

	_, err = store.Close(ctx, s, "admin", a.ID, nil, "again")
	assert.ErrorIs(t, err, domain.ErrClosed)
	re, err := store.Reopen(ctx, s, "admin", a.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, re.Status)
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
	assert.ErrorIs(t, err, domain.ErrNotHolder)
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

	released, err := s.Expire(ctx, "controller", "s1", "session died")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{a.ID, b.ID}, released)
	for _, id := range released {
		got, err := s.Get(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, domain.StatusOpen, got.Status)
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

	swept, err := s.Expire(ctx, "controller", "", "")
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, swept)
	_, err = s.Heartbeat(ctx, ca, lease)
	assert.ErrorIs(t, err, domain.ErrNotHolder, "swept holder lost the claim")

	h, err := store.History(ctx, s, a.ID)
	require.NoError(t, err)
	assert.Equal(t, "lease_expired", h[len(h)-1].Op)
}

func testDeferUntil(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	later := c.T.Add(time.Hour)
	a := task(t, s, "a", "dev", func(n *store.NewTicket) { n.DeferUntil = later })
	r, err := s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Empty(t, r)
	c.Advance(time.Hour)
	r, err = s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ids(r))
}

// CommentsFor reads many tickets in one call, in the order Thread gives.
func testCommentsFor(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a, b, none := task(t, s, "a", ""), task(t, s, "b", ""), task(t, s, "none", "")
	for _, c := range []struct{ id, body string }{{a.ID, "a1"}, {b.ID, "b1"}, {a.ID, "a2"}, {b.ID, "b2"}, {a.ID, "a3"}} {
		_, err := store.AddComment(ctx, s, "u", c.id, c.body)
		require.NoError(t, err)
	}
	got, err := s.CommentsFor(ctx, []string{a.ID, b.ID, none.ID, "js-missing"})
	require.NoError(t, err)
	assert.Len(t, got, 2)
	for _, tk := range []domain.Ticket{a, b} {
		want, err := store.Comments(ctx, s, tk.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got[tk.ID])
	}
	assert.Len(t, got[a.ID], 3)
	assert.Equal(t, "a1", got[a.ID][0].Body)
	assert.Equal(t, "a3", got[a.ID][2].Body)
	assert.Empty(t, got[none.ID])

	got, err = s.CommentsFor(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, got)
}

func testLabelsCommentsHistory(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "")
	require.NoError(t, store.AddLabel(ctx, s, "u", a.ID, "bug"))
	require.NoError(t, store.AddLabel(ctx, s, "u", a.ID, "bug"))
	require.NoError(t, store.RemoveLabel(ctx, s, "u", a.ID, "missing"))
	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, []string{"bug"}, got.Labels)
	assert.ErrorIs(t, store.AddLabel(ctx, s, "u", "js-missing", "x"), domain.ErrNotFound)

	_, err = store.AddComment(ctx, s, "u", a.ID, "first")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "u", a.ID, "second")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "u", "js-missing", "x")
	assert.ErrorIs(t, err, domain.ErrNotFound)
	cs, err := store.Comments(ctx, s, a.ID)
	require.NoError(t, err)
	require.Len(t, cs, 2)
	assert.Equal(t, "first", cs[0].Body)

	h, err := store.History(ctx, s, a.ID)
	require.NoError(t, err)
	ops := make([]string, len(h))
	for i, e := range h {
		ops[i] = e.Op
	}
	assert.Equal(t, []string{"create", "label_add", "comment", "comment"}, ops)

	// Changes is the global history after a seq, in order.
	ch, err := s.Changes(ctx, h[0].Seq, 0)
	require.NoError(t, err)
	require.Len(t, ch.Changes, 3)
	assert.Equal(t, h[1:], ch.Changes)
	assert.Equal(t, h[3].Seq, ch.Last)
	last, err := store.LastSeq(ctx, s)
	require.NoError(t, err)
	assert.Equal(t, h[3].Seq, last)
	ch, err = s.Changes(ctx, 0, 1)
	require.NoError(t, err)
	assert.Len(t, ch.Changes, 1)
	assert.Equal(t, h[3].Seq, ch.Last)
	ch, err = s.Changes(ctx, 0, -1)
	require.NoError(t, err)
	assert.Empty(t, ch.Changes)
	assert.Equal(t, h[3].Seq, ch.Last)

	// Thread reads only the parts asked for.
	th, err := s.Thread(ctx, a.ID, store.Parts{Comments: true})
	require.NoError(t, err)
	assert.Len(t, th.Comments, 2)
	assert.Nil(t, th.Edges)
	assert.Nil(t, th.History)
	th, err = s.Thread(ctx, a.ID, store.AllParts)
	require.NoError(t, err)
	assert.Equal(t, cs, th.Comments)
	assert.Equal(t, h, th.History)
	assert.NotNil(t, th.Edges)

	// One Update carries fields, labels and a comment in one transaction
	// (#579); labels and the comment do not bump the revision.
	title := "a2"
	pre, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	got, err = s.Update(ctx, "u", a.ID, store.Update{
		Title: &title, AddLabels: []string{"x", "y"}, RemoveLabels: []string{"bug"}, Comment: "third",
	})
	require.NoError(t, err)
	assert.Equal(t, "a2", got.Title)
	assert.ElementsMatch(t, []string{"x", "y"}, got.Labels)
	assert.Equal(t, pre.Revision+1, got.Revision)
	got, err = s.Update(ctx, "u", a.ID, store.Update{AddLabels: []string{"z"}})
	require.NoError(t, err)
	assert.Equal(t, pre.Revision+1, got.Revision)
	_, err = s.Update(ctx, "u", a.ID, store.Update{Comment: "  "})
	require.ErrorIs(t, err, domain.ErrInvalid)
	// A failing part rolls back the rest.
	_, err = s.Update(ctx, "u", a.ID, store.Update{Reopen: true, Comment: "lost"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	cs, err = store.Comments(ctx, s, a.ID)
	require.NoError(t, err)
	assert.Len(t, cs, 3)

	// Record adds a row without changing the ticket (#542).
	before, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	require.NoError(t, s.Record(ctx, "desk:alice", a.ID, "desk_exec", map[string]any{"name": "x"}))
	h, err = store.History(ctx, s, a.ID)
	require.NoError(t, err)
	lastRow := h[len(h)-1]
	assert.Equal(t, "desk_exec", lastRow.Op)
	assert.Equal(t, "desk:alice", lastRow.Actor)
	assert.Equal(t, before.Revision, lastRow.Revision)
	assert.JSONEq(t, `{"name":"x"}`, string(lastRow.Diff))
	after, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	assert.Equal(t, before.Revision, after.Revision)
	assert.ErrorIs(t, s.Record(ctx, "u", "js-missing", "x", nil), domain.ErrNotFound)
}

func testCreateGraphAtomic(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: domain.KindTask}
	s1 := store.NewTicket{ID: "js-s1", Project: proj, Kind: domain.KindTask, ParentID: "js-run"}
	s2 := store.NewTicket{ID: "js-s2", Project: proj, Kind: domain.KindTask, ParentID: "js-run"}
	out, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, s1, s2}, Edges: []domain.Edge{{From: "js-s2", To: "js-s1", Kind: domain.EdgeBlocks}}})

	require.NoError(t, err)
	assert.Len(t, out, 3)
	kids, err := s.List(ctx, store.Filter{ParentID: "js-run"})
	require.NoError(t, err)
	assert.Equal(t, []string{"js-s1", "js-s2"}, ids(kids))

	// A cycle aborts the whole graph.
	x := store.NewTicket{ID: "js-x", Project: proj, Kind: domain.KindTask}
	y := store.NewTicket{ID: "js-y", Project: proj, Kind: domain.KindTask}
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{x, y}, Edges: []domain.Edge{
		{From: "js-x", To: "js-y", Kind: domain.EdgeBlocks},
		{From: "js-y", To: "js-x", Kind: domain.EdgeBlocks},
	}})

	assert.ErrorIs(t, err, domain.ErrCycle)
	_, err = s.Get(ctx, "js-x")
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func testSessionKeyUnique(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	sess := func() (domain.Ticket, error) {
		return store.Create(ctx, s, "test", store.NewTicket{
			Project: proj, Kind: domain.KindSession,
			Fields: json.RawMessage(`{"agent":"jcode","key":"matrix:!r"}`),
		})
	}
	a, err := sess()
	require.NoError(t, err)
	assert.Regexp(t, `^js-s-[0-9a-z]{13}$`, a.ID, "session ids carry the s tag (#615)")
	_, err = sess()
	assert.ErrorIs(t, err, domain.ErrAlreadyExist)
	_, err = store.Close(ctx, s, "test", a.ID, nil, "stopped")
	require.NoError(t, err)
	_, err = sess()
	assert.NoError(t, err, "key is reusable once the session closes")
	_, err = store.Create(ctx, s, "test", store.NewTicket{
		Project: proj, Kind: domain.KindSession,
		Fields: json.RawMessage(`{"agent":"jcode"}`),
	})

	assert.NoError(t, err)
	_, err = store.Create(ctx, s, "test", store.NewTicket{
		Project: proj, Kind: domain.KindSession,
		Fields: json.RawMessage(`{"agent":"jcode"}`),
	})

	assert.NoError(t, err, "sessions without a key never collide")
}

// testApplyGraph: re-parent, insert and edges land together; a held
// Unclaimed ticket aborts the whole apply.
func testApplyGraph(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"w","key":"issue-1","state":"running"}`)}
	step := store.NewTicket{
		ID: "js-s", Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: "js-run",
		Fields: json.RawMessage(`{"run":"js-run","step":"s","index":0,"retries":1}`),
	}
	_, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, step}, Edges: nil})
	require.NoError(t, err)
	other := task(t, s, "other", "dev")
	batch := func(id string) store.NewTicket {
		return store.NewTicket{ID: id, Project: proj, Kind: domain.KindBatch, Fields: json.RawMessage(`{"key":"k","issues":["1"],"state":"open"}`)}
	}

	// Held: nothing is written.
	_, _, err = s.Claim(ctx, "js-s", "sess", time.Minute)
	require.NoError(t, err)
	_, err = s.Apply(ctx, "test", store.Graph{
		Tickets: []store.NewTicket{batch("js-b1")}, Reparent: map[string]string{"js-run": "js-b1"},
		Edges: []domain.Edge{{From: "js-s", To: other.ID, Kind: domain.EdgeBlocks}}, Unclaimed: []string{"js-s"},
	})

	require.ErrorIs(t, err, domain.ErrNotHolder)
	_, err = s.Get(ctx, "js-b1")
	require.ErrorIs(t, err, domain.ErrNotFound)
	_, err = store.OpenBatch(ctx, s, proj, "k")
	require.ErrorIs(t, err, domain.ErrNotFound)

	out, err := s.Apply(ctx, "test", store.Graph{
		Tickets: []store.NewTicket{batch("js-b2")}, Reparent: map[string]string{"js-run": "js-b2"},
	})

	require.NoError(t, err)
	require.Len(t, out, 1)
	got, err := s.Get(ctx, "js-run")
	require.NoError(t, err)
	assert.Equal(t, "js-b2", got.ParentID)
	kids, err := store.Children(ctx, s, "js-b2")
	require.NoError(t, err)
	assert.Equal(t, []string{"js-run"}, ids(kids))
	b, err := store.OpenBatch(ctx, s, proj, "k")
	require.NoError(t, err)
	assert.Equal(t, "js-b2", b.ID)
	ready, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
	require.NoError(t, err)
	assert.NotContains(t, ids(ready), "js-b2", "batches are never claimable")
}
