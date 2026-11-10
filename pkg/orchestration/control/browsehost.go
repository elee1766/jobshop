package control

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/github"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/orchestration/browse"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// Desk's read-only Linear and GitHub views (JOB-163, pkg/orchestration/browse). The
// service reads through the identities the controller already holds and only
// for the teams and repos it is configured for. This file is the controller
// side: the grant, the clients, the Jobshop runs that name an item, and the
// revision that tells Desk's live stream the providers moved.

// grantBrowse lets a pool's sessions read through browse.*. Desk and the
// admin socket do not need it.
const grantBrowse = "browse"

// browseService is the browse service, made on first use.
func (c *Controller) browseService() *browse.Service {
	c.browseOnce.Do(func() { c.browse = browse.New(browseHost{c}, c.log) })
	return c.browse
}

// browseHost is the controller as browse.Host.
type browseHost struct{ c *Controller }

func (h browseHost) Config() *config.Config { return h.c.config() }
func (h browseHost) Now() time.Time         { return h.c.now() }

func (h browseHost) Linear(team string) (*linear.Client, error) {
	return h.c.linearForTeam(h.c.config(), team)
}

// GitHub is the covering identity's client. A repo that no identity covers
// is an error, not the host's gh login: Desk shows the repos the service
// identity may read, and nothing the host happens to be logged in to.
func (h browseHost) GitHub(repo string) (*github.Client, error) {
	return h.c.githubForRepo(h.c.config(), "", repo)
}

func (c *Controller) browseScope(_ context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, grantBrowse, "browse scope"); err != nil {
		return nil, err
	}
	return c.browseService().Scope(), nil
}

func (c *Controller) browseList(ctx context.Context, sc scope, p rpc.BrowseListParams) (any, error) {
	if err := c.requireGrant(sc, grantBrowse, "browse list"); err != nil {
		return nil, err
	}
	res, err := c.browseService().List(ctx, p)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Controller) browseShow(ctx context.Context, sc scope, p rpc.BrowseShowParams) (any, error) {
	if err := c.requireGrant(sc, grantBrowse, "browse show"); err != nil {
		return nil, err
	}
	d, err := c.browseService().Show(ctx, p)
	if err != nil {
		return nil, err
	}
	if p.Cursor == "" {
		d.Linked = append(d.Linked, c.browseRuns(ctx, d.Item, d.Linked)...)
	}
	return d, nil
}

func (c *Controller) browseCatalog(ctx context.Context, sc scope, p rpc.BrowseCatalogParams) (any, error) {
	if err := c.requireGrant(sc, grantBrowse, "browse catalog"); err != nil {
		return nil, err
	}
	return c.browseService().Catalog(ctx, p)
}

// browseRuns are the Jobshop runs that reference item (activityLinks),
// then the runs of the work it is tied to: a Linear issue's attached PRs,
// and the issues a PR closes or mentions. Those carry the relation
// "via <ref>". They come from the store at call time, never from the
// provider, so a link is there only when a run really exists.
func (c *Controller) browseRuns(ctx context.Context, it rpc.BrowseItem, linked []rpc.BrowseLinked) []rpc.BrowseLinked {
	links, err := c.activityLinks(ctx)
	if err != nil {
		c.log.WarnContext(ctx, "browse: run links", "ref", it.Ref, "err", err)
		return nil
	}
	key := it.Ref
	if it.Provider == rpc.BrowseGitHub {
		// activityLinks keys GitHub as repo#n.
		repo, n, ok := strings.Cut(it.Ref, "#")
		if _, err := strconv.Atoi(n); !ok || err != nil {
			return nil
		}
		key = repo + "#" + n
	}
	var out []rpc.BrowseLinked
	seen := map[string]bool{}
	add := func(key, relation string) {
		for _, id := range links[key] {
			if seen[id] {
				continue
			}
			seen[id] = true
			l := rpc.BrowseLinked{Kind: "run", Ref: id, Relation: relation}
			if t, err := c.store.Get(ctx, id); err == nil {
				l.Title, l.State = t.Title, runState(t)
			}
			out = append(out, l)
		}
	}
	add(key, "run")
	for _, l := range linked {
		if l.Kind == "pr" || l.Relation == "closes" || l.Relation == "mentions" {
			add(l.Ref, "via "+l.Ref)
		}
	}
	return out
}

// runState is the run's own state when it has one, else its ticket status.
func runState(t domain.Ticket) string {
	if f, err := domain.DecodeFields[domain.RunFields](t); err == nil && f.State != "" {
		return f.State
	}
	return fmt.Sprint(t.Status)
}

// browseChanged drops what Desk cached for the repos and teams the new
// activity items belong to. The feed polls every minute and takes webhook
// rows through the same inbox, so this is what makes an open view refresh
// without a poll of its own. The items are also what the activity event
// carries to Desk's live stream.
func (c *Controller) browseChanged(items []rpc.ActivityItem) {
	if len(items) == 0 {
		return
	}
	seen := map[string]bool{}
	for _, it := range items {
		if it.Repo == "" || seen[it.Source+it.Repo] {
			continue
		}
		seen[it.Source+it.Repo] = true
		c.browseService().Invalidate(it.Source, it.Repo)
	}
}
