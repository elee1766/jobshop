package storetest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
)

// testCountLearnings: CountLearnings filters by since, status, kind and
// scope globs and source on every backend (#524).
func testCountLearnings(t *testing.T, s store.Store, c *Clock) {
	ctx := t.Context()
	add := func(kind, scope, source string) {
		t.Helper()
		_, err := s.AddLearning(ctx, store.Learning{Kind: kind, Summary: kind, Scope: scope, Source: source})
		require.NoError(t, err)
	}
	add("review-nit", "repo:acme/widget", store.LearningSourceAgent)
	add("review-nit", "repo:acme/widget", store.LearningSourceAgent)
	add("friction", "repo:acme/widget", store.LearningSourceAgent)
	add("review-big", "repo:other/x", store.LearningSourceAgent)
	add("failure", "pool:dev", store.LearningSourceController)
	start := c.T
	c.Advance(time.Minute)
	add("review-nit", "pool:dev", store.LearningSourceAgent)

	count := func(since time.Time, f store.LearningCount) int {
		t.Helper()
		n, err := s.CountLearnings(ctx, since, f)
		require.NoError(t, err)
		return n
	}
	assert.Equal(t, 6, count(time.Time{}, store.LearningCount{}))
	assert.Equal(t, 6, count(time.Time{}, store.LearningCount{Status: store.LearningOpen}))
	assert.Equal(t, 0, count(time.Time{}, store.LearningCount{Status: store.LearningApplied}))
	assert.Equal(t, 4, count(time.Time{}, store.LearningCount{Kinds: []string{"review-*"}}))
	assert.Equal(t, 3, count(time.Time{}, store.LearningCount{Scopes: []string{"repo:acme/*"}}))
	assert.Equal(t, 2, count(time.Time{}, store.LearningCount{Kinds: []string{"review-*"}, Scopes: []string{"repo:acme/widget"}}))
	assert.Equal(t, 1, count(time.Time{}, store.LearningCount{Source: store.LearningSourceController}))
	assert.Equal(t, 1, count(start.Add(time.Millisecond), store.LearningCount{}))
}
