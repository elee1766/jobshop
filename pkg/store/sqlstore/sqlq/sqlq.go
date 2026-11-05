// Package sqlq renders the SQL for store collections (#579) from a
// CollectionSpec and a RecordFilter, for SQLite and Postgres. It is plain
// SQL in Go, not sqlc: the statements depend on the spec's declared
// fields, which sqlc cannot express.
//
// Identifiers (collection and field names, JSON paths) are written into
// the SQL text, so they must come from a spec that passed
// CollectionSpec.Validate. Values are always bound. The collection name
// is written as a literal, not bound, because SQLite uses a partial
// index only when the query's WHERE implies the index's.
package sqlq

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/elee1766/jobshop/pkg/store"
)

// Dialect is the engine-specific part of the SQL.
type Dialect interface {
	// Ph returns the placeholder for the n-th argument, from 1.
	Ph(n int) string
	// Field returns the text expression of a JSON path in value. Index
	// and query use the same expression, so the index applies.
	Field(path string) string
	// ValueIn wraps the placeholder for a Value argument.
	ValueIn(ph string) string
	// ValueOut is the expression that reads value as text.
	ValueOut(prefix string) string
	// Register returns the DDL that creates spec's partition (Postgres),
	// field indexes and search index if missing.
	Register(spec store.CollectionSpec) []string
	// SearchIndex returns the statements that refresh key's search
	// document after a write, each taking the key as its only argument.
	SearchIndex(spec store.CollectionSpec) []string
	// SearchDelete returns a statement removing the search documents of
	// the keys the subquery selects ("" if the index follows the row).
	SearchDelete(spec store.CollectionSpec, keys string) string
	// Search returns the full-text query: argument 1 is a match
	// expression, argument 2 the limit.
	Search(spec store.CollectionSpec) string
}

// Columns lists the record columns Scan reads, in order.
func Columns(d Dialect, prefix string) string {
	p := prefix
	return p + "key, " + p + "state, " + d.ValueOut(p) + ", " + p + "revision, " + p + "holder, " +
		p + "lease_until, " + p + "attempts, " + p + "created_at, " + p + "updated_at"
}

