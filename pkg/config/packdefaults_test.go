package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// routedPack mirrors packs/github's implement routing (JOB-59): the step
// pool is the dev_pool run var, whose default names the instance's own
// dev pool.
const routedPack = `defaults:
  dev_pool: "{{.pack_prefix}}dev"
  plain: dev
  label: "{{.pack_prefix}}{{.plain}}"
pools:
  dev: {agent: a, dir: /w}
  dev-sonnet: {agent: a, dir: /w}
workflows:
  pr-work:
    vars: [issue, dev_pool]
    steps:
      - {id: implement, pool: "{{.dev_pool}}", retry_pool: dev, prompt: x, title: "Implement {{.issue}}"}
`

func loadRouted(t *testing.T, packs string) *Config {
	t.Helper()
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"gh/pack.yaml": routedPack, "jobshop.yaml": hostBase + "packs:\n" + packs})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	return c
}

// TestPackDefaultPrefix: a templated default renders with the instance's
// prefix, so a prefixed pr-work run started without --var dev_pool lands
// in <prefix>dev, not the unprefixed instance's dev (#880).
func TestPackDefaultPrefix(t *testing.T) {
	c := loadRouted(t, "  - path: gh\n  - path: gh\n    prefix: gz-\n")
	for _, tc := range []struct{ workflow, want string }{
		{"pr-work", "dev"},
		{"gz-pr-work", "gz-dev"},
	} {
		w := c.Workflows[tc.workflow]
		assert.Equal(t, tc.want, w.PackVars["dev_pool"], tc.workflow)
		vars := map[string]any{"issue": "7"}
		for k, v := range w.PackVars {
			vars[k] = v
		}
		pool, err := c.ResolvePool(&w, w.Steps[0].Pool, vars)
		require.NoError(t, err)
		assert.Equal(t, tc.want, pool, tc.workflow)
		assert.Equal(t, tc.want, w.Steps[0].RetryPool, "literal retry_pool is renamed")
	}
	assert.Equal(t, "gz-dev", c.Instances[1].Vars["dev_pool"])
	assert.Equal(t, "gz-dev", c.Instances[1].Vars["label"], "a template may read a plain default")

	c = loadRouted(t, "  - path: gh\n    prefix: gz-\n    vars: {dev_pool: gz-dev-sonnet}\n")
	assert.Equal(t, "gz-dev-sonnet", c.Workflows["gz-pr-work"].PackVars["dev_pool"], "a given var wins over the default")
}

// TestCheckInstancePool: a step pool another pack instance declares is
// rejected; the workflow's own pools and host pools pass (#880).
func TestCheckInstancePool(t *testing.T) {
	c := loadRouted(t, "  - path: gh\n  - path: gh\n    prefix: gz-\n")
	require.NoError(t, c.CheckInstancePool("gz-pr-work", "gz-dev"))
	require.NoError(t, c.CheckInstancePool("pr-work", "dev-sonnet"))
	require.NoError(t, c.CheckInstancePool("gz-pr-work", "host-pool"), "a pool no instance declares")
	err := c.CheckInstancePool("gz-pr-work", "dev")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `pool "dev" belongs to pack instance gh`)
	assert.Contains(t, err.Error(), "prefix gz-")
	require.Error(t, c.CheckInstancePool("pr-work", "gz-dev-sonnet"))
}
