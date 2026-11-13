package packs_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control"
	"github.com/gfx-labs/jobshop/pkg/orchestration/control/controltest"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// fakeGitHub answers the reads the github pack's triggers make. Times
// are relative to controltest.Epoch.
type fakeGitHub struct {
	mu       sync.Mutex
	ciFailed bool
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body string
	switch {
	case r.URL.Path == "/repos/acme/widget/actions/runs":
		conclusion := "success"
		if f.ciFailed {
			conclusion = "failure"
		}
		body = fmt.Sprintf(`{"workflow_runs":[{"name":"build","conclusion":"success","head_sha":"x"},{"name":"ci","conclusion":%q,"head_sha":"abc123"}]}`, conclusion)
	case r.URL.Path == "/repos/acme/widget/issues" && r.Method == http.MethodGet:
		body = `[
			{"number":7,"title":"Add a thing","labels":[],"user":{"login":"Wile","type":"User"},"created_at":"2026-11-04T10:00:00Z"},
			{"number":8,"title":"Too new","labels":[],"user":{"login":"wile","type":"User"},"created_at":"2026-11-04T11:50:00Z"},
			{"number":9,"title":"Planning","labels":[{"name":"target"}],"user":{"login":"wile","type":"User"},"created_at":"2026-11-04T10:00:00Z"},
			{"number":10,"title":"Stranger","labels":[],"user":{"login":"someone","type":"User"},"created_at":"2026-11-04T10:00:00Z"},
			{"number":11,"title":"A PR","pull_request":{},"labels":[],"user":{"login":"wile","type":"User"},"created_at":"2026-11-04T10:00:00Z"}
		]`
	case r.URL.Path == "/repos/acme/widget/pulls":
		body = `[]`
	case r.URL.Path == "/repos/acme/widget/issues/7":
		body = `{"number":7,"title":"Add a thing"}`
	case r.URL.Path == "/repos/acme/gizmo/issues/5":
		body = `{"number":5,"title":"Gizmo thing"}`
	case r.URL.Path == "/repos/acme/gizmo/actions/runs":
		body = `{"workflow_runs":[]}`
	case strings.HasPrefix(r.URL.Path, "/repos/acme/gizmo/"):
		body = `[]`
	case r.URL.Path == "/repos/acme/widget/pulls/20":
		body = `{"number":20,"title":"feat: add the thing"}`
	case r.URL.Path == "/repos/acme/widget/pulls/21":
		body = `{"number":21,"title":"chore: release 1.2.3"}`
	case r.URL.Path == "/graphql":
		pr := func(n int, draft bool, author, label, checks string) string {
			labels := `[]`
			if label != "" {
				labels = fmt.Sprintf(`[{"name":%q}]`, label)
			}
			return fmt.Sprintf(`{"number":%d,"isDraft":%t,"updatedAt":"2026-11-04T08:00:00Z","author":{"login":%q},"labels":{"nodes":%s},
				"commits":{"nodes":[{"commit":{"statusCheckRollup":{"contexts":{"nodes":%s}}}}]}}`, n, draft, author, labels, checks)
		}
		green := `[{"__typename":"CheckRun","conclusion":"SUCCESS"},{"__typename":"StatusContext","state":"SUCCESS"}]`
		red := `[{"__typename":"CheckRun","conclusion":"FAILURE"}]`
		body = fmt.Sprintf(`{"data":{"repository":{"pullRequests":{"nodes":[%s,%s,%s,%s,%s]}}}}`,
			pr(20, false, "bot", "", green),
			pr(21, false, "bot", "autorelease: pending", green),
			pr(22, false, "bot", "Hold", green),
			pr(23, false, "bot", "", red),
			pr(24, true, "bot", "", green))
	default:
		http.Error(w, "unexpected "+r.Method+" "+r.URL.Path, http.StatusNotFound)
		return
	}
	_, _ = io.WriteString(w, body)
}

// packRoot composes the github and operator packs from this module.
func packRoot(t *testing.T, apiURL string) config.Root {
	return packRootWith(t, apiURL, "")
}

