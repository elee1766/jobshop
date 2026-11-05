package control

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Revert learnings (#272): when the activity feed sees a merged PR that
// reverts a PR merged through a pr-work run, record one correction on
// that run. The feed already polls every worked repo with ETags, so a
// check costs one PR fetch per new merge and nothing when idle.

// ghPR is the part of GET /repos/{r}/pulls/{n} revert detection reads.
type ghPR struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	HTMLURL        string `json:"html_url"`
	Merged         bool   `json:"merged"`
	MergeCommitSHA string `json:"merge_commit_sha"`
	// MergeableState is GitHub's mergeable_state (dirty, behind, clean,
	// ...), read by the rebase check (#511).
	MergeableState string `json:"mergeable_state"`
}

// revertSource fetches PRs for revert detection. liveActivity implements
// it; tests replace it.
type revertSource interface {
	PullRequest(ctx context.Context, repo string, n int) (ghPR, error)
	// CommitPulls lists the PRs a commit belongs to.
	CommitPulls(ctx context.Context, repo, sha string) ([]int, error)
}

var (
	revertsCommit = regexp.MustCompile(`(?i)this reverts commit ([0-9a-f]{7,40})`)
	// GitHub's revert button writes "Reverts owner/name#123".
	revertsPR = regexp.MustCompile(`(?i)reverts ([\w.-]+/[\w.-]+)?#(\d+)`)
)

// isRevert reports whether a PR looks like a revert at all.
func isRevert(pr ghPR) bool {
	return strings.HasPrefix(pr.Title, `Revert "`) || revertsCommit.MatchString(pr.Body) || revertsPR.MatchString(pr.Body)
}

// revertedPR finds the PR that pr reverts, checked against that PR's
// merge commit: a "Reverts #n" or the commit named in "This reverts
// commit <sha>" must be the original's merge commit. Zero means none.
func revertedPR(ctx context.Context, src revertSource, repo string, pr ghPR) (ghPR, error) {
	if m := revertsPR.FindStringSubmatch(pr.Body); m != nil && (m[1] == "" || strings.EqualFold(m[1], repo)) {
		n, _ := strconv.Atoi(m[2])
		orig, err := src.PullRequest(ctx, repo, n)
		if err != nil {
			return ghPR{}, err
		}
		if orig.Merged && orig.MergeCommitSHA != "" {
			return orig, nil
		}
	}
	m := revertsCommit.FindStringSubmatch(pr.Body)
	if m == nil {
		return ghPR{}, nil
	}
	sha := strings.ToLower(m[1])
	nums, err := src.CommitPulls(ctx, repo, sha)
	if err != nil {
		return ghPR{}, err
	}
	for _, n := range nums {
		if n == pr.Number {
			continue
		}
		orig, err := src.PullRequest(ctx, repo, n)
		if err != nil {
			return ghPR{}, err
		}
		if orig.Merged && strings.HasPrefix(strings.ToLower(orig.MergeCommitSHA), sha) {
			return orig, nil
		}
	}
	return ghPR{}, nil
}

// prWorkRun returns the pr-work run whose steps posted a link to
// repo#n (the "PR <url> head <sha>" handoff), so only PRs jobshop
// merged through a run count.
func (c *Controller) prWorkRun(ctx context.Context, repo string, n int) (store.Ticket, bool) {
	runs, err := c.store.ListRuns(ctx, "", "", 0)
	if err != nil {
		return store.Ticket{}, false
	}
	for _, r := range runs {
		rf, err := store.DecodeFields[store.RunFields](r)
		if err != nil || rf.Workflow != "pr-work" || !strings.EqualFold(metaString(r.Meta, "github", "repo"), repo) {
			continue
		}
		steps, err := c.store.ListChildren(ctx, r.ID)
		if err != nil {
			continue
		}
		for _, s := range append([]store.Ticket{r}, steps...) {
			cs, err := c.store.Comments(ctx, s.ID)
			if err != nil {
				continue
			}
			for _, th := range implicitThreads(s, cs) {
				if th.PR && th.Number == n && strings.EqualFold(th.Repo, repo) {
					return r, true
				}
			}
		}
	}
	return store.Ticket{}, false
}

// revertKey marks a revert learning so a later poll or a restart does
// not record the same revert twice.
func revertKey(repo string, n int) string { return fmt.Sprintf("revert %s#%d", repo, n) }

// checkReverts looks at newly seen merge items and records a correction
// for each revert of a pr-work PR. Errors only log: the feed goes on.
func (c *Controller) checkReverts(ctx context.Context, src revertSource, items []rpc.ActivityItem) {
	for _, it := range items {
		if it.Source != rpc.ActivityGitHub || it.Kind != "merge" || it.Ref == "" {
			continue
		}
		n, err := strconv.Atoi(it.Ref)
		if err != nil {
			continue
		}
		if err := c.checkRevert(ctx, src, it.Repo, n); err != nil {
			c.log.Warn("revert check", "repo", it.Repo, "pr", n, "err", err)
		}
	}
}

func (c *Controller) checkRevert(ctx context.Context, src revertSource, repo string, n int) error {
	pr, err := src.PullRequest(ctx, repo, n)
	if err != nil {
		return err
	}
	if !pr.Merged || !isRevert(pr) {
		return nil
	}
	orig, err := revertedPR(ctx, src, repo, pr)
	if err != nil || orig.Number == 0 {
		return err
	}
	run, ok := c.prWorkRun(ctx, repo, orig.Number)
	if !ok {
		return nil
	}
	key := revertKey(repo, pr.Number)
	prior, err := c.store.Learnings(ctx, store.LearningFilter{Ticket: run.ID, Kind: store.LearningCorrection})
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
		if i := refNumber(m.GitHub.Issue); i > 0 {
			issue = fmt.Sprintf(" for issue #%d", i)
		}
	}
	summary := fmt.Sprintf("%s: PR #%d%s (run %s) was reverted by PR #%d %s",
		key, orig.Number, issue, run.ID, pr.Number, pr.HTMLURL)
	c.record(ctx, store.LearningCorrection, run, summary, pr.Body)
	return nil
}
