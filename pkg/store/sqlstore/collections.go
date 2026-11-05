package sqlstore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// Collections (#579): the core validates the spec and hands it to the
// backend. Each backend writes its own collection SQL (with bqb) in its
// package; there is no shared dialect.

// CollectionOpener registers spec on s's database and returns the
// handle. spec has passed Validate.
type CollectionOpener func(ctx context.Context, s *Store, spec store.CollectionSpec) (store.Collection, error)

// Collection registers spec and returns its handle. Registration is
// idempotent.
func (s *Store) Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	if s.drv.Collection == nil {
		return nil, fmt.Errorf("%w: backend has no collections", store.ErrInvalid)
	}
	return s.drv.Collection(ctx, s, spec)
}

// DB is the store's database, for backend collection reads.
func (s *Store) DB() *sql.DB { return s.db }

// Now is the store's clock.
func (s *Store) Now() time.Time { return s.now() }

// WriteTx runs fn in a write transaction, rerunning it when it loses a
// write race, as the ticket writes do.
func (s *Store) WriteTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return s.sqlTx(ctx, fn)
}

// Classify maps a constraint failure to a store error naming what.
func (s *Store) Classify(err error, what string) error { return s.classify(err, what) }
