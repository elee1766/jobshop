package activity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// GET /repos/{r}/events answers 200 with [] for an installation token
// (JOB-108), so a github-app identity gets its feed from the per-resource
// endpoints below: pulls, issues, issue and review comments, reviews and
// commits on the default branch. Each has its own ETag in repoPoll.ETags,
// so an unchanged repo costs 304s, as with events.

// appReviewPRs bounds the per-PR review reads on one poll.
const appReviewPRs = 5

// appActivity reports whether repo's feed must come from the per-resource
// endpoints instead of events: a github-app identity covers it, or the
// token is an installation token ("ghs_") a github-token identity holds,
// for example one minted by a command secret.
func (l *Live) appActivity(ctx context.Context, cl *github.Client, repo string) bool {
	if tok, err := cl.Auth.Token(ctx); err == nil && strings.HasPrefix(tok, "ghs_") {
		return true
	}
	cfg := l.C.Config()
	name, ok := cfg.IdentityFor(cfg.IdentityNames(), repo)
	return ok && cfg.Identities[name].Kind == config.IdentityGitHubApp
}

// fetchAppActivity reads the per-resource endpoints. A failing endpoint
// (a permission the App lacks) does not stop the others: the items read
// so far come back with the joined errors.
func (l *Live) fetchAppActivity(ctx context.Context, cl *github.Client, repo string, st *RepoPoll) ([]rpc.ActivityItem, error) {
	var out []rpc.ActivityItem
	var errs []error
	read := func(name, path string, parse func([]byte) ([]rpc.ActivityItem, error)) []byte {
		body, h, err := l.get(ctx, cl, "/repos/"+repo+path, st.ETags[name])
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			return nil
		}
		if body == nil {
			return nil
		}
		items, err := parse(body)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", name, err))
			return nil
		}
		out = append(out, items...)
		if st.ETags == nil {
			st.ETags = map[string]string{}
		}
		st.ETags[name] = h.Get("ETag")
		return body
	}

	if body := read("pulls", "/pulls?state=all&sort=updated&direction=desc&per_page=30", func(b []byte) ([]rpc.ActivityItem, error) {
		return parseAppPulls(repo, b)
	}); body != nil {
		// A review does not show in the pulls list: read the newest PRs'.
		var prs []struct {
			Number int `json:"number"`
		}
		if json.Unmarshal(body, &prs) == nil {
			for _, p := range prs[:min(len(prs), appReviewPRs)] {
				b, _, err := l.get(ctx, cl, fmt.Sprintf("/repos/%s/pulls/%d/reviews?per_page=30", repo, p.Number), "")
				if err != nil {
					errs = append(errs, fmt.Errorf("reviews: %w", err))
					break
				}
				items, err := parseAppReviews(repo, p.Number, b)
				if err != nil {
					errs = append(errs, fmt.Errorf("reviews: %w", err))
					break
				}
				out = append(out, items...)
			}
		}
	}
	read("issues", "/issues?state=all&sort=updated&direction=desc&per_page=30", func(b []byte) ([]rpc.ActivityItem, error) {
		return parseAppIssues(repo, b)
	})
	read("issue_comments", "/issues/comments?sort=created&direction=desc&per_page=50", func(b []byte) ([]rpc.ActivityItem, error) {
		return parseAppComments(repo, b, false)
	})
	read("review_comments", "/pulls/comments?sort=created&direction=desc&per_page=50", func(b []byte) ([]rpc.ActivityItem, error) {
		return parseAppComments(repo, b, true)
	})
	if st.DefaultBranch != "" {
		read("commits", "/commits?per_page=30&sha="+url.QueryEscape(st.DefaultBranch), func(b []byte) ([]rpc.ActivityItem, error) {
			return parseAppCommits(repo, b)
		})
	}
	return out, errors.Join(errs...)
}

// lastSegment is the trailing number of an API url such as .../issues/12.
func lastSegment(u string) int {
	n, _ := strconv.Atoi(u[strings.LastIndex(u, "/")+1:])
	return n
}

