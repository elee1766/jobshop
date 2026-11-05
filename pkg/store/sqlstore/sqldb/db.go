// Package sqldb holds the ticket store's hand-written SQL. Every query
// runs as is on SQLite and Postgres (#694): $N placeholders, BIGINT
// casts, booleans cast to INT before BIGINT. Engine-specific SQL (full-
// text search, DDL) lives in the backend packages. Each query text
// keeps a leading "-- name: X" line, which TestEveryQueryPrepares
// reports failures by.
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
