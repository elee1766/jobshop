// Package sqldb holds the SQLite store's hand-written SQL (JOB-42).
// Postgres has its own copy in postgres/pgdb, so each can diverge. Each
// query text keeps a leading "-- name: X" line, which
// TestEveryQueryPrepares reports failures by.
package sqldb

import (
	"context"
	"database/sql"
)

// DBTX is a database or transaction the queries run on.
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// New returns the queries over db.
func New(db DBTX) *Queries {
	return &Queries{db: db}
}

// Queries runs each query on a DBTX.
type Queries struct {
	db DBTX
}
