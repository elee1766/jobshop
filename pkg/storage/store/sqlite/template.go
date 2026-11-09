package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// A fresh database file is a copy of a migrated template instead of
// running every migration again (#433). Production opens one database
// per process, so this saves nothing there; the control tests open one
// per test. The template is built once per process and kept as file
// bytes: a VACUUM INTO per open re-ran every CREATE statement and was
// 14% of the control tests' CPU.
var schemaTmpl struct {
	once sync.Once
	b    []byte
}

// templateBytes returns a migrated, empty database file, or nil if it
// could not be built.
func templateBytes(ctx context.Context) []byte {
	schemaTmpl.once.Do(func() {
		dir, err := os.MkdirTemp("", "jobshop-tmpl")
		if err != nil {
			return
		}
		defer func() { _ = os.RemoveAll(dir) }()
		// One connection: each connection to :memory: is its own database.
		db, err := sql.Open("sqlite", "file::memory:?_pragma=foreign_keys(1)")
		if err != nil {
			return
		}
		defer func() { _ = db.Close() }()
		db.SetMaxOpenConns(1)
		ctx := context.WithoutCancel(ctx)
		if err := migrate(ctx, db); err != nil {
			return
		}
		out := filepath.Join(dir, "db")
		if _, err := db.ExecContext(ctx, "VACUUM INTO ?", out); err != nil {
			return
		}
		b, err := os.ReadFile(out)
		if err != nil {
			return
		}
		schemaTmpl.b = b
	})
	return schemaTmpl.b
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
	b := templateBytes(ctx)
	if b == nil {
		return
	}
	// Into a temp name first, then linked: a link never replaces a file
	// another opener created meanwhile, and a half-written copy is never
	// at path.
	f, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".seed*")
	if err != nil {
		return
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	_, err = f.Write(b)
	if cerr := f.Close(); err != nil || cerr != nil {
		return
	}
	_ = os.Link(tmp, path)
}
