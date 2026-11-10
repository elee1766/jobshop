package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func timelineOf(t *testing.T, c *Controller, id string) rpc.TimelineResult {
	t.Helper()
	out, err := c.ticketTimeline(t.Context(), scope{admin: true, actor: "admin"}, rpc.GraphParams{ID: id})
	require.NoError(t, err)
	return out.(rpc.TimelineResult)
}

func rowOf(t *testing.T, tl rpc.TimelineResult, id string) rpc.TimelineRow {
	t.Helper()
	for _, r := range tl.Rows {
		if r.ID == id {
			return r
		}
	}
	require.Failf(t, "no row", id)
	return rpc.TimelineRow{}
}

func segStates(r rpc.TimelineRow) []string {
	var out []string
	for _, s := range r.Segments {
		out = append(out, s.State)
	}
	return out
}

// TestTimelineSegments (#1103): a step's bar is replayed from its history:
// queued, a claim, a wait that ended met, a failed attempt and the claim
// that finished it. Bars are contiguous and a closed ticket's ends at its
// close.
func TestTimelineSegments(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	ctx := t.Context()
	s1, s2 := mkSession(t, c), mkSession(t, c)
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step := func(name string, idx int) domain.Ticket {
		return mkTicket(t, c, store.NewTicket{
			Kind: domain.KindStep, Title: name, Pool: "dev", ParentID: run.ID,
			Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"` + name + `","index":` + string(rune('0'+idx)) + `,"retries":3,"failures":0}`),
		})
	}
	a, b := step("implement", 0), step("review", 1)
	require.NoError(t, store.AddEdge(ctx, c.store, "admin", domain.Edge{From: b.ID, To: a.ID, Kind: domain.EdgeBlocks}))
	t0 := c.now()

	clk.advance(time.Minute)
	_, _, err := c.store.Claim(ctx, a.ID, s1.ID, time.Hour)
	require.NoError(t, err)
	clk.advance(2 * time.Minute)
	prs.set(recPR, `[[]]`, recChecks)
	_, err = c.waitAdd(ctx, sessScope(s1), rpc.WaitAddParams{
		Conds:  []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success", Required: true}},
		Reason: "CI on PR 5", Resume: true,
	})
	require.NoError(t, err)

	mid := rowOf(t, timelineOf(t, c, run.ID), a.ID)
	assert.Equal(t, []string{"queued", "working", "waiting"}, segStates(mid), "a live wait is a running segment")
	assert.True(t, mid.Segments[2].End.IsZero())
	assert.Equal(t, "CI on PR 5", mid.Segments[2].Detail)
	require.Len(t, mid.Segments[2].Wait, 1)
	assert.Equal(t, "waiting", mid.Segments[2].Wait[0].State)
	assert.Contains(t, mid.Segments[2].Wait[0].URL, "acme/widget/pull/5")

	clk.advance(5 * time.Minute)
	prs.set(recPR, `[[]]`, `[{"name":"build","bucket":"pass"}]`)
	c.pollWaits(ctx, prs, threads)
	clk.advance(time.Minute)
	_, cl, err := c.store.Claim(ctx, a.ID, s1.ID, time.Hour)
	require.NoError(t, err)
	clk.advance(time.Minute)
	_, err = c.store.Release(ctx, "", a.ID, &cl, store.Outcome{Reason: "tests failed"})
	require.NoError(t, err)
	_, cl, err = c.store.Claim(ctx, a.ID, s2.ID, time.Hour)
	require.NoError(t, err)
	clk.advance(2 * time.Minute)
	_, err = store.Close(ctx, c.store, "", a.ID, &cl, "done")
	require.NoError(t, err)

	tl := timelineOf(t, c, run.ID)
	row := rowOf(t, tl, a.ID)
	assert.Equal(t, []string{"queued", "working", "waiting", "queued", "retry", "working"}, segStates(row))
	assert.Equal(t, s1.ID, row.Segments[1].Actor)
	assert.Equal(t, "met", row.Segments[2].Ended)
	assert.Equal(t, "met", row.Segments[2].Wait[0].State)
	assert.Equal(t, "tests failed", row.Segments[4].Detail)
	assert.Equal(t, s1.ID, row.Segments[4].Actor)
	assert.Equal(t, s2.ID, row.Segments[5].Actor)
	assert.Equal(t, row.CreatedAt, row.Segments[0].Start)
	assert.WithinDuration(t, t0, row.CreatedAt, time.Millisecond)
	for i := 1; i < len(row.Segments); i++ {
		assert.Equal(t, row.Segments[i-1].End, row.Segments[i].Start, "segment %d follows the one before", i)
	}
	assert.Equal(t, row.ClosedAt, row.Segments[len(row.Segments)-1].End)
	assert.False(t, row.Partial)
	assert.Equal(t, row.Segments[1].Start, row.Started)
	assert.Contains(t, markerKinds(row), rpc.MarkerSession, "s2 took over from s1")

	// Nesting: the run is the root, its steps are one level down.
	assert.Equal(t, run.ID, tl.Rows[0].ID)
	assert.Equal(t, 0, tl.Rows[0].Depth)
	assert.Equal(t, 1, row.Depth)
	assert.Equal(t, run.ID, row.Parent)
	assert.Equal(t, []string{"working"}, segStates(tl.Rows[0]), "an open run is working")

	// B is blocked by A: A closed before B started, so the edge is met.
	assert.Contains(t, tl.Arrows, rpc.TimelineArrow{From: a.ID, To: b.ID, Kind: "blocks", State: rpc.ArrowMet})
	assert.Equal(t, []string{"queued"}, segStates(rowOf(t, tl, b.ID)))
}

func markerKinds(r rpc.TimelineRow) []string {
	var out []string
	for _, m := range r.Markers {
		out = append(out, m.Kind)
	}
	return out
}

// TestTimelineHumanAndConflict (#1103): a needs-human label is its own
// segment from creation to removal, and a blocking edge added after the
// blocked ticket started is a conflict.
func TestTimelineHumanAndConflict(t *testing.T) {
	t.Parallel()
	c, clk, _, _ := waitController(t)
	ctx := t.Context()
	s1 := mkSession(t, c)
	q := mkTicket(t, c, store.NewTicket{Title: "q", Pool: "dev", Labels: []string{needsHumanLabel, questionLabel}})
	clk.advance(time.Minute)
	require.NoError(t, store.RemoveLabel(ctx, c.store, "admin", q.ID, questionLabel))
	clk.advance(time.Minute)
	require.NoError(t, store.RemoveLabel(ctx, c.store, "admin", q.ID, needsHumanLabel))
	clk.advance(time.Minute)

	blocker := mkTicket(t, c, store.NewTicket{Title: "blocker", Pool: "dev"})
	blocked := mkTicket(t, c, store.NewTicket{Title: "blocked", Pool: "dev"})
	clk.advance(time.Minute)
	_, _, err := c.store.Claim(ctx, blocked.ID, s1.ID, time.Hour)
	require.NoError(t, err)
	require.NoError(t, store.AddEdge(ctx, c.store, "admin", domain.Edge{From: blocked.ID, To: blocker.ID, Kind: domain.EdgeBlocks}))
	require.NoError(t, store.AddEdge(ctx, c.store, "admin", domain.Edge{From: q.ID, To: blocker.ID, Kind: domain.EdgeWaitsFor}))

	tl := timelineOf(t, c, blocked.ID)
	assert.Contains(t, tl.Arrows, rpc.TimelineArrow{From: blocker.ID, To: blocked.ID, Kind: "blocks", State: rpc.ArrowConflict})
	assert.Contains(t, tl.Arrows, rpc.TimelineArrow{From: blocker.ID, To: q.ID, Kind: "waits_for", State: rpc.ArrowBlocking})
	row := rowOf(t, tl, q.ID)
	assert.Equal(t, []string{"human", "queued"}, segStates(row), "the human span lasts until the last label goes")
	assert.Equal(t, needsHumanLabel, row.Segments[0].Detail)
	assert.WithinDuration(t, row.CreatedAt.Add(2*time.Minute), row.Segments[1].Start, time.Millisecond)
	assert.Equal(t, []string{"queued", "working"}, segStates(rowOf(t, tl, blocked.ID)))
}

// TestTimelineDeferral (#1103): a release until a time is a deferred span
// that ends at that time.
func TestTimelineDeferral(t *testing.T) {
	t.Parallel()
	c, clk, _, _ := waitController(t)
	ctx := t.Context()
	s1 := mkSession(t, c)
	task := mkTicket(t, c, store.NewTicket{Title: "later", Pool: "dev"})
	_, cl, err := c.store.Claim(ctx, task.ID, s1.ID, time.Hour)
	require.NoError(t, err)
	clk.advance(time.Minute)
	until := c.now().Add(10 * time.Minute)
	_, err = c.store.Release(ctx, "", task.ID, &cl, store.Outcome{Reason: "not yet", Until: until})
	require.NoError(t, err)
	clk.advance(30 * time.Minute)

	row := rowOf(t, timelineOf(t, c, task.ID), task.ID)
	assert.Equal(t, []string{"working", "deferred", "queued"}, segStates(row))
	assert.Equal(t, until, row.Segments[1].End)
	assert.Equal(t, until, row.Segments[1].Until)
	assert.True(t, row.Segments[2].End.IsZero())

	// Not a retry: a deferral is not a failed attempt.
	assert.NotContains(t, segStates(row), "retry")
}

// TestTimelinePartial (#1103): a ticket with no history gets an unknown
// span up to now and is marked partial.
func TestTimelinePartial(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	tk := domain.Ticket{ID: "js-e", Kind: domain.KindTask, Status: domain.StatusOpen, CreatedAt: now.Add(-time.Hour)}
	row := timelineRow(&config.Config{}, tk, store.Thread{}, now, nil)
	assert.True(t, row.Partial)
	assert.Equal(t, []string{"unknown", "queued"}, segStates(row))
	assert.Equal(t, now, row.Segments[0].End)
}
