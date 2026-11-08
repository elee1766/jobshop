package control

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/github"
	"github.com/elee1766/jobshop/pkg/ingest"
)

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
		Name: pc.Name, Events: pc.EventSet(), Store: c.store, Now: c.now,
		Client: func(repo string) (*github.Client, error) {
			return c.githubOrHost(cfg, pc.Identity, repo), nil
		},
	}
	for _, repo := range pc.Repos {
		res, err := p.Poll(ctx, repo)
		c.queue().Appended(res.New)
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
