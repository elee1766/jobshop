package sqlite

import (
	"context"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqlcdb"
)

// AddLearning inserts a learning (#264). It is not ticket history: learnings
// have their own table so no pool can claim them.
func (s *Store) AddLearning(ctx context.Context, l store.Learning) (out store.Learning, err error) {
	if err := store.ValidLearning(&l); err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		r, err := q.InsertLearning(ctx, sqlcdb.InsertLearningParams{
			Kind: l.Kind, Summary: l.Summary, Fix: l.Fix, Scope: l.Scope, Ticket: l.Ticket,
			Session: l.Session, Source: l.Source, Status: l.Status, CreatedAt: ms(s.now()),
		})
		if err != nil {
			return err
		}
		out = toLearning(r)
		return nil
	})
	return out, err
}

// Learnings lists learnings newest first.
func (s *Store) Learnings(ctx context.Context, f store.LearningFilter) ([]store.Learning, error) {
	rows, err := s.q.ListLearnings(ctx, sqlcdb.ListLearningsParams{
		Since: ms(f.Since), Scope: narg(f.Scope), Kind: narg(f.Kind), Status: narg(f.Status),
		Ticket: narg(f.Ticket), MaxRows: limit(f.Limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]store.Learning, len(rows))
	for i, r := range rows {
		out[i] = toLearning(r)
	}
	return out, nil
}

func toLearning(r sqlcdb.Learning) store.Learning {
	return store.Learning{
		ID: r.ID, Kind: r.Kind, Summary: r.Summary, Fix: r.Fix, Scope: r.Scope, Ticket: r.Ticket,
		Session: r.Session, Source: r.Source, Status: r.Status, CreatedAt: fromMS(r.CreatedAt),
	}
}