// packRootWith is packRoot with extra CUE appended to the root pack.
func packRootWith(t *testing.T, apiURL, extra string) config.Root {
	t.Helper()
	root := controltest.NewRoot(t, "pack: pack.cue\n")
	wd, err := os.Getwd()
	require.NoError(t, err)
	module := filepath.Dir(wd)
	write(t, root.Dir, "cue.mod/module.cue", "module: \"acceptance.example/host@v0\"\nlanguage: version: \"v0.17.0\"\ndeps: \"github.com/gfx-labs/jobshop@v0\": {v: \"v0.0.0\", default: true}\n")
	write(t, root.Dir, "cue.mod/local-module.cue", fmt.Sprintf("deps: \"github.com/gfx-labs/jobshop@v0\": replaceWith: %q\n", module))
	write(t, root.Dir, "pack.cue", fmt.Sprintf(`package acceptance
import (
    githubpack "github.com/gfx-labs/jobshop/packs/github"
    operatorpack "github.com/gfx-labs/jobshop/packs/operator"
    selfdevpack "github.com/gfx-labs/jobshop/packs/selfdev"
)
default_project: "js"
projects: js: {}
agents: local: {command: "false"}
packs: github: githubpack.#Pack & {
    prefix: "github-"
    path: %q
    vars: {repo: "acme/widget", checkout: %q, agent: "local", human: "wile", dev_sessions: 0, refinery_sessions: 0, dev_sonnet_sessions: 0, deploy_sessions: 0}
}
packs: operator: operatorpack.#Pack & {
    path: %q
    vars: {jobshop_repo: "acme/jobshop", checkout: %q, agent: "local"}
}
packs: selfdev: selfdevpack.#Pack & {
    path: %q
    #Config: {repo: %q, agent: "local", dev_sessions: 0, refinery_sessions: 0, dev_sonnet_sessions: 0, learn_sessions: 0}
}
`, filepath.Join(module, "packs/github"), root.Dir, filepath.Join(module, "packs/operator"), root.Dir, filepath.Join(module, "packs/selfdev"), root.Dir)+strings.ReplaceAll(extra, "$MODULE", module))
	write(t, root.Dir, "gh-token", "ghp_test\n")
	write(t, root.Dir, config.FileName, fmt.Sprintf(`pack: pack.cue
automation: {github: true, shell: true}
identities:
  bot: {kind: github-token, token_file: %s, api_url: %s, repos: [acme/widget, acme/gizmo]}
# The operator guard, unscoped so the admin socket reaches it.
hooks:
  - {on: action, actions: [push], handler: %s, args: {protected: [master, main]}, timeout: 5s}
`, filepath.Join(root.Dir, "gh-token"), apiURL, filepath.Join(module, "packs/operator/guard.star:push")))
	return root
}

func write(t *testing.T, root, name, data string) {
	t.Helper()
	file := filepath.Join(root, name)
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o700))
	require.NoError(t, os.WriteFile(file, []byte(data), 0o600))
}

func runKeys(t *testing.T, c *controltest.Controller, workflow string) []string {
	t.Helper()
	keys := make([]string, 0)
	for _, r := range workflowRuns(t, c, workflow) {
		f, err := domain.DecodeFields[domain.RunFields](r)
		require.NoError(t, err)
		keys = append(keys, f.Key)
	}
	slices.Sort(keys)
	return keys
}

// runTitle is the title of the one run of workflow.
func runTitle(t *testing.T, c *controltest.Controller, workflow string) string {
	t.Helper()
	runs := workflowRuns(t, c, workflow)
	require.Len(t, runs, 1, workflow)
	return runs[0].Title
}

func workflowRuns(t *testing.T, c *controltest.Controller, workflow string) []domain.Ticket {
	t.Helper()
	var runs, out []domain.Ticket
	c.Call(rpc.MethodRunList, rpc.RunListParams{}, &runs)
	for _, r := range runs {
		f, err := domain.DecodeFields[domain.RunFields](r)
		require.NoError(t, err)
		if f.Workflow == workflow {
			out = append(out, r)
		}
	}
	return out
}

