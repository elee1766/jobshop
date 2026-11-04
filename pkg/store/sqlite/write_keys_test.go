package sqlite_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

func TestWriteKeys(t *testing.T) {
	ctx := t.Context()
	s, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	defer func() { _ = s.Close() }()

	k, err := s.ClaimWriteKey(ctx, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.True(t, k.New)
	assert.False(t, k.Done)

	k, err = s.ClaimWriteKey(ctx, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.False(t, k.New, "second claim of a pending key: an earlier attempt")
	assert.False(t, k.Done)

	_, err = s.ClaimWriteKey(ctx, "js/k1", "linear.create", "sess")
	require.ErrorIs(t, err, store.ErrInvalid, "a key is bound to its op")

	k, err = s.FinishWriteKey(ctx, "js/k1", `{"id":"a"}`)
	require.NoError(t, err)
	assert.True(t, k.Done)
	k, err = s.FinishWriteKey(ctx, "js/k1", `{"id":"b"}`)
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":"a"}`, k.Result, "first result wins")

	k, err = s.ClaimWriteKey(ctx, "js/k1", "linear.ask", "sess")
	require.NoError(t, err)
	assert.True(t, k.Done)
	assert.JSONEq(t, `{"id":"a"}`, k.Result)

	_, err = s.FinishWriteKey(ctx, "js/none", "x")
	require.ErrorIs(t, err, store.ErrNotFound)
}
