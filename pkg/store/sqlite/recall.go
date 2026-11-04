package sqlite

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqlcdb"
)

// defaultRecallLimit bounds Recall when no limit is given.
const defaultRecallLimit = 10

// IndexRecall replaces the document with d's kind and ref.
func (s *Store) IndexRecall(ctx context.Context, d store.RecallDoc) error {
	if d.Kind == "" || d.Ref == "" {
		return fmt.Errorf("%w: recall doc needs kind and ref", store.ErrInvalid)
	}
	return s.tx(ctx, func(q *sqlcdb.Queries) error {
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

// Recall runs an FTS5 query. Words must all match; when none match that
// way, any word may match, so a long question still finds something.
func (s *Store) Recall(ctx context.Context, rq store.RecallQuery) ([]store.RecallHit, error) {
	words := ftsWords(rq.Text)
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
	rows, err := search(strings.Join(words, " "))
	if err == nil && len(rows) == 0 && len(words) > 1 {
		rows, err = search(strings.Join(words, " OR "))
	}
	if err != nil {
		return nil, err
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

// ftsWords quotes each word so FTS5 syntax in a query (quotes, colons,
// hyphens, AND/NOT) is matched as text instead of failing to parse.
func ftsWords(text string) []string {
	var out []string
	for _, w := range strings.Fields(text) {
		w = strings.ReplaceAll(w, `"`, `""`)
		out = append(out, `"`+w+`"`)
	}
	return out
}
