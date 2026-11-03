package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite"
)

const subYAML = `projects:
  js: {}
agents:
  fake: {command: /bin/false}
pools:
  dev: {agent: fake, dir: %q}
subscriptions:
  resume_window: 24h
  ignore: [ci-person]
`

// fakeThreads serves canned comments per thread key and records what
// each fetch asked for.
type fakeThreads struct {
	mu    sync.Mutex
	cs    map[string][]ghComment
	since map[string]time.Time
	pr    map[string]bool
}

func (f *fakeThreads) add(key string, cm ghComment) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cs[key] = append(f.cs[key], cm)
}

func (f *fakeThreads) Comments(_ context.Context, th thread, since time.Time) ([]ghComment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.since[th.key()] = since
	f.pr[th.key()] = th.PR
	var out []ghComment
	for _, c := range f.cs[th.key()] {
		if c.At.After(since) {
			out = append(out, c)
		}
	}
	return out, nil
}

func subController(t *testing.T) (*Controller, *stallClock, *fakeThreads) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, config.FileName), fmt.Appendf(nil, subYAML, dir), 0o600))
	root := config.Root{Dir: dir}
	cfg, err := config.Load(root)
	require.NoError(t, err)
	lc, err := newLiveConfig(cfg)
	require.NoError(t, err)
	live := new(atomic.Pointer[liveConfig])
	live.Store(lc)
	clk := &stallClock{t: time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC)}
	st, err := sqlite.Open(t.Context(), filepath.Join(dir, "db"), sqlite.Options{Now: clk.now, Prefix: func(p string) string { return p }})
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	c := &Controller{
		root: root, live: live, store: st, opt: Options{Now: clk.now}, log: slog.New(slog.DiscardHandler),
		sessions: map[string]context.CancelFunc{}, runners: map[string]*runner{},
	}
	c.ctx = t.Context()
	src := &fakeThreads{cs: map[string][]ghComment{}, since: map[string]time.Time{}, pr: map[string]bool{}}
	return c, clk, src
}

func threadsOf(t *testing.T, c *Controller, id string) []thread {
	t.Helper()
	tk, err := c.store.Get(t.Context(), id)
	require.NoError(t, err)
	sm, _ := decodeSubMeta(tk.Meta)
	return sm.Threads
}

func messagesTo(t *testing.T, c *Controller, to string) []store.Ticket {
	t.Helper()
	msgs, err := c.store.List(t.Context(), store.Filter{Kind: store.KindMessage})
	require.NoError(t, err)
	var out []store.Ticket
	for _, m := range msgs {
		f, err := store.DecodeFields[store.MessageFields](m)
		require.NoError(t, err)
		if f.To == to {
			out = append(out, m)
		}
	}
	return out
}

func mkSession(t *testing.T, c *Controller) store.Ticket {
	t.Helper()
	return mkTicket(t, c, store.NewTicket{Kind: store.KindSession, Title: "s", Pool: "dev", Fields: json.RawMessage(`{"agent":"fake"}`)})
}

