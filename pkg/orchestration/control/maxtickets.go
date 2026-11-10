package control

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"

	"github.com/gfx-labs/jobshop/pkg/config"
	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Single-use sessions (JOB-189). A ticket pool with max_tickets > 0 runs
// sessions that close that many tickets and end ticket_limit. While a
// session's ticket waits on a wait with resume, or a timed defer the
// session made, the session is suspended: its host is stopped and it
// counts toward neither max_sessions nor limits.max_sessions. The
// supervisor resumes it (session/load) when its ticket is ready, oldest
// first. A resume may take the pool one over max_sessions, so a finished
// wait does not queue behind new work, but never past the host limit.

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

// countedSessions is how many of open take a slot: a suspended session
// takes none, whatever suspended it (JOB-189).
func countedSessions(open []domain.Ticket) int {
	n := 0
	for _, t := range open {
		if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil && f.Suspended {
			continue
		}
		n++
	}
	return n
}

// hostCounted is the sessions that count toward limits.max_sessions: open,
// not suspended, in a configured ticket pool (the scaler's own count).
func (c *Controller) hostCounted(ctx context.Context, cfg *config.Config) int {
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusOpen})
	if err != nil {
		return 0
	}
	n := 0
	for _, t := range open {
		if p, ok := cfg.Pools[t.Pool]; ok && !p.Chat() {
			n += countedSessions([]domain.Ticket{t})
		}
	}
	return n
}

// ticketsWorked is how many tickets the session has closed.
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

// markClaimed records the ticket the session just claimed. It counts when
// settleClaimed finds it closed.
func (r *runner) markClaimed(ctx context.Context, id string) {
	err := r.c.updateSessionFields(ctx, r.id, func(f *domain.SessionFields) { f.Claimed = id })
	if err != nil {
		r.log.WarnContext(ctx, "record claimed ticket", "ticket", id, "err", err)
	}
}

// settleClaimed counts the claimed ticket against max_tickets once it is
// closed, with any resolution. A ticket the session released (a wait, a
// defer, a drop) counts nothing, and ends a resumed session's exemption
// from being cordoned as surplus. It reports whether the session released
// the ticket, so it may park on it.
func (r *runner) settleClaimed(ctx context.Context) (released bool) {
	t, err := r.c.store.Get(ctx, r.id)
	if err != nil {
		return false
	}
	f, err := domain.DecodeFields[domain.SessionFields](t)
	if err != nil || f.Claimed == "" {
		return false
	}
	ct, err := r.c.store.Get(ctx, f.Claimed)
	if err == nil && ct.Status == domain.StatusInProgress && ct.Holder == r.id {
		return false // still working it
	}
	closed := err == nil && ct.Status == domain.StatusClosed
	err = r.c.updateSessionFields(ctx, r.id, func(f *domain.SessionFields) {
		if closed {
			f.Tickets++
		}
		f.Claimed, f.Resumed = "", false
	})
	if err != nil {
		r.log.WarnContext(ctx, "count closed ticket", "ticket", f.Claimed, "err", err)
	}
	return !closed
}

// parkedTicket is the open ticket this session is waiting for: one that
// waits with resume for it, or one it deferred to a future time. "" if none.
func (r *runner) parkedTicket(ctx context.Context) string {
	kinds := []domain.Kind{domain.KindTask, domain.KindStep}
	// A wait with resume names the session in the ticket's fields.
	ts, err := r.c.store.List(ctx, store.Filter{
		Project: r.project(), Kinds: kinds, Status: domain.StatusOpen, Query: r.id, Limit: store.NoLimit,
	})
	if err != nil {
		return ""
	}
	for _, t := range ts {
		if w := waitOf(t); w != nil && w.Resume && w.Session == r.id {
			return t.ID
		}
	}
	// A timed defer the session made is assigned back to it.
	ts, err = r.c.store.List(ctx, store.Filter{Project: r.project(), Kinds: kinds, Status: domain.StatusOpen, Holder: r.id})
	if err != nil {
		return ""
	}
	now := r.c.now()
	for _, t := range ts {
		if waitOf(t) == nil && t.DeferUntil.After(now) {
			return t.ID
		}
	}
	return ""
}

