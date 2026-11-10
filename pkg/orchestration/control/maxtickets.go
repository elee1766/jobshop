package control

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Single-use sessions (JOB-189). A ticket pool with max_tickets > 0 runs
// sessions that work that many tickets and end ticket_limit. While a
// session's ticket is parked on a wait with resume, the session is
// suspended: its host is stopped, it takes no max_sessions slot, and the
// supervisor resumes it (session/load) once the ticket is ready and the
// pool has a free slot. The wait pre-assigns the ticket to the session, so
// the resumed session claims it first and counts it once.

const (
	resolutionTicketLimit = "ticket_limit"
	// resolutionParkedGone ends a suspended session whose parked ticket was
	// closed, deleted or handed to another session.
	resolutionParkedGone = "parked_gone"
)

// parkedSession is a suspended single-use session and the ticket it waits on.
type parkedSession struct {
	ID, Ticket string
}

// ticketsWorked is how many tickets the session has counted.
func (r *runner) ticketsWorked(ctx context.Context) int {
	t, err := r.c.store.Get(ctx, r.id)
	if err != nil {
		return 0
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil {
		return 0
	}
	return f.Tickets
}

// countTicket counts a claimed ticket against max_tickets, once per
// ticket: a host restart or a resume claims the same one again.
func (r *runner) countTicket(ctx context.Context, id string) {
	err := r.c.updateSessionFields(ctx, r.id, func(f *domain.SessionFields) {
		if f.LastTicket != id {
			f.Tickets++
			f.LastTicket = id
		}
	})
	if err != nil {
		r.log.WarnContext(ctx, "count ticket", "ticket", id, "err", err)
	}
}

// parkedTicket is the open ticket that waits with resume for this
// session, or "". The query matches the session id in the ticket fields.
func (r *runner) parkedTicket(ctx context.Context) string {
	ts, err := r.c.store.List(ctx, store.Filter{
		Project: r.project(), Kinds: []domain.Kind{domain.KindTask, domain.KindStep},
		Status: domain.StatusOpen, Query: r.id,
	})
	if err != nil {
		return ""
	}
	for _, t := range ts {
		if w := waitOf(t); w != nil && w.Resume && w.Session == r.id {
			return t.ID
		}
	}
	return ""
}

// park suspends the session on ticket: it stops the host and keeps the
// session open and out of the slot count until the supervisor resumes it.
func (r *runner) park(nc net.Conn, ticket string) error {
	err := r.c.updateSessionFields(r.ctx, r.id, func(f *domain.SessionFields) {
		f.Suspended, f.Parked = true, ticket
	})
	if err != nil {
		return err
	}
	r.log.InfoContext(r.ctx, "single-use session parked", "ticket", ticket)
	r.lifecycle("park")
	_ = nc.Close()
	r.c.stopHost(r.dir)
	return errSuspended
}

// parkedGone says why a parked session has nothing to resume for; ""
// while the ticket still waits or is the session's to claim.
func (r *runner) parkedGone(ctx context.Context, ticket string) string {
	t, err := r.c.store.Get(ctx, ticket)
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return fmt.Sprintf("parked ticket %s is gone", ticket)
	case err != nil:
		return ""
	case t.Status == domain.StatusClosed:
		return fmt.Sprintf("parked ticket %s was closed (%s)", ticket, t.Resolution)
	}
	if w := waitOf(t); w != nil {
		if w.Resume && w.Session == r.id {
			return ""
		}
		return fmt.Sprintf("parked ticket %s no longer waits to resume this session", ticket)
	}
	if t.Holder != r.id {
		// The wait ended with the ticket back in the pool (or on another
		// session): nothing is pre-assigned to this one.
		return fmt.Sprintf("parked ticket %s is no longer assigned to this session", ticket)
	}
	return ""
}

// parkedReady: a ticket pre-assigned to the session is ready to claim.
func (c *Controller) parkedReady(ctx context.Context, project, session string) bool {
	ts, err := c.store.List(ctx, store.Filter{Project: project, Holder: session, Ready: true, Limit: 1})
	return err == nil && len(ts) > 0
}

// resumeParked takes a session out of suspension; its runner starts the
// host again and loads the conversation.
func (c *Controller) resumeParked(ctx context.Context, id string) error {
	err := c.updateSessionFields(ctx, id, func(f *domain.SessionFields) { f.Suspended, f.Parked = false, "" })
	if err == nil {
		c.wakeRunner(id)
	}
	return err
}

// resumeReady resumes the parked sessions of a pool whose ticket is ready,
// while active (the pool's sessions that take a slot) is below
// max_sessions. It returns how many it resumed.
func (c *Controller) resumeReady(ctx context.Context, name string, p config.Pool, active int, parked []parkedSession) int {
	n := 0
	for _, ps := range parked {
		if p.MaxSessions > 0 && active+n >= p.MaxSessions {
			break
		}
		if !c.parkedReady(ctx, p.Project, ps.ID) {
			continue
		}
		if err := c.resumeParked(ctx, ps.ID); err != nil {
			c.log.WarnContext(ctx, "resume parked session", "pool", name, "session", ps.ID, "err", err)
			continue
		}
		c.scaleEvent(ctx, name, "resume", ps.ID, c.poolScale(name).desired)
		n++
	}
	return n
}

// parkEnd says whether a parked session should end instead of waiting or
// resuming: its pool is drained, or the ticket it waits on is gone.
func (r *runner) parkEnd(ctx context.Context, f domain.SessionFields) (errSessionEnd, bool) {
	if !f.Suspended || f.Parked == "" {
		return errSessionEnd{}, false
	}
	if r.c.poolDrained(r.pool) {
		return errSessionEnd{resolution: resolutionDrained}, true
	}
	if why := r.parkedGone(ctx, f.Parked); why != "" {
		return errSessionEnd{resolution: resolutionParkedGone, reason: why}, true
	}
	return errSessionEnd{}, false
}
