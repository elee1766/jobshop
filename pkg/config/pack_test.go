package config

import (
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

// TestSelfdevTwoTrackers (#652): selfdev with linear_team set loads
// pr-work-linear and the Linear intake signal next to GitHub pr-work;
// with it empty, the signal is off and pr-work-linear has no workspace.
func TestSelfdevTwoTrackers(t *testing.T) {
	pack, err := filepath.Abs("../../packs/selfdev")
	require.NoError(t, err)
	load := func(vars string) *Config {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{"jobshop.yaml": "projects:\n  js: {}\nagents:\n  jcode: {command: jcode}\npacks:\n  - path: " + pack + "\n    vars: {repo: /src/jobshop, agent: jcode" + vars + "}\n"})
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		return c
	}
	linearSignal := func(c *Config) StallSignal {
		for _, s := range c.Triggers["stall"].When.Stall {
			if s.Signal == StallLinearIssueNoRun {
				return s
			}
		}
		t.Fatal("no linear_issue_no_run signal")
		return StallSignal{}
	}

	c := load(", linear_workspace: gfx-labs, linear_team: JOB, linear_intake_labels: \"repo:github.com/elee1766/jobshop\", linear_human: eddy")
	gh, lin := c.Workflows["pr-work"], c.Workflows["pr-work-linear"]
	assert.Equal(t, TrackerGitHub, gh.Tracker)
	assert.Empty(t, gh.LinearWorkspace, "pr-work stays on GitHub")
	assert.Equal(t, TrackerLinear, lin.Tracker)
	assert.Equal(t, "gfx-labs", lin.LinearWorkspace)
	assert.Equal(t, "elee1766/jobshop", lin.Repo)
	require.Len(t, lin.Steps, 3)
	assert.Equal(t, []string{"dev", "refinery", "deploy-check"}, []string{lin.Steps[0].Pool, lin.Steps[1].Pool, lin.Steps[2].Pool})
	assert.Equal(t, "eddy", lin.PackVars["linear_human"])
	sig := linearSignal(c)
	assert.Equal(t, "JOB", sig.Team)
	assert.Equal(t, "repo:github.com/elee1766/jobshop", sig.Labels)

	c = load("")
	assert.Empty(t, linearSignal(c).Team, "signal off")
	assert.Empty(t, c.Workflows["pr-work-linear"].LinearWorkspace)
	assert.Equal(t, TrackerGitHub, c.Workflows["pr-work"].Tracker)
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
	assert.Equal(t, "/src/jobshop", c.Pools["dev"].Repo, "selfdev repo var is the checkout")
	assert.Contains(t, c.Triggers["master-red"].When.Check, "--repo elee1766/jobshop --branch master --workflow ci")
	assert.Equal(t, "go build ./... && go vet ./... && go test ./... && golangci-lint run ./... && nix fmt -- --check $(git ls-files '*.nix')", c.Workflows["pr-work"].PackVars["check_cmd"])
	var mem *RuleMemory
	var guide bool
	for _, r := range c.PromptRules {
		if r.Name == "selfdev-memory" {
			mem = r.Memory
		}
		if r.Name == "selfdev-guide" {
			guide = true
			assert.Equal(t, []string{"elee1766/jobshop"}, r.Match.Repos)
		}
	}
	require.NotNil(t, mem, "memory rule (#266)")
	assert.Equal(t, RuleMemory{Dir: "memory", Cap: 2200}, *mem, "local <root>/memory by default (#592)")
	assert.True(t, guide, "selfdev guide rule")
	assert.Equal(t, []string{"memory-write"}, c.Pools["learn"].Admin, "the curator writes local memory (#592)")
	assert.NotContains(t, c.Pools, "operator", "the operator pack is included once by the host (#248)")
	w := c.Workflows["pr-work"]
	require.Len(t, w.Steps, 3)
	assert.Equal(t, "plan", w.EscalatePool)
	assert.Equal(t, "squash", w.PackVars["merge_method"], "selfdev merges by squash")
	assert.Empty(t, w.PackVars["required_approver"], "selfdev needs no approval")
	assert.Equal(t, "deploy-check", w.Steps[2].Pool, "post-merge leaves the refinery free (#167)")
	b := c.Workflows["stall-check"]
	require.Len(t, b.Steps, 1)
	assert.Equal(t, 1, b.Steps[0].RetryLimit())
	assert.Equal(t, "plan", b.Steps[0].Pool)
	assert.Equal(t, "plan", b.EscalatePool)
	assert.Equal(t, []string{"signal", "subject", "detail"}, b.Vars)
	assert.NotContains(t, c.Workflows, "blocker-check", "folded into stall-check")
	assert.Contains(t, c.Pools["plan"].Admin, "session-restart")
	assert.Contains(t, c.Pools["plan"].Admin, "attention", "planner reads jobshop attention (#185)")
	st := c.Triggers["stall"]
	require.NotNil(t, st.Run)
	assert.Equal(t, "stall-check", st.Run.Workflow)
	assert.Equal(t, "{{.stall_subject}}", st.Run.Vars["subject"], "stall vars survive pack substitution")
	assert.Equal(t, 2, st.StallMaxRuns())
	sigs := map[string]StallSignal{}
	for _, s := range st.When.Stall {
		sigs[s.Signal] = s
	}
	assert.Equal(t, "3h", sigs[StallStepStalled].After)
	assert.Contains(t, sigs[StallIssueNoRun].Check, "gh issue list --repo elee1766/jobshop")
	assert.Contains(t, sigs[StallIssueNoRun].Check, `. == "proposed"`, "planner proposals wait on a human (#506)")
	assertIssueNoRunFilter(t, sigs[StallIssueNoRun].Check)
	assert.Equal(t, "elee1766/jobshop", sigs[StallIssueNoRun].Repo, "issue_no_run matches runs of its repo (#183)")
	assert.Equal(t, "elee1766/jobshop", c.Workflows["pr-work"].Repo)
	assert.Contains(t, sigs["no_merge"].Check, "repos/elee1766/jobshop/commits/master")
	assert.Contains(t, sigs["pr_unmerged"].Check, "statusCheckRollup")
	assertPRUnmergedFilter(t, sigs["pr_unmerged"].Check)
	mr := c.Triggers["master-red"]
	assert.Nil(t, mr.Ticket)
	require.NotNil(t, mr.Run, "master-red starts a run")
	assert.Equal(t, "ci-repair", mr.Run.Workflow)
	assert.Equal(t, "{{.check_output}}", mr.Run.Key)
	assert.Equal(t, map[string]string{"sha": "{{.check_output}}"}, mr.Run.Vars)
	ci := c.Workflows["ci-repair"]
	require.Len(t, ci.Steps, 2)
	assert.Equal(t, "plan", ci.EscalatePool)
	assert.Equal(t, 2, ci.Steps[0].RetryLimit())
	assert.Equal(t, []string{"dev", "refinery"}, []string{ci.Steps[0].Pool, ci.Steps[1].Pool})
	assert.Equal(t, "plan", c.Triggers["replan"].Ticket.Pool)
	// Curator (#273).
	assert.Equal(t, 1, c.Pools["learn"].MaxSessions)
	cu := c.Workflows["curate"]
	require.Len(t, cu.Steps, 1)
	assert.Equal(t, "learn", cu.Steps[0].Pool)
	ct := c.Triggers["curate"]
	require.NotNil(t, ct.Run)
	assert.Equal(t, "curate", ct.Run.Workflow)
	assert.Equal(t, "curate", ct.Run.Key)
	require.NotNil(t, ct.When.Learnings)
	assert.Equal(t, LearningsWhen{Min: 10, Every: "24h", Scope: []string{
		"repo:elee1766/jobshop", "pool:dev", "pool:plan", "pool:refinery", "pool:deploy-check", "pool:learn",
	}}, *ct.When.Learnings, "other instances' learnings do not start this curator (#524)")
	assert.Equal(t, "${SESSION_DIR}/jcode.sock", c.Agents["jcode"].Args[2], "host strings are not substituted")
}

