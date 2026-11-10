package control

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

const suspendedYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %[1]q, max_sessions: 1}
`

// TestSuspendedSessionsTakeNoSlot (JOB-189): a suspended session of any
// kind, not only a parked one, counts against neither the pool's human
// slot (chat.go) nor the pre-start cap (attentionchat.go).
func TestSuspendedSessionsTakeNoSlot(t *testing.T) {
	c, _, _ := scanControllerYAML(t, suspendedYAML)
	ctx := t.Context()
	mk := func(f domain.SessionFields) domain.Ticket {
		f.Agent = "fake"
		b, err := json.Marshal(f)
		require.NoError(t, err)
		tk, err := store.Create(ctx, c.store, "test", store.NewTicket{Project: "js", Kind: domain.KindSession, Title: "s", Pool: "dev", Fields: b})
		require.NoError(t, err)
		return tk
	}
	mk(domain.SessionFields{})
	mk(domain.SessionFields{})
	require.Error(t, c.humanSlot(ctx, c.config(), "dev"), "two live sessions fill max_sessions + 1")

	// A suspended attention chat frees its slot; so does a parked session.
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: "dev"})
	require.NoError(t, err)
	require.Len(t, open, 2)
	require.NoError(t, c.updateSessionFields(ctx, open[0].ID, func(f *domain.SessionFields) { f.Suspended = true }))
	require.NoError(t, c.humanSlot(ctx, c.config(), "dev"), "a suspended chat takes no slot")
	require.NoError(t, c.updateSessionFields(ctx, open[0].ID, func(f *domain.SessionFields) { f.Parked = "t-1" }))
	require.NoError(t, c.humanSlot(ctx, c.config(), "dev"), "nor does a parked session")

	open, err = c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen, Pool: "dev"})
	require.NoError(t, err)
	assert.Equal(t, 1, countedSessions(open))
}
