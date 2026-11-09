package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/elee1766/jobshop/pkg/storage/store/postgres"
	"github.com/elee1766/jobshop/pkg/storage/store/storetest"
)

// BenchmarkStore runs the hot-path benchmarks (JOB-43) on a database
// cloned from the template. Skips without JOBSHOP_TEST_POSTGRES. Not
// part of CI: go test -run '^$' -bench .
func BenchmarkStore(b *testing.B) {
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		b.Skip(dsnEnv + " is not set")
	}
	name := fmt.Sprintf("jobshop_bench_%d_%d", os.Getpid(), seq.Add(1))
	ctx := context.WithoutCancel(b.Context())
	_, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE "+templateDB)
	require.NoError(b, err)
	s, err := postgres.Open(ctx, withDB(dsn, name), postgres.Options{})
	require.NoError(b, err)
	b.Cleanup(func() {
		_ = s.Close()
		_, _ = admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	})
	storetest.Bench(b, s, storetest.DefaultBenchSize())
}
