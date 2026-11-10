package control

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// fakeLinearWait serves Linear issue contexts by key.
type fakeLinearWait struct {
	mu sync.Mutex
	is map[string]linear.Context
}

func (f *fakeLinearWait) IssueContext(_ context.Context, key string) (linear.Context, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.is[key], nil
}

func (f *fakeLinearWait) set(key, state, typ string, labels ...string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.is[key] = linear.Context{Issue: linear.Issue{Identifier: key, State: linear.State{Name: state, Type: typ}}, Labels: labels}
}

// TestWaitEventKinds (JOB-11): github.tag, deploy, run, linear.state and
// linear.label each hold the ticket until their condition holds, then
// release it, and not before.
func TestWaitEventKinds(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	lin := &fakeLinearWait{is: map[string]linear.Context{}}
	c.linearWait = lin
	c.opt.Commit = "dddd0000"
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	lin.set("ENG-1", "In Progress", "started", "Blocked")

	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "dep", Fields: json.RawMessage(`{"workflow":"w","state":"running"}`)})
	park := func(title string, cd domain.WaitCond) domain.Ticket {
		tk := mkTicket(t, c, store.NewTicket{Title: title, Pool: "dev"})
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: tk.ID, Reason: title, Conds: []domain.WaitCond{cd}})
		require.NoError(t, err)
		return tk
	}
	tag := park("tag", domain.WaitCond{Type: domain.WaitTag, Repo: "acme/widget", Tag: "v1.2.3"})
	dep := park("deploy", domain.WaitCond{Type: domain.WaitDeploy, Repo: "acme/widget", SHA: "aaaa1111"})
	rn := park("run", domain.WaitCond{Type: domain.WaitRun, Run: run.ID, State: "done"})
	st := park("state", domain.WaitCond{Type: domain.WaitLinearState, Thread: "ENG-1", State: "done"})
	lb := park("label", domain.WaitCond{Type: domain.WaitLinearLabel, Thread: "ENG-1", Label: "blocked", Absent: true})

	prs.issues["compare:aaaa1111...dddd0000"] = "behind"
	c.pollWaits(ctx, prs, threads)
	assert.Empty(t, readyIDs(t, c), "nothing holds yet")
	// #716: an unmet condition's note must not read as met.
	assert.Equal(t, "deployed dddd000 does not contain aaaa111 yet", waitOf(mustGet(t, c, dep.ID)).Conds[0].Note)
	assert.Equal(t, "tag v1.2.3 not pushed to acme/widget yet", waitOf(mustGet(t, c, tag.ID)).Conds[0].Note)

	prs.issues["tag:acme/widget@v1.2.3"] = ""
	prs.issues["compare:aaaa1111...dddd0000"] = "ahead"
	c.pollWaits(ctx, prs, threads)
	assert.ElementsMatch(t, []string{tag.ID, dep.ID}, readyIDs(t, c))
	cs, err := store.Comments(ctx, c.store, dep.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "deployed dddd000 contains aaaa111")

	_, err = c.store.Update(ctx, "controller", run.ID, store.Update{Fields: json.RawMessage(`{"workflow":"w","state":"done"}`)})
	require.NoError(t, err)
	lin.set("ENG-1", "Done", "completed", "Blocked")
	c.pollWaits(ctx, prs, threads)
	assert.ElementsMatch(t, []string{tag.ID, dep.ID, rn.ID, st.ID}, readyIDs(t, c), "state matches the type, case-insensitive")
	assert.NotNil(t, waitOf(mustGet(t, c, lb.ID)), "label still present")

	lin.set("ENG-1", "Done", "completed")
	c.pollWaits(ctx, prs, threads)
	assert.Nil(t, waitOf(mustGet(t, c, lb.ID)))
}

