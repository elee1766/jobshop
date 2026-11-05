package store_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

func TestParseScope(t *testing.T) {
	for in, want := range map[string]string{
		"":                      "global",
		"global":                "global",
		"repo:elee1766/jobshop": "repo:elee1766/jobshop",
		"pool:dev":              "pool:dev",
	} {
		got, err := store.ParseScope(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"repo:", "repo:jobshop", "repo:a/b/c", "pool:", "pool:a b", "team:eng", "Global"} {
		_, err := store.ParseScope(in)
		require.ErrorIs(t, err, store.ErrInvalid, in)
	}
}

func TestValidLearning(t *testing.T) {
	l := store.Learning{Kind: store.LearningWorkaround, Summary: " x ", Source: store.LearningSourceAgent}
	require.NoError(t, store.ValidLearning(&l))
	assert.Equal(t, store.Learning{Kind: store.LearningWorkaround, Summary: "x", Source: store.LearningSourceAgent, Scope: "global", Status: "open"}, l)
	bad := []store.Learning{
		{Kind: "Bad kind", Summary: "s", Source: store.LearningSourceAgent},
		{Kind: store.LearningFailure, Source: store.LearningSourceAgent},
		{Kind: store.LearningFailure, Summary: "s", Source: "human"},
		{Kind: store.LearningFailure, Summary: "s", Source: store.LearningSourceAgent, Status: "done"},
	}
	for _, b := range bad {
		require.ErrorIs(t, store.ValidLearning(&b), store.ErrInvalid)
	}
}

func TestLearningKinds(t *testing.T) {
	// Default config: today's three kinds.
	for _, k := range []string{"correction", "failure", "workaround"} {
		require.NoError(t, store.CheckLearningKind(k, nil), k)
	}
	err := store.CheckLearningKind("friction", nil)
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Contains(t, err.Error(), "correction, failure, workaround")

	pats := []string{"friction", "review-*"}
	require.NoError(t, store.CheckLearningKind("friction", pats))
	require.NoError(t, store.CheckLearningKind("review-nit", pats))
	require.NoError(t, store.CheckLearningKind("failure", pats), "controller kinds always allowed")
	err = store.CheckLearningKind("blocker", pats)
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Contains(t, err.Error(), "friction, review-*")

	for _, p := range []string{"friction", "review-*", "*", "b?ocker", "[a-c]*"} {
		require.NoError(t, store.ValidKindPattern(p), p)
	}
	for _, p := range []string{"", "Friction", "1x", "-x", "a b", "a/b", "[", "x_y"} {
		require.Error(t, store.ValidKindPattern(p), p)
	}

	l := store.Learning{Kind: "Bad Kind", Summary: "s", Source: store.LearningSourceAgent}
	require.ErrorIs(t, store.ValidLearning(&l), store.ErrInvalid)
}

func TestLearningCountMatches(t *testing.T) {
	f := store.LearningCount{Kinds: []string{"review-*"}, Scopes: []string{"repo:elee1766/*", "pool:acphost-*"}, Source: store.LearningSourceAgent}
	assert.True(t, f.Matches("review-nit", "repo:elee1766/jobshop", "agent"))
	assert.True(t, f.Matches("review-x", "pool:acphost-dev", "agent"))
	assert.False(t, f.Matches("failure", "repo:elee1766/jobshop", "agent"))
	assert.False(t, f.Matches("review-nit", "repo:other/x", "agent"))
	assert.False(t, f.Matches("review-nit", "global", "agent"))
	assert.False(t, f.Matches("review-nit", "repo:elee1766/jobshop", "controller"))
	assert.True(t, store.LearningCount{}.Matches("anything", "global", "controller"))
	assert.True(t, store.Glob("repo:*", "repo:a/b"), "* crosses /")
}
