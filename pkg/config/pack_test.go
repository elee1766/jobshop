package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o700))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
}

const hostBase = "projects:\n  js: {}\nagents:\n  a:\n    command: x\n"

func TestPackMerge(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p1/pack.yaml": `pools:
  dev:
    agent: a
    repo: ${repo}
    max_sessions: ${n}
    prompt_file: prompts/dev.md
    idle_timeout: 1m
triggers:
  t:
    when: {pool: dev, ready_below: 1}
    ticket: {title: "Plan ${repo}", labels: [a, b]}
`,
		"p1/prompts/dev.md": "pack prompt\n",
		"p2/pack.yaml":      "pools:\n  dev:\n    idle_timeout: 2m\n",
		"jobshop.yaml": hostBase + `packs:
  - path: p1
    vars: {repo: /src, n: 2}
  - path: p2
pools:
  dev:
    max_sessions: 5
triggers:
  t:
    ticket: {labels: [c]}
`,
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	p := c.Pools["dev"]
	assert.Equal(t, "/src", p.Repo)
	assert.Equal(t, 5, p.MaxSessions, "host overrides one field")
	assert.Equal(t, "2m", p.IdleTimeout, "later pack overrides earlier")
	assert.Equal(t, "pack prompt\n", p.Prompt, "prompt_file read relative to pack")
	assert.Equal(t, "js", p.Project)
	tr := c.Triggers["t"]
	assert.Equal(t, "Plan /src", tr.Ticket.Title, "unchanged fields survive")
	assert.Equal(t, []string{"c"}, tr.Ticket.Labels, "lists replace")
	assert.Equal(t, []string{
		filepath.Join(dir, "p1", "pack.yaml"),
		filepath.Join(dir, "p2", "pack.yaml"),
		filepath.Join(dir, "jobshop.yaml"),
	}, c.Sources["pools.dev"])
}

func TestPromptFileInHost(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"jobshop.yaml": hostBase + "pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt_file: p.md\n",
		"p.md":         "hello",
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "hello", c.Pools["dev"].Prompt)
	assert.Empty(t, c.Pools["dev"].PromptFile)
}

func TestPackErrors(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string
	}{
		{"missing var", map[string]string{
			"p/pack.yaml":  "pools:\n  dev:\n    agent: a\n    repo: ${repo}\n",
			"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
		}, []string{"p/pack.yaml", "pools.dev.repo", `var "repo" is not set`}},
		{"prompt and prompt_file", map[string]string{
			"p/pack.yaml":  "pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt: x\n    prompt_file: y.md\n",
			"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
		}, []string{"p/pack.yaml", "pools.dev", "not both"}},
		{"prompt and prompt_file in host", map[string]string{
			"jobshop.yaml": hostBase + "pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt: x\n    prompt_file: y.md\n",
		}, []string{"jobshop.yaml", "not both"}},
		{"missing prompt file", map[string]string{
			"p/pack.yaml":  "pools:\n  dev:\n    agent: a\n    dir: /w\n    prompt_file: nope.md\n",
			"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
		}, []string{"p/pack.yaml", "pools.dev.prompt_file"}},
		{"unknown pack path", map[string]string{
			"jobshop.yaml": hostBase + "packs:\n  - path: nope\n",
		}, []string{"pack ", "nope"}},
		{"unknown pack key", map[string]string{
			"p/pack.yaml":  "projects:\n  x: {}\n",
			"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
		}, []string{"p/pack.yaml", "projects"}},
		{"validation names pack", map[string]string{
			"p/pack.yaml":  "pools:\n  dev:\n    agent: zz\n    dir: /w\n",
			"jobshop.yaml": hostBase + "packs:\n  - path: p\n",
		}, []string{"p/pack.yaml", "pool dev", `unknown agent "zz"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, tt.files)
			_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.Error(t, err)
			for _, w := range tt.want {
				assert.Contains(t, err.Error(), w)
			}
		})
	}
}

func TestSelfdevPackLoads(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/selfdev")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  jcode:
    command: jcode
    args: [acp, --socket, "${SESSION_DIR}/jcode.sock"]
packs:
  - path: ` + pack + `
    vars: {repo: /src/jobshop, agent: jcode, dev_sessions: 2}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 2, c.Pools["dev"].MaxSessions)
	assert.Equal(t, 1, c.Pools["plan"].MaxSessions)
	assert.Equal(t, 1, c.Pools["refinery"].MaxSessions)
	assert.Contains(t, c.Pools["refinery"].Prompt, "--delete-branch")
	assert.NotContains(t, c.Pools["dev"].Prompt, "--pool refinery", "no prompt-only handoff")
	assert.Contains(t, c.Pools["dev"].Prompt, "You work on Jobshop")
	assert.Contains(t, c.Pools["plan"].Prompt, "target")
	assert.Contains(t, c.Pools["plan"].Prompt, "jobshop run start pr-work --var issue=<n> --key issue-<n>")
	w := c.Workflows["pr-work"]
	require.Len(t, w.Steps, 3)
	assert.Equal(t, "plan", w.EscalatePool)
	assert.Contains(t, w.Steps[1].Prompt, "--match-head-commit")
	assert.Contains(t, w.Steps[2].Prompt, "jobshop version")
	b := c.Workflows["blocker-check"]
	require.Len(t, b.Steps, 1)
	assert.Equal(t, 1, b.Steps[0].RetryLimit())
	assert.Equal(t, "plan", b.Steps[0].Pool)
	assert.Equal(t, "plan", b.EscalatePool)
	assert.Equal(t, []string{"ticket", "issue"}, b.Vars)
	assert.Contains(t, b.Steps[0].Prompt, "jobshop ticket list --orphaned")
	assert.Contains(t, c.Pools["plan"].Prompt, "jobshop run start blocker-check --var ticket=<id> --var issue=<n> --key blocker-<id>")
	assert.Equal(t, "plan", c.Triggers["replan"].Ticket.Pool)
	assert.Equal(t, "${SESSION_DIR}/jcode.sock", c.Agents["jcode"].Args[2], "host strings are not substituted")
}
