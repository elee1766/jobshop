package pack_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/pack"
)

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
}

// assemble reads dir/jobshop.yaml the way config does: decode, then hand the
// document and its path to pack with no knowledge of the config types.
func assemble(t *testing.T, dir string, opts pack.Options) (*pack.Result, map[string]any, error) {
	t.Helper()
	path := filepath.Join(dir, "jobshop.yaml")
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	top, err := pack.DecodeMap(b)
	require.NoError(t, err)
	res, err := pack.Assemble(pack.Input{Path: path, Top: top, Watch: []string{path}}, opts)
	if err != nil {
		return nil, nil, err
	}
	var doc map[string]any
	require.NoError(t, yaml.Unmarshal(res.Document, &doc))
	return res, doc, nil
}

func TestAssembleMergesInstancesWithPrefixAndSources(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"jobshop.yaml": "packs:\n  - path: gh\n  - path: gh\n    prefix: gz-\n    vars: {repo: o/r}\nagents:\n  a: {command: x}\n",
		"gh/pack.yaml": "defaults: {repo: none, dev_pool: \"{{.pack_prefix}}dev\"}\n" +
			"pools:\n  dev: {agent: a, repo: \"{{.repo}}\", prompt_file: dev.md}\n" +
			"workflows:\n  w:\n    steps:\n      - {id: s, pool: \"{{.dev_pool}}\", title: \"T {{.issue}}\"}\n",
		"gh/dev.md": "repo {{.repo}}\n",
	})
	res, doc, err := assemble(t, dir, pack.Options{StepTemplateKeys: []string{"title", "pool"}})
	require.NoError(t, err)

	pools := doc["pools"].(map[string]any)
	assert.Contains(t, pools, "dev")
	assert.Contains(t, pools, "gz-dev")
	assert.Equal(t, "o/r", pools["gz-dev"].(map[string]any)["repo"])
	assert.Equal(t, "repo o/r\n", pools["gz-dev"].(map[string]any)["prompt"])

	// Step templates survive for run start, the instance's own pool is a default.
	step := doc["workflows"].(map[string]any)["gz-w"].(map[string]any)["steps"].([]any)[0].(map[string]any)
	assert.Equal(t, "{{.dev_pool}}", step["pool"])
	assert.Equal(t, "T {{.issue}}", step["title"])
	assert.Equal(t, "gz-dev", doc["workflows"].(map[string]any)["gz-w"].(map[string]any)["pack_vars"].(map[string]any)["dev_pool"])

	require.Len(t, res.Instances, 2)
	assert.Equal(t, "gz-", res.Instances[1].Prefix)
	assert.Equal(t, []string{"gz-dev"}, res.Instances[1].Pools)
	assert.Equal(t, []string{"instance:gz", "pack:gh"}, res.Instances[1].Labels())

	assert.Equal(t, []string{filepath.Join(dir, "gh", "pack.yaml")}, res.Sources["pools.gz-dev"])
	assert.Equal(t, []string{filepath.Join(dir, "gh", "dev.md")}, res.Sources["pools.dev.prompt"])
	assert.Contains(t, res.Watch.Dirs, filepath.Join(dir, "gh"))
	assert.Contains(t, res.Watch.Files, filepath.Join(dir, "jobshop.yaml"))
}

func TestAssembleKeepsLateVarsAndEscalationVars(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"jobshop.yaml": "packs:\n  - path: p\n    vars: {repo: o/r}\n",
		"p/pack.yaml": "triggers:\n  t:\n    run: {workflow: w, vars: {sha: \"{{.check_output}}\", s: \"{{.signal}}\"}}\n" +
			"escalation:\n  prompt: \"{{.ticket}} in {{.repo}}\"\n",
	})
	_, doc, err := assemble(t, dir, pack.Options{
		LateVars:       map[string]any{"check_output": pack.Placeholder("check_output"), "signal": pack.Placeholder("signal")},
		EscalationVars: []string{"ticket"},
	})
	require.NoError(t, err)
	run := doc["triggers"].(map[string]any)["t"].(map[string]any)["run"].(map[string]any)["vars"].(map[string]any)
	assert.Equal(t, "{{.check_output}}", run["sha"])
	assert.Equal(t, "{{.signal}}", run["s"])
	assert.Equal(t, "{{.ticket}} in o/r", doc["escalation"].(map[string]any)["prompt"])
}

func TestAssembleCheckLayerNamesThePackFile(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, map[string]string{
		"jobshop.yaml": "packs:\n  - path: p\n",
		"p/pack.yaml":  "pools:\n  dev: {bogus: 1}\n",
	})
	_, _, err := assemble(t, dir, pack.Options{CheckLayer: func(m map[string]any, _ bool) error {
		if _, ok := m["pools"].(map[string]any)["dev"].(map[string]any)["bogus"]; ok {
			return errors.New("unknown key bogus")
		}
		return nil
	}})
	require.ErrorContains(t, err, filepath.Join(dir, "p", "pack.yaml"))
	require.ErrorContains(t, err, "unknown key bogus")
	assert.ErrorIs(t, err, pack.ErrPack, "a bad pack file is a pack error")
}

func TestAssembleErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		files map[string]string
		want  string
	}{
		"cycle": {
			map[string]string{"jobshop.yaml": "packs:\n  - path: a\n", "a/pack.yaml": "extends: [{path: ../a}]\n"},
			"extends nested more than",
		},
		"removed syntax": {
			map[string]string{"jobshop.yaml": "packs:\n  - path: a\n", "a/pack.yaml": "pools:\n  p: {dir: \"${x}\"}\n"},
			"the ${name} syntax was removed",
		},
		"bad prefix": {
			map[string]string{"jobshop.yaml": "packs:\n  - path: a\n    prefix: BAD\n", "a/pack.yaml": "pools: {}\n"},
			"must match",
		},
		"missing var": {
			map[string]string{"jobshop.yaml": "packs:\n  - path: a\n", "a/pack.yaml": "pools:\n  p: {dir: \"{{.nope}}\"}\n"},
			`var "nope" is not set`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			write(t, dir, tc.files)
			_, _, err := assemble(t, dir, pack.Options{})
			require.ErrorContains(t, err, tc.want)
		})
	}
}

func TestCheckInstancePool(t *testing.T) {
	instances := []pack.Instance{
		{Path: "gh", Pools: []string{"dev"}, Workflows: []string{"pr-work"}},
		{Path: "gh", Prefix: "gz-", Pools: []string{"gz-dev"}, Workflows: []string{"gz-pr-work"}},
	}
	require.NoError(t, pack.CheckInstancePool(instances, "gz-pr-work", "gz-dev"))
	require.NoError(t, pack.CheckInstancePool(instances, "gz-pr-work", "host-pool"), "a pool no instance declares")
	err := pack.CheckInstancePool(instances, "gz-pr-work", "dev")
	require.ErrorContains(t, err, `pool "dev" belongs to pack instance gh`)
	require.ErrorContains(t, err, "prefix gz-")
}

func TestIgnoredChangeAndPrefix(t *testing.T) {
	for _, p := range []string{"/p/.x.swp", "/p/4913", "/p/x~", "/p/.#x"} {
		assert.True(t, pack.IgnoredChange(p), p)
	}
	assert.False(t, pack.IgnoredChange("/p/pack.yaml"))
	assert.True(t, pack.ValidPrefix("gz-"))
	assert.False(t, pack.ValidPrefix("Gz"))
}
