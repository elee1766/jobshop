package sqlstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// defaultRecallLimit bounds Recall when no limit is given.
const defaultRecallLimit = 10

// Recall searches ticket comments. It runs the driver's match
// expressions in order and returns the first that has hits (SQLite: all
// words, then any word).
func (s *Store) Recall(ctx context.Context, rq store.RecallQuery) ([]store.RecallHit, error) {
	words := strings.Fields(rq.Text)
	if len(words) == 0 {
		return nil, fmt.Errorf("%w: empty recall query", store.ErrInvalid)
	}
	lim := rq.Limit
	if lim <= 0 {
		lim = defaultRecallLimit
	}
	var rows []sqlcdb.RecallSearchRow
	for _, match := range s.drv.RecallMatch(words) {
		var err error
		rows, err = s.q.RecallSearch(ctx, sqlcdb.RecallSearchParams{
			Query: match, Project: rq.Project, TicketID: rq.TicketID, Lim: int64(lim),
		})
		if err != nil {
			return nil, err
		}
		if len(rows) > 0 {
			break
		}
	}
	out := make([]store.RecallHit, len(rows))
	for i, r := range rows {
		out[i] = store.RecallHit{
			Kind: r.Kind, Ref: r.Ref, TicketID: r.TicketID, Project: r.Project,
			At: fromMS(r.AtMs), Snippet: r.Snippet, Score: r.Score,
		}
	}
	return out, nil
}
