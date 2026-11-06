package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/config"
	"github.com/elee1766/jobshop/pkg/ingest"
	"github.com/elee1766/jobshop/pkg/linear"
	"github.com/elee1766/jobshop/pkg/store"
)

// Linear webhook intake (#300). Linear pushes Issue and Comment events
// to POST /linear/webhook on linear.webhook_listen. The shared ingest
// handler (#334) checks the signature, age and actor and writes the
// delivery to the controller's spool, like every other receiver (#679).
// The listener only wakes the drain; the drain hands each new Linear
// inbox item to linearDelivered. The payload is only a hint: the
// controller then checks organization and actor against the identity,
// re-reads the issue and runs the polling code early. Polling stays the
// fallback, so a lost or refused event only costs latency.

// linearWebhookHandler is the ingest handler for the controller's
// listener: the Linear endpoints from ingest.endpoints and the implicit
// identity ones, writing to the same spool database as `ingest serve`.
func (c *Controller) linearWebhookHandler(sp *ingest.Spool) *ingest.Handler {
	var eps []config.IngestEndpoint
	for _, e := range c.config().IngestEndpoints() {
		if e.Source == config.IngestLinear {
			eps = append(eps, e)
		}
	}
	return &ingest.Handler{
		Spool:     sp,
		Endpoints: eps,
		Secret: func(e config.IngestEndpoint) (ingest.Secret, error) {
			return ingest.SecretFor(c.root, c.config())(e)
		},
		Now:     c.now,
		Log:     c.log,
		OnSpool: func(ingest.Delivery) { c.wakeDrain() },
	}
}

// linearDelivered runs the Linear consumer for new Linear inbox items
// from the drain, in the background: it reads Linear. The endpoint names
// the identity whose secret signed the delivery.
func (c *Controller) linearDelivered(items []store.InboxItem) {
	eps := map[string]string{}
	for _, e := range c.config().IngestEndpoints() {
		if e.Source == config.IngestLinear {
			eps[e.Name] = e.Identity
		}
	}
	for _, it := range items {
		name, ok := eps[it.Endpoint]
		if it.Source != ingest.SourceLinear || !ok {
			continue
		}
		var ev ingest.LinearEvent
		if json.Unmarshal(it.Payload, &ev) != nil {
			continue
		}
		c.wg.Go(func() {
			if err := c.linearWebhookEvent(c.ctx, name, ev); err != nil {
				c.log.Warn("linear webhook", "type", ev.Type, "issue", ev.IssueKey(), "err", err)
			}
		})
	}
}

// serveLinearWebhook serves the Linear endpoints on linear.webhook_listen
// until the controller stops. Empty listen address: off. Endpoints are
// read at start; a new one needs a controller restart.
func (c *Controller) serveLinearWebhook() error {
	addr := c.config().Linear.WebhookListen
	if addr == "" {
		return nil
	}
	sp, err := c.spool(c.ctx)
	if err != nil {
		return fmt.Errorf("linear webhook: %w", err)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("linear webhook listen: %w", err)
	}
	srv := &http.Server{Handler: c.linearWebhookHandler(sp).Mux(), ReadHeaderTimeout: 10 * time.Second}
	c.wg.Go(func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			c.log.Error("linear webhook server", "err", err)
		}
	})
	c.wg.Go(func() {
		<-c.ctx.Done()
		// Shutdown waits for handlers in flight, so none writes to a
		// closed spool.
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
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
	if ev.Type == ingest.AgentSessionEvent {
		// Linear shows the app unresponsive after 10s without a reply.
		c.sweepLinearAgentSessions(ctx, true)
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
	cl := c.newLinear(c.linearURL, name, id, team, tok.Token)
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