type appUser struct {
	Login string `json:"login"`
}

// parseAppPulls maps GET /pulls to opened, merged and closed items. The
// list has no merger, so the author is the actor of each.
func parseAppPulls(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var prs []struct {
		Number    int        `json:"number"`
		Title     string     `json:"title"`
		HTMLURL   string     `json:"html_url"`
		User      appUser    `json:"user"`
		CreatedAt time.Time  `json:"created_at"`
		ClosedAt  *time.Time `json:"closed_at"`
		MergedAt  *time.Time `json:"merged_at"`
	}
	if err := json.Unmarshal(body, &prs); err != nil {
		return nil, fmt.Errorf("pulls: %w", err)
	}
	var out []rpc.ActivityItem
	for _, p := range prs {
		ref := strconv.Itoa(p.Number)
		add := func(state, kind string, at time.Time) {
			out = append(out, rpc.ActivityItem{
				ID: fmt.Sprintf("gh:%s:pr:%d:%s", repo, p.Number, state), Source: rpc.ActivityGitHub, Repo: repo,
				Kind: kind, Actor: p.User.Login, Bot: ghBot(p.User.Login), At: at, Ref: ref, PR: true, URL: p.HTMLURL,
				Title: fmt.Sprintf("PR #%d %s: %s", p.Number, state, p.Title),
			})
		}
		add("opened", "pr", p.CreatedAt)
		switch {
		case p.MergedAt != nil:
			add("merged", "merge", *p.MergedAt)
		case p.ClosedAt != nil:
			add("closed", "pr", *p.ClosedAt)
		}
	}
	return out, nil
}

// parseAppReviews maps GET /pulls/{n}/reviews. Pending reviews are the
// author's draft and not activity.
func parseAppReviews(repo string, pr int, body []byte) ([]rpc.ActivityItem, error) {
	var rs []struct {
		ID          int64     `json:"id"`
		User        appUser   `json:"user"`
		Body        string    `json:"body"`
		State       string    `json:"state"`
		HTMLURL     string    `json:"html_url"`
		SubmittedAt time.Time `json:"submitted_at"`
	}
	if err := json.Unmarshal(body, &rs); err != nil {
		return nil, fmt.Errorf("reviews: %w", err)
	}
	var out []rpc.ActivityItem
	for _, r := range rs {
		if r.State == "PENDING" || r.SubmittedAt.IsZero() {
			continue
		}
		state := strings.ToLower(strings.ReplaceAll(r.State, "_", " "))
		out = append(out, rpc.ActivityItem{
			ID: fmt.Sprintf("gh:%s:review:%d", repo, r.ID), Source: rpc.ActivityGitHub, Repo: repo, Kind: "review",
			Actor: r.User.Login, Bot: ghBot(r.User.Login), At: r.SubmittedAt, Ref: strconv.Itoa(pr), PR: true,
			URL: r.HTMLURL, Body: Truncate(r.Body), Title: fmt.Sprintf("review %s on PR #%d", state, pr),
		})
	}
	return out, nil
}

// parseAppIssues maps GET /issues (which also lists PRs, skipped here) to
// opened and closed items.
func parseAppIssues(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var is []struct {
		Number      int        `json:"number"`
		Title       string     `json:"title"`
		HTMLURL     string     `json:"html_url"`
		User        appUser    `json:"user"`
		CreatedAt   time.Time  `json:"created_at"`
		ClosedAt    *time.Time `json:"closed_at"`
		PullRequest any        `json:"pull_request"`
	}
	if err := json.Unmarshal(body, &is); err != nil {
		return nil, fmt.Errorf("issues: %w", err)
	}
	var out []rpc.ActivityItem
	for _, i := range is {
		if i.PullRequest != nil {
			continue
		}
		add := func(state string, at time.Time) {
			out = append(out, rpc.ActivityItem{
				ID: fmt.Sprintf("gh:%s:issue:%d:%s", repo, i.Number, state), Source: rpc.ActivityGitHub, Repo: repo,
				Kind: "issue", Actor: i.User.Login, Bot: ghBot(i.User.Login), At: at, Ref: strconv.Itoa(i.Number),
				URL: i.HTMLURL, Title: fmt.Sprintf("issue #%d %s: %s", i.Number, state, i.Title),
			})
		}
		add("opened", i.CreatedAt)
		if i.ClosedAt != nil {
			add("closed", *i.ClosedAt)
		}
	}
	return out, nil
}

