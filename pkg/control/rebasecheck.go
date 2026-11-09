package control

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
<<<<<<< HEAD
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/store"
=======
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/storage/store"
>>>>>>> 7fea25a3 (refactor(storage): group store and journal packages under pkg/storage (JOB-173))
)

// Rebase check (#511): close -r rebase routes a run back without counting
// a failure, which relies on the agent judging the conflict. Before
// accepting it on a step whose run has a PR, read the PR's mergeable
// state: only dirty (conflicts) or behind justify a rebase. Anything else
// closes as failed. An unreadable or not yet computed state keeps the
// rebase and logs a WARN.

// runPR returns the repo and number of the last PR linked from run or
// its steps, zero when none.
func (c *Controller) runPR(ctx context.Context, run domain.Ticket) (string, int) {
	steps, err := store.Children(ctx, c.store, run.ID)
	if err != nil {
		return "", 0
	}
	repo, n := "", 0
	for _, s := range append([]domain.Ticket{run}, steps...) {
		cs, err := store.Comments(ctx, c.store, s.ID)
		if err != nil {
			continue
		}
		for _, th := range threads.Implicit(s, cs) {
			if th.PR {
				repo, n = th.Repo, th.Number
			}
		}
	}
	return repo, n
}

// rebaseResolution returns the resolution a rebase close on step gets:
// "failed" when the run's PR has no conflict with its base, else
// store.ResolutionRebase. The second value is the comment explaining a
// change.
func (c *Controller) rebaseResolution(ctx context.Context, step domain.Ticket) (string, string) {
	if step.Kind != domain.KindStep || step.ParentID == "" {
		return domain.ResolutionRebase, ""
	}
	run, err := c.store.Get(ctx, step.ParentID)
	if err != nil {
		return domain.ResolutionRebase, ""
	}
	repo, n := c.runPR(ctx, run)
	if n == 0 {
		return domain.ResolutionRebase, ""
	}
	src, ok := c.activitySource().(github.PullSource)
	if !ok {
		c.log.WarnContext(ctx, "rebase check: no PR source", "step", step.ID, "repo", repo, "pr", n)
		return domain.ResolutionRebase, ""
	}
	pr, err := src.PullRequest(ctx, repo, n)
	state := strings.ToLower(pr.MergeableState)
	needs, known := github.NeedsRebase(pr.MergeableState)
	if err != nil || !known {
		c.log.WarnContext(ctx, "rebase check: merge state unreadable, accepting rebase", "step", step.ID, "repo", repo, "pr", n, "state", state, "err", err)
		return domain.ResolutionRebase, ""
	}
	if needs {
		return domain.ResolutionRebase, ""
	}
	return "failed", fmt.Sprintf("close -r rebase counted as failed: PR %s#%d merge state is %s, not dirty or behind, so no rebase is needed.", repo, n, state)
}
