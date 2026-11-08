package control

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// A github-tracker instance (pr-work on GitHub issues) that also takes
// Linear team JOB, as packs/github renders it with linear_team set (#652).
const twoTrackersYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
identities:
  lin:
    kind: linear
    token_file: %[1]s/token
    workspace: acme
    teams: [JOB]
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
workflows:
  triage:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Finding {{.signal}} {{.subject}}", prompt: x}
  pr-work:
    vars: [issue]
    issue: issue
    repo: acme/app
    tracker: github
    steps:
      - {id: implement, pool: dev, title: "Implement #{{.issue}}", prompt: x}
  pr-work-linear:
    vars: [issue]
    issue: issue
    repo: acme/app
    tracker: linear
    linear_workspace: acme
    steps:
      - {id: implement, pool: dev, title: "Implement {{.issue}}", prompt: x}
triggers:
  scan:
    when:
      scan:
        - {signal: linear_issue_no_run, team: JOB, states: Todo, poll: 1m}
    run:
      workflow: triage
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`

// TestTwoTrackersLinearIntake (#652): on a github-tracker instance, a
// Linear intake issue fires linear_issue_no_run; the triage's start
// of pr-work-linear with key linear-JOB-1 makes a run carrying
// meta.linear, which then hides the issue from intake. GitHub pr-work
// still takes GitHub issue numbers.
func TestTwoTrackersLinearIntake(t *testing.T) {
	t.Parallel()
	c, clk, dir := scanControllerYAML(t, twoTrackersYAML)
	ctx := t.Context()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_x\n"), 0o600))
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query string `json:"query"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		if strings.Contains(q.Query, "issues(") {
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"issues": map[string]any{
				"nodes":    []any{map[string]any{"identifier": "JOB-1", "title": "Moved #631", "url": "u", "updatedAt": "2026-11-01T00:00:00Z", "state": map[string]any{"id": "s", "name": "Todo", "type": "unstarted"}, "assignee": nil, "labels": map[string]any{"nodes": []any{}}}},
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			}}})
			return
		}
		_, _ = io.WriteString(w, `{"data":{}}`)
	}))
	t.Cleanup(srv.Close)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL

	assert.Equal(t, []string{"scan:linear_issue_no_run:JOB-1"}, findingKeys(t, c))
	c.runTriggers(t.Context(), c.config())
	runs, err := c.store.List(ctx, store.Filter{Kind: store.KindRun})
	require.NoError(t, err)
	require.Len(t, runs, 1, "one triage run")

	admin := scope{admin: true, actor: "admin"}
	// Force: this fake serves no issue context for the intake gate (JOB-69).
	res, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work-linear", Vars: map[string]string{"issue": "JOB-1"}, Key: "linear-JOB-1", Force: true})
	require.NoError(t, err)
	run := res.(rpc.RunResult).Run
	var meta struct {
		Linear struct{ Workspace, Issue string }
		GitHub struct{ Repo string } `json:"github"`
	}
	require.NoError(t, json.Unmarshal(run.Meta, &meta))
	assert.Equal(t, "acme", meta.Linear.Workspace)
	assert.Equal(t, "JOB-1", meta.Linear.Issue)
	assert.Equal(t, "acme/app", meta.GitHub.Repo, "PRs stay on GitHub")
	again, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work-linear", Vars: map[string]string{"issue": "JOB-1"}, Key: "linear-JOB-1"})
	require.NoError(t, err)
	assert.Equal(t, run.ID, again.(rpc.RunResult).Run.ID, "same key, same run")
	clk.advance(2 * time.Minute)
	assert.Empty(t, findingKeys(t, c), "the open run hides JOB-1 from intake")

	gh, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "12"}, Key: "issue-12"})
	require.NoError(t, err)
	assert.JSONEq(t, `{"github":{"repo":"acme/app","issue":12}}`, string(gh.(rpc.RunResult).Run.Meta))
}