// parseAppComments maps GET /issues/comments (review false) and
// GET /pulls/comments (review true). Neither carries the thread's title.
func parseAppComments(repo string, body []byte, review bool) ([]rpc.ActivityItem, error) {
	var cs []struct {
		ID        int64     `json:"id"`
		HTMLURL   string    `json:"html_url"`
		Body      string    `json:"body"`
		User      appUser   `json:"user"`
		CreatedAt time.Time `json:"created_at"`
		IssueURL  string    `json:"issue_url"`
		PullURL   string    `json:"pull_request_url"`
	}
	if err := json.Unmarshal(body, &cs); err != nil {
		return nil, fmt.Errorf("comments: %w", err)
	}
	var out []rpc.ActivityItem
	for _, c := range cs {
		it := rpc.ActivityItem{
			Source: rpc.ActivityGitHub, Repo: repo, Actor: c.User.Login, Bot: ghBot(c.User.Login),
			At: c.CreatedAt, URL: c.HTMLURL, Body: Truncate(c.Body),
		}
		if review {
			n := lastSegment(c.PullURL)
			it.ID, it.Kind, it.PR = fmt.Sprintf("gh:%s:review_comment:%d", repo, c.ID), "review_comment", true
			it.Ref, it.Title = strconv.Itoa(n), fmt.Sprintf("review comment on PR #%d", n)
		} else {
			n := lastSegment(c.IssueURL)
			what := "issue"
			if strings.Contains(c.HTMLURL, "/pull/") {
				it.PR, what = true, "PR"
			}
			it.ID, it.Kind = fmt.Sprintf("gh:%s:comment:%d", repo, c.ID), "comment"
			it.Ref, it.Title = strconv.Itoa(n), fmt.Sprintf("comment on %s #%d", what, n)
		}
		out = append(out, it)
	}
	return out, nil
}

// parseAppCommits maps GET /commits on the default branch to push items,
// one per commit: the list has no push boundaries.
func parseAppCommits(repo string, body []byte) ([]rpc.ActivityItem, error) {
	var cs []struct {
		SHA     string  `json:"sha"`
		HTMLURL string  `json:"html_url"`
		Author  appUser `json:"author"`
		Commit  struct {
			Message string `json:"message"`
			Author  struct {
				Name string    `json:"name"`
				Date time.Time `json:"date"`
			} `json:"author"`
			Committer struct {
				Date time.Time `json:"date"`
			} `json:"committer"`
		} `json:"commit"`
	}
	if err := json.Unmarshal(body, &cs); err != nil {
		return nil, fmt.Errorf("commits: %w", err)
	}
	var out []rpc.ActivityItem
	for _, c := range cs {
		subject, _, _ := strings.Cut(c.Commit.Message, "\n")
		actor := c.Author.Login
		if actor == "" {
			actor = c.Commit.Author.Name
		}
		out = append(out, rpc.ActivityItem{
			ID: fmt.Sprintf("gh:%s:commit:%s", repo, c.SHA), Source: rpc.ActivityGitHub, Repo: repo, Kind: "push",
			Actor: actor, Bot: ghBot(actor), At: c.Commit.Committer.Date, URL: c.HTMLURL,
			Title: fmt.Sprintf("commit %.7s: %s", c.SHA, subject), Body: Truncate(c.Commit.Message),
		})
	}
	return out, nil
}
