package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/waitengine"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Recorded GitHub responses (trimmed) for the wait source tests.
const (
	recPR = `{"number":5,"state":"open","merged":false,"head":{"sha":"aaaa1111bbbb","ref":"issue-5"}}`
	// --slurp wraps the pages in an array.
	recReviews = `[[{"id":1,"user":{"login":"bob"},"state":"APPROVED","commit_id":"0000old","submitted_at":"2026-11-03T08:01:00Z"},
{"id":2,"user":{"login":"bob"},"state":"APPROVED","commit_id":"aaaa1111bbbb","submitted_at":"2026-11-03T08:05:00Z"}]]`
	recChecks    = `[{"name":"build","bucket":"pass"},{"name":"lint","bucket":"pending"}]`
	recCheckRuns = `{"total_count":2,"check_runs":[{"name":"build","status":"completed","conclusion":"success"},{"name":"nix","status":"completed","conclusion":"failure"}]}`
	// GET .../actions/runs?head_sha= on a master commit, ci queued.
	recWorkflowRuns = `{"total_count":2,"workflow_runs":[{"id":42172458357,"name":"ci","path":".github/workflows/ci.yml","status":"queued","conclusion":null,"event":"push"},{"id":42172458351,"name":"release","path":".github/workflows/release.yml","status":"completed","conclusion":"success","event":"push"}]}`
)

func TestWaitParseRecorded(t *testing.T) {
	t.Parallel()
	pr, err := github.ParsePR([]byte(recPR))
	require.NoError(t, err)
	assert.Equal(t, github.PRInfo{Head: "aaaa1111bbbb", State: "open"}, pr)
	rs, err := github.ParseReviews([]byte(recReviews))
	require.NoError(t, err)
	require.Len(t, rs, 2)
	assert.Equal(t, github.Review{Login: "bob", State: "APPROVED", Commit: "aaaa1111bbbb", At: time.Date(2026, 11, 3, 8, 5, 0, 0, time.UTC)}, rs[1])
	cs, err := github.ParsePRChecks([]byte(recChecks))
	require.NoError(t, err)
	assert.Equal(t, []github.Check{{Name: "build", Bucket: "pass"}, {Name: "lint", Bucket: "pending"}}, cs)
	cs, err = github.ParseCheckRuns([]byte(recCheckRuns))
	require.NoError(t, err)
	assert.Equal(t, []github.Check{{Name: "build", Bucket: "pass"}, {Name: "nix", Bucket: "fail"}}, cs)
	runs, err := github.ParseRunList([]byte(recWorkflowRuns))
	require.NoError(t, err)
	assert.Equal(t, []github.WorkflowRun{
		{ID: 42172458357, Name: "ci", Path: ".github/workflows/ci.yml", Status: "queued"},
		{ID: 42172458351, Name: "release", Path: ".github/workflows/release.yml", Status: "completed", Conclusion: "success"},
	}, runs)
}

// TestWaitChecksCancelledPR: a check cancelled on a PR head is not a
// failure either (JOB-75): the wait is invalidated so the agent decides
// again, instead of resuming as red.
func TestWaitChecksCancelledPR(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	park := func() domain.Ticket {
		tk := mkTicket(t, c, store.NewTicket{Title: "ci", Pool: "dev"})
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{
			ID: tk.ID, Reason: "CI", Any: true,
			Conds: []domain.WaitCond{
				{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"},
				{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "failure"},
			},
		})
		require.NoError(t, err)
		return tk
	}
	last := func(id string) string {
		cs, err := store.Comments(ctx, c.store, id)
		require.NoError(t, err)
		require.NotEmpty(t, cs)
		return cs[len(cs)-1].Body
	}

	a := park()
	prs.set(recPR, `[[]]`, `[{"name":"go","bucket":"cancel"},{"name":"nix","bucket":"pending"}]`)
	c.pollWaits(ctx, prs, threads)
	require.NotNil(t, waitOf(mustGet(t, c, a.ID)), "cancelled with another check pending: keep waiting")

	prs.set(recPR, `[[]]`, `[{"name":"go","bucket":"cancel"},{"name":"nix","bucket":"pass"},{"name":"ui","bucket":"skipping"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	assert.Contains(t, last(a.ID), "invalidated: checks cancelled on aaaa111: go")

	b := park()
	prs.set(recPR, `[[]]`, `[{"name":"go","bucket":"fail"},{"name":"nix","bucket":"cancel"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)))
	assert.Contains(t, last(b.ID), "condition met: checks failed on aaaa111: go")
}

