// Package storemigrate copies a SQLite ticket store into Postgres
// (#632). Both stores are opened and migrated by their backends first,
// so the copy runs between two schemas at the same version; it copies
// rows table by table over database/sql and never goes through the
// store API, so ids, revisions, history and timestamps are kept.
package storemigrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/elee1766/jobshop/pkg/store"
)

// Tables are the copied tables in dependency order: tickets before the
// tables that reference them. A table added to both schemas must be
// added here; TestTablesComplete fails otherwise.
var Tables = []string{
	"tickets",
	"edges",
	"labels",
	"history",
	"comments",
	"recall",
	"resource_caps",
	"ticket_meta",
	"resource_providers",
	"resource_slots",
	"ticket_resources",
	"records",
	"id_seqs",
}

// identity lists tables whose id column is a Postgres identity; its
// sequence is moved past the copied ids.
var identity = map[string]string{
	"history":  "seq",
	"comments": "id",
	"recall":   "id",
}

// ErrNotEmpty is returned when the target holds rows and Force is off.
var ErrNotEmpty = errors.New("target store is not empty")

// Options configures Copy.
type Options struct {
	// Force empties the target first instead of refusing.
	Force bool
}

// Result is the copied row count per table.
type Result map[string]int64

// Copy copies every row of src (a migrated SQLite database) into dst (a
// migrated Postgres database) in one Postgres transaction, then compares
// row counts per table and rolls back on a mismatch. The transaction is
// serializable: Postgres connections default to read committed since
// JOB-41, and the count check must see only this copy's rows.
func Copy(ctx context.Context, src, dst *sql.DB, opt Options) (Result, error) {
	tx, err := dst.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	if err := prepareTarget(ctx, tx, opt.Force); err != nil {
		return nil, err
	}
	for _, t := range Tables {
		if err := copyTable(ctx, src, tx, t); err != nil {
			return nil, fmt.Errorf("copy %s: %w", t, err)
		}
	}
	for t, col := range identity {
		q := fmt.Sprintf("SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE(MAX(%s), 0) + 1, false) FROM %s", t, col, col, t)
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return nil, fmt.Errorf("reset %s sequence: %w", t, err)
		}
	}
	res := Result{}
	var bad []string
	for _, t := range Tables {
		want, err := count(ctx, src, t)
		if err != nil {
			return nil, err
		}
		got, err := count(ctx, tx, t)
		if err != nil {
			return nil, err
		}
		res[t] = got
		if got != want {
			bad = append(bad, fmt.Sprintf("%s: sqlite %d, postgres %d", t, want, got))
		}
	}
	if len(bad) > 0 {
		return res, fmt.Errorf("row counts differ, nothing written: %s", strings.Join(bad, "; "))
	}
	return res, tx.Commit()
}

type queryer interface {
	QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row
}

func count(ctx context.Context, db queryer, t string) (int64, error) {
	var n int64
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM "+t).Scan(&n); err != nil {
		return 0, fmt.Errorf("count %s: %w", t, err)
	}
	return n, nil
}

func prepareTarget(ctx context.Context, tx *sql.Tx, force bool) error {
	if force {
		_, err := tx.ExecContext(ctx, "TRUNCATE "+strings.Join(Tables, ", ")+" RESTART IDENTITY CASCADE")
		return err
	}
	var full []string
	for _, t := range Tables {
		n, err := count(ctx, tx, t)
		if err != nil {
			return err
		}
		if n > 0 {
			full = append(full, fmt.Sprintf("%s (%d)", t, n))
		}
	}
	if len(full) > 0 {
		return fmt.Errorf("%w: %s; use --force to replace it", ErrNotEmpty, strings.Join(full, ", "))
	}
	return nil
}

// source returns the query that reads t from SQLite. Comment recall
// rows are written by the Postgres trigger as comments are copied, so
// no other recall rows remain to read (journals and learnings are
// records, #579); recall ids are not kept. Postgres searches records
// through an index, so SQLite's fts_* tables are not copied.
func source(t string) string {
	switch t {
	case "recall":
		return "SELECT body, kind, ref, ticket_id, project, at FROM recall WHERE kind != 'comment'"
	case "tickets":
		// Parents are set after every ticket is in (self reference).
		return "SELECT * FROM tickets ORDER BY created_at, id"
	}
	return "SELECT * FROM " + t
}

