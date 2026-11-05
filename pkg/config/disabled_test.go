package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// disabledPack is a fixture pack (not packs/, JOB-2) whose pool, workflow
// and trigger are switched by vars.
const disabledPack = `defaults: {pool_on: true, wf_on: true, tr_on: true}
pools:
  dev:
    enabled: "{{.pool_on}}"
    agent: a
    dir: /w
workflows:
  flow:
    enabled: "{{.wf_on}}"
    steps:
      - {id: one, pool: dev, title: t, prompt: p}
triggers:
  kick:
    enabled: "{{.tr_on}}"
    when: {cron: "0 * * * *"}
    run: {workflow: flow, key: k}
`

func loadDisabled(t *testing.T, vars, extra string) (*Config, error) {
	t.Helper()
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  disabledPack,
		"jobshop.yaml": hostBase + "packs:\n  - path: p\n    vars: " + vars + "\n" + extra,
	})
	return LoadFile(filepath.Join(dir, "jobshop.yaml"))
}

func TestPackEnabled(t *testing.T) {
	c, err := loadDisabled(t, "{}", "")
	require.NoError(t, err)
	assert.Contains(t, c.Pools, "dev")
	assert.Contains(t, c.Workflows, "flow")
	assert.Contains(t, c.Triggers, "kick")
	assert.Empty(t, c.Disabled)

	c, err = loadDisabled(t, "{pool_on: false, wf_on: false, tr_on: false}", "")
	require.NoError(t, err)
	assert.Empty(t, c.Pools)
	assert.Empty(t, c.Workflows)
	assert.Empty(t, c.Triggers)
	assert.Equal(t, Disabled{Pools: []string{"dev"}, Workflows: []string{"flow"}, Triggers: []string{"kick"}}, c.Disabled)

	_, err = loadDisabled(t, "{tr_on: false}", "workflows:\n  flow:\n    enabled: false\n")
	require.NoError(t, err, "host overrides the pack's enabled")
}

func TestPackDisabledRefs(t *testing.T) {
	_, err := loadDisabled(t, "{wf_on: false}", "")
	require.ErrorContains(t, err, "trigger kick: run: workflow flow is disabled")

	_, err = loadDisabled(t, "{pool_on: false, tr_on: false}", "")
	require.ErrorContains(t, err, "workflow flow: step one: pool dev is disabled")

	_, err = loadDisabled(t, "{pool_on: false, wf_on: false, tr_on: false}", "hooks:\n  - {on: claim, command: x, pools: [dev]}\n")
	require.ErrorContains(t, err, "hooks[0]: pool dev is disabled")
}

// A disabled entry is still validated, so it fails to load when broken.
func TestPackDisabledValidated(t *testing.T) {
	_, err := loadDisabled(t, "{pool_on: false, wf_on: false, tr_on: false}", "pools:\n  dev:\n    agent: nosuch\n")
	require.ErrorContains(t, err, "nosuch")
}
