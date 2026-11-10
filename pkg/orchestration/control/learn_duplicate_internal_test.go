package control

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/knowledge"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestLearnDuplicate (JOB-199): duplicates link to the learning kept open,
// the list gives it a count, and the count rules (trigger, gate window)
// still see every occurrence.
func TestLearnDuplicate(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	add := func(summary string) int64 {
		t.Helper()
		out, err := c.learnAdd(ctx, admin, rpc.LearnParams{Kind: domain.LearningFailure, Summary: summary, Scope: "global"})
		require.NoError(t, err)
		clk.advance(time.Minute)
		return out.(domain.Learning).ID
	}
	update := func(p rpc.LearnUpdateParams) ([]domain.Learning, error) {
		t.Helper()
		out, err := c.learnUpdate(ctx, admin, p)
		if err != nil {
			return nil, err
		}
		return out.([]domain.Learning), nil
	}
	list := func(p rpc.LearnListParams) map[int64]domain.Learning {
		t.Helper()
		out, err := c.learnList(ctx, admin, p)
		require.NoError(t, err)
		m := map[int64]domain.Learning{}
		for _, l := range out.([]domain.Learning) {
			m[l.ID] = l
		}
		return m
	}

	start := clk.Now()
	kept := add("step review-merge closed failed")
	d1 := add("step review-merge closed failed")
	d2 := add("step review-merge closed failed")
	other := add("something else")
	applied := add("already handled")
	_, err := update(rpc.LearnUpdateParams{IDs: []int64{applied}, Status: domain.LearningApplied})
	require.NoError(t, err)
	// A window over every learning, measured before and after marking.
	pools := []string{"dev"}
	before, err := knowledge.Measure(ctx, c.store, nil, pools, start.Add(-time.Hour), clk.Now(), false)
	require.NoError(t, err)
	countBefore, err := knowledge.CountLearnings(ctx, c.store, time.Time{}, domain.LearningCount{})
	require.NoError(t, err)

	// The checks: no target, itself, unknown, an applied one, status mix.
	for name, p := range map[string]rpc.LearnUpdateParams{
		"itself":         {IDs: []int64{kept, d1}, DuplicateOf: kept},
		"applied target": {IDs: []int64{d1}, DuplicateOf: applied},
		"applied id":     {IDs: []int64{applied}, DuplicateOf: kept},
		"with status":    {IDs: []int64{d1}, DuplicateOf: kept, Status: domain.LearningRejected},
		"with pr":        {IDs: []int64{d1}, DuplicateOf: kept, PR: "https://github.com/a/b/pull/1"},
		"status flag":    {IDs: []int64{d1}, Status: domain.LearningDuplicate},
	} {
		_, err := update(p)
		require.ErrorIs(t, err, domain.ErrInvalid, name)
	}
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{d1}, DuplicateOf: 9999})
	require.ErrorIs(t, err, domain.ErrNotFound, "unknown target")
	assert.Empty(t, list(rpc.LearnListParams{Status: domain.LearningDuplicate}), "failed calls change nothing")

	// Two duplicates of a third.
	got, err := update(rpc.LearnUpdateParams{IDs: []int64{d1, d2}, DuplicateOf: kept})
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, l := range got {
		assert.Equal(t, domain.LearningDuplicate, l.Status)
		assert.Equal(t, kept, l.DuplicateOf)
	}
	open := list(rpc.LearnListParams{Status: domain.LearningOpen})
	assert.ElementsMatch(t, []int64{kept, other}, keys(open), "duplicates are not open")
	assert.Equal(t, 2, open[kept].Duplicates)
	assert.Zero(t, open[other].Duplicates)
	dups := list(rpc.LearnListParams{DuplicateOf: kept})
	assert.ElementsMatch(t, []int64{d1, d2}, keys(dups))
	assert.ElementsMatch(t, []int64{d1, d2}, keys(list(rpc.LearnListParams{Status: domain.LearningDuplicate})))

	// No chains: a duplicate cannot be kept, and a kept one cannot become one.
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{other}, DuplicateOf: d1})
	require.ErrorIs(t, err, domain.ErrInvalid, "duplicate of a duplicate")
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{kept}, DuplicateOf: other})
	require.ErrorIs(t, err, domain.ErrInvalid, "a learning others repeat")
	// A duplicate can move to another kept learning.
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{d2}, DuplicateOf: other})
	require.NoError(t, err)
	assert.Equal(t, 1, list(rpc.LearnListParams{Status: domain.LearningOpen})[kept].Duplicates)
	assert.Equal(t, 1, list(rpc.LearnListParams{Status: domain.LearningOpen})[other].Duplicates)
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{d2}, DuplicateOf: kept})
	require.NoError(t, err)

	// A duplicate takes no outcome of its own.
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{d1}, Status: domain.LearningApplied})
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{d1}, PR: "https://github.com/a/b/pull/1"})
	require.ErrorIs(t, err, domain.ErrInvalid)

	// Counts that must see every occurrence: the trigger count (open only)
	// leaves duplicates out, the date-window reads do not.
	n, err := knowledge.CountLearnings(ctx, c.store, time.Time{}, domain.LearningCount{Status: domain.LearningOpen})
	require.NoError(t, err)
	assert.Equal(t, 2, n, "kept and other; duplicates and applied are not open")
	countAfter, err := knowledge.CountLearnings(ctx, c.store, time.Time{}, domain.LearningCount{})
	require.NoError(t, err)
	assert.Equal(t, countBefore, countAfter)
	after, err := knowledge.Measure(ctx, c.store, nil, pools, start.Add(-time.Hour), clk.Now(), false)
	require.NoError(t, err)
	assert.Equal(t, before, after, "EditMetrics sees the same window")

	// Applying the kept learning leaves its duplicates and the count.
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{kept}, Status: domain.LearningApplied, PR: "memory:1"})
	require.NoError(t, err)
	appliedList := list(rpc.LearnListParams{Status: domain.LearningApplied})
	assert.Equal(t, 2, appliedList[kept].Duplicates)
	assert.ElementsMatch(t, []int64{d1, d2}, keys(list(rpc.LearnListParams{DuplicateOf: kept})))
	assert.Equal(t, domain.LearningDuplicate, list(rpc.LearnListParams{DuplicateOf: kept})[d1].Status)
	_, err = update(rpc.LearnUpdateParams{IDs: []int64{other}, DuplicateOf: kept})
	require.ErrorIs(t, err, domain.ErrInvalid, "an applied learning is not kept open")

	// A memory edit cannot cite a duplicate: it would detach from its group.
	_, err = c.memoryWrite(ctx, admin, rpc.MemoryWriteParams{Scope: "global", Text: "x", Learnings: []int64{d1}})
	require.ErrorIs(t, err, domain.ErrInvalid)

	// Reopening a duplicate clears the link and drops the count.
	got, err = update(rpc.LearnUpdateParams{IDs: []int64{d2}, Status: domain.LearningOpen})
	require.NoError(t, err)
	assert.Equal(t, domain.LearningOpen, got[0].Status)
	assert.Zero(t, got[0].DuplicateOf)
	assert.Equal(t, 1, list(rpc.LearnListParams{Status: domain.LearningApplied})[kept].Duplicates)
	assert.ElementsMatch(t, []int64{d1}, keys(list(rpc.LearnListParams{DuplicateOf: kept})))
	assert.Contains(t, keys(list(rpc.LearnListParams{Status: domain.LearningOpen})), d2)

	stored, err := knowledge.GetLearning(ctx, c.store, d2)
	require.NoError(t, err)
	assert.Zero(t, stored.DuplicateOf)
}

func keys(m map[int64]domain.Learning) []int64 {
	out := make([]int64, 0, len(m))
	for id := range m {
		out = append(out, id)
	}
	return out
}
