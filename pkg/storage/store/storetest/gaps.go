package storetest

// Cases the JOB-36 audit found missing (JOB-37): wait fields, Recall
// filters, reopen readiness, tracks edges and the remaining List filters.

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
)

func waitJSON(t *testing.T, since time.Time) json.RawMessage {
	t.Helper()
	w := domain.TaskFields{Wait: &domain.Wait{
		Conds:  []domain.WaitCond{{Type: domain.WaitChecks, Repo: "o/r", PR: 1, State: "success"}},
		Mode:   domain.WaitAny,
		Reason: "ci",
		Since:  since,
	}}
	b, err := json.Marshal(w)
	require.NoError(t, err)
	return b
}

// testWaitGates: fields.wait keeps a ticket out of Ready, Unheld
// listings of ready work and ClaimNext; clearing it restores all three.
func testWaitGates(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev", func(n *store.NewTicket) { n.Fields = waitJSON(t, c.Now()) })

	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	f, err := domain.DecodeFields[domain.TaskFields](got)
	require.NoError(t, err)
	require.NotNil(t, f.Wait)
	assert.Equal(t, "ci", f.Wait.Reason)
	assert.Equal(t, "o/r", f.Wait.Conds[0].Repo)

	ready := func() []string {
		t.Helper()
		l, err := s.List(ctx, store.Filter{Project: proj, Pool: "dev", Ready: true, Unheld: true})
		require.NoError(t, err)
		return ids(l)
	}
	assert.Empty(t, ready())
	_, _, err = s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s1", Lease: lease})
	require.ErrorIs(t, err, domain.ErrNoWork)

	// An invalid wait is rejected.
	_, err = s.Update(ctx, "test", a.ID, store.Update{Fields: json.RawMessage(`{"wait":{"conds":[],"reason":"x"}}`)})
	require.ErrorIs(t, err, domain.ErrInvalid)

	_, err = s.Update(ctx, "test", a.ID, store.Update{Fields: json.RawMessage(`{}`)})
	require.NoError(t, err)
	assert.Equal(t, []string{a.ID}, ready())
	tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s1", Lease: lease})
	require.NoError(t, err)
	assert.Equal(t, a.ID, tk.ID)
}

