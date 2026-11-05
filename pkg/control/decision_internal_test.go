package control

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// ghCall is one request the fake GitHub saw.
type ghCall struct {
	Method, Path string
	Body         map[string]any
}

// decisionGitHub serves testdata/github_decision.json (responses
// recorded from acphost#56) and records each call.
func decisionGitHub(t *testing.T) (*httptest.Server, func() []ghCall) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "github_decision.json"))
	require.NoError(t, err)
	var rec map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &rec))
	var mu sync.Mutex
	var calls []ghCall
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer ghp_x", r.Header.Get("Authorization"))
		b, _ := io.ReadAll(r.Body)
		var body map[string]any
		_ = json.Unmarshal(b, &body)
		mu.Lock()
		calls = append(calls, ghCall{r.Method, r.URL.Path, body})
		mu.Unlock()
		key := ""
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/app/issues/56/labels":
			key = "labels_add"
		case r.Method == http.MethodDelete && r.URL.Path == "/repos/acme/app/issues/56/labels/proposed":
			key = "labels_remove"
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/app/issues/56/comments":
			w.WriteHeader(http.StatusCreated)
			key = "comment"
		case r.Method == http.MethodPatch && r.URL.Path == "/repos/acme/app/issues/56":
			key = "issue_closed"
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Not Found"}`))
			return
		}
		_, _ = w.Write(rec[key])
	}))
	t.Cleanup(srv.Close)
	return srv, func() []ghCall {
		mu.Lock()
		defer mu.Unlock()
		return append([]ghCall(nil), calls...)
	}
}

