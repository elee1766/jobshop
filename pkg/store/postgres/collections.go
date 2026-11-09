package postgres

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

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/store"
)

// Collections on Postgres (#579, docs/notes/579-collections-store.md).
// The SQL is built here with bqb and rendered with $n placeholders;
// SQLite has its own. records is partitioned by collection: registration
// creates the partition, and a write to an unregistered collection fails
// (no DEFAULT partition). Identifiers come only from a spec that passed
// Validate; values are always bound, the collection name included.

// claimTries bounds how often Claim moves to the next row after losing
// a race for one, as InboxClaim.
const claimTries = 10

// quoteLit quotes a validated name as an SQL string literal, for DDL,
// which takes no bound parameters.
func quoteLit(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// field is the text expression of a JSON path in value.
func field(path string) string {
	return "(value #>> '{" + strings.ReplaceAll(path, ".", ",") + "}')"
}

// fieldIndex names the parent index of a JSON path. Indexes on the
// partitioned parent apply to every partition, and collections that
// declare the same path share one.
func fieldIndex(path string) string { return "rf_" + strings.ReplaceAll(path, ".", "__") }

const columns = "key, state, CAST(value AS TEXT), revision, holder, lease_until, attempts, created_at, updated_at"

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

// searchText joins the search fields into one text, for the GIN index
// and the query alike.
func searchText(spec store.CollectionSpec) string {
	parts := make([]string, len(spec.Search))
	for i, f := range spec.Search {
		parts[i] = "coalesce(" + field(f.Path) + ", '')"
	}
	return strings.Join(parts, " || ' ' || ")
}

func tsv(spec store.CollectionSpec) string {
	return "to_tsvector('english', " + searchText(spec) + ")"
}

type registry struct {
	mu   sync.Mutex
	done map[string]string // name -> spec fingerprint
}

func (g *registry) open(ctx context.Context, s *Store, spec store.CollectionSpec) (store.Collection, error) {
	fp := fmt.Sprintf("%v", spec)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.done[spec.Name] != fp {
		err := s.writeTx(ctx, func(tx *sql.Tx) error {
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

// registerDDL creates spec's partition, field indexes, unique indexes
// and search index if missing.
func registerDDL(spec store.CollectionSpec) []string {
	part := "records_" + spec.Name
	out := []string{"CREATE TABLE IF NOT EXISTS " + part + " PARTITION OF records FOR VALUES IN (" + quoteLit(spec.Name) + ")"}
	for _, f := range spec.Index {
		out = append(out, "CREATE INDEX IF NOT EXISTS "+fieldIndex(f.Path)+" ON records ("+field(f.Path)+", created_at)")
	}
	// A unique index on a partitioned parent must include the partition
	// key, so each one lives on its collection's partition.
	for _, f := range spec.Unique {
		out = append(out, "CREATE UNIQUE INDEX IF NOT EXISTS ru_"+spec.Name+"_"+f.Name+" ON "+part+" ("+
			field(f.Path)+") WHERE "+field(f.Path)+" != ''")
	}
	if len(spec.Search) > 0 {
		out = append(out, "CREATE INDEX IF NOT EXISTS rs_"+spec.Name+" ON "+part+" USING gin ("+tsv(spec)+")")
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
	text, args, err := q.ToPgsql()
	if err != nil {
		return store.Record{}, err
	}
	return c.scan(db.QueryRowContext(ctx, text, args...))
}

func (c *collection) many(ctx context.Context, db querier, q *bqb.Query) ([]store.Record, error) {
	text, args, err := q.ToPgsql()
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
	text, args, err := q.ToPgsql()
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

// where builds the WHERE conditions of f. Binding the collection lets
// the planner prune to its partition at execution.
func (c *collection) where(f store.RecordFilter) (*bqb.Query, error) {
	w := bqb.New("collection = ?", c.spec.Name)
	if len(f.Keys) > 0 {
		w.And("key IN (?)", f.Keys)
	}
	if len(f.States) > 0 {
		w.And("state IN (?)", f.States)
	}
	for _, n := range slices.Sorted(maps.Keys(f.Where)) {
		fld, unique, ok := c.spec.Lookup(n)
		if !ok {
			return nil, fmt.Errorf("%w: collection %s: field %q is not declared", domain.ErrInvalid, c.spec.Name, n)
		}
		w.And(field(fld.Path)+" = ?", f.Where[n])
		if unique {
			// The unique index is partial on non-empty values.
			w.And(field(fld.Path) + " != ''")
		}
	}
	if !f.Since.IsZero() {
		w.And("created_at >= ?", ns(f.Since))
	}
	if !f.Until.IsZero() {
		w.And("created_at < ?", ns(f.Until))
	}
	if !f.UpdatedBefore.IsZero() {
		w.And("updated_at < ?", ns(f.UpdatedBefore))
	}
	return w, nil
}

// listQuery is the List statement; the index-plan test explains it.
func (c *collection) listQuery(f store.RecordFilter) (*bqb.Query, error) {
	w, err := c.where(f)
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
	return bqb.New("SELECT "+columns+" FROM records WHERE ? ORDER BY "+ord+" LIMIT ?", w, lim), nil
}

func (c *collection) get(key string) *bqb.Query {
	return bqb.New("SELECT "+columns+" FROM records WHERE collection = ? AND key = ?", c.spec.Name, key)
}

func (c *collection) Put(ctx context.Context, r store.Record) (rec store.Record, created bool, err error) {
	if r.Key == "" {
		return rec, false, fmt.Errorf("%w: %s: record needs a key", domain.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.spec.CheckValue(r.Value); err != nil {
		return rec, false, err
	}
	err = c.s.writeTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		at := r.CreatedAt
		if at.IsZero() {
			at = now
		}
		n, err := exec(ctx, tx, bqb.New("INSERT INTO records (collection, key, state, value, revision, holder, lease_until, attempts, created_at, updated_at)"+
			" VALUES (?, ?, ?, CAST(? AS jsonb), 1, '', 0, 0, ?, ?) ON CONFLICT DO NOTHING",
			c.spec.Name, r.Key, r.State, string(r.Value), ns(at), ns(now)))
		if err != nil {
			return c.s.classify(err, c.spec.Name+" "+r.Key)
		}
		created = n == 1
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
		rec, err := c.one(ctx, tx, bqb.New("SELECT "+columns+" FROM records WHERE collection = ? AND "+
			field(f.Path)+" = ? AND "+field(f.Path)+" != '' LIMIT 1", c.spec.Name, v))
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
	rec, err := c.one(ctx, c.s.readDB(), c.get(key))
	return rec, c.notFound(err, key)
}

func (c *collection) Update(ctx context.Context, r store.Record) (rec store.Record, err error) {
	if r.Key == "" {
		return rec, fmt.Errorf("%w: %s: record needs a key", domain.ErrInvalid, c.spec.Name)
	}
	if r.Value, err = c.spec.CheckValue(r.Value); err != nil {
		return rec, err
	}
	err = c.s.writeTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		q := bqb.New("UPDATE records SET state = ?, value = CAST(? AS jsonb), revision = revision + 1, updated_at = ?", r.State, string(r.Value), ns(now))
		if r.LeaseUntil.IsZero() {
			q.Space(", holder = '', lease_until = 0")
		}
		q.Space("WHERE collection = ? AND key = ? AND revision = ?", c.spec.Name, r.Key, r.Revision)
		if r.Holder == "" {
			q.Space("AND (holder = '' OR lease_until <= ?)", ns(now))
		} else {
			q.Space("AND holder = ? AND lease_until > ?", r.Holder, ns(now))
		}
		q.Space("RETURNING " + columns)
		rec, err = c.one(ctx, tx, q)
		if errors.Is(err, sql.ErrNoRows) {
			return c.whyNot(ctx, tx, r, now)
		}
		return c.s.classify(err, c.spec.Name+" "+r.Key)
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
	return c.many(ctx, c.s.readDB(), q)
}

func (c *collection) Count(ctx context.Context, f store.RecordFilter) (map[string]int, error) {
	w, err := c.where(f)
	if err != nil {
		return nil, err
	}
	text, args, err := bqb.New("SELECT state, COUNT(*) FROM records WHERE ? GROUP BY state", w).ToPgsql()
	if err != nil {
		return nil, err
	}
	rows, err := c.s.readDB().QueryContext(ctx, text, args...)
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
		return rec, fmt.Errorf("%w: %s: claim needs a filter, a holder and a lease", domain.ErrInvalid, c.spec.Name)
	}
	err = c.s.writeTx(ctx, func(tx *sql.Tx) error {
		for range claimTries {
			now := c.s.Now()
			w, err := c.where(f)
			if err != nil {
				return err
			}
			text, args, err := bqb.New("SELECT key FROM records WHERE ? AND (holder = '' OR lease_until <= ?) ORDER BY created_at, key LIMIT 1", w, ns(now)).ToPgsql()
			if err != nil {
				return err
			}
			var key string
			if err := tx.QueryRowContext(ctx, text, args...).Scan(&key); err != nil {
				return c.notFound(err, "claimable record")
			}
			rec, err = c.one(ctx, tx, bqb.New("UPDATE records SET holder = ?, lease_until = ?, attempts = attempts + 1, revision = revision + 1, updated_at = ?"+
				" WHERE collection = ? AND key = ? AND (holder = '' OR lease_until <= ?) RETURNING "+columns,
				holder, ns(now.Add(lease)), ns(now), c.spec.Name, key, ns(now)))
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
	err = c.s.writeTx(ctx, func(tx *sql.Tx) error {
		now := c.s.Now()
		rec, err = c.one(ctx, tx, bqb.New("UPDATE records SET lease_until = ? WHERE collection = ?"+
			" AND key = ? AND holder = ? AND holder != '' AND lease_until > ? RETURNING "+columns,
			ns(now.Add(lease)), c.spec.Name, key, holder, ns(now)))
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
	w, err := c.where(f)
	if err != nil {
		return 0, err
	}
	// The GIN expression index follows the row: nothing else to delete.
	err = c.s.writeTx(ctx, func(tx *sql.Tx) error {
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
	v := tsv(c.spec)
	// As the recall query: ts_rank divided by the log of the word count,
	// so a short focused match beats a long one, as bm25 does.
	nwords := "cardinality(regexp_split_to_array(trim(" + searchText(c.spec) + "), '\\s+'))"
	out := []store.Record{}
	for _, m := range recallMatch(words) {
		q := bqb.New("SELECT "+columns+" FROM records, to_tsquery('english', ?) q WHERE collection = ? AND "+v+
			" @@ q ORDER BY ts_rank("+v+", q) / ln(2 + "+nwords+") DESC, created_at, key LIMIT ?", m, c.spec.Name, limit)
		var err error
		if out, err = c.many(ctx, c.s.readDB(), q); err != nil {
			return nil, err
		}
		if len(out) > 0 {
			break
		}
	}
	return out, nil
}
