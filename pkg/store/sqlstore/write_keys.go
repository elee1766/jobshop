package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

// ClaimWriteKey returns the stored write for key, inserting it pending
// when new (#280). A key reused for another op is ErrInvalid.
func (s *Store) ClaimWriteKey(ctx context.Context, key, op, actor string) (out store.WriteKey, err error) {
	key, op = strings.TrimSpace(key), strings.TrimSpace(op)
	if key == "" || op == "" {
		return out, fmt.Errorf("%w: write key needs a key and an op", store.ErrInvalid)
	}
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		now := ms(s.now())
		n, err := q.InsertWriteKey(ctx, sqlcdb.InsertWriteKeyParams{Key: key, Op: op, Actor: actor, Now: now})
		if err != nil {
			return err
		}
		r, err := q.GetWriteKey(ctx, key)
		if err != nil {
			return err
		}
		if r.Op != op {
			return fmt.Errorf("%w: write key %q was used for %s, not %s", store.ErrInvalid, key, r.Op, op)
		}
		out = toWriteKey(r)
		out.New = n > 0
		return nil
	})
	return out, err
}

// FinishWriteKey stores the result of a pending write. A key already done
// keeps its first result, which is returned.
func (s *Store) FinishWriteKey(ctx context.Context, key, result string) (out store.WriteKey, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		r, err := q.FinishWriteKey(ctx, sqlcdb.FinishWriteKeyParams{Result: result, Now: ms(s.now()), Key: key})
		if errors.Is(err, sql.ErrNoRows) {
			r, err = q.GetWriteKey(ctx, key)
			if err != nil {
				return notFound(err, "write key "+key)
			}
		} else if err != nil {
			return err
		}
		out = toWriteKey(r)
		return nil
	})
	return out, err
}

// PruneWriteKeys deletes keys created before before, done or pending
// (#415): the local table is a cache, the remote key or marker is the
// record.
func (s *Store) PruneWriteKeys(ctx context.Context, before time.Time) (n int64, err error) {
	err = s.tx(ctx, func(q sqlcdb.Querier) error {
		n, err = q.PruneWriteKeys(ctx, ms(before))
		return err
	})
	return n, err
}

func toWriteKey(r sqlcdb.WriteKey) store.WriteKey {
	return store.WriteKey{
		Key: r.Key, Op: r.Op, Result: r.Result, Done: r.Done != 0, Actor: r.Actor,
		CreatedAt: fromMS(r.CreatedAt), UpdatedAt: fromMS(r.UpdatedAt),
	}
}
