package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/identity"
	"github.com/gfx-labs/jobshop/pkg/orchestration/linearsvc"
	"github.com/gfx-labs/jobshop/pkg/orchestration/threads"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// askLinear is a fake Linear for ENG-5: comments are stored, labels kept
// by name, and team members resolvable for mentions.
type askLinear struct {
	mu       sync.Mutex
	now      func() time.Time
	comments []map[string]any // nodes as GetIssue returns them, oldest first
	labels   []string
	creates  int
}

func (f *askLinear) serve(t *testing.T) *httptest.Server {
	srv := httptest.NewServer(f.handler(t))
	t.Cleanup(srv.Close)
	return srv
}

func (f *askLinear) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var q struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		require.NoError(t, json.Unmarshal(b, &q))
		f.mu.Lock()
		defer f.mu.Unlock()
		out := func(v any) { _ = json.NewEncoder(w).Encode(map[string]any{"data": v}) }
		switch {
		case strings.Contains(q.Query, "commentCreate"):
			in, _ := q.Variables["input"].(map[string]any)
			f.creates++
			id, _ := in["id"].(string)
			f.comments = append(f.comments, map[string]any{"id": id, "body": in["body"], "url": "https://linear.app/c/" + id, "createdAt": f.now().Format(time.RFC3339Nano), "user": map[string]any{"name": "jobshop", "isMe": true}})
			out(map[string]any{"commentCreate": map[string]any{"success": true, "comment": map[string]any{"id": id, "url": "https://linear.app/c/" + id}}})
		case strings.Contains(q.Query, "team(id"):
			out(map[string]any{"team": map[string]any{
				"id": "t", "key": "ENG", "states": map[string]any{"nodes": []any{}}, "labels": map[string]any{"nodes": []any{}},
				"members": map[string]any{"nodes": []any{map[string]any{"id": "u1", "name": "Wile", "displayName": "wile", "email": "w@acme", "url": "https://linear.app/acme/profiles/wile"}}},
			}})
		case strings.Contains(q.Query, "issueLabels("):
			out(map[string]any{"issueLabels": map[string]any{"nodes": []any{map[string]any{"id": "lab-blocked", "team": nil}}}})
		case strings.Contains(q.Query, "issueAddLabel"):
			if !slices.Contains(f.labels, linearsvc.BlockedLabel) {
				f.labels = append(f.labels, linearsvc.BlockedLabel)
			}
			out(map[string]any{"issueAddLabel": map[string]any{"success": true}})
		case strings.Contains(q.Query, "issueRemoveLabel"):
			f.labels = slices.DeleteFunc(f.labels, func(s string) bool { return s == linearsvc.BlockedLabel })
			out(map[string]any{"issueRemoveLabel": map[string]any{"success": true}})
		case strings.Contains(q.Query, "labels { nodes { id name } }"):
			var ns []any
			for _, l := range f.labels {
				ns = append(ns, map[string]any{"id": "lab-blocked", "name": l})
			}
			out(map[string]any{"issue": map[string]any{"id": "iss-5", "labels": map[string]any{"nodes": ns}}})
		default:
			rev := slices.Clone(f.comments)
			slices.Reverse(rev)
			out(map[string]any{"issue": map[string]any{"id": "iss-5", "identifier": "ENG-5", "title": "Fix login", "url": "https://linear.app/acme/issue/ENG-5", "team": map[string]any{"id": "t", "key": "ENG"}, "comments": map[string]any{"nodes": rev}}})
		}
	}
}

func (f *askLinear) reply(parent, body, user string, at time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.comments = append(f.comments, map[string]any{"id": fmt.Sprintf("r%d", len(f.comments)), "body": body, "url": "", "createdAt": at.Format(time.RFC3339Nano), "parent": map[string]any{"id": parent}, "user": map[string]any{"name": user, "isMe": false}})
}

func (f *askLinear) state() (int, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.creates, slices.Clone(f.labels)
}

