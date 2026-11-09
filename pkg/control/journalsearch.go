package control

import (
	"context"
	"fmt"

	"github.com/elee1766/jobshop/pkg/control/journalread"
	"github.com/elee1766/jobshop/pkg/control/journalship"
	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/rpc"
)

// grantJournalSearch lets a pool search its own project's journals.
const grantJournalSearch = "journal-search"

// errNoVictoriaLogs is journal.search and backfill without config.
var errNoVictoriaLogs = fmt.Errorf("%w: journal search needs journal.victorialogs", domain.ErrInvalid)

// journalSearch runs a LogsQL query on VictoriaLogs (#530). Admin sees
// every project; a pool with the journal-search grant only its own,
// enforced by VictoriaLogs' extra_stream_filters, outside the query text.
// The grant and the project scope are decided here; journalread runs the
// query (JOB-140).
func (c *Controller) journalSearch(ctx context.Context, sc scope, p rpc.JournalSearchParams) (any, error) {
	if err := c.requireGrant(sc, grantJournalSearch, "journal search"); err != nil {
		return nil, err
	}
	cl := journalship.Client(c.config())
	if cl == nil {
		return nil, errNoVictoriaLogs
	}
	s := journalread.Search{Client: cl, Now: c.now(), JournalSearchParams: p}
	if !sc.admin {
		s.Project = sc.project
	}
	res, err := journalread.Run(ctx, s)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// journalBackfill ships every tracked journal from its cursor now (#530),
// adding sessions closed within Since (all closed sessions when zero), and
// flushes VictoriaLogs so the rows are searchable when it returns. The
// shipper does the work; this is the admin check.
func (c *Controller) journalBackfill(ctx context.Context, sc scope, p rpc.JournalBackfillParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("journal backfill needs the admin socket")
	}
	cl := journalship.Client(c.config())
	if cl == nil {
		return nil, errNoVictoriaLogs
	}
	res, err := c.shipper().Backfill(ctx, cl, p.Since)
	if err != nil {
		return nil, err
	}
	return res, nil
}
