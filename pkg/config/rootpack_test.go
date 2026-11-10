package config

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/pack"
)

func TestRootPack(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"jobshop.yaml": "pack: work/pack.yaml\ntelemetry: {instance: root-pack-test}\n",
		"work/pack.yaml": `defaults: {child: child, agent: a}
default_project: js
projects:
  js: {}
agents:
  a: {command: x}
resources:
  browser: {max: 2}
prompt_rules:
  - name: shared
    at: system
    file: prompts/shared.md
packs:
  - path: "{{.child}}"
    prefix: app-
    vars: {agent: "{{.agent}}"}
pools:
  app-dev: {max_sessions: 2}
`,
		"work/prompts/shared.md": "Shared instructions.",
		"work/child/pack.yaml": `extends:
  - path: ../base
    vars: {agent: "{{.agent}}"}
workflows:
  change:
    steps:
      - id: implement
        pool: dev
        needs_resources: [browser]
        title: Implement
        prompt: Work.
`,
		"work/base/pack.yaml": "pools:\n  dev: {agent: '{{.agent}}', dir: /w, max_sessions: 1}\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	assert.Equal(t, "js", c.Pools["app-dev"].Project)
	assert.Equal(t, "a", c.Pools["app-dev"].Agent)
	assert.Equal(t, 2, c.Pools["app-dev"].MaxSessions)
	assert.Equal(t, 2, c.ResourceCaps()["browser"])
	assert.Equal(t, "Shared instructions.", c.PromptRules[0].Text)
	assert.Equal(t, "app-dev", c.Workflows["app-change"].Steps[0].Pool)
	assert.Equal(t, []string{"browser"}, c.Workflows["app-change"].Steps[0].NeedsResources)
	require.Len(t, c.Instances, 1)
	assert.Equal(t, "app-", c.Instances[0].Prefix)
	assert.Equal(t, []string{filepath.Join(dir, "work/base/pack.yaml"), filepath.Join(dir, "work/pack.yaml")}, c.Sources["pools.app-dev"])
	assert.Equal(t, []string{filepath.Join(dir, "work/pack.yaml")}, c.Sources["projects.js"])
	assert.Equal(t, []string{filepath.Join(dir, "work/pack.yaml")}, c.Sources["resources.browser"])
	assert.Contains(t, c.Watch.Files, filepath.Join(dir, "work/pack.yaml"))
	assert.Contains(t, c.Watch.Files, filepath.Join(dir, "work/prompts/shared.md"))
	assert.Contains(t, c.Watch.Dirs, filepath.Join(dir, "work/child"))
	assert.Contains(t, c.Watch.Dirs, filepath.Join(dir, "work/base"))
	assert.NotContains(t, c.Watch.Dirs, filepath.Join(dir, "work"), "do not recursively watch the whole root checkout")
	require.Len(t, c.Warnings, 3)
	for _, warning := range c.Warnings {
		assert.Contains(t, warning, "YAML packs are deprecated")
	}

	writeFiles(t, dir, map[string]string{FileName: "pack: work/pack.yaml\npools: {app-dev: {max_sessions: 3}}\n"})
	legacy, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	assert.Equal(t, 3, legacy.Pools["app-dev"].MaxSessions)
	assert.Equal(t, "a", legacy.Pools["app-dev"].Agent)
	require.Len(t, legacy.Warnings, 4)
	assert.Contains(t, legacy.Warnings[0], "service work definitions are deprecated")

	root := Root{Dir: dir}
	require.NoError(t, SaveLastGood(root.LastGood(), c))
	writeFiles(t, dir, map[string]string{"work/pack.yaml": "store: {driver: sqlite}\n"})
	last, err := LoadOrLastGood(root)
	require.ErrorIs(t, err, pack.ErrPack)
	require.NotNil(t, last)
	assert.Equal(t, c.Pools["app-dev"].Agent, last.Pools["app-dev"].Agent)
	assert.Equal(t, c.Pools["app-dev"].Project, last.Pools["app-dev"].Project)
	assert.Equal(t, c.Pools["app-dev"].MaxSessions, last.Pools["app-dev"].MaxSessions)
	assert.Equal(t, c.Resources, last.Resources)
}

func TestRootPackErrors(t *testing.T) {
	for _, tc := range []struct {
		name      string
		service   string
		pack      string
		want      string
		packError bool
	}{
		{"missing file", "pack: missing.yaml\n", "", "missing.yaml", true},
		{"invalid reference", "pack: [pack.yaml]\n", "", "pack", false},
		{"empty reference", "pack: ''\n", "", "file path is required", false},
		{"host-only setting", "pack: pack.yaml\n", "telemetry: {}\n", "telemetry", true},
		{"invalid child", "pack: pack.yaml\n", "packs: [{prefix: app-}]\n", "path is required", true},
		{"custom filename validation", "pack: work.yaml\n", "projects: {bad/name: {}}\n", "project name", true},
		{"project validation", "pack: pack.yaml\n", "projects: {bad/name: {}}\n", "project name", true},
		{"resource validation", "pack: pack.yaml\n", "resources: {browser: {max: -1}}\n", "max must", true},
		{"host validation is fatal", "pack: pack.yaml\nstore: {driver: unknown}\n", "", "unknown driver", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{FileName: tc.service, pack.FileName: tc.pack, "work.yaml": tc.pack})
			_, err := LoadFile(filepath.Join(dir, FileName))
			require.ErrorContains(t, err, tc.want)
			assert.Equal(t, tc.packError, errors.Is(err, pack.ErrPack))
		})
	}
}
