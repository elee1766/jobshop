package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/journal"
	"github.com/elee1766/jobshop/pkg/rpc"
	"github.com/elee1766/jobshop/pkg/store"
)

const (
	journalSearchDefaultLimit = 1000
	journalSearchMaxLimit     = 10000
	// grantJournalSearch lets a pool search its own project's journals.
	grantJournalSearch = "journal-search"
)

// errNoVictoriaLogs is journal.search and backfill without config.
var errNoVictoriaLogs = fmt.Errorf("%w: journal search needs journal.victorialogs", store.ErrInvalid)

// journalSearch runs a LogsQL query on VictoriaLogs (#530). Admin sees
// every project; a pool with the journal-search grant only its own,
// enforced by VictoriaLogs' extra_stream_filters, outside the query text.
func (c *Controller) journalSearch(ctx context.Context, sc scope, p rpc.JournalSearchParams) (any, error) {
	if err := c.requireGrant(sc, grantJournalSearch, "journal search"); err != nil {
		return nil, err
	}
	cl := c.vlClient(c.config())
	if cl == nil {
		return nil, errNoVictoriaLogs
	}
	if strings.TrimSpace(p.Query) == "" {
		return nil, fmt.Errorf("%w: query is required (LogsQL, e.g. '*')", store.ErrInvalid)
	}
	limit := p.Limit
	if limit <= 0 {
		limit = journalSearchDefaultLimit
	}
	limit = min(limit, journalSearchMaxLimit)
	q := journal.Query{LogsQL: p.Query, Start: p.Start, End: p.Until, Limit: limit + 1}
	if q.Start.IsZero() && p.Since > 0 {
		q.Start = c.now().Add(-p.Since)
	}
	if !sc.admin {
		q.Project = sc.project
	}
	res := rpc.JournalSearchResult{Rows: []json.RawMessage{}}
	err := cl.Run(ctx, q, func(row json.RawMessage) error {
		if len(res.Rows) == limit {
			res.Truncated = true
			return errStopSearch
		}
		res.Rows = append(res.Rows, row)
		return nil
	})
	if errors.Is(err, errStopSearch) {
		err = nil
	}
	var se *journal.StatusError
	if errors.As(err, &se) && se.Code == 400 {
		return nil, fmt.Errorf("%w: %s", store.ErrInvalid, se.Body)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

var errStopSearch = errors.New("limit reached")

// journalBackfill ships every tracked journal from its cursor now (#530),
// adding sessions closed within Since (all closed sessions when zero), and
// flushes VictoriaLogs so the rows are searchable when it returns.
func (c *Controller) journalBackfill(ctx context.Context, sc scope, p rpc.JournalBackfillParams) (any, error) {
	if !sc.admin {
		return nil, forbidden("journal backfill needs the admin socket")
	}
	cl := c.vlClient(c.config())
	if cl == nil {
		return nil, errNoVictoriaLogs
	}
	closed, err := c.store.List(ctx, store.Filter{Kind: store.KindSession, Status: store.StatusClosed, Order: store.OrderNewest, Limit: 1 << 20})
	if err != nil {
		return nil, err
	}
	c.ship.mu.Lock()
	defer c.ship.mu.Unlock()
	if c.ship.sessions == nil {
		c.ship.sessions = map[string]*shipSession{}
	}
	if err := c.trackSessions(ctx, !c.ship.started); err != nil {
		return nil, err
	}
	c.ship.started = true
	for _, t := range closed {
		if p.Since > 0 && c.since(t.ClosedAt) > p.Since {
			continue
		}
		c.trackSession(t)
	}
	before := map[*shipSession]int64{}
	for _, s := range c.ship.sessions {
		before[s] = s.shipped
	}
	res := rpc.JournalBackfillResult{}
	for {
		n, err := c.shipPass(ctx, cl, true)
		res.Frames += n
		if err != nil {
			return nil, err
		}
		// A session reading in chunks has more after a pass that shipped.
		more := false
		for _, s := range c.ship.sessions {
			if !s.caughtUp() {
				more = true
			}
		}
		if !more || n == 0 {
			break
		}
	}
	c.saveCursors(ctx, true)
	for _, s := range c.ship.sessions {
		if s.closed && s.caughtUp() && s.saved == s.shipped {
			delete(c.ship.sessions, s.t.ID)
		}
	}
	for s, seq := range before {
		if s.shipped > seq {
			res.Sessions++
		}
	}
	if err := cl.Flush(ctx); err != nil {
		return nil, err
	}
	return res, nil
}
