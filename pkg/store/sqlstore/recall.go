package sqlstore

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// defaultRecallLimit bounds Recall when no limit is given.
const defaultRecallLimit = 10

// IndexRecall replaces the document with d's kind and ref.
func (s *Store) IndexRecall(ctx context.Context, d store.RecallDoc) error {
	if d.Kind == "" || d.Ref == "" {
		return fmt.Errorf("%w: recall doc needs kind and ref", store.ErrInvalid)
	}
	return s.tx(ctx, func(q sqlcdb.Querier) error {
		if err := q.RecallDelete(ctx, sqlcdb.RecallDeleteParams{Kind: d.Kind, Ref: d.Ref}); err != nil {
			return err
		}
		if strings.TrimSpace(d.Body) == "" {
			return nil
		}
		return q.RecallInsert(ctx, sqlcdb.RecallInsertParams{
			Body: d.Body, Kind: d.Kind, Ref: d.Ref, TicketID: d.TicketID, Project: d.Project,
			At: strconv.FormatInt(ms(d.At), 10),
		})
	})
}

// Recall runs the driver's match expressions in order and returns the
// first that has hits (SQLite: all words, then any word).
func (s *Store) Recall(ctx context.Context, rq store.RecallQuery) ([]store.RecallHit, error) {
	words := strings.Fields(rq.Text)
	if len(words) == 0 {
		return nil, fmt.Errorf("%w: empty recall query", store.ErrInvalid)
	}
	lim := rq.Limit
	if lim <= 0 {
		lim = defaultRecallLimit
	}
	search := func(match string) ([]sqlcdb.RecallSearchRow, error) {
		return s.q.RecallSearch(ctx, sqlcdb.RecallSearchParams{
			Query: match, Project: rq.Project, Kind: rq.Kind, TicketID: rq.TicketID, Lim: int64(lim),
		})
	}
	var rows []sqlcdb.RecallSearchRow
	for _, match := range s.drv.RecallMatch(words) {
		var err error
		if rows, err = search(match); err != nil {
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