// TestOperatorPackLoads loads packs/operator at host scope next to
// selfdev and a prefixed packs/github (#248): one operator pool, a
// prompt that reads every instance from config show instead of naming
// one repo, and instances that say which pools and workflows are whose.
func TestOperatorPackLoads(t *testing.T) {
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
    vars: {repo: /src/jobshop, agent: a}
  - path: ` + abs("github") + `
    prefix: acphost-
    vars: {repo: elee1766/acphost, checkout: /src/acphost, agent: a, human: wile, check_cmd: make test}
  - path: ` + abs("operator") + `
    vars: {jobshop_repo: acme/jobshop-fork, checkout: /src/jobshop, agent: a}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	var ops []string
	for n, p := range c.Pools {
		if p.Chat() {
			ops = append(ops, n)
		}
	}
	assert.Equal(t, []string{"operator"}, ops, "one operator pool for the host")
	p := c.Pools["operator"]
	assert.Equal(t, "/src/jobshop", p.Repo)
	assert.Equal(t, 0, p.MaxSessions)
	assert.Equal(t, []string{"config-reload", "config-show", "session-list", "pool-list", "pool-scale", "session-cordon", "session-config", "trigger-list", "attention", "activity", "ticket-release", "ticket-undefer", "decision-resolve"}, p.Admin)
	assert.NotContains(t, p.Prompt, "${")
	assert.NotContains(t, p.Prompt, "elee1766/jobshop", "the Jobshop repo comes from the var")

	require.Len(t, c.Instances, 3)
	sd, gh, op := c.Instances[0], c.Instances[1], c.Instances[2]
	assert.Equal(t, "elee1766/jobshop", sd.Vars["repo"], "the github base's repo, not selfdev's checkout path")
	assert.Equal(t, "/src/jobshop", sd.Vars["checkout"])
	assert.Equal(t, "github", sd.Vars["tracker"])
	assert.Contains(t, sd.Pools, "dev")
	assert.Contains(t, sd.Workflows, "pr-work")
	assert.Equal(t, "acphost-", gh.Prefix)
	assert.Equal(t, "elee1766/acphost", gh.Vars["repo"])
	assert.Contains(t, gh.Pools, "acphost-dev")
	assert.Contains(t, gh.Workflows, "acphost-pr-work")
	assert.Contains(t, gh.Triggers, "acphost-master-red")
	assert.Equal(t, []string{"operator"}, op.Pools)

	// Improvement sessions (#249): the rule matches improve:* keys of the
	// operator pool, and a push hook guards the pool.
	var rule *PromptRule
	for i := range c.PromptRules {
		if c.PromptRules[i].Name == "operator-improve" {
			rule = &c.PromptRules[i]
		}
	}
	require.NotNil(t, rule)
	assert.Equal(t, RuleAfter, rule.At)
	assert.True(t, rule.Match.Matches(RuleSubject{Pool: "operator", Chat: true, Key: "improve:prompts"}))
	assert.False(t, rule.Match.Matches(RuleSubject{Pool: "operator", Chat: true, Key: "matrix:!r:x"}))
	require.Len(t, c.Hooks, 1)
	assert.True(t, c.Hooks[0].Applies(HookAction, "push", "operator"))
	assert.False(t, c.Hooks[0].Applies(HookAction, "push", "dev"), "dev pools are not this pack's to guard")
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
	assert.ElementsMatch(t, []string{"pr-work", "pr-work-linear", "ci-repair", "stall-check", "drill"}, slices.Collect(maps.Keys(c.Workflows)))
	assert.ElementsMatch(t, []string{"replan", "master-red", "stall"}, slices.Collect(maps.Keys(c.Triggers)))
	assert.Equal(t, 2, c.Pools["dev"].MaxSessions, "dev_sessions default")
	assert.Equal(t, 3, c.Pools["deploy-check"].MaxSessions, "deploy_sessions default: one sleeping post-merge must not block the rest (#174)")
	for n, p := range c.Pools {
		assert.Equal(t, "/src/widget", p.Repo, n)
		assert.NotContains(t, p.Prompt, "elee1766", n)
		assert.NotContains(t, p.Prompt, "Jobshop", n)
	}
	mr := c.Triggers["master-red"]
	assert.Contains(t, mr.When.Check, "--repo acme/widget --branch drill/red-ci --workflow build")
	require.NotNil(t, mr.Run)
	assert.Equal(t, "ci-repair", mr.Run.Workflow)
	w := c.Workflows["pr-work"]
	assert.Equal(t, "make test", w.PackVars["check_cmd"])
}

