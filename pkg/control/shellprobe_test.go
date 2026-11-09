package control_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestShellProbeQuarantines: a host env whose PATH lacks a probed command
// fails the spawn naming it, quarantines the pool with an attention item,
// and leaves the ticket ready for the next session (JOB-67).
func TestShellProbeQuarantines(t *testing.T) {
	root, start := poolRootOpt(t, "1m", control.Options{ShellProbe: []string{"sh", "no-such-tool-67"}})
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	var task domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	var st rpc.PoolStatusResult
	passEventually(t, root.Dir, func() bool {
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
		return st.Quarantined["dev"] != ""
	}, "pool quarantined")
	assert.Contains(t, st.Quarantined["dev"], "agent shell probe: missing no-such-tool-67")
	assert.NotContains(t, st.Quarantined["dev"], "missing sh")

	var got struct {
		Ticket domain.Ticket `json:"ticket"`
	}
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketShow, rpc.IDParams{ID: task.ID}, &got))
	assert.Equal(t, domain.StatusOpen, got.Ticket.Status, "no attempt burned")
	assert.Empty(t, got.Ticket.Holder)

	var res rpc.AttentionResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodAttention, nil, &res))
	found := false
	for _, it := range res.Items {
		if strings.Contains(it.Title, "no-such-tool-67") {
			found = true
		}
	}
	assert.True(t, found, "attention shows the probe failure: %+v", res.Items)
}
