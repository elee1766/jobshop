package writekey_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
	"github.com/elee1766/jobshop/pkg/writekey"
)

func open(t *testing.T, now *time.Time) *sqlite.Store {
	t.Helper()
	s, err := sqlite.Open(t.Context(), filepath.Join(t.TempDir(), "j.db"), sqlite.Options{Now: func() time.Time { return *now }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestClaimFinish(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)

	k, err := writekey.Claim(ctx, s, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.True(t, k.New)
	assert.False(t, k.Done)
	assert.Equal(t, "sess", k.Actor)

	k, err = writekey.Claim(ctx, s, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.False(t, k.New, "second claim of a pending key: an earlier attempt")
	assert.False(t, k.Done)

	_, err = writekey.Claim(ctx, s, "js/k1", "linear.create", "sess")
	require.ErrorIs(t, err, domain.ErrInvalid, "a key is bound to its op")
	_, err = writekey.Claim(ctx, s, " ", "op", "sess")
	require.ErrorIs(t, err, domain.ErrInvalid)

	k, err = writekey.Finish(ctx, s, "js/k1", `{"id":"a"}`)
	require.NoError(t, err)
	assert.True(t, k.Done)
	k, err = writekey.Finish(ctx, s, "js/k1", `{"id":"b"}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"a"}`, k.Result, "first result wins")

	k, err = writekey.Claim(ctx, s, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.True(t, k.Done)
	assert.JSONEq(t, `{"id":"a"}`, k.Result)

	_, err = writekey.Finish(ctx, s, "js/none", "x")
	require.ErrorIs(t, err, domain.ErrNotFound)
}

// TestPrune: keys older than the cutoff go, done or pending; newer keys
// stay (#415).
func TestPrune(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	now := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	s := open(t, &now)
	_, err := writekey.Claim(ctx, s, "js/old-done", "op", "a")
	require.NoError(t, err)
	_, err = writekey.Finish(ctx, s, "js/old-done", "r")
	require.NoError(t, err)
	_, err = writekey.Claim(ctx, s, "js/old-pending", "op", "a")
	require.NoError(t, err)
	now = now.Add(47 * time.Hour)
	_, err = writekey.Claim(ctx, s, "js/new", "op", "a")
	require.NoError(t, err)
	now = now.Add(2 * time.Hour)

	n, err := writekey.Prune(ctx, s, now.Add(-48*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, int64(2), n)
	k, err := writekey.Claim(ctx, s, "js/old-done", "op", "a")
	require.NoError(t, err)
	assert.True(t, k.New, "pruned key is claimed fresh")
	k, err = writekey.Claim(ctx, s, "js/new", "op", "a")
	require.NoError(t, err)
	assert.False(t, k.New, "newer key kept")
}
