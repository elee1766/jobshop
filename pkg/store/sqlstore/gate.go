package sqlstore

import (
	"context"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// Curator gate reads (#274). Memory edits are in pkg/memedit (#579).

// SetLearningGate writes the gate's result on learning id.
func (s *Store) SetLearningGate(ctx context.Context, id int64, gate string) error {
	return s.q.SetLearningGate(ctx, sqlcdb.SetLearningGateParams{Gate: gate, ID: id})
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
