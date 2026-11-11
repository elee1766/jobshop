package sqlite

import (
	"context"
	"database/sql"

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
	return q.ToSql()
}

// CountSQL returns the Count statement of spec on s, for plan tests.
func CountSQL(ctx context.Context, s *Store, spec store.CollectionSpec, f store.RecordFilter) (string, []any, error) {
	col, err := s.Collection(ctx, spec)
	if err != nil {
		return "", nil, err
	}
	return col.(*collection).countQuery(f)
}

// RegisterDDL is registerDDL, for tests.
var RegisterDDL = registerDDL

// QuoteLit is quoteLit, for tests.
var QuoteLit = quoteLit

// MigrateDownTo rolls db back to version with the migrations' Down
// sections, for migration tests.
func MigrateDownTo(ctx context.Context, db *sql.DB, version int64) error {
	p, err := goose.NewProvider(goose.DialectSQLite3, db, mustSub(migrations, "migrations"))
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

// TicketListSQL returns the List statement for f, for plan tests.
func TicketListSQL(s *Store, f store.Filter) (string, []any, error) {
	q, err := listQuery(f, ns(s.now()))
	if err != nil {
		return "", nil, err
	}
	return s.render(q)
}