func copyTable(ctx context.Context, src *sql.DB, tx *sql.Tx, t string) error {
	types, err := targetTypes(ctx, tx, t)
	if err != nil {
		return err
	}
	rows, err := src.QueryContext(ctx, source(t))
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	parent := -1
	marks := make([]string, len(cols))
	quoted := make([]string, len(cols))
	for i, c := range cols {
		typ, ok := types[c]
		if !ok {
			return fmt.Errorf("column %s missing in postgres", c)
		}
		marks[i] = "$" + strconv.Itoa(i+1)
		if typ == "jsonb" {
			marks[i] = "CAST(" + marks[i] + " AS jsonb)"
		}
		quoted[i] = `"` + c + `"`
		if t == "tickets" && c == "parent_id" {
			parent = i
		}
	}
	ins, err := tx.PrepareContext(ctx, "INSERT INTO "+t+" ("+strings.Join(quoted, ", ")+") VALUES ("+strings.Join(marks, ", ")+")")
	if err != nil {
		return err
	}
	defer func() { _ = ins.Close() }()
	parts := map[string]bool{}
	type link struct{ id, parent any }
	var links []link
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return err
		}
		args := make([]any, len(cols))
		for i, v := range vals {
			if args[i], err = convert(v, types[cols[i]]); err != nil {
				return fmt.Errorf("column %s: %w", cols[i], err)
			}
		}
		if parent >= 0 && args[parent] != nil {
			links = append(links, link{id: args[0], parent: args[parent]})
			args[parent] = nil
		}
		if t == "records" {
			name, _ := args[0].(string)
			if !parts[name] {
				if err := partition(ctx, tx, name); err != nil {
					return err
				}
				parts[name] = true
			}
		}
		if _, err := ins.ExecContext(ctx, args...); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if parent >= 0 && cols[0] != "id" {
		return errors.New("tickets: first column is not id")
	}
	for _, l := range links {
		if _, err := tx.ExecContext(ctx, "UPDATE tickets SET parent_id = $1 WHERE id = $2", l.parent, l.id); err != nil {
			return err
		}
	}
	return nil
}

// partition creates a collection's records partition, as registration
// does at controller start; registration adds its indexes.
func partition(ctx context.Context, tx *sql.Tx, name string) error {
	if err := (store.CollectionSpec{Name: name}).Validate(); err != nil {
		return fmt.Errorf("records: collection %q: %w", name, err)
	}
	_, err := tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS records_"+name+" PARTITION OF records FOR VALUES IN ('"+name+"')")
	return err
}

// targetTypes maps t's Postgres columns to their data types.
func targetTypes(ctx context.Context, tx *sql.Tx, t string) (map[string]string, error) {
	rows, err := tx.QueryContext(ctx, "SELECT column_name, data_type FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = $1", t)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := map[string]string{}
	for rows.Next() {
		var c, typ string
		if err := rows.Scan(&c, &typ); err != nil {
			return nil, err
		}
		out[c] = typ
	}
	if len(out) == 0 && rows.Err() == nil {
		return nil, fmt.Errorf("no postgres table %s", t)
	}
	return out, rows.Err()
}

// convert turns a SQLite value into one Postgres takes for a column of
// type typ. SQLite is loosely typed: an integer may sit in a TEXT
// column (recall.at), and text may come back as bytes.
func convert(v any, typ string) (any, error) {
	switch x := v.(type) {
	case nil:
		return nil, nil
	case []byte:
		v = string(x)
	}
	text := typ == "text" || typ == "jsonb"
	switch x := v.(type) {
	case int64:
		if text {
			return strconv.FormatInt(x, 10), nil
		}
	case float64:
		if text {
			return strconv.FormatFloat(x, 'f', -1, 64), nil
		}
	case bool:
		if !text {
			if x {
				return int64(1), nil
			}
			return int64(0), nil
		}
	case string:
		if typ == "bigint" {
			n, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("text %q in an integer column", x)
			}
			return n, nil
		}
	}
	return v, nil
}