func askController(t *testing.T) (*Controller, *scanClock, *askLinear) {
	t.Helper()
	c, clk, _ := subController(t)
	tokFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokFile, []byte("lin_api_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"lin": {Kind: config.IdentityLinear, TokenFile: config.SecretPath(tokFile), Workspace: "acme", Teams: []string{"ENG"}}}
	cfg.Linear.AskOnIssue = true // opt-in since JOB-114
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter = identity.NewMinter()
	f := &askLinear{now: clk.now}
	c.linearURL = f.serve(t).URL
	return c, clk, f
}

// TestLinearQuestion: a worker question on a Linear run is posted once
// with the mention and Blocked; a reply under it answers the question,
// reaches the session and removes Blocked.
func TestLinearQuestion(t *testing.T) {
	t.Parallel()
	c, clk, f := askController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"linear":{"workspace":"acme","issue":"ENG-5","human":"wile"},"github":{"repo":"acme/app"}}`)})
	_, _, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	sc := scope{actor: sess.ID, session: sess.ID, project: tk.Project, pool: "dev"}
	out, err := c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportWaiting, Question: "Keep the old login page?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question

	c.sweepLinearAsks(ctx, c.config())
	c.sweepLinearAsks(ctx, c.config())
	n, labels := f.state()
	assert.Equal(t, 1, n, "posted once")
	assert.Equal(t, []string{linearsvc.BlockedLabel}, labels)
	f.mu.Lock()
	body, _ := f.comments[0]["body"].(string)
	askID, _ := f.comments[0]["id"].(string)
	f.mu.Unlock()
	assert.True(t, strings.HasPrefix(body, "https://linear.app/acme/profiles/wile "), body)
	assert.Contains(t, body, "Keep the old login page?")
	got, err := c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	a, ok := linearsvc.AskOf(got)
	require.True(t, ok)
	assert.Equal(t, askID, a.Comment)

	// The session's ticket follows ENG-5; the reply comes through
	// subscriptions.
	src := threads.Mux{GH: &fakeThreads{}, Linear: c.threadEngine().LinearSource()}
	c.pollSubscriptions(ctx, src)
	clk.advance(time.Minute)
	f.reply("other", "unrelated note", "Ann", clk.now())
	f.reply(askID, "yes, keep it", "Wile", clk.now())
	c.pollSubscriptions(ctx, src)

	got, err = c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusClosed, got.Status)
	assert.Equal(t, resolutionAnswered, got.Resolution)
	msgs := messagesTo(t, c, sess.ID)
	var answer, news bool
	for _, m := range msgs {
		answer = answer || (strings.Contains(m.Title, "Answer to your question") && m.Body == "yes, keep it")
		news = news || (strings.Contains(m.Body, "unrelated note") && !strings.Contains(m.Body, "yes, keep it"))
	}
	assert.True(t, answer, "the reply is the answer")
	assert.True(t, news, "other comments are delivered as usual, without the answer")
	_, labels = f.state()
	assert.Empty(t, labels, "Blocked removed on answer")
}

// TestLinearEscalation: an escalation ticket for a Linear run is posted
// on the issue without Blocked; a GitHub-only one is not posted.
func TestLinearEscalation(t *testing.T) {
	t.Parallel()
	c, _, f := askController(t)
	ctx := t.Context()
	esc := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindTask, Title: "Escalate run r1", Body: "Run r1 needs a human.\n\n## Step 0 implement: failed\n- long", Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"linear":{"issue":"ENG-5"},"run":{"id":"r1","escalation":"r1/1"}}`),
	})
	gh := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindTask, Title: "Escalate run r2", Body: "Run r2", Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"github":{"repo":"acme/app","issue":3},"run":{"id":"r2","escalation":"r2/1"}}`),
	})
	c.sweepLinearAsks(ctx, c.config())
	c.sweepLinearAsks(ctx, c.config())
	n, labels := f.state()
	assert.Equal(t, 1, n)
	assert.Empty(t, labels, "escalations do not block intake")
	f.mu.Lock()
	body, _ := f.comments[0]["body"].(string)
	f.mu.Unlock()
	assert.Contains(t, body, "Run r1 needs a human.")
	assert.NotContains(t, body, "## Step", "summary only")
	assert.Contains(t, body, esc.ID)
	got, err := c.store.Get(ctx, gh.ID)
	require.NoError(t, err)
	_, ok := linearsvc.AskOf(got)
	assert.False(t, ok, "non-Linear runs unchanged")
}

// TestLinearAskRetry: a crash after the post but before meta.linear_ask
// was saved replays the stored write instead of posting again.
func TestLinearAskRetry(t *testing.T) {
	t.Parallel()
	c, _, f := askController(t)
	ctx := t.Context()
	esc := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindTask, Title: "Escalate run r1", Body: "Run r1", Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"linear":{"issue":"ENG-5"},"run":{"id":"r1","escalation":"r1/1"}}`),
	})
	c.sweepLinearAsks(ctx, c.config())
	// The crash: meta.linear_ask never saved.
	got, err := c.store.Get(ctx, esc.ID)
	require.NoError(t, err)
	_, err = c.store.Update(ctx, "admin", esc.ID, store.Update{Meta: json.RawMessage(`{"linear":{"issue":"ENG-5"},"run":{"id":"r1","escalation":"r1/1"}}`), IfRevision: got.Revision})
	require.NoError(t, err)
	n, _ := f.state()
	require.Equal(t, 1, n)
	c.sweepLinearAsks(ctx, c.config())
	n, _ = f.state()
	assert.Equal(t, 1, n, "key done: stored result replayed")
	got, err = c.store.Get(ctx, esc.ID)
	require.NoError(t, err)
	_, ok := linearsvc.AskOf(got)
	assert.True(t, ok, "meta.linear_ask restored from the stored result")
}

