package sqlite_test

import (
	"database/sql"
	"testing"

	"github.com/gfx-labs/jobshop/pkg/storage/store/sqlite"
)

// openAt21 opens the database at path and runs the Down of every
// migration after 00021, so tests that undo older migrations by hand
// start from the schema they were written against.
func openAt21(t *testing.T, path string) (*sql.DB, error) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		return nil, err
	}
	if err := sqlite.MigrateDownTo(t.Context(), db, 21); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}
