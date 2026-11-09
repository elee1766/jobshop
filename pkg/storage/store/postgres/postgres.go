// Package postgres is the Postgres store backend (#493, JOB-41). It owns
// its SQL (pgdb, list.go), transactions, retries and connection pools;
// it shares only the pkg/storage/store types and the pure rules in storerules
// with the SQLite store.
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"time"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/elee1766/jobshop/pkg/domain"
	"github.com/elee1766/jobshop/pkg/ids"
)

//go:embed migrations/*.sql
var migrations embed.FS

// recallSQL searches comments: $1 a to_tsquery expression, $2 project
// and $3 ticket id (” for any), $4 the limit. Lower score is better.
// ts_rank's own length normalization counts lexemes, which drops
// stopwords; it is divided by the log of the body's word count
// instead, so a short focused match beats a long one.
const recallSQL = `SELECT kind, ref, ticket_id, project, CAST(at AS BIGINT) AS at_ns,
       ts_headline('english', body, q,
         'StartSel=[, StopSel=], MaxWords=16, MinWords=4, FragmentDelimiter=...') AS snippet,
       CAST(-ts_rank(tsv, q) / ln(2 + cardinality(regexp_split_to_array(trim(body), '\s+')))
         AS DOUBLE PRECISION) AS score
FROM recall, to_tsquery('english', $1) q
WHERE tsv @@ q
  AND kind = 'comment'
  AND ($2 = '' OR project = $2)
  AND ($3 = '' OR ticket_id = $3)
ORDER BY score, id
LIMIT CAST($4 AS BIGINT)`

// Options configures Open.
type Options struct {
	// Now overrides the clock, for tests.
	Now func() time.Time
	// Prefix maps a project to its id prefix. Default: the project name.
	Prefix func(project string) string
	// IDs picks a project's id generator (JOB-57). Default: sonyflake.
	IDs func(project string) ids.Spec
	// OnCommit, if set, runs after every committed write transaction.
	// The controller uses it to wake events.subscribe (#147).
	OnCommit func()
}

// maxConns bounds each pool. Writers run serializable and retry on a
// serialization failure, so more write connections mostly add retries;
// reads have their own pool so they never wait on one.
const maxConns = 8

// Open connects to the database at dsn (a libpq URL or key=value
// string) and migrates it.
func Open(ctx context.Context, dsn string, opt Options) (*Store, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres dsn: %w", err)
	}
	// Isolation is set per write transaction (BeginTx); connections keep
	// the server default, read committed (JOB-41).
	db := stdlib.OpenDB(*cfg)
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("postgres: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	// Reads outside a transaction (Get, List, Recall, collection reads)
	// get their own pool, for capacity: Desk reads do not queue behind
	// claims for a write connection. Read-only, so a stray write through
	// it fails instead of running outside the retry loop.
	rcfg := cfg.Copy()
	if rcfg.RuntimeParams == nil {
		rcfg.RuntimeParams = map[string]string{}
	}
	rcfg.RuntimeParams["default_transaction_read_only"] = "on"
	rdb := stdlib.OpenDB(*rcfg)
	rdb.SetMaxOpenConns(maxConns)
	rdb.SetMaxIdleConns(maxConns)
	s, err := newStore(db, rdb, opt)
	if err != nil {
		_ = db.Close()
		_ = rdb.Close()
		return nil, err
	}
	return s, nil
}

// Migrate applies the migrations to db. Tests use it to build a
// template database.
func Migrate(ctx context.Context, db *sql.DB) error { return migrate(ctx, db) }

func migrate(ctx context.Context, db *sql.DB) error {
	sub, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	p, err := goose.NewProvider(goose.DialectPostgres, db, sub)
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// isBusy reports a transaction that lost a race: a serialization
// failure or a deadlock. Both are safe to rerun.
func isBusy(err error) bool {
	var pe *pgconn.PgError
	return errors.As(err, &pe) && (pe.Code == "40001" || pe.Code == "40P01")
}

// classify maps Postgres constraint failures to store errors.
func classify(err error) error {
	var pe *pgconn.PgError
	if !errors.As(err, &pe) {
		return nil
	}
	switch pe.Code {
	case "23505":
		return domain.ErrAlreadyExist
	case "23503":
		return domain.ErrNotFound
	case "23514":
		return domain.ErrInvalid
	}
	return nil
}

// recallMatch builds to_tsquery expressions: all words, then any word.
// Each word is reduced to letters and digits, so query syntax in the
// text (quotes, colons, &, |, !) is matched as text and cannot fail to
// parse, as sqlite quotes FTS5 terms.
func recallMatch(words []string) []string {
	var q []string
	for _, w := range words {
		for _, part := range strings.FieldsFunc(w, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			q = append(q, "'"+part+"'")
		}
	}
	if len(q) == 0 {
		return nil
	}
	out := []string{strings.Join(q, " & ")}
	if len(q) > 1 {
		out = append(out, strings.Join(q, " | "))
	}
	return out
}
