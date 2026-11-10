package github

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/support/telemetry"
)

// Read-only observation of GitHub state (JOB-143): what the controller's
// waits and subscriptions ask about a PR, an issue, a commit or a thread.
// Everything here reads and returns plain values. It decides nothing: a
// wait clearing, a ticket reopening or a rebase being refused stays with
// the controller.

// PRInfo is the part of a PR a wait reads.
type PRInfo struct {
	Head   string
	State  string // open or closed
	Merged bool
	// Conflicting: GitHub reports mergeable false. No pull_request CI runs
	// on such a PR, so a checks wait on it would never resolve.
	Conflicting bool
}

// Review is one review on a PR.
type Review struct {
	Login  string
	State  string // APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED
	Commit string
	At     time.Time
}

// Check is one check of a PR or commit. Bucket is gh's: pass, fail,
// pending, skipping, cancel.
type Check struct {
	Name   string
	Bucket string
}

// IssueInfo is the state and labels of an issue.
type IssueInfo struct {
	State  string // open or closed
	Labels []string
}

// WorkflowRun is one GitHub Actions run on a commit.
type WorkflowRun struct {
	ID         int64
	Name       string
	Path       string // .github/workflows/ci.yml
	Status     string
	Conclusion string
}

// Thread is a GitHub issue or PR conversation.
type Thread struct {
	Repo   string
	Number int
	PR     bool
}

// Comment is a comment or review on a thread.
type Comment struct {
	ID     string
	Author string
	Bot    bool
	Kind   string // comment, review, review comment
	State  string // review state
	Body   string
	URL    string
	At     time.Time
}

// CLI observes GitHub through the gh CLI, which has the host's token.
// It is the controller's default; tests replace the controller's
// interfaces instead.
type CLI struct{}