// The in-repo packs load together (selfdev's workflows and triggers
// included); the github pack's triggers read GitHub through the identity
// and start their runs; the operator guard denies pushes to master.
func TestPacks(t *testing.T) {
	gh := &fakeGitHub{}
	server := httptest.NewServer(gh)
	defer server.Close()
	root := packRoot(t, server.URL)
	_, err := config.Load(root)
	require.NoError(t, err)
	c := controltest.Start(t, root, control.Options{})

	c.PassUntil(func() bool { return len(runKeys(t, c, "github-pr-review")) > 0 }, "pr-unmerged did not start a review")
	assert.Equal(t, []string{"pr-20"}, runKeys(t, c, "github-pr-review"), "release, held, red and draft PRs are skipped")
	assert.Equal(t, "feat: add the thing", runTitle(t, c, "github-pr-review"), "a PR run is titled with the PR")
	c.PassUntil(func() bool { return len(runKeys(t, c, "github-release-merge")) > 0 }, "release-pr did not start a merge")
	assert.Equal(t, []string{"release-pr-21"}, runKeys(t, c, "github-release-merge"))
	assert.Equal(t, "chore: release 1.2.3", runTitle(t, c, "github-release-merge"))
	c.PassUntil(func() bool { return len(runKeys(t, c, "github-triage")) > 0 }, "issue_no_run did not start triage")
	assert.Equal(t, []string{"scan:issue_no_run:7"}, runKeys(t, c, "github-triage"), "new, planning, outside and PR items are skipped")
	assert.Equal(t, "Add a thing", runTitle(t, c, "github-triage"), "an issue's triage run is titled with the issue")

	var planned []domain.Ticket
	c.PassUntil(func() bool {
		c.Call(rpc.MethodTicketList, rpc.ListParams{Kind: "task", Pool: "github-plan", Meta: map[string]string{"trigger.name": "github-replan"}}, &planned)
		return len(planned) > 0
	}, "replan did not file a planning task")
	assert.Len(t, planned, 1)

	assert.Empty(t, runKeys(t, c, "github-ci-repair"))
	gh.mu.Lock()
	gh.ciFailed = true
	gh.mu.Unlock()
	c.Clock.Advance(6 * 60e9)
	c.PassUntil(func() bool { return len(runKeys(t, c, "github-ci-repair")) > 0 }, "master-red did not start a repair")
	assert.Equal(t, []string{"abc123"}, runKeys(t, c, "github-ci-repair"))
	assert.Equal(t, "Repair red master CI at abc123", runTitle(t, c, "github-ci-repair"), "operational runs say what the job is")

	guard := func(target string) rpc.GuardResult {
		var r rpc.GuardResult
		c.Call(rpc.MethodGuard, rpc.GuardParams{Action: "push", Target: target}, &r)
		return r
	}
	assert.Equal(t, domain.GuardDeny, guard("").Decision, "no target")
	for _, ref := range []string{"master", "origin/main", "HEAD:master", "refs/heads/main"} {
		r := guard(ref)
		assert.Equal(t, domain.GuardDeny, r.Decision, ref)
		assert.Contains(t, r.Reason, "never push to", ref)
	}
	assert.Equal(t, domain.GuardAllow, guard("fix-301").Decision)
}