// TestWaitRunCancelled: a done wait on a cancelled run is invalidated; a
// closed wait holds.
func TestWaitRunCancelled(t *testing.T) {
	t.Parallel()
	c, _, prs, threads := waitController(t)
	ctx := t.Context()
	admin := scope{admin: true, actor: "admin"}
	run := mkTicket(t, c, store.NewTicket{Kind: domain.KindRun, Title: "dep", Fields: json.RawMessage(`{"workflow":"w","state":"cancelled"}`)})
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	b := mkTicket(t, c, store.NewTicket{Title: "b", Pool: "dev"})
	for id, state := range map[string]string{a.ID: "done", b.ID: "closed"} {
		_, err := c.waitAdd(ctx, admin, rpc.WaitAddParams{ID: id, Reason: "dep", Conds: []domain.WaitCond{{Type: domain.WaitRun, Run: run.ID, State: state}}})
		require.NoError(t, err)
	}
	c.pollWaits(ctx, prs, threads)
	cs, err := store.Comments(ctx, c.store, a.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "invalidated: run "+run.ID+" was cancelled")
	cs, err = store.Comments(ctx, c.store, b.ID)
	require.NoError(t, err)
	assert.Contains(t, cs[len(cs)-1].Body, "condition met: run "+run.ID+" ended (cancelled)")
}

// TestWaitDeployNoCommit: a controller with no build commit refuses a
// deploy wait instead of parking a ticket that can never clear.
func TestWaitDeployNoCommit(t *testing.T) {
	t.Parallel()
	c, _, _, _ := waitController(t)
	a := mkTicket(t, c, store.NewTicket{Title: "a", Pool: "dev"})
	_, err := c.waitAdd(t.Context(), scope{admin: true, actor: "admin"}, rpc.WaitAddParams{
		ID: a.ID, Reason: "deploy",
		Conds: []domain.WaitCond{{Type: domain.WaitDeploy, Repo: "acme/widget", SHA: "aaaa1111"}},
	})
	require.ErrorIs(t, err, domain.ErrInvalid)
	assert.ErrorContains(t, err, "reports no build commit")
}

func TestWaitEventKindsValidate(t *testing.T) {
	t.Parallel()
	for _, cd := range []domain.WaitCond{
		{Type: domain.WaitTag, Repo: "acme/widget"},
		{Type: domain.WaitDeploy, Repo: "acme/widget", SHA: "main"},
		{Type: domain.WaitRun, Run: "js-1", State: "running"},
		{Type: domain.WaitLinearState, Thread: "acme/widget#1", State: "Done"},
		{Type: domain.WaitLinearLabel, Thread: "ENG-1"},
	} {
		w := &domain.Wait{Reason: "x", Conds: []domain.WaitCond{cd}}
		require.ErrorIs(t, w.Validate(), domain.ErrInvalid, "%+v", cd)
	}
}

// TestReleaseGuard (JOB-11): release --until for CI, a PR, a deploy, a
// tag or a merge is refused with the wait add line; --time-wait and a
// reason that is a real timer pass.
func TestReleaseGuard(t *testing.T) {
	t.Parallel()
	c, clk, _ := scanController(t)
	ctx := t.Context()
	until := clk.now().Add(30 * time.Minute)
	_, _, sc := startWork(t, c, "r1")
	for _, reason := range []string{"CI on PR #672", "waiting for checks", "Deploy pending", "tag v1 not pushed", "merge queue"} {
		_, err := c.release(ctx, sc, rpc.ReleaseParams{Reason: reason, Until: &until})
		require.ErrorIs(t, err, domain.ErrInvalid, reason)
		assert.ErrorContains(t, err, "jobshop wait add", reason)
	}
	_, err := c.release(ctx, sc, rpc.ReleaseParams{Reason: "precise: the deploy window opens at 18:00", Until: &until, TimeWait: true})
	require.NoError(t, err)

	_, _, sc = startWork(t, c, "r2")
	_, err = c.release(ctx, sc, rpc.ReleaseParams{Reason: "rate limit backoff", Until: &until})
	require.NoError(t, err, "not an event")
}