// TestGithubPackProposeTargets: the planner prompt has the proposal
// rules by default and omits them with propose_targets false (#488).
func TestGithubPackProposeTargets(t *testing.T) {
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	for _, tc := range []struct {
		extra string
		want  bool
	}{{"", true}, {", propose_targets: false", false}} {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + pack + `
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile` + tc.extra + `}
`})
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		p := c.Pools["plan"].Prompt
		for _, s := range []string{"Proposing a target", "`gh issue create --label proposed", "never add the `target` label yourself", "-l proposed --meta"} {
			if tc.want {
				assert.Contains(t, p, s)
			} else {
				assert.NotContains(t, p, s)
			}
		}
		if tc.want {
			assert.NotContains(t, p, "comment that on the ticket and close it")
		} else {
			assert.Contains(t, p, "comment that on the ticket and close it")
			assert.NotContains(t, p, "proposed")
		}
	}
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
				out, err := w.Render(s.Prompt, vars)
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

// TestGithubPackTwoRepos: two packs/github instances with prefixes load
// side by side, each with its own pools, workflows and triggers (#184).
func TestGithubPackTwoRepos(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + pack + `
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test}
  - path: ` + pack + `
    prefix: gz-
    vars: {repo: acme/gizmo, checkout: /src/gizmo, agent: a, human: wile, check_cmd: make check}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"dev", "plan", "refinery", "deploy-check", "gz-dev", "gz-plan", "gz-refinery", "gz-deploy-check"}, slices.Collect(maps.Keys(c.Pools)))
	assert.Equal(t, "/src/widget", c.Pools["dev"].Repo)
	assert.Equal(t, "/src/gizmo", c.Pools["gz-dev"].Repo)
	assert.Equal(t, "a", c.Pools["gz-dev"].Agent, "host agent is not renamed")
	w := c.Workflows["gz-pr-work"]
	assert.Equal(t, "acme/gizmo", w.Repo)
	assert.Equal(t, "gz-plan", w.EscalatePool)
	assert.Equal(t, []string{"gz-dev", "gz-refinery", "gz-deploy-check"}, []string{w.Steps[0].Pool, w.Steps[1].Pool, w.Steps[2].Pool})
	mr := c.Triggers["gz-master-red"]
	assert.Equal(t, "gz-ci-repair", mr.Run.Workflow)
	assert.Contains(t, mr.When.Check, "--repo acme/gizmo")
	assert.Equal(t, "gz-dev", c.Triggers["gz-replan"].When.Pool)
	assert.Equal(t, "gz-plan", c.Triggers["gz-replan"].Ticket.Pool)
	assert.Equal(t, "gz-stall-check", c.Triggers["gz-stall"].Run.Workflow)
	assert.Equal(t, "gz-", w.PackVars[PackPrefixVar])

	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: " + pack + "\n    prefix: Bad\n"})
	_, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.ErrorContains(t, err, "prefix")
}

// TestGithubPackLinearTracker: tracker: linear swaps the plan, dev and
// pr-work prompts for ones that use jobshop linear, and the workflow
// carries the tracker and workspace (#232).
func TestGithubPackLinearTracker(t *testing.T) {
	dir := t.TempDir()
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
packs:
  - path: ` + pack + `
    prefix: eng-
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test, tracker: linear, linear_workspace: acme, linear_team: ENG, linear_target: ENG-1}
`})
	c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	w := c.Workflows["eng-pr-work"]
	assert.Equal(t, TrackerLinear, w.Tracker)
	assert.Equal(t, "acme", w.LinearWorkspace)
	assert.Equal(t, "acme/widget", w.Repo)
	for _, n := range []string{"eng-plan", "eng-dev"} {
		assert.NotContains(t, c.Pools[n].Prompt, "gh issue", n)
	}
	vars := maps.Clone(w.PackVars)
	vars["issue"] = "ENG-7"
	for _, s := range w.Steps {
		b, err := w.Render(s.Prompt, vars)
		require.NoError(t, err, s.ID)
		assert.NotContains(t, b, "gh issue", s.ID)
	}

	// The planner lists the team by the intake vars (#292).
	assert.Equal(t, "Todo", w.PackVars["linear_intake_states"])
	assert.Contains(t, w.PackVars, "linear_intake_assignee")

	// The github default keeps the gh prompts and no Linear workspace.
	writeFiles(t, dir, map[string]string{"jobshop.yaml": hostBase + "packs:\n  - path: " + pack + "\n    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile, check_cmd: make test}\n"})
	c, err = LoadFile(filepath.Join(dir, "jobshop.yaml"))
	require.NoError(t, err)
	assert.Equal(t, TrackerGitHub, c.Workflows["pr-work"].Tracker)
	assert.Empty(t, c.Workflows["pr-work"].LinearWorkspace)
	assert.Contains(t, c.Workflows["pr-work"].Steps[0].Prompt, "gh issue view")
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