func decisionController(t *testing.T) (*Controller, *stallClock, func() []ghCall) {
	t.Helper()
	c, clk, _ := subController(t)
	srv, calls := decisionGitHub(t)
	pat := filepath.Join(t.TempDir(), "pat")
	require.NoError(t, os.WriteFile(pat, []byte("ghp_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"me": {Kind: config.IdentityGitHubToken, TokenFile: config.SecretPath(pat), Repos: []string{"acme/*"}, APIURL: srv.URL}}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter = identity.NewMinter()
	m, err := telemetry.NewMetrics()
	require.NoError(t, err)
	c.metrics = m
	return c, clk, calls
}

// proposal makes the decision ticket and a planner ticket parked on it
// for 24h, as the planner did before #603.
func proposal(t *testing.T, c *Controller, clk *stallClock) (store.Ticket, store.Ticket) {
	t.Helper()
	d := mkTicket(t, c, store.NewTicket{Title: "Proposed target: v0.5.0 (#56)", Pool: "dev", Labels: []string{proposedLabel}, Meta: json.RawMessage(`{"github":{"repo":"acme/app","issue":56}}`)})
	parked := mkTicket(t, c, store.NewTicket{Title: "Plan: waiting on approval of acme/app#56", Pool: "dev", DeferUntil: clk.now().Add(24 * time.Hour)})
	other := mkTicket(t, c, store.NewTicket{Title: "Unrelated #560", Pool: "dev", DeferUntil: clk.now().Add(24 * time.Hour)})
	t.Cleanup(func() {
		got, err := c.store.Get(context.Background(), other.ID)
		require.NoError(t, err)
		assert.True(t, got.DeferUntil.After(clk.now()), "#560 is not #56")
	})
	return d, parked
}

func TestDecisionApprove(t *testing.T) {
	t.Parallel()
	c, clk, calls := decisionController(t)
	ctx := t.Context()
	d, parked := proposal(t, c, clk)
	out, err := c.decisionResolve(ctx, scope{admin: true, actor: "desk:alice"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionApprove, Note: "go"})
	require.NoError(t, err)
	res := out.(rpc.DecisionResolveResult)
	assert.Equal(t, "approved", res.Ticket.Resolution)
	assert.Equal(t, []string{parked.ID}, res.Undeferred)
	got := calls()
	require.Len(t, got, 3)
	assert.Equal(t, ghCall{"POST", "/repos/acme/app/issues/56/labels", map[string]any{"labels": []any{"target"}}}, got[0])
	assert.Equal(t, ghCall{"DELETE", "/repos/acme/app/issues/56/labels/proposed", nil}, got[1])
	assert.Equal(t, "POST", got[2].Method)
	assert.Equal(t, "Approved as target by desk:alice.\n\ngo", got[2].Body["body"])
	p, err := c.store.Get(ctx, parked.ID)
	require.NoError(t, err)
	assert.True(t, p.DeferUntil.IsZero(), "the parked planner ticket is ready now")

	_, err = c.decisionResolve(ctx, scope{admin: true, actor: "x"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionApprove})
	require.ErrorIs(t, err, store.ErrClosed, "a second click does not act twice")
}

func TestDecisionReject(t *testing.T) {
	t.Parallel()
	c, clk, calls := decisionController(t)
	d, parked := proposal(t, c, clk)
	out, err := c.decisionResolve(t.Context(), scope{admin: true, actor: "desk:alice"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionReject})
	require.NoError(t, err)
	res := out.(rpc.DecisionResolveResult)
	assert.Equal(t, "rejected", res.Ticket.Resolution)
	assert.Equal(t, []string{parked.ID}, res.Undeferred)
	got := calls()
	require.Len(t, got, 2)
	assert.Equal(t, "/repos/acme/app/issues/56/comments", got[0].Path)
	assert.Equal(t, ghCall{"PATCH", "/repos/acme/app/issues/56", map[string]any{"state": "closed", "state_reason": "not_planned"}}, got[1])
}

func TestDecisionDismiss(t *testing.T) {
	t.Parallel()
	c, clk, calls := decisionController(t)
	ctx := t.Context()
	d, parked := proposal(t, c, clk)
	out, err := c.decisionResolve(ctx, scope{admin: true, actor: "desk:alice"}, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionDismiss})
	require.NoError(t, err)
	res := out.(rpc.DecisionResolveResult)
	assert.Equal(t, "dismissed", res.Ticket.Resolution)
	assert.Equal(t, []string{parked.ID}, res.Undeferred)
	assert.Empty(t, calls(), "the issue is left as is")
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	for _, it := range items {
		assert.NotEqual(t, d.ID, it.ID, "dismissed is gone from attention")
	}
}

// TestDecisionGrant: sessions need decision-resolve; ticket.undefer needs
// ticket-undefer and stays in the session's project.
func TestDecisionGrant(t *testing.T) {
	t.Parallel()
	c, clk, _ := decisionController(t)
	ctx := t.Context()
	d, parked := proposal(t, c, clk)
	sc := scope{actor: "js-s1", session: "js-s1", project: "js", pool: "dev"}
	_, err := c.decisionResolve(ctx, sc, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionDismiss})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = c.ticketUndefer(ctx, sc, rpc.IDParams{ID: parked.ID})
	require.ErrorIs(t, err, rpc.ErrForbidden)

	cfg := *c.config()
	pools := map[string]config.Pool{}
	for k, v := range cfg.Pools {
		pools[k] = v
	}
	p := pools["dev"]
	p.Admin = []string{grantTicketUndefer, grantDecisionResolve}
	pools["dev"] = p
	cfg.Pools = pools
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)

	out, err := c.ticketUndefer(ctx, sc, rpc.IDParams{ID: parked.ID})
	require.NoError(t, err)
	assert.True(t, out.(store.Ticket).DeferUntil.IsZero())
	_, err = c.ticketUndefer(ctx, scope{actor: "x", session: "x", project: "other", pool: "dev"}, rpc.IDParams{ID: parked.ID})
	require.ErrorIs(t, err, rpc.ErrForbidden)
	_, err = c.decisionResolve(ctx, sc, rpc.DecisionResolveParams{ID: d.ID, Action: rpc.DecisionDismiss})
	require.NoError(t, err)
}

// TestWaitLabel: github.label clears when the issue gains the label, and
// when it is closed.
func TestWaitLabel(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	raw, err := os.ReadFile(filepath.Join("testdata", "github_decision.json"))
	require.NoError(t, err)
	var rec map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(raw, &rec))
	tk := mkTicket(t, c, store.NewTicket{Title: "proposal", Pool: "dev"})
	_, err = c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{ID: tk.ID, Reason: "approval", Conds: []store.WaitCond{{Type: store.WaitLabel, Repo: "acme/app", Issue: 56, Label: "target"}}})
	require.NoError(t, err)
	prs.issues["acme/app#56"] = string(rec["issue_proposed"])
	c.pollWaits(ctx, prs, threads)
	got, err := c.store.Get(ctx, tk.ID)
	require.NoError(t, err)
	require.NotNil(t, waitOf(got), "still proposed")
	prs.mu.Lock()
	prs.issues["acme/app#56"] = string(rec["issue_target"])
	prs.mu.Unlock()
	c.pollWaits(ctx, prs, threads)
	got, err = c.store.Get(ctx, tk.ID)
	require.NoError(t, err)
	assert.Nil(t, waitOf(got), "labeled target")

	is, err := parseIssue(rec["issue_closed"])
	require.NoError(t, err)
	assert.Equal(t, "closed", is.State)
	require.Error(t, (&store.Wait{Reason: "x", Conds: []store.WaitCond{{Type: store.WaitLabel, Repo: "acme/app"}}}).Validate(), "issue and label required")
}

// TestSendWakesRunner: a message for a session ends its runner's idle
// wait at once instead of at the next poll (#603).
func TestSendWakesRunner(t *testing.T) {
	t.Parallel()
	c, _, _ := subController(t)
	sess := mkSession(t, c)
	r := &runner{restartWake: make(chan struct{}, 1)}
	c.runners[sess.ID] = r
	_, err := c.sessionSend(t.Context(), scope{admin: true, actor: "desk:alice"}, rpc.SessionSendParams{ID: sess.ID, Text: "hi"})
	require.NoError(t, err)
	select {
	case <-r.restartWake:
	default:
		t.Fatal("runner not woken")
	}
}