// TestWaitChecksWorkflow: github.checks with workflow= follows that
// workflow's newest run on the commit (JOB-75), not the commit's checks,
// for a sha and for a PR head.
func TestWaitChecksWorkflow(t *testing.T) {
	t.Parallel()
	run := func(id int, name, file, status, conclusion string) string {
		conc := "null"
		if conclusion != "" {
			conc = `"` + conclusion + `"`
		}
		return fmt.Sprintf(`{"id":%d,"name":%q,"path":".github/workflows/%s","status":%q,"conclusion":%s}`, id, name, file, status, conc)
	}
	list := func(rs ...string) string {
		return `{"workflow_runs":[` + strings.Join(rs, ",") + `]}`
	}
	release := run(9, "release", "release.yml", "completed", "success")

	for _, tc := range []struct {
		name     string
		workflow string
		state    string
		runs     string
		met      bool
		invalid  bool
		note     string // in the wait's note or invalidation comment
	}{
		{"none yet", "ci", "success", list(release), false, false, "ci not started on aaaa111"},
		{"queued", "ci", "success", list(release, run(10, "ci", "ci.yml", "queued", "")), false, false, "ci in progress on aaaa111"},
		{"in progress", "ci", "failure", list(release, run(10, "ci", "ci.yml", "in_progress", "")), false, false, "ci in progress on aaaa111"},
		{"success", "ci", "success", list(release, run(10, "ci", "ci.yml", "completed", "success")), true, false, "ci green on aaaa111"},
		{"success, failure wait", "ci", "failure", list(run(10, "ci", "ci.yml", "completed", "success")), false, false, "ci green on aaaa111"},
		{"failure", "ci", "failure", list(release, run(10, "ci", "ci.yml", "completed", "failure")), true, false, "ci failed on aaaa111"},
		{"failure, success wait", "ci", "success", list(run(10, "ci", "ci.yml", "completed", "failure")), false, false, "ci failed on aaaa111"},
		{"cancelled", "ci", "success", list(release, run(10, "ci", "ci.yml", "completed", "cancelled")), false, true, "ci cancelled on aaaa111"},
		{"cancelled, failure wait", "ci", "failure", list(run(10, "ci", "ci.yml", "completed", "cancelled")), false, true, "ci cancelled on aaaa111"},
		{"newest run wins", "ci", "success", list(run(10, "ci", "ci.yml", "completed", "cancelled"), run(11, "ci", "ci.yml", "completed", "success")), true, false, "ci green on aaaa111"},
		{"rerun of a failed run", "ci", "success", list(run(12, "ci", "ci.yml", "in_progress", ""), run(11, "ci", "ci.yml", "completed", "failure")), false, false, "ci in progress on aaaa111"},
		{"matched by file", "ci.yml", "success", list(run(10, "ci", "ci.yml", "completed", "success")), true, false, "ci.yml green on aaaa111"},
		{"matched by path", ".github/workflows/ci.yml", "success", list(run(10, "ci", "ci.yml", "completed", "success")), true, false, ".github/workflows/ci.yml green on aaaa111"},
		{"name case-insensitive", "ci", "success", list(release, run(10, "CI", "build.yml", "completed", "success")), true, false, "ci green on aaaa111"},
		{"matched by file stem", "ci", "success", list(release, run(10, "Build and test", "ci.yml", "completed", "success")), true, false, "ci green on aaaa111"},
		{"file case-insensitive", "CI.yml", "success", list(run(10, "Build", "ci.yml", "completed", "success")), true, false, "CI.yml green on aaaa111"},
		{"other workflow does not match", "ci", "success", list(release, run(10, "lint", "lint.yml", "completed", "success")), false, false, "ci not started on aaaa111"},
	} {
		for _, form := range []string{"sha", "pr"} {
			t.Run(tc.name+"/"+form, func(t *testing.T) {
				t.Parallel()
				c, _, prs, threads := waitController(t)
				ctx := t.Context()
				cond := domain.WaitCond{Type: domain.WaitChecks, Repo: "acme/widget", State: tc.state, Workflow: tc.workflow}
				head := "aaaa1111"
				if form == "pr" {
					head = "aaaa1111bbbb"
					cond.PR = 5
					prs.set(recPR, `[[]]`, `[{"name":"go","bucket":"pass"}]`)
				} else {
					cond.SHA = head
				}
				prs.runs[head] = tc.runs
				tk := mkTicket(t, c, store.NewTicket{Title: "ci", Pool: "dev"})
				_, err := c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{ID: tk.ID, Reason: "CI", Conds: []domain.WaitCond{cond}})
				require.NoError(t, err)
				c.pollWaits(ctx, prs, threads)
				w := waitOf(mustGet(t, c, tk.ID))
				if tc.met || tc.invalid {
					assert.Nil(t, w, "wait ended")
					cs, err := store.Comments(ctx, c.store, tk.ID)
					require.NoError(t, err)
					require.NotEmpty(t, cs)
					body := cs[len(cs)-1].Body
					assert.Contains(t, body, tc.note)
					if tc.invalid {
						assert.Contains(t, body, "invalidated:")
					} else {
						assert.Contains(t, body, "condition met:")
					}
					return
				}
				require.NotNil(t, w, "wait keeps waiting")
				assert.False(t, w.Conds[0].Met)
				assert.Contains(t, w.Conds[0].Note, tc.note)
			})
		}
	}
}