// ghCmdName names a gh call for its exec span: the subcommand only ("api",
// "pr"), never the paths or flags after it.
func ghCmdName(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

// run executes gh with a two minute bound on top of ctx.
func run(ctx context.Context, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cmd, done := telemetry.ExecContext(ctx, "gh-"+ghCmdName(args), "gh", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	done(err)
	if err != nil {
		return b, fmt.Errorf("gh %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return b, nil
}

// PR reads GET /repos/{repo}/pulls/{n}.
func (CLI) PR(ctx context.Context, repo string, n int) (PRInfo, error) {
	b, err := run(ctx, "api", fmt.Sprintf("repos/%s/pulls/%d", repo, n))
	if err != nil {
		return PRInfo{}, err
	}
	return ParsePR(b)
}

// Reviews reads every page of GET .../pulls/{n}/reviews.
func (CLI) Reviews(ctx context.Context, repo string, n int) ([]Review, error) {
	b, err := run(ctx, "api", "--paginate", "--slurp", fmt.Sprintf("repos/%s/pulls/%d/reviews?per_page=100", repo, n))
	if err != nil {
		return nil, err
	}
	return ParseReviews(b)
}

// Checks lists the checks of PR n's current head (required only when
// asked), or of sha when n is 0.
func (CLI) Checks(ctx context.Context, repo string, n int, sha string, required bool) ([]Check, error) {
	if n == 0 {
		return commitChecks(ctx, repo, sha)
	}
	args := []string{"pr", "checks", strconv.Itoa(n), "--repo", repo, "--json", "name,bucket"}
	if required {
		args = append(args, "--required")
	}
	// gh pr checks exits non-zero while checks are pending or failing;
	// its JSON is still the answer.
	b, err := run(ctx, args...)
	if cs, perr := ParsePRChecks(b); perr == nil && len(bytes.TrimSpace(b)) > 0 {
		return cs, nil
	}
	if err != nil && (strings.Contains(err.Error(), "no required checks") || strings.Contains(err.Error(), "no checks reported")) {
		return nil, nil // nothing reported yet: not green
	}
	return nil, err
}

// Tag reads GET /repos/{repo}/git/ref/tags/{tag}; 404 is not yet.
func (CLI) Tag(ctx context.Context, repo, tag string) (bool, error) {
	_, err := run(ctx, "api", fmt.Sprintf("repos/%s/git/ref/tags/%s", repo, tag))
	if err != nil && strings.Contains(err.Error(), "404") {
		return false, nil
	}
	return err == nil, err
}

// Contains compares sha...head: identical or ahead means head has sha.
func (CLI) Contains(ctx context.Context, repo, sha, head string) (bool, error) {
	b, err := run(ctx, "api", fmt.Sprintf("repos/%s/compare/%s...%s", repo, sha, head), "--jq", ".status")
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil // head not pushed (a local build): not deployed
		}
		return false, err
	}
	return CompareContains(string(b)), nil
}

// Issue reads GET /repos/{repo}/issues/{n}.
func (CLI) Issue(ctx context.Context, repo string, n int) (IssueInfo, error) {
	b, err := run(ctx, "api", fmt.Sprintf("repos/%s/issues/%d", repo, n))
	if err != nil {
		return IssueInfo{}, err
	}
	return ParseIssue(b)
}

// WorkflowRuns lists the Actions runs on commit sha, every workflow.
func (CLI) WorkflowRuns(ctx context.Context, repo, sha string) ([]WorkflowRun, error) {
	full := sha
	if len(full) < 40 { // head_sha takes only a full sha
		out, err := run(ctx, "api", fmt.Sprintf("repos/%s/commits/%s", repo, sha), "--jq", ".sha")
		if err != nil {
			return nil, err
		}
		full = strings.TrimSpace(string(out))
	}
	b, err := run(ctx, "api", fmt.Sprintf("repos/%s/actions/runs?head_sha=%s&per_page=100", repo, full))
	if err != nil {
		return nil, err
	}
	return ParseRunList(b)
}

// commitChecks lists the check runs of sha plus its Actions workflow
// runs (#831). A workflow run queued behind a concurrency group has no
// check runs yet; without it a commit looks green once another workflow
// (release) finishes, and a run cancelled while queued would not show.
func commitChecks(ctx context.Context, repo, sha string) ([]Check, error) {
	b, err := run(ctx, "api", fmt.Sprintf("repos/%s/commits/%s/check-runs?per_page=100", repo, sha))
	if err != nil {
		return nil, err
	}
	cs, err := ParseCheckRuns(b)
	if err != nil {
		return nil, err
	}
	rs, err := CLI{}.WorkflowRuns(ctx, repo, sha)
	if err != nil {
		return nil, err
	}
	for _, r := range rs {
		cs = append(cs, Check{Name: r.Name + " (workflow)", Bucket: CheckBucket(r.Status, r.Conclusion)})
	}
	return cs, nil
}

// Comments reads the comments of th newer than since: issue comments,
// and for a PR its reviews and review comments. Reviews carry no since
// filter, so the caller filters by At.
func (CLI) Comments(ctx context.Context, th Thread, since time.Time) ([]Comment, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	base := fmt.Sprintf("repos/%s", th.Repo)
	s := since.UTC().Format(time.RFC3339)
	const user = `author: .user.login, bot: (.user.type == "Bot")`
	queries := [][2]string{{
		fmt.Sprintf("%s/issues/%d/comments?since=%s&per_page=100", base, th.Number, s),
		`.[] | {id: ("c" + (.id|tostring)), kind: "comment", body, url: .html_url, at: .updated_at, ` + user + `}`,
	}}
	if th.PR {
		queries = append(queries,
			[2]string{
				fmt.Sprintf("%s/pulls/%d/reviews?per_page=100", base, th.Number),
				`.[] | select(.submitted_at != null) | {id: ("r" + (.id|tostring)), kind: "review", state, body, url: .html_url, at: .submitted_at, ` + user + `}`,
			},
			[2]string{
				fmt.Sprintf("%s/pulls/%d/comments?since=%s&per_page=100", base, th.Number, s),
				`.[] | {id: ("p" + (.id|tostring)), kind: "review comment", body, url: .html_url, at: .updated_at, ` + user + `}`,
			})
	}
	var out []Comment
	for _, q := range queries {
		cmd, done := telemetry.ExecContext(ctx, "gh-api", "gh", "api", "--paginate", q[0], "--jq", q[1])
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		b, err := cmd.Output()
		done(err)
		if err != nil {
			return nil, fmt.Errorf("gh api %s: %w: %s", q[0], err, strings.TrimSpace(stderr.String()))
		}
		cs, err := ParseCommentLines(b)
		if err != nil {
			return nil, err
		}
		out = append(out, cs...)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out, nil
}

// ParseCommentLines reads the JSON lines the gh --jq filters print.
func ParseCommentLines(b []byte) ([]Comment, error) {
	var out []Comment
	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var r struct {
			ID     string    `json:"id"`
			Kind   string    `json:"kind"`
			State  string    `json:"state"`
			Body   string    `json:"body"`
			URL    string    `json:"url"`
			At     time.Time `json:"at"`
			Author string    `json:"author"`
			Bot    bool      `json:"bot"`
		}
		if err := json.Unmarshal(line, &r); err != nil {
			return nil, fmt.Errorf("gh api output: %w", err)
		}
		if r.Kind == "review" && strings.TrimSpace(r.Body) == "" && r.State == "COMMENTED" {
			continue // the review's line comments arrive on their own
		}
		out = append(out, Comment{ID: r.ID, Author: r.Author, Bot: r.Bot, Kind: r.Kind, State: r.State, Body: r.Body, URL: r.URL, At: r.At})
	}
	return out, sc.Err()
}

// CompareContains reads the compare API's status.
func CompareContains(status string) bool {
	s := strings.TrimSpace(status)
	return s == "identical" || s == "ahead"
}

// ParseIssue reads GET /repos/{repo}/issues/{n}.
func ParseIssue(b []byte) (IssueInfo, error) {
	var r struct {
		State  string `json:"state"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return IssueInfo{}, fmt.Errorf("issue: %w", err)
	}
	out := IssueInfo{State: r.State}
	for _, l := range r.Labels {
		out.Labels = append(out.Labels, l.Name)
	}
	return out, nil
}

// ParsePR reads GET /repos/{repo}/pulls/{n}.
func ParsePR(b []byte) (PRInfo, error) {
	var r struct {
		State     string `json:"state"`
		Merged    bool   `json:"merged"`
		Mergeable *bool  `json:"mergeable"` // null while GitHub computes it
		Head      struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return PRInfo{}, fmt.Errorf("pull request: %w", err)
	}
	return PRInfo{Head: r.Head.SHA, State: r.State, Merged: r.Merged, Conflicting: r.Mergeable != nil && !*r.Mergeable}, nil
}

// ParseReviews reads the slurped pages of GET .../pulls/{n}/reviews.
func ParseReviews(b []byte) ([]Review, error) {
	var pages [][]struct {
		State       string    `json:"state"`
		CommitID    string    `json:"commit_id"`
		SubmittedAt time.Time `json:"submitted_at"`
		User        struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := json.Unmarshal(b, &pages); err != nil {
		return nil, fmt.Errorf("reviews: %w", err)
	}
	var out []Review
	for _, p := range pages {
		for _, r := range p {
			out = append(out, Review{Login: r.User.Login, State: r.State, Commit: r.CommitID, At: r.SubmittedAt})
		}
	}
	return out, nil
}

// ParsePRChecks reads `gh pr checks --json name,bucket`.
func ParsePRChecks(b []byte) ([]Check, error) {
	var out []Check
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, fmt.Errorf("pr checks: %w", err)
	}
	return out, nil
}

// ParseRunList reads GET .../actions/runs?head_sha=.
func ParseRunList(b []byte) ([]WorkflowRun, error) {
	var r struct {
		WorkflowRuns []struct {
			ID         int64  `json:"id"`
			Name       string `json:"name"`
			Path       string `json:"path"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"workflow_runs"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("workflow runs: %w", err)
	}
	var out []WorkflowRun
	for _, wr := range r.WorkflowRuns {
		out = append(out, WorkflowRun{ID: wr.ID, Name: wr.Name, Path: wr.Path, Status: wr.Status, Conclusion: wr.Conclusion})
	}
	return out, nil
}

// ParseCheckRuns reads GET .../commits/{sha}/check-runs into gh buckets.
func ParseCheckRuns(b []byte) ([]Check, error) {
	var r struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		} `json:"check_runs"`
	}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("check runs: %w", err)
	}
	var out []Check
	for _, cr := range r.CheckRuns {
		out = append(out, Check{Name: cr.Name, Bucket: CheckBucket(cr.Status, cr.Conclusion)})
	}
	return out, nil
}

