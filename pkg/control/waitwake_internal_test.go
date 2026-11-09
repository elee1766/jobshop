package control

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/waitengine"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// TestWaitWebhookResume (#679): a CI completion webhook resumes a
// --resume waiting step through the one ingest path (receiver, spool,
// drain, inbox, targeted wait wake) with the real pumps running and the
// fake clock never advanced, so no timer is involved.
func TestWaitWebhookResume(t *testing.T) {
	t.Parallel()
	c, clk, prs, threads := waitController(t)
	c.waitSource, c.subSource = prs, threads
	c.waitWake = make(chan struct{}, 1)
	ctx := t.Context()
	sess := mkSession(t, c)
	step := mkTicket(t, c, store.NewTicket{Title: "implement", Pool: "dev"})
	_, _, err := c.store.Claim(ctx, step.ID, sess.ID, time.Hour)
	require.NoError(t, err)
	prs.set(recPR, `[[]]`, recChecks)
	_, err = c.waitAdd(ctx, sessScope(sess), rpc.WaitAddParams{
		Conds: []domain.WaitCond{
			{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "success"},
			{Type: domain.WaitChecks, Repo: "acme/widget", PR: 5, State: "failure"},
		},
		Any: true, Resume: true, Reason: "CI on PR #5",
	})
	require.NoError(t, err)
	// The pump's first full pass is the startup reconcile; drop the wake
	// waitAdd queued so only the webhook can wake it below.
	c.waitWakes.Take()
	select {
	case <-c.waitWake:
	default:
	}

	pctx, cancel := context.WithCancel(ctx)
	c.ctx = pctx
	t.Cleanup(func() { cancel(); c.wg.Wait(); c.queue().CloseSpool() })
	c.wg.Go(c.pumpWaits)
	c.wg.Go(c.pumpDrain)
	// Both pumps parked on their timers: the first full pass is done.
	require.True(t, clk.WaitFor(2, ctx.Done()))
	require.NotNil(t, waitOf(mustGet(t, c, step.ID)), "checks still pending")

	prs.set(recPR, `[[]]`, `[{"name":"build","bucket":"pass"},{"name":"lint","bucket":"pass"}]`)
	sp, err := c.queue().Spool(ctx)
	require.NoError(t, err)
	ep := config.IngestEndpoint{Name: "gh", Source: config.IngestGitHub}
	h := &ingest.Handler{
		Spool: sp, Endpoints: []config.IngestEndpoint{ep},
		Secret:  func(config.IngestEndpoint) (ingest.Secret, error) { return ingest.Secret{Key: []byte("s")}, nil },
		Now:     c.now,
		Log:     slog.New(slog.DiscardHandler),
		OnSpool: func(ingest.Delivery) { c.queue().WakeDrain() },
	}
	body := `{"action":"completed","repository":{"full_name":"acme/widget"},"check_suite":{"conclusion":"success","pull_requests":[{"number":5}]}}`
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/github/gh", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "check_suite")
	req.Header.Set("X-GitHub-Delivery", "ci-1")
	req.Header.Set("X-Hub-Signature-256", "sha256="+sign("s", []byte(body)))
	w := httptest.NewRecorder()
	h.Mux().ServeHTTP(w, req)
	require.Equal(t, http.StatusAccepted, w.Code)

	start := clk.now()
	// clearWait clears the wait, then comments: wait for both, or the
	// read can land between the two writes.
	lastComment := func() string {
		th, err := c.store.Thread(ctx, step.ID, store.Parts{Comments: true})
		if err != nil || len(th.Comments) == 0 {
			return ""
		}
		return th.Comments[len(th.Comments)-1].Body
	}
	untilCommit(t, c, func() bool {
		return waitOf(mustGet(t, c, step.ID)) == nil && strings.Contains(lastComment(), "Wait (CI on PR #5) ended")
	}, "the webhook resumes the step without a timer")
	assert.Equal(t, start, clk.now(), "fake time did not move")
	got := mustGet(t, c, step.ID)
	assert.Equal(t, sess.ID, got.Holder, "resume pre-assigns the parking session")
	assert.Contains(t, lastComment(), "checks green on aaaa111")
}

