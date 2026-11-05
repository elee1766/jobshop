package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Loader tests on small fixture packs (JOB-2). They exercise one loader
// feature each and never read packs/, so editing a real pack's text,
// caps or vars needs no change here.

// fixturePack is a pack with a pool, a workflow and triggers that name
// each other, for the prefix and instance tests.
var fixturePack = map[string]string{
	"gh/pack.yaml": `defaults: {n: 2, chat: ""}
pools:
  dev: {agent: "{{.agent}}", repo: "{{.checkout}}", max_sessions: "{{.n}}", prompt_file: "prompts/{{.tracker}}/dev.md"}
  plan: {agent: "{{.agent}}", repo: "{{.checkout}}"}
workflows:
  pr-work:
    vars: [issue]
    issue: issue
    repo: "{{.repo}}"
    tracker: "{{.tracker}}"
    linear_workspace: '{{if eq .tracker "linear"}}{{.ws}}{{end}}'
    escalate_pool: plan
    escalate:
      chat_pool: "{{.chat}}"
      prompt_file: prompts/escalation.md
    steps:
      - {id: implement, pool: dev, title: "Implement {{.issue}}", prompt_file: "prompts/{{.tracker}}/implement.md"}
      - {id: merge, pool: plan, title: "Merge {{.issue}}", prompt: "merge in {{.repo}}"}
  ci-repair:
    vars: [sha]
    steps:
      - {id: fix, pool: dev, title: "Fix {{.sha}}", prompt: x}
triggers:
  replan:
    when: {pool: dev, ready_below: 1}
    ticket: {title: plan, pool: plan}
  red:
    when: {check: "gh run list --repo {{.repo}}"}
    run: {workflow: ci-repair, key: "{{.check_output}}", vars: {sha: "{{.check_output}}"}}
`,
	"gh/prompts/github/dev.md":       "dev with gh issue\n",
	"gh/prompts/github/implement.md": "gh issue view {{.issue}}\n",
	"gh/prompts/linear/dev.md":       "dev with jobshop linear\n",
	"gh/prompts/linear/implement.md": "jobshop linear issue {{.issue}}\n",
	"gh/prompts/escalation.md":       "run {{.run}} escalated: {{.body}}\n",
}

func loadFixture(t *testing.T, host string) *Config {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{"jobshop.yaml": hostBase + host}
	for k, v := range fixturePack {
		files[k] = v
	}
	writeFiles(t, dir, files)
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	return c
}

const ghVars = "{repo: acme/widget, checkout: /src/widget, agent: a, tracker: github}"

// TestPackPrefix: a pack included twice with a prefix gets its own pool,
// workflow and trigger names, and every reference to them is renamed;
// host names (the agent) are not (#184).
func TestPackPrefix(t *testing.T) {
	c := loadFixture(t, "packs:\n  - path: gh\n    vars: "+ghVars+"\n  - path: gh\n    prefix: gz-\n    vars: {repo: acme/gizmo, checkout: /src/gizmo, agent: a, tracker: github}\n")
	assert.ElementsMatch(t, []string{"dev", "plan", "gz-dev", "gz-plan"}, keys(c.Pools))
	assert.Equal(t, "/src/widget", c.Pools["dev"].Repo)
	assert.Equal(t, "/src/gizmo", c.Pools["gz-dev"].Repo)
	assert.Equal(t, "a", c.Pools["gz-dev"].Agent, "host agent is not renamed")
	w := c.Workflows["gz-pr-work"]
	assert.Equal(t, "acme/gizmo", w.Repo)
	assert.Equal(t, "gz-plan", w.EscalatePool)
	assert.Equal(t, []string{"gz-dev", "gz-plan"}, []string{w.Steps[0].Pool, w.Steps[1].Pool})
	assert.Equal(t, "gz-", w.PackVars[PackPrefixVar])
	assert.Equal(t, "gz-dev", c.Triggers["gz-replan"].When.Pool)
	assert.Equal(t, "gz-plan", c.Triggers["gz-replan"].Ticket.Pool)
	red := c.Triggers["gz-red"]
	assert.Equal(t, "gz-ci-repair", red.Run.Workflow)
	assert.Contains(t, red.When.Check, "--repo acme/gizmo")
	assert.Equal(t, "{{.check_output}}", red.Run.Vars["sha"], "fire-time vars survive pack substitution")

	require.Len(t, c.Instances, 2)
	assert.Empty(t, c.Instances[0].Prefix)
	assert.Equal(t, "gz-", c.Instances[1].Prefix)
	assert.Equal(t, "acme/gizmo", c.Instances[1].Vars["repo"])
	assert.ElementsMatch(t, []string{"gz-dev", "gz-plan"}, c.Instances[1].Pools)
	assert.ElementsMatch(t, []string{"gz-pr-work", "gz-ci-repair"}, c.Instances[1].Workflows)
	assert.ElementsMatch(t, []string{"gz-replan", "gz-red"}, c.Instances[1].Triggers)

	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"gh/pack.yaml": fixturePack["gh/pack.yaml"], "jobshop.yaml": hostBase + "packs:\n  - path: gh\n    prefix: Bad\n"})
	_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "prefix")
}

