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
		{Kind: "x", Summary: "s", Source: store.LearningSourceAgent},
		{Kind: store.LearningFailure, Source: store.LearningSourceAgent},
		{Kind: store.LearningFailure, Summary: "s", Source: "human"},
		{Kind: store.LearningFailure, Summary: "s", Source: store.LearningSourceAgent, Status: "done"},
	}
	for _, b := range bad {
		require.ErrorIs(t, store.ValidLearning(&b), store.ErrInvalid)
	}
}
