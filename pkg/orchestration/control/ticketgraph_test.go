package control_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

func TestTicketGraph(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, "projects:\n  js: {}\n")
	// The parent has more children than List's default page, the last
	// one closed. They are written in one transaction: one ticket.create
	// each took 3-12s under -race (JOB-117).
	const parentID = "js-plan"
	last := domain.Ticket{ID: fmt.Sprintf("js-sibling%d", store.DefaultLimit)}
	seedStore(t, root, func(st store.Store) {
		g := store.Graph{Tickets: []store.NewTicket{{ID: parentID, Project: "js", Kind: domain.KindTask, Title: "Plan"}}, Close: map[string]string{last.ID: "done"}}
		for i := 0; i <= store.DefaultLimit; i++ {
			g.Tickets = append(g.Tickets, store.NewTicket{ID: fmt.Sprintf("js-sibling%d", i), Project: "js", Kind: domain.KindTask, Title: fmt.Sprintf("Sibling %d", i), ParentID: parentID})
		}
		_, err := st.Apply(t.Context(), "test", g)
		require.NoError(t, err)
	})
	c := controltest.Start(t, root, control.Options{})
	ctx := t.Context()
	create := func(p rpc.CreateParams) domain.Ticket {
		t.Helper()
		var ticket domain.Ticket
		require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketCreate, p, &ticket))
		return ticket
	}
	parent := domain.Ticket{ID: parentID}
	child := create(rpc.CreateParams{Title: "Implement", ParentID: parent.ID})
	leaf := child
	for i := range 6 {
		leaf = create(rpc.CreateParams{Title: fmt.Sprintf("Step %d", i), ParentID: leaf.ID})
	}
	dependency := create(rpc.CreateParams{Title: "Prerequisite", Body: "https://linear.app/acme/issue/ENG-2", Meta: json.RawMessage(`{"github":{"repo":"acme/widget","pr":"42"},"linear":{"workspace":"acme","issue":"ENG-2"}}`)})
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: dependency.ID}, nil))
	dependent := create(rpc.CreateParams{Title: "Downstream"})
	for _, edge := range []rpc.EdgeParams{
		{From: child.ID, To: dependency.ID, Kind: domain.EdgeBlocks},
		{From: dependent.ID, To: child.ID, Kind: domain.EdgeWaitsFor},
		{From: dependency.ID, To: parent.ID, Kind: domain.EdgeTracks},
	} {
		require.NoError(t, c.Admin.Call(ctx, rpc.MethodEdgeAdd, edge, nil))
	}
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodCommentAdd, rpc.CommentParams{ID: dependency.ID, Body: "See https://github.com/acme/widget/issues/9"}, nil))
	var graph rpc.GraphResult
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: leaf.ID, Limit: 2000}, &graph))
	assert.Equal(t, parent.ID, graph.Root, "the top ticket, not the one asked for (JOB-190)")
	assert.Equal(t, leaf.ID, graph.Requested)
	assert.False(t, graph.Truncated)
	nodes := map[string]rpc.GraphNode{}
	for _, node := range graph.Nodes {
		require.NotContains(t, nodes, node.ID, "shared dependencies and cycles appear once")
		nodes[node.ID] = node
	}
	assert.Len(t, nodes, store.DefaultLimit+11)
	assert.Equal(t, domain.StatusClosed, nodes[last.ID].Status)
	assert.Equal(t, domain.StatusClosed, nodes[dependency.ID].Status)
	assert.Contains(t, graph.Relations, rpc.GraphRelation{From: dependent.ID, To: child.ID, Kind: "waits_for"})
	assert.Contains(t, graph.Relations, rpc.GraphRelation{From: parent.ID, To: child.ID, Kind: "parent"})
	assert.ElementsMatch(t, []string{"https://github.com/acme/widget/pull/42", "https://linear.app/acme/issue/ENG-2", "https://github.com/acme/widget/issues/9"}, nodes[dependency.ID].Links)
	var bounded rpc.GraphResult
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: child.ID, Limit: 2}, &bounded))
	assert.Len(t, bounded.Nodes, 2)
	assert.True(t, bounded.Truncated)
	for _, edge := range bounded.Relations {
		assert.Condition(t, func() bool {
			return (edge.From == bounded.Nodes[0].ID || edge.From == bounded.Nodes[1].ID) && (edge.To == bounded.Nodes[0].ID || edge.To == bounded.Nodes[1].ID)
		})
	}
}

