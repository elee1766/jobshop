package control

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
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

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const webhookSecret = "whsec"

const linearWebhookYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
identities:
  lapp:
    kind: linear-app
    client_id_file: %[1]s/cid
    client_secret_file: %[1]s/csecret
    webhook_secret_file: %[1]s/whsecret
    token_url: %[1]s/TOKENURL
    workspace: acme
    teams: [ENG]
linear:
  webhook_listen: 127.0.0.1:0
pools:
  dev: {agent: fake, dir: %[1]q}
  plan: {agent: fake, dir: %[1]q}
subscriptions:
  resume_window: 24h
workflows:
  stall-check:
    vars: [signal, subject, detail]
    escalate_pool: plan
    steps:
      - {id: diagnose, pool: plan, retries: 1, title: "Stall ${signal} ${subject}", prompt: x}
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
        - {signal: linear_issue_no_run, team: ENG, states: Todo, poll: 1h}
    run:
      workflow: stall-check
      vars: {signal: "${stall_signal}", subject: "${stall_subject}", detail: "${stall_detail}"}
`

// webhookLinear is a fake Linear for the webhook tests: an OAuth token
// endpoint, viewer (the App, org-1), a team list that is empty until
// listed is set, and askLinear for the rest.
type webhookLinear struct {
	mu     sync.Mutex
	listed bool
	ask    *askLinear
}

func (f *webhookLinear) serve(t *testing.T) *httptest.Server {
	inner := f.ask.handler(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = io.WriteString(w, `{"access_token":"oauth-tok","expires_in":3600}`)
			return
		}
		b, _ := io.ReadAll(r.Body)
		switch {
		case bytes.Contains(b, []byte("viewer")):
			_, _ = io.WriteString(w, `{"data":{"viewer":{"id":"app-user","organization":{"id":"org-1"}}}}`)
			return
		case bytes.Contains(b, []byte("issues(")):
			f.mu.Lock()
			var nodes []any
			if f.listed {
				nodes = []any{map[string]any{"identifier": "ENG-5", "title": "Fix login", "url": "u", "updatedAt": "2026-11-01T00:00:00Z", "state": map[string]any{"id": "s", "name": "Todo", "type": "unstarted"}, "assignee": nil, "labels": map[string]any{"nodes": []any{}}}}
			}
			f.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"issues": map[string]any{
				"nodes": nodes, "pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			}}})
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(b))
		inner(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func webhookController(t *testing.T) (*Controller, *stallClock, *webhookLinear) {
	t.Helper()
	f := &webhookLinear{}
	c, clk, dir := stallControllerYAML(t, linearWebhookYAML)
	f.ask = &askLinear{now: clk.now}
	srv := f.serve(t)
	for name, s := range map[string]string{"cid": "cid", "csecret": "cs", "whsecret": webhookSecret + "\n"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(s), 0o600))
	}
	cfg := *c.config()
	id := cfg.Identities["lapp"]
	id.TokenURL = srv.URL + "/token"
	cfg.Identities = map[string]config.Identity{"lapp": id}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter, c.linearURL = identity.NewMinter(), srv.URL
	sp, err := ingest.OpenSpool(t.Context(), c.config().Ingest.SpoolPath(c.root))
	require.NoError(t, err)
	testSpools.Store(c, sp)
	t.Cleanup(func() { testSpools.Delete(c); _ = sp.Close() })
	return c, clk, f
}

func webhookBody(t *testing.T, at time.Time, typ, actorType, actorID, org string) []byte {
	t.Helper()
	data := map[string]any{"identifier": "ENG-5", "number": 5, "team": map[string]any{"key": "ENG"}}
	if typ == "Comment" {
		data = map[string]any{"id": "c1", "userId": actorID, "issue": map[string]any{"identifier": "ENG-5"}}
	}
	b, err := json.Marshal(map[string]any{
		"action": "create", "type": typ, "organizationId": org, "webhookTimestamp": at.UnixMilli(),
		"actor": map[string]any{"id": actorID, "type": actorType}, "data": data,
	})
	require.NoError(t, err)
	return b
}

func sign(secret string, b []byte) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write(b)
	return hex.EncodeToString(m.Sum(nil))
}

// testSpools holds one open spool per controller under test; closed by
// webhookController's cleanup.
var testSpools sync.Map

func testSpool(c *Controller) *ingest.Spool {
	v, _ := testSpools.Load(c)
	return v.(*ingest.Spool)
}

// postWebhook sends b and waits for the async work.
func postWebhook(c *Controller, b []byte, sig string) int {
	req := httptest.NewRequest(http.MethodPost, config.LinearWebhookPath, bytes.NewReader(b))
	req.Header.Set("Linear-Signature", sig)
	rec := httptest.NewRecorder()
	c.linearWebhookHandler(testSpool(c)).Mux().ServeHTTP(rec, req)
	c.wg.Wait()
	return rec.Code
}

// TestLinearWebhookIssue (#300): a signed issue event from a human
// starts the intake run without waiting for the poll interval; bad
// signatures, stale timestamps, other orgs, bots and the App itself do
// nothing.
func TestLinearWebhookIssue(t *testing.T) {
	c, clk, f := webhookController(t)
	// First read is cached for the 1h poll: ENG-5 does not exist yet.
	c.runTriggers()
	require.Equal(t, 0, runCount(t, c))
	f.mu.Lock()
	f.listed = true
	f.mu.Unlock()
	c.runTriggers()
	require.Equal(t, 0, runCount(t, c), "cached read: polling alone waits for the interval")

	now := clk.now()
	good := webhookBody(t, now, "Issue", "user", "u-ann", "org-1")
	assert.Equal(t, http.StatusUnauthorized, postWebhook(c, good, sign("wrong", good)), "bad signature")
	assert.Equal(t, http.StatusUnauthorized, postWebhook(c, good, ""), "no signature")
	stale := webhookBody(t, now.Add(-2*time.Minute), "Issue", "user", "u-ann", "org-1")
	assert.Equal(t, http.StatusUnauthorized, postWebhook(c, stale, sign(webhookSecret, stale)), "stale")
	bot := webhookBody(t, now, "Issue", "app", "a-bot", "org-1")
	assert.Equal(t, http.StatusNoContent, postWebhook(c, bot, sign(webhookSecret, bot)), "bot filtered")
	for name, b := range map[string][]byte{
		"other org": webhookBody(t, now, "Issue", "user", "u-ann", "org-2"),
		"self":      webhookBody(t, now, "Issue", "user", "app-user", "org-1"),
	} {
		assert.Equal(t, http.StatusAccepted, postWebhook(c, b, sign(webhookSecret, b)), name)
		assert.Equal(t, 0, runCount(t, c), "%s ignored", name)
	}

	assert.Equal(t, http.StatusAccepted, postWebhook(c, good, sign(webhookSecret, good)))
	runs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindRun})
	require.NoError(t, err)
	require.Len(t, runs, 1, "the webhook started the intake run")
	assert.Contains(t, string(runs[0].Meta), "ENG-5")
}

// TestLinearWebhookComment (#300): a signed reply under a question's ask
// comment answers it right away, through the subscription path.
func TestLinearWebhookComment(t *testing.T) {
	c, clk, f := webhookController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"linear":{"workspace":"acme","issue":"ENG-5"},"github":{"repo":"acme/app"}}`)})
	_, _, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	out, err := c.report(ctx, scope{actor: sess.ID, session: sess.ID, project: tk.Project, pool: "dev"}, rpc.ReportParams{State: rpc.ReportWaiting, Question: "Keep it?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question
	c.sweepLinearAsks(ctx)
	f.ask.mu.Lock()
	askID, _ := f.ask.comments[0]["id"].(string)
	f.ask.mu.Unlock()
	// The pump registers the thread; then the reply arrives by webhook.
	c.subSource = muxSource{gh: &fakeThreads{}, linear: linearSource{c: c, url: c.linearURL}}
	c.pollSubscriptions(ctx, c.subSource)
	clk.advance(time.Minute)
	f.ask.reply(askID, "yes", "Wile", clk.now())

	b := webhookBody(t, clk.now(), "Comment", "user", "u-wile", "org-1")
	require.Equal(t, http.StatusAccepted, postWebhook(c, b, sign(webhookSecret, b)))
	rows, err := testSpool(c).List(ctx, 0)
	require.NoError(t, err)
	require.NotEmpty(t, rows, "spooled before processing")
	assert.Equal(t, "linear-lapp", rows[len(rows)-1].Endpoint, "implicit identity endpoint")
	got, err := c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, store.StatusClosed, got.Status)
	assert.Equal(t, resolutionAnswered, got.Resolution)
}

// TestLinearWebhookOff: no listen address, no server.
func TestLinearWebhookOff(t *testing.T) {
	c, _, _ := stallControllerYAML(t, strings.Replace(linearWebhookYAML, "webhook_listen: 127.0.0.1:0", `webhook_listen: ""`, 1))
	require.NoError(t, c.serveLinearWebhook())
	c.wg.Wait()
	assert.Empty(t, fmt.Sprint(c.config().Linear.WebhookListen))
}
