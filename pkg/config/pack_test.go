package config

import (
	"os"
	"path/filepath"
	"strings"
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
    repo: "{{.repo}}"
    max_sessions: "{{.n}}"
    prompt_file: prompts/dev.md
    idle_timeout: 1m
triggers:
  t:
    when: {pool: dev, ready_below: 1}
    ticket: {title: "Plan {{.repo}}", labels: [a, b]}
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
			"p/pack.yaml":  "pools:\n  dev:\n    agent: a\n    repo: \"{{.repo}}\"\n",
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

func TestPackExtends(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"base/pack.yaml": "defaults: {n: 1}\npools:\n  dev: {agent: a, dir: /w, max_sessions: \"{{.n}}\", prompt_file: p.md}\n",
		"base/p.md":      "for {{.who}}\n",
		"inst/pack.yaml": "extends:\n  - path: ../base\n    vars: {who: \"{{.team}}\"}\npools:\n  dev: {idle_timeout: 3m}\n",
		"jobshop.yaml":   hostBase + "packs:\n  - path: inst\n    vars: {team: ops}\n",
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	p := c.Pools["dev"]
	assert.Equal(t, "for ops\n", p.Prompt, "base prompt file gets the instance's value")
	assert.Equal(t, 1, p.MaxSessions, "base default")
	assert.Equal(t, "3m", p.IdleTimeout, "instance overrides base")
	assert.Equal(t, []string{
		filepath.Join(dir, "base", "pack.yaml"),
		filepath.Join(dir, "inst", "pack.yaml"),
	}, c.Sources["pools.dev"])

	writeFiles(t, dir, map[string]string{"loop/pack.yaml": "extends:\n  - path: .\n"})
	_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: loop\n"})
	_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "cycle")
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: inst\n"})
	_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, `var "team" is not set`)
}

func TestPackRenamedVars(t *testing.T) {
	pack := "renamed_vars: {repo: jobshop_repo}\npools:\n  op:\n    agent: a\n    repo: /src\n    prompt_file: p.md\n    idle_timeout: \"{{.jobshop_repo}}\"\n"
	for _, tc := range []struct {
		name, vars, want string
		warn             bool
	}{
		{"old name", "{repo: 1m}", "1m", true},
		{"new name", "{jobshop_repo: 2m}", "2m", false},
		{"both, new wins", "{repo: 1m, jobshop_repo: 2m}", "2m", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				"op/pack.yaml": pack,
				"op/p.md":      "x\n",
				"jobshop.yaml": hostBase + "packs:\n  - path: op\n    vars: " + tc.vars + "\n",
			})
			c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.NoError(t, err)
			assert.Equal(t, tc.want, c.Pools["op"].IdleTimeout)
			if tc.warn {
				require.Len(t, c.Warnings, 1)
				assert.Contains(t, c.Warnings[0], `var "repo" is renamed to "jobshop_repo"`)
			} else {
				assert.Empty(t, c.Warnings)
			}
		})
	}
}

func TestLoadFileWithRewrite(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"old/op/pack.yaml": "pools:\n  op:\n    agent: a\n    repo: /src\n    idle_timeout: 1m\n",
		"new/op/pack.yaml": "pools:\n  op:\n    agent: a\n    repo: \"{{.jobshop_repo}}\"\n    idle_timeout: 1m\n",
		"jobshop.yaml":     hostBase + "packs:\n  - path: old/op\n",
	})
	f := filepath.Join(dir, "jobshop.yaml")
	_, err := LoadFile(f)
	require.NoError(t, err)
	_, err = LoadFileWith(f, func(p string) string { return strings.Replace(p, "old", "new", 1) })
	require.ErrorContains(t, err, `var "jobshop_repo" is not set`)
}

// TestMinimalPackDoc loads the minimal pack from docs/packs.md and the
// README (a copy in testdata/minimal; keep them in sync).
func TestMinimalPackDoc(t *testing.T) {
	c, err := LoadFile(filepath.Join("testdata", "minimal", "jobshop.yaml"))
	require.NoError(t, err)
	require.Equal(t, "/srv/widget", c.Pools["work"].Repo)
	require.Equal(t, 1, c.Pools["work"].MaxSessions)
	require.Len(t, c.Workflows["change"].Steps, 2)
}

// TestEscalationConfig: chat_pool must be a chat pool with a prompt, and
// prompts may only use the escalation vars.
func TestEscalationConfig(t *testing.T) {
	base := "projects:\n  js: {}\nagents:\n  a: {command: x}\npools:\n  dev: {agent: a, dir: /w}\n  chat: {agent: a, type: chat, dir: /w}\n"
	for name, tc := range map[string]struct{ yaml, err string }{
		"ok":        {"escalation: {chat_pool: chat, prompt: \"t {{.ticket}} {{.body}}\", pool_prompt: \"p {{.pool}} {{.error}}\"}\n", ""},
		"not chat":  {"escalation: {chat_pool: dev, prompt: x}\n", "not a chat pool"},
		"unknown":   {"escalation: {chat_pool: nope, prompt: x}\n", "unknown pool"},
		"no prompt": {"escalation: {chat_pool: chat}\n", "needs prompt"},
		"bad var":   {"escalation: {chat_pool: chat, prompt: \"{{.issue}}\"}\n", "uses {{.issue}}"},
		"hold var":  {"escalation: {chat_pool: chat, hold_prompt: \"{{.issue}}\"}\n", "uses {{.issue}}"},
		"empty":     {"escalation: {}\n", ""},
	} {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{"jobshop.yaml": base + tc.yaml})
		_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		if tc.err == "" {
			assert.NoError(t, err, name)
		} else {
			require.Error(t, err, name)
			assert.Contains(t, err.Error(), tc.err, name)
		}
	}
}

// TestLearningKindsConfig: learnings.kinds (#524) defaults to the three
// built-in kinds, packs append to the host's list, and a pattern that can
// match no kind is a load error.
func TestLearningKindsConfig(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"correction", "failure", "workaround"}, c.LearningKinds(), "default unchanged")

	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  "learnings:\n  kinds: [blocker]\n",
		"jobshop.yaml": hostBase + "packs:\n  - path: p\nlearnings:\n  kinds: [friction, \"review-*\"]\n",
	})
	c, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"blocker", "friction", "review-*"}, c.LearningKinds())

	for _, bad := range []string{"Friction", "a/b", "["} {
		writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "learnings:\n  kinds: [\"" + bad + "\"]\n"})
		_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.ErrorContains(t, err, "learnings.kinds", bad)
	}
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + `workflows:
  w:
    steps: [{id: s, pool: dev, title: t, prompt: p}]
pools:
  dev: {agent: a, dir: /tmp}
triggers:
  t:
    when: {learnings: {min: 1, source: human}}
    run: {workflow: w}
`})
	_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "learnings.source")
}
