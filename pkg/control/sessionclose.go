package control

import (
	"cmp"
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel/metric"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/telemetry"
)

// resolutionClosed: ended on request by `session close` or Desk (JOB-111).
const resolutionClosed = "closed"

// grantSessionClose lets a pool close sessions in its project.
const grantSessionClose = "session-close"

// sessionCloseRPC ends a session now and frees its slot (JOB-111). The
// runner and host stop first (a running turn is cancelled), so nothing is
// claimed after the held list is read. Claimed tickets are handed back to
// their pool, or to p.Handoff, which reopens them without counting a step
// failure; endSession then unassigns anything left. The session ticket,
// journal and worktree history stay as for any ended session.
func (c *Controller) sessionCloseRPC(ctx context.Context, sc scope, p rpc.SessionCloseParams) (any, error) {
	if err := c.requireGrant(sc, grantSessionClose, "session close"); err != nil {
		return nil, err
	}
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	if p.ID == sc.session {
		// stopRunner would wait for the turn that is making this call.
		return nil, fmt.Errorf("%w: a session cannot close itself; close or release its ticket instead", domain.ErrInvalid)
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if !sc.admin && t.Project != sc.project {
		return nil, forbidden("session %s is not in project %s", p.ID, sc.project)
	}
	if t.Status == domain.StatusClosed {
		return nil, fmt.Errorf("%w: session %s", domain.ErrClosed, p.ID)
	}
	cfg := c.config()
	if p.Handoff != "" {
		if _, ok := cfg.Pools[p.Handoff]; !ok {
			return nil, fmt.Errorf("%w: unknown pool %q", domain.ErrInvalid, p.Handoff)
		}
	}
	if p.DryRun {
		held, err := c.heldBy(ctx, p.ID, p.Handoff)
		if err != nil {
			return nil, err
		}
		return rpc.SessionCloseResult{ID: p.ID, DryRun: true, Held: held}, nil
	}

	c.log.InfoContext(ctx, "session close requested", "session", p.ID, "by", sc.actor)
	c.stopRunner(p.ID)
	held, err := c.heldBy(ctx, p.ID, p.Handoff)
	if err != nil {
		return nil, err
	}
	why := cmp.Or(strings.TrimSpace(p.Reason), "no reason given")
	for _, h := range held {
		// Pre-assigned tickets were never started: endSession unassigns
		// them. A claimed one with no pool to go back to is released there
		// too, which counts a failed attempt for a step.
		if h.Pool == "" || (h.Status != domain.StatusInProgress && p.Handoff == "") {
			continue
		}
		note := fmt.Sprintf("Session %s was closed by %s (%s); handed back to pool %s.", p.ID, sc.actor, why, h.Pool)
		rt, err := c.store.Release(ctx, sc.actor, h.ID, nil, store.Outcome{Pool: h.Pool, Note: note})
		if err != nil {
			c.log.WarnContext(ctx, "session close: hand back", "session", p.ID, "ticket", h.ID, "err", err)
			continue
		}
		c.stepEvent(ctx, cfg, rt)
	}
	closed, err := store.Close(ctx, c.store, sc.actor, p.ID, nil, resolutionClosed)
	if err != nil {
		return nil, err
	}
	c.countClose(ctx, closed)
	c.metrics.SessionsEnded.Add(ctx, 1, metric.WithAttributes(telemetry.AttrPool.String(closed.Pool),
		telemetry.AttrResolution.String(resolutionClosed)))
	_, _ = store.AddComment(ctx, c.store, "controller", p.ID, fmt.Sprintf("Session closed by %s: %s", sc.actor, why))
	c.endSession(ctx, cfg, p.ID, "session closed by "+sc.actor)
	return rpc.SessionCloseResult{ID: p.ID, Held: held, Session: &closed}, nil
}

// heldBy lists what session id holds, claimed or pre-assigned, with the
// pool each would go to: handoff when set, else its own.
func (c *Controller) heldBy(ctx context.Context, id, handoff string) ([]rpc.HeldTicket, error) {
	ts, err := c.store.List(ctx, store.Filter{Holder: id, Statuses: []domain.Status{domain.StatusInProgress, domain.StatusOpen}})
	if err != nil {
		return nil, err
	}
	out := make([]rpc.HeldTicket, 0, len(ts))
	for _, t := range ts {
		out = append(out, rpc.HeldTicket{ID: t.ID, Title: t.Title, Kind: t.Kind, Status: t.Status, Pool: cmp.Or(handoff, t.Pool)})
	}
	return out, nil
}
