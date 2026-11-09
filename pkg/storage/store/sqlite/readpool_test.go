package sqlite_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// TestReadsDoNotWaitOnWriter checks Get and List run while a write
// transaction holds the writer connection, and see only committed data.
func TestReadsDoNotWaitOnWriter(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, err := sqlite.Open(ctx, testPath(t), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	a, err := store.Create(ctx, s, "test", store.NewTicket{Project: "p", Kind: domain.KindTask, Title: "a"})
	require.NoError(t, err)

	err = s.WriteTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "UPDATE tickets SET title = 'uncommitted' WHERE id = ?", a.ID); err != nil {
			return err
		}
		// With one shared connection these would block until the
		// transaction ends, which is never.
		got, err := s.Get(ctx, a.ID)
		require.NoError(t, err)
		require.Equal(t, "a", got.Title)
		ts, err := s.List(ctx, store.Filter{Project: "p"})
		require.NoError(t, err)
		require.Len(t, ts, 1)
		return nil
	})
	require.NoError(t, err)

	got, err := s.Get(ctx, a.ID)
	require.NoError(t, err)
	require.Equal(t, "uncommitted", got.Title)
}
