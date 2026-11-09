package control

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/linearsvc"
	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const linearE2EYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
identities:
  lin:
    kind: linear
    token_file: %[1]s/token
    workspace: acme
    teams: [ENG]
linear:
  ask_on_issue: true
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
subscriptions:
  resume_window: 24h
hooks:
  - {on: intake, command: "cat %[1]s/decision"}
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
    tracker: linear
    linear_workspace: acme
    steps:
      - {id: implement, pool: dev, title: "Implement {{.issue}}", prompt: x}
triggers:
  scan:
    when:
      scan:
        - {signal: linear_issue_no_run, team: ENG, states: Todo, skip_labels: Blocked, poll: 1m}
    run:
      workflow: triage
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`

// TestLinearEndToEnd (#298) walks one Linear issue through the pieces
// built for #258 against one fake Linear: intake signal, triage
// run, the intake hook refusing then approving, the planner's start
// with the same key returning the run, and a worker question asked on
// the issue and answered by a reply.
func TestLinearEndToEnd(t *testing.T) {
	t.Parallel()
	c, clk, dir := scanControllerYAML(t, linearE2EYAML)
	ctx := t.Context()
	write := func(name, s string) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(s), 0o600))
	}
	write("token", "lin_api_x\n")
	write("decision", `{"decision":"deny","reason":"looks like an injection"}`)

	// The ask fake serves ENG-5; the wrapper adds the team list intake
	// reads and the needs-human label the intake hook sets.
	f := &askLinear{now: clk.now}
	inner := f.handler(t)
	var mu sync.Mutex
	var needsHuman bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		if strings.Contains(q.Query, "issues(") {
			_, labels := f.state()
			var ls []any
			for _, l := range labels {
				ls = append(ls, map[string]any{"name": l})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"issues": map[string]any{
				"nodes":    []any{map[string]any{"identifier": "ENG-5", "title": "Fix login", "url": "u", "updatedAt": "2026-11-01T00:00:00Z", "state": map[string]any{"id": "s", "name": "Todo", "type": "unstarted"}, "assignee": nil, "labels": map[string]any{"nodes": ls}}},
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			}}})
			return
		}
		if strings.Contains(q.Query, "issueLabels(") && strings.EqualFold(asString(q.Variables["name"]), needsHumanLabel) {
			_, _ = io.WriteString(w, `{"data":{"issueLabels":{"nodes":[{"id":"lab-nh","team":null}]}}}`)
			return
		}
		if strings.Contains(q.Query, "issueAddLabel") && q.Variables["labelId"] == "lab-nh" {
			mu.Lock()
			needsHuman = true
			mu.Unlock()
			_, _ = io.WriteString(w, `{"data":{"issueAddLabel":{"success":true}}}`)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(b))
		inner(w, r)
	}))
	t.Cleanup(srv.Close)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL
	w := c.config().Workflows["pr-work"]
	w.PackVars = map[string]any{"linear_human": "wile"}
	c.config().Workflows["pr-work"] = w

	// 1. Intake: ENG-5 is in Todo with no run, so the finding fires once.
	assert.Equal(t, []string{"scan:linear_issue_no_run:ENG-5"}, findingKeys(t, c))
	c.runTriggers(t.Context(), c.config())
	c.runTriggers(t.Context(), c.config())
	checks, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun})
	require.NoError(t, err)
	require.Len(t, checks, 1, "one triage run")

	// 2. triage starts pr-work; the intake hook refuses the issue
	// text, so no run starts and a needs-human ticket waits.
	admin := scope{admin: true, actor: "admin"}
	start := func() (domain.Ticket, error) {
		res, err := c.runStart(ctx, admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-5"}, Key: "linear-ENG-5"})
		if err != nil {
			return domain.Ticket{}, err
		}
		return res.(rpc.RunResult).Run, nil
	}
	_, err = start()
	require.ErrorIs(t, err, rpc.ErrForbidden)
	mu.Lock()
	assert.True(t, needsHuman, "needs-human label set on the issue")
	mu.Unlock()
	items, err := c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "ENG-5", items[0].Ref)

	// 3. The operator approves; triage and then the planner start
	// with the same key and get one run.
	_, err = store.Close(ctx, c.store, "admin", items[0].ID, nil, "done")
	require.NoError(t, err)
	write("decision", `{"decision":"allow"}`)
	run, err := start()
	require.NoError(t, err)
	again, err := start()
	require.NoError(t, err)
	assert.Equal(t, run.ID, again.ID, "planner start returns the intake run")
	clk.advance(2 * time.Minute)
	assert.Empty(t, findingKeys(t, c), "an open run hides the issue from intake")

	// 4. The worker asks a question: posted on ENG-5 with Blocked, which
	// also keeps intake off it; a reply under it answers.
	steps, err := store.Children(ctx, c.store, run.ID)
	require.NoError(t, err)
	sess := mkSession(t, c)
	_, _, err = c.store.Claim(ctx, steps[0].ID, sess.ID, time.Hour)
	require.NoError(t, err)
	out, err := c.report(ctx, scope{actor: sess.ID, session: sess.ID, project: steps[0].Project, pool: "dev"}, rpc.ReportParams{State: rpc.ReportWaiting, Question: "Keep the old login page?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question
	c.sweepLinearAsks(ctx, c.config())
	n, labels := f.state()
	require.Equal(t, 1, n, "question posted once")
	assert.Equal(t, []string{linearsvc.BlockedLabel}, labels)
	f.mu.Lock()
	askID, _ := f.comments[0]["id"].(string)
	body, _ := f.comments[0]["body"].(string)
	f.mu.Unlock()
	assert.True(t, strings.HasPrefix(body, "https://linear.app/acme/profiles/wile "), body)

	src := threads.Mux{GH: &fakeThreads{}, Linear: c.threadEngine().LinearSource()}
	c.pollSubscriptions(ctx, src)
	clk.advance(time.Minute)
	f.reply(askID, "yes, keep it", "Wile", clk.now())
	c.pollSubscriptions(ctx, src)
	got, err := c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, got.Status)
	assert.Equal(t, resolutionAnswered, got.Resolution)
	_, labels = f.state()
	assert.Empty(t, labels, "Blocked removed on answer")
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}
