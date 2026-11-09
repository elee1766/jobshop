// Package sqlite is the modernc.org/sqlite store backend (JOB-42). It
// owns its SQL (sqldb, list.go), transactions and connections; it shares
// only the pkg/store types and the pure rules in storerules with the
// Postgres store. Schema lives in migrations/ (goose).
package sqlite