// testReopenReady: a closed step or run reopens to open and the step is
// claimable again.
func testReopenReady(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run := store.NewTicket{ID: "js-run", Project: proj, Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"w","key":"k","state":"running"}`)}
	step := store.NewTicket{
		ID: "js-st", Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: "js-run",
		Fields: json.RawMessage(`{"run":"js-run","step":"s","index":0,"retries":1}`),
	}
	_, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{run, step}})
	require.NoError(t, err)
	_, cl, err := s.Claim(ctx, "js-st", "s1", lease)
	require.NoError(t, err)
	_, err = store.Close(ctx, s, "s1", "js-st", &cl, "done")
	require.NoError(t, err)
	_, err = store.Close(ctx, s, "test", "js-run", nil, "done")
	require.NoError(t, err)

	r, err := store.Reopen(ctx, s, "test", "js-run")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, r.Status)
	st, err := store.Reopen(ctx, s, "test", "js-st")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, st.Status)
	assert.Empty(t, st.Holder)

	l, err := s.List(ctx, store.Filter{Project: proj, Ready: true})
	require.NoError(t, err)
	assert.Equal(t, []string{"js-st"}, ids(l))
	tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s2", Lease: lease})
	require.NoError(t, err)
	assert.Equal(t, "js-st", tk.ID)

	_, err = store.Reopen(ctx, s, "test", "js-run")
	require.ErrorIs(t, err, domain.ErrInvalid, "reopen of an open ticket")
}

// testTracksDoesNotGate: a tracks edge to an open ticket leaves the
// tracker claimable.
func testTracksDoesNotGate(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "dev", prio(1))
	b := task(t, s, "b", "other")
	require.NoError(t, store.AddEdge(ctx, s, "test", domain.Edge{From: a.ID, To: b.ID, Kind: domain.EdgeTracks}))
	tk, _, err := s.ClaimNext(ctx, store.ClaimNext{Project: proj, Pool: "dev", Holder: "s1", Lease: lease})
	require.NoError(t, err)
	assert.Equal(t, a.ID, tk.ID)
}

// testListMoreFilters: Holder, ParentID, PoolNotIn and Statuses.
func testListMoreFilters(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	parent := task(t, s, "parent", "")
	kid := task(t, s, "kid", "dev", func(n *store.NewTicket) { n.ParentID = parent.ID })
	held := task(t, s, "held", "ops", func(n *store.NewTicket) { n.Holder = "s1" })
	done := task(t, s, "done", "ops")
	_, err := store.Close(ctx, s, "test", done.ID, nil, "done")
	require.NoError(t, err)
	_, _, err = s.Claim(ctx, kid.ID, "s2", lease)
	require.NoError(t, err)

	list := func(f store.Filter) []string {
		t.Helper()
		l, err := s.List(ctx, f)
		require.NoError(t, err)
		return ids(l)
	}
	assert.Equal(t, []string{held.ID}, list(store.Filter{Holder: "s1"}))
	assert.Equal(t, []string{kid.ID}, list(store.Filter{Holder: "s2"}))
	assert.Equal(t, []string{kid.ID}, list(store.Filter{ParentID: parent.ID}))
	assert.ElementsMatch(t, []string{parent.ID, kid.ID}, list(store.Filter{PoolNotIn: []string{"ops"}}))
	assert.ElementsMatch(t, []string{parent.ID}, list(store.Filter{PoolNotIn: []string{"ops", "dev"}}))
	assert.ElementsMatch(t, []string{kid.ID, done.ID}, list(store.Filter{Statuses: []domain.Status{domain.StatusInProgress, domain.StatusClosed}}))
	assert.ElementsMatch(t, []string{parent.ID, held.ID}, list(store.Filter{Statuses: []domain.Status{domain.StatusOpen}}))
}

// testRecallFilters: Project, TicketID, Limit (default 10), ranking and
// query text carrying search operators.
func testRecallFilters(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	a := task(t, s, "a", "")
	b, err := store.Create(ctx, s, "test", store.NewTicket{Project: "other", Kind: domain.KindTask, Title: "b"})
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "u", a.ID, "zebra failure in the build")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "u", a.ID, "zebra zebra zebra crossing zebra timeout")
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "u", b.ID, "zebra in another project")
	require.NoError(t, err)
	for i := range 12 {
		_, err = store.AddComment(ctx, s, "u", a.ID, fmt.Sprintf("quokka note %d", i))
		require.NoError(t, err)
	}

	recall := func(q store.RecallQuery) []store.RecallHit {
		t.Helper()
		h, err := s.Recall(ctx, q)
		require.NoError(t, err)
		return h
	}
	assert.Len(t, recall(store.RecallQuery{Text: "zebra"}), 3)
	h := recall(store.RecallQuery{Text: "zebra", Project: "other"})
	require.Len(t, h, 1)
	assert.Equal(t, b.ID, h[0].TicketID)
	h = recall(store.RecallQuery{Text: "zebra", TicketID: a.ID})
	require.Len(t, h, 2)
	for _, x := range h {
		assert.Equal(t, a.ID, x.TicketID)
	}
	h = recall(store.RecallQuery{Text: "zebra timeout"})
	require.NotEmpty(t, h)
	assert.Contains(t, h[0].Snippet, "timeout", "the comment matching every word ranks first")
	for i := 1; i < len(h); i++ {
		assert.LessOrEqual(t, h[i-1].Score, h[i].Score, "best match first")
	}

	assert.Len(t, recall(store.RecallQuery{Text: "quokka"}), 10, "default limit")
	assert.Len(t, recall(store.RecallQuery{Text: "quokka", Limit: 3}), 3)

	for _, q := range []string{`zebra "unmatched AND: & | OR NOT * ( ) -x`, `NEAR(zebra) ^col:x`, `"`} {
		_, err := s.Recall(ctx, store.RecallQuery{Text: q})
		require.NoError(t, err, q)
	}
	h = recall(store.RecallQuery{Text: `crossing "nothing-here" AND: |`})
	require.Len(t, h, 1, "any-word fallback")
}
