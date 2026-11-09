package control_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// ticket.list's filters for plugins and custom workflows (#541).
func TestTicketListFilters(t *testing.T) {
	admin := runRoot(t)
	ctx := t.Context()
	r1 := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "1"}})
	r2 := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "2"}})

	list := func(p rpc.ListParams) []string {
		t.Helper()
		var ts []domain.Ticket
		require.NoError(t, admin.Call(ctx, rpc.MethodTicketList, p, &ts))
		out := make([]string, len(ts))
		for i, tk := range ts {
			out[i] = tk.ID
		}
		return out
	}
	// Only string meta values are indexed; github.issue is a number.
	assert.Equal(t, []string{r1.Run.ID, r2.Run.ID}, list(rpc.ListParams{Kind: domain.KindRun, Meta: map[string]string{"github.repo": "acme/widget"}}))
	assert.Empty(t, list(rpc.ListParams{Kind: domain.KindRun, Meta: map[string]string{"github.repo": "other/repo"}}))
	assert.Equal(t, []string{r1.Run.ID, r2.Run.ID}, list(rpc.ListParams{Kind: domain.KindRun, Workflow: "pr", State: domain.RunRunning}))
	assert.Equal(t, []string{r2.Run.ID, r1.Run.ID}, list(rpc.ListParams{Kind: domain.KindRun, Order: "newest"}))
	assert.Equal(t, []string{r2.Run.ID}, list(rpc.ListParams{Kind: domain.KindRun, After: r1.Run.ID}))
	assert.Len(t, list(rpc.ListParams{Run: r1.Run.ID}), len(r1.Steps))

	var ts []domain.Ticket
	require.ErrorIs(t, admin.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Order: "random"}, &ts), domain.ErrInvalid)
	require.ErrorIs(t, admin.Call(ctx, rpc.MethodTicketList, rpc.ListParams{Meta: map[string]string{"nodot": "x"}}, &ts), domain.ErrInvalid)
}
