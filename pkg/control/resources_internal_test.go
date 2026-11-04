package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestResourceWaitStallAndAttention: a ticket held back by a full
// resource is a resource_wait stall and an attention item once it has
// waited long enough (#408).
func TestResourceWaitStallAndAttention(t *testing.T) {
	c, clk, _ := subController(t)
	ctx := t.Context()
	require.NoError(t, c.store.SetResourceCaps(ctx, map[string]int{"emu": 1}))
	res := json.RawMessage(`{"resources":["emu"]}`)
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev", Fields: res})
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev", Fields: res})
	_, _, err := c.store.Claim(ctx, a.ID, "s1", time.Hour)
	require.NoError(t, err)

	found, err := c.stallResourceWait(ctx, clk.now(), 30*time.Minute)
	require.NoError(t, err)
	assert.Empty(t, found, "not waited long enough")
	c.config().Attention.ResourceWait = "30m"
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	clk.advance(time.Hour)
	found, err = c.stallResourceWait(ctx, clk.now(), 30*time.Minute)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, stall{config.StallResourceWait, b.ID, "waiting on emu, 0 ahead (1/1 held) for 1h0m0s"}, found[0])
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionResource, items[0].Kind)
	assert.Equal(t, b.ID, items[0].ID)
	assert.Contains(t, items[0].Title, "waiting on emu")
}
