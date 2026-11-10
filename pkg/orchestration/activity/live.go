package activity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Clients is what Live needs from the controller: the clients for the
// identity that covers a repo or team (identity scopes stay the
// controller's), the host's own login and the clock.
type Clients interface {
	Config() *config.Config
	Now() time.Time
	// GitHub is the client for repo: the covering identity's, else the
	// host's gh login.
	GitHub(repo string) *github.Client
	// GitHubSelf is the login the covering identity acts as, "" when it is
	// an App or unknown.
	GitHubSelf(ctx context.Context, repo string) string
	// Linear is the client for the identity covering team.
	Linear(team string) (*linear.Client, error)
}

// Live reads GitHub over HTTP (ETag, X-Poll-Interval) and Linear
// with pkg/integrations/linear.
type Live struct {
	C Clients
}

// get fetches path with If-None-Match. A 304 returns nil body, no error.
func (l *Live) get(ctx context.Context, cl *github.Client, path, etag string) ([]byte, http.Header, error) {
	resp, err := cl.Do(ctx, github.Request{Path: path, ETag: etag, Route: "/repos/{repo}/activity"})
	if err != nil {
		return nil, nil, err
	}
	if resp.NotModified() {
		return nil, resp.Header, nil
	}
	return resp.Body, resp.Header, nil
}

// GitHub reads repo as its covering identity and marks that identity's own
// events as bot.
func (l *Live) GitHub(ctx context.Context, repo string, st *RepoPoll) ([]rpc.ActivityItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cl := l.C.GitHub(repo)
	items, err := l.fetchGitHub(ctx, cl, repo, st)
	// A github-token acts as a user, not a [bot]: mark its own events (#340).
	if self := l.C.GitHubSelf(ctx, repo); self != "" {
		for i := range items {
			if strings.EqualFold(items[i].Actor, self) {
				items[i].Bot = true
			}
		}
	}
	return items, err
}

// fetchGitHub reads repo events and completed default-branch runs, each
// with its ETag: an unchanged repo is two 304s, which cost no rate limit.
// A github-app identity reads the per-resource endpoints instead of
// events, which are empty for an installation token (JOB-108).
func (l *Live) fetchGitHub(ctx context.Context, cl *github.Client, repo string, st *RepoPoll) ([]rpc.ActivityItem, error) {
	var out []rpc.ActivityItem
	if st.DefaultBranch == "" {
		b, _, err := l.get(ctx, cl, "/repos/"+repo, "")
		if err != nil {
			return nil, err
		}
		var r struct {
			DefaultBranch string `json:"default_branch"`
		}
		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
		st.DefaultBranch = r.DefaultBranch
	}
	var appErr error
	if l.appActivity(ctx, cl, repo) {
		out, appErr = l.fetchAppActivity(ctx, cl, repo, st)
	} else {
		body, h, err := l.get(ctx, cl, "/repos/"+repo+"/events?per_page=100", st.EventsETag)
		if err != nil {
			return nil, err
		}
		if s, err := strconv.Atoi(h.Get("X-Poll-Interval")); err == nil && s > 0 {
			st.Next = l.C.Now().Add(time.Duration(s) * time.Second)
		}
		if body != nil {
			items, err := parseRepoEvents(repo, body)
			if err != nil {
				return nil, err
			}
			out = append(out, items...)
			st.EventsETag = h.Get("ETag")
		}
	}
	u := fmt.Sprintf("/repos/%s/actions/runs?status=completed&per_page=30&branch=%s", repo, url.QueryEscape(st.DefaultBranch))
	body, h, err := l.get(ctx, cl, u, st.RunsETag)
	if err != nil {
		return out, err
	}
	if body != nil {
		items, err := parseWorkflowRuns(repo, body)
		if err != nil {
			return out, err
		}
		out = append(out, items...)
		st.RunsETag = h.Get("ETag")
	}
	return out, appErr
}

// Linear reads the changes in team since the given time.
func (l *Live) Linear(ctx context.Context, team string, since time.Time) ([]rpc.ActivityItem, error) {
	cl, err := l.C.Linear(team)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	chs, err := cl.Changes(ctx, team, since)
	if err != nil {
		return nil, err
	}
	out := make([]rpc.ActivityItem, 0, len(chs))
	for _, ch := range chs {
		it := rpc.ActivityItem{
			ID: "linear:" + ch.ItemID, Source: rpc.ActivityLinear, Repo: team, Kind: ch.Kind,
			Actor: ch.Actor, Bot: ch.Bot || ch.Self, At: ch.At, URL: ch.URL, Ref: ch.Issue,
		}
		switch ch.Kind {
		case "comment":
			it.Title, it.Body = ch.Issue+": "+ch.Title, Truncate(ch.Text)
		case "issue":
			it.Title = ch.Issue + " created: " + ch.Title
		default:
			it.Title = fmt.Sprintf("%s %s %s: %s", ch.Issue, ch.Kind, ch.Text, ch.Title)
		}
		out = append(out, it)
	}
	return out, nil
}
