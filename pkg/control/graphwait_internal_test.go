package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

func graphOf(t *testing.T, c *Controller, id string) rpc.GraphResult {
	t.Helper()
	out, err := c.ticketGraph(t.Context(), scope{admin: true, actor: "admin"}, rpc.GraphParams{ID: id})
	require.NoError(t, err)
	return out.(rpc.GraphResult)
}

func waitRelations(g rpc.GraphResult) []rpc.GraphRelation {
	var out []rpc.GraphRelation
	for _, r := range g.Relations {
		if r.Kind == "wait" {
			out = append(out, r)
		}
	}
	return out
}

// TestGraphWaitEdges (#1091): a parked ticket has a wait edge to its
// target. A run wait points at the run's node, an external wait at an
// external node with its canonical URL. The edge says live or why it
// ended, and survives the wait being gone.
func TestGraphWaitEdges(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: []byte(`{"workflow":"w","state":"running"}`)})
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	prs.set(recPR, `[[]]`, recChecks)
	_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: a.ID, Reason: "CI and run", Any: true,
		Conds: []domain.WaitCond{
			{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success", Workflow: "ci"},
			{Type: domain.WaitRun, Run: run.ID, State: "done"},
			{Type: domain.WaitLinearState, Thread: "JOB-150", State: "completed"},
			{Type: domain.WaitTime, At: c.now().Add(1)},
		},
	})
	require.NoError(t, err)

	g := graphOf(t, c, a.ID)
	got := map[string]rpc.GraphRelation{}
	for _, r := range waitRelations(g) {
		got[r.To] = r
	}
	require.Len(t, got, 3, "a time condition has nothing to point at")
	runRel := got[run.ID]
	assert.Equal(t, rpc.GraphRelation{From: a.ID, To: run.ID, Kind: "wait", Type: domain.WaitRun, Reason: "CI and run", State: "waiting", Since: runRel.Since}, runRel)
	assert.False(t, runRel.Since.IsZero())
	ext := map[string]rpc.GraphExternal{}
	for _, e := range g.External {
		ext[e.ID] = e
	}
	checks := ext["ext:github.checks:acme/widget PR #5 checks: success (ci)"]
	assert.Equal(t, "https://github.com/acme/widget/pull/5", checks.URL)
	assert.Equal(t, runRel.Since, checks.At)
	linear := ext["ext:linear.state:Linear JOB-150: completed"]
	assert.Equal(t, "Linear JOB-150: completed", linear.Label)
	assert.Empty(t, linear.URL, "no configured workspace covers JOB")
	var nodes []string
	for _, n := range g.Nodes {
		nodes = append(nodes, n.ID)
	}
	assert.ElementsMatch(t, []string{a.ID, run.ID}, nodes, "the run is a node, externals are not")

	// A hand clear leaves the edges, marked cleared.
	_, err = c.waitClear(ctx, admin, rpc.WaitClearParams{ID: a.ID, Reason: "not needed"})
	require.NoError(t, err)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	for _, r := range waitRelations(graphOf(t, c, a.ID)) {
		assert.Equal(t, "cleared", r.State, r.To)
	}

	// A new wait on the same target replaces the cleared edge. The head
	// moving invalidates it.
	_, err = c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: a.ID, Reason: "approval",
		Conds: []domain.WaitCond{{Type: domain.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob", Head: "aaaa1111bbbb"}},
	})
	require.NoError(t, err)
	prs.set(`{"state":"open","merged":false,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	require.Nil(t, waitOf(mustGet(t, c, a.ID)))
	states := map[string]string{}
	for _, r := range waitRelations(graphOf(t, c, a.ID)) {
		states[r.Type+" "+r.Reason] = r.State
	}
	assert.Equal(t, "invalidated", states["github.review approval"])
	assert.Equal(t, "cleared", states["run CI and run"])

	// A wait that is met says so.
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	_, err = c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: b.ID, Reason: "merge",
		Conds: []domain.WaitCond{{Type: domain.WaitPR, Repo: "acme/widget", PR: 5, State: "merged"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "waiting", waitRelations(graphOf(t, c, b.ID))[0].State)
	prs.set(`{"state":"closed","merged":true,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	require.Nil(t, waitOf(mustGet(t, c, b.ID)))
	rels := waitRelations(graphOf(t, c, b.ID))
	require.Len(t, rels, 1)
	assert.Equal(t, "met", rels[0].State)
}
