package packs_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/orchestration/control"
	"github.com/elee1766/jobshop/pkg/orchestration/control/controltest"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
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
	t.Helper()
	root := controltest.NewRoot(t, "pack: pack.cue\n")
	wd, err := os.Getwd()
	require.NoError(t, err)
	module := filepath.Dir(wd)
	write(t, root.Dir, "cue.mod/module.cue", "module: \"acceptance.example/host@v0\"\nlanguage: version: \"v0.17.0\"\ndeps: \"github.com/elee1766/jobshop@v0\": {v: \"v0.0.0\", default: true}\n")
	write(t, root.Dir, "cue.mod/local-module.cue", fmt.Sprintf("deps: \"github.com/elee1766/jobshop@v0\": replaceWith: %q\n", module))
	write(t, root.Dir, "pack.cue", fmt.Sprintf(`package acceptance
import (
    githubpack "github.com/elee1766/jobshop/packs/github"
    operatorpack "github.com/elee1766/jobshop/packs/operator"
    selfdevpack "github.com/elee1766/jobshop/packs/selfdev"
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
`, filepath.Join(module, "packs/github"), root.Dir, filepath.Join(module, "packs/operator"), root.Dir, filepath.Join(module, "packs/selfdev"), root.Dir))
	write(t, root.Dir, "gh-token", "ghp_test\n")
	write(t, root.Dir, config.FileName, fmt.Sprintf(`pack: pack.cue
automation: {github: true, shell: true}
identities:
  bot: {kind: github-token, token_file: %s, api_url: %s, repos: [acme/widget]}
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
