package control_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/control/controltest"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/storage/store"
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
	assert.Equal(t, leaf.ID, graph.Root)
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