// commitBound fails a commit wait that never comes. It is only reached on
// failure, so it is generous for a CPU-starved -race run (JOB-84).
const commitBound = 2 * time.Minute

// untilCommit checks cond after each store commit until it holds. cond
// runs on the test goroutine, so nothing asserts after the test returns
// (require.Eventually's condition goroutine could, JOB-84). It needs a
// controller whose store kicks c.events (subController), with no
// pumpEvents draining the wake.
func untilCommit(t *testing.T, c *Controller, cond func() bool, msg string) {
	t.Helper()
	bound := time.After(commitBound)
	for !cond() {
		select {
		case <-c.events.wake:
		case <-bound:
			require.True(t, cond(), msg)
			return
		}
	}
}

// TestWaitWakeTargeted (#679): a wake for one PR evaluates only the
// waits on it; a run state change wakes the waits on that run.
func TestWaitWakeTargeted(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	green := `[{"name":"build","bucket":"pass"}]`
	prs.set(recPR, `[[]]`, green)
	prs.mu.Lock()
	prs.pr["acme/widget#6"], prs.checks["acme/widget#6"] = recPR, green
	prs.mu.Unlock()
	park := func(n int) domain.Ticket {
		tk := mkTicket(t, c, store.NewTicket{Title: "t", Pool: "dev"})
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: tk.ID, Reason: "ci", Conds: []domain.WaitCond{{Type: domain.WaitChecks, Repo: "acme/widget", PR: n, State: "success"}}})
		require.NoError(t, err)
		return tk
	}
	five, six := park(5), park(6)
	c.waitWakes.Take()

	c.wakeWaitsFor([]domain.InboxItem{{Source: "github", RepoOrTeam: "Acme/Widget", Number: 5}})
	w := c.waitWakes.Take()
	assert.False(t, w.All)
	c.pollWaitsOn(ctx, prs, threads, w.Subjects)
	assert.Nil(t, waitOf(mustGet(t, c, five.ID)), "PR 5's wait evaluated and cleared")
	assert.NotNil(t, waitOf(mustGet(t, c, six.ID)), "PR 6's wait not evaluated")

	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "r", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	f, err := domain.DecodeFields[domain.RunFields](run)
	require.NoError(t, err)
	_, err = c.setRunState(ctx, "controller", run, f, domain.RunDone, "")
	require.NoError(t, err)
	assert.True(t, c.waitWakes.Take().Subjects[waitengine.RunSubject(run.ID)], "run state change wakes its waits")

	c.wakeWaitsFor([]domain.InboxItem{{Source: "github", RepoOrTeam: "acme/widget"}})
	assert.True(t, c.waitWakes.Take().Subjects[waitengine.RepoSubject("acme/widget")], "an event with no number names the repo")

	c.wakeWaitsFor([]domain.InboxItem{{Source: "generic"}})
	assert.True(t, c.waitWakes.Take().All, "an item with no subject wakes a full pass")
}

// TestWaitNextTime (#679): the pump sleeps until the earliest time
// condition, not the whole interval.
func TestWaitNextTime(t *testing.T) {
	t.Parallel()
	c, clk, _, _ := waitController(t)
	at := clk.now().Add(30 * time.Second)
	tk := mkTicket(t, c, store.NewTicket{Title: "t", Pool: "dev"})
	_, err := c.waitAdd(t.Context(), scope{admin: true, actor: "admin"}, rpc.WaitAddParams{ID: tk.ID, Reason: "later", Conds: []domain.WaitCond{{Type: domain.WaitTime, At: at}}})
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, c.untilNextTime(t.Context()))
}
