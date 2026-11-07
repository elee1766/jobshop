package config

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// designPack is the three example workflows from docs/plan.md section 6,
// "Runs and workflows", with the pools they name.
const designPack = `pools:
  dev: {agent: a, dir: /w}
  plan: {agent: a, dir: /w}
  refinery: {agent: a, dir: /w}
workflows:
  pr-work:
    vars: [issue]
    issue: issue
    repo: acme/widget
    escalate_pool: plan
    steps:
      - {id: implement, pool: dev, retries: 3, prompt_file: prompts/implement.md,
         title: "Implement #{{.issue}}"}
      - {id: review-merge, pool: refinery, retries: 2, prompt_file: prompts/merge.md,
         title: "Review and merge PR for #{{.issue}}"}
      - {id: post-merge, pool: refinery, retries: 2, prompt_file: prompts/post-merge.md,
         title: "Check master CI and self-deploy for #{{.issue}}"}
  ci-repair:
    vars: [sha]
    escalate_pool: plan
    steps:
      - {id: fix, pool: dev, retries: 2, prompt_file: prompts/ci-fix.md,
         title: "Fix red master CI at {{.sha}}"}
      - {id: merge, pool: refinery, retries: 2, prompt_file: prompts/merge.md,
         title: "Merge CI fix for {{.sha}}"}
  blocker-check:
    vars: [ticket, issue]
    issue: issue
    repo: acme/widget
    escalate_pool: plan
    steps:
      - {id: assess, pool: plan, retries: 1, prompt_file: prompts/blocker.md,
         title: "Re-check blocker on {{.ticket}}"}
`

func designFiles() map[string]string {
	return map[string]string{
		"p/pack.yaml":             designPack,
		"p/prompts/implement.md":  "Implement issue {{.issue}} in {{.repo}}.\n",
		"p/prompts/merge.md":      "Merge it. Cost: $5.\n",
		"p/prompts/post-merge.md": "Check master.\n",
		"p/prompts/ci-fix.md":     "Fix {{.sha}}.\n",
		"p/prompts/blocker.md":    "Re-check {{.ticket}} (#{{.issue}}).\n",
		"jobshop.yaml":            hostBase + "packs:\n  - path: p\n    vars: {repo: /src}\n",
	}
}

func TestWorkflowsLoad(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, designFiles())
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	require.Len(t, c.Workflows, 3)

	pr := c.Workflows["pr-work"]
	assert.Equal(t, []string{"issue"}, pr.Vars)
	assert.Equal(t, "issue", pr.Issue)
	assert.Equal(t, "plan", pr.EscalatePool)
	require.Len(t, pr.Steps, 3)
	assert.Equal(t, []string{"implement", "review-merge", "post-merge"},
		[]string{pr.Steps[0].ID, pr.Steps[1].ID, pr.Steps[2].ID})
	assert.Equal(t, "Implement #{{.issue}}", pr.Steps[0].Title, "run vars stay templates")
	assert.Equal(t, "Implement issue {{.issue}} in {{.repo}}.\n", pr.Steps[0].Prompt,
		"prompt_file read, templates kept for run start")
	assert.Equal(t, "Merge it. Cost: $5.\n", pr.Steps[1].Prompt)
	assert.Empty(t, pr.Steps[0].PromptFile)
	assert.Equal(t, 3, pr.Steps[0].RetryLimit())
	assert.Equal(t, map[string]any{"repo": "/src", PackPrefixVar: ""}, pr.PackVars)
	assert.Equal(t, 1, c.Workflows["blocker-check"].Steps[0].RetryLimit())
	assert.Equal(t, []string{filepath.Join(dir, "p", PackFileName)}, c.Sources["workflows.pr-work"])
}

func TestWorkflowDefaults(t *testing.T) {
	c, err := Parse([]byte(hostBase + "pools:\n  dev: {agent: a, dir: /w}\nworkflows:\n  w:\n    steps:\n      - {id: s, pool: dev, title: t, prompt: p}\n"))
	require.NoError(t, err)
	assert.Equal(t, DefaultStepRetries, c.Workflows["w"].Steps[0].RetryLimit())
}

