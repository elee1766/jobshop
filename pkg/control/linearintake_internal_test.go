package control

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
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
  stall-check:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Stall ${signal} ${subject}", prompt: "x"}
  pr-work:
    vars: [issue]
    issue: issue
    repo: acme/app
    tracker: linear
    linear_workspace: acme
    steps:
      - {id: implement, pool: dev, title: "Implement ${issue}", prompt: x}
triggers:
  stall:
    when:
      stall:
        - {signal: linear_issue_no_run, team: ENG, states: "Todo, Backlog", labels: agent, skip_labels: "needs-human", poll: 5m}
    run:
      workflow: stall-check
      vars: {signal: "${stall_signal}", subject: "${stall_subject}", detail: "${stall_detail}"}
`

// TestLinearIntake (#287): a matching Linear issue with no run fires one
// stall-check run; skip-labeled, wrong-state, closed, unlabeled and
// already-running issues do not, and Linear is read once per poll.
func TestLinearIntake(t *testing.T) {
	var reads atomic.Int32
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "lin_api_test" {
			_, _ = w.Write([]byte(`{"errors":[{"message":"Authentication required"}]}`))
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

	c, clk, dir := stallControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL

	// ENG-4 already has an open run.
	_, err := c.startRun(t.Context(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "eng-4"}, Key: "linear-ENG-4"}, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{"stall:linear_issue_no_run:ENG-1"}, stallKeys(t, c))
	assert.Equal(t, int32(1), reads.Load())

	c.runTriggers()
	stallRuns := func() []store.Ticket {
		rs, err := c.store.ListRuns(t.Context(), "", "", 0)
		require.NoError(t, err)
		var out []store.Ticket
		for _, r := range rs {
			if f, _ := store.DecodeFields[store.RunFields](r); f.Workflow == "stall-check" {
				out = append(out, r)
			}
		}
		return out
	}
	rs := stallRuns()
	require.Len(t, rs, 1)
	f, err := store.DecodeFields[store.RunFields](rs[0])
	require.NoError(t, err)
	assert.Equal(t, "stall:linear_issue_no_run:ENG-1", f.Key)
	assert.Equal(t, "ENG-1", f.Vars["subject"])
	assert.Equal(t, int32(1), reads.Load(), "within the poll interval the cached read is used")

	c.runTriggers()
	assert.Len(t, stallRuns(), 1, "fires once: the open run holds the key")

	// stall-check starts pr-work with key linear-ENG-1; the issue then
	// has an open run and is no longer a stall, and the start repeats safely.
	for range 2 {
		_, err = c.startRun(t.Context(), scope{admin: true, actor: "admin"}, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-1"}, Key: "linear-ENG-1"}, nil)
		require.NoError(t, err)
	}
	clk.advance(6 * time.Minute)
	assert.Empty(t, stallKeys(t, c))
	assert.Equal(t, int32(2), reads.Load(), "re-read after the poll interval")
}

// TestLinearIntakeOff: no team, no Linear read; an unreachable Linear
// finds nothing instead of failing the trigger.
func TestLinearIntakeOff(t *testing.T) {
	c, _, _ := stallControllerYAML(t, linearIntakeYAML)
	c.minter, c.linearURL = identity.NewMinter(), "http://127.0.0.1:1"
	found, err := c.stallLinearIssues(t.Context(), config.StallSignal{Signal: config.StallLinearIssueNoRun})
	require.NoError(t, err)
	assert.Empty(t, found)
	found, err = c.stallLinearIssues(t.Context(), config.StallSignal{Signal: config.StallLinearIssueNoRun, Team: "ENG"})
	require.NoError(t, err, "token file missing: logged, no stall")
	assert.Empty(t, found)
}

func TestLinearIntakeMatch(t *testing.T) {
	is := linear.Summary{State: linear.State{Name: "Todo", Type: "unstarted"}, Labels: []string{"Agent", "x"}}
	assert.True(t, linearIntakeMatch(is, nil, nil, nil))
	assert.True(t, linearIntakeMatch(is, []string{"todo"}, []string{"agent"}, nil))
	assert.False(t, linearIntakeMatch(is, []string{"Backlog"}, nil, nil))
	assert.False(t, linearIntakeMatch(is, nil, []string{"bug"}, nil))
	assert.False(t, linearIntakeMatch(is, nil, nil, []string{"X"}))
	is.State.Type = "canceled"
	assert.False(t, linearIntakeMatch(is, nil, nil, nil))
}

// TestLinearPlannerIntakeKey (#292): the planner and the intake
// stall-check both start pr-work with key linear-<KEY>, so whichever is
// second gets the first run, and the stall clears.
func TestLinearPlannerIntakeKey(t *testing.T) {
	gql := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"issues":{"nodes":[{"identifier":"ENG-1","title":"t","url":"u","updatedAt":"2026-11-01T00:00:00Z","state":{"id":"s","name":"Todo","type":"unstarted"},"assignee":null,"labels":{"nodes":[{"name":"agent"}]}}],"pageInfo":{"hasNextPage":false,"endCursor":null}}}}`))
	}))
	t.Cleanup(gql.Close)
	c, clk, dir := stallControllerYAML(t, linearIntakeYAML)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "token"), []byte("lin_api_test\n"), 0o600))
	c.minter, c.linearURL = identity.NewMinter(), gql.URL
	require.Equal(t, []string{"stall:linear_issue_no_run:ENG-1"}, stallKeys(t, c))

	admin := scope{admin: true, actor: "admin"}
	start := func() store.Ticket {
		res, err := c.startRun(t.Context(), admin, rpc.RunStartParams{Workflow: "pr-work", Vars: map[string]string{"issue": "ENG-1"}, Key: "linear-ENG-1"}, nil)
		require.NoError(t, err)
		return res.(rpc.RunResult).Run
	}
	planner, intake := start(), start()
	assert.Equal(t, planner.ID, intake.ID, "one run for both starters")
	clk.advance(6 * time.Minute)
	assert.Empty(t, stallKeys(t, c), "the planner's run hides the issue from intake")
}