// park suspends the session on ticket: it stops the host and keeps the
// session open and out of the slot counts until the supervisor resumes it.
func (r *runner) park(nc net.Conn, ticket string) error {
	err := r.c.updateSessionFields(r.ctx, r.id, func(f *domain.SessionFields) {
		f.Suspended, f.Parked, f.Claimed, f.Resumed = true, ticket, "", false
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
		// session): nothing is assigned to this one.
		return fmt.Sprintf("parked ticket %s is no longer assigned to this session", ticket)
	}
	return ""
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

// resumeParked takes a session out of suspension; its runner starts the
// host again and loads the conversation. Resumed keeps the scaler from
// cordoning it as surplus while the pool is over max_sessions.
func (c *Controller) resumeParked(ctx context.Context, id string) error {
	err := c.updateSessionFields(ctx, id, func(f *domain.SessionFields) {
		f.Suspended, f.Parked, f.Resumed = false, "", true
	})
	if err == nil {
		c.wakeRunner(id)
	}
	return err
}

// readyParked is a parked session whose ticket is ready to claim.
type readyParked struct {
	parkedSession
	since int64 // when the ticket became ready, for oldest first
}

// resumeReady resumes the parked sessions of a pool whose ticket is ready,
// oldest first. count is the pool's sessions that take a slot. A resume
// may take the pool to max_sessions + 1, not past it, and hostFree is the
// sessions the host limit still allows (-1 for no limit). It returns how
// many it resumed.
func (c *Controller) resumeReady(ctx context.Context, name string, p config.Pool, count, hostFree int, parked []parkedSession) int {
	var ready []readyParked
	for _, ps := range parked {
		ts, err := c.store.List(ctx, store.Filter{Project: p.Project, Holder: ps.ID, Ready: true, Limit: 1})
		if err != nil || len(ts) == 0 {
			continue
		}
		ready = append(ready, readyParked{ps, ts[0].UpdatedAt.UnixNano()})
	}
	slices.SortStableFunc(ready, func(a, b readyParked) int {
		if a.since != b.since {
			return int(a.since - b.since)
		}
		return compareStrings(a.ID, b.ID)
	})
	n := 0
	for _, rp := range ready {
		if p.MaxSessions > 0 && count+n >= p.MaxSessions+1 {
			break
		}
		if hostFree >= 0 && n >= hostFree {
			break
		}
		if err := c.resumeParked(ctx, rp.ID); err != nil {
			c.log.WarnContext(ctx, "resume parked session", "pool", name, "session", rp.ID, "err", err)
			continue
		}
		c.scaleEvent(ctx, name, "resume", rp.ID, c.poolScale(name).desired)
		n++
	}
	return n
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	}
	return 0
}

// assignDeferred pre-assigns a ticket a single-use session deferred to a
// future time back to it, so the session can park on it (JOB-189). Other
// pools and plain sessions are unchanged.
func (c *Controller) assignDeferred(ctx context.Context, t domain.Ticket, session string) {
	if session == "" || t.Status != domain.StatusOpen {
		return
	}
	s, err := c.store.Get(ctx, session)
	if err != nil || s.Kind != domain.KindSession || s.Status != domain.StatusOpen {
		return
	}
	if p, ok := c.config().Pools[s.Pool]; !ok || p.MaxTickets == 0 {
		return
	}
	if _, err := c.store.Update(ctx, "controller", t.ID, store.Update{Holder: &session}); err != nil {
		c.log.WarnContext(ctx, "assign deferred ticket to its session", "ticket", t.ID, "session", session, "err", err)
	}
}
