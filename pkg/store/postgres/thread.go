package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/nullism/bqb"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Thread and Changes are built in Go (#579 step 3), like List.

// Thread reads the parts of id's thread that p asks for.
func (s *Store) Thread(ctx context.Context, id string, p store.Parts) (store.Thread, error) {
	var th store.Thread
	var err error
	if p.Edges {
		th.Edges = []domain.Edge{}
		err = s.each(ctx, bqb.New("SELECT from_id, to_id, kind, created_at FROM edges "+
			"WHERE from_id = ? OR to_id = ? ORDER BY created_at, from_id, to_id, kind", id, id),
			func(r *sql.Rows) error {
				var e domain.Edge
				var at int64
				if err := r.Scan(&e.From, &e.To, &e.Kind, &at); err != nil {
					return err
				}
				e.CreatedAt = fromNS(at)
				th.Edges = append(th.Edges, e)
				return nil
			})
		if err != nil {
			return store.Thread{}, err
		}
	}
	if p.Comments {
		th.Comments = []domain.Comment{}
		err = s.each(ctx, bqb.New("SELECT id, ticket_id, author, body, created_at FROM comments "+
			"WHERE ticket_id = ? AND deleted = 0 ORDER BY id", id),
			func(r *sql.Rows) error {
				var c domain.Comment
				var at int64
				if err := r.Scan(&c.ID, &c.TicketID, &c.Author, &c.Body, &at); err != nil {
					return err
				}
				c.CreatedAt = fromNS(at)
				th.Comments = append(th.Comments, c)
				return nil
			})
		if err != nil {
			return store.Thread{}, err
		}
	}
	if p.History {
		if th.History, err = s.historyRows(ctx, bqb.New("WHERE ticket_id = ? ORDER BY seq", id)); err != nil {
			return store.Thread{}, err
		}
	}
	return th, nil
}

// Changes returns history rows of every ticket after seq, oldest first,
// and the newest seq.
func (s *Store) Changes(ctx context.Context, after int64, n int) (store.ChangePage, error) {
	var p store.ChangePage
	err := s.each(ctx, bqb.New("SELECT CAST(COALESCE(MAX(seq), 0) AS BIGINT) FROM history"),
		func(r *sql.Rows) error { return r.Scan(&p.Last) })
	if err != nil || n < 0 {
		return p, err
	}
	if n == 0 {
		n = store.DefaultLimit
	}
	// Rows past Last are left for the next page, so Last never trails a
	// row the caller has seen.
	p.Changes, err = s.historyRows(ctx, bqb.New("WHERE seq > ? AND seq <= ? ORDER BY seq LIMIT ?", after, p.Last, n))
	return p, err
}

func (s *Store) historyRows(ctx context.Context, where *bqb.Query) ([]store.HistoryEntry, error) {
	out := []store.HistoryEntry{}
	err := s.each(ctx, bqb.New("SELECT seq, ticket_id, at, actor, op, revision, diff FROM history ?", where),
		func(r *sql.Rows) error {
			var h store.HistoryEntry
			var at int64
			var diff string
			if err := r.Scan(&h.Seq, &h.TicketID, &at, &h.Actor, &h.Op, &h.Revision, &diff); err != nil {
				return err
			}
			h.At, h.Diff = fromNS(at), json.RawMessage(diff)
			out = append(out, h)
			return nil
		})
	return out, err
}

// each runs q on the database and calls fn for every row.
func (s *Store) each(ctx context.Context, q *bqb.Query, fn func(*sql.Rows) error) error {
	text, args, err := q.ToPgsql()
	if err != nil {
		return err
	}
	rows, err := s.rdb.QueryContext(ctx, text, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
