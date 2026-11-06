// Package pgdb holds the Postgres store's hand-written SQL (JOB-41),
// separate from the SQLite queries in sqlstore/sqldb so each engine can
// use its own features (FOR UPDATE SKIP LOCKED for claims). Each query
// text keeps a leading "-- name: X" line, which TestEveryQueryPrepares
// reports failures by.
package pgdb

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
