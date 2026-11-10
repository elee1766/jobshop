package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nullism/bqb"

	"github.com/gfx-labs/jobshop/pkg/domain"
	"github.com/gfx-labs/jobshop/pkg/storage/store"
)

// Collections on SQLite (#579, docs/notes/579-collections-store.md).
// The SQL is built here with bqb; Postgres has its own. Identifiers come
// only from a spec that passed Validate; values are always bound. The
// collection name is written as a literal (quoteLit), not bound, because
// SQLite uses a partial index only when the query's WHERE implies it.

// claimTries bounds how often Claim moves to the next row after losing
// a race for one, as InboxClaim.
const claimTries = 10

// quoteLit quotes a validated name as an SQL string literal.
func quoteLit(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// field is the text expression of a JSON path in value. Index and query
// use the same text, so the index applies.
func field(path string) string {
	return "CAST(json_extract(value, '$." + path + "') AS TEXT)"
}

func columns(p string) string {
	return p + "key, " + p + "state, " + p + "value, " + p + "revision, " + p + "holder, " +
		p + "lease_until, " + p + "attempts, " + p + "created_at, " + p + "updated_at"
}

func ns(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}

func fromNS(v int64) time.Time {
	if v == 0 {
		return time.Time{}
	}
	return time.Unix(0, v).UTC()
}

// registry remembers registered specs, so a repeated Collection call
// with the same spec skips the DDL.
type registry struct {
	mu   sync.Mutex
	done map[string]string // name -> spec fingerprint
}

func (g *registry) open(ctx context.Context, s *Store, spec store.CollectionSpec) (store.Collection, error) {
	fp := fmt.Sprintf("%v", spec)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done[spec.Name] != fp {
		err := s.WriteTx(ctx, func(tx *sql.Tx) error {
			for _, ddl := range registerDDL(spec) {
				if _, err := tx.ExecContext(ctx, ddl); err != nil {
					return fmt.Errorf("register collection %s: %w", spec.Name, err)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if g.done == nil {
			g.done = map[string]string{}
		}
		g.done[spec.Name] = fp
	}
	return &collection{s: s, spec: spec}, nil
}

// registerDDL creates spec's field indexes, unique indexes and FTS5
// table if missing.
func registerDDL(spec store.CollectionSpec) []string {
	var out []string
	name := quoteLit(spec.Name)
	for _, f := range spec.Index {
		out = append(out, "CREATE INDEX IF NOT EXISTS ri_"+spec.Name+"_"+f.Name+" ON records ("+
			field(f.Path)+", created_at, key) WHERE collection = "+name)
	}
	for _, f := range spec.Unique {
		out = append(out, "CREATE UNIQUE INDEX IF NOT EXISTS ru_"+spec.Name+"_"+f.Name+" ON records ("+
			field(f.Path)+") WHERE collection = "+name+" AND "+field(f.Path)+" != ''")
	}
	if len(spec.Search) > 0 {
		out = append(out, "CREATE VIRTUAL TABLE IF NOT EXISTS fts_"+spec.Name+" USING fts5(key UNINDEXED, body)")
	}
	return out
}

type collection struct {
	s    *Store
	spec store.CollectionSpec
}

var _ store.Collection = (*collection)(nil)

type querier interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

type scanner interface{ Scan(dest ...any) error }

func (c *collection) lit() string { return quoteLit(c.spec.Name) }

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
	rec.LeaseUntil, rec.CreatedAt, rec.UpdatedAt = fromNS(lease), fromNS(created), fromNS(updated)
	return rec, nil
}

func (c *collection) one(ctx context.Context, db querier, q *bqb.Query) (store.Record, error) {
	text, args, err := q.ToSql()
	if err != nil {
		return store.Record{}, err
	}
	return c.scan(db.QueryRowContext(ctx, text, args...))
}

func (c *collection) many(ctx context.Context, db querier, q *bqb.Query) ([]store.Record, error) {
	text, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, text, args...)
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

func exec(ctx context.Context, db querier, q *bqb.Query) (int64, error) {
	text, args, err := q.ToSql()
	if err != nil {
		return 0, err
	}
	res, err := db.ExecContext(ctx, text, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (c *collection) notFound(err error, key string) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: %s %s", domain.ErrNotFound, c.spec.Name, key)
	}
	return err
}

// where builds the WHERE conditions of f.
func (c *collection) where(f store.RecordFilter, p string) (*bqb.Query, error) {
	w := bqb.New(p + "collection = " + c.lit())
	if len(f.Keys) > 0 {
		w.And(p+"key IN (?)", f.Keys)
	}
	if len(f.States) > 0 {
		w.And(p+"state IN (?)", f.States)
	}
	for _, n := range slices.Sorted(maps.Keys(f.Where)) {
		fld, unique, ok := c.spec.Lookup(n)
		if !ok {
			return nil, fmt.Errorf("%w: collection %s: field %q is not declared", domain.ErrInvalid, c.spec.Name, n)
		}
		w.And(field(fld.Path)+" = ?", f.Where[n])
		if unique {
			// The unique index is partial on non-empty values; repeat
			// its condition so the planner can use it.
			w.And(field(fld.Path) + " != ''")
		}
	}
	if !f.Since.IsZero() {
		w.And(p+"created_at >= ?", ns(f.Since))
	}
	if !f.Until.IsZero() {
		w.And(p+"created_at < ?", ns(f.Until))
	}
	if !f.UpdatedBefore.IsZero() {
		w.And(p+"updated_at < ?", ns(f.UpdatedBefore))
	}
	return w, nil
}

// listQuery is the List statement; the index-plan test explains it.
func (c *collection) listQuery(f store.RecordFilter) (*bqb.Query, error) {
	w, err := c.where(f, "")
	if err != nil {
		return nil, err
	}
	ord, err := store.RecordOrder(f.Order)
	if err != nil {
		return nil, err
	}
	lim := f.Limit
	if lim <= 0 {
		lim = store.DefaultLimit
	}
	return bqb.New("SELECT "+columns("")+" FROM records WHERE ? ORDER BY "+ord+" LIMIT ?", w, lim), nil
}

func (c *collection) get(key string) *bqb.Query {
	return bqb.New("SELECT "+columns("")+" FROM records WHERE collection = "+c.lit()+" AND key = ?", key)
}

// reindex refreshes key's FTS document.
func (c *collection) reindex(ctx context.Context, tx *sql.Tx, key string) error {
	if len(c.spec.Search) == 0 {
		return nil
	}
	parts := make([]string, len(c.spec.Search))
	for i, f := range c.spec.Search {
		parts[i] = "coalesce(" + field(f.Path) + ", '')"
	}
	t := "fts_" + c.spec.Name
	if _, err := exec(ctx, tx, bqb.New("DELETE FROM "+t+" WHERE key = ?", key)); err != nil {
		return err
	}
	_, err := exec(ctx, tx, bqb.New("INSERT INTO "+t+" (key, body) SELECT key, "+strings.Join(parts, " || char(10) || ")+
		" FROM records WHERE collection = "+c.lit()+" AND key = ?", key))
	return err
}

func (c *collection) Put(ctx context.Context, r store.Record) (rec store.Record, created bool, err error) {
	if r.Key == "" {
		return rec, false, fmt.Errorf("%w: %s: record needs a key", domain.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.spec.CheckValue(r.Value); err != nil {
		return rec, false, err
	}
	err = c.s.WriteTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		at := r.CreatedAt
		if at.IsZero() {
			at = now
		}
		n, err := exec(ctx, tx, bqb.New("INSERT INTO records (collection, key, state, value, revision, holder, lease_until, attempts, created_at, updated_at) VALUES ("+
			c.lit()+", ?, ?, ?, 1, '', 0, 0, ?, ?) ON CONFLICT DO NOTHING", r.Key, r.State, string(r.Value), ns(at), ns(now)))
		if err != nil {
			return c.s.Classify(err, c.spec.Name+" "+r.Key)
		}
		if created = n == 1; created {
			if err := c.reindex(ctx, tx, r.Key); err != nil {
				return err
			}
		}
		rec, err = c.one(ctx, tx, c.get(r.Key))
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
	for _, f := range c.spec.Unique {
		v, ok := store.FieldText(r.Value, f.Path)
		if !ok || v == "" {
			continue
		}
		rec, err := c.one(ctx, tx, bqb.New("SELECT "+columns("")+" FROM records WHERE collection = "+c.lit()+
			" AND "+field(f.Path)+" = ? AND "+field(f.Path)+" != '' LIMIT 1", v))
		if err == nil {
			*out = rec
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	return fmt.Errorf("%w: %s %s: insert ignored but no conflicting record found", domain.ErrConflict, c.spec.Name, r.Key)
}

func (c *collection) Get(ctx context.Context, key string) (store.Record, error) {
	rec, err := c.one(ctx, c.s.DB(), c.get(key))
	return rec, c.notFound(err, key)
}

func (c *collection) Update(ctx context.Context, r store.Record) (rec store.Record, err error) {
	if r.Key == "" {
		return rec, fmt.Errorf("%w: %s: record needs a key", domain.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.spec.CheckValue(r.Value); err != nil {
		return rec, err
	}
	err = c.s.WriteTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		q := bqb.New("UPDATE records SET state = ?, value = ?, revision = revision + 1, updated_at = ?", r.State, string(r.Value), ns(now))
		if r.LeaseUntil.IsZero() {
			q.Space(", holder = '', lease_until = 0")
		}
		q.Space("WHERE collection = "+c.lit()+" AND key = ? AND revision = ?", r.Key, r.Revision)
		if r.Holder == "" {
			q.Space("AND (holder = '' OR lease_until <= ?)", ns(now))
		} else {
			q.Space("AND holder = ? AND lease_until > ?", r.Holder, ns(now))
		}
		q.Space("RETURNING " + columns(""))
		rec, err = c.one(ctx, tx, q)
		if errors.Is(err, sql.ErrNoRows) {
			return c.whyNot(ctx, tx, r, now)
		}
		if err != nil {
			return c.s.Classify(err, c.spec.Name+" "+r.Key)
		}
		return c.reindex(ctx, tx, r.Key)
	})
	return rec, err
}

// whyNot explains an Update that matched no row.
func (c *collection) whyNot(ctx context.Context, tx *sql.Tx, r store.Record, now time.Time) error {
	cur, err := c.one(ctx, tx, c.get(r.Key))
	if err != nil {
		return c.notFound(err, r.Key)
	}
	live := cur.Holder != "" && cur.LeaseUntil.After(now)
	switch {
	case r.Holder != "" && (!live || cur.Holder != r.Holder):
		return fmt.Errorf("%w: %s %s is not leased to %s", domain.ErrNotHolder, c.spec.Name, r.Key, r.Holder)
	case r.Holder == "" && live:
		return fmt.Errorf("%w: %s %s is leased to %s", domain.ErrNotHolder, c.spec.Name, r.Key, cur.Holder)
	}
	return fmt.Errorf("%w: %s %s at revision %d, not %d", domain.ErrConflict, c.spec.Name, r.Key, cur.Revision, r.Revision)
}

func (c *collection) List(ctx context.Context, f store.RecordFilter) ([]store.Record, error) {
	q, err := c.listQuery(f)
	if err != nil {
		return nil, err
	}
	return c.many(ctx, c.s.DB(), q)
}

func (c *collection) Count(ctx context.Context, f store.RecordFilter) (map[string]int, error) {
	text, args, err := c.countQuery(f)
	if err != nil {
		return nil, err
	}
	rows, err := c.s.DB().QueryContext(ctx, text, args...)
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

// countQuery is Count's statement. records_state_created (JOB-61) serves
// a state filter with a created_at range.
func (c *collection) countQuery(f store.RecordFilter) (string, []any, error) {
	w, err := c.where(f, "")
	if err != nil {
		return "", nil, err
	}
	return bqb.New("SELECT state, COUNT(*) FROM records WHERE ? GROUP BY state", w).ToSql()
}

func (c *collection) Claim(ctx context.Context, f store.RecordFilter, holder string, lease time.Duration) (rec store.Record, err error) {
	if f.Empty() || holder == "" || lease <= 0 {
		return rec, fmt.Errorf("%w: %s: claim needs a filter, a holder and a lease", domain.ErrInvalid, c.spec.Name)
	}
	err = c.s.WriteTx(ctx, func(tx *sql.Tx) error {
		for range claimTries {
			now := c.s.Now()
			w, err := c.where(f, "")
			if err != nil {
				return err
			}
			text, args, err := bqb.New("SELECT key FROM records WHERE ? AND (holder = '' OR lease_until <= ?) ORDER BY created_at, key LIMIT 1", w, ns(now)).ToSql()
			if err != nil {
				return err
			}
			var key string
			if err := tx.QueryRowContext(ctx, text, args...).Scan(&key); err != nil {
				return c.notFound(err, "claimable record")
			}
			rec, err = c.one(ctx, tx, bqb.New("UPDATE records SET holder = ?, lease_until = ?, attempts = attempts + 1, revision = revision + 1, updated_at = ?"+
				" WHERE collection = "+c.lit()+" AND key = ? AND (holder = '' OR lease_until <= ?) RETURNING "+columns(""),
				holder, ns(now.Add(lease)), ns(now), key, ns(now)))
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return err
		}
		return fmt.Errorf("%w: %s: claim lost every race", domain.ErrConflict, c.spec.Name)
	})
	return rec, err
}

func (c *collection) Renew(ctx context.Context, key, holder string, lease time.Duration) (rec store.Record, err error) {
	if holder == "" || lease <= 0 {
		return rec, fmt.Errorf("%w: %s: renew needs a holder and a lease", domain.ErrInvalid, c.spec.Name)
	}
	err = c.s.WriteTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		rec, err = c.one(ctx, tx, bqb.New("UPDATE records SET lease_until = ? WHERE collection = "+c.lit()+
			" AND key = ? AND holder = ? AND holder != '' AND lease_until > ? RETURNING "+columns(""),
			ns(now.Add(lease)), key, holder, ns(now)))
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if _, err := c.one(ctx, tx, c.get(key)); err != nil {
			return c.notFound(err, key)
		}
		return fmt.Errorf("%w: %s %s is not leased to %s", domain.ErrNotHolder, c.spec.Name, key, holder)
	})
	return rec, err
}

func (c *collection) Delete(ctx context.Context, f store.RecordFilter) (n int, err error) {
	if f.Empty() {
		return 0, fmt.Errorf("%w: %s: delete needs a filter", domain.ErrInvalid, c.spec.Name)
	}
	w, err := c.where(f, "")
	if err != nil {
		return 0, err
	}
	err = c.s.WriteTx(ctx, func(tx *sql.Tx) error {
		if len(c.spec.Search) > 0 {
			if _, err := exec(ctx, tx, bqb.New("DELETE FROM fts_"+c.spec.Name+" WHERE key IN (SELECT key FROM records WHERE ?)", w)); err != nil {
				return err
			}
		}
		m, err := exec(ctx, tx, bqb.New("DELETE FROM records WHERE ?", w))
		n = int(m)
		return err
	})
	return n, err
}

// defaultSearchLimit is Search's limit when none is given, as recall's.
const defaultSearchLimit = 10

func (c *collection) Search(ctx context.Context, query string, limit int) ([]store.Record, error) {
	if len(c.spec.Search) == 0 {
		return nil, fmt.Errorf("%w: collection %s has no search fields", domain.ErrInvalid, c.spec.Name)
	}
	words := strings.Fields(query)
	if len(words) == 0 {
		return nil, fmt.Errorf("%w: empty search query", domain.ErrInvalid)
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	t := "fts_" + c.spec.Name
	out := []store.Record{}
	for _, m := range recallMatch(words) {
		q := bqb.New("SELECT "+columns("r.")+" FROM "+t+" JOIN records r ON r.collection = "+c.lit()+
			" AND r.key = "+t+".key WHERE "+t+" MATCH ? ORDER BY bm25("+t+"), r.created_at, r.key LIMIT ?", m, limit)
		var err error
		if out, err = c.many(ctx, c.s.DB(), q); err != nil {
			return nil, err
		}
		if len(out) > 0 {
			break
		}
	}
	return out, nil
}
