package control

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Recorded GitHub responses (trimmed) for the wait source tests.
const (
	recPR = `{"number":5,"state":"open","merged":false,"head":{"sha":"aaaa1111bbbb","ref":"issue-5"}}`
	// --slurp wraps the pages in an array.
	recReviews = `[[{"id":1,"user":{"login":"bob"},"state":"APPROVED","commit_id":"0000old","submitted_at":"2026-11-03T08:01:00Z"},
{"id":2,"user":{"login":"bob"},"state":"APPROVED","commit_id":"aaaa1111bbbb","submitted_at":"2026-11-03T08:05:00Z"}]]`
	recChecks    = `[{"name":"build","bucket":"pass"},{"name":"lint","bucket":"pending"}]`
	recCheckRuns = `{"total_count":2,"check_runs":[{"name":"build","status":"completed","conclusion":"success"},{"name":"nix","status":"completed","conclusion":"failure"}]}`
)

func TestWaitParseRecorded(t *testing.T) {
	t.Parallel()
	pr, err := parsePR([]byte(recPR))
	require.NoError(t, err)
	assert.Equal(t, prInfo{Head: "aaaa1111bbbb", State: "open"}, pr)
	rs, err := parseReviews([]byte(recReviews))
	require.NoError(t, err)
	require.Len(t, rs, 2)
	assert.Equal(t, prReview{Login: "bob", State: "APPROVED", Commit: "aaaa1111bbbb", At: time.Date(2026, 11, 3, 8, 5, 0, 0, time.UTC)}, rs[1])
	cs, err := parsePRChecks([]byte(recChecks))
	require.NoError(t, err)
	assert.Equal(t, []prCheck{{"build", "pass"}, {"lint", "pending"}}, cs)
	cs, err = parseCheckRuns([]byte(recCheckRuns))
	require.NoError(t, err)
	assert.Equal(t, []prCheck{{"build", "pass"}, {"nix", "fail"}}, cs)
}

// fakePRs serves recorded responses per PR, replaced as a test goes on.
type fakePRs struct {
	mu      sync.Mutex
	pr      map[string]string
	reviews map[string]string
	checks  map[string]string
	issues  map[string]string
	calls   int
}

func (f *fakePRs) set(pr, reviews, checks string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pr["acme/widget#5"], f.reviews["acme/widget#5"], f.checks["acme/widget#5"] = pr, reviews, checks
}

func (f *fakePRs) PR(_ context.Context, repo string, n int) (prInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return parsePR([]byte(f.pr[prKey(repo, n)]))
}

func (f *fakePRs) Reviews(_ context.Context, repo string, n int) ([]prReview, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return parseReviews([]byte(f.reviews[prKey(repo, n)]))
}

func (f *fakePRs) Checks(_ context.Context, repo string, n int, _ string, _ bool) ([]prCheck, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return parsePRChecks([]byte(f.checks[prKey(repo, n)]))
}

func (f *fakePRs) Issue(_ context.Context, repo string, n int) (issueInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return parseIssue([]byte(f.issues[prKey(repo, n)]))
}

func waitController(t *testing.T) (*Controller, *stallClock, *fakePRs, *fakeThreads) {
	t.Helper()
	c, clk, threads := subController(t)
	return c, clk, &fakePRs{pr: map[string]string{}, reviews: map[string]string{}, checks: map[string]string{}, issues: map[string]string{}}, threads
}

func sessScope(s store.Ticket) scope {
	return scope{actor: s.ID, session: s.ID, project: s.Project, pool: s.Pool}
}

func readyIDs(t *testing.T, c *Controller) []string {
	t.Helper()
	r, err := c.store.Ready(t.Context(), "js", "", 0)
	require.NoError(t, err)
	var ids []string
	for _, x := range r {
		ids = append(ids, x.ID)
	}
	return ids
}

