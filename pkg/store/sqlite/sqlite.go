package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
	msqlite "modernc.org/sqlite" // database/sql driver "sqlite"
	sqlite3 "modernc.org/sqlite/lib"

	"github.com/elee1766/jobshop/pkg/store"
	"github.com/elee1766/jobshop/pkg/store/sqlstore"
	"github.com/elee1766/jobshop/pkg/store/sqlstore/sqlq"
)

//go:generate go tool sqlc generate

//go:embed migrations/*.sql
var migrations embed.FS

// Options configures Open.
type Options = sqlstore.Options

// Store is the SQLite-backed store. In production all access goes
// through one connection, but claims and sweeps do not rely on that
// (#406).
type Store = sqlstore.Store

// Open opens (creating if needed) the database at path and migrates it.
// Use ":memory:" for a private in-memory database.
func Open(ctx context.Context, path string, opt Options) (*Store, error) {
	// _txlock=immediate takes the write lock at BEGIN, so a read-then-write
	// transaction cannot fail with SQLITE_BUSY at its first write.
	// Concurrent (tests, #406) uses deferred locks on several connections.
	lock, conns := "immediate", 1
	if opt.Concurrent {
		lock, conns = "deferred", 8
	}
	seed(ctx, path)
	dsn := "file:" + path + "?_txlock=" + lock +
		"&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)" +
		"&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(conns)
	// database/sql keeps 2 idle connections by default: with more open,
	// each extra one was closed after use and reopened, rerunning the
	// pragmas and every statement prepare (#433).
	db.SetMaxIdleConns(conns)
	db.SetConnMaxIdleTime(0)
	db.SetConnMaxLifetime(0)
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	// Statements are prepared once per text and closed with db (#433).
	cache := newStmtCache(db)
	s, err := sqlstore.New(sqlstore.Driver{
		DB:          db,
		Bind:        cache.bind,
		Classify:    classify,
		RecallMatch: recallMatch,
		IsBusy:      isBusy,
		Dialect:     sqlq.SQLite,
	}, opt)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	p, err := goose.NewProvider(goose.DialectSQLite3, db, mustSub(migrations, "migrations"))
	if err != nil {
		return fmt.Errorf("migrations: %w", err)
	}
	if _, err := p.Up(ctx); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// isBusy reports SQLITE_BUSY on a deferred read-to-write upgrade, which
// busy_timeout does not wait on.
func isBusy(err error) bool {
	var se *msqlite.Error
	return errors.As(err, &se) && se.Code()&0xff == sqlite3.SQLITE_BUSY
}

// classify maps SQLite constraint failures to store errors.
func classify(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "UNIQUE constraint failed"), strings.Contains(msg, "PRIMARY KEY"):
		return store.ErrAlreadyExist
	case strings.Contains(msg, "FOREIGN KEY constraint failed"):
		return store.ErrNotFound
	case strings.Contains(msg, "CHECK constraint failed"):
		return store.ErrInvalid
	}
	return nil
}

// recallMatch quotes each word so FTS5 syntax in a query (quotes, colons,
// hyphens, AND/NOT) is matched as text instead of failing to parse. All
// words must match first; failing that any word may, so a long question
// still finds something.
func recallMatch(words []string) []string {
	q := make([]string, len(words))
	for i, w := range words {
		q[i] = `"` + strings.ReplaceAll(w, `"`, `""`) + `"`
	}
	out := []string{strings.Join(q, " ")}
	if len(q) > 1 {
		out = append(out, strings.Join(q, " OR "))
	}
	return out
}
