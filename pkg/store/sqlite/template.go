package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"strings"
	"sync"
)

// A fresh database file is copied from a migrated in-memory template
// instead of running every migration again (#433). Production opens one
// database per process, so this saves nothing there; the control tests
// open one per test, and running the migrations was a quarter of their
// CPU.
var schemaTmpl struct {
	once sync.Once
	mu   sync.Mutex // VACUUM INTO runs one at a time on the template
	db   *sql.DB
}

// templateDB returns the migrated in-memory template, or nil if it could
// not be built.
func templateDB(ctx context.Context) *sql.DB {
	schemaTmpl.once.Do(func() {
		// One connection: each connection to :memory: is its own database.
		db, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)")
		if err != nil {
			return
		}
		db.SetMaxOpenConns(1)
		db.SetConnMaxIdleTime(0)
		db.SetConnMaxLifetime(0)
		if err := migrate(context.WithoutCancel(ctx), db); err != nil {
			_ = db.Close()
			return
		}
		schemaTmpl.db = db
	})
	return schemaTmpl.db
}

// seed writes a migrated copy of the template to path if no file is
// there yet. Any failure leaves path alone: Open then migrates as usual.
func seed(ctx context.Context, path string) {
	if path == "" || path == ":memory:" || strings.HasPrefix(path, "file:") {
		return
	}
	if _, err := os.Stat(path); !errors.Is(err, fs.ErrNotExist) {
		return
	}
	db := templateDB(ctx)
	if db == nil {
		return
	}
	schemaTmpl.mu.Lock()
	defer schemaTmpl.mu.Unlock()
	// Into a temp name first, then linked: a link never replaces a file
	// another opener created meanwhile, and a half-written copy is never
	// at path.
	tmp := path + ".seed"
	_ = os.Remove(tmp)
	defer func() { _ = os.Remove(tmp) }()
	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", tmp); err != nil {
		return
	}
	_ = os.Link(tmp, path)
}
