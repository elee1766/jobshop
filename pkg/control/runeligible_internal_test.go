package control

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// gateLinear serves ENG-5 for GetContext with a state, labels and an
// optional blocker the test sets.
type gateLinear struct {
	mu      sync.Mutex
	blocker bool
	labels  []string
	reads   int
}

func (f *gateLinear) serve(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query string `json:"query"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		f.mu.Lock()
		defer f.mu.Unlock()
		f.reads++
		var inv []any
		if f.blocker {
			inv = append(inv, map[string]any{"type": "blocks", "issue": map[string]any{"identifier": "ENG-4", "title": "first", "state": map[string]any{"name": "Todo", "type": "unstarted"}}})
		}
		var ls []any
		for _, l := range f.labels {
			ls = append(ls, map[string]any{"name": l})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"issue": map[string]any{
			"id": "iss-5", "identifier": "ENG-5", "title": "Fix login", "description": "d",
			"state":            map[string]any{"id": "s", "name": "Todo", "type": "unstarted"},
			"team":             map[string]any{"id": "t", "key": "ENG"},
			"labels":           map[string]any{"nodes": ls},
			"children":         map[string]any{"nodes": []any{}},
			"relations":        map[string]any{"nodes": []any{}},
			"inverseRelations": map[string]any{"nodes": inv},
			"comments":         map[string]any{"nodes": []any{}},
		}}})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (f *gateLinear) set(blocker bool, labels ...string) {
	f.mu.Lock()
	f.blocker, f.labels = blocker, labels
	f.mu.Unlock()
}

// TestRunStartEligibleGate (JOB-69): run start of a Linear workflow runs
// the intake gate. A blocked issue or one with a skip label is refused
// with the reason and no run; --force starts it; a repeated start with
// the key returns the open run without reading Linear again.
func TestRunStartEligibleGate(t *testing.T) {
	t.Parallel()
	c, _, _ := subController(t)
	tokFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokFile, []byte("lin_api_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"lin": {Kind: config.IdentityLinear, TokenFile: config.SecretPath(tokFile), Workspace: "acme", Teams: []string{"ENG"}}}
	cfg.Workflows = map[string]config.Workflow{"lw": {
		Vars: []string{"issue"}, Issue: "issue", Repo: "acme/app", Tracker: config.TrackerLinear, LinearWorkspace: "acme",
		Steps: []config.Step{{ID: "implement", Pool: "dev", Title: "Implement {{.issue}}", Prompt: "do {{.issue}}"}},
	}}
	cfg.Triggers = map[string]config.Trigger{"intake": {When: config.TriggerWhen{Stall: []config.StallSignal{
		{Signal: config.StallLinearIssueNoRun, Team: "ENG", SkipLabels: "Blocked, tracking"},
	}}}}
	cfg.Hooks = nil
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	f := &gateLinear{}
	c.minter, c.linearURL = identity.NewMinter(), f.serve(t).URL
	admin := scope{admin: true, actor: "admin"}
	start := func(force bool) (rpc.RunResult, error) {
		res, err := c.runStart(t.Context(), admin, rpc.RunStartParams{Workflow: "lw", Vars: map[string]string{"issue": "ENG-5"}, Key: "linear-ENG-5", Force: force})
		if err != nil {
			return rpc.RunResult{}, err
		}
		return res.(rpc.RunResult), nil
	}

	f.set(true)
	_, err = start(false)
	require.ErrorIs(t, err, store.ErrInvalid)
	assert.Contains(t, err.Error(), "blocked by ENG-4")
	assert.Contains(t, err.Error(), "--force")
	assert.Zero(t, runCount(t, c))

	f.set(false, "tracking")
	_, err = start(false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label tracking is skipped")
	assert.Zero(t, runCount(t, c))

	f.set(true)
	first, err := start(true)
	require.NoError(t, err)
	assert.False(t, first.Existing)
	assert.Equal(t, 1, runCount(t, c))

	f.mu.Lock()
	reads := f.reads
	f.mu.Unlock()
	again, err := start(false)
	require.NoError(t, err, "key reuse returns the open run")
	assert.True(t, again.Existing)
	assert.Equal(t, first.Run.ID, again.Run.ID)
	f.mu.Lock()
	assert.Equal(t, reads, f.reads, "key reuse does not read Linear")
	f.mu.Unlock()

	// Eligible: starts without --force.
	f.set(false)
	_, err = c.runCancel(t.Context(), admin, rpc.IDParams{ID: first.Run.ID})
	require.NoError(t, err)
	third, err := start(false)
	require.NoError(t, err)
	assert.False(t, third.Existing)
	assert.NotEqual(t, first.Run.ID, third.Run.ID)
}
