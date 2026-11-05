package storetest

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/store"
)

// Filled names what Fill wrote, for CheckFilled.
type Filled struct {
	Run, Step, Task, Blocker string
	Learning                 int64
	Inbox                    int64
}

// Fill writes at least one row to every store table through the store
// API (#632): tickets with a parent, edges, labels, meta, history,
// comments, learnings, recall, memory edits, inbox, resources, slots
// and a collection record. Store copy tests fill a source with it and
// read the copy back with CheckFilled.
func Fill(t *testing.T, s store.Store, c *Clock) Filled {
	t.Helper()
	ctx := t.Context()
	var f Filled
	require.NoError(t, s.SetResourceCaps(ctx, map[string]int{"emu": 2}))
	require.NoError(t, s.SetResourceProviders(ctx, []string{"dev"}))
	_, err := s.AddSlot(ctx, "dev", "d1", map[string]string{"X": "1"})
	require.NoError(t, err)

	var err2 error
	f.Run, err2 = s.NewID(proj)
	require.NoError(t, err2)
	f.Step, err2 = s.NewID(proj)
	require.NoError(t, err2)
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{
		{ID: f.Run, Project: proj, Kind: store.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"w","state":"running","key":"k"}`)},
		{
			ID: f.Step, Project: proj, Kind: store.KindStep, Pool: "dev", ParentID: f.Run, Title: "step",
			Fields: json.RawMessage(fmt.Sprintf(`{"run":%q,"step":"s","index":0,"retries":2,"resources":["emu"]}`, f.Run)),
		},
	}, Edges: nil})

	require.NoError(t, err)
	c.Advance(time.Second)
	tk := task(t, s, "task", "dev", func(n *store.NewTicket) {
		n.Meta = json.RawMessage(`{"github":{"issue":"7"}}`)
		n.Labels = []string{"bug"}
		n.Body = "body text"
	})
	f.Task = tk.ID
	f.Blocker = task(t, s, "blocker", "dev").ID
	require.NoError(t, store.AddEdge(ctx, s, "test", store.Edge{From: f.Task, To: f.Blocker, Kind: store.EdgeBlocks}))
	_, err = store.AddComment(ctx, s, "u", f.Task, "vendorHash mismatch after go.sum change")
	require.NoError(t, err)
	_, _, err = s.Claim(ctx, f.Step, "sess", lease)
	require.NoError(t, err)

	l, err := learning.Add(ctx, s, store.Learning{Kind: store.LearningFailure, Summary: "flaky dial timeout", Fix: "rerun alone", Ticket: f.Task, Source: "agent"}, "")
	require.NoError(t, err)
	f.Learning = l.ID
	require.NoError(t, s.IndexRecall(ctx, store.RecallDoc{Kind: store.RecallJournal, Ref: "sess:1", TicketID: f.Task, Project: proj, At: c.T, Body: "journal about nftables"}))
	_, err = memedit.AddWrite(ctx, s, store.MemoryEdit{Path: "memory/global.md", Before: "old", Existed: true, Learnings: []int64{l.ID}, Pools: []string{"dev"}, MergedAt: c.T})
	require.NoError(t, err)
	it, _, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: "d1", Kind: "issues", RepoOrTeam: "a/b", Number: 7, ReceivedAt: c.T, Payload: json.RawMessage(`{"x":1}`)})
	require.NoError(t, err)
	f.Inbox = it.ID

	col, err := s.Collection(ctx, TestSpec)
	require.NoError(t, err)
	_, _, err = col.Put(ctx, store.Record{Key: "r1", State: "pending", Value: json.RawMessage(`{"source":"github","summary":"cache miss","natural_key":"n1"}`)})
	require.NoError(t, err)
	return f
}

// CheckFilled reads back what Fill wrote, on a copy of the filled store,
// and checks that the copy accepts new writes (identity sequences moved).
func CheckFilled(t *testing.T, s store.Store, f Filled) {
	t.Helper()
	ctx := t.Context()
	tk, err := s.Get(ctx, f.Task)
	require.NoError(t, err)
	assert.Equal(t, "body text", tk.Body)
	assert.Equal(t, []string{"bug"}, tk.Labels)
	l, err := s.List(ctx, store.Filter{MetaPack: "github", MetaKey: "issue", MetaValue: "7"})
	require.NoError(t, err)
	assert.Equal(t, []string{f.Task}, ids(l))
	r, err := s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true})
	require.NoError(t, err)
	assert.Equal(t, []string{f.Blocker}, ids(r), "task blocked by its edge, step held")
	kids, err := store.Children(ctx, s, f.Run)
	require.NoError(t, err)
	assert.Equal(t, []string{f.Step}, ids(kids))
	step, err := s.Get(ctx, f.Step)
	require.NoError(t, err)
	assert.Equal(t, store.StatusInProgress, step.Status)
	assert.Equal(t, "sess", step.Holder)
	ru, err := s.Resources(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, ru)
	h, err := store.History(ctx, s, f.Task)
	require.NoError(t, err)
	assert.NotEmpty(t, h)
	cs, err := store.Comments(ctx, s, f.Task)
	require.NoError(t, err)
	require.Len(t, cs, 1)

	for _, q := range []string{"vendorHash mismatch", "nftables"} {
		hits, err := s.Recall(ctx, store.RecallQuery{Text: q})
		require.NoError(t, err)
		assert.Len(t, hits, 1, q)
	}
	lhits, err := learning.Recall(ctx, s, "flaky dial", "", "", 0)
	require.NoError(t, err)
	assert.Len(t, lhits, 1, "learning search after the copy")
	lg, err := learning.Get(ctx, s, f.Learning)
	require.NoError(t, err)
	assert.Equal(t, "rerun alone", lg.Fix)
	me, err := memedit.List(ctx, s, "", 0)
	require.NoError(t, err)
	require.Len(t, me, 1)
	assert.Equal(t, "memory/global.md", me[0].Path)
	assert.True(t, me[0].Existed)
	in, err := ingest.InboxGet(ctx, s, f.Inbox)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":1}`, string(in.Payload))
	sl, err := s.Slots(ctx)
	require.NoError(t, err)
	require.Len(t, sl, 1)
	assert.Equal(t, map[string]string{"X": "1"}, sl[0].Env)

	col, err := s.Collection(ctx, TestSpec)
	require.NoError(t, err)
	rec, err := col.Get(ctx, "r1")
	require.NoError(t, err)
	assert.Equal(t, "pending", rec.State)
	found, err := col.Search(ctx, "cache", 0)
	require.NoError(t, err)
	assert.Len(t, found, 1)

	// New rows get ids past the copied ones.
	c2, err := store.AddComment(ctx, s, "u", f.Task, "after the copy")
	require.NoError(t, err)
	assert.Greater(t, c2.ID, cs[0].ID)
	l2, err := learning.Add(ctx, s, store.Learning{Kind: store.LearningFailure, Summary: "x", Source: "agent"}, "")
	require.NoError(t, err)
	assert.Greater(t, l2.ID, f.Learning)
	it, dup, err := ingest.InboxPut(ctx, s, store.InboxItem{Source: "github", DeliveryID: "d2", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
	assert.False(t, dup)
	assert.Greater(t, it.ID, f.Inbox)
	_, err = store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: store.KindTask, Title: "new"})
	require.NoError(t, err)
}
