package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	host "github.com/elee1766/acphost"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestPreflightQuarantine (#728): a sandboxed pool whose host fails the
// isolation preflight is quarantined at start, before any session
// spawns, with the report as the reason. A reload once the host passes
// clears it.
func TestPreflightQuarantine(t *testing.T) {
	var fail atomic.Bool
	fail.Store(true)
	var seen atomic.Pointer[host.Spec]
	check := func(s host.Spec) (host.PreflightReport, error) {
		seen.Store(&s)
		if fail.Load() {
			return host.PreflightReport{Checks: []host.PreflightCheck{
				{Name: "network.pasta", Reason: "pasta not found", Hint: "install passt"},
			}}, nil
		}
		return host.PreflightReport{Checks: []host.PreflightCheck{{Name: "network.pasta", OK: true}}}, nil
	}
	root, start := poolRootOpt(t, "1m", control.Options{Preflight: check})
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n    sandbox: {backend: bubblewrap, network: pasta}\n  plain:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n", os.Args[0], root.Dir, root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	status := func() rpc.PoolStatusResult {
		var st rpc.PoolStatusResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
		return st
	}

	var task store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketCreate, rpc.CreateParams{Title: "t", Pool: "dev"}, &task))
	pass(t, root.Dir, 2)
	st := status()
	assert.Contains(t, st.Quarantined["dev"], "FAIL network.pasta: pasta not found")
	assert.Contains(t, st.Quarantined["dev"], "fix: install passt")
	assert.NotContains(t, st.Quarantined, "plain", "an unisolated pool is not checked")
	require.NotNil(t, seen.Load())
	require.NotNil(t, seen.Load().Sandbox)
	assert.Equal(t, "pasta", seen.Load().Sandbox.Network)

	var sessions []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Kind: store.KindSession}, &sessions))
	assert.Empty(t, sessions, "no spawn while quarantined")
	var esc []store.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Label: "needs-human"}, &esc))
	require.Len(t, esc, 1)
	assert.Contains(t, esc[0].Body, "fix: install passt")

	fail.Store(false)
	require.NoError(t, admin.Call(t.Context(), rpc.MethodReload, struct{}{}, nil))
	assert.Empty(t, status().Quarantined, "a passing preflight after reload clears it")
}
