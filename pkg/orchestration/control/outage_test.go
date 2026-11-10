package control_test

import (
	"fmt"
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
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// outageRoot writes a root whose dev pool acts as identity gh: a
// github-token read by a command that fails until <dir>/token exists, as
// the host's token provider did on 2026-11-10 (JOB-194). The API answers
// /rate_limit with the status in *rate (the probe). A workflow with one
// retrying step runs in dev; escalations go to ops, which starts no
// sessions.
func outageRoot(t *testing.T, rate *atomic.Int32) (config.Root, func() func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rate_limit" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(int(rate.Load()))
		_, _ = w.Write([]byte(`{"message":"Bad credentials"}`))
	}))
	t.Cleanup(srv.Close)
	root, start := poolRoot(t, "1m")
	cfg := fmt.Sprintf(`projects:
  js: {}
agents:
  fake:
    command: %q
    args: [fake-agent]
identities:
  gh:
    kind: github-token
    token_file: {type: command, command: "cat %s/token"}
    repos: ["acme/*"]
    api_url: %q
pools:
  dev:
    project: js
    agent: fake
    max_sessions: 1
    idle_timeout: 1m
    dir: %q
    identities: [gh]
  ops:
    project: js
    agent: fake
    max_sessions: 0
    dir: %q
workflows:
  pr:
    vars: [issue]
    issue: issue
    repo: acme/widget
    escalate_pool: ops
    steps:
      - {id: implement, pool: dev, retries: 3, prompt: "do it", title: "Implement #{{.issue}}"}
`, os.Args[0], root.Dir, srv.URL, root.Dir, root.Dir)
	require.NoError(t, os.WriteFile(root.Config(), []byte(cfg), 0o600))
	return root, start
}

func attentionOf(t *testing.T, admin *rpc.Client, kind string) []rpc.AttentionItem {
	t.Helper()
	var res rpc.AttentionResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodAttention, struct{}{}, &res))
	var out []rpc.AttentionItem
	for _, it := range res.Items {
		if it.Kind == kind {
			out = append(out, it)
		}
	}
	return out
}

// TestHostOutagePausesPool (JOB-194): the token provider fails, so every
// step that needs a token fails for the host's reason. The pool pauses with
// one attention row, no attempt is counted and nothing is escalated, the
// paused steps stay unclaimed, and when the provider recovers the probe
// lifts the pause and the steps run, with no undefer.
func TestHostOutagePausesPool(t *testing.T) {
	var rate atomic.Int32
	rate.Store(http.StatusUnauthorized)
	root, start := outageRoot(t, &rate)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "mint"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())
	status := func() rpc.PoolStatusResult {
		var st rpc.PoolStatusResult
		require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
		return st
	}

	var steps []string
	for _, issue := range []string{"1", "2"} {
		r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": issue}})
		steps = append(steps, r.Steps[0].ID)
	}
	passEventually(t, root.Dir, func() bool { return status().Outages["dev"].Identity == "gh" }, "dev paused by the outage")

	out := status().Outages["dev"]
	assert.Contains(t, out.Reason, "secret command", "the pause names the failure")
	assert.False(t, out.Since.IsZero())
	assert.Empty(t, status().Paused, "not a crash-loop pause")
	assert.Empty(t, status().Quarantined)

	// Paused: nothing is claimed, however many passes run.
	pass(t, root.Dir, 5)
	for _, id := range steps {
		got := getTicket(t, admin, id)
		assert.Equal(t, domain.StatusOpen, got.Ticket.Status, "paused steps stay open")
		assert.Zero(t, failures(t, got.Ticket), "the host's failure is no attempt")
	}
	var esc []domain.Ticket
	require.NoError(t, admin.Call(t.Context(), rpc.MethodTicketList, rpc.ListParams{Label: "needs-human"}, &esc))
	assert.Empty(t, esc, "no escalation")
	items := attentionOf(t, admin, rpc.AttentionOutage)
	require.Len(t, items, 1, "one attention item for two steps")
	assert.Equal(t, "gh", items[0].ID)
	assert.Contains(t, items[0].Title, "dev")
	assert.Empty(t, attentionOf(t, admin, rpc.AttentionRun), "no run needs a human")

	// The provider recovers but the API still refuses the token: the
	// probe fails and the pause holds.
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "token"), []byte("ghp_x\n"), 0o600))
	advance(t, root.Dir, 2*time.Minute)
	passEventually(t, root.Dir, func() bool { return strings.Contains(status().Outages["dev"].Reason, "401") }, "probe sees the 401")
	pass(t, root.Dir, 3)
	assert.Contains(t, status().Outages, "dev", "a failed probe keeps the pause")
	assert.Equal(t, domain.StatusOpen, getTicket(t, admin, steps[0]).Ticket.Status)

	// The API accepts it: the next probe lifts the pause and both steps run.
	rate.Store(http.StatusOK)
	advance(t, root.Dir, 2*time.Minute)
	passEventually(t, root.Dir, func() bool { return len(status().Outages) == 0 }, "probe lifts the pause")
	passEventually(t, root.Dir, func() bool {
		for _, id := range steps {
			if getTicket(t, admin, id).Ticket.Status != domain.StatusClosed {
				return false
			}
		}
		return true
	}, "steps run after the pause")
	for _, id := range steps {
		got := getTicket(t, admin, id)
		assert.Equal(t, "done", got.Ticket.Resolution)
		assert.Zero(t, failures(t, got.Ticket))
	}
	assert.Empty(t, attentionOf(t, admin, rpc.AttentionOutage), "the attention item goes with the pause")
}

// TestStepFailureWithoutOutage (JOB-194): a step failing for its own
// reasons, with the host's credentials fine, still counts attempts and
// escalates at the limit.
func TestStepFailureWithoutOutage(t *testing.T) {
	var rate atomic.Int32
	rate.Store(http.StatusOK)
	root, start := outageRoot(t, &rate)
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "token"), []byte("ghp_x\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "mint"), nil, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root.Dir, "failclose"), nil, 0o600))
	stop := start()
	defer stop()
	admin := dial(t, root.AdminSocket())

	r := startRun(t, admin, rpc.RunStartParams{Workflow: "pr", Vars: map[string]string{"issue": "9"}})
	step := r.Steps[0].ID
	passEventually(t, root.Dir, func() bool {
		return getTicket(t, admin, step).Ticket.Status == domain.StatusClosed
	}, "the step fails to its limit")
	got := getTicket(t, admin, step)
	assert.Equal(t, "failed", got.Ticket.Resolution)
	assert.Equal(t, 3, failures(t, got.Ticket))
	passEventually(t, root.Dir, func() bool { return len(escalations(t, admin, r.Run.ID)) == 1 }, "the run escalates")
	var st rpc.PoolStatusResult
	require.NoError(t, admin.Call(t.Context(), rpc.MethodPoolList, nil, &st))
	assert.Empty(t, st.Outages, "its own failure is no host outage")
}
