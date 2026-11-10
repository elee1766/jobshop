package postgres

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

// DB is the store's write pool. storemigrate copies into it.
func (s *Store) DB() *sql.DB { return s.db }

// readDB is the read-only pool, for collection reads outside a
// transaction.
func (s *Store) readDB() *sql.DB { return s.rdb }

// Now is the store's clock.
func (s *Store) Now() time.Time { return s.now() }

// writeTx runs fn in a serializable write transaction, rerunning it
// when it loses a race, as the ticket writes do.
func (s *Store) writeTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	return s.sqlTx(ctx, fn)
}