// CheckBucket maps a check or workflow run's status and conclusion to
// gh's buckets.
func CheckBucket(status, conclusion string) string {
	if status != "completed" {
		return "pending"
	}
	switch conclusion {
	case "success", "neutral":
		return "pass"
	case "skipped":
		return "skipping"
	case "cancelled":
		return "cancel"
	}
	return "fail"
}

// NewestRun is the highest run id among the runs of workflow, matched
// case-insensitively by name or by workflow file (ci, ci.yml or
// .github/workflows/ci.yml), as trigger matching does: workflow=ci must
// find a workflow named CI in ci.yml (JOB-88).
func NewestRun(rs []WorkflowRun, workflow string) (WorkflowRun, bool) {
	var best WorkflowRun
	found := false
	for _, r := range rs {
		base := path.Base(r.Path)
		stem := strings.TrimSuffix(base, path.Ext(base))
		matched := false
		for _, s := range []string{r.Name, r.Path, base, stem} {
			if s != "" && strings.EqualFold(s, workflow) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		if !found || r.ID > best.ID {
			best, found = r, true
		}
	}
	return best, found
}

// NeedsRebase classifies a PR's mergeable_state for the rebase check
// (#511): only dirty (conflicts) or behind justify a rebase. known is
// false while GitHub has not computed the state (empty or unknown).
func NeedsRebase(mergeableState string) (needs, known bool) {
	state := strings.ToLower(mergeableState)
	if state == "" || state == "unknown" {
		return false, false
	}
	return state == "dirty" || state == "behind", true
}

// prLink finds PR URLs in comments, e.g. the pr-work "PR <url> head <sha>".
var prLink = regexp.MustCompile(`https://github\.com/([\w.-]+/[\w.-]+)/pull/(\d+)`)

// PRRefs returns the PR URLs in text, in order, with repeats.
func PRRefs(text string) []Thread {
	var out []Thread
	for _, sm := range prLink.FindAllStringSubmatch(text, -1) {
		n, _ := strconv.Atoi(sm[2])
		out = append(out, Thread{Repo: sm[1], Number: n, PR: true})
	}
	return out
}

// RefNumber reads an issue or PR number from ticket meta: a JSON number,
// or a string such as "12" or "#12". Zero is none.
func RefNumber(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case string:
		n, _ := strconv.Atoi(strings.TrimPrefix(x, "#"))
		return n
	}
	return 0
}
