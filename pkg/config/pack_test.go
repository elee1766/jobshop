package config

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
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
	assert.Equal(t, 3, c.Pools["deploy-check"].MaxSessions, "post-merge has its own pool (#167), several sessions (#174)")
	assert.Contains(t, c.Pools["refinery"].Prompt, "--delete-branch")
	assert.NotContains(t, c.Pools["dev"].Prompt, "--pool refinery", "no prompt-only handoff")
	assert.Contains(t, c.Pools["dev"].Prompt, "You work on elee1766/jobshop")
	assert.Contains(t, c.Pools["dev"].Prompt, "Read AGENTS.md")
	assert.Equal(t, "/src/jobshop", c.Pools["dev"].Repo, "selfdev repo var is the checkout")
	assert.Contains(t, c.Pools["plan"].Prompt, "@elee1766")
	assert.Contains(t, c.Triggers["master-red"].When.Check, "--repo elee1766/jobshop --branch master --workflow ci")
	assert.Contains(t, c.Workflows["pr-work"].PackVars["deploy_check"], "jobshop version")
	assert.Equal(t, "go build ./... && go vet ./... && go test ./...", c.Workflows["pr-work"].PackVars["check_cmd"])
	assert.Contains(t, c.Workflows["pr-work"].PackVars["docs_rule"], "docs/notes/<issue>-<slug>.md")
	assert.Contains(t, c.Pools["plan"].Prompt, "target")
	assert.Contains(t, c.Pools["plan"].Prompt, "jobshop run start pr-work --var issue=<n> --key issue-<n>")
	op := c.Pools["operator"]
	assert.Equal(t, 0, op.MaxSessions, "operator sessions are started by hand or by key")
	assert.Equal(t, "/src/jobshop", op.Repo)
	assert.Contains(t, op.Prompt, "agents working elee1766/jobshop")
	assert.Contains(t, op.Prompt, "Read AGENTS.md")
	w := c.Workflows["pr-work"]
	require.Len(t, w.Steps, 3)
	assert.Equal(t, "plan", w.EscalatePool)
	assert.Contains(t, w.Steps[1].Prompt, "--match-head-commit")
	assert.Equal(t, "squash", w.PackVars["merge_method"], "selfdev merges by squash")
	assert.Empty(t, w.PackVars["required_approver"], "selfdev needs no approval")
	assert.Contains(t, w.Steps[2].Prompt, "${deploy_check}", "filled at run start")
	assert.Equal(t, "deploy-check", w.Steps[2].Pool, "post-merge leaves the refinery free (#167)")
	b := c.Workflows["stall-check"]
	require.Len(t, b.Steps, 1)
	assert.Equal(t, 1, b.Steps[0].RetryLimit())
	assert.Equal(t, "plan", b.Steps[0].Pool)
	assert.Equal(t, "plan", b.EscalatePool)
	assert.Equal(t, []string{"signal", "subject", "detail"}, b.Vars)
	assert.Contains(t, b.Steps[0].Prompt, "jobshop ticket list --orphaned")
	assert.Contains(t, b.Steps[0].Prompt, "Always escalate these")
	assert.NotContains(t, c.Workflows, "blocker-check", "folded into stall-check")
	assert.Contains(t, c.Pools["plan"].Prompt, "jobshop label add <id> blocked")
	assert.Contains(t, c.Pools["plan"].Prompt, "--key stall:blocked:<n>")
	assert.Contains(t, c.Pools["plan"].Admin, "session-restart")
	st := c.Triggers["stall"]
	require.NotNil(t, st.Run)
	assert.Equal(t, "stall-check", st.Run.Workflow)
	assert.Equal(t, "${stall_subject}", st.Run.Vars["subject"], "stall vars survive pack substitution")
	assert.Equal(t, 2, st.StallMaxRuns())
	sigs := map[string]StallSignal{}
	for _, s := range st.When.Stall {
		sigs[s.Signal] = s
	}
	assert.Equal(t, "3h", sigs[StallStepStalled].After)
	assert.Contains(t, sigs[StallIssueNoRun].Check, "gh issue list --repo elee1766/jobshop")
	assert.Contains(t, sigs["no_merge"].Check, "repos/elee1766/jobshop/commits/master")
	assert.Contains(t, sigs["pr_unmerged"].Check, "statusCheckRollup")
	mr := c.Triggers["master-red"]
	assert.Nil(t, mr.Ticket)
	require.NotNil(t, mr.Run, "master-red starts a run")
	assert.Equal(t, "ci-repair", mr.Run.Workflow)
	assert.Equal(t, "${check_output}", mr.Run.Key)
	assert.Equal(t, map[string]string{"sha": "${check_output}"}, mr.Run.Vars)
	ci := c.Workflows["ci-repair"]
	require.Len(t, ci.Steps, 2)
	assert.Equal(t, "plan", ci.EscalatePool)
	assert.Equal(t, 2, ci.Steps[0].RetryLimit())
	assert.Equal(t, []string{"dev", "refinery"}, []string{ci.Steps[0].Pool, ci.Steps[1].Pool})
	assert.Contains(t, ci.Steps[0].Prompt, "--log-failed")
	assert.Contains(t, ci.Steps[1].Prompt, "--match-head-commit")
	assert.Equal(t, "plan", c.Triggers["replan"].Ticket.Pool)
	assert.Equal(t, "${SESSION_DIR}/jcode.sock", c.Agents["jcode"].Args[2], "host strings are not substituted")
}

