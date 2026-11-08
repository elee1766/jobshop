package waitengine_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/control/waitengine"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// fakeGitHub serves one PR and its checks.
type fakeGitHub struct {
	pr     github.PRInfo
	checks []github.Check
	prRead int
}

func (f *fakeGitHub) PR(context.Context, string, int) (github.PRInfo, error) {
	f.prRead++
	return f.pr, nil
}

func (f *fakeGitHub) Reviews(context.Context, string, int) ([]github.Review, error) {
	return nil, nil
}

func (f *fakeGitHub) Checks(context.Context, string, int, string, bool) ([]github.Check, error) {
	return f.checks, nil
}

func (f *fakeGitHub) WorkflowRuns(context.Context, string, string) ([]github.WorkflowRun, error) {
	return nil, nil
}

func (f *fakeGitHub) Issue(context.Context, string, int) (github.IssueInfo, error) {
	return github.IssueInfo{}, nil
}
func (f *fakeGitHub) Tag(context.Context, string, string) (bool, error) { return false, nil }
func (f *fakeGitHub) Contains(context.Context, string, string, string) (bool, error) {
	return false, nil
}

type noLinear struct{}

func (noLinear) IssueContext(context.Context, string) (linear.Context, error) {
	return linear.Context{}, nil
}

var t0 = time.Date(2026, 11, 8, 12, 0, 0, 0, time.UTC)

func pass(gh *fakeGitHub) *waitengine.Pass {
	return waitengine.NewPass(waitengine.Sources{GitHub: gh, Linear: noLinear{}})
}

func checksWait(state string) store.WaitCond {
	return store.WaitCond{Type: store.WaitChecks, Repo: "acme/widget", PR: 5, State: state}
}

// TestEvaluateDoesNotMutate: Evaluate returns the new wait and leaves the
// snapshot as the controller read it, so the controller can still compare
// and apply at the snapshot's revision.
func TestEvaluateDoesNotMutate(t *testing.T) {
	t.Parallel()
	gh := &fakeGitHub{pr: github.PRInfo{Head: "aaaa1111", State: "open"}, checks: []github.Check{{Name: "go", Bucket: "pending"}}}
	w := store.Wait{Conds: []store.WaitCond{checksWait("success")}, Reason: "ci", Since: t0}
	d := waitengine.Evaluate(t.Context(), pass(gh), w, t0)
	assert.False(t, d.Clear)
	assert.True(t, d.Save, "the PR head was bound")
	assert.Equal(t, "aaaa1111", d.Wait.Conds[0].Head)
	assert.Empty(t, w.Conds[0].Head, "the snapshot is untouched")
	assert.Empty(t, w.Conds[0].Note)
}

// TestEvaluateAnyAll: any clears on one met condition, all needs every
// one; a condition error keeps the others' progress.
func TestEvaluateAnyAll(t *testing.T) {
	t.Parallel()
	gh := &fakeGitHub{pr: github.PRInfo{Head: "aaaa1111", State: "open"}, checks: []github.Check{{Name: "go", Bucket: "pass"}}}
	later := store.WaitCond{Type: store.WaitTime, At: t0.Add(time.Hour)}
	anyWait := store.Wait{Conds: []store.WaitCond{checksWait("success"), later}, Mode: store.WaitAny, Reason: "r", Since: t0}
	d := waitengine.Evaluate(t.Context(), pass(gh), anyWait, t0)
	require.True(t, d.Clear)
	assert.Contains(t, d.Why, "condition met: checks green on aaaa111")

	all := anyWait
	all.Mode = ""
	d = waitengine.Evaluate(t.Context(), pass(gh), all, t0)
	require.False(t, d.Clear)
	assert.True(t, d.Wait.Conds[0].Met)
	assert.False(t, d.Wait.Conds[1].Met)
	assert.Equal(t, "time "+later.At.Format(time.RFC3339), d.Wait.Conds[1].Note)

	d = waitengine.Evaluate(t.Context(), pass(gh), all, t0.Add(time.Hour))
	assert.True(t, d.Clear)
}

// TestEvaluateTimeout: the timeout escalates once.
func TestEvaluateTimeout(t *testing.T) {
	t.Parallel()
	gh := &fakeGitHub{pr: github.PRInfo{Head: "aaaa1111", State: "open"}, checks: []github.Check{{Name: "go", Bucket: "pending"}}}
	w := store.Wait{Conds: []store.WaitCond{checksWait("success")}, Reason: "ci", Since: t0, Timeout: t0.Add(time.Minute)}
	d := waitengine.Evaluate(t.Context(), pass(gh), w, t0)
	assert.False(t, d.TimedOut)
	d = waitengine.Evaluate(t.Context(), pass(gh), w, t0.Add(time.Minute))
	require.True(t, d.TimedOut)
	assert.True(t, d.Save)
	assert.True(t, d.Wait.TimedOut)
	d = waitengine.Evaluate(t.Context(), pass(gh), d.Wait, t0.Add(2*time.Minute))
	assert.False(t, d.TimedOut, "already escalated")
}

