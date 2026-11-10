package control

import (
	"context"
	"fmt"
	"os"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/execution/session"
	"github.com/gfx-labs/jobshop/pkg/storage/journal"
	"github.com/gfx-labs/jobshop/pkg/transport/rpc"
)

// session.journal (#529): read-only, paged, redacted access to a
// session's ACP journal. It reads the host's journal.ndjson through
// acphost's reader, so it works for live and ended sessions while the
// session dir exists. It takes the session id only, so a later
// multi-controller setup can forward it to the owner unchanged.
//
// The reading lives in pkg/storage/journal/read (JOB-140). This file is
// the access boundary: it finds the ticket, decides with canReadJournal
// and hands the reader an authorized journal.Access.

// grantSessionJournal lets a pool read journals in its project.
const grantSessionJournal = "session-journal"

func (c *Controller) sessionJournal(ctx context.Context, sc scope, p rpc.SessionJournalParams) (any, error) {
	if p.ID == "" {
		return nil, fmt.Errorf("%w: session id is required", domain.ErrInvalid)
	}
	if err := journal.ValidatePage(p); err != nil {
		return nil, err
	}
	t, err := c.store.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	if t.Kind != domain.KindSession {
		return nil, fmt.Errorf("%w: %s is a %s, not a session", domain.ErrInvalid, p.ID, t.Kind)
	}
	if err := c.canReadJournal(sc, t); err != nil {
		return nil, err
	}
	dir := c.root.SessionDir(t.ID)
	res, err := journal.Page(journal.Access{Dir: dir, Redact: c.redactor(sessionEnv(dir))}, p)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// sessionEnv is the env the session's agent ran with, for redacting its
// secret values: this controller's inherited env plus the spec's. The
// controller reads the spec so the journal package does not import
// session ownership (JOB-173).
func sessionEnv(dir string) []string {
	env := os.Environ()
	if s, err := session.Read(dir); err == nil {
		env = append(env, s.Env...)
	}
	return env
}

// canReadJournal: admin; a session for itself; a pool with the
// session-journal grant for sessions in its project.
func (c *Controller) canReadJournal(sc scope, t domain.Ticket) error {
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
