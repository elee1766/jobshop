package storetest

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/knowledge"
	"github.com/gfx-labs/jobshop/pkg/orchestration/ingest"
	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
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
	require.NoError(t, s.SetResources(ctx, domain.ResourceConfig{Caps: map[string]int{"emu": 2}, Providers: []string{"dev"}}))
	_, err := addSlot(ctx, s, "dev", "d1", map[string]string{"X": "1"})
	require.NoError(t, err)

	var err2 error
	f.Run, err2 = s.NewID(proj)
	require.NoError(t, err2)
	f.Step, err2 = s.NewID(proj)
	require.NoError(t, err2)
	_, err = s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{
		{ID: f.Run, Project: proj, Kind: domain.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"w","state":"running","key":"k"}`)},
		{
			ID: f.Step, Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: f.Run, Title: "step",
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
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: f.Task, To: f.Blocker, Kind: domain.EdgeBlocks}))
	_, err = store.AddComment(ctx, s, "u", f.Task, "vendorHash mismatch after go.sum change")
	require.NoError(t, err)
	_, _, err = s.Claim(ctx, f.Step, "sess", lease)
	require.NoError(t, err)

	l, err := knowledge.AddLearning(ctx, s, domain.Learning{Kind: domain.LearningFailure, Summary: "flaky dial timeout", Fix: "rerun alone", Ticket: f.Task, Source: "agent"}, "")
	require.NoError(t, err)
	f.Learning = l.ID
	require.NoError(t, journal.Index(ctx, s, journal.Doc{Session: "sess", Ticket: f.Task, Project: proj, At: c.T, Text: "journal about nftables"}))
	_, err = knowledge.AddWriteEdit(ctx, s, domain.MemoryEdit{Path: "memory/global.md", Before: "old", Existed: true, Learnings: []int64{l.ID}, Pools: []string{"dev"}, MergedAt: c.T})
	require.NoError(t, err)
	it, _, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: "d1", Kind: "issues", RepoOrTeam: "a/b", Number: 7, ReceivedAt: c.T, Payload: json.RawMessage(`{"x":1}`)})
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
	assert.Equal(t, domain.StatusInProgress, step.Status)
	assert.Equal(t, "sess", step.Holder)
	ru, err := resourceUse(ctx, s)
	require.NoError(t, err)
	assert.NotEmpty(t, ru)
	h, err := store.History(ctx, s, f.Task)
	require.NoError(t, err)
	assert.NotEmpty(t, h)
	cs, err := store.Comments(ctx, s, f.Task)
	require.NoError(t, err)
	require.Len(t, cs, 1)

	hits, err := s.Recall(ctx, store.RecallQuery{Text: "vendorHash mismatch"})
	require.NoError(t, err)
	assert.Len(t, hits, 1, "comment search after the copy")
	jhits, err := journal.Recall(ctx, s, "nftables", "", "", 0)
	require.NoError(t, err)
	assert.Len(t, jhits, 1, "journal search after the copy")
	lhits, err := knowledge.RecallLearnings(ctx, s, "flaky dial", "", "", 0)
	require.NoError(t, err)
	assert.Len(t, lhits, 1, "learning search after the copy")
	lg, err := knowledge.GetLearning(ctx, s, f.Learning)
	require.NoError(t, err)
	assert.Equal(t, "rerun alone", lg.Fix)
	me, err := knowledge.ListEdits(ctx, s, "", 0)
	require.NoError(t, err)
	require.Len(t, me, 1)
	assert.Equal(t, "memory/global.md", me[0].Path)
	assert.True(t, me[0].Existed)
	in, err := ingest.InboxGet(ctx, s, f.Inbox)
	require.NoError(t, err)
	assert.JSONEq(t, `{"x":1}`, string(in.Payload))
	sl, err := allSlots(ctx, s)
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
	l2, err := knowledge.AddLearning(ctx, s, domain.Learning{Kind: domain.LearningFailure, Summary: "x", Source: "agent"}, "")
	require.NoError(t, err)
	assert.Greater(t, l2.ID, f.Learning)
	it, dup, err := ingest.InboxPut(ctx, s, domain.InboxItem{Source: "github", DeliveryID: "d2", Payload: json.RawMessage(`{}`)})
	require.NoError(t, err)
	assert.False(t, dup)
	assert.Greater(t, it.ID, f.Inbox)
	_, err = store.Create(ctx, s, "test", store.NewTicket{Project: proj, Kind: domain.KindTask, Title: "new"})
	require.NoError(t, err)
}
