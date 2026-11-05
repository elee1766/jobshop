package sqlstore

import (
	"context"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// Curator gate reads (#274). Memory edits and learnings are in
// pkg/memedit and pkg/learning (#579).

// StepsCreated returns step tickets created in [since, until), oldest first.
func (s *Store) StepsCreated(ctx context.Context, since, until time.Time) ([]store.Ticket, error) {
	rows, err := s.q.ListStepsCreated(ctx, sqlcdb.ListStepsCreatedParams{Since: ms(since), Until: ms(until)})
	if err != nil {
		return nil, err
	}
	return toTickets(rows), nil
}