// TestWaitApprovalAndChecks: a held step waits on an exact-head approval
// and green required checks. It is released without a failed attempt,
// stays out of the frontier while either is missing, and becomes ready,
// pre-assigned to the session, once both hold.
func TestWaitApprovalAndChecks(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	run := mkTicket(t, c, store.NewTicket{Kind: store.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step := mkTicket(t, c, store.NewTicket{
		Kind: store.KindStep, Title: "review", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"review","index":0,"retries":3,"failures":0}`),
	})
	_, _, err := c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)

	prs.set(recPR, `[[]]`, recChecks)
	_, err = c.waitAdd(ctx, sessScope(sess), rpc.WaitAddParams{
		Conds: []store.WaitCond{
			{Type: store.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob"},
			{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "success", Required: true},
		},
		Reason: "needs bob's approval", Owner: "bob", Resume: true,
	})
	require.NoError(t, err)
	got, err := c.store.Get(ctx, step.ID)
	require.NoError(t, err)
	assert.Equal(t, store.StatusOpen, got.Status)
	sf, err := store.DecodeFields[store.StepFields](got)
	require.NoError(t, err)
	assert.Equal(t, 0, sf.Failures, "parking is not a failed attempt")
	assert.Empty(t, readyIDs(t, c))
	_, _, err = c.store.ClaimNext(ctx, store.ClaimNext{Project: "js", Pool: "dev", Holder: sess.ID, Lease: time.Hour})
	require.ErrorIs(t, err, store.ErrNoWork, "a waiting ticket cannot be claimed")

	c.pollWaits(ctx, prs, threads)
	w := waitOf(mustGet(t, c, step.ID))
	require.NotNil(t, w)
	assert.Equal(t, "aaaa1111bbbb", w.Conds[0].Head, "head recorded on first evaluation")

	// Approval on the exact head; checks still pending.
	clk.advance(time.Minute)
	prs.set(recPR, recReviews, recChecks)
	c.pollWaits(ctx, prs, threads)
	w = waitOf(mustGet(t, c, step.ID))
	require.NotNil(t, w)
	assert.True(t, w.Conds[0].Met)
	assert.False(t, w.Conds[1].Met)

	prs.set(recPR, recReviews, `[{"name":"build","bucket":"pass"},{"name":"lint","bucket":"pass"}]`)
	c.pollWaits(ctx, prs, threads)
	got = mustGet(t, c, step.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder, "resume pre-assigns the parking session")
	assert.Equal(t, []string{step.ID}, readyIDs(t, c))
	cs, err := c.store.Comments(ctx, step.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "approved by bob on aaaa111")
}

func mustGet(t *testing.T, c *Controller, id string) store.Ticket {
	t.Helper()
	tk, err := c.store.Get(t.Context(), id)
	require.NoError(t, err)
	return tk
}

// TestWaitHeadMoved: an approval wait on a head is invalidated when the
// PR head moves, and a merged wait resolves on merge.
func TestWaitHeadMoved(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	prs.set(recPR, `[[]]`, `[]`)
	_, err := c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: a.ID, Reason: "approval",
		Conds: []store.WaitCond{{Type: store.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob", Head: "aaaa1111bbbb"}},
	})
	require.NoError(t, err)
	_, err = c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: b.ID, Reason: "merge",
		Conds: []store.WaitCond{{Type: store.WaitPR, Repo: "acme/widget", PR: 5, State: "merged"}},
	})
	require.NoError(t, err)
	c.pollWaits(ctx, prs, threads)
	assert.Equal(t, 1, prs.calls, "one PR fetch per pass")
	assert.Empty(t, readyIDs(t, c))

	prs.set(`{"state":"open","merged":false,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	cs, err := c.store.Comments(ctx, a.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "invalidated: acme/widget#5 head moved aaaa111 -> cccc222")
	assert.NotNil(t, waitOf(mustGet(t, c, b.ID)))

	prs.set(`{"state":"closed","merged":true,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)))
	assert.ElementsMatch(t, []string{a.ID, b.ID}, readyIDs(t, c))
}

// TestWaitTimeoutAttention: past its timeout a wait is labeled
// needs-human once and listed in attention with reason and owner; a
// time condition and a reply clear waits.
func TestWaitTimeoutAttention(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	timeout := clk.now().Add(time.Hour)
	_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: a.ID, Reason: "reply from alice", Owner: "alice", Timeout: &timeout,
		Conds: []store.WaitCond{{Type: store.WaitReply, Thread: "acme/widget#7", From: "alice"}},
	})
	require.NoError(t, err)
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	_, err = c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: b.ID, Reason: "recheck tomorrow",
		Conds: []store.WaitCond{{Type: store.WaitTime, At: clk.now().Add(2 * time.Hour)}},
	})
	require.NoError(t, err)

	c.pollWaits(ctx, prs, threads)
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	clk.advance(90 * time.Minute)
	threads.add("acme/widget#7", ghComment{Author: "mallory", Body: "not alice", At: clk.now()})
	c.pollWaits(ctx, prs, threads)
	c.pollWaits(ctx, prs, threads)
	got := mustGet(t, c, a.ID)
	assert.Contains(t, got.Labels, needsHumanLabel)
	assert.True(t, waitOf(got).TimedOut)
	cs, err := c.store.Comments(ctx, a.ID)
	require.NoError(t, err)
	assert.Len(t, cs, 1, "timeout noted once")
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionWait, items[0].Kind)
	assert.Equal(t, "alice", items[0].Owner)
	assert.Contains(t, items[0].Title, "reply from alice")

	clk.advance(time.Hour)
	threads.add("acme/widget#7", ghComment{Author: "Alice", Body: "go ahead", At: clk.now()})
	c.pollWaits(ctx, prs, threads)
	got = mustGet(t, c, a.ID)
	assert.Nil(t, waitOf(got))
	assert.NotContains(t, got.Labels, needsHumanLabel)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)), "time passed")

	list, err := c.waitList(ctx, admin, rpc.WaitListParams{})
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestWaitConflictingPR: a checks wait on a PR GitHub reports as dirty
// (conflicting) is invalidated in one pass with the conflict reason;
// unknown and clean with pending checks keep waiting (#609).
func TestWaitConflictingPR(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	sess := mkSession(t, c)
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	_, _, err := c.store.Claim(ctx, a.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	pr := func(state string) string {
		return `{"number":5,"state":"open","merged":false,"mergeable":false,"mergeable_state":"` + state + `","head":{"sha":"aaaa1111bbbb"},"base":{"ref":"master"}}`
	}
	prs.set(pr("unknown"), `[[]]`, `[]`)
	_, err = c.waitAdd(ctx, sessScope(sess), rpc.WaitAddParams{
		Reason: "CI on PR #5", Resume: true, Any: true,
		Conds: []store.WaitCond{
			{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"},
			{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "failure"},
		},
	})
	require.NoError(t, err)
	c.pollWaits(ctx, prs, threads)
	assert.NotNil(t, waitOf(mustGet(t, c, a.ID)), "unknown is retried")

	prs.set(pr("clean"), `[[]]`, recChecks)
	c.pollWaits(ctx, prs, threads)
	assert.NotNil(t, waitOf(mustGet(t, c, a.ID)), "clean with pending checks waits")

	prs.set(pr("dirty"), `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	got := mustGet(t, c, a.ID)
	assert.Nil(t, waitOf(got))
	assert.Equal(t, sess.ID, got.Holder, "resumed")
	cs, err := c.store.Comments(ctx, a.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "PR #5 conflicts with master; CI will not run")
}

// TestWaitCINotStarted: a repo+sha checks wait with no check after the
// grace period shows in attention; one check seen keeps it out (#609).
func TestWaitCINotStarted(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	for _, x := range []store.Ticket{a, b} {
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{
			ID: x.ID, Reason: "master CI",
			Conds: []store.WaitCond{{Type: store.WaitChecks, Repo: "acme/widget", SHA: "dddd3333eeee" + x.ID, State: "success"}},
		})
		require.NoError(t, err)
	}
	prs.checks["acme/widget#0"] = `[]`
	c.pollWaits(ctx, prs, threads)
	clk.advance(10 * time.Minute)
	items, err := c.attentionItems(ctx)
	require.NoError(t, err)
	assert.Empty(t, items, "within the grace period")

	// b's sha starts CI; a's never does.
	w := waitOf(mustGet(t, c, b.ID))
	w.Conds[0].Started = true
	fields, err := withWait(mustGet(t, c, b.ID), w)
	require.NoError(t, err)
	_, err = c.store.Update(ctx, "controller", b.ID, store.Update{Fields: fields})
	require.NoError(t, err)

	clk.advance(10 * time.Minute)
	items, err = c.attentionItems(ctx)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, a.ID, items[0].ID)
	assert.Contains(t, items[0].Title, "CI did not start for dddd333")
}

func TestWaitValidate(t *testing.T) {
	t.Parallel()
	c, _, _, _ := waitController(t)
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	admin := scope{admin: true, actor: "admin"}
	for _, conds := range [][]store.WaitCond{
		nil,
		{{Type: "nope"}},
		{{Type: store.WaitReview, Repo: "acme/widget", PR: 5}},
		{{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: "green"}},
		{{Type: store.WaitReply, Thread: "bad"}},
	} {
		_, err := c.waitAdd(t.Context(), admin, rpc.WaitAddParams{ID: a.ID, Reason: "x", Conds: conds})
		require.ErrorIs(t, err, store.ErrInvalid, "%v", conds)
	}
	_, err := c.waitAdd(t.Context(), admin, rpc.WaitAddParams{ID: a.ID, Conds: []store.WaitCond{{Type: store.WaitTime, At: time.Now()}}})
	require.ErrorIs(t, err, store.ErrInvalid, "reason is required")
}