// TestEvaluateHeadMoved: a checks condition bound to a head is invalidated
// when the PR head moves, whatever the checks of the new head say.
func TestEvaluateHeadMoved(t *testing.T) {
	t.Parallel()
	gh := &fakeGitHub{pr: github.PRInfo{Head: "cccc2222", State: "open"}, checks: []github.Check{{Name: "go", Bucket: "pass"}}}
	cd := checksWait("success")
	cd.Head = "aaaa1111"
	d := waitengine.Evaluate(t.Context(), pass(gh), store.Wait{Conds: []store.WaitCond{cd}, Reason: "ci", Since: t0}, t0)
	require.True(t, d.Clear)
	assert.Equal(t, "invalidated: acme/widget#5 head moved aaaa111 -> cccc222", d.Why)
}

// TestPassSharesReads: waits on one PR share a fetch within a pass, and a
// new pass reads again.
func TestPassSharesReads(t *testing.T) {
	t.Parallel()
	gh := &fakeGitHub{pr: github.PRInfo{Head: "aaaa1111", State: "open"}, checks: []github.Check{{Name: "go", Bucket: "pending"}}}
	w := store.Wait{Conds: []store.WaitCond{checksWait("success")}, Reason: "ci", Since: t0}
	p := pass(gh)
	waitengine.Evaluate(t.Context(), p, w, t0)
	waitengine.Evaluate(t.Context(), p, w, t0)
	assert.Equal(t, 1, gh.prRead)
	waitengine.Evaluate(t.Context(), pass(gh), w, t0)
	assert.Equal(t, 2, gh.prRead)
}

// TestWakesKeepLateRequests: a request queued after Take, as when an event
// arrives while a pass runs, is not lost; the next Take returns it.
func TestWakesKeepLateRequests(t *testing.T) {
	t.Parallel()
	var w waitengine.Wakes
	assert.True(t, w.Subjects(waitengine.IssueSubject("Acme/Widget", 5)))
	first := w.Take()
	assert.Equal(t, map[string]bool{"gh:acme/widget#5": true}, first.Subjects)
	assert.Equal(t, waitengine.Pending{}, w.Take(), "Take clears")

	w.Items([]store.InboxItem{{Source: "github", RepoOrTeam: "acme/widget", Number: 5}})
	w.Items([]store.InboxItem{{Source: "generic"}})
	got := w.Take()
	assert.True(t, got.All)
	assert.True(t, got.Subjects["gh:acme/widget#5"])
	assert.False(t, w.Subjects(), "nothing to queue")
}

func TestSubjects(t *testing.T) {
	t.Parallel()
	cd := store.WaitCond{Type: store.WaitChecks, Repo: "Acme/Widget", PR: 5}
	assert.Equal(t, []string{"gh:acme/widget#5", "repo:acme/widget"}, waitengine.CondSubjects(cd))
	cd.PR = 0
	assert.Equal(t, []string{"repo:acme/widget"}, waitengine.CondSubjects(cd))
	assert.Equal(t, []string{"linear:ENG-1"}, waitengine.CondSubjects(store.WaitCond{Type: store.WaitReply, Thread: "eng-1"}))
	assert.Empty(t, waitengine.CondSubjects(store.WaitCond{Type: store.WaitDeploy}))
	assert.Equal(t, []string{"linear:ENG-7"}, waitengine.ItemSubjects(store.InboxItem{Source: "linear", RepoOrTeam: "ENG", Number: 7}))

	met := store.Wait{Conds: []store.WaitCond{{Type: store.WaitRun, Run: "r1", Met: true}}}
	assert.False(t, waitengine.Names(&met, map[string]bool{"run:r1": true}), "a met condition names nothing")
	assert.False(t, waitengine.Names(nil, map[string]bool{"run:r1": true}))
}

func TestNextDue(t *testing.T) {
	t.Parallel()
	every := time.Minute
	ws := []store.Wait{
		{Conds: []store.WaitCond{{Type: store.WaitTime, At: t0.Add(30 * time.Second)}}},
		{Timeout: t0.Add(10 * time.Second), Conds: []store.WaitCond{checksWait("success")}},
		{Timeout: t0.Add(time.Second), TimedOut: true},
	}
	assert.Equal(t, 10*time.Second, waitengine.NextDue(ws, t0, every), "an escalated timeout no longer sets the timer")
	assert.Equal(t, 0*time.Second, waitengine.NextDue(ws, t0.Add(time.Hour), every), "overdue means now")
	assert.Equal(t, every, waitengine.NextDue(nil, t0, every))
}

func TestTimerHint(t *testing.T) {
	t.Parallel()
	require.ErrorIs(t, waitengine.TimerHint("CI on PR #672"), store.ErrInvalid)
	assert.NoError(t, waitengine.TimerHint("rate limit backoff"))
}
