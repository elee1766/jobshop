package sqlite_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// TestRecallDuringCollectionDDL: Recall runs on the read pool while other
// collections register their FTS5 tables on the writer. A schema change
// under a read must not surface as "database schema has changed" (#1069).
func TestRecallDuringCollectionDDL(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, err := sqlite.Open(ctx, testPath(t), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	tk, err := store.Create(ctx, s, "admin", store.NewTicket{Project: "p", Kind: domain.KindTask, Title: "t"})
	require.NoError(t, err)
	_, err = store.AddComment(ctx, s, "a", tk.ID, "vendorHash mismatch fixed")
	require.NoError(t, err)

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	wg.Go(func() {
		for i := range 40 {
			spec := store.CollectionSpec{Name: fmt.Sprintf("ddl_%d", i), Search: []store.Field{{Name: "body", Path: "body"}}}
			if _, err := s.Collection(ctx, spec); err != nil {
				errs <- err
				return
			}
		}
	})
	for range 3 {
		wg.Go(func() {
			for range 300 {
				hits, err := s.Recall(ctx, store.RecallQuery{Text: "vendorHash mismatch"})
				if err != nil {
					errs <- err
					return
				}
				if len(hits) != 1 {
					errs <- fmt.Errorf("want 1 hit, got %d", len(hits))
					return
				}
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
}
