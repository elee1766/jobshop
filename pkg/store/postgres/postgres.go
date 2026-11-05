// Package postgres is the Postgres store backend (#493). It runs the
// backend-neutral core in pkg/store/sqlstore on the same sqlc queries as
// pkg/store/sqlite: the generated SQLite SQL is rewritten for Postgres
// when it runs (rewrite.go), and the few queries that use an engine
// feature (full-text search) have Postgres versions in queries/.
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
	"github.com/pressly/goose/v3"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlcdb"
)

//go:embed migrations/*.sql
var migrations embed.FS

//go:embed queries/overrides.sql
var overridesSQL string

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
	rw := newRewriter(parseOverrides(overridesSQL))
	s, err := sqlstore.New(sqlstore.Driver{
		DB:          db,
		Bind:        func(d sqlcdb.DBTX) sqlcdb.Querier { return sqlcdb.New(pgDB{d: d, rw: rw}) },
		Classify:    classify,
		RecallMatch: recallMatch,
		IsBusy:      isBusy,
	}, opt)
	if err != nil {
		_ = db.Close()
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

// pgDB runs sqlc's queries on d with the SQL rewritten for Postgres.
type pgDB struct {
	d  sqlcdb.DBTX
	rw *rewriter
}

func (p pgDB) ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error) {
	return p.d.ExecContext(ctx, p.rw.rewrite(q), args...)
}

func (p pgDB) PrepareContext(ctx context.Context, q string) (*sql.Stmt, error) {
	return p.d.PrepareContext(ctx, p.rw.rewrite(q))
}

func (p pgDB) QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error) {
	return p.d.QueryContext(ctx, p.rw.rewrite(q), args...)
}

func (p pgDB) QueryRowContext(ctx context.Context, q string, args ...any) *sql.Row {
	return p.d.QueryRowContext(ctx, p.rw.rewrite(q), args...)
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
