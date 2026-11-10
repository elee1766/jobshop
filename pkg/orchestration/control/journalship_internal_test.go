package control

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestJournalSearchNotConfigured: without journal.victorialogs, search
// and backfill give the clear error.
func TestJournalSearchNotConfigured(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, journalYAML)
	_, err := c.journalSearch(t.Context(), scope{admin: true}, rpc.JournalSearchParams{Query: "*"})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
	_, err = c.journalBackfill(t.Context(), scope{admin: true}, rpc.JournalBackfillParams{})
	assert.Contains(t, err.Error(), "journal search needs journal.victorialogs")
}