// TestWaitChecksCommit: a checks wait on a commit (#831). A cancelled
// run (superseded on master) is not a failure, and invalidates the wait
// once nothing else is pending; skipped checks alone are not green; a
// queued workflow run keeps it pending.
func TestWaitChecksCommit(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	park := func() domain.Ticket {
		tk := mkTicket(t, c, store.NewTicket{Title: "post-merge", Pool: "dev"})
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{
			ID: tk.ID, Reason: "master CI", Any: true,
			Conds: []domain.WaitCond{
				{Type: domain.WaitChecks, Repo: "acme/widget", SHA: "aaaa1111", State: "success"},
				{Type: domain.WaitChecks, Repo: "acme/widget", SHA: "aaaa1111", State: "failure"},
			},
		})
		require.NoError(t, err)
		return tk
	}
	set := func(cs string) {
		prs.mu.Lock()
		defer prs.mu.Unlock()
		prs.checks["acme/widget#0"] = cs
	}
	lastComment := func(id string) string {
		cs, err := store.Comments(ctx, c.store, id)
		require.NoError(t, err)
		require.NotEmpty(t, cs)
		return cs[len(cs)-1].Body
	}

	a := park()
	// Only release's skipped jobs, ci queued: not green.
	set(`[{"name":"nix","bucket":"skipping"},{"name":"upload","bucket":"skipping"},{"name":"ci (workflow)","bucket":"pending"}]`)
	c.pollWaits(ctx, prs, threads)
	require.NotNil(t, waitOf(mustGet(t, c, a.ID)))
	set(`[{"name":"nix","bucket":"skipping"},{"name":"upload","bucket":"skipping"}]`)
	c.pollWaits(ctx, prs, threads)
	require.NotNil(t, waitOf(mustGet(t, c, a.ID)), "skipped checks alone are not green")

	// go cancelled while nix still runs: neither failed nor green yet.
	set(`[{"name":"go","bucket":"cancel"},{"name":"nix","bucket":"pending"},{"name":"release-please","bucket":"pass"}]`)
	c.pollWaits(ctx, prs, threads)
	require.NotNil(t, waitOf(mustGet(t, c, a.ID)), "a cancelled check is not a failure")

	// Everything else finished: the commit will not go green.
	set(`[{"name":"go","bucket":"cancel"},{"name":"nix","bucket":"pass"},{"name":"ci (workflow)","bucket":"cancel"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	assert.Contains(t, lastComment(a.ID), "invalidated: checks cancelled on aaaa111: go, ci (workflow)")

	// A real failure still fires the failure condition.
	b := park()
	set(`[{"name":"go","bucket":"fail"},{"name":"nix","bucket":"cancel"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)))
	assert.Contains(t, lastComment(b.ID), "condition met: checks failed on aaaa111: go")

	// Passed and skipped: green.
	d := park()
	set(`[{"name":"go","bucket":"pass"},{"name":"upload","bucket":"skipping"},{"name":"ci (workflow)","bucket":"pass"}]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, d.ID)))
	assert.Contains(t, lastComment(d.ID), "condition met: checks green on aaaa111")
}

// fakePRs serves recorded responses per PR, replaced as a test goes on.
type fakePRs struct {
	mu      sync.Mutex
	pr      map[string]string
	reviews map[string]string
	checks  map[string]string
	issues  map[string]string
	// runs holds the workflow runs per commit, as the actions/runs JSON.
	runs  map[string]string
	calls int
}

func (f *fakePRs) set(pr, reviews, checks string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pr["acme/widget#5"], f.reviews["acme/widget#5"], f.checks["acme/widget#5"] = pr, reviews, checks
}

func (f *fakePRs) PR(_ context.Context, repo string, n int) (github.PRInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return github.ParsePR([]byte(f.pr[waitengine.PRKey(repo, n)]))
}

func (f *fakePRs) Reviews(_ context.Context, repo string, n int) ([]github.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return github.ParseReviews([]byte(f.reviews[waitengine.PRKey(repo, n)]))
}

func (f *fakePRs) Checks(_ context.Context, repo string, n int, _ string, _ bool) ([]github.Check, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return github.ParsePRChecks([]byte(f.checks[waitengine.PRKey(repo, n)]))
}

func (f *fakePRs) Issue(_ context.Context, repo string, n int) (github.IssueInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return github.ParseIssue([]byte(f.issues[waitengine.PRKey(repo, n)]))
}

// Tag: issues["tag:<repo>@<tag>"] set means the tag exists.
func (f *fakePRs) Tag(_ context.Context, repo, tag string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, ok := f.issues["tag:"+repo+"@"+tag]
	return ok, nil
}

// Contains: issues["compare:<sha>...<head>"] holds the compare status.
func (f *fakePRs) Contains(_ context.Context, _ string, sha, head string) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return github.CompareContains(f.issues["compare:"+sha+"..."+head]), nil
}

func (f *fakePRs) WorkflowRuns(_ context.Context, _, sha string) ([]github.WorkflowRun, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.runs[sha] == "" {
		return nil, nil
	}
	return github.ParseRunList([]byte(f.runs[sha]))
}

func waitController(t *testing.T) (*Controller, *scanClock, *fakePRs, *fakeThreads) {
	t.Helper()
	c, clk, threads := subController(t)
	return c, clk, &fakePRs{pr: map[string]string{}, reviews: map[string]string{}, checks: map[string]string{}, issues: map[string]string{}, runs: map[string]string{}}, threads
}

func sessScope(s domain.Ticket) scope {
	return scope{actor: s.ID, session: s.ID, project: s.Project, pool: s.Pool}
}

func readyIDs(t *testing.T, c *Controller) []string {
	t.Helper()
	r, err := c.store.List(t.Context(), store.Filter{Project: "js", Ready: true})
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
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	step := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindStep, Title: "review", Pool: "dev", ParentID: run.ID,
		Fields: json.RawMessage(`{"run":"` + run.ID + `","step":"review","index":0,"retries":3,"failures":0}`),
	})
	_, _, err := c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)

	prs.set(recPR, `[[]]`, recChecks)
	_, err = c.waitAdd(ctx, sessScope(sess), rpc.WaitAddParams{
		Conds: []domain.WaitCond{
			{Type: domain.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob"},
			{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success", Required: true},
		},
		Reason: "needs bob's approval", Owner: "bob", Resume: true,
	})
	require.NoError(t, err)
	got, err := c.store.Get(ctx, step.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusOpen, got.Status)
	sf, err := domain.DecodeFields[domain.StepFields](got)
	require.NoError(t, err)
	assert.Equal(t, 0, sf.Failures, "parking is not a failed attempt")
	assert.Empty(t, readyIDs(t, c))
	_, _, err = c.store.ClaimNext(ctx, store.ClaimNext{Project: "js", Pool: "dev", Holder: sess.ID, Lease: time.Hour})
	require.ErrorIs(t, err, domain.ErrNoWork, "a waiting ticket cannot be claimed")

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
	cs, err := store.Comments(ctx, c.store, step.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "wants APPROVED by bob on aaaa111, observed APPROVED")
}

func mustGet(t *testing.T, c *Controller, id string) domain.Ticket {
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
		Conds: []domain.WaitCond{{Type: domain.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob", Head: "aaaa1111bbbb"}},
	})
	require.NoError(t, err)
	_, err = c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: b.ID, Reason: "merge",
		Conds: []domain.WaitCond{{Type: domain.WaitPR, Repo: "acme/widget", PR: 5, State: "merged"}},
	})
	require.NoError(t, err)
	c.pollWaits(ctx, prs, threads)
	assert.Equal(t, 1, prs.calls, "one PR fetch per pass")
	assert.Empty(t, readyIDs(t, c))

	prs.set(`{"state":"open","merged":false,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	cs, err := store.Comments(ctx, c.store, a.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "invalidated: acme/widget#5 head moved aaaa111 -> cccc222")
	assert.NotNil(t, waitOf(mustGet(t, c, b.ID)))

	prs.set(`{"state":"closed","merged":true,"head":{"sha":"cccc2222"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)))
	assert.ElementsMatch(t, []string{a.ID, b.ID}, readyIDs(t, c))
}

// TestWaitChecksConflicting: GitHub runs no pull_request CI on a PR that
// conflicts with its base, so a checks wait on it is invalidated rather
// than left to time out (#609). mergeable null (still computing) waits.
func TestWaitChecksConflicting(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	prs.set(`{"state":"open","merged":false,"mergeable":null,"head":{"sha":"aaaa1111bbbb"}}`, `[[]]`, `[]`)
	_, err := c.waitAdd(ctx, scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: a.ID, Reason: "ci",
		Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"}},
	})
	require.NoError(t, err)
	c.pollWaits(ctx, prs, threads)
	assert.NotNil(t, waitOf(mustGet(t, c, a.ID)))

	prs.set(`{"state":"open","merged":false,"mergeable":false,"mergeable_state":"dirty","head":{"sha":"aaaa1111bbbb"}}`, `[[]]`, `[]`)
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, a.ID)))
	cs, err := store.Comments(ctx, c.store, a.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "invalidated: acme/widget#5 conflicts with its base")
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
		Conds: []domain.WaitCond{{Type: domain.WaitReply, Thread: "acme/widget#7", From: "alice"}},
	})
	require.NoError(t, err)
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	_, err = c.waitAdd(ctx, admin, rpc.WaitAddParams{
		ID: b.ID, Reason: "recheck tomorrow",
		Conds: []domain.WaitCond{{Type: domain.WaitTime, At: clk.now().Add(2 * time.Hour)}},
	})
	require.NoError(t, err)

	c.pollWaits(ctx, prs, threads)
	items, err := c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	assert.Empty(t, items)

	clk.advance(90 * time.Minute)
	threads.add("acme/widget#7", github.Comment{Author: "mallory", Body: "not alice", At: clk.now()})
	c.pollWaits(ctx, prs, threads)
	c.pollWaits(ctx, prs, threads)
	got := mustGet(t, c, a.ID)
	assert.Contains(t, got.Labels, needsHumanLabel)
	assert.True(t, waitOf(got).TimedOut)
	cs, err := store.Comments(ctx, c.store, a.ID)
	require.NoError(t, err)
	assert.Len(t, cs, 1, "timeout noted once")
	items, err = c.attentionItems(ctx, c.config())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, rpc.AttentionWait, items[0].Kind)
	assert.Equal(t, "alice", items[0].Owner)
	assert.Contains(t, items[0].Title, "reply from alice")

	clk.advance(time.Hour)
	threads.add("acme/widget#7", github.Comment{Author: "Alice", Body: "go ahead", At: clk.now()})
	c.pollWaits(ctx, prs, threads)
	got = mustGet(t, c, a.ID)
	assert.Nil(t, waitOf(got))
	assert.NotContains(t, got.Labels, needsHumanLabel)
	assert.Nil(t, waitOf(mustGet(t, c, b.ID)), "time passed")

	list, err := c.waitList(ctx, admin, rpc.WaitListParams{})
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestWaitValidate(t *testing.T) {
	t.Parallel()
	c, _, _, _ := waitController(t)
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	admin := scope{admin: true, actor: "admin"}
	for _, conds := range [][]domain.WaitCond{
		nil,
		{{Type: "nope"}},
		{{Type: domain.WaitReview, Repo: "acme/widget", PR: 5}},
		{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "green"}},
		{{Type: domain.WaitReview, Repo: "acme/widget", PR: 5, Login: "bob", Workflow: "ci"}},
		{{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success", Workflow: "ci", Required: true}},
		{{Type: domain.WaitReply, Thread: "bad"}},
	} {
		_, err := c.waitAdd(t.Context(), admin, rpc.WaitAddParams{ID: a.ID, Reason: "x", Conds: conds})
		require.ErrorIs(t, err, domain.ErrInvalid, "%v", conds)
	}
	_, err := c.waitAdd(t.Context(), admin, rpc.WaitAddParams{ID: a.ID, Conds: []domain.WaitCond{{Type: domain.WaitTime, At: time.Now()}}})
	require.ErrorIs(t, err, domain.ErrInvalid, "reason is required")
}
