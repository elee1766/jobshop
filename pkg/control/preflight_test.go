package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	host "github.com/elee1766/acphost/pkg/acphost"
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
	var opts atomic.Int32
	check := func(s host.Spec, o ...host.PreflightOption) (host.PreflightReport, error) {
		seen.Store(&s)
		opts.Store(int32(len(o)))
		if fail.Load() {
			return host.PreflightReport{Checks: []host.PreflightCheck{
				{Name: "network.pasta", Reason: "pasta not found", Hint: "install passt"},
			}}, nil
		}
		return host.PreflightReport{Checks: []host.PreflightCheck{{Name: "network.pasta", OK: true}}}, nil
	}
	root, start := poolRootOpt(t, "1m", control.Options{Preflight: check})
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n  dev:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n    sandbox: {backend: bubblewrap, network: pasta, deny: [10.0.0.0/8], dns_check: example.com}\n  plain:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n", os.Args[0], root.Dir, root.Dir)
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
	assert.Equal(t, []string{"10.0.0.0/8"}, seen.Load().Sandbox.Deny, "deny reaches Preflight for its nft check")
	assert.Equal(t, int32(1), opts.Load(), "dns_check adds WithDNSCheck")

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

// TestPreflightSharedParallel (JOB-102): pools that ask for the same
// isolation share one check, and distinct checks run at once, so a
// reload of many sandboxed pools is not one dns_check per pool in
// series. Every pool still gets its own quarantine.
func TestPreflightSharedParallel(t *testing.T) {
	var calls, arrived atomic.Int32
	both := make(chan struct{})
	var overlapped atomic.Bool
	check := func(host.Spec, ...host.PreflightOption) (host.PreflightReport, error) {
		calls.Add(1)
		// The two distinct checks meet here: run in series, the first
		// would wait out the timeout alone.
		if arrived.Add(1) == 2 {
			close(both)
		}
		select {
		case <-both:
			overlapped.Store(true)
		case <-time.After(10 * time.Second):
		}
		return host.PreflightReport{Checks: []host.PreflightCheck{{Name: "network.pasta", Reason: "pasta not found"}}}, nil
	}
	root, start := poolRootOpt(t, "1m", control.Options{Preflight: check})
	pool := func(name, deny string) string {
		return fmt.Sprintf("  %s:\n    project: js\n    agent: fake\n    max_sessions: 1\n    idle_timeout: 1m\n    dir: %q\n    sandbox: {backend: bubblewrap, network: pasta, deny: [%s], dns_check: example.com}\n", name, root.Dir, deny)
	}
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n", os.Args[0]) +
		pool("a", "10.0.0.0/8") + pool("b", "10.0.0.0/8") + pool("c", "192.168.0.0/16")
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	pass(t, root.Dir, 1)

	assert.Equal(t, int32(2), calls.Load(), "a and b share one check, c has its own")
	assert.True(t, overlapped.Load(), "distinct checks run in parallel")
	admin := dial(t, root.AdminSocket())
	var st rpc.PoolStatusResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
	for _, p := range []string{"a", "b", "c"} {
		assert.Contains(t, st.Quarantined[p], "FAIL network.pasta: pasta not found", p)
	}
}