// replan counts claimable tickets in its own instance's dev pool (#1221).
// packRoot composes the github pack twice: unprefixed through selfdev
// (dev, plan) and as github- (github-dev, github-plan). A deferred ticket
// in dev holds neither: selfdev's replan does not count it, and
// github-replan reads github-dev, not dev.
func TestReplanIgnoresDeferred(t *testing.T) {
	server := httptest.NewServer(&fakeGitHub{})
	defer server.Close()
	c := controltest.Start(t, packRoot(t, server.URL), control.Options{})
	planned := func(trigger, pool string, status domain.Status) []domain.Ticket {
		var out []domain.Ticket
		c.Call(rpc.MethodTicketList, rpc.ListParams{Kind: "task", Pool: pool, Status: status, Meta: map[string]string{"trigger.name": trigger}}, &out)
		return out
	}
	both := func() bool {
		return len(planned("replan", "plan", domain.StatusOpen)) == 1 && len(planned("github-replan", "github-plan", domain.StatusOpen)) == 1
	}
	c.PassUntil(both, "replan did not file planning tasks on empty dev pools")

	var deferred domain.Ticket
	c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Title: "waits a week", Pool: "dev"}, &deferred)
	until := controltest.Epoch.Add(7 * 24 * time.Hour)
	c.Call(rpc.MethodTicketUpdate, rpc.UpdateParams{ID: deferred.ID, DeferUntil: &until}, nil)
	for _, p := range append(planned("replan", "plan", domain.StatusOpen), planned("github-replan", "github-plan", domain.StatusOpen)...) {
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: p.ID}, nil)
	}
	c.Clock.Advance(11 * time.Minute)
	c.PassUntil(both, "a deferred ticket in dev held replan back")
	assert.Len(t, planned("replan", "plan", ""), 2)
	assert.Len(t, planned("github-replan", "github-plan", ""), 2)

	// Ready work in dev holds selfdev's replan, not github-replan.
	c.Call(rpc.MethodTicketCreate, rpc.CreateParams{Title: "ready now", Pool: "dev"}, nil)
	for _, p := range append(planned("replan", "plan", domain.StatusOpen), planned("github-replan", "github-plan", domain.StatusOpen)...) {
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: p.ID}, nil)
	}
	c.Clock.Advance(11 * time.Minute)
	c.PassUntil(func() bool { return len(planned("github-replan", "github-plan", domain.StatusOpen)) == 1 }, "ready work in dev held github-replan back")
	c.Pass(3)
	assert.Empty(t, planned("replan", "plan", domain.StatusOpen), "dev has ready work")
}

// pr-work routes before it implements: the route step's outcome picks the
// pool of the one implement step, and no step is made for the other route.
func TestPRWorkRoutes(t *testing.T) {
	server := httptest.NewServer(&fakeGitHub{})
	defer server.Close()
	c := controltest.Start(t, packRoot(t, server.URL), control.Options{})
	steps := func(run string) []domain.Ticket {
		var r rpc.RunResult
		c.Call(rpc.MethodRunShow, rpc.IDParams{ID: run}, &r)
		return r.Steps
	}
	for outcome, pool := range map[string]string{"scoped": "github-dev-sonnet", "primary": "github-dev"} {
		var started rpc.RunResult
		c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "github-pr-work", Key: "route-" + outcome, Vars: map[string]string{"issue": "7"}}, &started)
		run := started.Run.ID
		c.PassUntil(func() bool { return len(steps(run)) == 1 }, "no route step")
		route := steps(run)[0]
		assert.Equal(t, "github-triage", route.Pool)
		c.Call(rpc.MethodClaim, rpc.ClaimParams{ID: route.ID, Holder: "s1"}, nil)
		c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: route.ID, Outcome: outcome}, nil)
		c.PassUntil(func() bool { return len(steps(run)) == 2 }, "no implement step after route "+outcome)
		implement := steps(run)[1]
		f, err := domain.DecodeFields[domain.StepFields](implement)
		require.NoError(t, err)
		assert.Equal(t, "implement", f.Step)
		assert.Equal(t, pool, implement.Pool, outcome)
		c.Call(rpc.MethodRunCancel, rpc.IDParams{ID: run}, nil)
	}
	// A word the router does not know makes no step: the run waits on a person.
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "github-pr-work", Key: "route-unknown", Vars: map[string]string{"issue": "7"}}, &started)
	run := started.Run.ID
	c.PassUntil(func() bool { return len(steps(run)) == 1 }, "no route step")
	route := steps(run)[0]
	c.Call(rpc.MethodClaim, rpc.ClaimParams{ID: route.ID, Holder: "s1"}, nil)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: route.ID, Outcome: "luna"}, nil)
	c.PassUntil(func() bool {
		var r rpc.RunResult
		c.Call(rpc.MethodRunShow, rpc.IDParams{ID: run}, &r)
		f, err := domain.DecodeFields[domain.RunFields](r.Run)
		require.NoError(t, err)
		return f.State == domain.RunNeedsHuman
	}, "unknown route did not escalate")
	assert.Len(t, steps(run), 1)
}

