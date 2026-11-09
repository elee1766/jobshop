package control

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/control/activity"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/integrations/github"
	"github.com/elee1766/jobshop/pkg/integrations/linear"
	"github.com/elee1766/jobshop/pkg/storage/store"
	"github.com/elee1766/jobshop/pkg/transport/rpc"
)

// Activity feed (#255, JOB-149). The cache, poll loop, listing filters
// and GitHub/Linear readers are pkg/control/activity. This file is the
// controller side: the RPC grant, the ticket and run links computed from
// the store when listing (so they follow meta changes), the inbox and
// ingest admission new items pass through, revert detection, and the
// identity-scoped clients the readers use.

// activitySvc is the activity service, made on first use so controllers
// built without Run get one too.
func (c *Controller) activitySvc() *activity.Service {
	c.activityOnce.Do(func() { c.activity = activity.New(activityHost{c}, c.log, c.root.Dir) })
	return c.activity
}

// pumpActivity polls until the controller stops.
func (c *Controller) pumpActivity() { c.activitySvc().Run(c.ctx, c.activitySource()) }

// activityHost is the controller as activity.Host.
type activityHost struct{ c *Controller }

func (h activityHost) Config() *config.Config                 { return h.c.config() }
func (h activityHost) Now() time.Time                         { return h.c.now() }
func (h activityHost) After(d time.Duration) <-chan time.Time { return h.c.after(d) }

func (h activityHost) Inbox(ctx context.Context, items []rpc.ActivityItem) {
	h.c.inboxActivity(ctx, items)
}

func (h activityHost) Ingest(ctx context.Context) { h.c.ingestPass(ctx) }

// Admitted checks newly admitted merges for reverts. It runs before the
// cache is saved, as before the extraction.
func (h activityHost) Admitted(ctx context.Context, items []rpc.ActivityItem) {
	h.c.browseChanged(items)
	if rs, ok := h.c.activitySource().(github.PullSource); ok {
		h.c.checkReverts(ctx, rs, items)
	}
}

func (h activityHost) Changed(now time.Time) {
	if h.c.events != nil {
		h.c.events.publish([]rpc.Event{{Kind: rpc.EventKindActivity, Time: now}})
	}
}

// activityClients is the controller as activity.Clients: the identity
// that covers a repo or team picks the token.
type activityClients struct{ c *Controller }

func (a activityClients) Config() *config.Config { return a.c.config() }
func (a activityClients) Now() time.Time         { return a.c.now() }

func (a activityClients) GitHub(repo string) *github.Client {
	return a.c.githubOrHost(a.c.config(), "", repo)
}

func (a activityClients) GitHubSelf(ctx context.Context, repo string) string {
	return a.c.githubSelf(ctx, repo)
}

func (a activityClients) Linear(team string) (*linear.Client, error) {
	return a.c.linearForTeam(a.c.config(), team)
}

// activitySource is the configured activity source: the test fake, else
// the live APIs.
func (c *Controller) activitySource() activity.Source {
	if c.activitySrc != nil {
		return c.activitySrc
	}
	return c.liveActivity()
}

// liveActivity is the activity feed readers plus the PR reads revert
// detection needs (#272), which are not part of the feed.
type liveActivity struct {
	activity.Live
	c *Controller
}

func (c *Controller) liveActivity() *liveActivity {
	return &liveActivity{Live: activity.Live{C: activityClients{c}}, c: c}
}

// PullRequest and CommitPulls read as the repo's identity through
// github.Client (JOB-159).
func (l *liveActivity) PullRequest(ctx context.Context, repo string, n int) (github.Pull, error) {
	return l.client(repo).PullRequest(ctx, repo, n)
}

func (l *liveActivity) CommitPulls(ctx context.Context, repo, sha string) ([]int, error) {
	return l.client(repo).CommitPulls(ctx, repo, sha)
}

func (l *liveActivity) client(repo string) *github.Client {
	return l.c.githubOrHost(l.c.config(), "", repo)
}

// truncate bounds feed text to what the feed stores.
func truncate(s string) string { return activity.Truncate(s) }

// addActivity appends one controller-made item (a recorded `jobshop gh`
// call, a redaction) to the feed, if one is configured.
func (c *Controller) addActivity(it rpc.ActivityItem) { c.activitySvc().Record(it) }

// activityLinks maps "repo#n" and Linear keys to the runs and tickets
// that reference them. A step stands for its run.
func (c *Controller) activityLinks(ctx context.Context) (map[string][]string, error) {
	all, err := c.store.List(ctx, store.Filter{Limit: math.MaxInt32})
	if err != nil {
		return nil, err
	}
	out := map[string][]string{}
	addLink := func(k, id string) {
		if !slices.Contains(out[k], id) {
			out[k] = append(out[k], id)
		}
	}
	for _, t := range all {
		if t.Kind != domain.KindRun && t.Kind != domain.KindStep && t.Kind != domain.KindTask {
			continue
		}
		var m struct {
			GitHub struct {
				Repo  string `json:"repo"`
				Issue any    `json:"issue"`
				PR    any    `json:"pr"`
			} `json:"github"`
			Linear struct {
				Issue string `json:"issue"`
			} `json:"linear"`
		}
		if len(t.Meta) == 0 || json.Unmarshal(t.Meta, &m) != nil {
			continue
		}
		id := t.ID
		if t.Kind == domain.KindStep && t.ParentID != "" {
			id = t.ParentID
		}
		if m.GitHub.Repo != "" {
			for _, v := range []any{m.GitHub.Issue, m.GitHub.PR} {
				if n := github.RefNumber(v); n > 0 {
					addLink(fmt.Sprintf("%s#%d", m.GitHub.Repo, n), id)
				}
			}
		}
		if m.Linear.Issue != "" {
			addLink(m.Linear.Issue, id)
		}
	}
	return out, nil
}

func (c *Controller) activityList(ctx context.Context, sc scope, p rpc.ActivityListParams) (any, error) {
	if err := c.requireGrant(sc, "activity", "activity list"); err != nil {
		return nil, err
	}
	if !activity.ValidBots(p.Bots) {
		return nil, fmt.Errorf("%w: bots must be humans, bots or all", domain.ErrInvalid)
	}
	cfg := c.config()
	res := rpc.ActivityListResult{Items: []rpc.ActivityItem{}, Sources: []string{}}
	for _, r := range cfg.ActivityRepos() {
		res.Sources = append(res.Sources, rpc.ActivityGitHub+":"+r)
	}
	for _, t := range cfg.ActivityTeams() {
		res.Sources = append(res.Sources, rpc.ActivityLinear+":"+t)
	}
	if cfg.Activity == nil {
		return res, nil
	}
	if !c.activitySvc().Loaded() {
		return res, nil // nothing polled yet
	}
	links, err := c.activityLinks(ctx)
	if err != nil {
		return nil, err
	}
	res.Items = c.activitySvc().List(p, links, c.now())
	return res, nil
}
