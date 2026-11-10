package control

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

const linearIntakeYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
identities:
  lin:
    kind: linear
    token_file: %[1]s/token
    workspace: acme
    teams: [ENG]
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
workflows:
  triage:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Finding {{.signal}} {{.subject}}", prompt: "x"}
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
        - {signal: linear_issue_no_run, team: ENG, states: "Todo, Backlog", labels: agent, skip_labels: "needs-human", poll: 5m}
    run:
      workflow: triage
      vars: {signal: "{{.signal}}", subject: "{{.subject}}", detail: "{{.detail}}"}
`

// TestLinearIntake (#287): a matching Linear issue with no run fires one
// triage run; skip-labeled, wrong-state, closed, unlabeled and
// already-running issues do not, and Linear is read once per poll.
func TestLinearIntake(t *testing.T) {
	t.Parallel()
	var reads atomic.Int32
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "lin_api_test" {
			_, _ = w.Write([]byte(`{"errors":[{"message":"Authentication required"}]}`))
			return
		}
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "issue(id:") {
			_, _ = w.Write([]byte(linearContextReply(b)))
			return
		}
		reads.Add(1)
		issue := func(id, state, typ, label string) string {
			return `{"identifier":"` + id + `","title":"t ` + id + `","url":"u","updatedAt":"2026-11-01T00:00:00Z","state":{"id":"s","name":"` + state + `","type":"` + typ + `"},"assignee":null,"labels":{"nodes":[{"name":"` + label + `"}]}}`
		}
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[` +
			issue("ENG-1", "Todo", "unstarted", "agent") + `,` +
			issue("ENG-2", "Todo", "unstarted", "Needs-Human") + `,` +
			issue("ENG-3", "In Review", "started", "agent") + `,` +
			issue("ENG-4", "Backlog", "backlog", "agent") + `,` +
			issue("ENG-5", "Todo", "unstarted", "other") + `,` +
			issue("ENG-6", "Todo", "completed", "agent") +
			`],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)

	c, clk, dir := scanControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL

	// ENG-4 already has an open run.
	_, err := c.startRun(t.Context(), c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "eng-4"}, Key: "linear-ENG-4"}, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"scan:linear_issue_no_run:ENG-1"}, findingKeys(t, c))
	assert.Equal(t, int32(1), reads.Load())

	c.runTriggers(t.Context(), c.config())
	scanRuns := func() []domain.Ticket {
		rs, err := c.store.List(t.Context(), store.Filter{Kind: domain.KindRun, Order: store.OrderNewest})
		require.NoError(t, err)
		var out []domain.Ticket
		for _, r := range rs {
			if f, _ := domain.DecodeFields[domain.RunFields](r); f.Workflow == "triage" {
				out = append(out, r)
			}
		}
		return out
	}
	rs := scanRuns()
	require.Len(t, rs, 1)
	f, err := domain.DecodeFields[domain.RunFields](rs[0])
	require.NoError(t, err)
	assert.Equal(t, "scan:linear_issue_no_run:ENG-1", f.Key)
	assert.Equal(t, "ENG-1", f.Vars["subject"])
	assert.Equal(t, int32(1), reads.Load(), "within the poll interval the cached read is used")

	c.runTriggers(t.Context(), c.config())
	assert.Len(t, scanRuns(), 1, "fires once: the open run holds the key")

	// triage starts pr-work with key linear-ENG-1; the issue then
	// has an open run and is no longer a finding, and the start repeats safely.
	for range 2 {
		_, err = c.startRun(t.Context(), c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-1"}, Key: "linear-ENG-1"}, nil)
		require.NoError(t, err)
	}
	clk.advance(6 * time.Minute)
	assert.Empty(t, findingKeys(t, c))
	assert.Equal(t, int32(2), reads.Load(), "re-read after the poll interval")
}

// linearContextReply answers a context read (JOB-68): ENG-7 is blocked in
// Linear, ENG-8 depends on an open GitHub issue (acme/lib#155), ENG-9 on
// a closed one; any other issue has neither.
func linearContextReply(body []byte) string {
	var q struct {
		Variables struct {
			ID string `json:"id"`
		} `json:"variables"`
	}
	_ = json.Unmarshal(body, &q)
	desc, inv := "", ""
	switch q.Variables.ID {
	case "ENG-7":
		inv = `{"type":"blocks","issue":{"identifier":"OPS-1","title":"x","state":{"name":"Todo","type":"unstarted"}}}`
	case "ENG-8":
		desc = "Depends on lib#155 (https://github.com/acme/lib/issues/155): x"
	case "ENG-9":
		desc = "Depends on acme/lib#154"
	}
	db, _ := json.Marshal(desc)
	return `{"data":{"issue":{"id":"i","identifier":"` + q.Variables.ID + `","title":"t","description":` + string(db) +
		`,"state":{"name":"Todo","type":"unstarted"},"team":{"key":"ENG"},"labels":{"nodes":[]},"children":{"nodes":[]},"relations":{"nodes":[]},"inverseRelations":{"nodes":[` + inv + `]},"comments":{"nodes":[]}}}}`
}

// TestLinearIntakeBlocked (JOB-68): an issue with an open Linear blocker
// or an open stated GitHub dependency is no finding; a closed dependency
// does not block.
func TestLinearIntakeBlocked(t *testing.T) {
	t.Parallel()
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "issue(id:") {
			_, _ = w.Write([]byte(linearContextReply(b)))
			return
		}
		issue := func(id string) string {
			return `{"identifier":"` + id + `","title":"t","url":"u","updatedAt":"2026-11-01T00:00:00Z","state":{"id":"s","name":"Todo","type":"unstarted"},"assignee":null,"labels":{"nodes":[{"name":"agent"}]}}`
		}
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[` + issue("ENG-7") + `,` + issue("ENG-8") + `,` + issue("ENG-9") +
			`],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)
	c, _, dir := scanControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	var asked []string
	c.ghIssueOpen = func(_ context.Context, d linear.GitHubDep) (bool, error) {
		asked = append(asked, d.String())
		return d.String() == "acme/lib#155", nil
	}
	assert.Equal(t, []string{"scan:linear_issue_no_run:ENG-9"}, findingKeys(t, c))
	assert.Equal(t, []string{"acme/lib#155", "acme/lib#154"}, asked, "bare repo takes the pr-work repo owner")
}

// TestLinearIntakeOff: no team, no Linear read; an unreachable Linear
// finds nothing instead of failing the trigger.
func TestLinearIntakeOff(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, linearIntakeYAML)
	c.minter, c.linearURL = identity.NewMinter(), "http://127.0.0.1:1"
	found, err := c.signalLinearIssues(t.Context(), c.config(), config.ScanSignal{Signal: config.SignalLinearIssueNoRun})
	require.NoError(t, err)
	assert.Empty(t, found)
	found, err = c.signalLinearIssues(t.Context(), c.config(), config.ScanSignal{Signal: config.SignalLinearIssueNoRun, Team: "ENG"})
	require.NoError(t, err, "token file missing: logged, no finding")
	assert.Empty(t, found)
}

// TestLinearPlannerIntakeKey (#292): the planner and the intake
// triage both start pr-work with key linear-<KEY>, so whichever is
// second gets the first run, and the finding clears.
func TestLinearPlannerIntakeKey(t *testing.T) {
	t.Parallel()
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"identifier":"ENG-1","title":"t","url":"u","updatedAt":"2026-11-01T00:00:00Z","state":{"id":"s","name":"Todo","type":"unstarted"},"assignee":null,"labels":{"nodes":[{"name":"agent"}]}}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)
	c, clk, dir := scanControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	require.Equal(t, []string{"scan:linear_issue_no_run:ENG-1"}, findingKeys(t, c))

	admin := scope{admin: true, actor: "admin"}
	start := func() domain.Ticket {
		res, err := c.startRun(t.Context(), c.config(), admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-1"}, Key: "linear-ENG-1"}, nil)
		require.NoError(t, err)
		return res.(rpc.RunResult).Run
	}
	planner, intake := start(), start()
	assert.Equal(t, planner.ID, intake.ID, "one run for both starters")
	clk.advance(6 * time.Minute)
	assert.Empty(t, findingKeys(t, c), "the planner's run hides the issue from intake")
}

// TestLinearIntakeRunDuringScan (JOB-26): a Todo issue whose run the
// planner starts while Linear is being read creates no triage run,
// and with grace set an issue updated within it is skipped until it passes.
func TestLinearIntakeRunDuringScan(t *testing.T) {
	t.Parallel()
	var c *Controller
	var started atomic.Bool
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if started.CompareAndSwap(false, true) {
			_, err := c.startRun(r.Context(), c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-1"}, Key: "linear-ENG-1"}, nil)
			assert.NoError(t, err)
		}
		issue := func(id, updated string) string {
			return `{"identifier":"` + id + `","title":"t","url":"u","updatedAt":"` + updated + `","state":{"id":"s","name":"Todo","type":"unstarted"},"assignee":null,"labels":{"nodes":[{"name":"agent"}]}}`
		}
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[` +
			issue("ENG-1", "2026-11-01T00:00:00Z") + `,` +
			issue("ENG-2", "2026-11-02T07:58:00Z") +
			`],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)

	c2, clk, dir := scanControllerYAML(t, strings.Replace(linearIntakeYAML, "poll: 5m}", "poll: 5m, grace: 5m}", 1))
	c = c2
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL

	scanRuns := func() []string {
		rs, err := c.store.List(t.Context(), store.Filter{Kind: domain.KindRun})
		require.NoError(t, err)
		var out []string
		for _, r := range rs {
			if f, _ := domain.DecodeFields[domain.RunFields](r); f.Workflow == "triage" {
				out = append(out, f.Key)
			}
		}
		return out
	}
	c.runTriggers(t.Context(), c.config())
	assert.Empty(t, scanRuns(), "ENG-1 got a run during the scan, ENG-2 is inside the grace window")

	clk.advance(6 * time.Minute)
	c.runTriggers(t.Context(), c.config())
	assert.Equal(t, []string{"scan:linear_issue_no_run:ENG-2"}, scanRuns(), "past the grace window")
}

// TestLinearIssueRecheck (JOB-26): the fire-time recheck sees a run
// started after detection.
func TestLinearIssueRecheck(t *testing.T) {
	t.Parallel()
	c, _, _ := scanControllerYAML(t, linearIntakeYAML)
	busy, err := c.linearSvc().HasOpenRun(t.Context(), "eng-7")
	require.NoError(t, err)
	assert.False(t, busy)
	_, err = c.startRun(t.Context(), c.config(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-7"}, Key: "linear-ENG-7"}, nil)
	require.NoError(t, err)
	busy, err = c.linearSvc().HasOpenRun(t.Context(), "eng-7")
	require.NoError(t, err)
	assert.True(t, busy)
}

// TestLinearIntakeRPC (JOB-121): linear.intake reports the states the
// team's linear_issue_no_run signals take, any state when one has none,
// and not watched for another team.
func TestLinearIntakeRPC(t *testing.T) {
	t.Parallel()
	sess := scope{session: "s1", actor: "s1"}
	c, _, _ := scanControllerYAML(t, linearIntakeYAML)
	res, err := c.linearIntakeRPC(t.Context(), sess, rpc.LinearIntakeParams{Team: "eng"})
	require.NoError(t, err)
	assert.Equal(t, rpc.LinearIntakeResult{Watched: true, States: []string{"Todo", "Backlog"}}, res)

	res, err = c.linearIntakeRPC(t.Context(), sess, rpc.LinearIntakeParams{Team: "OPS"})
	require.NoError(t, err)
	assert.Equal(t, rpc.LinearIntakeResult{}, res)

	_, err = c.linearIntakeRPC(t.Context(), sess, rpc.LinearIntakeParams{})
	require.ErrorIs(t, err, domain.ErrInvalid)

	c, _, _ = scanControllerYAML(t, strings.Replace(linearIntakeYAML, `states: "Todo, Backlog", `, "", 1))
	res, err = c.linearIntakeRPC(t.Context(), sess, rpc.LinearIntakeParams{Team: "ENG"})
	require.NoError(t, err)
	assert.Equal(t, rpc.LinearIntakeResult{Watched: true}, res)
}
