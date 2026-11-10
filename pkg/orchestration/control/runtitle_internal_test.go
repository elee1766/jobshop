package control

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/orchestration/triggers"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// titleGitHub serves issue and PR titles of acme/app and counts reads.
type titleGitHub struct {
	mu    sync.Mutex
	reads map[string]int
}

func (g *titleGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mu.Lock()
	g.reads[r.URL.Path]++
	g.mu.Unlock()
	switch r.URL.Path {
	case "/repos/acme/app/issues/7":
		_, _ = w.Write([]byte(`{"number":7,"title":"Add  a\nthing"}`))
	case "/repos/acme/app/pulls/12":
		_, _ = w.Write([]byte(`{"number":12,"title":"fix: the thing"}`))
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`))
	}
}

func (g *titleGitHub) count(path string) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.reads[path]
}

func titleController(t *testing.T) (*Controller, *titleGitHub) {
	t.Helper()
	c, _, _ := subController(t)
	gh := &titleGitHub{reads: map[string]int{}}
	srv := httptest.NewServer(gh)
	t.Cleanup(srv.Close)
	pat := filepath.Join(t.TempDir(), "pat")
	require.NoError(t, os.WriteFile(pat, []byte("ghp_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"me": {Kind: config.IdentityGitHubToken, TokenFile: config.SecretPath(pat), Repos: []string{"acme/*"}, APIURL: srv.URL}}
	step := []config.Step{{ID: "work", Pool: "dev", Title: "Work", Prompt: "do it"}}
	cfg.Workflows = map[string]config.Workflow{
		"pw":    {Vars: []string{"issue"}, Issue: "issue", Repo: "acme/app", Steps: step},
		"rv":    {Vars: []string{"pr"}, PR: "pr", Repo: "acme/app", Steps: step},
		"plain": {Steps: step},
	}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter = identity.NewMinter()
	return c, gh
}

func startTitled(t *testing.T, c *Controller, p rpc.RunStartParams) domain.Ticket {
	t.Helper()
	res, err := c.startRun(t.Context(), c.config(), scope{admin: true, actor: "admin"}, p, nil)
	require.NoError(t, err)
	return res.(rpc.RunResult).Run
}

// TestRunTitle (JOB-133): a run of an issue or PR workflow is titled with
// the source's own title, one line; the key and meta are unchanged.
func TestRunTitle(t *testing.T) {
	t.Parallel()
	c, gh := titleController(t)

	r := startTitled(t, c, rpc.RunStartParams{Workflow: "pw", Vars: map[string]string{"issue": "7"}, Key: "issue-7"})
	assert.Equal(t, "Add a thing", r.Title)
	assert.Equal(t, 7, int(metaNumber(t, r, "issue")))
	f, err := domain.DecodeFields[domain.RunFields](r)
	require.NoError(t, err)
	assert.Equal(t, "issue-7", f.Key, "the dedup key does not depend on the title")

	// The open run for the key is returned without another read.
	again := startTitled(t, c, rpc.RunStartParams{Workflow: "pw", Vars: map[string]string{"issue": "7"}, Key: "issue-7"})
	assert.Equal(t, r.ID, again.ID)
	assert.Equal(t, 1, gh.count("/repos/acme/app/issues/7"))

	pr := startTitled(t, c, rpc.RunStartParams{Workflow: "rv", Vars: map[string]string{"pr": "#12"}, Key: "pr-12"})
	assert.Equal(t, "fix: the thing", pr.Title)

	named := startTitled(t, c, rpc.RunStartParams{Workflow: "pw", Vars: map[string]string{"issue": "7"}, Key: "other", Title: "  Chosen  title "})
	assert.Equal(t, "Chosen title", named.Title, "an explicit title wins and is not read")
	assert.Equal(t, 1, gh.count("/repos/acme/app/issues/7"))

	plain := startTitled(t, c, rpc.RunStartParams{Workflow: "plain", Key: "k"})
	assert.Equal(t, "Run plain k", plain.Title, "no source: the default title")
}

// A title that cannot be read does not stop the run: it is titled by its
// source.
func TestRunTitleFallback(t *testing.T) {
	t.Parallel()
	c, gh := titleController(t)
	r := startTitled(t, c, rpc.RunStartParams{Workflow: "pw", Vars: map[string]string{"issue": "99"}, Key: "issue-99"})
	assert.Equal(t, "Issue acme/app#99", r.Title)
	assert.Equal(t, 1, gh.count("/repos/acme/app/issues/99"))
	pr := startTitled(t, c, rpc.RunStartParams{Workflow: "rv", Vars: map[string]string{"pr": "98"}, Key: "pr-98"})
	assert.Equal(t, "PR acme/app#98", pr.Title)
}

// In a pass the read runs off the tick: the start returns triggers.ErrCheckPending
// with the read noted, and the replay with its result starts the run.
func TestRunTitleOffTick(t *testing.T) {
	t.Parallel()
	c, gh := titleController(t)
	p := rpc.RunStartParams{Workflow: "pw", Vars: map[string]string{"issue": "7"}, Key: "issue-7"}
	cp := &triggers.Pass{}
	_, err := c.startRun(triggers.WithPass(t.Context(), cp), c.config(), scope{admin: true, actor: "admin"}, p, nil)
	require.ErrorIs(t, err, triggers.ErrCheckPending)
	require.Len(t, cp.Need, 1)
	assert.Zero(t, gh.count("/repos/acme/app/issues/7"), "nothing is read inside the pass")
	assert.Zero(t, runCount(t, c))

	results := c.triggerEngine().RunChecks(t.Context(), cp.Need)
	res, err := c.startRun(triggers.WithPass(t.Context(), &triggers.Pass{Results: results}), c.config(), scope{admin: true, actor: "admin"}, p, nil)
	require.NoError(t, err)
	assert.Equal(t, "Add a thing", res.(rpc.RunResult).Run.Title)
}

// A scan finding about an issue titles its run, whatever the workflow
// (triage), and the key stays scan:<signal>:<subject>.
func TestFindingTitle(t *testing.T) {
	t.Parallel()
	c, _ := titleController(t)
	tr := config.Trigger{
		When: config.TriggerWhen{Scan: []config.ScanSignal{{Signal: config.SignalIssueNoRun, Repo: "acme/app"}, {Signal: config.SignalNeedsHuman}}},
		Run:  &config.TriggerRun{Workflow: "plain"},
	}
	cfg := c.config()
	run, _, err := c.fireFinding(t.Context(), cfg, "scan", tr, triggers.Finding{Signal: config.SignalIssueNoRun, Subject: "7", Detail: "open, no run"})
	require.NoError(t, err)
	assert.Equal(t, "Add a thing", run.Title)
	assert.Equal(t, "scan:issue_no_run:7", metaString(run.Meta, triggerPack, "key")[len("scan/"):])

	stuck, _, err := c.fireFinding(t.Context(), cfg, "scan", tr, triggers.Finding{Signal: config.SignalNeedsHuman, Subject: "js-1", Detail: "stuck"})
	require.NoError(t, err)
	assert.Equal(t, "Run plain scan:needs_human:js-1", stuck.Title, "stuck work keeps its descriptive title")
}

func metaNumber(t *testing.T, r domain.Ticket, key string) float64 {
	t.Helper()
	var m map[string]map[string]any
	require.NoError(t, json.Unmarshal(r.Meta, &m))
	n, _ := m["github"][key].(float64)
	return n
}