// TestOperatorPackLoads loads packs/operator for another repo: the prompt
// covers what the pool may do (#124) and names no Jobshop values.
func TestOperatorPackLoads(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/operator")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + pack + `
    vars: {repo: acme/widget, checkout: /src/widget, agent: a}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	require.Len(t, c.Pools, 1)
	p := c.Pools["operator"]
	assert.Equal(t, "/src/widget", p.Repo)
	assert.Equal(t, 0, p.MaxSessions)
	assert.Equal(t, []string{"config-reload", "session-list"}, p.Admin)
	assert.NotContains(t, p.Prompt, "needs the admin socket")
	for _, s := range []string{"acme/widget", "jobshop ticket create", "gh issue create", "jobshop send", "jobshop run start", "jobshop reload", "ticket trace", "session list", "gh pr create", "Replies are short"} {
		assert.Contains(t, p.Prompt, s)
	}
	assert.NotContains(t, p.Prompt, "elee1766")
	assert.NotContains(t, p.Prompt, "Jobshop")
}

// TestGithubPackLoads loads packs/github for another repo: no Jobshop
// values leak in, and the vars reach pools, prompts and the trigger.
func TestGithubPackLoads(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + pack + `
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test, ci_workflow: build, ci_branch: drill/red-ci}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dev", "plan", "refinery", "deploy-check"}, slices.Collect(maps.Keys(c.Pools)))
	assert.ElementsMatch(t, []string{"pr-work", "ci-repair", "stall-check", "drill"}, slices.Collect(maps.Keys(c.Workflows)))
	assert.ElementsMatch(t, []string{"replan", "master-red", "stall"}, slices.Collect(maps.Keys(c.Triggers)))
	assert.Equal(t, 2, c.Pools["dev"].MaxSessions, "dev_sessions default")
	assert.Equal(t, 3, c.Pools["deploy-check"].MaxSessions, "deploy_sessions default: one sleeping post-merge must not block the rest (#174)")
	for n, p := range c.Pools {
		assert.Equal(t, "/src/widget", p.Repo, n)
		assert.NotContains(t, p.Prompt, "elee1766", n)
		assert.NotContains(t, p.Prompt, "Jobshop", n)
	}
	assert.Contains(t, c.Pools["dev"].Prompt, "GitHub issues on acme/widget")
	assert.Contains(t, c.Pools["plan"].Prompt, "@wile")
	mr := c.Triggers["master-red"]
	assert.Contains(t, mr.When.Check, "--repo acme/widget --branch drill/red-ci --workflow build")
	require.NotNil(t, mr.Run)
	assert.Equal(t, "ci-repair", mr.Run.Workflow)
	w := c.Workflows["pr-work"]
	assert.Equal(t, "make test", w.PackVars["check_cmd"])
	assert.Contains(t, w.PackVars["docs_rule"], "Never append a numbered section", "docs_rule default (#131)")
	body, err := FillTemplate(w.Steps[0].Prompt, map[string]any{"issue": 7, "check_cmd": "make test", "docs_rule": w.PackVars["docs_rule"]})
	require.NoError(t, err)
	assert.Contains(t, body, "`make test`")
	assert.Contains(t, body, "Docs: put a design note")
	assert.Contains(t, w.Steps[1].Prompt, "Conflicts only in docs", "refinery resolves docs-only conflicts (#131)")
	assert.Contains(t, c.Workflows["ci-repair"].Steps[1].Prompt, "Conflicts only in docs")
}

// TestGithubPackGeneric: nothing in packs/github names the Jobshop repo.
func TestGithubPackGeneric(t *testing.T) {
	err := filepath.WalkDir("../../packs/github", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.NotContains(t, string(b), "elee1766/jobshop", path)
		return nil
	})
	require.NoError(t, err)
}

func TestPackExtends(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"base/pack.yaml": "defaults: {n: 1}\npools:\n  dev: {agent: a, dir: /w, max_sessions: \"${n}\", prompt_file: p.md}\n",
		"base/p.md":      "for ${who}\n",
		"inst/pack.yaml": "extends:\n  - path: ../base\n    vars: {who: \"${team}\"}\npools:\n  dev: {idle_timeout: 3m}\n",
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

// TestGithubMergePolicy: merge policy vars default to squash with no
// approver and fill the merge prompts at run start.
func TestGithubMergePolicy(t *testing.T) {
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	load := func(extra string) *Config {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{"jobshop.yaml": "projects:\n  js: {}\nagents:\n  a: {command: x}\npacks:\n  - path: " + pack +
			"\n    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test" + extra + "}\n"})
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		return c
	}
	render := func(c *Config, wf, step string) string {
		w := c.Workflows[wf]
		vars := map[string]any{"issue": 7, "sha": "abc"}
		maps.Copy(vars, w.PackVars)
		for _, s := range w.Steps {
			if s.ID == step {
				out, err := FillTemplate(s.Prompt, vars)
				require.NoError(t, err)
				return out
			}
		}
		t.Fatalf("no step %s/%s", wf, step)
		return ""
	}

	def := load("")
	pv := def.Workflows["pr-work"].PackVars
	assert.Equal(t, "squash", pv["merge_method"])
	assert.Empty(t, pv["required_approver"])
	assert.Empty(t, pv["required_checks"])
	m := render(def, "pr-work", "review-merge")
	assert.Contains(t, m, "gh pr merge <n> --squash --match-head-commit <headRefOid>")
	assert.Contains(t, m, "Required approver: ``. Empty means no approval")
	assert.Contains(t, render(def, "ci-repair", "merge"), "--squash --match-head-commit")

	set := load(", merge_method: merge, required_approver: abs3nt-agent, required_checks: \"build, test\"")
	m = render(set, "pr-work", "review-merge")
	assert.Contains(t, m, "gh pr merge <n> --merge --match-head-commit <headRefOid>")
	assert.Contains(t, m, `gh api --paginate repos/acme/widget/pulls/<n>/reviews --jq '[.[] | select(.user.login == "abs3nt-agent" and .state == "APPROVED" and .commit_id == "<headRefOid>")] | length'`)
	assert.Contains(t, m, "Required checks: `build, test`")
	assert.Contains(t, render(set, "ci-repair", "merge"), `.user.login == "abs3nt-agent"`)
	assert.NotContains(t, m, "${")
}

// TestSelfdevPoolCaps: selfdev extends packs/github without passing
// deploy_sessions, so the base default applies (#177). The gasserver bug
// was a missed reload, not this.
func TestSelfdevPoolCaps(t *testing.T) {
	pack, err := filepath.Abs("../../packs/selfdev")
	require.NoError(t, err)
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: " + pack +
		"\n    vars: {repo: /src/jobshop, agent: a, dev_sessions: 4}\n"})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, 4, c.Pools["dev"].MaxSessions)
	assert.Equal(t, 3, c.Pools["deploy-check"].MaxSessions, "github pack default")
}