// TestSubscriptionsWorkingSession: the issue and a PR linked in a comment
// are followed; a human comment reaches the holding session as an
// interrupt message and a ticket comment; bots and ignored logins do not;
// the read position is stored and used on the next pass.
func TestSubscriptionsWorkingSession(t *testing.T) {
	c, clk, src := subController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"github":{"repo":"acme/widget","issue":7}}`)})
	_, _, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	_, err = c.store.AddComment(ctx, sess.ID, tk.ID, "PR https://github.com/other/lib/pull/12 head abc")
	require.NoError(t, err)

	c.pollSubscriptions(ctx, src)
	ths := threadsOf(t, c, tk.ID)
	require.Len(t, ths, 2)
	assert.Equal(t, thread{Repo: "acme/widget", Number: 7, Seen: clk.now()}, ths[0])
	assert.Equal(t, thread{Repo: "other/lib", Number: 12, PR: true, Seen: clk.now()}, ths[1])
	assert.True(t, src.pr["other/lib#12"])

	at := clk.now().Add(time.Minute)
	src.add("other/lib#12", ghComment{Author: "github-actions[bot]", Body: "CI", At: at})
	src.add("other/lib#12", ghComment{Author: "ci-person", Body: "noise", At: at})
	src.add("other/lib#12", ghComment{Author: "renovate", Bot: true, Body: "bump", At: at})
	c.pollSubscriptions(ctx, src)
	assert.Empty(t, messagesTo(t, c, sess.ID), "bots and ignored logins are filtered")
	assert.Equal(t, at, threadsOf(t, c, tk.ID)[1].Seen, "read position advances past filtered comments")

	at2 := at.Add(time.Minute)
	src.add("other/lib#12", ghComment{Author: "alice", Kind: "review", State: "CHANGES_REQUESTED", Body: "rename foo", URL: "u", At: at2})
	c.pollSubscriptions(ctx, src)
	msgs := messagesTo(t, c, sess.ID)
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].Body, "rename foo")
	assert.Contains(t, msgs[0].Body, "review (changes_requested)")
	f, err := store.DecodeFields[store.MessageFields](msgs[0])
	require.NoError(t, err)
	assert.True(t, f.Interrupt)
	cs, err := c.store.Comments(ctx, tk.ID)
	require.NoError(t, err)
	assert.Equal(t, "github:alice", cs[len(cs)-1].Author)

	c.pollSubscriptions(ctx, src)
	assert.Len(t, messagesTo(t, c, sess.ID), 1, "delivered once")
	assert.Equal(t, at2, src.since["other/lib#12"])
}

// TestSubscriptionsResume: a human comment on a closed ticket's thread
// reopens it and assigns it to the session that last held it; once the
// ticket is closed longer than the resume window it is not followed.
func TestSubscriptionsResume(t *testing.T) {
	c, clk, src := subController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"github":{"repo":"acme/widget","pr":3}}`)})
	_, cl, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	c.pollSubscriptions(ctx, src)
	_, err = c.store.CloseTicket(ctx, sess.ID, tk.ID, &cl, "done")
	require.NoError(t, err)

	clk.advance(time.Hour)
	src.add("acme/widget#3", ghComment{Author: "bob", Body: "one more thing", At: clk.now()})
	c.pollSubscriptions(ctx, src)
	got, err := c.store.Get(ctx, tk.ID)
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, got.Status)
	assert.Equal(t, sess.ID, got.Holder, "same session resumes it")

	_, cl, err = c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	_, err = c.store.CloseTicket(ctx, sess.ID, tk.ID, &cl, "done")
	require.NoError(t, err)
	clk.advance(25 * time.Hour)
	src.add("acme/widget#3", ghComment{Author: "bob", Body: "late", At: clk.now()})
	c.pollSubscriptions(ctx, src)
	got, err = c.store.Get(ctx, tk.ID)
	require.NoError(t, err)
	assert.Equal(t, store.StatusClosed, got.Status, "outside the window")
}

// TestSubscribeMethod: a session adds a thread to its current ticket.
func TestSubscribeMethod(t *testing.T) {
	c, clk, src := subController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev"})
	_, _, err := c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	sc := scope{actor: sess.ID, session: sess.ID, project: "js", pool: "dev"}
	_, err = c.subscribe(ctx, sc, rpc.SubscribeParams{Repo: "bad", Number: 1})
	require.ErrorIs(t, err, store.ErrInvalid)
	_, err = c.subscribe(ctx, sc, rpc.SubscribeParams{Repo: "x/y", Number: 9})
	require.NoError(t, err)
	_, err = c.subscribe(ctx, sc, rpc.SubscribeParams{Repo: "x/y", Number: 9})
	require.NoError(t, err, "idempotent")
	assert.Equal(t, []thread{{Repo: "x/y", Number: 9, Seen: clk.now()}}, threadsOf(t, c, tk.ID))

	src.add("x/y#9", ghComment{Author: "carol", Body: "hi", At: clk.now().Add(time.Second)})
	c.pollSubscriptions(ctx, src)
	assert.Len(t, messagesTo(t, c, sess.ID), 1)
}

func TestParseGHLines(t *testing.T) {
	out, err := parseGHLines([]byte(`{"id":"r1","kind":"review","state":"COMMENTED","body":"","at":"2026-11-03T00:00:00Z","author":"a"}
{"id":"r2","kind":"review","state":"APPROVED","body":"","at":"2026-11-03T00:00:01Z","author":"a"}
{"id":"c1","kind":"comment","body":"x","at":"2026-11-03T00:00:02Z","author":"b","bot":true}
`))
	require.NoError(t, err)
	require.Len(t, out, 2, "empty COMMENTED review dropped")
	assert.Equal(t, "APPROVED", out[0].State)
	assert.True(t, out[1].Bot)
}