func TestTicketGraphWorkflowReference(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, `projects:
  js: {}
workflows:
  review:
    vars: [subject, repo, pr, issue]
    repo: oku-trade/trade
    steps:
      - id: route
        type: command
        run: 'true'
        title: 'Route review of oku-trade/trade#{{.subject}}'
`)
	c := controltest.Start(t, root, control.Options{})
	ctx := t.Context()
	var run rpc.RunResult
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodRunStart, rpc.RunStartParams{
		Workflow: "review", Vars: map[string]string{"subject": "4520@252b4e7f", "repo": "oku-trade/trade", "pr": "4520", "issue": "OKU-219"}, Key: "first",
	}, &run))
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodRunShow, rpc.IDParams{ID: run.Run.ID}, &run))
	assert.Contains(t, run.Links, "https://github.com/oku-trade/trade/pull/4520")
	assert.Len(t, run.Links, 1, "an issue key cannot determine a Linear workspace")
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketUpdate, rpc.UpdateParams{ID: run.Run.ID, Meta: json.RawMessage(`{"linear":{"workspace":"acme"}}`)}, nil))
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodRunShow, rpc.IDParams{ID: run.Run.ID}, &run))
	assert.Contains(t, run.Links, "https://linear.app/acme/issue/OKU-219")
	var wrong domain.Ticket
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketCreate, rpc.CreateParams{Title: "Other oku-trade/trade#45200@abc"}, &wrong))
	var graph rpc.GraphResult
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: "oku-trade/trade#4520"}, &graph))
	assert.Equal(t, run.Run.ID, graph.Root)
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodRunStart, rpc.RunStartParams{
		Workflow: "review", Vars: map[string]string{"subject": "4520@newhead", "repo": "oku-trade/trade", "pr": "4520", "issue": "OKU-219"}, Key: "second",
	}, &run))
	err := c.Admin.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: "oku-trade/trade#4520"}, &graph)
	require.ErrorContains(t, err, "multiple roots")
}

