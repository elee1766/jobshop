package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlq"
)

// claimTries bounds how often Claim moves to the next row after losing
// a race for one, as InboxClaim.
const claimTries = 10

// collections remembers registered specs, so a repeated Collection call
// with the same spec skips the DDL.
type collections struct {
	mu   sync.Mutex
	done map[string]string // name -> spec fingerprint
}

// Collection registers spec (#579): it validates it, creates the
// partition (Postgres), field indexes and search index if missing, and
// returns a handle. Registration is idempotent.
func (s *Store) Collection(ctx context.Context, spec store.CollectionSpec) (store.Collection, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	if s.drv.Dialect == nil {
		return nil, fmt.Errorf("%w: backend has no collection dialect", store.ErrInvalid)
	}
	fp := fmt.Sprintf("%v", spec)
	s.colls.mu.Lock()
	defer s.colls.mu.Unlock()
	if s.colls.done[spec.Name] != fp {
		err := s.sqlTx(ctx, func(tx *sql.Tx) error {
			for _, ddl := range s.drv.Dialect.Register(spec) {
				if _, err := tx.ExecContext(ctx, ddl); err != nil {
					return fmt.Errorf("register collection %s: %w", spec.Name, err)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if s.colls.done == nil {
			s.colls.done = map[string]string{}
		}
		s.colls.done[spec.Name] = fp
	}
	return &collection{s: s, d: s.drv.Dialect, spec: spec}, nil
}

type collection struct {
	s    *Store
	d    sqlq.Dialect
	spec store.CollectionSpec
}

var _ store.Collection = (*collection)(nil)

// querier is what both *sql.DB and *sql.Tx offer.
type querier interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

type scanner interface{ Scan(dest ...any) error }

func (c *collection) scan(r scanner) (store.Record, error) {
	var (
		rec                     store.Record
		value                   string
		lease, created, updated int64
	)
	if err := r.Scan(&rec.Key, &rec.State, &value, &rec.Revision, &rec.Holder, &lease, &rec.Attempts, &created, &updated); err != nil {
		return store.Record{}, err
	}
	rec.Collection = c.spec.Name
	rec.Value = json.RawMessage(value)
	rec.LeaseUntil = fromMS(lease)
	rec.CreatedAt = fromMS(created)
	rec.UpdatedAt = fromMS(updated)
	return rec, nil
}

func (c *collection) one(ctx context.Context, q querier, query sqlq.Query) (store.Record, error) {
	return c.scan(q.QueryRowContext(ctx, query.SQL, query.Args...))
}

func (c *collection) many(ctx context.Context, q querier, query sqlq.Query) ([]store.Record, error) {
	rows, err := q.QueryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []store.Record{}
	for rows.Next() {
		r, err := c.scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (c *collection) notFound(err error, key string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s %s", store.ErrNotFound, c.spec.Name, key)
	}
	return err
}

// value checks v is a JSON object; empty means {}.
func (c *collection) value(v json.RawMessage) (json.RawMessage, error) {
	if len(v) == 0 {
		return json.RawMessage("{}"), nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(v, &m); err != nil || m == nil {
		return nil, fmt.Errorf("%w: %s: value must be a JSON object", store.ErrInvalid, c.spec.Name)
	}
	return v, nil
}

// reindex refreshes key's search document (SQLite FTS); Postgres's
// expression index needs nothing.
func (c *collection) reindex(ctx context.Context, tx *sql.Tx, key string) error {
	if len(c.spec.Search) == 0 {
		return nil
	}
	for _, q := range c.d.SearchIndex(c.spec) {
		if _, err := tx.ExecContext(ctx, q, key); err != nil {
			return err
		}
	}
	return nil
}

func (c *collection) Put(ctx context.Context, r store.Record) (rec store.Record, created bool, err error) {
	if r.Key == "" {
		return rec, false, fmt.Errorf("%w: %s: record needs a key", store.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.value(r.Value); err != nil {
		return rec, false, err
	}
	err = c.s.sqlTx(ctx, func(tx *sql.Tx) error {
		ins := sqlq.Insert(c.d, c.spec, r, c.s.now())
		res, err := tx.ExecContext(ctx, ins.SQL, ins.Args...)
		if err != nil {
			return c.s.classify(err, c.spec.Name+" "+r.Key)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return err
		}
		if created = n == 1; created {
			if err := c.reindex(ctx, tx, r.Key); err != nil {
				return err
			}
		}
		rec, err = c.one(ctx, tx, sqlq.Get(c.d, c.spec, r.Key))
		if created || !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		// Not inserted and no such key: a unique field holds it.
		return c.existingUnique(ctx, tx, r, &rec)
	})
	return rec, created, err
}

// existingUnique finds the record whose unique field value r repeats.
func (c *collection) existingUnique(ctx context.Context, tx *sql.Tx, r store.Record, out *store.Record) error {
	var m map[string]any
	if err := json.Unmarshal(r.Value, &m); err != nil {
		return err
	}
	for _, f := range c.spec.Unique {
		v, ok := lookupPath(m, f.Path)
		if !ok || v == "" {
			continue
		}
		rec, err := c.one(ctx, tx, sqlq.ByField(c.d, c.spec, f, v))
		if err == nil {
			*out = rec
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	return fmt.Errorf("%w: %s %s: insert ignored but no conflicting record found", store.ErrConflict, c.spec.Name, r.Key)
}

// lookupPath reads a dotted path as text, as the dialects' Field does.
func lookupPath(m map[string]any, path string) (string, bool) {
	var cur any = m
	for p := range strings.SplitSeq(path, ".") {
		obj, ok := cur.(map[string]any)
		if !ok {
			return "", false
		}
		if cur, ok = obj[p]; !ok {
			return "", false
		}
	}
	switch v := cur.(type) {
	case nil:
		return "", false
	case string:
		return v, true
	default:
		b, err := json.Marshal(v)
		return string(b), err == nil
	}
}

func (c *collection) Get(ctx context.Context, key string) (store.Record, error) {
	rec, err := c.one(ctx, c.s.db, sqlq.Get(c.d, c.spec, key))
	return rec, c.notFound(err, key)
}

func (c *collection) Update(ctx context.Context, r store.Record) (rec store.Record, err error) {
	if r.Key == "" {
		return rec, fmt.Errorf("%w: %s: record needs a key", store.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.value(r.Value); err != nil {
		return rec, err
	}
	err = c.s.sqlTx(ctx, func(tx *sql.Tx) error {
		now := c.s.now()
		rec, err = c.one(ctx, tx, sqlq.Update(c.d, c.spec, r, now))
		if errors.Is(err, sql.ErrNoRows) {
			return c.whyNot(ctx, tx, r, now)
		}
		if err != nil {
			return c.s.classify(err, c.spec.Name+" "+r.Key)
		}
		return c.reindex(ctx, tx, r.Key)
	})
	return rec, err
}

// whyNot explains an Update that matched no row.
func (c *collection) whyNot(ctx context.Context, tx *sql.Tx, r store.Record, now time.Time) error {
	cur, err := c.one(ctx, tx, sqlq.Get(c.d, c.spec, r.Key))
	if err != nil {
		return c.notFound(err, r.Key)
	}
	live := cur.Holder != "" && cur.LeaseUntil.After(now)
	switch {
	case r.Holder != "" && (!live || cur.Holder != r.Holder):
		return fmt.Errorf("%w: %s %s is not leased to %s", store.ErrNotHolder, c.spec.Name, r.Key, r.Holder)
	case r.Holder == "" && live:
		return fmt.Errorf("%w: %s %s is leased to %s", store.ErrNotHolder, c.spec.Name, r.Key, cur.Holder)
	}
	return fmt.Errorf("%w: %s %s at revision %d, not %d", store.ErrConflict, c.spec.Name, r.Key, cur.Revision, r.Revision)
}

func (c *collection) List(ctx context.Context, f store.RecordFilter) ([]store.Record, error) {
	q, err := sqlq.List(c.d, c.spec, f)
	if err != nil {
		return nil, err
	}
	return c.many(ctx, c.s.db, q)
}

func (c *collection) Count(ctx context.Context, f store.RecordFilter) (map[string]int, error) {
	q, err := sqlq.Count(c.d, c.spec, f)
	if err != nil {
		return nil, err
	}
	rows, err := c.s.db.QueryContext(ctx, q.SQL, q.Args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]int{}
	for rows.Next() {
		var (
			st string
			n  int
		)
		if err := rows.Scan(&st, &n); err != nil {
			return nil, err
		}
		out[st] = n
	}
	return out, rows.Err()
}

func (c *collection) Claim(ctx context.Context, f store.RecordFilter, holder string, lease time.Duration) (rec store.Record, err error) {
	if f.Empty() || holder == "" || lease <= 0 {
		return rec, fmt.Errorf("%w: %s: claim needs a filter, a holder and a lease", store.ErrInvalid, c.spec.Name)
	}
	err = c.s.sqlTx(ctx, func(tx *sql.Tx) error {
		for range claimTries {
			now := c.s.now()
			next, err := sqlq.NextClaimable(c.d, c.spec, f, now)
			if err != nil {
				return err
			}
			var key string
			if err := tx.QueryRowContext(ctx, next.SQL, next.Args...).Scan(&key); err != nil {
				return c.notFound(err, "claimable record")
			}
			rec, err = c.one(ctx, tx, sqlq.Take(c.d, c.spec, key, holder, now, now.Add(lease)))
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		return fmt.Errorf("%w: %s: claim lost every race", store.ErrConflict, c.spec.Name)
	})
	return rec, err
}

func (c *collection) Renew(ctx context.Context, key, holder string, lease time.Duration) (rec store.Record, err error) {
	if holder == "" || lease <= 0 {
		return rec, fmt.Errorf("%w: %s: renew needs a holder and a lease", store.ErrInvalid, c.spec.Name)
	}
	err = c.s.sqlTx(ctx, func(tx *sql.Tx) error {
		now := c.s.now()
		rec, err = c.one(ctx, tx, sqlq.Renew(c.d, c.spec, key, holder, now, now.Add(lease)))
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if _, err := c.one(ctx, tx, sqlq.Get(c.d, c.spec, key)); err != nil {
			return c.notFound(err, key)
		}
		return fmt.Errorf("%w: %s %s is not leased to %s", store.ErrNotHolder, c.spec.Name, key, holder)
	})
	return rec, err
}

func (c *collection) Delete(ctx context.Context, f store.RecordFilter) (n int, err error) {
	if f.Empty() {
		return 0, fmt.Errorf("%w: %s: delete needs a filter", store.ErrInvalid, c.spec.Name)
	}
	del, search, err := sqlq.Delete(c.d, c.spec, f)
	if err != nil {
		return 0, err
	}
	err = c.s.sqlTx(ctx, func(tx *sql.Tx) error {
		if search.SQL != "" {
			if _, err := tx.ExecContext(ctx, search.SQL, search.Args...); err != nil {
				return err
			}
		}
		res, err := tx.ExecContext(ctx, del.SQL, del.Args...)
		if err != nil {
			return err
		}
		m, err := res.RowsAffected()
		n = int(m)
		return err
	})
	return n, err
}

func (c *collection) Search(ctx context.Context, query string, limit int) ([]store.Record, error) {
	if len(c.spec.Search) == 0 {
		return nil, fmt.Errorf("%w: collection %s has no search fields", store.ErrInvalid, c.spec.Name)
	}
	words := strings.Fields(query)
	if len(words) == 0 {
		return nil, fmt.Errorf("%w: empty search query", store.ErrInvalid)
	}
	if limit <= 0 {
		limit = defaultRecallLimit
	}
	match := c.s.drv.SearchMatch
	if match == nil {
		match = c.s.drv.RecallMatch
	}
	sqlText := c.d.Search(c.spec)
	out := []store.Record{}
	for _, m := range match(words) {
		var err error
		if out, err = c.many(ctx, c.s.db, sqlq.Query{SQL: sqlText, Args: []any{m, limit}}); err != nil {
			return nil, err
		}
		if len(out) > 0 {
			break
		}
	}
	return out, nil
}