// TestGithubPackEscalationChat (#379): with escalation_chat_pool set,
// every workflow's escalate block points at it, with the pack's prompts
// keeping the escalation vars for the controller; the policy rule is a
// system rule for every pool. Unset: no chat.
func TestGithubPackEscalationChat(t *testing.T) {
	pack, err := filepath.Abs("../../packs/github")
	require.NoError(t, err)
	load := func(extra string) *Config {
		dir := t.TempDir()
		writeFiles(t, dir, map[string]string{"jobshop.yaml": `projects:
  js: {}
agents:
  a: {command: x}
pools:
  esc: {agent: a, type: chat, dir: /w}
packs:
  - path: ` + pack + `
    vars: {repo: acme/widget, checkout: /src/widget, agent: a, human: wile` + extra + `}
`})
		c, err := LoadFile(filepath.Join(dir, "jobshop.yaml"))
		require.NoError(t, err)
		return c
	}
	c := load(", escalation_chat_pool: esc")
	for n, w := range c.Workflows {
		require.NotNil(t, w.Escalate, n)
		assert.Equal(t, "esc", w.Escalate.ChatPool, n)
	}
	e := c.Workflows["pr-work"].Escalate
	vars := map[string]any{}
	for _, k := range EscalationVars {
		vars[k] = ""
	}
	maps.Copy(vars, map[string]any{"ticket": "js-e", "run": "js-r", "workflow": "pr-work", "issue_ref": "issue 7", "body": "BODY"})
	out, err := Fill(e.Prompt, vars)
	require.NoError(t, err)
	assert.Contains(t, out, "run js-r (workflow pr-work)")
	assert.Contains(t, out, "`jobshop run resume js-r")
	assert.Contains(t, out, "issue 7")
	assert.True(t, strings.HasSuffix(strings.TrimSpace(out), "BODY"))
	assert.NotContains(t, out, "gateway")
	assert.NotContains(t, out, "thread")
	maps.Copy(vars, map[string]any{"pool": "dev", "error": "boom", "body": "B"})
	pout, err := Fill(e.PoolPrompt, vars)
	require.NoError(t, err)
	assert.Contains(t, pout, "Pool dev starts no sessions")
	assert.Contains(t, pout, "jobshop pool unquarantine dev")

	var policy *PromptRule
	for i := range c.PromptRules {
		if c.PromptRules[i].Name == "escalate-policy" {
			policy = &c.PromptRules[i]
		}
	}
	require.NotNil(t, policy, "escalation policy rule")
	assert.Equal(t, "system", policy.At)
	assert.Contains(t, policy.Text, "only for a decision, a credential")
	assert.Contains(t, policy.Text, "jobshop report progress")

	for n, w := range load("").Workflows {
		require.NotNil(t, w.Escalate, n)
		assert.Empty(t, w.Escalate.ChatPool, "no chat by default: "+n)
	}
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

// assertIssueNoRunFilter runs the issue_no_run jq filter on fixture
// issues: a proposed issue is not emitted, an unlabeled one is (#506).
func assertIssueNoRunFilter(t *testing.T, check string) {
	t.Helper()
	jq, err := exec.LookPath("jq")
	if err != nil {
		t.Log("jq not on PATH, filter not run")
		return
	}
	_, filter, ok := strings.Cut(check, "--jq '")
	require.True(t, ok, "check has a --jq filter")
	filter = strings.TrimSuffix(filter, "'")
	issues := `[
	 {"number":1,"title":"plain","labels":[],"createdAt":"2020-01-01T00:00:00Z","author":{"login":"bot","is_bot":true}},
	 {"number":2,"title":"Proposed target: x","labels":[{"name":"proposed"}],"createdAt":"2020-01-01T00:00:00Z","author":{"login":"bot","is_bot":true}}
	]`
	cmd := exec.CommandContext(t.Context(), jq, "-r", filter)
	cmd.Stdin = strings.NewReader(issues)
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "1 open, no run\n", string(out))
}

