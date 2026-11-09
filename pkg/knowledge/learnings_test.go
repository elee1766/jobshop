package knowledge_test

import (
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/knowledge"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

func open(t *testing.T, now *time.Time) *sqlite.Store {
	t.Helper()
	s, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "j.db"), sqlite.Options{Now: func() time.Time { return *now }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func add(t *testing.T, s *sqlite.Store, l domain.Learning, project string) domain.Learning {
	t.Helper()
	if l.Source == "" {
		l.Source = domain.LearningSourceAgent
	}
	out, err := knowledge.AddLearning(t.Context(), s, l, project)
	require.NoError(t, err)
	return out
}

// TestAddListUpdate: ids count up past 9 within one millisecond, List
// filters and orders newest first by id, Update and SetGate change one
// learning.
func TestAddListUpdate(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)

	_, err := knowledge.AddLearning(ctx, s, domain.Learning{Kind: "Bad", Summary: "x", Source: domain.LearningSourceAgent}, "")
	require.ErrorIs(t, err, domain.ErrInvalid)

	var last domain.Learning
	for i := range 11 {
		last = add(t, s, domain.Learning{Kind: domain.LearningFailure, Summary: "s" + strconv.Itoa(i), Ticket: "js-1", Scope: "pool:dev"}, "p")
	}
	assert.Equal(t, int64(11), last.ID)
	assert.Equal(t, domain.LearningOpen, last.Status)
	now = now.Add(time.Minute)
	add(t, s, domain.Learning{Kind: domain.LearningWorkaround, Summary: "other"}, "")

	ls, err := knowledge.ListLearnings(ctx, s, store.LearningFilter{Kind: domain.LearningFailure})
	require.NoError(t, err)
	require.Len(t, ls, 11)
	assert.Equal(t, int64(11), ls[0].ID, "same millisecond: highest id first")
	ls, err = knowledge.ListLearnings(ctx, s, store.LearningFilter{Ticket: "js-1", Limit: 2})
	require.NoError(t, err)
	assert.Len(t, ls, 2)
	ls, err = knowledge.ListLearnings(ctx, s, store.LearningFilter{Scope: domain.ScopeGlobal})
	require.NoError(t, err)
	require.Len(t, ls, 1)
	assert.Equal(t, int64(12), ls[0].ID)

	u, err := knowledge.UpdateLearning(ctx, s, 3, domain.LearningApplied, " https://x/pull/1 ")
	require.NoError(t, err)
	assert.Equal(t, domain.LearningApplied, u.Status)
	assert.Equal(t, "https://x/pull/1", u.PR)
	_, err = knowledge.UpdateLearning(ctx, s, 3, "done", "")
	require.ErrorIs(t, err, domain.ErrInvalid)
	_, err = knowledge.UpdateLearning(ctx, s, 99, domain.LearningApplied, "")
	require.ErrorIs(t, err, domain.ErrNotFound)
	require.NoError(t, knowledge.SetLearningGate(ctx, s, 3, "helped: fewer retries"))
	got, err := knowledge.GetLearning(ctx, s, 3)
	require.NoError(t, err)
	assert.Equal(t, "helped: fewer retries", got.Gate)
	assert.Equal(t, domain.LearningApplied, got.Status)

	byPR, err := knowledge.LearningsByPR(ctx, s, "https://x/pull/1")
	require.NoError(t, err)
	require.Len(t, byPR, 1)
	open, err := knowledge.ListLearnings(ctx, s, store.LearningFilter{Status: domain.LearningOpen})
	require.NoError(t, err)
	assert.Len(t, open, 11)

	between, err := knowledge.LearningsBetween(ctx, s, now, now.Add(time.Millisecond))
	require.NoError(t, err)
	require.Len(t, between, 1)
	assert.Equal(t, int64(12), between[0].ID)
}

// TestCount: Count filters by since, status, kind and scope globs and
// source (#524).
func TestCount(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)
	a := func(kind, scope, source string) {
		add(t, s, domain.Learning{Kind: kind, Summary: kind, Scope: scope, Source: source}, "")
	}
	a("review-nit", "repo:acme/widget", domain.LearningSourceAgent)
	a("review-nit", "repo:acme/widget", domain.LearningSourceAgent)
	a("friction", "repo:acme/widget", domain.LearningSourceAgent)
	a("review-big", "repo:other/x", domain.LearningSourceAgent)
	a("failure", "pool:dev", domain.LearningSourceController)
	start := now
	now = now.Add(time.Minute)
	a("review-nit", "pool:dev", domain.LearningSourceAgent)

	count := func(since time.Time, f domain.LearningCount) int {
		t.Helper()
		n, err := knowledge.CountLearnings(ctx, s, since, f)
		require.NoError(t, err)
		return n
	}
	assert.Equal(t, 6, count(time.Time{}, domain.LearningCount{}))
	assert.Equal(t, 6, count(time.Time{}, domain.LearningCount{Status: domain.LearningOpen}))
	assert.Equal(t, 0, count(time.Time{}, domain.LearningCount{Status: domain.LearningApplied}))
	assert.Equal(t, 4, count(time.Time{}, domain.LearningCount{Kinds: []string{"review-*"}}))
	assert.Equal(t, 3, count(time.Time{}, domain.LearningCount{Scopes: []string{"repo:acme/*"}}))
	assert.Equal(t, 2, count(time.Time{}, domain.LearningCount{Kinds: []string{"review-*"}, Scopes: []string{"repo:acme/widget"}}))
	assert.Equal(t, 1, count(time.Time{}, domain.LearningCount{Source: domain.LearningSourceController}))
	assert.Equal(t, 1, count(start.Add(time.Millisecond), domain.LearningCount{}))
}

// TestRecall: summary and fix are searchable; a global learning (no
// project) matches every project; a project learning only its own.
func TestRecall(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)
	g := add(t, s, domain.Learning{Kind: domain.LearningWorkaround, Summary: "nix vendorHash mismatch", Fix: "use lib.fakeHash then paste got"}, "")
	add(t, s, domain.Learning{Kind: domain.LearningFailure, Summary: "fakeHash in other project", Ticket: "q-1"}, "q")

	hits, err := knowledge.RecallLearnings(ctx, s, "fakeHash", "p", "", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, store.RecallLearning, hits[0].Kind)
	assert.Equal(t, strconv.FormatInt(g.ID, 10), hits[0].Ref)
	assert.Contains(t, hits[0].Snippet, "fakeHash")

	hits, err = knowledge.RecallLearnings(ctx, s, "fakeHash", "", "", 0)
	require.NoError(t, err)
	assert.Len(t, hits, 2)
	hits, err = knowledge.RecallLearnings(ctx, s, "fakeHash", "", "q-1", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	assert.Equal(t, "q", hits[0].Project)

	// An update of the gate keeps the record searchable.
	require.NoError(t, knowledge.SetLearningGate(ctx, s, g.ID, "helped: x"))
	hits, err = knowledge.RecallLearnings(ctx, s, "vendorHash", "", "", 0)
	require.NoError(t, err)
	assert.Len(t, hits, 1)
}
