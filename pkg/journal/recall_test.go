package journal_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

// TestRecall: journal summaries rank a short focused match first, a
// re-index replaces the session's record, an empty text removes it, and
// project and ticket filter.
func TestRecall(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	s, err := sqlite.Open(ctx, filepath.Join(t.TempDir(), "j.db"), sqlite.Options{})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	at := time.UnixMilli(1_700_000_000_000)
	require.NoError(t, journal.Index(ctx, s, journal.Doc{
		Session: "s1", Ticket: "t1", Project: "p", At: at,
		Text: "vendor hash mismatch once, then lots of other text about reviewing and merging and waiting",
	}))
	require.NoError(t, journal.Index(ctx, s, journal.Doc{
		Session: "s2", Ticket: "t2", Project: "p", At: at,
		Text: "nix vendor hash mismatch: set vendorHash to the got value",
	}))
	hits, err := journal.Recall(ctx, s, "vendor hash mismatch", "", "", 0)
	require.NoError(t, err)
	require.Len(t, hits, 2)
	require.Equal(t, "s2", hits[0].Ref)
	require.Equal(t, store.RecallJournal, hits[0].Kind)
	require.Equal(t, at.UTC(), hits[0].At)
	require.Contains(t, hits[0].Snippet, "vendor")

	hits, err = journal.Recall(ctx, s, "vendor", "other", "", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
	hits, err = journal.Recall(ctx, s, "vendor", "", "t1", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, "s1", hits[0].Ref)

	require.NoError(t, journal.Index(ctx, s, journal.Doc{Session: "s1", Project: "p", At: at, Text: "now about nftables"}))
	hits, err = journal.Recall(ctx, s, "nftables", "", "", 0)
	require.NoError(t, err)
	require.Len(t, hits, 1)
	require.Equal(t, "s1", hits[0].Ref)

	require.NoError(t, journal.Index(ctx, s, journal.Doc{Session: "s2"}))
	hits, err = journal.Recall(ctx, s, "vendor", "", "", 0)
	require.NoError(t, err)
	require.Empty(t, hits)
}