// assertPRUnmergedFilter runs the pr_unmerged jq filter (#509): PRs by the
// operator (elee1766 in selfdev) or labeled hold are skipped, other green
// PRs are flagged.
func assertPRUnmergedFilter(t *testing.T, check string) {
	t.Helper()
	jq, err := exec.LookPath("jq")
	if err != nil {
		t.Log("jq not on PATH, filter not run")
		return
	}
	_, filter, ok := strings.Cut(check, "--jq '")
	require.True(t, ok, "check has a --jq filter")
	filter = strings.TrimSuffix(filter, "'")
	green := `"isDraft":false,"updatedAt":"2020-01-01T00:00:00Z","statusCheckRollup":[{"conclusion":"SUCCESS"}]`
	prs := `[
	 {"number":1,` + green + `,"labels":[],"author":{"login":"elee-agent[bot]"}},
	 {"number":2,` + green + `,"labels":[{"name":"Hold"}],"author":{"login":"elee-agent[bot]"}},
	 {"number":3,` + green + `,"labels":[],"author":{"login":"Elee1766"}},
	 {"number":4,` + green + `,"labels":[{"name":"holding"}],"author":{"login":"elee"}}
	]`
	cmd := exec.CommandContext(t.Context(), jq, "-r", filter)
	cmd.Stdin = strings.NewReader(prs)
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "1 green, unmerged since 2020-01-01T00:00:00Z\n4 green, unmerged since 2020-01-01T00:00:00Z\n", string(out))
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
