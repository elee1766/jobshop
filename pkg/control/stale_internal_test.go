package control

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// TestStaleAttention: an open ticket no pool will claim is listed in
// attention once it is older than attention.orphan_age, with a close
// hint, and only once, also when it is labeled needs-human (#495).
func TestStaleAttention(t *testing.T) {
	t.Parallel()
	c, clk, _ := subController(t)
	ctx := t.Context()
	none := mkTicket(t, c, store.NewTicket{Title: "Self-deploy stalled"})
	gone := mkTicket(t, c, store.NewTicket{Title: "old pool", Pool: "removed"})
	mkTicket(t, c, store.NewTicket{Title: "claimable", Pool: "dev"})
	human := mkTicket(t, c, store.NewTicket{Title: "ask", Labels: []string{needsHumanLabel}})

	stale := func() []rpc.AttentionItem {
		items, err := c.attentionItems(ctx, c.config())
		require.NoError(t, err)
		var out []rpc.AttentionItem
		for _, it := range items {
			if it.Kind == rpc.AttentionStale {
				out = append(out, it)
			}
		}
		return out
	}
	assert.Empty(t, stale(), "younger than the 72h default")

	clk.advance(71 * time.Hour)
	assert.Empty(t, stale(), "still younger than orphan_age")

	clk.advance(2 * time.Hour)
	got := stale()
	require.Len(t, got, 2)
	assert.Equal(t, []string{none.ID, gone.ID}, []string{got[0].ID, got[1].ID})
	assert.Contains(t, got[0].Title, "no claimable pool (no_pool) for 73h0m0s")
	assert.Contains(t, got[0].Title, "jobshop close "+none.ID+" -r cancelled")
	assert.Contains(t, got[1].Title, "(unknown_pool)")
	assert.Len(t, stale(), 2, "a second call lists the same items, not more")

	items, err := c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	n := 0
	for _, it := range items {
		if it.ID == human.ID {
			n++
			assert.Equal(t, rpc.AttentionTicket, it.Kind)
		}
	}
	assert.Equal(t, 1, n, "a needs-human ticket with no pool is listed once")

	c.config().Attention.OrphanAge = "200h"
	assert.Empty(t, stale(), "orphan_age is read from config")
	c.config().Attention.OrphanAge = "off"
	assert.Empty(t, stale())
}
