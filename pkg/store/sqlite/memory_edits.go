package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlite/sqlcdb"
)

// AddMemoryEdit records a merged curator PR (#274). A PR already
// recorded returns its row unchanged, so a repeated record is harmless.
func (s *Store) AddMemoryEdit(ctx context.Context, e store.MemoryEdit) (out store.MemoryEdit, err error) {
	e.PR = strings.TrimSpace(e.PR)
	if e.PR == "" {
		return out, fmt.Errorf("%w: memory edit needs a pr", store.ErrInvalid)
	}
	if e.MergedAt.IsZero() {
		return out, fmt.Errorf("%w: memory edit needs merged_at", store.ErrInvalid)
	}
	ls, err := json.Marshal(nonNil(e.Learnings))
	if err != nil {
		return out, err
	}
	ps, err := json.Marshal(nonNil(e.Pools))
	if err != nil {
		return out, err
	}
	base, err := json.Marshal(e.Baseline)
	if err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		if r, err := q.GetMemoryEditByPR(ctx, e.PR); err == nil {
			out, err = toMemoryEdit(r)
			return err
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		r, err := q.InsertMemoryEdit(ctx, sqlcdb.InsertMemoryEditParams{
			Pr: e.PR, Learnings: string(ls), Pools: string(ps), MergedAt: ms(e.MergedAt),
			Baseline: string(base), CreatedAt: ms(s.now()),
		})
		if err != nil {
			return err
		}
		out, err = toMemoryEdit(r)
		return err
	})
	return out, err
}

// MemoryEdits lists memory edits newest merge first; empty state is all.
func (s *Store) MemoryEdits(ctx context.Context, state string, n int) ([]store.MemoryEdit, error) {
	rows, err := s.q.ListMemoryEdits(ctx, sqlcdb.ListMemoryEditsParams{State: narg(state), MaxRows: limit(n)})
	if err != nil {
		return nil, err
	}
	out := make([]store.MemoryEdit, 0, len(rows))
	for _, r := range rows {
		e, err := toMemoryEdit(r)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, nil
}

// SetMemoryEditResult stores the gate's decision on an edit and writes
// "<state>: <reason> (<pr>)" on each cited learning, in one transaction.
func (s *Store) SetMemoryEditResult(ctx context.Context, id int64, state string, after store.PoolMetrics, reason string) (out store.MemoryEdit, err error) {
	if state != store.EditHelped && state != store.EditNotHelped {
		return out, fmt.Errorf("%w: memory edit state %q", store.ErrInvalid, state)
	}
	a, err := json.Marshal(after)
	if err != nil {
		return out, err
	}
	err = s.tx(ctx, func(q *sqlcdb.Queries) error {
		r, err := q.SetMemoryEditResult(ctx, sqlcdb.SetMemoryEditResultParams{
			State: state, After: string(a), Reason: reason, EvaluatedAt: ms(s.now()), ID: id,
		})
		if err != nil {
			return notFound(err, fmt.Sprintf("memory edit %d", id))
		}
		if out, err = toMemoryEdit(r); err != nil {
			return err
		}
		gate := fmt.Sprintf("%s: %s (%s)", state, reason, out.PR)
		for _, l := range out.Learnings {
			if err := q.SetLearningGate(ctx, sqlcdb.SetLearningGateParams{Gate: gate, ID: l}); err != nil {
				return err
			}
		}
		return nil
	})
	return out, err
}

// StepsCreated returns step tickets created in [since, until), oldest first.
func (s *Store) StepsCreated(ctx context.Context, since, until time.Time) ([]store.Ticket, error) {
	rows, err := s.q.ListStepsCreated(ctx, sqlcdb.ListStepsCreatedParams{Since: ms(since), Until: ms(until)})
	if err != nil {
		return nil, err
	}
	return toTickets(rows), nil
}

// LearningsBetween returns learnings recorded in [since, until), oldest first.
func (s *Store) LearningsBetween(ctx context.Context, since, until time.Time) ([]store.Learning, error) {
	rows, err := s.q.ListLearningsBetween(ctx, sqlcdb.ListLearningsBetweenParams{Since: ms(since), Until: ms(until)})
	if err != nil {
		return nil, err
	}
	return toLearnings(rows), nil
}

// LearningsByPR returns the learnings whose pr is pr.
func (s *Store) LearningsByPR(ctx context.Context, pr string) ([]store.Learning, error) {
	rows, err := s.q.ListLearningsByPR(ctx, strings.TrimSpace(pr))
	if err != nil {
		return nil, err
	}
	return toLearnings(rows), nil
}

func toLearnings(rows []sqlcdb.Learning) []store.Learning {
	out := make([]store.Learning, len(rows))
	for i, r := range rows {
		out[i] = toLearning(r)
	}
	return out
}

func toMemoryEdit(r sqlcdb.MemoryEdit) (store.MemoryEdit, error) {
	e := store.MemoryEdit{
		ID: r.ID, PR: r.Pr, MergedAt: fromMS(r.MergedAt), State: r.State, Reason: r.Reason,
		EvaluatedAt: fromMS(r.EvaluatedAt), CreatedAt: fromMS(r.CreatedAt),
	}
	for _, f := range []struct {
		raw string
		v   any
	}{{r.Learnings, &e.Learnings}, {r.Pools, &e.Pools}, {r.Baseline, &e.Baseline}, {r.After, &e.After}} {
		if err := json.Unmarshal([]byte(f.raw), f.v); err != nil {
			return e, fmt.Errorf("memory edit %d: %w", r.ID, err)
		}
	}
	return e, nil
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