// Lit quotes a validated name as an SQL string literal.
func Lit(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// Query is SQL text and its arguments.
type Query struct {
	SQL  string
	Args []any
}

// builder collects arguments and hands out their placeholders.
type builder struct {
	d    Dialect
	args []any
}

func (b *builder) arg(v any) string {
	b.args = append(b.args, v)
	return b.d.Ph(len(b.args))
}

func (b *builder) in(col string, vs []string) string {
	ph := make([]string, len(vs))
	for i, v := range vs {
		ph[i] = b.arg(v)
	}
	return col + " IN (" + strings.Join(ph, ", ") + ")"
}

func ms(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

// where renders the WHERE conditions of f for spec, with the column
// prefix p. It returns ErrInvalid for an undeclared Where field.
func (b *builder) where(spec store.CollectionSpec, f store.RecordFilter, p string) (string, error) {
	conds := []string{p + "collection = " + Lit(spec.Name)}
	if len(f.Keys) > 0 {
		conds = append(conds, b.in(p+"key", f.Keys))
	}
	if len(f.States) > 0 {
		conds = append(conds, b.in(p+"state", f.States))
	}
	names := make([]string, 0, len(f.Where))
	for n := range f.Where {
		names = append(names, n)
	}
	slices.Sort(names)
	for _, n := range names {
		fld, unique, ok := spec.Lookup(n)
		if !ok {
			return "", fmt.Errorf("%w: collection %s: field %q is not declared", store.ErrInvalid, spec.Name, n)
		}
		conds = append(conds, b.d.Field(fld.Path)+" = "+b.arg(f.Where[n]))
		if unique {
			// The unique index is partial on non-empty values; repeat its
			// condition so the planner can use it.
			conds = append(conds, b.d.Field(fld.Path)+" != ''")
		}
	}
	if !f.Since.IsZero() {
		conds = append(conds, p+"created_at >= "+b.arg(ms(f.Since)))
	}
	if !f.Until.IsZero() {
		conds = append(conds, p+"created_at < "+b.arg(ms(f.Until)))
	}
	if !f.UpdatedBefore.IsZero() {
		conds = append(conds, p+"updated_at < "+b.arg(ms(f.UpdatedBefore)))
	}
	return strings.Join(conds, " AND "), nil
}

func order(o string) (string, error) {
	switch o {
	case "", store.OrderCreated:
		return "created_at, key", nil
	case store.OrderCreatedDesc:
		return "created_at DESC, key DESC", nil
	case store.OrderUpdated:
		return "updated_at, key", nil
	}
	return "", fmt.Errorf("%w: record order %q", store.ErrInvalid, o)
}

// List selects the records matching f.
func List(d Dialect, spec store.CollectionSpec, f store.RecordFilter) (Query, error) {
	b := &builder{d: d}
	w, err := b.where(spec, f, "")
	if err != nil {
		return Query{}, err
	}
	ord, err := order(f.Order)
	if err != nil {
		return Query{}, err
	}
	lim := f.Limit
	if lim <= 0 {
		lim = store.DefaultLimit
	}
	sql := "SELECT " + Columns(d, "") + " FROM records WHERE " + w + " ORDER BY " + ord + " LIMIT " + b.arg(lim)
	return Query{sql, b.args}, nil
}

// Count counts the records matching f by state.
func Count(d Dialect, spec store.CollectionSpec, f store.RecordFilter) (Query, error) {
	b := &builder{d: d}
	w, err := b.where(spec, f, "")
	if err != nil {
		return Query{}, err
	}
	return Query{"SELECT state, COUNT(*) FROM records WHERE " + w + " GROUP BY state", b.args}, nil
}

// Delete removes the records matching f. The second query, when not
// empty, removes their search documents and runs first.
func Delete(d Dialect, spec store.CollectionSpec, f store.RecordFilter) (del, search Query, err error) {
	b := &builder{d: d}
	w, err := b.where(spec, f, "")
	if err != nil {
		return Query{}, Query{}, err
	}
	del = Query{"DELETE FROM records WHERE " + w, b.args}
	if len(spec.Search) > 0 {
		if s := d.SearchDelete(spec, "SELECT key FROM records WHERE "+w); s != "" {
			search = Query{s, b.args}
		}
	}
	return del, search, nil
}

// NextClaimable selects the key of the oldest record matching f that is
// unleased at now.
func NextClaimable(d Dialect, spec store.CollectionSpec, f store.RecordFilter, now time.Time) (Query, error) {
	b := &builder{d: d}
	w, err := b.where(spec, f, "")
	if err != nil {
		return Query{}, err
	}
	sql := "SELECT key FROM records WHERE " + w +
		" AND (holder = '' OR lease_until <= " + b.arg(ms(now)) + ") ORDER BY created_at, key LIMIT 1"
	return Query{sql, b.args}, nil
}

// Take leases key to holder if it is still unleased at now.
func Take(d Dialect, spec store.CollectionSpec, key, holder string, now, until time.Time) Query {
	b := &builder{d: d}
	sql := "UPDATE records SET holder = " + b.arg(holder) + ", lease_until = " + b.arg(ms(until)) +
		", attempts = attempts + 1, revision = revision + 1, updated_at = " + b.arg(ms(now)) +
		" WHERE collection = " + Lit(spec.Name) + " AND key = " + b.arg(key) +
		" AND (holder = '' OR lease_until <= " + b.arg(ms(now)) + ") RETURNING " + Columns(d, "")
	return Query{sql, b.args}
}

// Renew extends holder's live lease on key.
func Renew(d Dialect, spec store.CollectionSpec, key, holder string, now, until time.Time) Query {
	b := &builder{d: d}
	sql := "UPDATE records SET lease_until = " + b.arg(ms(until)) +
		" WHERE collection = " + Lit(spec.Name) + " AND key = " + b.arg(key) +
		" AND holder = " + b.arg(holder) + " AND holder != '' AND lease_until > " + b.arg(ms(now)) +
		" RETURNING " + Columns(d, "")
	return Query{sql, b.args}
}

// Get selects the record with key.
func Get(d Dialect, spec store.CollectionSpec, key string) Query {
	b := &builder{d: d}
	return Query{"SELECT " + Columns(d, "") + " FROM records WHERE collection = " + Lit(spec.Name) +
		" AND key = " + b.arg(key), b.args}
}

// ByField selects the record whose unique field has value.
func ByField(d Dialect, spec store.CollectionSpec, f store.Field, value string) Query {
	b := &builder{d: d}
	return Query{"SELECT " + Columns(d, "") + " FROM records WHERE collection = " + Lit(spec.Name) +
		" AND " + d.Field(f.Path) + " = " + b.arg(value) + " AND " + d.Field(f.Path) + " != '' LIMIT 1", b.args}
}

// Insert adds r unless its key or a unique field exists.
func Insert(d Dialect, spec store.CollectionSpec, r store.Record, now time.Time) Query {
	b := &builder{d: d}
	created := r.CreatedAt
	if created.IsZero() {
		created = now
	}
	sql := "INSERT INTO records (collection, key, state, value, revision, holder, lease_until, attempts, created_at, updated_at) VALUES (" +
		Lit(spec.Name) + ", " + b.arg(r.Key) + ", " + b.arg(r.State) + ", " + d.ValueIn(b.arg(string(r.Value))) +
		", 1, '', 0, 0, " + b.arg(ms(created)) + ", " + b.arg(ms(now)) + ") ON CONFLICT DO NOTHING"
	return Query{sql, b.args}
}

// Update writes r's state and value if its revision matches and the
// lease allows it (see store.Collection.Update).
func Update(d Dialect, spec store.CollectionSpec, r store.Record, now time.Time) Query {
	b := &builder{d: d}
	set := "state = " + b.arg(r.State) + ", value = " + d.ValueIn(b.arg(string(r.Value))) +
		", revision = revision + 1, updated_at = " + b.arg(ms(now))
	if r.LeaseUntil.IsZero() {
		set += ", holder = '', lease_until = 0"
	}
	sql := "UPDATE records SET " + set + " WHERE collection = " + Lit(spec.Name) +
		" AND key = " + b.arg(r.Key) + " AND revision = " + b.arg(r.Revision)
	// Placeholders bind in text order on SQLite: add args as they appear.
	if r.Holder == "" {
		sql += " AND (holder = '' OR lease_until <= " + b.arg(ms(now)) + ")"
	} else {
		sql += " AND holder = " + b.arg(r.Holder) + " AND lease_until > " + b.arg(ms(now))
	}
	return Query{sql + " RETURNING " + Columns(d, ""), b.args}
}

// SQLite is the SQLite dialect.
var SQLite Dialect = sqliteDialect{}

// Postgres is the Postgres dialect.
var Postgres Dialect = postgresDialect{}

type sqliteDialect struct{}

func (sqliteDialect) Ph(int) string { return "?" }

func (sqliteDialect) Field(path string) string {
	return "CAST(json_extract(value, '$." + path + "') AS TEXT)"
}

func (sqliteDialect) ValueIn(ph string) string { return ph }

func (sqliteDialect) ValueOut(p string) string { return p + "value" }

func (d sqliteDialect) Register(spec store.CollectionSpec) []string {
	var out []string
	name := Lit(spec.Name)
	for _, f := range spec.Index {
		out = append(out, "CREATE INDEX IF NOT EXISTS ri_"+spec.Name+"_"+f.Name+" ON records ("+
			d.Field(f.Path)+", created_at, key) WHERE collection = "+name)
	}
	for _, f := range spec.Unique {
		out = append(out, "CREATE UNIQUE INDEX IF NOT EXISTS ru_"+spec.Name+"_"+f.Name+" ON records ("+
			d.Field(f.Path)+") WHERE collection = "+name+" AND "+d.Field(f.Path)+" != ''")
	}
	if len(spec.Search) > 0 {
		out = append(out, "CREATE VIRTUAL TABLE IF NOT EXISTS fts_"+spec.Name+" USING fts5(key UNINDEXED, body)")
	}
	return out
}

func (d sqliteDialect) SearchIndex(spec store.CollectionSpec) []string {
	parts := make([]string, len(spec.Search))
	for i, f := range spec.Search {
		parts[i] = "coalesce(" + d.Field(f.Path) + ", '')"
	}
	t := "fts_" + spec.Name
	return []string{
		"DELETE FROM " + t + " WHERE key = ?",
		"INSERT INTO " + t + " (key, body) SELECT key, " + strings.Join(parts, " || char(10) || ") +
			" FROM records WHERE collection = " + Lit(spec.Name) + " AND key = ?1",
	}
}

func (sqliteDialect) SearchDelete(spec store.CollectionSpec, keys string) string {
	return "DELETE FROM fts_" + spec.Name + " WHERE key IN (" + keys + ")"
}

func (d sqliteDialect) Search(spec store.CollectionSpec) string {
	t := "fts_" + spec.Name
	return "SELECT " + Columns(d, "r.") + " FROM " + t + " JOIN records r ON r.collection = " + Lit(spec.Name) +
		" AND r.key = " + t + ".key WHERE " + t + " MATCH ?1 ORDER BY bm25(" + t + "), r.created_at, r.key LIMIT ?2"
}

type postgresDialect struct{}

func (postgresDialect) Ph(n int) string { return "$" + strconv.Itoa(n) }

func (postgresDialect) Field(path string) string {
	return "(value #>> '{" + strings.ReplaceAll(path, ".", ",") + "}')"
}

func (postgresDialect) ValueIn(ph string) string { return "CAST(" + ph + " AS jsonb)" }

func (postgresDialect) ValueOut(p string) string { return "CAST(" + p + "value AS TEXT)" }

// FieldIndex names the parent index of a JSON path. Indexes live on the
// partitioned parent, so every partition gets them, and collections
// that declare the same path share one.
func FieldIndex(path string) string { return "rf_" + strings.ReplaceAll(path, ".", "__") }

func (d postgresDialect) tsv(spec store.CollectionSpec) string {
	parts := make([]string, len(spec.Search))
	for i, f := range spec.Search {
		parts[i] = "coalesce(" + d.Field(f.Path) + ", '')"
	}
	return "to_tsvector('english', " + strings.Join(parts, " || ' ' || ") + ")"
}

func (d postgresDialect) Register(spec store.CollectionSpec) []string {
	part := "records_" + spec.Name
	out := []string{"CREATE TABLE IF NOT EXISTS " + part + " PARTITION OF records FOR VALUES IN (" + Lit(spec.Name) + ")"}
	for _, f := range spec.Index {
		out = append(out, "CREATE INDEX IF NOT EXISTS "+FieldIndex(f.Path)+" ON records ("+d.Field(f.Path)+", created_at)")
	}
	// A unique index on a partitioned parent must include the partition
	// key, so each one lives on its collection's partition.
	for _, f := range spec.Unique {
		out = append(out, "CREATE UNIQUE INDEX IF NOT EXISTS ru_"+spec.Name+"_"+f.Name+" ON "+part+" ("+
			d.Field(f.Path)+") WHERE "+d.Field(f.Path)+" != ''")
	}
	if len(spec.Search) > 0 {
		out = append(out, "CREATE INDEX IF NOT EXISTS rs_"+spec.Name+" ON "+part+" USING gin ("+d.tsv(spec)+")")
	}
	return out
}

// SearchIndex is empty: the GIN expression index follows the row.
func (postgresDialect) SearchIndex(store.CollectionSpec) []string { return nil }

func (postgresDialect) SearchDelete(store.CollectionSpec, string) string { return "" }

func (d postgresDialect) Search(spec store.CollectionSpec) string {
	tsv := d.tsv(spec)
	// As the recall query: ts_rank divided by the log of the word count,
	// so a short focused match beats a long one, as bm25 does.
	parts := make([]string, len(spec.Search))
	for i, f := range spec.Search {
		parts[i] = "coalesce(" + d.Field(f.Path) + ", '')"
	}
	words := "cardinality(regexp_split_to_array(trim(" + strings.Join(parts, " || ' ' || ") + "), '\\s+'))"
	return "SELECT " + Columns(d, "") + " FROM records, to_tsquery('english', $1) q WHERE collection = " +
		Lit(spec.Name) + " AND " + tsv + " @@ q ORDER BY ts_rank(" + tsv + ", q) / ln(2 + " + words +
		") DESC, created_at, key LIMIT $2"
}
