package control

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigWatchCUEPackages(t *testing.T) {
	t.Parallel()
	c, clk, dir, res := watchController(t)
	packDir := filepath.Join(dir, "cue")
	require.NoError(t, os.Mkdir(packDir, 0o755))
	require.NoError(t, os.CopyFS(packDir, os.DirFS("../config/testdata/cue-starlark-example")))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "jobshop.yaml"), []byte("pack: cue/pack.cue\nreload: {watch: true, debounce: 2s}\n"), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	require.Contains(t, c.config().Pools, "demo-dev")

	prompt := filepath.Join(packDir, "packs/demo/prompt.md")
	require.NoError(t, os.WriteFile(prompt, []byte("Edited native pack prompt."), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	assert.Equal(t, "Edited native pack prompt.", c.config().Pools["demo-dev"].Prompt)

	// New package members must reload even though they were absent at load.
	sibling := filepath.Join(packDir, "resources.cue")
	require.NoError(t, os.WriteFile(sibling, []byte("package city\nresources: browser: max: 2\n"), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	assert.Equal(t, 2, c.config().ResourceCaps()["browser"])

	imported := filepath.Join(packDir, "packs/demo/constraint.cue")
	require.NoError(t, os.WriteFile(imported, []byte("package demo\n#Pack: vars: repo: int\n"), 0o600))
	require.ErrorContains(t, fireDebounce(t, clk, res).err, "constraint.cue")
	assert.Equal(t, "Edited native pack prompt.", c.config().Pools["demo-dev"].Prompt)
	require.NoError(t, os.Remove(imported))
	require.NoError(t, fireDebounce(t, clk, res).err)
}
