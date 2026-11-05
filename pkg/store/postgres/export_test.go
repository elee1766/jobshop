package postgres

import (
	"context"

	"github.com/elee1766/jobshop/pkg/store"
)

// ListSQL returns the List statement of spec on s, for plan tests.
func ListSQL(ctx context.Context, s *Store, spec store.CollectionSpec, f store.RecordFilter) (string, []any, error) {
	col, err := s.Collection(ctx, spec)
	if err != nil {
		return "", nil, err
	}
	q, err := col.(*collection).listQuery(f)
	if err != nil {
		return "", nil, err
	}
	return q.ToPgsql()
}

// QuoteLit is quoteLit, for tests.
var QuoteLit = quoteLit
