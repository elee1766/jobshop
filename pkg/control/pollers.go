package control

import (
	"context"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// pollerHTTP is the client for ingest pollers.
var pollerHTTP = telemetry.HTTPClient("poller", time.Minute)

// pumpPollers runs each ingest.pollers entry (#341) on its interval
// until the controller stops. It wakes at the shortest interval and polls
// the pollers that are due, so a config reload takes effect on the next
// wake without restarting goroutines.
func (c *Controller) pumpPollers() {
	last := map[string]time.Time{}
	for {
		cfg := c.config()
		wait := time.Minute
		now := c.now()
		for _, pc := range cfg.Ingest.Pollers {
			every := pc.Every()
			wait = min(wait, every)
			if now.Sub(last[pc.Name]) < every {
				continue
			}
			last[pc.Name] = now
			c.pollOnce(c.ctx, cfg, pc)
		}
		select {
		case <-c.ctx.Done():
			return
		case <-c.after(wait):
		}
	}
}

// pollOnce polls each repo of one poller; a repo's error is logged and
// does not stop the others.
func (c *Controller) pollOnce(ctx context.Context, cfg *config.Config, pc config.IngestPoller) {
	p := &ingest.Poller{
		Name: pc.Name, Events: pc.EventSet(), HTTP: pollerHTTP, Store: c.store, Now: c.now,
		Auth: func(ctx context.Context, repo string) (string, string, error) {
			return c.pollerAuth(ctx, cfg, pc.Identity, repo)
		},
	}
	for _, repo := range pc.Repos {
		res, err := p.Poll(ctx, repo)
		c.inboxAppended(res.New)
		if err != nil {
			if ctx.Err() == nil {
				c.log.WarnContext(ctx, "ingest poller", "poller", pc.Name, "repo", repo, "err", err)
			}
			continue
		}
		if res.Fresh > 0 {
			c.log.InfoContext(ctx, "ingest poller", "poller", pc.Name, "repo", repo, "items", res.Items, "fresh", res.Fresh)
		}
	}
}

// pollerAuth returns the API base and token: the named identity, else the
// identity covering repo, else the host's gh token (as activity does).
func (c *Controller) pollerAuth(ctx context.Context, cfg *config.Config, name, repo string) (string, string, error) {
	if name == "" {
		name, _ = cfg.IdentityFor(cfg.IdentityNames(), repo)
	}
	if name == "" {
		l := &liveActivity{c: c}
		return l.githubAuth(ctx, repo)
	}
	id := cfg.Identities[name]
	tok, err := c.minter.Token(ctx, name, id, repo)
	if err != nil {
		return "", "", err
	}
	base := id.APIURL
	if base == "" {
		base = githubAPI
	}
	return strings.TrimRight(base, "/"), tok.Token, nil
}
