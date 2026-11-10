package postgres

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/pressly/goose/v3"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
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

// MigrateDownTo rolls db back to version, for migration tests.
func MigrateDownTo(ctx context.Context, db *sql.DB, version int64) error {
	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, sub)
	if err != nil {
		return err
	}
	_, err = p.DownTo(ctx, version)
	return err
}

// ClaimSQL returns the candidate statements Claim runs for f, one per
// state, for plan tests.
func ClaimSQL(ctx context.Context, s *Store, spec store.CollectionSpec, f store.RecordFilter) ([]string, [][]any, error) {
	col, err := s.Collection(ctx, spec)
	if err != nil {
		return nil, nil, err
	}
	var texts []string
	var args [][]any
	for _, g := range claimFilters(f) {
		q, a, err := col.(*collection).claimQuery(g, s.Now())
		if err != nil {
			return nil, nil, err
		}
		texts, args = append(texts, q), append(args, a)
	}
	return texts, args, nil
}
