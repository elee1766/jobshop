package sqlstore

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// AddLearning inserts a learning (#264). It is not ticket history: learnings
// have their own table so no pool can claim them.
func (s *Store) AddLearning(ctx context.Context, l store.Learning) (out store.Learning, err error) {
	if err := store.ValidLearning(&l); err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
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
	return toLearnings(rows), nil
}

func toLearning(r sqlcdb.Learning) store.Learning {
	return store.Learning{
		ID: r.ID, Kind: r.Kind, Summary: r.Summary, Fix: r.Fix, Scope: r.Scope, Ticket: r.Ticket,
		Session: r.Session, Source: r.Source, Status: r.Status, PR: r.Pr, Gate: r.Gate, CreatedAt: fromMS(r.CreatedAt),
	}
}

// UpdateLearning sets a learning's status and PR (#273).
func (s *Store) UpdateLearning(ctx context.Context, id int64, status, pr string) (out store.Learning, err error) {
	if status != store.LearningOpen && status != store.LearningApplied && status != store.LearningRejected {
		return out, fmt.Errorf("%w: learning status %q", store.ErrInvalid, status)
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := q.UpdateLearning(ctx, sqlcdb.UpdateLearningParams{Status: status, Pr: strings.TrimSpace(pr), ID: id})
		if err != nil {
			return notFound(err, fmt.Sprintf("learning %d", id))
		}
		out = toLearning(r)
		return nil
	})
	return out, err
}

// CountLearnings counts learnings created at or after since with status
// (empty: any).
func (s *Store) CountLearnings(ctx context.Context, since time.Time, status string) (int, error) {
	n, err := s.q.CountLearnings(ctx, sqlcdb.CountLearningsParams{Since: ms(since), Status: narg(status)})
	return int(n), err
}

// GetLearning reads one learning.
func (s *Store) GetLearning(ctx context.Context, id int64) (store.Learning, error) {
	r, err := s.q.GetLearning(ctx, id)
	if err != nil {
		return store.Learning{}, notFound(err, fmt.Sprintf("learning %d", id))
	}
	return toLearning(r), nil
}
