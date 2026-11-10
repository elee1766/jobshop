package control

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/triggers"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestResourceWaitSignalAndAttention: a ticket held back by a full
// resource is a resource_wait finding and an attention item once it has
// waited long enough (#408).
func TestResourceWaitSignalAndAttention(t *testing.T) {
	t.Parallel()
	c, clk, _ := subController(t)
	ctx := t.Context()
	require.NoError(t, c.store.SetResources(ctx, domain.ResourceConfig{Caps: map[string]int{"emu": 1}}))
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step := func(title string) domain.Ticket {
		return mkTicket(t, c, store.NewTicket{
			Kind: domain.KindStep, Title: title, Pool: "dev", ParentID: run.ID,
			Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"` + title + `","retries":3,"resources":["emu"]}`),
		})
	}
	a := step("a")
	b := step("b")
	_, _, err := c.store.Claim(ctx, a.ID, "s1", time.Hour)
	require.NoError(t, err)

	found, err := c.signalResourceWait(ctx, clk.now(), 30*time.Minute)
	require.NoError(t, err)
	assert.Empty(t, found, "not waited long enough")
	c.config().Attention.ResourceWait = "30m"
	items, err := c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	assert.Empty(t, items)

	clk.advance(time.Hour)
	found, err = c.signalResourceWait(ctx, clk.now(), 30*time.Minute)
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, triggers.Finding{Signal: config.SignalResourceWait, Subject: b.ID, Detail: "waiting on emu, 0 ahead (1/1 held) for 1h0m0s"}, found[0])
	items, err = c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionResource, items[0].Kind)
	assert.Equal(t, b.ID, items[0].ID)
	assert.Contains(t, items[0].Title, "waiting on emu")
}
