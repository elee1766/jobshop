package sqlite_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
)

// template is a database migrated once per test binary. Running the
// migrations per case was most of the package's time under -race (#363),
// so each test copies this file and Open finds nothing to apply.
var template []byte

func TestMain(m *testing.M) {
	code, err := run(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(code)
}

func run(m *testing.M) (int, error) {
	dir, err := os.MkdirTemp("", "jobshop-sqlite-template")
	if err != nil {
		return 0, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	path := filepath.Join(dir, "template.db")
	s, err := sqlite.Open(context.Background(), path, sqlite.Options{})
	if err != nil {
		return 0, err
	}
	// Close checkpoints the WAL, so the main file holds the whole schema.
	if err := s.Close(); err != nil {
		return 0, err
	}
	if template, err = os.ReadFile(path); err != nil {
		return 0, err
	}
	return m.Run(), nil
}

// testPath returns a fresh database path in t's temp dir holding a copy
// of the migrated template.
func testPath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "j.db")
	require.NoError(t, os.WriteFile(path, template, 0o600))
	return path
}

// openTest opens a migrated store from the template. The caller closes it.
func openTest(t *testing.T, opt sqlite.Options) *sqlite.Store {
	t.Helper()
	s, err := sqlite.Open(t.Context(), testPath(t), opt)
	require.NoError(t, err)
	return s
}