// TestTicketGraphWhole (JOB-190): every member of a graph returns the same
// tickets, relations and root, whichever one is asked for. A session that
// holds a step of run A and creates run B connects B to that step, and a
// ticket made while two were held hangs from both.
func TestTicketGraphWhole(t *testing.T) {
	t.Parallel()
	root := controltest.NewRoot(t, "default_project: js\nprojects:\n  js: {}\n  ox: {}\n")
	c := controltest.Start(t, root, control.Options{})
	ctx := t.Context()
	create := func(by string, p rpc.CreateParams) domain.Ticket {
		t.Helper()
		var ticket domain.Ticket
		require.NoError(t, c.Admin.Call(rpc.WithActor(ctx, by), rpc.MethodTicketCreate, p, &ticket))
		// The fake clock only moves when told: ticket order is explicit.
		c.Clock.Advance(time.Second)
		return ticket
	}
	graph := func(cl *rpc.Client, id string) rpc.GraphResult {
		t.Helper()
		var g rpc.GraphResult
		require.NoError(t, cl.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: id, Limit: 2000}, &g))
		return g
	}
	ids := func(g rpc.GraphResult) []string {
		var out []string
		for _, n := range g.Nodes {
			out = append(out, n.ID)
		}
		return out
	}

	var sess rpc.SessionNewResult
	require.NoError(t, c.Admin.Call(ctx, rpc.MethodSessionNew, rpc.SessionNewParams{Agent: "jcode"}, &sess))
	s := sess.Ticket.ID
	sock := c.Dial(sess.Socket)

	runA := create("admin", rpc.CreateParams{Title: "Run A"})
	a1 := create("admin", rpc.CreateParams{Title: "A step 1", ParentID: runA.ID})
	a2 := create("admin", rpc.CreateParams{Title: "A step 2", ParentID: runA.ID})
	// The session holds a step of A and starts B and a step of B.
	claimAs(t, c.Admin, a1.ID, s)
	runB := create(s, rpc.CreateParams{Title: "Run B"})
	b1 := create("admin", rpc.CreateParams{Title: "B step 1", ParentID: runB.ID})
	// Holding two, it makes a third.
	claimAs(t, c.Admin, a2.ID, s)
	both := create(s, rpc.CreateParams{Title: "Made while holding two"})
	// A ticket from another project made in the same window.
	foreign := create(s, rpc.CreateParams{Title: "Other project", Project: "ox"})
	// Both holds end: later tickets of the same session are not pulled in.
	for _, id := range []string{a1.ID, a2.ID} {
		require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketClose, rpc.CloseParams{ID: id}, nil))
		c.Clock.Advance(time.Second)
	}
	outside := create(s, rpc.CreateParams{Title: "Made after the holds"})

	want := graph(c.Admin, runA.ID)
	assert.Equal(t, runA.ID, want.Root)
	assert.Equal(t, runA.ID, want.Requested)
	assert.ElementsMatch(t, []string{runA.ID, a1.ID, a2.ID, runB.ID, b1.ID, both.ID, foreign.ID}, ids(want))
	assert.NotContains(t, ids(want), outside.ID, "outside any hold window")
	assert.Contains(t, want.Relations, rpc.GraphRelation{From: a1.ID, To: runB.ID, Kind: "created"})
	for _, held := range []string{a1.ID, a2.ID} {
		assert.Contains(t, want.Relations, rpc.GraphRelation{From: held, To: both.ID, Kind: "created"})
	}

	for _, ask := range []domain.Ticket{runB, b1, a1, a2, both, foreign} {
		got := graph(c.Admin, ask.ID)
		assert.Equal(t, runA.ID, got.Root, "root from %s", ask.Title)
		assert.Equal(t, ask.ID, got.Requested)
		assert.ElementsMatch(t, ids(want), ids(got), "tickets from %s", ask.Title)
		assert.ElementsMatch(t, want.Relations, got.Relations, "relations from %s", ask.Title)
	}
	// Nothing above the outside ticket: it is its own graph.
	alone := graph(c.Admin, outside.ID)
	assert.Equal(t, outside.ID, alone.Root)
	assert.Equal(t, []string{outside.ID}, ids(alone))

	// The timeline lists the same rows from B as from A.
	timeline := func(id string) rpc.TimelineResult {
		t.Helper()
		var tl rpc.TimelineResult
		require.NoError(t, c.Admin.Call(ctx, rpc.MethodTicketTimeline, rpc.GraphParams{ID: id, Limit: 2000}, &tl))
		return tl
	}
	rowIDs := func(tl rpc.TimelineResult) []string {
		var out []string
		for _, r := range tl.Rows {
			out = append(out, r.ID)
		}
		return out
	}
	fromA, fromB := timeline(runA.ID), timeline(runB.ID)
	assert.Equal(t, runA.ID, fromB.Root)
	assert.Equal(t, runB.ID, fromB.Requested)
	assert.Equal(t, rowIDs(fromA), rowIDs(fromB))
	assert.Len(t, fromA.Rows, 7)

	// A caller scoped to a project gets nothing from another.
	scoped := graph(sock, runB.ID)
	assert.NotContains(t, ids(scoped), foreign.ID)
	assert.Equal(t, runA.ID, scoped.Root)
	err := sock.Call(ctx, rpc.MethodTicketGraph, rpc.GraphParams{ID: foreign.ID}, &rpc.GraphResult{})
	require.ErrorIs(t, err, rpc.ErrForbidden)
}