// One planner over two repos (JOB-207): a second github entry with
// pools_from shares github-'s pools. It declares no pools or replan of
// its own, adds its repo to the shared code pools, and its runs' steps
// land in those pools carrying its repo.
func TestGithubPackSharedPools(t *testing.T) {
	gh := &fakeGitHub{}
	server := httptest.NewServer(gh)
	defer server.Close()
	gizmo := t.TempDir()
	root := packRootWith(t, server.URL, fmt.Sprintf(`packs: gizmo: githubpack.#Pack & {
    prefix: "gz-"
    pools_from: "github"
    path: "$MODULE/packs/github"
    vars: {repo: "acme/gizmo", checkout: %q, agent: "local", human: "wile"}
}
`, gizmo))
	cfg, err := config.Load(root)
	require.NoError(t, err)

	for _, p := range []string{"gz-dev", "gz-plan", "gz-refinery"} {
		assert.NotContains(t, cfg.Pools, p, "a pools_from entry declares no pools")
	}
	assert.NotContains(t, cfg.Triggers, "gz-replan", "one planner: the shared entry does not replan")
	assert.Contains(t, cfg.Triggers, "gz-master-red", "repo triggers stay per repo")
	for _, p := range []string{"github-dev", "github-dev-sonnet", "github-triage", "github-refinery", "github-deploy-check"} {
		pool := cfg.Pools[p]
		assert.Equal(t, gizmo, pool.Repos["acme/gizmo"].Checkout, p)
		assert.Equal(t, cfg.Pools["github-plan"].Repo, pool.Repo, "%s keeps github-'s checkout as its default", p)
		assert.NotContains(t, pool.Labels, "instance:gz", p)
	}
	assert.Empty(t, cfg.Pools["github-plan"].Repos, "the planner does not edit code")
	w := cfg.Workflows["gz-pr-work"]
	assert.Equal(t, "acme/gizmo", w.Repo)
	assert.Equal(t, "github-plan", w.EscalatePool)
	var rules []string
	for _, r := range cfg.PromptRules {
		rules = append(rules, r.Name)
	}
	assert.Contains(t, rules, "gz-shared-plan")
	assert.NotContains(t, rules, "gz-recall", "general rules come once, from github-")

	c := controltest.Start(t, root, control.Options{})
	steps := func(run string) []domain.Ticket {
		var r rpc.RunResult
		c.Call(rpc.MethodRunShow, rpc.IDParams{ID: run}, &r)
		return r.Steps
	}
	var started rpc.RunResult
	c.Call(rpc.MethodRunStart, rpc.RunStartParams{Workflow: "gz-pr-work", Key: "issue-5", Vars: map[string]string{"issue": "5"}}, &started)
	run := started.Run.ID
	c.PassUntil(func() bool { return len(steps(run)) == 1 }, "no route step")
	route := steps(run)[0]
	assert.Equal(t, "github-triage", route.Pool, "the step lands in the shared pool")
	assert.Contains(t, string(route.Meta), `"repo":"acme/gizmo"`, "a session keyed to acme/gizmo claims it")
	c.Call(rpc.MethodClaim, rpc.ClaimParams{ID: route.ID, Holder: "s1"}, nil)
	c.Call(rpc.MethodTicketClose, rpc.CloseParams{ID: route.ID, Outcome: "primary"}, nil)
	c.PassUntil(func() bool { return len(steps(run)) == 2 }, "no implement step")
	assert.Equal(t, "github-dev", steps(run)[1].Pool)
}