// TestSubscriptionsLinear: a ticket with meta.linear follows the Linear
// issue through the team's identity. Comments by the identity's own user
// (isMe) and bots advance the read position; a human comment reaches the
// working session.
func TestSubscriptionsLinear(t *testing.T) {
	c, clk, gh := subController(t)
	ctx := t.Context()
	tokFile := filepath.Join(t.TempDir(), "tok")
	require.NoError(t, os.WriteFile(tokFile, []byte("lin_api_x\n"), 0o600))
	cfg := *c.config()
	cfg.Identities = map[string]config.Identity{"lin": {Kind: config.IdentityLinear, TokenFile: tokFile, Workspace: "acme", Teams: []string{"ENG"}}}
	lc, err := newLiveConfig(&cfg)
	require.NoError(t, err)
	c.live.Store(lc)
	c.minter = identity.NewMinter()

	var mu sync.Mutex
	var nodes []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "lin_api_x", r.Header.Get("Authorization"))
		mu.Lock()
		defer mu.Unlock()
		// Linear lists newest first.
		rev := slices.Clone(nodes)
		slices.Reverse(rev)
		_, _ = fmt.Fprintf(w, `{"data":{"issue":{"id":"iss-1","identifier":"ENG-5","url":"https://linear.app/acme/issue/ENG-5","team":{"id":"t","key":"ENG"},"comments":{"nodes":[%s]}}}}`, strings.Join(rev, ","))
	}))
	t.Cleanup(srv.Close)
	addNode := func(body, user string, me bool, at time.Time) {
		mu.Lock()
		defer mu.Unlock()
		u := "null"
		if user != "" {
			u = fmt.Sprintf(`{"name":%q,"isMe":%t}`, user, me)
		}
		nodes = append(nodes, fmt.Sprintf(`{"id":"c%d","body":%q,"url":"","createdAt":%q,"user":%s}`, len(nodes), body, at.Format(time.RFC3339Nano), u))
	}
	src := muxSource{gh: gh, linear: linearSource{c: c, url: srv.URL}}

	sess := mkSession(t, c)
	tk := mkTicket(t, c, store.NewTicket{Title: "work", Pool: "dev", Meta: json.RawMessage(`{"linear":{"workspace":"acme","issue":"ENG-5"},"github":{"repo":"acme/app"}}`)})
	_, _, err = c.store.Claim(ctx, tk.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	addNode("old", "Ann", false, clk.now().Add(-time.Hour))

	c.pollSubscriptions(ctx, src)
	assert.Equal(t, []thread{{Linear: "ENG-5", Seen: clk.now()}}, threadsOf(t, c, tk.ID))
	assert.Empty(t, messagesTo(t, c, sess.ID), "older comments are not news")

	clk.advance(time.Minute)
	addNode("PR https://github.com/acme/app/pull/9", "jobshop", true, clk.now())
	addNode("synced", "", false, clk.now())
	c.pollSubscriptions(ctx, src)
	assert.Empty(t, messagesTo(t, c, sess.ID), "own and bot comments dropped")
	assert.Equal(t, clk.now(), threadsOf(t, c, tk.ID)[0].Seen)

	clk.advance(time.Minute)
	addNode("please also fix logout", "Ann", false, clk.now())
	c.pollSubscriptions(ctx, src)
	msgs := messagesTo(t, c, sess.ID)
	require.Len(t, msgs, 1)
	assert.Contains(t, msgs[0].Body, "New on Linear issue ENG-5")
	assert.Contains(t, msgs[0].Body, "please also fix logout")
	assert.Contains(t, msgs[0].Body, "https://linear.app/acme/issue/ENG-5")
	cs, err := c.store.Comments(ctx, tk.ID)
	require.NoError(t, err)
	assert.Equal(t, "linear:Ann", cs[len(cs)-1].Author)

	c.pollSubscriptions(ctx, src)
	assert.Len(t, messagesTo(t, c, sess.ID), 1, "delivered once")
}
