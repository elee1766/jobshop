package sqlite_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/storage/store/sqlite"
	"github.com/elee1766/jobshop/pkg/storage/store/storetest"
)

// BenchmarkStore runs the hot-path benchmarks (JOB-43) on a file-backed
// database, as a host runs it. Not part of CI: go test -run '^$' -bench .
func BenchmarkStore(b *testing.B) {
	path := filepath.Join(b.TempDir(), "j.db")
	require.NoError(b, os.WriteFile(path, template, 0o600))
	s, err := sqlite.Open(b.Context(), path, sqlite.Options{})
	require.NoError(b, err)
	b.Cleanup(func() { _ = s.Close() })
	storetest.Bench(b, s, storetest.DefaultBenchSize())
}