func TestWorkflowErrors(t *testing.T) {
	const pools = "pools:\n  dev: {agent: a, dir: /w}\n"
	step := "{id: s, pool: dev, title: t, prompt: p}"
	tests := []struct {
		name string
		pack string // workflows section body of p/pack.yaml
		want string
	}{
		{"unknown step pool", "  w:\n    steps:\n      - {id: s, pool: nope, title: t, prompt: p}\n", `step s: unknown pool "nope"`},
		{"unknown escalate_pool", "  w:\n    escalate_pool: nope\n    steps:\n      - " + step + "\n", `escalate_pool: unknown pool "nope"`},
		{"duplicate step id", "  w:\n    steps:\n      - " + step + "\n      - " + step + "\n", "step s: duplicate step id"},
		{"empty steps", "  w:\n    vars: [x]\n", "at least one step"},
		{"prompt and prompt_file", "  w:\n    steps:\n      - {id: s, pool: dev, title: t, prompt: p, prompt_file: x.md}\n", "steps[0]: set prompt or prompt_file, not both"},
		{"neither prompt nor prompt_file", "  w:\n    steps:\n      - {id: s, pool: dev, title: t}\n", "step s: set prompt or prompt_file"},
		{"unknown title var", "  w:\n    steps:\n      - {id: s, pool: dev, title: \"t {{.who}}\", prompt: p}\n", "title uses {{.who}}, not in workflow vars or pack vars"},
		{"unknown prompt var", "  w:\n    vars: [a]\n    steps:\n      - {id: s, pool: dev, title: t, prompt: \"{{.a}} {{.b}}\"}\n", "prompt uses {{.b}}"},
		{"issue not in vars", "  w:\n    vars: [sha]\n    issue: issue\n    steps:\n      - " + step + "\n", `issue "issue" is not in vars`},
		{"issue without repo", "  w:\n    vars: [issue]\n    issue: issue\n    steps:\n      - " + step + "\n", "repo (owner/name) is required"},
		{"bad repo", "  w:\n    vars: [issue]\n    issue: issue\n    repo: /src\n    steps:\n      - " + step + "\n", `repo "/src" must be owner/name`},
		{"bad tracker", "  w:\n    vars: [issue]\n    issue: issue\n    repo: a/b\n    tracker: jira\n    steps:\n      - " + step + "\n", `tracker "jira" must be`},
		{"linear without issue", "  w:\n    vars: [x]\n    repo: a/b\n    tracker: linear\n    linear_workspace: acme\n    steps:\n      - " + step + "\n", "tracker linear needs issue"},
		{"workspace without linear", "  w:\n    vars: [issue]\n    issue: issue\n    repo: a/b\n    linear_workspace: acme\n    steps:\n      - " + step + "\n", "linear_workspace needs tracker: linear"},
		{"issue and pr", "  w:\n    vars: [issue, pr]\n    issue: issue\n    pr: pr\n    repo: a/b\n    steps:\n      - " + step + "\n", "set issue or pr, not both"},
		{"pr not in vars", "  w:\n    vars: [x]\n    pr: pr\n    repo: a/b\n    steps:\n      - " + step + "\n", `pr "pr" is not in vars`},
		{"missing step id", "  w:\n    steps:\n      - {pool: dev, title: t, prompt: p}\n", `steps[0]: id ""`},
		{"missing title", "  w:\n    steps:\n      - {id: s, pool: dev, prompt: p}\n", "step s: title is required"},
		{"negative retries", "  w:\n    steps:\n      - {id: s, pool: dev, title: t, prompt: p, retries: -1}\n", "retries must be >= 0"},
		{"bad step type", "  w:\n    steps:\n      - {id: s, type: shell, pool: dev, title: t, prompt: p}\n", `type "shell" must be agent, command or http`},
		{"command without run", "  w:\n    steps:\n      - {id: s, type: command}\n", "step s: run is required"},
		{"command with url", "  w:\n    steps:\n      - {id: s, type: command, run: x, url: http://x}\n", "are for http steps"},
		{"command with prompt", "  w:\n    steps:\n      - {id: s, type: command, run: x, prompt: p}\n", "prompt is for agent steps"},
		{"http without url", "  w:\n    steps:\n      - {id: s, type: http}\n", "url is required"},
		{"http bad method", "  w:\n    steps:\n      - {id: s, type: http, url: http://x, method: TRACE}\n", `method "TRACE"`},
		{"bad timeout", "  w:\n    steps:\n      - {id: s, type: command, run: x, timeout: soon}\n", `timeout "soon"`},
		{"bad on_restart", "  w:\n    steps:\n      - {id: s, type: command, run: x, on_restart: maybe}\n", `on_restart "maybe"`},
		{"unknown run var", "  w:\n    steps:\n      - {id: a, type: command, run: 'echo {{.nope}}'}\n", "{{.nope}} is not in workflow vars"},
		{"unknown key", "  w:\n    steps:\n      - {id: s, pool: dev, title: t, prompt: p, pol: x}\n", "pol"},
		{"pack_vars set by user", "  w:\n    pack_vars: {x: 1}\n    steps:\n      - " + step + "\n", "pack_vars is set by the loader"},
		{"needs later step", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p, needs: [b]}\n      - {id: b, pool: dev, title: t, prompt: p}\n", `step a: needs "b", which is not an earlier step`},
		{"needs itself", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p}\n      - {id: b, pool: dev, title: t, prompt: p, needs: [b]}\n", `step b: needs "b"`},
		{"when on unrelated step", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p}\n      - {id: b, pool: dev, title: t, prompt: p, needs: [a]}\n      - {id: c, pool: dev, title: t, prompt: p, needs: [a], when: {step: b, outcome: ok}}\n", `step c: when.step "b" is not a step it needs`},
		{"bad outcome", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p}\n      - {id: b, pool: dev, title: t, prompt: p, when: {step: a, outcome: \"Not ok\"}}\n", `when.outcome "Not ok" must match`},
		{"on_request_changes not a need", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p}\n      - {id: b, pool: dev, title: t, prompt: p}\n      - {id: c, pool: dev, title: t, prompt: p, on_request_changes: a}\n", `step c: on_request_changes "a" must be a step this agent step needs directly`},
		{"on_request_changes itself", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p, on_request_changes: a}\n", `step a: on_request_changes "a"`},
		{"negative max_review_rounds", "  w:\n    steps:\n      - {id: a, pool: dev, title: t, prompt: p}\n      - {id: b, pool: dev, title: t, prompt: p, max_review_rounds: -1}\n", "max_review_rounds must be >= 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, map[string]string{
				"p/pack.yaml":  pools + "workflows:\n" + tt.pack,
				"jobshop.yaml": hostBase + "packs:\n  - path: p\n    vars: {repo: /src}\n",
			})
			_, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
			assert.True(t, strings.Contains(err.Error(), filepath.Join("p", PackFileName)), "names the pack: %v", err)
		})
	}
}

