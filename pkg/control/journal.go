package control

import (
	"context"
	"fmt"

	"github.com/elee1766/jobshop/pkg/control/journalread"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

// session.journal (#529): read-only, paged, redacted access to a
// session's ACP journal. It reads the host's journal.ndjson through
// acphost's reader, so it works for live and ended sessions while the
// session dir exists. It takes the session id only, so a later
// multi-controller setup can forward it to the owner unchanged.
//
// The reading lives in pkg/control/journalread (JOB-140). This file is
// the access boundary: it finds the ticket, decides with canReadJournal
// and hands the reader an authorized journalread.Access.

// grantSessionJournal lets a pool read journals in its project.
const grantSessionJournal = "session-journal"

func (c *Controller) sessionJournal(ctx context.Context, sc scope, p rpc.SessionJournalParams) (any, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", store.ErrInvalid)
	}
	if err := journalread.ValidatePage(p); err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != store.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", store.ErrInvalid, p.ID, t.Kind)
	}
	if err := c.canReadJournal(sc, t); err != nil {
		return nil, err
	}
	dir := c.root.SessionDir(t.ID)
	res, err := journalread.Page(journalread.Access{Dir: dir, Redact: c.redactor(journalread.Env(dir))}, p)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// canReadJournal: admin; a session for itself; a pool with the
// session-journal grant for sessions in its project.
func (c *Controller) canReadJournal(sc scope, t store.Ticket) error {
	if sc.admin || sc.session == t.ID {
		return nil
	}
	if p, ok := c.config().Pools[sc.pool]; ok && p.Allows(grantSessionJournal) {
		if t.Project != sc.project {
			return forbidden("session %s is not in project %s", t.ID, sc.project)
		}
		return nil
	}
	return forbidden("reading another session's journal needs the admin socket or pool admin grant %q", grantSessionJournal)
}
