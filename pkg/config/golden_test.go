package config

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata golden files")

// TestPacksGolden renders the in-repo packs (pools, triggers, rules,
// hooks and every step with placeholder run vars) and compares them with
// testdata/packs.golden, so a change of template syntax (#383) cannot
// change the text agents get.
func TestPacksGolden(t *testing.T) {
	dir := t.TempDir()
	abs := func(p string) string {
		a, err := filepath.Abs("../../packs/" + p)
		require.NoError(t, err)
		return a
	}
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + abs("selfdev") + `
    vars: {repo: /src/jobshop, agent: a, memory_dir: /mem}
  - path: ` + abs("github") + `
    prefix: acphost-
    vars: {repo: elee1766/acphost, checkout: /src/acphost, agent: a, human: wile, check_cmd: make test, authors: "x,y"}
  - path: ` + abs("github") + `
    prefix: eng-
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test, tracker: linear, linear_workspace: acme, linear_team: ENG, linear_target: ENG-1, linear_human: bob}
  - path: ` + abs("operator") + `
    vars: {jobshop_repo: acme/jobshop-fork, checkout: /src/jobshop, agent: a}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	steps := map[string][]map[string]string{}
	names := make([]string, 0, len(c.Workflows))
	for n := range c.Workflows {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		w := c.Workflows[n]
		vars := map[string]any{}
		for k, v := range w.PackVars {
			vars[k] = v
		}
		for _, k := range w.Vars {
			vars[k] = "<" + k + ">"
		}
		for _, s := range w.Steps {
			title, err := w.Render(s.Title, vars)
			require.NoError(t, err, n)
			body, err := w.Render(s.Prompt, vars)
			require.NoError(t, err, n)
			steps[n] = append(steps[n], map[string]string{"id": s.ID, "title": title, "prompt": body})
		}
	}
	// Trigger run vars and key render when the trigger fires.
	fire := map[string]any{CheckOutputVar: "<check_output>"}
	for _, k := range StallVars {
		fire[k] = "<" + k + ">"
	}
	for n, tr := range c.Triggers {
		if tr.Run == nil {
			continue
		}
		r := *tr.Run
		r.Vars = map[string]string{}
		for k, v := range tr.Run.Vars {
			r.Vars[k], err = tr.Run.Fill(v, fire)
			require.NoError(t, err, n)
		}
		r.Key, err = tr.Run.Fill(r.Key, fire)
		require.NoError(t, err, n)
		r.Template = ""
		tr.Run = &r
		c.Triggers[n] = tr
	}
	out, err := yaml.Marshal(map[string]any{
		"pools": c.Pools, "triggers": c.Triggers, "steps": steps,
		"prompt_rules": c.PromptRules, "hooks": c.Hooks,
	})
	require.NoError(t, err)
	golden := filepath.Join("testdata", "packs.golden")
	if *updateGolden {
		require.NoError(t, os.MkdirAll("testdata", 0o750))
		require.NoError(t, os.WriteFile(golden, out, 0o600))
	}
	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.Equal(t, string(want), string(out), "rendered packs changed; rerun with -update if intended")
}
