package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestPoolShowSandbox (JOB-83): pool.show returns the merged sandbox and
// isolation of a pool with the level that set each entry.
func TestPoolShowSandbox(t *testing.T) {
	root, start := poolRoot(t, "1m")
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
sandbox_profiles:
  host: {backend: bubblewrap, ro: [/nix/store]}
  egress-deny: {network: pasta, deny: [10.0.0.0/8]}
isolation_profiles:
  mapped: {user: {uid: 1000, map_uid: 100000, map_gid: 100000, groups: [jobshop]}}
defaults:
  isolation_by_label:
    - {labels: ["tier:iso"], profiles: [mapped]}
pools:
  iso:
    project: js
    agent: fake
    max_sessions: 1
    dir: %q
    labels: ["tier:iso"]
    sandbox: {profiles: [host, egress-deny], ro: [/extra]}
`, os.Args[0], root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	var res rpc.PoolShowResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolShow, rpc.PoolShowParams{Pool: "iso"}, &res))
	assert.Equal(t, "iso", res.Pool)
	assert.Equal(t, []string{"tier:iso"}, res.Labels)
	has := func(es []config.MergedEntry, field, value, source string) bool {
		for _, e := range es {
			if e.Field == field && e.Value == value && e.Source == source {
				return true
			}
		}
		return false
	}
	assert.True(t, has(res.Sandbox, "backend", "bubblewrap", "profile host (pool)"), "%v", res.Sandbox)
	assert.True(t, has(res.Sandbox, "ro", "/nix/store", "profile host (pool)"))
	assert.True(t, has(res.Sandbox, "ro", "/extra", "pool"))
	assert.True(t, has(res.Sandbox, "network", "pasta", "profile egress-deny (pool)"))
	assert.True(t, has(res.Sandbox, "deny", "10.0.0.0/8", "profile egress-deny (pool)"))
	require.Len(t, res.Isolation, 1)
	assert.Equal(t, "user", res.Isolation[0].Field)
	assert.Equal(t, `profile mapped (label rule 1 [tier:iso])`, res.Isolation[0].Source)

	err := admin.Call(t.Context(), rpc.MethodPoolShow, rpc.PoolShowParams{Pool: "nope"}, &res)
	require.Error(t, err)
	assert.Contains(t, err.Error(), domain.ErrNotFound.Error())
}
