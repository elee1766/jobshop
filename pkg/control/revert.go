package control

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/control/threads"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/learning"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Revert learnings (#272): when the activity feed sees a merged PR that
// reverts a PR merged through a pr-work run, record one correction on
// that run. The feed already polls every worked repo with ETags, so a
// check costs one PR fetch per new merge and nothing when idle.

// prWorkRun returns the pr-work run whose steps posted a link to
// repo#n (the "PR <url> head <sha>" handoff), so only PRs jobshop
// merged through a run count.
func (c *Controller) prWorkRun(ctx context.Context, repo string, n int) (domain.Ticket, bool) {
	runs, err := c.store.List(ctx, store.Filter{Kind: domain.KindRun, Order: store.OrderNewest})
	if err != nil {
		return domain.Ticket{}, false
	}
	for _, r := range runs {
		rf, err := domain.DecodeFields[domain.RunFields](r)
		if err != nil || rf.Workflow != "pr-work" || !strings.EqualFold(metaString(r.Meta, "github", "repo"), repo) {
			continue
		}
		steps, err := store.Children(ctx, c.store, r.ID)
		if err != nil {
			continue
		}
		for _, s := range append([]domain.Ticket{r}, steps...) {
			cs, err := store.Comments(ctx, c.store, s.ID)
			if err != nil {
				continue
			}
			for _, th := range threads.Implicit(s, cs) {
				if th.PR && th.Number == n && strings.EqualFold(th.Repo, repo) {
					return r, true
				}
			}
		}
	}
	return domain.Ticket{}, false
}

// revertKey marks a revert learning so a later poll or a restart does
// not record the same revert twice.
func revertKey(repo string, n int) string { return fmt.Sprintf("revert %s#%d", repo, n) }

// checkReverts looks at newly seen merge items and records a correction
// for each revert of a pr-work PR. Errors only log: the feed goes on.
func (c *Controller) checkReverts(ctx context.Context, src github.PullSource, items []rpc.ActivityItem) {
	for _, it := range items {
		if it.Source != rpc.ActivityGitHub || it.Kind != "merge" || it.Ref == "" {
			continue
		}
		n, err := strconv.Atoi(it.Ref)
		if err != nil {
			continue
		}
		if err := c.checkRevert(ctx, src, it.Repo, n); err != nil {
			c.log.WarnContext(ctx, "revert check", "repo", it.Repo, "pr", n, "err", err)
		}
	}
}

func (c *Controller) checkRevert(ctx context.Context, src github.PullSource, repo string, n int) error {
	pr, err := src.PullRequest(ctx, repo, n)
	if err != nil {
		return err
	}
	if !pr.Merged || !github.IsRevert(pr) {
		return nil
	}
	orig, err := github.RevertedPR(ctx, src, repo, pr)
	if err != nil || orig.Number == 0 {
		return err
	}
	run, ok := c.prWorkRun(ctx, repo, orig.Number)
	if !ok {
		return nil
	}
	key := revertKey(repo, pr.Number)
	prior, err := learning.List(ctx, c.store, store.LearningFilter{Ticket: run.ID, Kind: domain.LearningCorrection})
	if err != nil {
		return err
	}
	for _, l := range prior {
		if strings.HasPrefix(l.Summary, key+":") {
			return nil
		}
	}
	issue := ""
	var m struct {
		GitHub struct {
			Issue any `json:"issue"`
		} `json:"github"`
	}
	if json.Unmarshal(run.Meta, &m) == nil {
		if i := github.RefNumber(m.GitHub.Issue); i > 0 {
			issue = fmt.Sprintf(" for issue #%d", i)
		}
	}
	summary := fmt.Sprintf("%s: PR #%d%s (run %s) was reverted by PR #%d %s",
		key, orig.Number, issue, run.ID, pr.Number, pr.HTMLURL)
	c.record(ctx, domain.LearningCorrection, run, summary, pr.Body)
	return nil
}
