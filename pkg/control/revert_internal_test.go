package control

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// fakeReverts is fakeActivity plus the PR reads revert detection needs.
type fakeReverts struct {
	*fakeActivity
	prs     map[int]github.Pull
	commits map[string][]int
}

func (f *fakeReverts) PullRequest(_ context.Context, _ string, n int) (github.Pull, error) {
	pr, ok := f.prs[n]
	if !ok {
		return github.Pull{}, fmt.Errorf("no PR %d", n)
	}
	return pr, nil
}

func (f *fakeReverts) CommitPulls(_ context.Context, _, sha string) ([]int, error) {
	return f.commits[sha], nil
}

const revRepo = "elee1766/jobshop"

func mergeItem(id string, n int, at time.Time) rpc.ActivityItem {
	return ghItem(revRepo, id, "merge", "elee1766", fmt.Sprint(n), at)
}

func revertLearnings(t *testing.T, c *Controller) []domain.Learning {
	t.Helper()
	ls, err := learning.List(t.Context(), c.store, store.LearningFilter{Kind: domain.LearningCorrection})
	require.NoError(t, err)
	return ls
}

func TestRevertLearning(t *testing.T) {
	t.Parallel()
	c, clk, fa := activityController(t)
	now := clk.now()
	run := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindRun, Title: "run", Fields: json.RawMessage(`{"workflow":"pr-work","state":"done"}`),
		Meta: json.RawMessage(`{"github":{"repo":"elee1766/jobshop","issue":265}}`),
	})
	step := mkTicket(t, c, store.NewTicket{
		Kind: domain.KindStep, Title: "implement", ParentID: run.ID,
		Fields: json.RawMessage(`{"step":"implement","run":"` + run.ID + `"}`),
	})
	_, err := store.AddComment(t.Context(), c.store, "s1", step.ID, "PR https://github.com/elee1766/jobshop/pull/268 head abc")
	require.NoError(t, err)

	src := &fakeReverts{
		fakeActivity: fa,
		prs: map[int]github.Pull{
			268: {Number: 268, Title: "recall (#265)", Merged: true, MergeCommitSHA: "aaaa1111bbbb"},
			270: {Number: 270, Title: "other fix", Body: "plain change", Merged: true, MergeCommitSHA: "cccc"},
			271: {Number: 271, Title: `Revert "recall (#265)"`, Body: "This reverts commit aaaa1111bbbb.\n\nbroke deploy", Merged: true, HTMLURL: "https://github.com/elee1766/jobshop/pull/271"},
			// A PR jobshop did not merge, and its revert.
			280: {Number: 280, Title: "manual", Merged: true, MergeCommitSHA: "dddd"},
			281: {Number: 281, Title: `Revert "manual"`, Body: "Reverts elee1766/jobshop#280", Merged: true},
		},
		commits: map[string][]int{"aaaa1111bbbb": {268}},
	}
	c.activitySrc = src
	fa.gh[revRepo] = []rpc.ActivityItem{
		mergeItem("m270", 270, now.Add(-3*time.Minute)),
		mergeItem("m281", 281, now.Add(-2*time.Minute)),
	}
	c.pollActivity(t, src)
	assert.Empty(t, revertLearnings(t, c), "a non-revert and a revert of a non-jobshop PR record nothing")

	fa.gh[revRepo] = append(fa.gh[revRepo], mergeItem("m271", 271, now.Add(-time.Minute)))
	c.pollActivity(t, src)
	ls := revertLearnings(t, c)
	require.Len(t, ls, 1)
	l := ls[0]
	assert.Equal(t, run.ID, l.Ticket)
	assert.Equal(t, domain.LearningSourceController, l.Source)
	assert.Equal(t, "repo:"+revRepo, l.Scope)
	assert.Contains(t, l.Summary, "PR #268 for issue #265")
	assert.Contains(t, l.Summary, "reverted by PR #271")
	assert.Contains(t, l.Fix, "broke deploy")

	// Seen again (cache lost, the event re-listed): still one learning.
	c.forgetActivity(t)
	c.pollActivity(t, src)
	assert.Len(t, revertLearnings(t, c), 1, "a revert is recorded once")
}
