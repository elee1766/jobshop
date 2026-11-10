package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/execution/automation"
	"github.com/gfx-labs/jobshop/pkg/pack"
)

// TestLastGood: a merged config round-trips through the last good file,
// and only pack errors fall back to it (#227).
func TestLastGood(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"good/pack.yaml": "pools:\n  plan: {agent: a, dir: /w}\nworkflows:\n  pr-work:\n    steps: [{id: s, pool: plan, title: t, prompt: \"in {{.repo}}\"}]\n",
		"p/pack.yaml":    "pools:\n  dev: {agent: a, dir: /w, prompt: \"hi {{.who}}\"}\n",
		"jobshop.yaml":   hostBase + "packs:\n  - path: good\n    vars: {repo: /src}\n",
	})
	r := Root{Dir: dir}
	c, err := Load(r)
	require.NoError(t, err)
	require.NoError(t, SaveLastGood(r.LastGood(), c))
	lg, err := LoadLastGood(r.LastGood())
	require.NoError(t, err)
	assert.Len(t, lg.Pools, len(c.Pools))
	assert.Equal(t, c.Workflows["pr-work"].Steps[0].Prompt, lg.Workflows["pr-work"].Steps[0].Prompt)
	assert.Equal(t, c.Workflows["pr-work"].PackVars, lg.Workflows["pr-work"].PackVars)
	assert.Equal(t, c.Instances, lg.Instances)

	// A pack with a var this binary does not set: last good, plus the error.
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: p\n"})
	got, err := LoadOrLastGood(r)
	require.ErrorIs(t, err, pack.ErrPack)
	assert.ErrorContains(t, err, `var "who" is not set`)
	require.NotNil(t, got)
	assert.Contains(t, got.Pools, "plan", "the saved config")

	// A pack entry the validator rejects is the pack's too.
	writeFiles(t, dir, map[string]string{"p/pack.yaml": "pools:\n  dev: {agent: nosuch, dir: /w}\n"})
	got, err = LoadOrLastGood(r)
	require.ErrorIs(t, err, pack.ErrPack)
	require.NotNil(t, got)

	// Host errors still fail.
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "bogus: 1\n"})
	got, err = LoadOrLastGood(r)
	require.Error(t, err)
	assert.False(t, errors.Is(err, pack.ErrPack))
	assert.Nil(t, got)

	// No last good: the pack error fails the start.
	require.NoError(t, os.Remove(r.LastGood()))
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: p\n"})
	got, err = LoadOrLastGood(r)
	require.ErrorIs(t, err, pack.ErrPack)
	assert.Nil(t, got)
}

func TestLastGoodHookMetadata(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		FileName: hostBase + "packs:\n  - {path: app, prefix: first-}\n  - {path: app, prefix: second-}\n",
		"app/pack.yaml": `pools:
  dev: {agent: a, dir: /w}
workflows:
  change:
    on_start: hooks.star:begin
    on_step_completed: hooks.star:completed
triggers:
  incoming:
    when: {cron: "0 * * * *"}
    handler: hooks.star:handle
`,
		"app/hooks.star": "def begin(run):\n    return []\ndef completed(run, step):\n    return []\ndef handle(event, args):\n    return []\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	c.HookDigests = map[string]string{}
	for _, entry := range []string{c.Workflows["first-change"].OnStart, c.Workflows["first-change"].OnStepCompleted, c.Triggers["first-incoming"].Handler} {
		file, fn, _ := strings.Cut(entry, ":")
		engine, err := automation.Load(context.Background(), filepath.Dir(file), []string{filepath.Base(file) + ":" + fn}, automation.Options{})
		require.NoError(t, err)
		c.HookDigests[entry] = engine.Digest()
	}
	root := Root{Dir: dir}
	require.NoError(t, SaveLastGood(root.LastGood(), c))

	// Recovery must not need the source pack or scripts to restore its authority.
	require.NoError(t, os.Remove(filepath.Join(dir, "app/pack.yaml")))
	require.NoError(t, os.Remove(filepath.Join(dir, "app/hooks.star")))
	last, err := LoadOrLastGood(root)
	require.ErrorIs(t, err, pack.ErrPack)
	require.NotNil(t, last)
	assert.Equal(t, c.Instances, last.Instances)
	assert.Equal(t, c.HookDigests, last.HookDigests)
	assert.Equal(t, last.Pools["first-dev"].Project, last.Pools["second-dev"].Project)
	for _, owner := range []string{"workflows.first-change", "triggers.first-incoming"} {
		for _, reference := range []struct{ kind, local, sibling string }{
			{"pool", "dev", "second-dev"},
			{"workflow", "change", "second-change"},
		} {
			name, err := last.ResolveHookReference(owner, reference.kind, reference.local)
			require.NoError(t, err)
			assert.Equal(t, "first-"+reference.local, name)
			_, err = last.ResolveHookReference(owner, reference.kind, reference.sibling)
			require.ErrorContains(t, err, "outside")
		}
	}

	saved, err := os.ReadFile(root.LastGood())
	require.NoError(t, err)
	for _, tc := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"missing metadata", func(snapshot map[string]any) { delete(snapshot, "metadata") }},
		{"missing instances", func(snapshot map[string]any) { delete(snapshot["metadata"].(map[string]any), "instances") }},
		{"missing manifest", func(snapshot map[string]any) { delete(snapshot["metadata"].(map[string]any), "hook_digests") }},
		{"missing entry digest", func(snapshot map[string]any) {
			delete(snapshot["metadata"].(map[string]any)["hook_digests"].(map[string]any), c.Triggers["first-incoming"].Handler)
		}},
		{"unsupported version", func(snapshot map[string]any) { snapshot["metadata"].(map[string]any)["version"] = 2 }},
		{"unknown metadata field", func(snapshot map[string]any) { snapshot["metadata"].(map[string]any)["unknown"] = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var snapshot map[string]any
			require.NoError(t, yaml.Unmarshal(saved, &snapshot))
			tc.mutate(snapshot)
			data, err := yaml.Marshal(snapshot)
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(root.LastGood(), data, 0o600))
			_, err = LoadLastGood(root.LastGood())
			require.Error(t, err)
		})
	}

	flat := *c
	flat.Pack, flat.Packs = "", nil
	legacy, err := yaml.Marshal(&flat)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(root.LastGood(), legacy, 0o600))
	_, err = LoadLastGood(root.LastGood())
	require.ErrorContains(t, err, "metadata")

	// An incomplete accepted manifest cannot replace the previously saved snapshot.
	require.NoError(t, os.WriteFile(root.LastGood(), saved, 0o600))
	delete(c.HookDigests, c.Triggers["first-incoming"].Handler)
	require.Error(t, SaveLastGood(root.LastGood(), c))
	unchanged, err := os.ReadFile(root.LastGood())
	require.NoError(t, err)
	assert.Equal(t, saved, unchanged)
}

func TestLastGoodLegacyWithoutHooks(t *testing.T) {
	path := filepath.Join(t.TempDir(), LastGoodName)
	require.NoError(t, os.WriteFile(path, []byte(hostBase+"pools:\n  dev: {agent: a, dir: /w}\n"), 0o600))
	c, err := LoadLastGood(path)
	require.NoError(t, err)
	assert.Contains(t, c.Pools, "dev")
}
