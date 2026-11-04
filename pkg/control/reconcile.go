package control

import (
	"context"
	"net/http"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/identity"
	"github.com/elee1766/jobshop/pkg/ingest"
)

// reconcileHTTP is the client for the deliveries API.
var reconcileHTTP = &http.Client{Timeout: 30 * time.Second}

// pumpReconcile runs the GitHub App delivery reconcile (#335) every
// ingest.github_reconcile until the controller stops. Each pass covers
// deliveries since the previous pass (the first looks back one
// interval), so a delivery that failed while the controller was down
// for less than that is still found.
func (c *Controller) pumpReconcile() {
	for {
		every := c.config().Ingest.ReconcileEvery()
		if every > 0 {
			c.reconcileOnce(c.ctx, every)
		} else {
			every = time.Minute // re-read the config
		}
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(every):
		}
	}
}

// reconcileOnce runs one pass for every github-app identity.
func (c *Controller) reconcileOnce(ctx context.Context, every time.Duration) {
	now := c.now()
	since := c.reconcileAt
	if since.IsZero() {
		since = now.Add(-every)
	}
	cfg := c.config()
	ok := true
	for _, name := range cfg.IdentityNames() {
		id := cfg.Identities[name]
		if id.Kind != config.IdentityGitHubApp {
			continue
		}
		r := &ingest.Reconcile{
			HTTP: reconcileHTTP, APIURL: id.APIURL,
			JWT: func() (string, error) { return identity.JWTFor(id, c.now()) },
		}
		res, err := r.Pass(ctx, since)
		if err != nil {
			ok = false
			if ctx.Err() == nil {
				c.log.Warn("github reconcile", "identity", name, "err", err)
			}
			continue
		}
		if res.Failed > 0 {
			c.log.Info("github reconcile", "identity", name, "failed", res.Failed, "redelivered", res.Redelivered)
		}
	}
	// A failed pass keeps the old cutoff, so the next pass retries it.
	if ok {
		c.reconcileAt = now
	}
}
