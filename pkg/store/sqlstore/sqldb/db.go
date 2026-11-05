// Package sqldb holds the ticket store's SQL, written for SQLite and
// rewritten for Postgres when it runs (pkg/store/postgres/rewrite.go).
// It started as sqlc output and is edited by hand since #579 removed
// sqlc. Each query text keeps its leading "-- name: X" line: Postgres
// overrides are looked up by that name.
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
