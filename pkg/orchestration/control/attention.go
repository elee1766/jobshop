package control

import (
	"context"
	"math"
	"time"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/conversation/attention"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/orchestration/triggers"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// The attention projection lives in pkg/conversation/attention (JOB-147).
// This file is the controller's side: the grant, and the Source that
// hands the projection the tickets and the subsystem rows.

const (
	questionLabel = attention.QuestionLabel
	proposedLabel = attention.ProposedLabel
)

// attention lists only what waits on a human, across all projects,
// oldest first (#185). Empty when healthy.
func (c *Controller) attention(ctx context.Context, sc scope, _ struct{}) (any, error) {
	if err := c.requireGrant(sc, "attention", "attention"); err != nil {
		return nil, err
	}
	items, err := c.attentionItems(ctx, c.config())
	if err != nil {
		return nil, err
	}
	return rpc.AttentionResult{Items: items}, nil
}

func (c *Controller) attentionItems(ctx context.Context, cfg *config.Config) ([]rpc.AttentionItem, error) {
	return attention.Build(ctx, attentionSource{c, cfg})
}

// attentionSource is the controller as attention.Source. cfg is the
// config the caller holds, so one pass reads one config.
type attentionSource struct {
	c   *Controller
	cfg *config.Config
}

func (s attentionSource) Now() time.Time { return s.c.now() }

func (s attentionSource) Ticket(ctx context.Context, id string) (domain.Ticket, error) {
	return s.c.store.Get(ctx, id)
}

func (s attentionSource) Steps(ctx context.Context, run string) ([]domain.Ticket, error) {
	return store.Children(ctx, s.c.store, run)
}

func (s attentionSource) LabeledOpen(ctx context.Context, label string) ([]domain.Ticket, error) {
	return triggers.LabeledOpen(ctx, s.c.store, label)
}

func (s attentionSource) NeedsHumanRuns(ctx context.Context) ([]domain.Ticket, error) {
	return s.c.store.List(ctx, store.Filter{Kind: domain.KindRun, State: domain.RunNeedsHuman, Order: store.OrderNewest, Limit: math.MaxInt32})
}

func (s attentionSource) TimedOut(ctx context.Context) ([]attention.Waiting, error) {
	ts, err := s.c.waiting(ctx, "")
	if err != nil {
		return nil, err
	}
	var out []attention.Waiting
	for _, t := range ts {
		if w := waitOf(t); !w.Timeout.IsZero() && !s.c.now().Before(w.Timeout) {
			out = append(out, attention.Waiting{Ticket: t, Wait: w})
		}
	}
	return out, nil
}

// Subsystems are the rows the resource, ingest and journal shipping
// subsystems own.
func (s attentionSource) Subsystems(ctx context.Context) ([]rpc.AttentionItem, error) {
	items, err := s.c.resourceAttention(ctx, s.cfg)
	if err != nil {
		return nil, err
	}
	ing, err := s.c.ingestAttention(ctx, s.cfg)
	if err != nil {
		return nil, err
	}
	items = append(items, ing...)
	return append(items, s.c.shipper().LagAttention(s.cfg)...), nil
}

func (s attentionSource) Stale(ctx context.Context) ([]rpc.AttentionItem, error) {
	return s.c.staleAttention(ctx, s.cfg)
}

func (s attentionSource) Health() attention.Health {
	c := s.c
	c.mu.Lock()
	defer c.mu.Unlock()
	h := attention.Health{
		Started: c.started, ConfigErr: c.configErr, ConfigErrAt: c.configErrAt, WatchErr: c.watchErr, WatchErrAt: c.watchErrAt,
		Quarantined: make(map[string]string, len(c.poolQuarantined)), PausedFor: poolPause,
	}
	for p, why := range c.poolQuarantined {
		h.Quarantined[p] = why
	}
	now := c.now()
	for p, until := range c.poolPausedUntil {
		if now.Before(until) {
			h.Paused = append(h.Paused, attention.Pause{Pool: p, Until: until, Why: c.poolPauseWhy(p)})
		}
	}
	return h
}

// ticketItem lists t as an attention row of kind (see attention.TicketItem).
func ticketItem(kind string, t domain.Ticket) rpc.AttentionItem { return attention.TicketItem(kind, t) }

const maxSummary = attention.MaxSummary