// TestPackTrackerSelection: a templated prompt_file picks the tracker's
// prompts, and the workflow carries the tracker and workspace (#232).
func TestPackTrackerSelection(t *testing.T) {
	c := loadFixture(t, "packs:\n  - path: gh\n    vars: "+ghVars+"\n")
	w := c.Workflows["pr-work"]
	assert.Equal(t, TrackerGitHub, w.Tracker)
	assert.Empty(t, w.LinearWorkspace)
	assert.Equal(t, "dev with gh issue\n", c.Pools["dev"].Prompt)
	assert.Equal(t, "gh issue view {{.issue}}\n", w.Steps[0].Prompt)

	c = loadFixture(t, "packs:\n  - path: gh\n    vars: {repo: acme/widget, checkout: /src/widget, agent: a, tracker: linear, ws: acme}\n")
	w = c.Workflows["pr-work"]
	assert.Equal(t, TrackerLinear, w.Tracker)
	assert.Equal(t, "acme", w.LinearWorkspace)
	assert.Equal(t, "dev with jobshop linear\n", c.Pools["dev"].Prompt)
	out, err := w.Render(w.Steps[0].Prompt, map[string]any{"issue": "ENG-7"})
	require.NoError(t, err)
	assert.Equal(t, "jobshop linear issue ENG-7\n", out)
}

// TestPackTypedVarsAndRender: a "{{.n}}" pool field becomes an int, and
// run-time step templates render with the pack vars (#383).
func TestPackTypedVarsAndRender(t *testing.T) {
	c := loadFixture(t, "packs:\n  - path: gh\n    vars: {repo: acme/widget, checkout: /src/widget, agent: a, tracker: github, n: 4}\n")
	assert.Equal(t, 4, c.Pools["dev"].MaxSessions)
	assert.Equal(t, 0, c.Pools["plan"].MaxSessions)
	w := c.Workflows["pr-work"]
	vars := map[string]any{"issue": 7}
	for k, v := range w.PackVars {
		vars[k] = v
	}
	title, err := w.Render(w.Steps[0].Title, vars)
	require.NoError(t, err)
	assert.Equal(t, "Implement 7", title)
	body, err := w.Render(w.Steps[1].Prompt, vars)
	require.NoError(t, err)
	assert.Equal(t, "merge in acme/widget", body)
}

// TestPackEscalateBlock: a workflow's escalate block reads its prompt
// file from the pack and keeps the escalation vars for the controller;
// an empty chat_pool var means no chat (#379).
func TestPackEscalateBlock(t *testing.T) {
	c := loadFixture(t, "pools:\n  esc: {agent: a, type: chat, dir: /w}\npacks:\n  - path: gh\n    vars: {repo: acme/widget, checkout: /src/widget, agent: a, tracker: github, chat: esc}\n")
	e := c.Workflows["pr-work"].Escalate
	require.NotNil(t, e)
	assert.Equal(t, "esc", e.ChatPool)
	out, err := Fill(e.Prompt, map[string]any{"run": "js-r", "body": "B"})
	require.NoError(t, err)
	assert.Equal(t, "run js-r escalated: B\n", out)

	c = loadFixture(t, "packs:\n  - path: gh\n    vars: "+ghVars+"\n")
	require.NotNil(t, c.Workflows["pr-work"].Escalate)
	assert.Empty(t, c.Workflows["pr-work"].Escalate.ChatPool)
}

// TestPackHostStringsKept: host config strings are not pack templates,
// so ${SESSION_DIR} and {{...}} in the host file stay as written.
func TestPackHostStringsKept(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{"jobshop.yaml": "projects:\n  js: {}\nagents:\n  a: {command: x, args: [\"${SESSION_DIR}/a.sock\"]}\npacks:\n  - path: gh\n    vars: " + ghVars + "\n"}
	for k, v := range fixturePack {
		files[k] = v
	}
	writeFiles(t, dir, files)
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, []string{"${SESSION_DIR}/a.sock"}, c.Agents["a"].Args)
}

// TestPacksLoad is the only test that reads packs/: each in-repo pack
// loads with minimal vars. It asserts nothing about what the packs say.
func TestPacksLoad(t *testing.T) {
	dirs, err := filepath.Glob("../../packs/*/pack.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, dirs)
	vars := map[string]string{
		"selfdev":  "{repo: /src/jobshop, agent: a}",
		"github":   "{repo: acme/widget, checkout: /src/widget, agent: a, human: wile}",
		"operator": "{jobshop_repo: acme/jobshop, checkout: /src/jobshop, agent: a}",
		"deploy":   "{state_dir: /tmp/state}",
		"agent":    "{repo: acme/app, checkout: /src/app, agent: a}",
	}
	for _, f := range dirs {
		pack, err := filepath.Abs(filepath.Dir(f))
		require.NoError(t, err)
		name := filepath.Base(pack)
		t.Run(name, func(t *testing.T) {
			v, ok := vars[name]
			require.True(t, ok, "add minimal vars for new pack %s to TestPacksLoad", name)
			dir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(dir, "jobshop.yaml"),
				[]byte(hostBase+"packs:\n  - path: "+pack+"\n    vars: "+v+"\n"), 0o600))
			_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.NoError(t, err)
		})
	}
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
