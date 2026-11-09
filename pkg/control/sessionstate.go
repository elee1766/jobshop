package control

import (
	"context"
	"math"
	"slices"

	host "github.com/elee1766/acphost/pkg/acphost"

	"github.com/elee1766/jobshop/pkg/control/sessionrt"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// Session list state (JOB-32). session.list used to walk every session
// directory and read the exit file and journal tail of every closed
// session on each request. A closed session's host state no longer
// changes, so it is saved on the ticket when the session ends, and the
// directory walk for orphaned hosts runs once per supervisor pass.
// Liveness of open sessions is still read at request time: a stored
// value goes stale on a host crash or controller restart.

// endedBackfillBatch bounds how many old closed sessions one pass fills.
const endedBackfillBatch = 100

// hostEnded reads a stopped host's state from its session directory.
func (c *Controller) hostEnded(id string) domain.SessionEnded {
	dir := c.root.SessionDir(id)
	e := domain.SessionEnded{HostPid: sessionrt.HostPid(dir)}
	if ex, ok := host.ReadExit(dir); ok {
		e.Exit = ex.Reason
	}
	if seq, at := host.JournalTail(dir); seq > 0 {
		e.Seq, e.SeqTime = seq, &at
	}
	return e
}

// saveEnded stores the host state of a session that just ended.
func (c *Controller) saveEnded(ctx context.Context, id string) {
	e := c.hostEnded(id)
	if err := c.updateSessionFields(ctx, id, func(f *domain.SessionFields) { f.Ended = &e }); err != nil {
		c.log.WarnContext(ctx, "save ended session state", "session", id, "err", err)
	}
}

// scanSessions is the supervisor pass's walk of the session directories:
// it keeps the ids with a live host and no open session ticket, and
// fills Ended on closed sessions that ended before it was recorded.
func (c *Controller) scanSessions(ctx context.Context) {
	open, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Statuses: []domain.Status{domain.StatusOpen, domain.StatusInProgress}, Limit: math.MaxInt32})
	if err != nil {
		return
	}
	orphans := c.orphans(open)
	c.sessScanMu.Lock()
	c.sessOrphans = orphans
	c.sessScanned = true
	done := c.endedFilled
	c.sessScanMu.Unlock()
	if !done {
		c.backfillEnded(ctx, orphans)
	}
}

// backfillEnded fills Ended on up to endedBackfillBatch closed sessions
// per pass, until none is left.
func (c *Controller) backfillEnded(ctx context.Context, orphans []string) {
	closed, err := c.store.List(ctx, store.Filter{Kind: domain.KindSession, Status: domain.StatusClosed, Limit: math.MaxInt32})
	if err != nil {
		return
	}
	n := 0
	for _, t := range closed {
		if slices.Contains(orphans, t.ID) {
			continue // host still running: filled once it is reaped
		}
		f, err := domain.DecodeFields[domain.SessionFields](t)
		if err != nil || f.Ended != nil {
			continue
		}
		if n == endedBackfillBatch {
			return
		}
		n++
		c.saveEnded(ctx, t.ID)
	}
	c.sessScanMu.Lock()
	c.endedFilled = true
	c.sessScanMu.Unlock()
}

// scannedOrphans returns the last pass's orphaned host ids. Before the
// first pass it walks the directories itself.
func (c *Controller) scannedOrphans(open []domain.Ticket) []string {
	c.sessScanMu.Lock()
	ids, ok := slices.Clone(c.sessOrphans), c.sessScanned
	c.sessScanMu.Unlock()
	if !ok {
		return c.orphans(open)
	}
	return ids
}

// closedSessionInfo is a closed session's list entry from its ticket
// alone: its host is stopped, so the saved Ended state is current.
func closedSessionInfo(t domain.Ticket) rpc.SessionInfo {
	info := rpc.SessionInfo{ID: t.ID}
	if f, err := domain.DecodeFields[domain.SessionFields](t); err == nil {
		info.DisplayName = f.DisplayName
		info.Model, info.Effort, info.Override = f.Model, f.Effort, f.Override
		if e := f.Ended; e != nil {
			info.HostPid, info.Seq, info.SeqTime, info.Exit = e.HostPid, e.Seq, e.SeqTime, e.Exit
		}
	}
	if info.DisplayName == "" {
		info.DisplayName = defaultDisplayName(t, nil)
	}
	return info
}
