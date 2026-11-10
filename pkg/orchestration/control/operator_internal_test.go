package control

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// TestOperatorGrant (#248): an operator-pool session can call each admin
// method the host-level operator needs across instances; a dev-pool
// session cannot call the granted ones.
func TestOperatorGrant(t *testing.T) {
	t.Parallel()
	// Fixture packs (JOB-2): a repo pack included with a prefix, and an
	// operator chat pool holding the grants under test.
	packs := t.TempDir()
	for name, body := range map[string]string{
		"gh/pack.yaml": "pools:\n  dev: {agent: \"{{.agent}}\", dir: /tmp}\nworkflows:\n  pr-work:\n    repo: \"{{.repo}}\"\n    steps: [{id: s, pool: dev, title: t, prompt: p}]\n",
		"op/pack.yaml": "pools:\n  operator: {type: chat, agent: \"{{.agent}}\", dir: /tmp, admin: [config-reload, config-show, session-list, pool-list, trigger-list, attention]}\n",
	} {
		p := filepath.Join(packs, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
	gh, op := filepath.Join(packs, "gh"), filepath.Join(packs, "op")
	c, _ := rulesController(t, fmt.Sprintf(`projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: %s
    prefix: acphost-
    vars: {repo: elee1766/acphost, agent: a}
  - path: %s
    vars: {agent: a}
`, gh, op))
	ctx := t.Context()
	type call func(context.Context, scope) (any, error)
	granted := map[string]call{
		// The gate only: a full reload needs a running controller.
		"config-reload": func(_ context.Context, sc scope) (any, error) {
			return nil, c.requireGrant(sc, "config-reload", "config reload")
		},
		"config-show": func(ctx context.Context, sc scope) (any, error) { return c.configShow(ctx, sc, struct{}{}) },
		"session-list": func(ctx context.Context, sc scope) (any, error) {
			return c.sessionList(ctx, sc, rpc.SessionListParams{})
		},
		"pool-list":    func(ctx context.Context, sc scope) (any, error) { return c.poolList(ctx, sc, rpc.PoolListParams{}) },
		"trigger-list": func(ctx context.Context, sc scope) (any, error) { return c.triggerList(ctx, sc, struct{}{}) },
		"attention":    func(ctx context.Context, sc scope) (any, error) { return c.attention(ctx, sc, struct{}{}) },
	}
	opSc := scope{actor: "js-op", session: "js-op", project: "js", pool: "operator"}
	devSc := scope{actor: "js-dev", session: "js-dev", project: "js", pool: "acphost-dev"}
	for name, f := range granted {
		_, err := f(ctx, opSc)
		require.NoError(t, err, "operator calls %s", name)
		_, err = f(ctx, devSc)
		require.ErrorIs(t, err, rpc.ErrForbidden, "dev cannot call %s", name)
	}
	// Runs need no grant: any session reads them.
	_, err := c.runList(ctx, opSc, rpc.RunListParams{})
	require.NoError(t, err)

	res, err := c.configShow(ctx, opSc, struct{}{})
	require.NoError(t, err)
	inst := res.(rpc.ConfigShowResult).Instances
	require.Len(t, inst, 2)
	assert.Equal(t, "acphost-", inst[0].Prefix)
	assert.Contains(t, inst[0].Workflows, "acphost-pr-work")
	assert.Equal(t, "elee1766/acphost", inst[0].Vars["repo"])
}
