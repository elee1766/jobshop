package sqlite_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
	"github.com/gfx-labs/jobshop/pkg/storage/store/storetest"
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

// TestCollectionReadsDoNotWaitOnWriter checks Get, List, Count and Search
// of a collection return while a write transaction holds the writer
// connection, and see only committed records (JOB-191). It is
// file-backed: an in-memory store has one pool. The transaction ends only
// after the reads, so a read on the writer would block forever.
func TestCollectionReadsDoNotWaitOnWriter(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, err := sqlite.Open(ctx, testPath(t), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	col, err := s.Collection(ctx, storetest.TestSpec)
	require.NoError(t, err)
	_, _, err = col.Put(ctx, store.Record{Key: "a", State: "pending", Value: json.RawMessage(`{"summary":"slow tick"}`)})
	require.NoError(t, err)

	done := make(chan struct{})
	reads := make(chan error, 1)
	holding := make(chan struct{})
	go func() {
		// The reads run in their own goroutine; the transaction below
		// waits for them, so the order is not left to timing.
		<-holding
		reads <- func() error {
			if _, err := col.Get(ctx, "a"); err != nil {
				return err
			}
			got, err := col.List(ctx, store.RecordFilter{})
			if err != nil {
				return err
			}
			if len(got) != 1 {
				return fmt.Errorf("list: %d records", len(got))
			}
			counts, err := col.Count(ctx, store.RecordFilter{})
			if err != nil {
				return err
			}
			if counts["pending"] != 1 {
				return fmt.Errorf("count: %v", counts)
			}
			found, err := col.Search(ctx, "slow", 5)
			if err != nil {
				return err
			}
			if len(found) != 1 {
				return fmt.Errorf("search: %d records", len(found))
			}
			return nil
		}()
		close(done)
	}()
	err = s.WriteTx(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, "UPDATE records SET state = 'done' WHERE collection = ?", storetest.TestSpec.Name); err != nil {
			return err
		}
		close(holding)
		<-done
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, <-reads)

	got, err := col.Count(ctx, store.RecordFilter{})
	require.NoError(t, err)
	require.Equal(t, map[string]int{"done": 1}, got, "committed after the transaction")
}
