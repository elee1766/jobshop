package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGoTemplatePack: a template: go pack (#383) renders conditionals,
// defaults and loops, keeps whole-value vars typed, escapes a literal
// {{, defers run vars to run start, and loads next to a ${} pack.
func TestGoTemplatePack(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"g/pack.yaml": `template: go
defaults: {n: 2, team: "", repos: [a, b], cmd: ""}
pools:
  dev:
    agent: a
    repo: /w
    max_sessions: "{{.n}}"
    prompt_file: dev.md
triggers:
  t:
    when: {check: "echo x"}
    run: {workflow: w, vars: {sha: "{{.check_output}}"}, key: "k-{{.check_output}}"}
workflows:
  w:
    vars: [sha]
    steps:
      - {id: s, pool: dev, title: "Fix {{.sha}}", prompt: "{{if .team}}team {{.team}}{{else}}no team{{end}} at {{.sha}}"}
`,
		"g/dev.md": "{{range .repos}}- {{.}}\n{{end}}run {{.cmd | default \"go test ./...\"}}\nGitHub: ${{\"{{\"}} github.sha }}\n",
		"d/pack.yaml": `pools:
  old: {agent: a, repo: /w, max_sessions: "${n}", prompt: "old ${n} $$x"}
`,
		"jobshop.yaml": hostBase + "packs:\n  - path: g\n  - path: d\n    vars: {n: 3}\n",
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 2, c.Pools["dev"].MaxSessions, "whole-value var stays a number")
	assert.Equal(t, "- a\n- b\nrun go test ./...\nGitHub: ${{ github.sha }}\n", c.Pools["dev"].Prompt)
	assert.Equal(t, 3, c.Pools["old"].MaxSessions)
	assert.Equal(t, "old 3 $x", c.Pools["old"].Prompt, "${} packs still load")
	assert.Empty(t, c.Warnings)

	w := c.Workflows["w"]
	assert.Equal(t, SyntaxGo, w.Template)
	vars := map[string]any{"sha": "abc"}
	for k, v := range w.PackVars {
		vars[k] = v
	}
	title, err := w.Render(w.Steps[0].Title, vars)
	require.NoError(t, err)
	assert.Equal(t, "Fix abc", title)
	body, err := w.Render(w.Steps[0].Prompt, vars)
	require.NoError(t, err)
	assert.Equal(t, "no team at abc", body)
	r := c.Triggers["t"].Run
	key, err := r.Fill(r.Key, map[string]any{CheckOutputVar: "deadbeef"})
	require.NoError(t, err)
	assert.Equal(t, "k-deadbeef", key)
}

func TestGoTemplatePackErrors(t *testing.T) {
	for _, tt := range []struct {
		name, pack string
		files      map[string]string
		want       []string
	}{
		{
			"missing var names the field", "template: go\npools:\n  dev: {agent: a, repo: /w, prompt: \"x {{.nope}}\"}\n", nil,
			[]string{"pack.yaml", "pools.dev.prompt", `var "nope" is not set`},
		},
		{
			"missing var in a prompt file", "template: go\npools:\n  dev: {agent: a, repo: /w, prompt_file: p.md}\n",
			map[string]string{"p/p.md": "{{.nope}}"},
			[]string{"pools.dev.prompt_file", `var "nope" is not set`},
		},
		{
			"syntax error", "template: go\npools:\n  dev: {agent: a, repo: /w, prompt: \"{{if .x}\"}\n", nil,
			[]string{"pools.dev.prompt", "template"},
		},
		{
			"unknown step var", "template: go\npools:\n  dev: {agent: a, repo: /w, prompt: x}\nworkflows:\n  w:\n    steps: [{id: s, pool: dev, title: t, prompt: \"{{.nope}}\"}]\n", nil,
			[]string{"step s: prompt uses {{.nope}}"},
		},
		{
			"trigger var", "template: go\npools:\n  dev: {agent: a, repo: /w, prompt: x}\nworkflows:\n  w:\n    vars: [v]\n    steps: [{id: s, pool: dev, title: t, prompt: p}]\ntriggers:\n  t:\n    when: {check: x}\n    run: {workflow: w, vars: {v: \"{{.other}}\"}}\n", nil,
			[]string{`var "other" is not set`},
		},
		{"bad syntax name", "template: jinja\n", nil, []string{`template: "jinja"`}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{"p/pack.yaml": tt.pack, "jobshop.yaml": hostBase + "packs:\n  - path: p\n"}
			for k, v := range tt.files {
				files[k] = v
			}
			writeFiles(t, dir, files)
			_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.Error(t, err)
			for _, w := range tt.want {
				assert.Contains(t, err.Error(), w)
			}
		})
	}
}

// TestGoTemplateLeftover: ${name} left in a template: go pack is a
// missed migration; the loader warns.
func TestGoTemplateLeftover(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  "template: go\npools:\n  dev: {agent: a, repo: /w, prompt: \"on ${repo}\"}\n",
		"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	require.Len(t, c.Warnings, 1)
	assert.Contains(t, c.Warnings[0], "pools.dev.prompt has ${repo}")
}
