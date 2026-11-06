package sqlite

import (
	"context"
	"fmt"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
)

// defaultRecallLimit bounds Recall when no limit is given.
const defaultRecallLimit = 10

// Recall searches ticket comments with FTS5: all words, then any word
// (recallMatch), returning the first that has hits.
func (s *Store) Recall(ctx context.Context, rq store.RecallQuery) ([]store.RecallHit, error) {
	words := strings.Fields(rq.Text)
	if len(words) == 0 {
		return nil, fmt.Errorf("%w: empty recall query", store.ErrInvalid)
	}
	lim := rq.Limit
	if lim <= 0 {
		lim = defaultRecallLimit
	}
	for _, match := range recallMatch(words) {
		out, err := s.recall(ctx, match, rq.Project, rq.TicketID, lim)
		if err != nil || len(out) > 0 {
			return out, err
		}
	}
	return []store.RecallHit{}, nil
}

func (s *Store) recall(ctx context.Context, match, project, ticket string, lim int) ([]store.RecallHit, error) {
	rows, err := s.conn(s.rdb).QueryContext(ctx, recallSQL, match, project, ticket, int64(lim))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []store.RecallHit{}
	for rows.Next() {
		var h store.RecallHit
		var at int64
		if err := rows.Scan(&h.Kind, &h.Ref, &h.TicketID, &h.Project, &at, &h.Snippet, &h.Score); err != nil {
			return nil, err
		}
		h.At = fromNS(at)
		out = append(out, h)
	}
	return out, rows.Err()
}