// A pack var used in a step template is allowed, and the host can
// override one workflow field without restating the steps.
func TestWorkflowPackVarAndOverride(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"p/pack.yaml":  "pools:\n  dev: {agent: a, dir: /w}\n  plan: {agent: a, dir: /w}\nworkflows:\n  w:\n    steps:\n      - {id: s, pool: dev, title: \"in {{.repo}}\", prompt: p}\n",
		"jobshop.yaml": hostBase + "packs:\n  - path: p\n    vars: {repo: /src}\nworkflows:\n  w:\n    escalate_pool: plan\n",
	})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	w := c.Workflows["w"]
	assert.Equal(t, "plan", w.EscalatePool)
	assert.Equal(t, "in {{.repo}}", w.Steps[0].Title)
	assert.Equal(t, "/src", w.PackVars["repo"])
	assert.Len(t, c.Sources["workflows.w"], 2)
}

func TestTriggerRun(t *testing.T) {
	base := hostBase + "pools:\n  dev: {agent: a, dir: /w}\nworkflows:\n  fix:\n    vars: [sha, why]\n    steps:\n      - {id: s, pool: dev, title: \"Fix {{.sha}}\", prompt: p}\ntriggers:\n  t:\n"
	c, err := Parse([]byte(base + "    when: {check: 'true'}\n    run: {workflow: fix, vars: {sha: \"{{.check_output}}\", why: red}, key: \"red-{{.check_output}}\"}\n"))
	require.NoError(t, err)
	r := c.Triggers["t"].Run
	require.NotNil(t, r)
	assert.Nil(t, c.Triggers["t"].Ticket)
	assert.Equal(t, "red-{{.check_output}}", r.Key, "template kept for fire time")

	tests := []struct{ name, yaml, want string }{
		{"both", "    when: {check: 'true'}\n    ticket: {title: x}\n    run: {workflow: fix, vars: {sha: a, why: b}}\n", "exactly one of ticket, run, command and handler"},
		{"neither", "    when: {check: 'true'}\n", "exactly one of ticket, run, command and handler"},
		{"unknown workflow", "    when: {check: 'true'}\n    run: {workflow: nope}\n", `unknown workflow "nope"`},
		{"missing var", "    when: {check: 'true'}\n    run: {workflow: fix, vars: {sha: a}}\n", "needs var why"},
		{"extra var", "    when: {check: 'true'}\n    run: {workflow: fix, vars: {sha: a, why: b, x: c}}\n", `no var "x"`},
		{"output without check", "    when: {ready_below: 1}\n    run: {workflow: fix, vars: {sha: a, why: b}, key: \"{{.check_output}}\"}\n", "needs when.check"},
		{"event without on", "    when: {ready_below: 1}\n    run: {workflow: fix, vars: {sha: \"{{.event.head_sha}}\", why: b}}\n", "needs an on: list"},
		{"on without source", "    on: [{}]\n    command: {run: x}\n", "exactly one of github, linear and source"},
		{"bad fallback", "    on: [{github: {event: push}}]\n    fallback: soon\n    command: {run: x}\n", "fallback"},
		{"empty command", "    on: [{github: {event: push}}]\n    command: {run: \" \"}\n", "command: run is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(base + tt.yaml))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.want)
		})
	}
}
