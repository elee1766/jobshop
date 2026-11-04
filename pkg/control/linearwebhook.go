package control

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/linear"
)

// Linear webhook intake (#300). Linear pushes Issue and Comment events
// to POST /linear/webhook on linear.webhook_listen. The shared ingest
// handler (#334) checks the signature, age and actor and writes the
// delivery to the controller's spool. The payload is only a hint: the
// controller then checks organization and actor against the identity,
// re-reads the issue and runs the polling code early. Polling stays the
// fallback, so a lost or refused event only costs latency. Until the
// spool drain lands, this in-memory step runs right after the write.

// linearWebhookHandler is the ingest handler for the controller's
// listener: Linear only, spooling under ingest.spool_dir.
func (c *Controller) linearWebhookHandler() *ingest.Handler {
	return &ingest.Handler{
		Spool: ingest.Spool{Dir: c.config().Ingest.Spool(c.root)},
		Linear: func() []ingest.Secret {
			return ingest.LinearSecrets(c.root, c.config(), func(name string, err error) {
				c.log.Warn("linear webhook: secret", "identity", name, "err", err)
			})
		},
		Now: c.now,
		Log: c.log,
		OnLinear: func(name string, ev ingest.LinearEvent) {
			c.wg.Go(func() {
				if err := c.linearWebhookEvent(c.ctx, name, ev); err != nil {
					c.log.Warn("linear webhook", "type", ev.Type, "issue", ev.IssueKey(), "err", err)
				}
			})
		},
	}
}

// serveLinearWebhook serves POST /linear/webhook on linear.webhook_listen
// until the controller stops. Empty listen address: off.
func (c *Controller) serveLinearWebhook() error {
	addr := c.config().Linear.WebhookListen
	if addr == "" {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("linear webhook listen: %w", err)
	}
	srv := &http.Server{Handler: c.linearWebhookHandler().Mux(), ReadHeaderTimeout: 10 * time.Second}
	c.wg.Go(func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.log.Error("linear webhook server", "err", err)
		}
	})
	c.wg.Go(func() {
		<-c.ctx.Done()
		_ = srv.Close()
	})
	c.log.Info("serving linear webhook", "addr", ln.Addr().String())
	return nil
}

// linearWebhookEvent checks a spooled event against the identity that
// signed it and runs the polling path for its issue. name is "" for the
// plain ingest.linear_secret_file: no identity to read Linear as, so
// polling handles it.
func (c *Controller) linearWebhookEvent(ctx context.Context, name string, ev ingest.LinearEvent) error {
	if name == "" || !ev.Wanted() {
		return nil
	}
	key := ev.IssueKey()
	team, ok := linear.Team(key)
	if !ok {
		return nil // polling picks it up
	}
	cfg := c.config()
	id := cfg.Identities[name]
	if !id.CoversTeam(team) {
		return nil
	}
	tok, err := c.minter.Token(ctx, name, id, team)
	if err != nil {
		return err
	}
	lctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cl := linear.New(c.linearURL, tok.Token)
	me, org, err := cl.Me(lctx)
	if err != nil {
		return err
	}
	if ev.OrganizationID != org || ev.Actor.ID == me || ev.Data.UserID == me {
		return nil
	}
	// Never trust the payload: the issue must exist as the identity
	// reads it.
	is, err := cl.GetIssue(lctx, key)
	if err != nil {
		return err
	}
	switch ev.Type {
	case "Issue":
		c.linearIntakeInvalidate(is.Team.Key)
		// Same pass the tick runs, so the stall fires now.
		c.tickMu.Lock()
		if !c.draining() {
			c.runTriggers()
		}
		c.tickMu.Unlock()
	case "Comment":
		c.pollSubscriptionThread(ctx, thread{Linear: is.Identifier})
	}
	return nil
}

// linearIntakeInvalidate drops cached linear_issue_no_run reads of team,
// so the next tick reads Linear again.
func (c *Controller) linearIntakeInvalidate(team string) {
	c.linIntake.mu.Lock()
	defer c.linIntake.mu.Unlock()
	for k := range c.linIntake.cache {
		if t, _, _ := strings.Cut(k, "|"); strings.EqualFold(t, team) {
			delete(c.linIntake.cache, k)
		}
	}
}
