package control_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// TestPoolListLabels (JOB-81): pool.list returns each pool's labels and
// project, and the label filter keeps only matching pools.
func TestPoolListLabels(t *testing.T) {
	root, start := poolRoot(t, "1m")
	cfg := fmt.Sprintf("projects:\n  js: {}\nagents:\n  fake:\n    command: %q\n    args: [fake-agent]\npools:\n"+
		"  dev: {project: js, agent: fake, max_sessions: 1, dir: %q, labels: [review, sonnet]}\n"+
		"  plan: {project: js, agent: fake, max_sessions: 1, dir: %q, labels: [review]}\n"+
		"  bare: {project: js, agent: fake, max_sessions: 1, dir: %q}\n", os.Args[0], root.Dir, root.Dir, root.Dir)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, config.FileName), []byte(cfg), 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	names := func(p rpc.PoolListParams) (map[string][]string, rpc.PoolStatusResult) {
		var st rpc.PoolStatusResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, p, &st))
		out := map[string][]string{}
		for _, s := range st.Pools {
			out[s.Name] = s.Labels
			assert.Equal(t, "js", s.Project)
		}
		return out, st
	}
	all, _ := names(rpc.PoolListParams{})
	assert.Equal(t, map[string][]string{"dev": {"review", "sonnet"}, "plan": {"review"}, "bare": nil}, all)
	got, _ := names(rpc.PoolListParams{Label: []string{"review"}})
	assert.ElementsMatch(t, []string{"dev", "plan"}, keysOf(got))
	got, _ = names(rpc.PoolListParams{Label: []string{"review", "sonnet"}})
	assert.ElementsMatch(t, []string{"dev"}, keysOf(got))
	got, _ = names(rpc.PoolListParams{Label: []string{"nope"}})
	assert.Empty(t, got)
}

func keysOf[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
