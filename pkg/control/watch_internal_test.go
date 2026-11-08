package control

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

type watchResult struct {
	files []string
	err   error
}

// watchController is a finding controller whose config watches a pack, with
// its watcher running on the fake clock.
func watchController(t *testing.T) (*Controller, *scanClock, string, <-chan watchResult) {
	t.Helper()
	dir := t.TempDir()
	write := func(name, s string) {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte(s), 0o600))
	}
	write("p/pack.yaml", "pools:\n  dev: {agent: a, dir: /w, max_sessions: 0, prompt_file: prompt.md}\n")
	write("p/prompt.md", "one")
	write("q/pack.yaml", "pools:\n  qa: {agent: a, dir: /w, max_sessions: 0, prompt: q}\n")
	c, clk, _ := scanControllerYAML(t, "projects:\n  js: {}\nagents:\n  a: {command: x}\n# %[1]s\n")
	// Its config is a placeholder: point it at ours and load that.
	write("jobshop.yaml", "pack: pack.yaml\nreload: {watch: true, debounce: 2s}\n")
	write("pack.yaml", "projects:\n  js: {}\nagents:\n  a: {command: x}\npacks:\n  - path: p\n")
	c.root.Dir = dir
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	c.configSwapped = make(chan struct{}, 1)
	require.NoError(t, c.reload())
	<-c.configSwapped
	c.poolQuarantined = map[string]string{}
	res := make(chan watchResult, 4)
	c.watcher = newConfigWatcher(c)
	c.watcher.reloaded = func(files []string, err error) { res <- watchResult{files, err} }
	ready := make(chan struct{})
	go func() {
		c.watcher.sync()
		close(ready)
		c.watcher.run()
	}()
	<-ready
	return c, clk, dir, res
}

// fireDebounce moves the fake clock past the debounce until the watcher
// reloads. One write can send several events, each restarting the
// debounce, so one advance may not be enough.
func fireDebounce(t *testing.T, clk *scanClock, res <-chan watchResult) watchResult {
	t.Helper()
	for range 50 {
		require.True(t, clk.WaitFor(1, t.Context().Done()))
		clk.advance(2 * time.Second)
		// Either the watcher reloads, or a later event re-armed the
		// debounce and the loop advances again.
		stop, rearmed := make(chan struct{}), make(chan struct{})
		go func() {
			if clk.WaitFor(1, stop) {
				close(rearmed)
			}
		}()
		select {
		case r := <-res:
			close(stop)
			return r
		case <-rearmed:
		}
	}
	t.Fatal("no watch reload")
	return watchResult{}
}

func prompt(c *Controller) string { return c.config().Pools["dev"].Prompt }

func TestConfigWatchReloadsPromptEdit(t *testing.T) {
	c, clk, dir, res := watchController(t)
	require.Equal(t, "one", prompt(c))
	// A vim swap file does not arm the debounce; the edit after it does.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "p", ".prompt.md.swp"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "p", "prompt.md"), []byte("two"), 0o600))
	r := fireDebounce(t, clk, res)
	require.NoError(t, r.err)
	assert.Equal(t, []string{filepath.Join(dir, "p", "prompt.md")}, r.files, "the swap file is ignored")
	assert.Equal(t, "two", prompt(c))
}

func TestConfigWatchBadPackRaisesAttention(t *testing.T) {
	c, clk, dir, res := watchController(t)
	pack := filepath.Join(dir, "p", "pack.yaml")
	require.NoError(t, os.WriteFile(pack, []byte("pools: [broken\n"), 0o600))
	require.Error(t, fireDebounce(t, clk, res).err)
	assert.Equal(t, "one", prompt(c), "the running config stays")
	items, err := c.attentionItems(t.Context(), c.config())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionConfigInvalid, items[0].Kind)

	require.NoError(t, os.WriteFile(pack, []byte("pools:\n  dev: {agent: a, dir: /w, max_sessions: 0, prompt: fixed}\n"), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	assert.Equal(t, "fixed", prompt(c))
	items, err = c.attentionItems(t.Context(), c.config())
	require.NoError(t, err)
	assert.Empty(t, items, "a good change clears it")
}

func TestConfigWatchFollowsNewPack(t *testing.T) {
	c, clk, dir, res := watchController(t)
	cfgFile := filepath.Join(dir, "pack.yaml")
	require.NoError(t, os.WriteFile(cfgFile, []byte("projects:\n  js: {}\nagents:\n  a: {command: x}\npacks:\n  - path: p\n  - path: q\n"), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	require.Contains(t, c.config().Pools, "qa")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "q", "pack.yaml"), []byte("pools:\n  qa: {agent: a, dir: /w, max_sessions: 0, prompt: q2}\n"), 0o600))
	require.NoError(t, fireDebounce(t, clk, res).err)
	assert.Equal(t, "q2", c.config().Pools["qa"].Prompt)
}
