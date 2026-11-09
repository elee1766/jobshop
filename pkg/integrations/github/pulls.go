package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// REST reads of one PR and revert detection (JOB-159). Pull is the REST
// view of GET /repos/{r}/pulls/{n}; PRInfo is the gh-CLI view the waits
// read. The two readers stay apart on purpose: Client uses the identity
// token of the repo, CLI uses the host gh token (JOB-143).

// Pull is the part of GET /repos/{r}/pulls/{n} that revert detection and
// the rebase check read.
type Pull struct {
	Number         int    `json:"number"`
	Title          string `json:"title"`
	Body           string `json:"body"`
	HTMLURL        string `json:"html_url"`
	Merged         bool   `json:"merged"`
	MergeCommitSHA string `json:"merge_commit_sha"`
	// MergeableState is GitHub's mergeable_state (dirty, behind, clean,
	// ...), read by the rebase check (#511); see NeedsRebase.
	MergeableState string `json:"mergeable_state"`
}

// PullSource reads PRs. Client implements it for one repo's identity;
// callers that pick a client per repo wrap it, and tests replace it.
type PullSource interface {
	PullRequest(ctx context.Context, repo string, n int) (Pull, error)
	// CommitPulls lists the PRs a commit belongs to.
	CommitPulls(ctx context.Context, repo, sha string) ([]int, error)
}

var _ PullSource = (*Client)(nil)

// pullRoute names these calls in traces. It stays the activity route the
// reads had before they moved here, so dashboards keep their series.
const pullRoute = "/repos/{repo}/activity"

// PullRequest reads PR n of repo.
func (c *Client) PullRequest(ctx context.Context, repo string, n int) (Pull, error) {
	var pr Pull
	resp, err := c.Do(ctx, Request{Path: fmt.Sprintf("/repos/%s/pulls/%d", repo, n), Route: pullRoute})
	if err != nil {
		return Pull{}, err
	}
	return pr, json.Unmarshal(resp.Body, &pr)
}

// CommitPulls lists the PRs a commit of repo is associated with.
func (c *Client) CommitPulls(ctx context.Context, repo, sha string) ([]int, error) {
	resp, err := c.Do(ctx, Request{Path: fmt.Sprintf("/repos/%s/commits/%s/pulls", repo, url.PathEscape(sha)), Route: pullRoute})
	if err != nil {
		return nil, err
	}
	var prs []struct {
		Number int `json:"number"`
	}
	if err := json.Unmarshal(resp.Body, &prs); err != nil {
		return nil, err
	}
	out := make([]int, len(prs))
	for i, p := range prs {
		out[i] = p.Number
	}
	return out, nil
}

var (
	revertsCommit = regexp.MustCompile(`(?i)this reverts commit ([0-9a-f]{7,40})`)
	// GitHub's revert button writes "Reverts owner/name#123".
	revertsPR = regexp.MustCompile(`(?i)reverts ([\w.-]+/[\w.-]+)?#(\d+)`)
)

// IsRevert reports whether a PR looks like a revert at all.
func IsRevert(pr Pull) bool {
	return strings.HasPrefix(pr.Title, `Revert "`) || revertsCommit.MatchString(pr.Body) || revertsPR.MatchString(pr.Body)
}

// RevertedPR finds the PR that pr reverts, checked against that PR's
// merge commit: a "Reverts #n" or the commit named in "This reverts
// commit <sha>" must be the original's merge commit. Zero means none.
func RevertedPR(ctx context.Context, src PullSource, repo string, pr Pull) (Pull, error) {
	if m := revertsPR.FindStringSubmatch(pr.Body); m != nil && (m[1] == "" || strings.EqualFold(m[1], repo)) {
		n, _ := strconv.Atoi(m[2])
		orig, err := src.PullRequest(ctx, repo, n)
		if err != nil {
			return Pull{}, err
		}
		if orig.Merged && orig.MergeCommitSHA != "" {
			return orig, nil
		}
	}
	m := revertsCommit.FindStringSubmatch(pr.Body)
	if m == nil {
		return Pull{}, nil
	}
	sha := strings.ToLower(m[1])
	nums, err := src.CommitPulls(ctx, repo, sha)
	if err != nil {
		return Pull{}, err
	}
	for _, n := range nums {
		if n == pr.Number {
			continue
		}
		orig, err := src.PullRequest(ctx, repo, n)
		if err != nil {
			return Pull{}, err
		}
		if orig.Merged && strings.HasPrefix(strings.ToLower(orig.MergeCommitSHA), sha) {
			return orig, nil
		}
	}
	return Pull{}, nil
}
