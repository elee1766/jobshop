package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Collection registers spec and returns its handle (#579). Registration
// is idempotent.
func (s *Store) Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	return s.reg.open(ctx, s, spec)
}

// DB is the store's write database. storemigrate copies into it.
func (s *Store) DB() *sql.DB { return s.db }

// Now is the store's clock.
func (s *Store) Now() time.Time { return s.now() }

// WriteTx runs fn in a write transaction, as the ticket writes do.
func (s *Store) WriteTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return s.sqlTx(ctx, fn)
}

// Classify maps a constraint failure to a store error naming what.
func (s *Store) Classify(err error, what string) error { return s.classify(err, what) }

// readDB is the pool for collection reads outside a transaction: the
// query-only readers when the store has them, else the writer. A read
// inside a write transaction uses the transaction.
func (s *Store) readDB() *sql.DB { return s.rdb }
