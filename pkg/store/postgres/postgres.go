// Package postgres is the Postgres store backend (#493). It runs the
// backend-neutral core in pkg/store/sqlstore on the same sqldb queries as
// pkg/store/sqlite: sqldb's SQL runs on both engines (#694). Full-text
// search, the one engine-specific query, is recallSQL.
package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/nullism/bqb"
	"github.com/pressly/goose/v3"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqldb"
)

//go:embed migrations/*.sql
var migrations embed.FS

// recallSQL is Driver.RecallSQL. Lower score is better, as with bm25.
// ts_rank's own length normalization counts lexemes, which drops
// stopwords; bm25 counts every word, so the rank is divided by the log
// of the body's word count instead: a short focused match beats a long
// one.
const recallSQL = `SELECT kind, ref, ticket_id, project, CAST(at AS BIGINT) AS at_ms,
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
type Options = sqlstore.Options

// Store is the Postgres-backed store.
type Store = sqlstore.Store

// maxConns bounds the pool. Writers run serializable and retry on a
// serialization failure, so more connections mostly add retries.
const maxConns = 8

// Open connects to the database at dsn (a libpq URL or key=value
// string) and migrates it.
func Open(ctx context.Context, dsn string, opt Options) (*Store, error) {
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres dsn: %w", err)
	}
	// Serializable: the claim and sweep guards (#406) read other rows in
	// the same statement as the write; under read committed a concurrent
	// claim could pass them on a stale snapshot. A serialization failure
	// is reported as busy and the core reruns the transaction.
	if cfg.RuntimeParams == nil {
		cfg.RuntimeParams = map[string]string{}
	}
	cfg.RuntimeParams["default_transaction_isolation"] = "serializable"
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
	// Reads outside a transaction (Get, List) run read committed: under
	// serializable a plain read can fail with 40001 while writers run
	// (JOB-38), and a read has no guard to protect.
	rcfg := cfg.Copy()
	rcfg.RuntimeParams["default_transaction_isolation"] = "read committed"
	rdb := stdlib.OpenDB(*rcfg)
	rdb.SetMaxOpenConns(maxConns)
	rdb.SetMaxIdleConns(maxConns)
	s, err := sqlstore.New(sqlstore.Driver{
		DB:          db,
		ReadDB:      rdb,
		Bind:        func(d sqldb.DBTX) sqldb.Querier { return sqldb.New(d) },
		Classify:    classify,
		RecallMatch: recallMatch,
		RecallSQL:   recallSQL,
		IsBusy:      isBusy,
		Collection:  (&registry{}).open,
		Render:      (*bqb.Query).ToPgsql,
	}, opt)
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
		return store.ErrAlreadyExist
	case "23503":
		return store.ErrNotFound
	case "23514":
		return store.ErrInvalid
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