// TestLinearAskOff: with linear.ask_on_issue off nothing new is posted
// or labelled, and a reply under an earlier ask comment is delivered as
// an ordinary comment instead of answering the question (JOB-114).
func TestLinearAskOff(t *testing.T) {
	t.Parallel()
	c, clk, f := askController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"linear":{"workspace":"acme","issue":"ENG-5","human":"wile"},"github":{"repo":"acme/app"}}`)})
	_, _, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	sc := scope{actor: sess.ID, session: sess.ID, project: tk.Project, pool: "dev"}
	out, err := c.report(ctx, sc, rpc.ReportParams{State: rpc.ReportWaiting, Question: "Keep the old login page?"})
	require.NoError(t, err)
	q := out.(rpc.ReportResult).Question
	// Posted while the switch was on.
	c.sweepLinearAsks(ctx, c.config())
	f.mu.Lock()
	askID, _ := f.comments[0]["id"].(string)
	f.mu.Unlock()

	cfg := *c.config()
	cfg.Linear.AskOnIssue = false // the default
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)

	esc := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindTask, Title: "Escalate run r1", Body: "Run r1 needs a human.", Labels: []string{needsHumanLabel},
		Meta: json.RawMessage(`{"linear":{"issue":"ENG-5"},"run":{"id":"r1","escalation":"r1/1"}}`),
	})
	c.sweepLinearAsks(ctx, c.config())
	n, _ := f.state()
	assert.Equal(t, 1, n, "nothing new posted")
	got, err := c.store.Get(ctx, esc.ID)
	require.NoError(t, err)
	_, ok := linearsvc.AskOf(got)
	assert.False(t, ok, "escalation stays on Desk")

	src := threads.Mux{GH: &fakeThreads{}, Linear: c.threadEngine().LinearSource()}
	c.pollSubscriptions(ctx, src)
	clk.advance(time.Minute)
	f.reply(askID, "yes, keep it", "Wile", clk.now())
	c.pollSubscriptions(ctx, src)

	got, err = c.store.Get(ctx, q.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, got.Status, "a Linear reply is not an answer")
	var news bool
	for _, m := range messagesTo(t, c, sess.ID) {
		assert.NotContains(t, m.Title, "Answer to your question")
		news = news || strings.Contains(m.Body, "yes, keep it")
	}
	assert.True(t, news, "the reply is delivered as a comment")
}
