package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
)

func TestParseScope(t *testing.T) {
	for in, want := range map[string]string{
		"":                      "global",
		"global":                "global",
		"repo:elee1766/jobshop": "repo:elee1766/jobshop",
		"pool:dev":              "pool:dev",
	} {
		got, err := domain.ParseScope(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	for _, in := range []string{"repo:", "repo:jobshop", "repo:a/b/c", "pool:", "pool:a b", "team:eng", "Global"} {
		_, err := domain.ParseScope(in)
		require.ErrorIs(t, err, domain.ErrInvalid, in)
	}
}

func TestValidLearning(t *testing.T) {
	l := domain.Learning{Kind: domain.LearningWorkaround, Summary: " x ", Source: domain.LearningSourceAgent}
	require.NoError(t, domain.ValidLearning(&l))
	assert.Equal(t, domain.Learning{Kind: domain.LearningWorkaround, Summary: "x", Source: domain.LearningSourceAgent, Scope: "global", Status: "open"}, l)
	bad := []domain.Learning{
		{Kind: "Bad kind", Summary: "s", Source: domain.LearningSourceAgent},
		{Kind: domain.LearningFailure, Source: domain.LearningSourceAgent},
		{Kind: domain.LearningFailure, Summary: "s", Source: "human"},
		{Kind: domain.LearningFailure, Summary: "s", Source: domain.LearningSourceAgent, Status: "done"},
	}
	for _, b := range bad {
		require.ErrorIs(t, domain.ValidLearning(&b), domain.ErrInvalid)
	}
}

func TestLearningKinds(t *testing.T) {
	// Default config: today's three kinds.
	for _, k := range []string{"correction", "failure", "workaround"} {
		require.NoError(t, domain.CheckLearningKind(k, nil), k)
	}
	err := domain.CheckLearningKind("friction", nil)
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "correction, failure, workaround")

	pats := []string{"friction", "review-*"}
	require.NoError(t, domain.CheckLearningKind("friction", pats))
	require.NoError(t, domain.CheckLearningKind("review-nit", pats))
	require.NoError(t, domain.CheckLearningKind("failure", pats), "controller kinds always allowed")
	err = domain.CheckLearningKind("blocker", pats)
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "friction, review-*")

	for _, p := range []string{"friction", "review-*", "*", "b?ocker", "[a-c]*"} {
		require.NoError(t, domain.ValidKindPattern(p), p)
	}
	for _, p := range []string{"", "Friction", "1x", "-x", "a b", "a/b", "[", "x_y"} {
		require.Error(t, domain.ValidKindPattern(p), p)
	}

	l := domain.Learning{Kind: "Bad Kind", Summary: "s", Source: domain.LearningSourceAgent}
	require.ErrorIs(t, domain.ValidLearning(&l), domain.ErrInvalid)
}

func TestLearningCountMatches(t *testing.T) {
	f := domain.LearningCount{Kinds: []string{"review-*"}, Scopes: []string{"repo:elee1766/*", "pool:acphost-*"}, Source: domain.LearningSourceAgent}
	assert.True(t, f.Matches("review-nit", "repo:elee1766/jobshop", "agent"))
	assert.True(t, f.Matches("review-x", "pool:acphost-dev", "agent"))
	assert.False(t, f.Matches("failure", "repo:elee1766/jobshop", "agent"))
	assert.False(t, f.Matches("review-nit", "repo:other/x", "agent"))
	assert.False(t, f.Matches("review-nit", "global", "agent"))
	assert.False(t, f.Matches("review-nit", "repo:elee1766/jobshop", "controller"))
	assert.True(t, domain.LearningCount{}.Matches("anything", "global", "controller"))
	assert.True(t, domain.Glob("repo:*", "repo:a/b"), "* crosses /")
}
