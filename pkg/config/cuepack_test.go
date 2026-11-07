package config

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/automation"
)

func TestCUEPack(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.CopyFS(dir, os.DirFS("testdata/cue-starlark-example")))
	writeFiles(t, dir, map[string]string{
		"instances.cue": `package city
import demopack "jobshop.example/automation/packs/demo"
packs: qa: demopack.#Pack & {vars: repo: "acme/second"}
pools: "demo-dev": max_sessions: 2
`,
		"packs/demo/hooks.cue": `package demo
#Pack: {
    workflows: callback: {
        on_start: "hooks.star:begin"
        on_step_completed: "hooks.star:completed"
    }
    triggers: wake: {
        when: cron: "0 * * * *"
        handler: "hooks.star:handle"
    }
}
`,
		"packs/demo/hooks.star": "def begin(run):\n    return []\ndef completed(run, step):\n    return []\ndef handle(event, args):\n    return []\n",
	})
	c, err := LoadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	assert.Equal(t, "demo", c.Pools["demo-dev"].Agent)
	assert.Equal(t, 2, c.Pools["demo-dev"].MaxSessions)
	assert.Contains(t, c.Workflows["demo-change"].Steps[0].Prompt, "{{.task}}", "CUE prompt files keep runtime templates")
	assert.Equal(t, "demo-dev", c.Workflows["demo-change"].Steps[0].Pool)
	assert.Contains(t, c.Disabled.Workflows, "qa-change")
	require.Len(t, c.Instances, 2)
	assert.Equal(t, "demo-", c.Instances[0].Prefix)
	assert.Equal(t, "main", c.Instances[0].Vars["branch"])
	assert.Equal(t, false, c.Instances[1].Vars["enabled"])
	assert.Equal(t, []string{"qa-dev"}, c.Instances[1].Pools)
	imported := filepath.Join(dir, "packs/demo/pack.cue")
	assert.Contains(t, c.Sources["pools.demo-dev"], imported)
	assert.Contains(t, c.Sources["pools.demo-dev"], filepath.Join(dir, "instances.cue"))
	assert.Contains(t, c.Sources["pools.demo-dev.prompt"], filepath.Join(dir, "packs/demo/prompt.md"))
	assert.Contains(t, c.Watch.Files, imported)
	assert.Contains(t, c.Watch.Files, filepath.Join(dir, "cue.mod/module.cue"))
	assert.Contains(t, c.Watch.CUEDirs, filepath.Join(dir, "packs/demo"))
	assert.Contains(t, c.Watch.CUEDirs, dir)
	assert.NotContains(t, c.Watch.Dirs, dir)
	script := filepath.Join(dir, "packs/demo/hooks.star")
	assert.Equal(t, script+":begin", c.Workflows["demo-callback"].OnStart)
	assert.Equal(t, script+":completed", c.Workflows["qa-callback"].OnStepCompleted)
	assert.Equal(t, script+":handle", c.Triggers["demo-wake"].Handler)
	assert.Contains(t, c.Sources["workflows.demo-callback.on_start"], script)
	assert.Contains(t, c.Sources["triggers.demo-wake.handler"], script)
	assert.Contains(t, c.Watch.Files, script)

	// Last-good snapshots require accepted hook identities, not only parsed config.
	c.HookDigests = map[string]string{}
	accept := func(entry string) {
		if entry == "" || c.HookDigests[entry] != "" {
			return
		}
		file, function, _ := strings.Cut(entry, ":")
		engine, err := automation.Load(context.Background(), filepath.Dir(file), []string{filepath.Base(file) + ":" + function}, automation.Options{})
		require.NoError(t, err)
		c.HookDigests[entry] = engine.Digest()
	}
	for _, workflow := range c.Workflows {
		accept(workflow.OnStart)
		accept(workflow.OnStepCompleted)
	}
	for _, trigger := range c.Triggers {
		accept(trigger.Handler)
	}
	root := Root{Dir: dir}
	require.NoError(t, SaveLastGood(root.LastGood(), c))
	writeFiles(t, dir, map[string]string{"packs/demo/broken.cue": "package demo\n#Pack: pools: broken: {agent: \"unknown\", dir: \".\", max_sessions: 0}\n"})
	last, err := LoadOrLastGood(root)
	require.ErrorIs(t, err, ErrPack)
	require.NotNil(t, last)
	assert.Equal(t, c.Pools["demo-dev"].Agent, last.Pools["demo-dev"].Agent)
	assert.Equal(t, c.Pools["demo-dev"].MaxSessions, last.Pools["demo-dev"].MaxSessions)
	assert.Equal(t, c.Pools["demo-dev"].Prompt, last.Pools["demo-dev"].Prompt)
	assert.Contains(t, err.Error(), "broken.cue")
}

func TestCUEPackErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		pack string
		want string
	}{
		{"constraint", "pools: dev: {agent: \"a\", dir: \"/w\", max_sessions: >=0 & -1}", "max_sessions"},
		{"incomplete", "pools: dev: {agent: \"a\", dir: \"/w\", max_sessions: int}", "incomplete value"},
		{"host setting", "telemetry: {}", "telemetry"},
		{"unknown work field", "pools: dev: {agent: \"a\", dir: \"/w\", unknown_setting: true}", "unknown_setting"},
		{"missing asset base", "packs: demo: {pools: {}}", "packs.demo.path"},
		{"missing asset directory", "packs: demo: {path: \"missing\"}", "packs.demo.path"},
		{"asset base is a file", "packs: demo: {path: \"pack.cue\"}", "path must name a directory"},
		{"bad instance shape", "packs: demo: []", "packs.demo must be a mapping"},
		{"bad prefix", "packs: demo: {path: \".\", prefix: \"INVALID\"}", "prefix"},
		{"bad vars", "packs: demo: {path: \".\", vars: []}", "vars must be a mapping"},
		{"bad hook entry", "workflows: bad: on_start: \"hook.star\"", "expected file.star:function"},
		{"missing hook file", "workflows: bad: on_start: \"missing.star:begin\"", "workflows.bad.on_start"},
		{"hook is directory", "triggers: bad: handler: \".:handle\"", "script must be a regular file"},
		{"invalid Go model", "resources: browser: {max: -1}", "max must"},
		{"external import", "import \"remote.example/packs/demo\"\npacks: demo: demo.#Pack", "offline"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				FileName:             "pack: pack.cue\n",
				"cue.mod/module.cue": "module: \"city.example/test@v0\"\nlanguage: version: \"v0.17.0\"\ndeps: \"remote.example/packs@v0\": {v: \"v0.1.0\", default: true}\n",
				CUEPackFileName:      "package city\n" + tc.pack + "\n",
			})
			_, err := LoadFile(filepath.Join(dir, FileName))
			require.ErrorContains(t, err, tc.want)
			assert.True(t, errors.Is(err, ErrPack))
			assert.Contains(t, err.Error(), CUEPackFileName)
		})
	}
}
