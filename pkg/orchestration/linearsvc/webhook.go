package linearsvc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/integrations/linear"
	"github.com/gfx-labs/jobshop/pkg/orchestration/ingest"
)

// Linear webhook intake (#300). Linear pushes Issue and Comment events
// to POST /linear/webhook on linear.webhook_listen. The shared ingest
// handler (#334) checks the signature, age and actor and writes the
// delivery to the controller's spool, like every other receiver (#679).
// The listener only wakes the drain; the drain hands each new Linear
// inbox item to Delivered. The payload is only a hint: the service then
// checks organization and actor against the identity, re-reads the issue
// and runs the polling code early. Polling stays the fallback, so a lost
// or refused event only costs latency.

// Handler is the ingest handler for the controller's listener: the Linear
// endpoints from ingest.endpoints and the implicit identity ones, writing
// to the same spool database as `ingest serve`.
func (s *Service) Handler(sp *ingest.Spool) *ingest.Handler {
	var eps []config.IngestEndpoint
	for _, e := range s.host.Config().IngestEndpoints() {
		if e.Source == config.IngestLinear {
			eps = append(eps, e)
		}
	}
	return &ingest.Handler{
		Spool:     sp,
		Endpoints: eps,
		Secret:    s.host.IngestSecret,
		Now:       s.host.Now,
		Log:       s.log(),
		OnSpool:   func(ingest.Delivery) { s.host.WakeDrain() },
	}
}

// Delivered runs the Linear consumer for new Linear inbox items from the
// drain, in the background: it reads Linear. The endpoint names the
// identity whose secret signed the delivery.
func (s *Service) Delivered(items []domain.InboxItem) {
	eps := map[string]string{}
	for _, e := range s.host.Config().IngestEndpoints() {
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
		s.host.Go(func() {
			if err := s.Event(s.host.Context(), name, ev); err != nil {
				s.log().Warn("linear webhook", "type", ev.Type, "issue", ev.IssueKey(), "err", err)
			}
		})
	}
}

// Serve serves the Linear endpoints on linear.webhook_listen until the
// controller stops. Empty listen address: off. Endpoints are read at
// start; a new one needs a controller restart.
func (s *Service) Serve() error {
	addr := s.host.Config().Linear.WebhookListen
	if addr == "" {
		return nil
	}
	ctx := s.host.Context()
	sp, err := s.host.Spool(ctx)
	if err != nil {
		return fmt.Errorf("linear webhook: %w", err)
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("linear webhook listen: %w", err)
	}
	srv := &http.Server{Handler: s.Handler(sp).Mux(), ReadHeaderTimeout: 10 * time.Second}
	s.host.Go(func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log().Error("linear webhook server", "err", err)
		}
	})
	s.host.Go(func() {
		<-ctx.Done()
		// Shutdown waits for handlers in flight, so none writes to a
		// closed spool.
		sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	})
	s.log().Info("serving linear webhook", "addr", ln.Addr().String())
	return nil
}

// Event checks a spooled event against the identity that signed it and
// runs the polling path for its issue. name is "" for the plain
// ingest.linear_secret_file: no identity to read Linear as, so polling
// handles it.
func (s *Service) Event(ctx context.Context, name string, ev ingest.LinearEvent) error {
	if name == "" || !ev.Wanted() {
		return nil
	}
	cfg := s.host.Config()
	if ev.Type == ingest.AgentSessionEvent {
		// Linear shows the app unresponsive after 10s without a reply.
		s.SweepSessions(ctx, cfg, true)
		return nil
	}
	key := ev.IssueKey()
	team, ok := linear.Team(key)
	if !ok {
		return nil // polling picks it up
	}
	id := cfg.Identities[name]
	if !id.CoversTeam(team) {
		return nil
	}
	lctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	cl := s.host.Client(name)
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
		s.InvalidateIntake(is.Team.Key)
		// Same pass the tick runs, so the finding fires now.
		s.host.RunTriggers(ctx)
	case "Comment":
		s.host.PollThread(ctx, is.Identifier)
	}
	return nil
}
