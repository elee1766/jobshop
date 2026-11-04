package control

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestOperatorGrant (#248): an operator-pool session can call each admin
// method the host-level operator needs across instances; a dev-pool
// session cannot call the granted ones.
func TestOperatorGrant(t *testing.T) {
	op, err := filepath.Abs("../../packs/operator")
	require.NoError(t, err)
	gh, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	c, _ := rulesController(t, fmt.Sprintf(`projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: %s
    prefix: acphost-
    vars: {repo: elee1766/acphost, checkout: /tmp, agent: a, human: wile, check_cmd: make test}
  - path: %s
    vars: {jobshop_repo: elee1766/jobshop, checkout: /tmp, agent: a}
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
		"pool-list":    func(ctx context.Context, sc scope) (any, error) { return c.poolList(ctx, sc, struct{}{}) },
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
	_, err = c.runList(ctx, opSc, rpc.RunListParams{})
	require.NoError(t, err)

	res, err := c.configShow(ctx, opSc, struct{}{})
	require.NoError(t, err)
	inst := res.(rpc.ConfigShowResult).Instances
	require.Len(t, inst, 2)
	assert.Equal(t, "acphost-", inst[0].Prefix)
	assert.Contains(t, inst[0].Workflows, "acphost-pr-work")
	assert.Equal(t, "elee1766/acphost", inst[0].Vars["repo"])
}
