package storetest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

func testApplyRevisionedGraph(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	run, err := store.Create(ctx, s, "test", store.NewTicket{
		ID: "js-a-run", Project: proj, Kind: domain.KindRun,
		Fields: json.RawMessage(`{"workflow":"w","key":"atomic","state":"running"}`),
	})
	require.NoError(t, err)
	observed := task(t, s, "observed input", "dev")
	history, err := store.History(ctx, s, run.ID)
	require.NoError(t, err)
	graph := store.Graph{
		Revisions: map[string]int64{run.ID: run.Revision, observed.ID: observed.Revision},
		Tickets:   []store.NewTicket{{ID: "js-next", Project: proj, Kind: domain.KindTask, Title: "Follow-up"}},
		Fields:    map[string]json.RawMessage{run.ID: json.RawMessage(`{"workflow":"w","key":"atomic","state":"done"}`)},
		Close:     map[string]string{run.ID: "done"},
		Comments:  []store.GraphComment{{TicketID: "js-next", Body: "filed by hook"}, {TicketID: observed.ID, Body: "seen"}},
	}
	assertUnchanged := func() {
		t.Helper()
		_, err := s.Get(ctx, "js-next")
		require.ErrorIs(t, err, domain.ErrNotFound)
		got, err := s.Get(ctx, run.ID)
		require.NoError(t, err)
		assert.Equal(t, run.Status, got.Status)
		assert.Equal(t, run.Revision, got.Revision)
		assert.JSONEq(t, string(run.Fields), string(got.Fields))
		gotHistory, err := store.History(ctx, s, run.ID)
		require.NoError(t, err)
		assert.Equal(t, history, gotHistory)
		comments, err := store.Comments(ctx, s, observed.ID)
		require.NoError(t, err)
		assert.Empty(t, comments, "a failed graph adds no comments")
	}

	// A read-only input belongs to the read set, not just the tickets written.
	title := "changed input"
	observed, err = s.Update(ctx, "other", observed.ID, store.Update{Title: &title})
	require.NoError(t, err)
	_, err = s.Apply(ctx, "hook", graph)
	require.ErrorIs(t, err, domain.ErrConflict)
	assertUnchanged()

	graph.Revisions[observed.ID] = observed.Revision
	graph.Close["js-z-missing"] = "done"
	_, err = s.Apply(ctx, "hook", graph)
	require.ErrorIs(t, err, domain.ErrNotFound)
	assertUnchanged()

	delete(graph.Close, "js-z-missing")
	created, err := s.Apply(ctx, "hook", graph)
	require.NoError(t, err)
	require.Len(t, created, 1)
	assert.Equal(t, "js-next", created[0].ID)
	_, err = s.Get(ctx, created[0].ID)
	require.NoError(t, err)
	closed, err := s.Get(ctx, run.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, closed.Status)
	assert.Equal(t, "done", closed.Resolution)
	fields, err := domain.DecodeFields[domain.RunFields](closed)
	require.NoError(t, err)
	assert.Equal(t, domain.RunDone, fields.State)
	for id, body := range map[string]string{"js-next": "filed by hook", observed.ID: "seen"} {
		comments, err := store.Comments(ctx, s, id)
		require.NoError(t, err)
		require.Len(t, comments, 1)
		assert.Equal(t, body, comments[0].Body)
		assert.Equal(t, "hook", comments[0].Author)
	}
}

func testStructuredStepResultFences(t *testing.T, s store.Store, _ *Clock) {
	ctx := t.Context()
	_, err := s.Apply(ctx, "test", store.Graph{Tickets: []store.NewTicket{
		{ID: "js-run", Project: proj, Kind: domain.KindRun, Fields: json.RawMessage(`{"workflow":"w","key":"result","state":"running"}`)},
		{
			ID: "js-step", Project: proj, Kind: domain.KindStep, Pool: "dev", ParentID: "js-run",
			Fields: json.RawMessage(`{"run":"js-run","step":"implement","index":0,"retries":1,"outcome":"pending","structured_output":{"plan":"original"}}`),
		},
	}})
	require.NoError(t, err)
	_, stale, err := s.Claim(ctx, "js-step", "worker", lease)
	require.NoError(t, err)
	_, err = s.Release(ctx, "worker", "js-step", &stale, store.Outcome{Pool: "dev"})
	require.NoError(t, err)
	claimed, current, err := s.Claim(ctx, "js-step", "worker", lease)
	require.NoError(t, err)
	title := "updated while held"
	before, err := s.Update(ctx, "other", "js-step", store.Update{Title: &title})
	require.NoError(t, err)
	history, err := store.History(ctx, s, "js-step")
	require.NoError(t, err)
	result := &store.StepResult{Revision: before.Revision, Outcome: "implemented", Output: map[string]any{"plan": "new", "checks": []any{"build", "test"}}}
	assertUnchanged := func() {
		t.Helper()
		got, err := s.Get(ctx, "js-step")
		require.NoError(t, err)
		assert.Equal(t, domain.StatusInProgress, got.Status)
		assert.Equal(t, before.Revision, got.Revision)
		assert.Equal(t, current.Holder, got.Holder)
		assert.Equal(t, current.Fence, got.Fence)
		assert.JSONEq(t, string(before.Fields), string(got.Fields))
		gotHistory, err := store.History(ctx, s, "js-step")
		require.NoError(t, err)
		assert.Equal(t, history, gotHistory)
	}

	_, err = s.Release(ctx, "worker", "js-step", &stale, store.Outcome{Close: "done", StepResult: result})
	require.ErrorIs(t, err, domain.ErrNotHolder)
	assertUnchanged()
	result.Revision = claimed.Revision
	_, err = s.Release(ctx, "worker", "js-step", &current, store.Outcome{Close: "done", StepResult: result})
	require.ErrorIs(t, err, domain.ErrConflict)
	assertUnchanged()

	result.Revision = before.Revision
	_, err = s.Release(ctx, "worker", "js-step", &current, store.Outcome{Close: "done", StepResult: result})
	require.NoError(t, err)
	closed, err := s.Get(ctx, "js-step")
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, closed.Status)
	assert.Equal(t, "done", closed.Resolution)
	fields, err := domain.DecodeFields[domain.StepFields](closed)
	require.NoError(t, err)
	assert.Equal(t, "implemented", fields.Outcome)
	assert.Equal(t, map[string]any{"plan": "new", "checks": []any{"build", "test"}}, fields.StructuredOutput)

	result.Output = map[string]any{"plan": "duplicate"}
	_, err = s.Release(ctx, "worker", "js-step", &current, store.Outcome{Close: "done", StepResult: result})
	require.ErrorIs(t, err, domain.ErrClosed)
	got, err := s.Get(ctx, "js-step")
	require.NoError(t, err)
	assert.Equal(t, closed.Revision, got.Revision)
	assert.JSONEq(t, string(closed.Fields), string(got.Fields))
}
