package memedit_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/memedit"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func open(t *testing.T, now *time.Time) *sqlite.Store {
	t.Helper()
	s, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "j.db"), sqlite.Options{Now: func() time.Time { return *now }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestAddListResult: PR edits dedupe by PR, writes always add, ids count
// up even within one millisecond, List sorts newest merge first and
// SetResult stores the decision.
func TestAddListResult(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)

	_, err := memedit.AddPR(ctx, s, store.MemoryEdit{MergedAt: now})
	require.ErrorIs(t, err, store.ErrInvalid)
	_, err = memedit.AddWrite(ctx, s, store.MemoryEdit{MergedAt: now})
	require.ErrorIs(t, err, store.ErrInvalid)

	pr, err := memedit.AddPR(ctx, s, store.MemoryEdit{PR: " https://x/pull/1 ", Learnings: []int64{4}, MergedAt: now.Add(-time.Hour)})
	require.NoError(t, err)
	assert.Equal(t, int64(1), pr.ID)
	assert.Equal(t, "https://x/pull/1", pr.PR)
	assert.Equal(t, store.EditPending, pr.State)
	again, err := memedit.AddPR(ctx, s, store.MemoryEdit{PR: "https://x/pull/1", MergedAt: now})
	require.NoError(t, err)
	assert.Equal(t, pr.ID, again.ID, "same PR: the first edit")
	assert.True(t, again.MergedAt.Equal(now.Add(-time.Hour)))

	// Same millisecond: ids still count up past 9 (text order is "10" < "9").
	var last store.MemoryEdit
	for range 10 {
		last, err = memedit.AddWrite(ctx, s, store.MemoryEdit{Path: "memory/global.md", Before: "b", Existed: true, MergedAt: now})
		require.NoError(t, err)
	}
	assert.Equal(t, int64(11), last.ID)
	assert.Empty(t, last.Learnings)

	all, err := memedit.List(ctx, s, "", 0)
	require.NoError(t, err)
	require.Len(t, all, 11)
	assert.Equal(t, int64(11), all[0].ID, "newest merge, then highest id")
	assert.Equal(t, int64(1), all[10].ID)

	now = now.Add(time.Minute)
	u, err := memedit.SetResult(ctx, s, 1, store.EditHelped, store.PoolMetrics{Runs: 3}, "fewer retries", now)
	require.NoError(t, err)
	assert.Equal(t, store.EditHelped, u.State)
	assert.Equal(t, 3, u.After.Runs)
	assert.True(t, u.EvaluatedAt.Equal(now))
	assert.Equal(t, []int64{4}, u.Learnings)
	_, err = memedit.SetResult(ctx, s, 1, "maybe", store.PoolMetrics{}, "", now)
	require.ErrorIs(t, err, store.ErrInvalid)
	_, err = memedit.SetResult(ctx, s, 99, store.EditHelped, store.PoolMetrics{}, "", now)
	require.ErrorIs(t, err, store.ErrNotFound)

	pending, err := memedit.List(ctx, s, store.EditPending, 0)
	require.NoError(t, err)
	assert.Len(t, pending, 10)
	got, err := memedit.Get(ctx, s, 1)
	require.NoError(t, err)
	assert.Equal(t, "fewer retries", got.Reason)
	assert.Equal(t, "https://x/pull/1", memedit.Ref(got))
	assert.Equal(t, "memory edit 11, memory/global.md", memedit.Ref(last))
}
