package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestJournalSearchNotConfigured: without journal.victorialogs, search
// and backfill give the clear error.
func TestJournalSearchNotConfigured(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, journalYAML)
	_, err := c.journalSearch(t.Context(), scope{admin: true}, rpc.JournalSearchParams{Query: "*"})
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
	_, err = c.journalBackfill(t.Context(), scope{admin: true}, rpc.JournalBackfillParams{})
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
}
